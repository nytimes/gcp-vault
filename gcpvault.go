package gcpvault

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iam/v1"
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

	TokenCache TokenCache
	// How long before the token expiration should it be regenerated (in seconds).
	// Default is 300 seconds.
	TokenCacheRefreshThreshold int `envconfig:"TOKEN_CACHE_REFRESH_THRESHOLD"`
	//Random refresh offset in seconds to avoid all the instances refreshing at once. Default is 1/2 the duration in seconds of the TOKEN_CACHE_REFRESH_THRESHOLD.
	TokenCacheRefreshRandomOffset int `envconfig:"TOKEN_CACHE_REFRESH_RANDOM_OFFSET"`
	// this value is in seconds. Default value is 30 seconds
	TokenCacheCtxTimeout int `envconfig:"TOKEN_CACHE_CTX_TIMEOUT"`
	// the object name to store. Default value is 'token-cache'
	TokenCacheKeyName string `envconfig:"TOKEN_CACHE_KEY_NAME"`
	// GCS bucket location where token can be stored for caching purposes
	TokenCacheStorageGCS string `envconfig:"TOKEN_CACHE_STORAGE_GCS"`
	// Host and port for Redis '10.200.30.4:6379'
	TokenCacheStorageRedis string `envconfig:"TOKEN_CACHE_STORAGE_REDIS"`
	//Database for Redis. Default is 0
	TokenCacheStorageRedisDB int `envconfig:"TOKEN_CACHE_STORAGE_REDIS_DB"`
}

type TokenCache interface {
	GetToken(ctx context.Context) (*Token, error)
	SaveToken(ctx context.Context, token Token) error
}

type Token struct {
	Token   string
	Expires time.Time
}

const (
	CachedTokenRefreshThresholdDefault   = 300
	TokenCacheCtxTimeoutDefault          = 30
	TokenCacheRefreshRandomOffsetDefault = 60
	TokenCacheKeyNameDefault             = "token-cache"
	TokenCacheMaxRetriesDefault          = 3
)

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
	err := checkDefaults(&cfg)
	if err != nil {
		return nil, err
	}

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
	err := checkDefaults(&cfg)
	if err != nil {
		return err
	}

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
	err := checkDefaults(&cfg)
	if err != nil {
		return nil, err
	}

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
	err := checkDefaults(&cfg)
	if err != nil {
		return err
	}

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

func checkDefaults(cfg *Config) error {
	if cfg == nil {
		return errors.New("configuration is empty")
	}

	if cfg.TokenCacheStorageGCS != "" && cfg.TokenCacheStorageRedis != "" {
		return errors.New("Both Cache types are configured")
	}

	if cfg.AuthPath == "" {
		cfg.AuthPath = "auth/gcp"
	}

	if cfg.TokenCacheStorageGCS != "" && cfg.TokenCache == nil {
		cfg.TokenCache = TokenCacheGCS{cfg: cfg}
	}

	if cfg.TokenCacheStorageRedis != "" && cfg.TokenCache == nil {
		cfg.TokenCache = TokenCacheRedis{cfg: cfg}
	}

	//if expiration is not set, use default
	if cfg.TokenCacheRefreshThreshold == 0 {
		cfg.TokenCacheRefreshThreshold = CachedTokenRefreshThresholdDefault
	}

	//if token cache timeout is not set, use default
	if cfg.TokenCacheCtxTimeout == 0 {
		cfg.TokenCacheCtxTimeout = TokenCacheCtxTimeoutDefault
	}

	//if the token cache name is not set, use default
	if cfg.TokenCacheKeyName == "" {
		cfg.TokenCacheKeyName = TokenCacheKeyNameDefault
	}

	//if max retries is not set, use default
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = TokenCacheMaxRetriesDefault
	}

	if cfg.TokenCacheRefreshRandomOffset == 0 && cfg.TokenCacheRefreshThreshold > 0 {
		// setting random offset to 1/2 of the refresh threshold
		seconds := cfg.TokenCacheRefreshThreshold / 2

		cfg.TokenCacheRefreshRandomOffset = seconds
	} else if cfg.TokenCacheRefreshRandomOffset == 0 {
		// TOKEN_CACHE_REFRESH_RANDOM_OFFSET is not set
		cfg.TokenCacheRefreshRandomOffset = TokenCacheRefreshRandomOffsetDefault
	}

	return nil
}

