package marvinexample

import (
	"context"
	"net/http"
	"os"
	"sync"

	"google.golang.org/appengine/log"

	"github.com/NYTimes/marvin"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/kelseyhightower/envconfig"
	gcpvault "github.com/nytimes/gcp-vault"
	"github.com/pkg/errors"
)

func NewService() *service {
	var cfg gcpvault.Config
	envconfig.Process("", &cfg)
	return &service{
		nytHost: os.Getenv("NYT_HOST"),
		vcfg:    cfg,
	}
}

type service struct {
	nytHost     string
	vcfg        gcpvault.Config
	secretsOnce sync.Once
	apiKey      string
}

func (s *service) HTTPMiddleware(h http.Handler) http.Handler {
	return h
}

func (s *service) Middleware(e endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (interface{}, error) {

		// attempt to fetch our secrets 1 time only when the first request comes in.
		// GAE standard only allows network access within the scope of an inbound request
		// so we must use our middleware to ensure the first request (hopefully a warmup
		// request) fetches the secrets before any other action happens on the service.
		// only attempt to call vault 1 time at startup
		var err error
		s.secretsOnce.Do(func() {
			s.apiKey, err = s.getKey(ctx)
		})
		if err != nil || s.apiKey == "" {
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
		"/svc/example/v1/top-stories": {
			"GET": {
				Endpoint: s.getTopStories,
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

func (s *service) getKey(ctx context.Context) (string, error) {
	secrets, err := gcpvault.GetSecrets(ctx, s.vcfg)
	if err != nil {
		return "", err
	}
	keyI, ok := secrets["APIKey"]
	if !ok {
		return "", errors.New("APIKey secret is not found")
	}
	key, ok := keyI.(string)
	if !ok {
		return "", errors.New("APIKey secret is not a string")
	}

	return key, nil
}
