/*
Explorer Platform, a platform for hosting and discovering Minecraft servers.
Copyright (C) 2024 Yannic Rieger <oss@76k.io>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package cni

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	current "github.com/containernetworking/cni/pkg/types/100"
	proxyv1alpha1 "github.com/spacechunks/platform/api/platformd/proxy/v1alpha1"
	"github.com/spacechunks/platform/internal/platformd/workload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	runtimev1 "k8s.io/cri-api/pkg/apis/runtime/v1"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
)

var (
	ErrHostIfaceNotFound  = errors.New("host interface not set")
	ErrIPAMConfigNotFound = errors.New("ipam config not set")
)

type Conf struct {
	types.NetConf
	HostIface           string `json:"hostIface"`
	PlatformdListenSock string `json:"platformdListenSock"`
	CRIListenSock       string `json:"criListenSock"`
}

type CNI struct {
	handler Handler
}

func NewCNI(h Handler) *CNI {
	return &CNI{
		handler: h,
	}
}

// ExecAdd sets up the veth pair for a container.
// internally the following happens:
// * first allocated ip address for host side veth using cni ipam plugin.
// * then create veth pair and move one peer into the containers netns.
// * configure ip address on container iface and bring it up.
// * configure ip address on host iface and bring it up.
// * attach snat bpf program to host-side veth peer (tc ingress)
func (c *CNI) ExecAdd(args *skel.CmdArgs) (err error) {
	ctx := context.Background()

	var conf Conf
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("parse network config: %v", err)
	}

	if conf.IPAM == (types.IPAM{}) {
		return ErrIPAMConfigNotFound
	}

	if conf.HostIface == "" {
		return ErrHostIfaceNotFound
	}

	if err := c.handler.AttachDNATBPF(conf.HostIface); err != nil {
		return fmt.Errorf("failed to attach dnat bpf to %s: %w", conf.HostIface, err)
	}

	defer func() {
		if err != nil {
			if err := c.handler.DeallocIPs(conf.IPAM.Type, args.StdinData); err != nil {
				log.Printf("could not deallocate ips after CNI ADD failure: %v\n", err)
			}
		}
	}()

	ips, err := c.handler.AllocIPs(conf.IPAM.Type, args.StdinData)
	if err != nil {
		return fmt.Errorf("alloc ips: %w", err)
	}

	hostVethName, podVethName, err := c.handler.CreateAndConfigureVethPair(args.Netns, ips)
	if err != nil {
		return fmt.Errorf("configure veth pair: %w", err)
	}

	if err := c.handler.AttachHostVethBPF(hostVethName); err != nil {
		return fmt.Errorf("attach snat: %w", err)
	}

	if err := c.handler.ConfigureSNAT(conf.HostIface); err != nil {
		return fmt.Errorf("configure snat: %w", err)
	}

	if err := c.handler.AddDefaultRoute(args.Netns); err != nil {
		return fmt.Errorf("add default route: %w", err)
	}

	proxyConn, err := grpc.NewClient(
		conf.PlatformdListenSock,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to create proxy service grpc client: %w", err)
	}

	id, err := getWorkloadID(ctx, conf.CRIListenSock, args.ContainerID)
	if err != nil {
		return fmt.Errorf("get workload id: %w", err)
	}

	client := proxyv1alpha1.NewProxyServiceClient(proxyConn)
	if _, err := client.CreateListeners(ctx, &proxyv1alpha1.CreateListenersRequest{
		WorkloadID: id,
		Ip:         ips[0].Address.IP.String(),
	}); err != nil {
		return fmt.Errorf("create proxy listeners: %w", err)
	}

	result := &current.Result{
		CNIVersion: supportedCNIVersion,
		Interfaces: []*current.Interface{
			{
				Name:    podVethName,
				Sandbox: args.Netns,
			},
		},
	}

	if err := result.PrintTo(os.Stdout); err != nil {
		return fmt.Errorf("print result: %w", err)
	}

	return nil
}

func getWorkloadID(ctx context.Context, criListenSock, ctrID string) (string, error) {
	criConn, err := grpc.NewClient(criListenSock, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", fmt.Errorf("failed to create cri grpc client: %w", err)
	}
	c := runtimev1.NewRuntimeServiceClient(criConn)
	ctrResp, err := c.ListContainers(ctx, &runtimev1.ListContainersRequest{
		Filter: &runtimev1.ContainerFilter{
			Id: ctrID,
		},
	})
	if err != nil {
		return "", fmt.Errorf("list containers: %w", err)
	}

	if len(ctrResp.Containers) == 0 {
		return "", errors.New("container not found")
	}

	podResp, err := c.ListPodSandbox(ctx, &runtimev1.ListPodSandboxRequest{
		Filter: &runtimev1.PodSandboxFilter{
			Id: ctrResp.Containers[0].PodSandboxId,
		},
	})
	if err != nil {
		return "", fmt.Errorf("list pods: %w", err)
	}

	if len(podResp.Items) == 0 {
		return "", errors.New("pod not found")
	}

	id, ok := podResp.Items[0].Labels[workload.LabelID]
	if !ok {
		return "", errors.New("workload id label not found")
	}
	return id, nil
}

func (c *CNI) ExecDel(args *skel.CmdArgs) error {
	log.Println("del")
	// TODO: remove veth pairs
	return nil
}
