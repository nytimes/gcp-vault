package marvinexample

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"

	gcpvault "github.com/NYTimes/gcp-vault"
	"github.com/NYTimes/gcp-vault/gcpvaulttest"
	"github.com/NYTimes/marvin"
)

func TestMySecretEndpoint(t *testing.T) {
	if !appengine.IsDevAppServer() {
		t.Skip()
		return
	}
	vaultSvr := gcpvaulttest.NewVaultServer(map[string]interface{}{"my-secret": "abcdefg"})
	defer vaultSvr.Close()

	cfg := gcpvault.Config{
		VaultAddress: vaultSvr.URL,
		// passing a local token so we only attempt to call the vault server
		// otherwise, we'd need to also start up the IAM server to mock out JWT signing
		LocalToken: "abcd",
	}
	svc := &service{vaultConfig: cfg}
	svr := marvin.NewServer(svc)

	testInst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("unable to setup aetest instance: %s", err)
		return
	}
	defer testInst.Close()

	r, err := testInst.NewRequest(http.MethodGet, "/svc/example/v1/my-secret", nil)
	if err != nil {
		t.Fatalf("unable to create new gae test request: %s", err)
	}
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
