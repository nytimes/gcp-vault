package gcpvaulttest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	iam "google.golang.org/api/iam/v1"

	"github.com/hashicorp/vault/api"
)

func NewVaultServer(secrets map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(api.Secret{
				Data: secrets,
			})
		case http.MethodPut:
			json.NewEncoder(w).Encode(api.Secret{
				Auth: &api.SecretAuth{
					ClientToken: "vault-test-token",
				},
			})
		}
	}))
}

func NewIAMServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(iam.SignJwtResponse{
			SignedJwt: "gcp-signed-jwt-for-vault",
		})
	}))
}

func NewMetadataServer(serviceAcctEmail string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, serviceAcctEmail)
	}))
}
