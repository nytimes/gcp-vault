package gcpvault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	iam "google.golang.org/api/iam/v1"
)

// Config contains fields for configuring access and secrets retrieval from a Vault
// server.
type Config struct {
	// SecretPath is the location of the secrets we wish to fetch from Vault.
	SecretPath string `envconfig:"VAULT_SECRET_PATH"`

	// VaultAddress is the location of the Vault server.
	VaultAddress string `envconfig:"VAULT_ADDR"`

	// Role is the role given to your service account when it was registered
	// with your Vault server. More information about creating roles for your service
	// account can be found here:
	// https://www.vaultproject.io/docs/auth/gcp.html#2-roles
	Role string `envconfig:"VAULT_GCP_IAM_ROLE"`

	// LocalToken is a Vault auth token obtained from logging into Vault via some outside
	// method like the command line tool. Users are only expected to pass this token
	// in local development scenarios.
	// This token can also be set in the `VAULT_TOKEN` environment variable and the
	// underlying Vault API client will use it.
	LocalToken string `envconfig:"VAULT_LOCAL_TOKEN"`

	// AuthPath is the path the GCP authentication method is mounted at.
	// Defaults to 'auth/gcp'.
	AuthPath string `envconfig:"VAULT_GCP_PATH"`

	// MaxRetries sets the number of retries that will be used in the case of certain
	// errors. The underlying Vault client will pull this value out of the environment
	// on it's own, but we're including it here so users can apply the same number of
	// attempts towards signing the JWT with Google's IAM services.
	MaxRetries int `envconfig:"VAULT_MAX_RETRIES"`

	// IAMAddress is the location of the GCP IAM server.
	// This should only used for testing.
	IAMAddress string `envconfig:"IAM_ADDR"`

	// MetadataAddress is the location of the GCP metadata
	// This should only used for testing.
	MetadataAddress string `envconfig:"METADATA_ADDR"`

	// HTTPClient can be optionally set if users wish to have more control over outbound
	// HTTP requests made by this library. If not set, an http.Client with a 1s
	// IdleConnTimeout will be used.
	HTTPClient *http.Client
}

// GetSecrets will use GCP Auth to access any secrets under the given SecretPath in
// Vault.
//
// This is comparable to the `vault read` command.
//
// Under the hood, this uses a JWT signed with the default Google application
// credentials to login to Vault via
// https://godoc.org/github.com/hashicorp/vault/api#Logical.Write and to read secrets via
// https://godoc.org/github.com/hashicorp/vault/api#Logical.Read. For more details about
// enabling GCP Auth and Vault visit: https://www.vaultproject.io/docs/auth/gcp.html
//
// The map[string]interface{} returned is the actual contents of the secret referenced in
// the Config.SecretPath.
//
// This is using the Vault API client's 'default config' to log in so users can provide
// additional environment variables to fine tune their Vault experience. For more
// information about configuring the Vault API client, view the code behind:
// https://godoc.org/github.com/hashicorp/vault/api#Config.ReadEnvironment
//
// If running in a local development environment (via 'goapp test' or dev_appserver.py)
// this tool will expect the LocalToken to be set in some way.
func GetSecrets(ctx context.Context, cfg Config) (map[string]interface{}, error) {
	checkDefaults(&cfg)

	vClient, err := login(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to login to vault")
	}

	// fetch secrets
	secrets, err := vClient.Logical().Read(cfg.SecretPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get secrets")
	}
	if secrets == nil {
		return nil, errors.New("no secrets found")
	}
	if (secrets.Data == nil || len(secrets.Data) == 0) && secrets.Warnings != nil {
		err := errors.New(strings.Join(secrets.Warnings, ","))
		return nil, errors.Wrap(err, "no secrets found")
	}
	return secrets.Data, nil
}

// PutSecrets writes secrets to Vault at the configured path.
// This is comparable to the `vault write` command.
func PutSecrets(ctx context.Context, cfg Config, secrets map[string]interface{}) error {
	checkDefaults(&cfg)
	vClient, err := login(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "unable to login to vault")
	}
	_, err = vClient.Logical().Write(cfg.SecretPath, secrets)
	return errors.Wrap(err, "unable to make vault request")
}

// GetVersionedSecrets reads versioned secrets from Vault.
// This is comparable to the `vault kv get` command.
func GetVersionedSecrets(ctx context.Context, cfg Config) (map[string]interface{}, error) {
	checkDefaults(&cfg)
	secs, err := GetSecrets(ctx, cfg)
	if err != nil {
		return nil, err
	}
	// versioned secrets are contained under a 'data' key
	s, ok := secs["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("no data in versioned secrets")
	}
	return s, nil
}

