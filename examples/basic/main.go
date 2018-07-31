package main

import (
	"net/http"
	"sync"

	"github.com/kelseyhightower/envconfig"
	gaevault "github.com/nytm/gae-vault"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func main() {
	// register secret-fetching middleware on all endpoints as warm up requests are NOT
	// guaranteed.
	http.HandleFunc("/_ah/warmup", secretsMiddleware(warmUpHandler))
	http.HandleFunc("/my-handler", secretsMiddleware(myHandler))
	appengine.Main()
}

var (
	secrets     map[string]interface{}
	secretsOnce sync.Once
)

func secretsMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// attempt to fetch our secrets 1 time only when the first request comes in.
		secretsOnce.Do(func() {
			ctx := appengine.NewContext(r)

			var cfg gaevault.Config
			envconfig.Process("", &cfg)

			var err error
			secrets, err = gaevault.GetSecrets(ctx, cfg)
			if err != nil {
				log.Errorf(ctx, "unable to fetch secrets: %s", err)
			}
		})

		h(w, r)
	}
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	// do something with our secrets here!
	_ = secrets["my-secret"].(string)

	w.Write([]byte("welcome to my secret service!"))
}

func warmUpHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
