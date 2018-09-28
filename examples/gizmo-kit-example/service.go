package gizmoexample

import (
	"context"
	"net/http"
	"sync"

	"google.golang.org/grpc"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/kelseyhightower/envconfig"
	gcpvault "github.com/nytimes/gcp-vault"
	"github.com/nytimes/gizmo/server/kit"
)

func NewService() (*service, error) {
	var cfg gcpvault.Config
	envconfig.Process("", &cfg)
	svc := &service{vaultConfig: cfg}
	// init the secrets on server startup
	err := svc.initSecrets(context.Background())
	return svc, err
}

type service struct {
	vaultConfig gcpvault.Config
	secretsOnce sync.Once

	mySecret string
}

func (s *service) HTTPMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ensure secrets are fully loaded before any request comes through
		err := s.initSecrets(context.Background())
		if err != nil {
			kit.LogErrorMsg(r.Context(), err, "unable to fetch secrets")
		}

		// call next layer down
		h.ServeHTTP(w, r)
	})
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
