package marvinexample

import (
	"context"
	"net/http"
	"sync"

	"google.golang.org/appengine/log"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/kelseyhightower/envconfig"
	gcpvault "github.com/nytimes/gcp-vault"
	"github.com/nytimes/marvin"
)

func NewService() *service {
	// configure from the environment/app.yaml
	var cfg gcpvault.Config
	envconfig.Process("", &cfg)
	return &service{vaultConfig: cfg}
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
	return func(ctx context.Context, r interface{}) (interface{}, error) {
		// attempt to fetch our secrets 1 time only when the first request comes in.
		err := s.initSecrets(ctx)
		if err != nil {
			log.Errorf(ctx, "unable to init secrets: %s", err)
			return nil, marvin.NewJSONStatusResponse("server error",
				http.StatusInternalServerError)
		}

		// call the actual endpoint
		return e(ctx, r)
	}
}

func (s *service) JSONEndpoints() map[string]map[string]marvin.HTTPEndpoint {
	return map[string]map[string]marvin.HTTPEndpoint{
		"/svc/example/v1/my-secret": {
			"GET": {
				Endpoint: s.getMySecret,
			},
		},
		"/_ah/warmup": {
			"GET": {
				Endpoint: func(ctx context.Context, r interface{}) (interface{}, error) {
					return "ok", nil
				},
			},
		},
	}
}

// to satisfy the marvin.Service interface
func (s *service) Options() []httptransport.ServerOption {
	return nil
}

// to satisfy the marvin.Service interface
func (s *service) RouterOptions() []marvin.RouterOption {
	return nil
}
