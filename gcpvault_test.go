package gcpvault_test

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
	iam "google.golang.org/api/iam/v1"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"

	gcpvault "github.com/NYTimes/gcp-vault"
)

func TestGetSecrets(t *testing.T) {
	tests := []struct {
		name          string
		givenCfg      gcpvault.Config
		givenSecrets  map[string]interface{}
		givenEmail    string
		givenVaultErr bool
		givenIAMErr   bool
		givenMetaErr  bool
		givenGAE      bool

		wantVaultLogin bool
		wantVaultRead  bool
		wantIAMHit     bool
		wantMetaHit    bool
		wantErr        bool
		wantSecrets    map[string]interface{}
	}{
		{
			name: "local token, success",

			givenCfg: gcpvault.Config{
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
			givenCfg: gcpvault.Config{
				Role:       "my-gcp-role",
				SecretPath: "my-secret-path",
			},
			givenSecrets: map[string]interface{}{
				"my-sec":       "123",
				"my-other-sec": "abcd",
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

			givenCfg: gcpvault.Config{
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
			name: "GCP standard login, no meta email, fail",

			givenCfg: gcpvault.Config{
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
			givenCfg: gcpvault.Config{
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
			givenCfg: gcpvault.Config{
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
			givenCfg: gcpvault.Config{
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
				cfg           gcpvault.Config
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

			gotSecrets, gotErr := gcpvault.GetSecrets(ctx, cfg)
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
