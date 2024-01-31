package chunk

import (
	"context"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spacechunks/chunks/internal/db"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	"log"
	"sync"
	"time"
)

const ns = "chunks-system"

func DeployAll(ctx context.Context, imgRepo string, meta Meta, conf Config /*conn *pgx.Conn*/) error {
	// TODO: run multiple versions at the same time?
	// TODO:
	// collect how many servers running this mode are live
	// * do this by quering all configured kubernetes endpoints
	//   for running pods with certain labels or smth
	// check if we are running into some server limit, if yes return
	// error and do nothing
	// * hardcode limit of 10 for POC, dont want to handle
	//   user data atm
	// project name is globally unique

	var deploys []db.VariantDeployment
	var err error = nil
	/*
		dao := db.New(conn)
		deploys, err := dao.ListVariantDeploys(ctx, pgtype.Text{
			String: meta.ChunkID,
		})
		if err != nil {
			return fmt.Errorf("list variant deploys %w", err)
		}*/

	// we have currently no deployments running
	// can happen if we deploy the chunk for the
	// first time
	if len(deploys) == 0 {
		// TODO: choose in which cluster variant will run
		// TODO: write to variant_deployment table
		log.Printf("no variant deployment found\n")
		for _, v := range conf.Variants {
			// TODO: need variant deployment as domain object
			deploys = append(deploys, db.VariantDeployment{
				Mode: pgtype.Text{
					String: meta.ChunkID,
				},
				Variant: pgtype.Text{
					String: v.ID,
				},
				ClusterUrl: pgtype.Text{
					String: "https://4513aec9-3a16-4b74-af2d-c99c3454af91.vultr-k8s.com:6443",
				},
				ClusterToken: pgtype.Text{
					String: "eyJhbGciOiJSUzI1NiIsImtpZCI6IjBOLTdtczRRUmpTNURwTWhpdUdUcjA2QTdkZGhZb3RMMFByZEdvTUFQclUifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJjaHVua3Mtc3lzdGVtIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImNodW5rZXItY2x1c3Rlci10b2tlbiIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJjaHVua2VyIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQudWlkIjoiODgyMmQ2MWEtMWIyZi00NWQwLTgyMzAtN2Q1ZGFmNzgzMDJiIiwic3ViIjoic3lzdGVtOnNlcnZpY2VhY2NvdW50OmNodW5rcy1zeXN0ZW06Y2h1bmtlciJ9.NM7XlG77ctLGUjYsH5G_gZ8z4Sp9o4_s_Zg56XIoGrQ0JKeK3ZQRkN1Vd9iG45IjZHhy6VhTfypJBE_geAGg4RqD7pjFN8zV7YiFSsO8B88gfQKdlQ_ntZ_pilayj6vYaIeQ8TtCj27edrVaMdGqQrVpcQTeVeU5IEG7aNtWc0F-dC8iU8NbbGqu-RRLeDexlxw_x4kP2V7ccRukEF98cEYp6u2SfXJLlZdsVt-MmCDiXHCNvMv1a9SJiGbemyDNc0hZIY9fwMPEvs23GBpiToWyD5T55UxgCQZjs-tFg1tiDa35Uf-tcz0mUAoYG6TyE0QRpaDRoIPq5NrwY3MeRA",
				},
			})
		}
	}
	// reconcile variant deployment
	for _, deploy := range deploys {
		log.Printf(
			"fetching current deployment state mode=%s variant=%s",
			deploy.Mode.String,
			deploy.Variant.String,
		)
		v, ok := variant(conf.Variants, deploy.Variant.String)
		if !ok {
			return fmt.Errorf("variant %s not found", deploy.Variant.String)
		}
		wg := sync.WaitGroup{}
		wg.Add(v.Replicas)
		for i := 0; i < v.Replicas; i++ {
			i := i
			go func() {
				defer wg.Done()
				// TODO: better reconcile:
				// * redeploy deleted resources
				imgRef := fmt.Sprintf("%s:%s-%s", imgRepo, meta.ChunkVersion, v.ID)
				err = reconcileVariantReplica(ctx, deploy, imgRef, i)
				if err != nil {
					log.Printf(
						"reconcile mode=%s variant=%s replica=%d err=%v",
						deploy.Mode.String,
						deploy.Variant.String,
						i,
						err,
					)
				}
			}()
		}
		wg.Wait()
		if err != nil {
			return fmt.Errorf("reconcile: %w", err)
		}
	}
	return nil
}

