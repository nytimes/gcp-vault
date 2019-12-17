package gcpvault

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/vault/api"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iam/v1"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

func TestGetSecrets(t *testing.T) {
	tests := []struct {
		name          string
		givenCfg      Config
		givenSecrets  map[string]interface{}
		givenEmail    string
		givenVaultErr bool
		givenIAMErr   bool
		givenMetaErr  bool
		givenGAE      bool
		givenCreds    *google.Credentials

		wantVaultLogin bool
		wantVaultRead  bool
		wantIAMHit     bool
		wantMetaHit    bool
		wantErr        bool
		wantSecrets    map[string]interface{}
	}{
		{
			name: "local token, success",

			givenCfg: Config{
				Role:       "my-gcp-role",
				LocalToken: "my-local-token",
				SecretPath: "my-secret-path",
			},
			givenSecrets: map[string]interface{}{
				"my-sec":       "123",
				"my-other-sec": "abcd",
			},

			wantVaultRead: true,
			wantSecrets: map[string]interface{}{
				"my-sec":       "123",
				"my-other-sec": "abcd",
			},
		},
		{
			name: "GCP standard login, success",

			givenEmail: "jp@example.com",
			givenCfg: Config{
				Role:       "my-gcp-role",
				SecretPath: "my-secret-path",
			},
			givenSecrets: map[string]interface{}{
				"my-sec":       "123",
				"my-other-sec": "abcd",
			},
			givenCreds: &google.Credentials{
				ProjectID:   "test-project",
				TokenSource: testTokenSource{},
				JSON: []byte(`{  "client_id": "1234.apps.googleusercontent.com",
	  "client_secret": "abcd",  "refresh_token": "blah",
  "client_email": "",  "type": "service_account"}`),
			},

			wantVaultRead:  true,
			wantVaultLogin: true,
			wantIAMHit:     true,
			wantMetaHit:    true,
			wantSecrets: map[string]interface{}{
				"my-sec":       "123",
				"my-other-sec": "abcd",
			},
		},
		{
			name: "GAE standard login, success",

			givenCfg: Config{
				Role:       "my-gcp-role",
				SecretPath: "my-secret-path",
			},
			givenSecrets: map[string]interface{}{
				"my-sec":       "123",
				"my-other-sec": "abcd",
			},
			givenGAE: true,

			wantVaultRead:  true,
			wantVaultLogin: true,
			wantIAMHit:     true,
			wantSecrets: map[string]interface{}{
				"my-sec":       "123",
				"my-other-sec": "abcd",
			},
		},
		{
			name:       "GCP standard login, no meta email, fail",
			givenCreds: &google.Credentials{},
			givenCfg: Config{
				Role:       "my-gcp-role",
				SecretPath: "my-secret-path",
			},
			givenSecrets: map[string]interface{}{
				"my-sec":       "123",
				"my-other-sec": "abcd",
			},

			wantErr: true,
		},
		{
			name: "GCP standard login, vault fail",

			givenEmail: "jp@example.com",
			givenCfg: Config{
				Role:       "my-gcp-role",
				SecretPath: "my-secret-path",
			},
			givenSecrets: map[string]interface{}{
				"my-sec":       "123",
				"my-other-sec": "abcd",
			},
			givenVaultErr: true,

			wantErr: true,
		},
		{
			name: "GCP standard login, iam fail",

			givenEmail: "jp@example.com",
			givenCfg: Config{
				Role:       "my-gcp-role",
				SecretPath: "my-secret-path",
			},
			givenSecrets: map[string]interface{}{
				"my-sec":       "123",
				"my-other-sec": "abcd",
			},
			givenIAMErr: true,

			wantErr: true,
		},
		{
			name: "GCP standard login, meta fail",

			givenEmail: "jp@example.com",
			givenCreds: &google.Credentials{},
			givenCfg: Config{
				Role:       "my-gcp-role",
				SecretPath: "my-secret-path",
			},
			givenSecrets: map[string]interface{}{
				"my-sec":       "123",
				"my-other-sec": "abcd",
			},
			givenMetaErr: true,

			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				cfg           Config
				gotVaultLogin bool
				gotVaultRead  bool
				gotIAMHit     bool
				gotMetaHit    bool
			)
			// ensure defaults are set
			envconfig.Process("", &cfg)

			vaultSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					gotVaultRead = true
					json.NewEncoder(w).Encode(api.Secret{
						Data: test.givenSecrets,
					})
				case http.MethodPut:
					gotVaultLogin = true
					json.NewEncoder(w).Encode(api.Secret{
						Auth: &api.SecretAuth{
							ClientToken: "vault-test-token",
						},
					})
				}
			}))
			if test.givenVaultErr {
				vaultSvr.Close()
			} else {
				defer vaultSvr.Close()
			}

			iamSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotIAMHit = true
				json.NewEncoder(w).Encode(iam.SignJwtResponse{
					SignedJwt: "gcp-signed-jwt-for-vault",
				})
			}))
			if test.givenIAMErr {
				iamSvr.Close()
			} else {
				defer iamSvr.Close()
			}

			metaSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMetaHit = true
				io.WriteString(w, test.givenEmail)
			}))
			if test.givenMetaErr {
				metaSvr.Close()
			} else {
				defer metaSvr.Close()
			}

			if test.givenCreds != nil {
				findDefaultCredentials = func(ctx context.Context, scopes ...string) (*google.Credentials, error) {
					return test.givenCreds, nil
				}
				defer func() {
					findDefaultCredentials = google.FindDefaultCredentials
				}()
			}

			cfg.AuthPath = test.givenCfg.AuthPath
			cfg.SecretPath = test.givenCfg.SecretPath
			cfg.LocalToken = test.givenCfg.LocalToken
			cfg.IAMAddress = iamSvr.URL
			cfg.MetadataAddress = metaSvr.URL
			cfg.VaultAddress = vaultSvr.URL

			if appengine.IsDevAppServer() && !test.givenGAE {
				t.Log("in an app engine environment, skipping non GAE test")
				t.SkipNow()
				return
			}

			ctx := context.Background()
			if test.givenGAE {
				if !appengine.IsDevAppServer() {
					t.Log("skipping GAE test outside GAE environment")
					t.SkipNow()
					return
				}
				var (
					err  error
					done func()
				)
				ctx, done, err = aetest.NewContext()
				if err != nil {
					t.Fatalf("unable to start app engine %s", err)
				}
				defer done()
			}

			gotSecrets, gotErr := GetSecrets(ctx, cfg)
			if test.wantErr != (gotErr != nil) {
				t.Errorf("expected error %t, but got %s", test.wantErr, gotErr)
			}
			if test.wantErr {
				return
			}

			if test.wantIAMHit != gotIAMHit {
				t.Errorf("expected IAM hit? %t - got %t", test.wantIAMHit, gotIAMHit)
			}
			if test.wantMetaHit != gotMetaHit {
				t.Errorf("expected Meta hit? %t - got %t", test.wantMetaHit, gotMetaHit)
			}
			if test.wantVaultRead != gotVaultRead {
				t.Errorf("expected Vault read? %t - got %t", test.wantVaultRead, gotVaultRead)
			}
			if test.wantVaultLogin != gotVaultLogin {
				t.Errorf("expected Vault login? %t - got %t", test.wantVaultLogin, gotVaultLogin)
			}

			if !cmp.Equal(test.wantSecrets, gotSecrets) {
				t.Errorf("secrets differ: (-want +got)\n%s", cmp.Diff(test.wantSecrets, gotSecrets))
			}
		})
	}
}

