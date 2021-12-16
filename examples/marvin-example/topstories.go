package marvinexample

import (
	"context"
	"net/http"

	"github.com/NYTimes/marvin"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/nytimes/gcp-vault/examples/nyt"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

func (s *service) getTopStories(ctx context.Context, _ interface{}) (interface{}, error) {
	// With GAE + Go<=1.9, the HTTP client cannot be reused across requests so the
	// client must get re-initiated each request with a client from GAE's "urlfetch".
	client := nyt.NewClient(s.nytHost, s.apiKey,
		kithttp.SetClient(urlfetch.Client(ctx)))

	stories, err := client.GetTopStories(context.Background(), "science")
	if err != nil {
		log.Errorf(ctx, "unable to get stories: %s", err)
		return nil, marvin.NewJSONStatusResponse("unable to get stories",
			http.StatusInternalServerError)
	}

	return stories, nil
}
