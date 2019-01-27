package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	gcpvault "github.com/NYTimes/gcp-vault"
	"github.com/NYTimes/gcp-vault/examples/nyt"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/common/log"

	"google.golang.org/appengine"
)

func main() {
	// register secret-fetching middleware on all endpoints as warm up requests are NOT
	// guaranteed.
	http.HandleFunc("/_ah/warmup", secretsMiddleware(warmUpHandler))
	http.HandleFunc("/my-handler", secretsMiddleware(myHandler))
	appengine.Main()
}

var client nyt.Client

func initClient(ctx context.Context) error {
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
	key, ok := keyI.(string)
	if !ok {
		return errors.New("APIKey secret is not a string")
	}

	client = nyt.NewClient(nyt.DefaultHost, key)
	return nil
}

var secretsOnce sync.Once

func secretsMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// attempt to fetch our secrets 1 time only when the first request comes in.
		secretsOnce.Do(func() {
			err := initClient(appengine.NewContext(r))
			if err != nil {
				log.Errorf(ctx, "unable to init secrets: %s", err)
			}
		})

		h(w, r)
	}
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	stories, err := client.GetTopStories(context.Background(), "science")
	if err != nil {
		http.Error(w, "unable to get top stories", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stories)
}

func warmUpHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
