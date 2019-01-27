package nyt

import (
	"context"
	"encoding/json"
	fmt "fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/pkg/errors"
)

type Client struct {
	getTopStories endpoint.Endpoint
}

const DefaultHost = "https://api.nytimes.com"

func NewClient(host, key string, opts ...kithttp.ClientOption) Client {
	if host == "" {
		host = DefaultHost
	}
	return Client{
		getTopStories: retryEndpoint(kithttp.NewClient(
			http.MethodGet,
			mustParseURL(host, "/svc/topstories/v2/{SECTION}.json"),
			encodeTopStories(key),
			decodeTopStories,
			opts...,
		).Endpoint()),
	}
}

func (c Client) GetTopStories(ctx context.Context, section string) (*TopStoriesResponse, error) {
	out, err := c.getTopStories(ctx, section)
	if out != nil {
		return out.(*TopStoriesResponse), err
	}
	return nil, err
}

func mustParseURL(host, path string) *url.URL {
	r, err := url.Parse(host + path)
	if err != nil {
		panic("invalid url: " + err.Error())
	}
	return r
}

func encodeTopStories(key string) kithttp.EncodeRequestFunc {
	return func(ctx context.Context, r *http.Request, req interface{}) error {
		r.URL.Path = strings.Replace(r.URL.Path, "{SECTION}", req.(string), 1)
		q := r.URL.Query()
		q.Add("api-key", key)
		r.URL.RawQuery = q.Encode()
		return nil
	}
}

func decodeTopStories(ctx context.Context, r *http.Response) (interface{}, error) {
	bod, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read response")
	}
	defer r.Body.Close()

	switch r.StatusCode {
	case http.StatusOK:
		var res TopStoriesResponse
		err = json.Unmarshal(bod, &res)
		return &res, errors.Wrap(err, "unable to decode response")
	default:
		return nil, fmt.Errorf("unpexpected response: [%d] %q",
			r.StatusCode, string(bod))
	}
}

func retryEndpoint(e endpoint.Endpoint) endpoint.Endpoint {
	bl := sd.NewEndpointer(
		sd.FixedInstancer{"1"},
		sd.Factory(func(_ string) (endpoint.Endpoint, io.Closer, error) {
			return e, nil, nil
		}),
		log.NewNopLogger(),
	)
	return lb.RetryWithCallback(5*time.Second, lb.NewRoundRobin(bl),
		func(n int, received error) (keepTrying bool, replacement error) {
			if n > 2 {
				return false, received
			}
			return true, nil
		})
}

type (
	TopStoriesResponse struct {
		Status      string    `json:"status"`
		Copyright   string    `json:"copyright"`
		Section     string    `json:"section"`
		LastUpdated string    `json:"last_updated"`
		NumResults  int       `json:"num_results"`
		Results     []Article `json:"results"`
	}

	Article struct {
		Section           string        `json:"section"`
		Subsection        string        `json:"subsection"`
		Title             string        `json:"title"`
		Abstract          string        `json:"abstract"`
		URL               string        `json:"url"`
		Byline            string        `json:"byline"`
		ItemType          string        `json:"item_type"`
		UpdatedDate       string        `json:"updated_date"`
		CreatedDate       string        `json:"created_date"`
		PublishedDate     string        `json:"published_date"`
		MaterialTypeFacet string        `json:"material_type_facet"`
		Kicker            string        `json:"kicker"`
		DesFacet          []string      `json:"des_facet"`
		OrgFacet          []string      `json:"org_facet"`
		PerFacet          []string      `json:"per_facet"`
		GeoFacet          []interface{} `json:"geo_facet"`
		Multimedia        []struct {
			URL       string `json:"url"`
			Format    string `json:"format"`
			Height    int    `json:"height"`
			Width     int    `json:"width"`
			Type      string `json:"type"`
			Subtype   string `json:"subtype"`
			Caption   string `json:"caption"`
			Copyright string `json:"copyright"`
		} `json:"multimedia"`
		ShortURL string `json:"short_url"`
	}
)
