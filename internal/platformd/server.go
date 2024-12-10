package platformd

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path"

	"github.com/spacechunks/platform/internal/datapath"

	"github.com/spacechunks/platform/internal/platformd/proxy/xds"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/hashicorp/go-multierror"
	proxyv1alpha1 "github.com/spacechunks/platform/api/platformd/proxy/v1alpha1"
	"github.com/spacechunks/platform/internal/platformd/proxy"
	"github.com/spacechunks/platform/internal/platformd/workload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	runtimev1 "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type Server struct {
	logger *slog.Logger
}

func NewServer(logger *slog.Logger) *Server {
	return &Server{
		logger: logger,
	}
}

func (s *Server) Run(ctx context.Context, cfg Config) error {
	criConn, err := grpc.NewClient("unix://"+cfg.CRIListenSock, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to create cri grpc client: %w", err)
	}

	var (
		xdsCfg = cache.NewSnapshotCache(true, cache.IDHash{}, nil)
		wlSvc  = workload.NewService(
			s.logger,
			runtimev1.NewRuntimeServiceClient(criConn),
			runtimev1.NewImageServiceClient(criConn),
		)
		proxySvc = proxy.NewService(xds.NewMap(xdsCfg))

		mgmtServer  = grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
		proxyServer = proxy.NewServer(proxySvc)
	)

	proxyv1alpha1.RegisterProxyServiceServer(mgmtServer, proxyServer)
	xds.CreateAndRegisterServer(ctx, mgmtServer, xdsCfg)

	bpf, err := datapath.LoadBPF()
	if err != nil {
		return fmt.Errorf("failed to load bpf: %w", err)
	}

	if err := proxySvc.ApplyOriginalDstCluster(ctx); err != nil {
		return fmt.Errorf("apply original dst cluster: %w", err)
	}

	if err := bpf.AttachAndPinGetsockopt(cfg.GetsockoptCGroup); err != nil {
		return fmt.Errorf("attach getsockopt: %w", err)
	}

	// before we start our grpc services make sure our system workloads are running
	labels := workload.SystemWorkloadLabels("envoy")
	if err := wlSvc.EnsureWorkload(ctx, workload.CreateOptions{
		Name:             "envoy",
		Image:            cfg.EnvoyImage,
		Namespace:        "system",
		Labels:           labels,
		NetworkNamespace: workload.NetworkNamespaceHost,
	}, labels); err != nil {
		return fmt.Errorf("ensure envoy: %w", err)
	}

	if err := os.MkdirAll(path.Dir(cfg.ProxyServiceListenSock), os.ModePerm); err != nil {
		return fmt.Errorf("create mgmt listen sock dir: %w", err)
	}

	unixSock, err := net.Listen("unix", cfg.ProxyServiceListenSock)
	if err != nil {
		return fmt.Errorf("failed to listen on unix socket: %v", err)
	}

	ctx, cancel := context.WithCancel(ctx)

	var g multierror.Group
	g.Go(func() error {
		if err := mgmtServer.Serve(unixSock); err != nil {
			cancel()
			return fmt.Errorf("failed to serve mgmt server: %w", err)
		}
		return nil
	})

	<-ctx.Done()

	// add stop related code below

	mgmtServer.GracefulStop()
	g.Go(func() error {
		if err := criConn.Close(); err != nil {
			return fmt.Errorf("cri conn close: %w", err)
		}
		return nil
	})

	return g.Wait().ErrorOrNil()
}
