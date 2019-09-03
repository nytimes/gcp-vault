package gcfexample

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	gcpvault "github.com/NYTimes/gcp-vault"
	"github.com/NYTimes/gcp-vault/examples/nyt"
	"github.com/kelseyhightower/envconfig"
	"github.com/GoogleCloudPlatform/berglas/pkg/berglas"
)

func init() {
	// Unlike GAE standard environment, GCF allows users to access the network on
	// startup. This allows us to fetch our secrets in the init() function instead of
	// hooking it in as a middleware.
	err := initClient(context.Background())
	if err != nil {
		log.Printf("unable to init client: %s", err)
	}
}

var client nyt.Client

func initClient(ctx context.Context) error {
	if err := ReplaceEnvVarIfBerglas(ctx, "VAULT_ADDR"); err != nil {
		return err
	}

	if err := ReplaceEnvVarIfBerglas(ctx, "VAULT_TOKEN"); err != nil {
		return err
	}

	var cfg gcpvault.Config
	envconfig.Process("", &cfg)

	secrets, err := gcpvault.GetSecrets(ctx, cfg)
	if err != nil {
		return err
	}

	client = nyt.NewClient(nyt.DefaultHost, secrets["APIKey"].(string))
	return nil
}

func ReplaceEnvVarIfBerglas(ctx, context.Context, envVar string) error {
	val := os.Getenv(envVar)
	if strings.HasPrefix(val, "berglas://") {
		if err := berglas.Replace(ctx, envVar); err != nil {
			return err
		}
	}

	return nil
}

func GetTopScienceStories(w http.ResponseWriter, r *http.Request) {
	stories, err := client.GetTopStories(r.Context(), "science")
	if err != nil {
		http.Error(w, "unable to get top stories", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stories)
}
