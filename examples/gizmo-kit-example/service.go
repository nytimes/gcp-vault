package gizmoexample

import (
	"context"
	"net/http"
	"sync"

	"google.golang.org/grpc"

	gcpvault "github.com/NYTimes/gcp-vault"
	"github.com/NYTimes/gizmo/server/kit"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/kelseyhightower/envconfig"
)

func NewService() (*service, error) {
	var cfg gcpvault.Config
	envconfig.Process("", &cfg)
	svc := &service{vaultConfig: cfg}

	// init the secrets as our server starts up so we can fail fast if something goes
	// wrong
	err := svc.initSecrets(context.Background())

	return svc, err
}

type service struct {
	vaultConfig gcpvault.Config
	secretsOnce sync.Once

	mySecret string
}

func (s *service) HTTPMiddleware(h http.Handler) http.Handler {
	return h
}

func (s *service) Middleware(e endpoint.Endpoint) endpoint.Endpoint {
	return e
}

func (s *service) HTTPEndpoints() map[string]map[string]kit.HTTPEndpoint {
	return map[string]map[string]kit.HTTPEndpoint{
		"/svc/example/v1/my-secret": {
			"GET": {
				Endpoint: s.getMySecret,
			},
		},
	}
}

func (s *service) HTTPOptions() []httptransport.ServerOption {
	return nil
}

func (s *service) HTTPRouterOptions() []kit.RouterOption {
	return nil
}

func (s *service) RPCMiddleware() grpc.UnaryServerInterceptor {
	return nil
}

func (s *service) RPCServiceDesc() *grpc.ServiceDesc {
	return nil
}

func (s *service) RPCOptions() []grpc.ServerOption {
	return nil
}
