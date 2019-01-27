package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	gcpvault "github.com/NYTimes/gcp-vault"
	"github.com/NYTimes/gcp-vault/examples/nyt"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/kelseyhightower/envconfig"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

func main() {
	// register secret-fetching middleware on all endpoints as warm up requests are NOT
	// guaranteed.
	http.HandleFunc("/_ah/warmup", secretsMiddleware(warmUpHandler))
	http.HandleFunc("/my-handler", secretsMiddleware(myHandler))
	appengine.Main()
}

var clientKey string

func initKey(ctx context.Context) error {
	var cfg gcpvault.Config
	envconfig.Process("", &cfg)
	secrets, err := gcpvault.GetSecrets(ctx, cfg)
	if err != nil {
		return err
	}
	keyI, ok := secrets["APIKey"]
	if !ok {
		return errors.New("APIKey secret is not found")
	}
	clientKey, ok = keyI.(string)
	if !ok {
		return errors.New("APIKey secret is not a string")
	}

	return nil
}

var secretsOnce sync.Once

func secretsMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// attempt to fetch our secrets 1 time only when the first request comes in.
		secretsOnce.Do(func() {
			ctx := appengine.NewContext(r)

			err := initKey(ctx)
			if err != nil {
				log.Errorf(ctx, "unable to init secrets: %s", err)
			}
		})

		h(w, r)
	}
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// With GAE + Go<=1.9, the HTTP client cannot be reused across requests so the
	// client must get re-initiated each request with a client from GAE's "urlfetch".
	client := nyt.NewClient(nyt.DefaultHost, clientKey,
		kithttp.SetClient(urlfetch.Client(ctx)))

	stories, err := client.GetTopStories(context.Background(), "science")
	if err != nil {
		ctx := appengine.NewContext(r)
		log.Errorf(ctx, "unable to get stories: %s", err)
		http.Error(w, "unable to get top stories", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stories)
}

func warmUpHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
