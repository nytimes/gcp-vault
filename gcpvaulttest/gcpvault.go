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
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": secrets,
			})
		case http.MethodPost, http.MethodPut:
			var incoming map[string]interface{}
			json.NewDecoder(r.Body).Decode(&incoming)
			secrets = incoming
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
