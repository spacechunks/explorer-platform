package proxy

import (
	"context"
	proxyv1alpha1 "github.com/spacechunks/platform/api/platformd/proxy/v1alpha1"
)

type Server struct {
	proxyv1alpha1.UnimplementedProxyServiceServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) CreateListener(
	ctx context.Context, req *proxyv1alpha1.CreateListenerRequest,
) (*proxyv1alpha1.CreateListenerResponse, error) {
	return &proxyv1alpha1.CreateListenerResponse{}, nil
}
