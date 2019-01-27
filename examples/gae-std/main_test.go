package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/NYTimes/gcp-vault/examples/nyt"
	"google.golang.org/appengine/aetest"
)

func TestTopStories(t *testing.T) {
	wr := httptest.NewRecorder()

	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer inst.Close()

	r, err := inst.NewRequest(http.MethodGet, "/my-handler", nil)
	if err != nil {
		t.Fatal(err)
	}

	secretsMiddleware(myHandler)(wr, r)

	w := wr.Result()

	if w.StatusCode != http.StatusOK {
		t.Errorf("non-200 response from function: %d", w.StatusCode)
	}

	var res nyt.TopStoriesResponse
	err = json.NewDecoder(w.Body).Decode(&res)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Body.Close()

	out := json.NewEncoder(os.Stdout)
	out.SetIndent("", "    ")
	out.Encode(res)
}
