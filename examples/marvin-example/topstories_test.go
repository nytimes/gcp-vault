package marvinexample

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"

	gcpvault "github.com/NYTimes/gcp-vault"
	"github.com/NYTimes/gcp-vault/gcpvaulttest"
	"github.com/NYTimes/marvin"
)

func TestTopStories(t *testing.T) {
	if !appengine.IsDevAppServer() {
		t.Skip()
	}
	testKey := "my-test-key"
	vaultSvr := gcpvaulttest.NewVaultServer(map[string]interface{}{"APIKey": testKey})
	defer vaultSvr.Close()

	nytSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("api-key"); got != testKey {
			t.Errorf("expected test key %q, got %q", testKey, got)
		}
		io.WriteString(w, "{}")
	}))
	defer nytSvr.Close()

	cfg := gcpvault.Config{
		VaultAddress: vaultSvr.URL,
		// passing a local token so we only attempt to call the vault server
		// otherwise, we'd need to also start up the IAM server to mock out JWT signing
		LocalToken: "abcd",
	}
	svc := &service{vcfg: cfg, nytHost: nytSvr.URL}
	svr := marvin.NewServer(svc)

	testInst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("unable to setup aetest instance: %s", err)
		return
	}
	defer testInst.Close()

	r, err := testInst.NewRequest(http.MethodGet, "/svc/example/v1/top-stories", nil)
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

	want := `{"status":"","copyright":"","section":"","last_updated":"","num_results":0,"results":null}`
	if strings.TrimSpace(string(got)) != want {
		t.Errorf("expected %q, got %q", want, string(got))
	}
}
