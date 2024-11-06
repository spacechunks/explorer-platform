package platformd

import (
	"fmt"
	proxyv1alpha1 "github.com/spacechunks/platform/api/platformd/proxy/v1alpha1"
	"github.com/spacechunks/platform/internal/platformd/proxy"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type Server struct {
	logger *slog.Logger
}

func NewServer(logger *slog.Logger) *Server {
	return &Server{}
}

func (s *Server) Run(cfg Config) error {
	unix, err := net.Listen("unix", cfg.ProxyServiceListenSock)
	if err != nil {
		return fmt.Errorf("failed to listen on unix socket: %v", err)
	}

	var (
		mgmtServer  = grpc.NewServer()
		proxyServer = proxy.NewServer()
	)

	proxyv1alpha1.RegisterProxyServiceServer(mgmtServer, proxyServer)

	if err := mgmtServer.Serve(unix); err != nil {
		return fmt.Errorf("failed to serve mgmt server: %w", err)
	}

	return nil
}
