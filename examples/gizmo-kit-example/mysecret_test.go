package gizmoexample

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	gcpvault "github.com/nytimes/gcp-vault"
	"github.com/nytimes/gcp-vault/gcpvaulttest"
	"github.com/nytimes/gizmo/server/kit"
)

func TestMySecretEndpoint(t *testing.T) {
	vaultSvr := gcpvaulttest.NewVaultServer(map[string]interface{}{"my-secret": "abcdefg"})
	defer vaultSvr.Close()

	cfg := gcpvault.Config{
		VaultAddress: vaultSvr.URL,
		// passing a local token so we only attempt to call the vault server
		// otherwise, we'd need to also start up the IAM server to mock out JWT signing
		LocalToken: "abcd",
	}
	svc := &service{vaultConfig: cfg}
	err := svc.initSecrets(context.Background())
	if err != nil {
		t.Fatalf("unable to init secrets: %s", err)
	}
	svr := kit.NewServer(svc)

	r := httptest.NewRequest(http.MethodGet, "/svc/example/v1/my-secret", nil)
	wr := httptest.NewRecorder()
	svr.ServeHTTP(wr, r)

	w := wr.Result()

	got, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("unable to read response: %s", err)
	}

	if string(got) != "{\"my-secret\":\"abcdefg\"}\n" {
		t.Errorf("expected `my-secret`, got %q", string(got))
	}
}