// PutVersionedSecrets writes versioned secrets to Vault at the configured path.
// This is comparable to the `vault kv put` command.
func PutVersionedSecrets(ctx context.Context, cfg Config, secrets map[string]interface{}) error {
	checkDefaults(&cfg)
	vClient, err := login(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "unable to login to vault")
	}

	req := vClient.NewRequest(http.MethodPost, "/v1/"+cfg.SecretPath)
	req.BodyBytes, err = json.Marshal(map[string]map[string]interface{}{
		"data": secrets,
	})
	if err != nil {
		return errors.Wrap(err, "unable to marshal request body")
	}
	_, err = vClient.RawRequestWithContext(ctx, req)
	return errors.Wrap(err, "unable to make vault request")
}

func checkDefaults(cfg *Config) {
	if cfg == nil {
		return
	}

	if cfg.AuthPath == "" {
		cfg.AuthPath = "auth/gcp"
	}
}

func login(ctx context.Context, cfg Config) (*api.Client, error) {
	if cfg.LocalToken != "" {
		return newLocalClient(ctx, cfg)
	}

	vClient, err := newClient(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init vault client")
	}

	// create signed JWT with our service account
	jwt, err := newJWT(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create JWT")
	}

	// 'login' to vault using GCP auth
	resp, err := vClient.Logical().Write(cfg.AuthPath+"/login", map[string]interface{}{
		"role": cfg.Role, "jwt": jwt,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to make login request")
	}

	vClient.SetToken(resp.Auth.ClientToken)

	return vClient, nil
}

func newClient(ctx context.Context, cfg Config) (*api.Client, error) {
	vcfg := api.DefaultConfig()
	vcfg.MaxRetries = cfg.MaxRetries
	vcfg.Address = cfg.VaultAddress
	vcfg.HttpClient = getHTTPClient(ctx, cfg)
	return api.NewClient(vcfg)
}

func newLocalClient(ctx context.Context, cfg Config) (*api.Client, error) {
	vcfg := api.DefaultConfig()
	vcfg.Address = cfg.VaultAddress
	vcfg.HttpClient = getHTTPClient(ctx, cfg)
	vClient, err := api.NewClient(vcfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init vault client")
	}

	vClient.SetToken(cfg.LocalToken)

	return vClient, nil
}

func newJWT(ctx context.Context, cfg Config) (string, error) {
	var (
		jwt string
		err error
	)
	for retries := 0; retries <= cfg.MaxRetries; retries++ {
		jwt, err = newJWTBase(ctx, cfg)
		if err == nil {
			return jwt, nil
		}
	}
	return "", errors.Wrapf(err, "unable to sign JWT after %d retries", cfg.MaxRetries)
}

// created JWT should match https://www.vaultproject.io/docs/auth/gcp.html#the-iam-authentication-token
func newJWTBase(ctx context.Context, cfg Config) (string, error) {
	serviceAccount, project, tokenSource, err := getServiceAccountInfo(ctx, cfg)
	if err != nil {
		return "", errors.Wrap(err, "unable to get service account from environment")
	}

	payload, err := json.Marshal(map[string]interface{}{
		"aud": "vault/" + cfg.Role,
		"sub": serviceAccount,
		"exp": time.Now().UTC().Add(5 * time.Minute).Unix(),
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to encode JWT payload")
	}

	hc := getHTTPClient(ctx, cfg)
	// reuse timeout but sprinkle on the token source for IAM access
	hcIAM := &http.Client{
		Timeout: hc.Timeout,
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}
	iamClient, err := iam.New(hcIAM)
	if err != nil {
		return "", errors.Wrap(err, "unable to init IAM client")
	}

	if cfg.IAMAddress != "" {
		iamClient.BasePath = cfg.IAMAddress
	}

	resp, err := iamClient.Projects.ServiceAccounts.SignJwt(
		fmt.Sprintf("projects/%s/serviceAccounts/%s",
			project, serviceAccount),
		&iam.SignJwtRequest{Payload: string(payload)}).Context(ctx).Do()
	if err != nil {
		return "", errors.Wrap(err, "unable to sign JWT")
	}
	return resp.SignedJwt, nil
}

var findDefaultCredentials = google.FindDefaultCredentials

func getServiceAccountInfo(ctx context.Context, cfg Config) (string, string, oauth2.TokenSource, error) {
	creds, err := findDefaultCredentials(ctx, iam.CloudPlatformScope)
	if err != nil {
		return "", "", nil, errors.Wrap(err, "unable to find credentials to sign JWT")
	}

	serviceAccountEmail, err := getEmailFromCredentials(creds)
	if err != nil {
		return "", "", nil, errors.Wrap(err, "unable to get email from given credentials")
	}

	if serviceAccountEmail == "" {
		serviceAccountEmail, err = getDefaultServiceAccountEmail(ctx, cfg)
		if err != nil {
			return "", "", nil, err
		}
	}

	return serviceAccountEmail, creds.ProjectID, creds.TokenSource, nil
}

func getEmailFromCredentials(creds *google.Credentials) (string, error) {
	if len(creds.JSON) == 0 {
		return "", nil
	}

	var data map[string]string
	err := json.Unmarshal(creds.JSON, &data)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse credentials")
	}

	return data["client_email"], nil
}
