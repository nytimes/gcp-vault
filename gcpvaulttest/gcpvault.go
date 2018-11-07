package gcpvaulttest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/hashicorp/vault/api"
	iam "google.golang.org/api/iam/v1"
)

func NewVaultServer(secrets map[string]interface{}) *httptest.Server {
	var mu sync.Mutex

	if secrets == nil {
		secrets = map[string]interface{}{}
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/login") {
			json.NewEncoder(w).Encode(api.Secret{
				Auth: &api.SecretAuth{ClientToken: "vault-test-token"},
			})
			return
		}

		mu.Lock()
		defer mu.Unlock()

		switch r.Method {
		case http.MethodGet:
			var v interface{}
			if strings.Contains(r.URL.Path, "/versioned/") {
				v = map[string]interface{}{"data": secrets}
			} else {
				v = api.Secret{Data: secrets}
			}
			json.NewEncoder(w).Encode(v)
		case http.MethodPut: // non-versioned secrets save
			var incoming map[string]interface{}
			json.NewDecoder(r.Body).Decode(&incoming)
			secrets = incoming
		case http.MethodPost: // versioned secrets save
			var incoming map[string]interface{}
			json.NewDecoder(r.Body).Decode(&incoming)
			secrets = incoming["data"].(map[string]interface{})
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
