package gaevault

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"

	iam "google.golang.org/api/iam/v1"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

// GetSecrets allows users to access Vault directly without the Teller layer.
// iamRole is the name of the Vault role given to your service account when configuring
// GCP and Vault. secretPath is the path of the secrets we wish to fetch from Vault
// with our IAM role.
//
// Under the hood this is using the Vault API client to log in, so make sure you inject
// the appropriate 'VAULT_*' environment variables like VAULT_ADDR.
//
// If running in a local development environment (via 'goapp test' or dev_appserver.py)
// this will look for a VAULT_LOCAL_TOKEN environment variable, which should contain
// the oken obtained after logging into Vault via the CLI tool.
func GetSecrets(ctx context.Context, iamRole, secretPath string) (map[string]interface{}, error) {
	if appengine.IsDevAppServer() {
		log.Debugf(ctx, "getting local secrets")
		return getLocalSecrets(ctx, secretPath)
	}

	// create signed JWT with our service account
	jwt, err := newJWT(ctx, iamRole)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create JWT")
	}

	// init vault client
	vcfg := api.DefaultConfig()
	vcfg.HttpClient = urlfetch.Client(ctx)
	vClient, err := api.NewClient(vcfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init vault client")
	}

	// 'login' to vault
	resp, err := vClient.Logical().Write("auth/gcp/login", map[string]interface{}{
		"role": iamRole, "jwt": jwt,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to login to vault")
	}

	vClient.SetToken(resp.Auth.ClientToken)

	// fetch secrets
	secrets, err := vClient.Logical().Read(secretPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get secrets")
	}

	return secrets.Data, nil
}

func getLocalSecrets(ctx context.Context, secretPath string) (map[string]interface{}, error) {
	// init vault client
	vcfg := api.DefaultConfig()
	vcfg.HttpClient = urlfetch.Client(ctx)
	vClient, err := api.NewClient(vcfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init vault client")
	}

	vClient.SetToken(os.Getenv("VAULT_LOCAL_TOKEN"))

	// fetch secrets
	secrets, err := vClient.Logical().Read(secretPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get secrets")
	}

	return secrets.Data, nil
}

// created JWT should match https://www.vaultproject.io/docs/auth/gcp.html#the-iam-authentication-token
func newJWT(ctx context.Context, iamRole string) (string, error) {
	serviceAccount, err := appengine.ServiceAccount(ctx)
	if err != nil {
		return "", errors.Wrap(err, "unable to find service account")
	}

	payload, err := json.Marshal(map[string]interface{}{
		"aud": "vault/" + iamRole,
		"sub": serviceAccount,
		"exp": time.Now().UTC().Add(5 * time.Minute).Unix(),
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to encode payload")
	}

	iamClient, err := iam.New(oauth2.NewClient(ctx,
		google.AppEngineTokenSource(ctx, iam.CloudPlatformScope)))
	if err != nil {
		return "", errors.Wrap(err, "unable to init IAM client")
	}

	resp, err := iamClient.Projects.ServiceAccounts.SignJwt(
		fmt.Sprintf("projects/%s/serviceAccounts/%s",
			appengine.AppID(ctx), serviceAccount),
		&iam.SignJwtRequest{Payload: string(payload)}).Context(ctx).Do()
	if err != nil {
		return "", errors.Wrap(err, "unable to sign JWT")
	}

	return resp.SignedJwt, nil
}
