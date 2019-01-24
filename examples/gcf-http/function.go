package gcfexample

import (
	"context"
	"net/http"

	gcpvault "github.com/NYTimes/gcp-vault"
	"github.com/kelseyhightower/envconfig"
)

func init() {
	// Unlike GAE standard environment, GCF allows users to access the network on
	// startup. This allows us to fetch our secrets in the init() function instead of
	// hooking it in as a middleware.
	initSecrets(context.Background())
}

var secrets map[string]interface{}

func initSecrets(ctx context.Context) error {
	var cfg gcpvault.Config
	envconfig.Process("", &cfg)

	var err error
	secrets, err = gcpvault.GetSecrets(ctx, cfg)
	return err
}

func MyFunction(w http.ResponseWriter, r *http.Request) {
	secret := secrets["my-secret"].(string)

	w.Write([]byte("the secret is: " + secret))
}
