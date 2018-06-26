package gaevault

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"

	iam "google.golang.org/api/iam/v1"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

// Config contains fields for configuring access and secrets retrieval from a Vault
// server.
type Config struct {
	// SecretPath is the location of the secrets we whish to fetch from Vault.
	SecretPath string `VAULT_SECRET_PATH`

	// Address is the location of the Vault server.
	Address string `VAULT_ADDR`

	// Role is the role given to your service account when it was registered
	// with your Vault server. More information about creating roles for your service
	// account can be found here:
	// https://www.vaultproject.io/docs/auth/gcp.html#2-roles
	Role string `VAULT_GCP_IAM_ROLE`

	// LocalToken is a Vault auth token obtained from logging into Vault via some outside
	// method like the command line tool. Users are only expected to pass this token
	// in local development scenarios.
	// This token can also be set in the `VAULT_TOKEN` environment variable and the
	// underlying Vault API client will use it.
	LocalToken string `VAULT_LOCAL_TOKEN`
}

// GetSecrets will use GCP Auth to access any secrets under the given SecretPath in
// Vault. Under the hood, this uses a JWT signed with the App Engine service account to
// login to Vault. For more details about enabling GCP Auth and Vault visit:
// https://www.vaultproject.io/docs/auth/gcp.html
//
// This is using the Vault API client's 'default config' to log in so users can provide
// additional environment variables to fine tune their Vault experience. For more
// information about configuring the Vault API client, visit:
// https://github.com/hashicorp/vault/blob/master/api/client.go#L215
//
// If running in a local development environment (via 'goapp test' or dev_appserver.py)
// this tool will expect the LocalToken to be set in some way.
func GetSecrets(ctx context.Context, cfg Config) (map[string]interface{}, error) {
	if appengine.IsDevAppServer() || cfg.LocalToken != "" {
		return getLocalSecrets(ctx, cfg)
	}

	// create signed JWT with our service account
	jwt, err := newJWT(ctx, cfg.Role)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create JWT")
	}

	// init vault client
	vcfg := api.DefaultConfig()
	vcfg.Address = cfg.Address
	vcfg.HttpClient = urlfetch.Client(ctx)
	vClient, err := api.NewClient(vcfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init vault client")
	}

	// 'login' to vault using GCP auth
	resp, err := vClient.Logical().Write("auth/gcp/login", map[string]interface{}{
		"role": cfg.Role, "jwt": jwt,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to login to vault")
	}

	vClient.SetToken(resp.Auth.ClientToken)

	// fetch secrets
	secrets, err := vClient.Logical().Read(cfg.SecretPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get secrets")
	}

	return secrets.Data, nil
}

func getLocalSecrets(ctx context.Context, cfg Config) (map[string]interface{}, error) {
	vcfg := api.DefaultConfig()
	vcfg.Address = cfg.Address
	vcfg.HttpClient = urlfetch.Client(ctx)
	vClient, err := api.NewClient(vcfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init vault client")
	}

	vClient.SetToken(cfg.LocalToken)

	// fetch secrets
	secrets, err := vClient.Logical().Read(cfg.SecretPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get secrets")
	}

	return secrets.Data, nil
}

// created JWT should match https://www.vaultproject.io/docs/auth/gcp.html#the-iam-authentication-token
func newJWT(ctx context.Context, role string) (string, error) {
	serviceAccount, err := appengine.ServiceAccount(ctx)
	if err != nil {
		return "", errors.Wrap(err, "unable to find service account")
	}

	payload, err := json.Marshal(map[string]interface{}{
		"aud": "vault/" + role,
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
