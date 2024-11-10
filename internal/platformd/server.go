package platformd

import (
	"context"
	"fmt"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/hashicorp/go-multierror"
	proxyv1alpha1 "github.com/spacechunks/platform/api/platformd/proxy/v1alpha1"
	"github.com/spacechunks/platform/internal/platformd/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	runtimev1 "k8s.io/cri-api/pkg/apis/runtime/v1"
	"log/slog"
	"net"
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
	unixSock, err := net.Listen("unix", cfg.ProxyServiceListenSock)
	if err != nil {
		return fmt.Errorf("failed to listen on unix socket: %v", err)
	}

	criConn, err := grpc.NewClient("unix://"+cfg.CRIListenSock, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to create cri grpc client: %w", err)
	}

	var (
		mgmtServer  = grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
		proxyServer = proxy.NewServer()
		xdsCfg      = cache.NewSnapshotCache(true, cache.IDHash{}, nil)
		criClient   = runtimev1.NewRuntimeServiceClient(criConn)
	)

	proxyv1alpha1.RegisterProxyServiceServer(mgmtServer, proxyServer)
	proxy.CreateAndRegisterXDSServer(ctx, mgmtServer, xdsCfg)

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

	mgmtServer.GracefulStop()
	g.Go(func() error {
		if err := criConn.Close(); err != nil {
			return fmt.Errorf("cri conn close: %w", err)
		}
		return nil
	})

	return g.Wait().ErrorOrNil()
}