func TestPutVersionedSecrets(t *testing.T) {
	tests := []struct {
		name            string
		givenCfg        Config
		startingSecrets map[string]interface{}
		givenEmail      string
		givenCreds      *google.Credentials

		wantVaultLogin bool
		wantVaultWrite bool
		wantIAMHit     bool
		wantMetaHit    bool
		putSecrets     map[string]interface{}
		wantSecrets    map[string]interface{}
	}{
		{
			name: "local token, success",

			givenCfg: Config{
				Role:       "my-gcp-role",
				LocalToken: "my-local-token",
				SecretPath: "my-secret-path",
			},
			startingSecrets: map[string]interface{}{
				"my-sec":       "123",
				"my-other-sec": "abcd",
			},
			wantVaultWrite: true,
			putSecrets: map[string]interface{}{
				"my-sec":       "456",
				"my-other-sec": "wxyz",
			},
			// versioned secrets are contained under a 'data' key
			wantSecrets: map[string]interface{}{
				"data": map[string]interface{}{
					"my-sec":       "456",
					"my-other-sec": "wxyz",
				},
			},
		},
		{
			name:       "GCP standard login, success",
			givenEmail: "jp@example.com",
			givenCfg: Config{
				Role:       "my-gcp-role",
				SecretPath: "my-secret-path",
			},
			startingSecrets: map[string]interface{}{
				"my-sec":       "123",
				"my-other-sec": "abcd",
			},
			wantVaultLogin: true,
			wantVaultWrite: true,
			wantIAMHit:     true,
			wantMetaHit:    true,
			putSecrets: map[string]interface{}{
				"my-sec":       "456",
				"my-other-sec": "wxyz",
			},
			givenCreds: &google.Credentials{
				ProjectID:   "test-project",
				TokenSource: testTokenSource{},
				JSON: []byte(`{  "client_id": "1234.apps.googleusercontent.com",
			  "client_secret": "abcd",  "refresh_token": "blah",
		  "client_email": "",  "type": "service_account"}`),
			},

			// versioned secrets are contained under a 'data' key
			wantSecrets: map[string]interface{}{
				"data": map[string]interface{}{
					"my-sec":       "456",
					"my-other-sec": "wxyz",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				cfg           Config
				gotVaultLogin bool
				gotVaultWrite bool
				gotIAMHit     bool
				gotMetaHit    bool
				secrets       map[string]interface{}
			)
			// ensure defaults are set
			envconfig.Process("", &cfg)

			secrets = test.startingSecrets

			cachedClient = nil

			vaultSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodPut:
					gotVaultLogin = true
					json.NewEncoder(w).Encode(api.Secret{
						Auth: &api.SecretAuth{
							ClientToken: "vault-test-token",
						},
					})
				case http.MethodPost:
					gotVaultWrite = true
					var incoming map[string]interface{}
					json.NewDecoder(r.Body).Decode(&incoming)
					secrets = incoming
				}
			}))

			iamSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotIAMHit = true
				json.NewEncoder(w).Encode(iam.SignJwtResponse{
					SignedJwt: "gcp-signed-jwt-for-vault",
				})
			}))
			defer iamSvr.Close()

			metaSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMetaHit = true
				io.WriteString(w, test.givenEmail)
			}))
			defer metaSvr.Close()

			if test.givenCreds != nil {
				findDefaultCredentials = func(ctx context.Context, scopes ...string) (*google.Credentials, error) {
					return test.givenCreds, nil
				}
				defer func() {
					findDefaultCredentials = google.FindDefaultCredentials
				}()
			}

			cfg.AuthPath = test.givenCfg.AuthPath
			cfg.SecretPath = test.givenCfg.SecretPath
			cfg.LocalToken = test.givenCfg.LocalToken
			cfg.IAMAddress = iamSvr.URL
			cfg.MetadataAddress = metaSvr.URL
			cfg.VaultAddress = vaultSvr.URL

			if appengine.IsDevAppServer() {
				t.Log("in an app engine environment, skipping non GAE test")
				t.SkipNow()
				return
			}

			ctx := context.Background()

			err := PutVersionedSecrets(ctx, cfg, test.putSecrets)
			if err != nil {
				t.Errorf("expected no error, got err: %s", err)
			}

			if test.wantIAMHit != gotIAMHit {
				t.Errorf("expected IAM hit? %t - got %t", test.wantIAMHit, gotIAMHit)
			}
			if test.wantMetaHit != gotMetaHit {
				t.Errorf("expected Meta hit? %t - got %t", test.wantMetaHit, gotMetaHit)
			}
			if test.wantVaultLogin != gotVaultLogin {
				t.Errorf("expected Vault login? %t - got %t", test.wantVaultLogin, gotVaultLogin)
			}
			if test.wantVaultWrite != gotVaultWrite {
				t.Errorf("expected Vault write? %t - got %t", test.wantVaultWrite, gotVaultWrite)
			}
			if !cmp.Equal(test.wantSecrets, secrets) {
				t.Errorf("secrets differ: (-want +got)\n%s", cmp.Diff(test.wantSecrets, secrets))
			}
		})
	}
}

type testTokenSource struct{}

func (t testTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}
