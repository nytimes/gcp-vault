package gcfexample

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/nytimes/gcp-vault/examples/nyt"
)

func TestTopStories(t *testing.T) {
	if os.Getenv("VAULT_SECRET_PATH") == "" {
		t.Skip()
		return
	}
	wr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	GetTopScienceStories(wr, r)

	w := wr.Result()

	if w.StatusCode != http.StatusOK {
		t.Errorf("non-200 response from function: %d", w.StatusCode)
	}

	var res nyt.TopStoriesResponse
	err := json.NewDecoder(w.Body).Decode(&res)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Body.Close()

	out := json.NewEncoder(os.Stdout)
	out.SetIndent("", "    ")
	out.Encode(res)
}
