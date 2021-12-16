package kitexample

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NYTimes/gizmo/server/kit"
	"github.com/nytimes/gcp-vault/examples/nyt"
)

func TestTopStories(t *testing.T) {
	testKey := "my-key"
	nytSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("api-key"); got != testKey {
			t.Errorf("expected test key %q, got %q", testKey, got)
		}
		io.WriteString(w, "{}")
	}))
	defer nytSvr.Close()

	svc := &service{client: nyt.NewClient(nytSvr.URL, testKey)}
	svr := kit.NewServer(svc)

	r := httptest.NewRequest(http.MethodGet, "/svc/example/v1/top-stories", nil)
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
