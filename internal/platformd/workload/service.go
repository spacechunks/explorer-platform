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
	logger := s.logger.With("pod_name", opts.Name, "namespace", opts.Namespace)

	if err := s.pullImageIfNotPresent(ctx, logger, opts.Image); err != nil {
		return fmt.Errorf("pull image if not present: %w", err)
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
		Config: sboxCfg,
	})
	if err != nil {
		return fmt.Errorf("create pod: %w", err)
	}

	logger = logger.With("pod_id", sboxResp.PodSandboxId)
	logger.InfoContext(ctx, "started pod sandbox")

	ctrResp, err := s.rtClient.CreateContainer(ctx, &runtimev1.CreateContainerRequest{
		PodSandboxId: sboxResp.PodSandboxId,
		Config: &runtimev1.ContainerConfig{
			Metadata: &runtimev1.ContainerMetadata{
				Name:    opts.Name,
				Attempt: 0,
			},
			Image: &runtimev1.ImageSpec{
				Image: opts.Image,
			},
			Labels:  opts.Labels,
			LogPath: fmt.Sprintf("%s_%s", opts.Namespace, opts.Name),
		},
		SandboxConfig: sboxCfg,
	})
	if err != nil {
		return fmt.Errorf("create container: %w", err)
	}

	if _, err := s.rtClient.StartContainer(ctx, &runtimev1.StartContainerRequest{
		ContainerId: ctrResp.ContainerId,
	}); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	logger.InfoContext(ctx, "started container", "container_id", ctrResp.ContainerId)
	return nil
}

// pullImageIfNotPresent first calls ListImages then checks if the image is contained in the response.
// if this is not the case PullImage is being called. this function does not access the services logger,
// and instead uses a passed one, to preserve arguments which provide additional context to the image pull.
func (s *criService) pullImageIfNotPresent(ctx context.Context, logger *slog.Logger, imageURL string) error {
	listResp, err := s.imgClient.ListImages(ctx, &runtimev1.ListImagesRequest{})
	if err != nil {
		return fmt.Errorf("list images: %w", err)
	}

	var img *runtimev1.Image
	for _, tmp := range listResp.Images {
		if slices.Contains(tmp.RepoTags, imageURL) {
			img = tmp
			break
		}
	}

	if img != nil {
		return nil
	}

	logger = logger.With("image", imageURL)
	logger.InfoContext(ctx, "pulling image")

	if _, err := s.imgClient.PullImage(ctx, &runtimev1.PullImageRequest{
		Image: &runtimev1.ImageSpec{
			Image: imageURL,
		},
	}); err != nil {
		return fmt.Errorf("pull image: %w", err)
	}

	logger.InfoContext(ctx, "image pulled")
	return nil
}