func login(ctx context.Context, cfg Config) (*api.Client, error) {
	if cfg.LocalToken != "" {
		return newLocalClient(ctx, cfg)
	}

	vClient, err := newClient(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init vault client")
	}

	timeout := time.Duration(cfg.TokenCacheCtxTimeout)
	ctx, cancel := context.WithTimeout(ctx, time.Second*timeout)
	defer cancel()

	b := backoff.NewExponentialBackOff()

	var token Token
	if cfg.TokenCache != nil {
		token, err = getVaultTokenFromCache(ctx, cfg, b)
	}
	//an error with gcs or redis
	if err != nil {
		return nil, err
	}

	//token is missing from cache or expired
	if token.Token == "" {
		//generate new token from Vault
		token, err := getToken(ctx, cfg, vClient)
		if err != nil {
			return nil, err
		}

		vClient.SetToken(token.Auth.ClientToken)
		//save to cache
		err = persistVaultTokenToCache(ctx, cfg, token, b)
		if err != nil {
			return nil, err
		}
		return vClient, nil
	}

	vClient.SetToken(token.Token)
	return vClient, nil
}

func getVaultTokenFromCache(ctx context.Context, cfg Config, b *backoff.ExponentialBackOff) (Token, error) {
	var (
		token *Token
		err   error
	)
	err = backoff.Retry(func() error {
		token, err = cfg.TokenCache.GetToken(ctx)
		return err
	}, backoff.WithMaxRetries(b, uint64(cfg.MaxRetries)))

	if err != nil {
		return Token{}, errors.Wrapf(err, "unable to retrieve Vault token from cache after %d retries", cfg.MaxRetries)
	}

	if !(isExpired(token, cfg) || isRevoked(ctx, cfg, token)) {
		return *token, nil
	}
	//token is expired
	return Token{}, nil
}

func persistVaultTokenToCache(ctx context.Context, cfg Config, token *api.Secret, b *backoff.ExponentialBackOff) error {
	if cfg.TokenCache != nil {
		tokenExpiration, err := token.TokenTTL()
		if err != nil {
			return errors.Wrap(err, "unable to retrieve token ttl")
		}

		now := time.Now()
		err = backoff.Retry(func() error {
			err = cfg.TokenCache.SaveToken(ctx, Token{Token: token.Auth.ClientToken, Expires: now.Add(tokenExpiration)})
			return err
		}, backoff.WithMaxRetries(b, uint64(cfg.MaxRetries)))

		if err != nil {
			return errors.Wrapf(err, "unable to save Vault token to cache after %d retries", cfg.MaxRetries)
		}

	}
	return nil
}

func getToken(ctx context.Context, cfg Config, vClient *api.Client) (*api.Secret, error) {

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

	return resp, nil
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

	b := backoff.NewExponentialBackOff()

	err = backoff.Retry(func() error {
		jwt, err = newJWTBase(ctx, cfg)
		return err
	}, backoff.WithMaxRetries(b, uint64(cfg.MaxRetries)))

	if err != nil {
		return "", errors.Wrapf(err, "unable to sign JWT after %d retries", cfg.MaxRetries)
	}

	return jwt, nil
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
	// reuse base transport and timeout but sprinkle on the token source for IAM access
	hcIAM := &http.Client{
		Timeout: hc.Timeout,
		Transport: &oauth2.Transport{
			Source: tokenSource,
			Base:   hc.Transport,
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

func isExpired(token *Token, cfg Config) bool {
	if token == nil {
		return true
	}

	refreshTime := time.Now().Add(time.Second * time.Duration(cfg.TokenCacheRefreshThreshold))
	//seed random generator
	rand.Seed(time.Now().UnixNano())
	//subtract random number of seconds from the expiration to avoid many simultaneous refresh events
	refreshTime = refreshTime.Add(time.Second * (-1 * time.Duration(rand.Intn(cfg.TokenCacheRefreshRandomOffset))))

	if refreshTime.After(token.Expires) {
		return true
	}

	return false
}

func isRevoked(ctx context.Context, cfg Config, token *Token) bool {
	vClient, err := newClient(ctx, cfg)
	if err != nil {
		return true
	}
	vClient.SetToken(token.Token)
	_, err = vClient.Auth().Token().LookupSelf()
	if err != nil {
		return true
	}
	return false
}
