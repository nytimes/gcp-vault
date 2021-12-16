package kitexample

import (
	"context"
	"net/http"
	"os"

	"github.com/NYTimes/gizmo/server/kit"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/kelseyhightower/envconfig"
	gcpvault "github.com/nytimes/gcp-vault"
	"github.com/nytimes/gcp-vault/examples/nyt"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func NewService() (*service, error) {
	var cfg gcpvault.Config
	envconfig.Process("", &cfg)
	secrets, err := gcpvault.GetSecrets(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	keyI, ok := secrets["APIKey"]
	if !ok {
		return nil, errors.New("APIKey secret is not found")
	}
	key, ok := keyI.(string)
	if !ok {
		return nil, errors.New("APIKey secret is not a string")
	}

	return &service{
		client: nyt.NewClient(os.Getenv("NYT_HOST"), key),
	}, nil
}

type service struct {
	client nyt.Client
}

func (s *service) HTTPEndpoints() map[string]map[string]kit.HTTPEndpoint {
	return map[string]map[string]kit.HTTPEndpoint{
		"/svc/example/v1/top-stories": {
			"GET": {
				Endpoint: s.getTopScienceStories,
			},
		},
	}
}

// to satisfy the kit.Service interface
func (s *service) HTTPMiddleware(h http.Handler) http.Handler {
	return h
}

// to satisfy the kit.Service interface
func (s *service) Middleware(e endpoint.Endpoint) endpoint.Endpoint {
	return e
}

// to satisfy the kit.Service interface
func (s *service) HTTPOptions() []httptransport.ServerOption {
	return nil
}

// to satisfy the kit.Service interface
func (s *service) HTTPRouterOptions() []kit.RouterOption {
	return nil
}

// to satisfy the kit.Service interface
func (s *service) RPCMiddleware() grpc.UnaryServerInterceptor {
	return nil
}

// to satisfy the kit.Service interface
func (s *service) RPCServiceDesc() *grpc.ServiceDesc {
	return nil
}

// to satisfy the kit.Service interface
func (s *service) RPCOptions() []grpc.ServerOption {
	return nil
}
