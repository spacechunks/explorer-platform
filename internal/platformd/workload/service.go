package workload

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/google/uuid"
	runtimev1 "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type CreateOptions struct {
	Name      string
	Image     string
	Namespace string
	Hostname  string
	Labels    map[string]string
}

const podLogDir = "/var/log/platformd/pods"

type Service interface {
	CreateWorkload(ctx context.Context, opts CreateOptions) error
	EnsureWorkload(ctx context.Context, opts CreateOptions, labelSelector map[string]string) error
}

type criService struct {
	logger    *slog.Logger
	rtClient  runtimev1.RuntimeServiceClient
	imgClient runtimev1.ImageServiceClient
}

func NewService(logger *slog.Logger, rtClient runtimev1.RuntimeServiceClient, imgClient runtimev1.ImageServiceClient) Service {
	return &criService{
		logger:    logger,
		rtClient:  rtClient,
		imgClient: imgClient,
	}
}

// EnsurePod ensures that a pod is created if not present.
// if ListPodSandbox returns 0 items, a pod with the passed configuration is created.
// Currently this function is designed for a single item returned by the label selector.
// If multiple items are returned the first one will be picked.
// TODO: what do we do if the pod found is in NOT_READY state
func (s *criService) EnsureWorkload(ctx context.Context, opts CreateOptions, labelSelector map[string]string) error {
	resp, err := s.rtClient.ListPodSandbox(ctx, &runtimev1.ListPodSandboxRequest{
		Filter: &runtimev1.PodSandboxFilter{
			LabelSelector: labelSelector,
		},
	})
	if err != nil {
		return fmt.Errorf("list pod sandbox: %w", err)
	}

	if len(resp.Items) > 0 {
		return nil
	}

	s.logger.InfoContext(ctx,
		"no matching workload found, creating pod",
		"pod_name", opts.Name,
		"namespace", opts.Namespace,
		"label_selector", labelSelector,
	)

	if err := s.CreateWorkload(ctx, opts); err != nil {
		return fmt.Errorf("create pod: %w", err)
	}
	return nil
}

func (s *criService) CreateWorkload(ctx context.Context, opts CreateOptions) error {
	imgResp, err := s.imgClient.ListImages(ctx, &runtimev1.ListImagesRequest{})
	if err != nil {
		return fmt.Errorf("list images: %w", err)
	}

	var img *runtimev1.Image
	for _, tmp := range imgResp.Images {
		if slices.Contains(tmp.RepoTags, opts.Image) {
			img = tmp
			break
		}
	}

	if img == nil {
		return fmt.Errorf("image not found")
	}

	sboxCfg := &runtimev1.PodSandboxConfig{
		Metadata: &runtimev1.PodSandboxMetadata{
			Name:      opts.Name,
			Uid:       uuid.New().String(),
			Namespace: opts.Namespace,
		},
		Hostname:     opts.Hostname,
		LogDirectory: podLogDir,
		Labels:       opts.Labels,
	}

	sboxResp, err := s.rtClient.RunPodSandbox(ctx, &runtimev1.RunPodSandboxRequest{
		Config:         sboxCfg,
		RuntimeHandler: "",
	})
	if err != nil {
		return fmt.Errorf("create pod: %w", err)
	}

	_, err = s.rtClient.CreateContainer(ctx, &runtimev1.CreateContainerRequest{
		PodSandboxId: sboxResp.PodSandboxId,
		Config: &runtimev1.ContainerConfig{
			Metadata: &runtimev1.ContainerMetadata{
				Name:    opts.Name,
				Attempt: 0,
			},
			Image: &runtimev1.ImageSpec{
				Image: img.Id,
			},
			Labels:  opts.Labels,
			LogPath: fmt.Sprintf("%s_%s", opts.Namespace, opts.Name),
		},
		SandboxConfig: sboxCfg,
	})
	if err != nil {
		return fmt.Errorf("create container: %w", err)
	}

	return nil
}