// TODO: rework reconcile
func reconcileVariantReplica(
	ctx context.Context,
	deploy db.VariantDeployment,
	imgRef string,
	replica int,
) error {
	kube, err := kubernetes.NewForConfig(&rest.Config{
		Host:        deploy.ClusterUrl.String,
		BearerToken: deploy.ClusterToken.String,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	})
	if err != nil {
		return fmt.Errorf("clientset: %w", err)
	}
	var (
		// freggy-flash-easy-01
		name = fmt.Sprintf("%s-%s-%d", deploy.Mode.String, deploy.Variant.String, replica)
		// we need this one in order to know if we need to deploy
		// a node port service first + create instead of update an
		// existing deployment
		deployed = true
	)
	handle, err := kube.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		deployed = false
	}
	if err != nil && deployed {
		return fmt.Errorf("get deploy: %w", err)
	}
	if !deployed {
		log.Printf("initial deploy name=%s", name)
		handle = deployment(name, imgRef, ns)
		if _, err := kube.AppsV1().
			Deployments(ns).
			Create(ctx, handle, metav1.CreateOptions{}); ignoreAlreadyExists(err) != nil {
			return fmt.Errorf("create deploy: %w", err)
		}
		svc := nodePortSvc(name, ns)
		svc, err = kube.CoreV1().
			Services(ns).
			Create(ctx, svc, metav1.CreateOptions{})
		if ignoreAlreadyExists(err) != nil {
			return fmt.Errorf("create svc: %w", err)
		}
		// retrieve service again, to retrieve node port from status
		svc, err = kube.CoreV1().Services(ns).Get(ctx, svc.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("get svc: %w", err)
		}
		// 1-freggy-flash-easy.chunks.76k.io
		domain := fmt.Sprintf("%d-%s-%s.chunks.76k.io", replica, deploy.Mode.String, deploy.Variant.String)
		if err := configureDNS(
			svc.Spec.Ports[0].NodePort,
			"157.90.167.132", // TODO: cluster LB IP
			domain,
			"76k.io",
			"43994d4a79a7e5c28dc5476589eb9da0f7ad9",
			"riegeryannic@gmail.com",
		); err != nil {
			return fmt.Errorf("dns: %w", err)
		}
		log.Printf("deploy complete name=%s", name)
		return nil
	}

	log.Printf("updating exsiting variant deployment name=%s img=%s", name, imgRef)
	handle.Spec.Template.Spec.Containers[0].Image = imgRef

	// note that  we do not need to update dns here, because
	// we will point to a static IP, so nothing regarding
	// DNS needs to be changed. records must be deleted
	// when the variant deployment is deleted.

	// rollout restart deployment
	handle.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
	if _, err := kube.AppsV1().
		Deployments(ns).
		Update(ctx, handle, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("update deploy: %w", err)
	}
	return nil
}

func variant(s []Variant, id string) (Variant, bool) {
	for _, v := range s {
		if v.ID == id {
			return v, true
		}
	}
	return Variant{}, false
}

func nodePortSvc(name string, ns string) *corev1.Service {
	return &corev1.Service{
		// cannot use external dns to create SRV record
		// see https://github.com/kubernetes-sigs/external-dns/pull/1890
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Protocol: corev1.ProtocolTCP,
					Port:     25565,
				},
			},
			Selector: map[string]string{
				"name": name,
			},
		},
	}
}

func deployment(name, imgRef, ns string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: ns,
					Labels: map[string]string{
						"name": name,
					},
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "reg1-auth",
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: imgRef,
						},
					},
				},
			},
		},
	}
}

func configureDNS(port int32, ip, domain, zoneName, key, mail string) error {
	ctx := context.Background()
	api, err := cloudflare.New(key, mail)
	if err != nil {
		return fmt.Errorf("cf client: %w", err)
	}
	zoneID, err := api.ZoneIDByName(zoneName)
	if err != nil {
		return fmt.Errorf("get zone: %w", err)
	}
	id := cloudflare.ZoneIdentifier(zoneID)
	records, _, err := api.ListDNSRecords(ctx, id, cloudflare.ListDNSRecordsParams{})
	if !recordExists(records, domain) {
		log.Printf("creating A record domain=%s ip=%s\n", domain, ip)
		if err := createDNSRecord(ctx, api, zoneID, domain, "A", ip, -1); err != nil {
			return fmt.Errorf("create a record: %w", err)
		}
	}
	log.Printf("updating A record domain=%s ip=%s\n", domain, ip)
	if _, err := api.UpdateDNSRecord(ctx, id, cloudflare.UpdateDNSRecordParams{
		ID:      zoneID,
		Type:    "A",
		Name:    domain,
		Content: ip,
		TTL:     1, // automatic
		Proxied: pointer.Bool(false),
	}); err != nil {
		return fmt.Errorf("update A record: %w", err)
	}
	if !recordExists(records, domain) {
		log.Printf("creating SRV record domain=%s port=%d\n", domain, port)
		if err := createDNSRecord(ctx, api, zoneID, domain, "SRV", "", port); err != nil {
			return fmt.Errorf("create SRV record: %w", err)
		}
	}
	log.Printf("updating SRV record domain=%s port=%d\n", domain, port)
	if _, err := api.UpdateDNSRecord(ctx, id, cloudflare.UpdateDNSRecordParams{
		ID:   zoneID,
		Type: "SRV",
		Data: map[string]any{
			"name":     domain,
			"port":     port,
			"priority": 1,
			"proto":    "_tcp",
			"service":  "_minecraft",
			"weight":   0,
			"target":   domain,
		},
		TTL:     1, // automatic
		Proxied: pointer.Bool(false),
	}); err != nil {
		return fmt.Errorf("update SRV record: %w", err)
	}
	return nil
}

func createDNSRecord(
	ctx context.Context,
	api *cloudflare.API,
	zoneID string,
	domain string,
	typeName string,
	ip string,
	port int32,
) error {
	if typeName == "A" {
		if _, err := api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.CreateDNSRecordParams{
			Type:      "A",
			Name:      domain,
			Content:   ip,
			TTL:       1, // automatic
			Proxiable: false,
		}); err != nil {
			return err
		}
		return nil
	}
	if typeName == "SRV" {
		if _, err := api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.CreateDNSRecordParams{
			Type: "SRV",
			Data: map[string]any{
				"name":     domain,
				"port":     port,
				"priority": 1,
				"proto":    "_tcp",
				"service":  "_minecraft",
				"weight":   0,
				"target":   domain,
			},
			TTL:       1, // automatic
			Proxiable: false,
		}); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func recordExists(records []cloudflare.DNSRecord, domain string) bool {
	for _, r := range records {
		if r.Type == "A" && r.Name == domain {
			return true
		}
		if r.Type == "SRV" && r.Data.(map[string]any)["name"] == domain {
			return true
		}
	}
	return false
}

func ignoreAlreadyExists(err error) error {
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
