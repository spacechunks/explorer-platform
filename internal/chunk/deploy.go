package chunk

import (
	"context"
	"fmt"
	"github.com/chunks76k/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	"log"
)

func DeployAll(ctx context.Context, regHost string, conf Config, conn *pgx.Conn) error {
	// TODO:
	// collect how many servers running this mode are live
	// * do this by quering all configured kubernetes endpoints
	//   for running pods with certain labels or smth
	// check if we are running into some server limit, if yes return
	// error and do nothing
	// * hardcode limit of 10 for POC, dont want to handle
	//   user data atm
	dao := db.New(conn)
	deploys, err := dao.ListVariantDeploys(ctx, pgtype.Text{
		String: conf.Name,
	})
	if err != nil {
		return fmt.Errorf("list variant deploys %w", err)
	}
	// reconcile variant deployment
	for _, deploy := range deploys {
		log.Printf(
			"fetch current deployment state mode=%s variant=%s",
			deploy.Mode,
			deploy.Variant,
		)
		v, ok := variant(conf.Variants, deploy.Variant.String)
		if !ok {
			return fmt.Errorf("variant %s not found", deploy.Variant)
		}
		for i := 0; i < v.Replicas; i++ {
			i := i
			go func() {
				if err := reconcileVariantReplicas(ctx, deploy, v, regHost, i); err != nil {
					log.Printf(
						"reconcile mode=%s variant=%s replica=%d err=%v",
						deploy.Mode,
						deploy.Variant,
						i,
						err,
					)
				}
			}()
		}
	}
	return nil
}

func reconcileVariantReplicas(
	ctx context.Context,
	deploy db.VariantDeployment,
	v Variant,
	regHost string,
	replica int,
) error {
	rConf := &rest.Config{
		Host:        deploy.ClusterUrl.String,
		BearerToken: deploy.ClusterToken.String,
	}
	rc, err := rest.RESTClientFor(rConf)
	if err != nil {
		return fmt.Errorf("rest client: %w", err)
	}
	kube := kubernetes.New(rc)
	var (
		// TODO: need to put project name here as well to have it unique
		// flash-easy-01
		name = fmt.Sprintf("%s-%s-%d", deploy.Mode, deploy.Variant, replica)
		// we need this one in order to know if we need to deploy
		// a node port service first + create instead of update an
		// existing deployment
		deployed = true
	)
	handle, err := kube.AppsV1().Deployments("chunks").Get(ctx, name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		deployed = false
	}
	if err != nil {
		return fmt.Errorf("get deploy: %w", err)
	}
	imgRef := fmt.Sprintf("%s/%s-%s", regHost, deploy.Mode.String, v.Name)
	if !deployed {
		svc := nodePortSvc(name, "chunks")
		if _, err := kube.CoreV1().
			Services("chunks").
			Create(ctx, svc, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("create service: %w", err)
		}
		handle = deployment(name, imgRef, "chunks")
		if _, err := kube.AppsV1().
			Deployments("chunks").
			Create(ctx, handle, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("create deploy: %w", err)
		}
		return nil
	}
	handle.Spec.Template.Spec.Containers[0].Image = imgRef
	if _, err := kube.AppsV1().
		Deployments("chunks").
		Update(ctx, handle, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("update deploy: %w", err)
	}
	return nil
}

func variant(s []Variant, name string) (Variant, bool) {
	for _, v := range s {
		if v.Name == name {
			return v, true
		}
	}
	return Variant{}, false
}

func nodePortSvc(name string, ns string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: map[string]string{}, // TODO: external dns
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Protocol: corev1.ProtocolTCP,
					Port:     25565,
				},
			},
			Selector: map[string]string{
				"name": name,
			},
			Type: corev1.ServiceTypeNodePort,
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
					Labels: map[string]string{
						"name": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: imgRef,
						},
					},
				},
			},
		},
	}
}
