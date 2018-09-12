// +build !appengine

package gcpvault

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
)

func getDefaultServiceAccountEmail(ctx context.Context, cfg Config) (string, error) {
	result, err := callMetadataService(ctx, cfg)
	if err != nil {
		return "", errors.Wrap(err, "unable to retrieve default service account email")
	}
	return result, nil
}

func callMetadataService(ctx context.Context, cfg Config) (string, error) {
	c, err := google.DefaultClient(ctx)
	if err != nil {
		return "", errors.Wrap(err, "unable to init default client")
	}
	if cfg.MetadataAddress == "" {
		cfg.MetadataAddress = "http://metadata"
	}
	resp, err := c.Get(cfg.MetadataAddress +
		"/computeMetadata/v1/instance/service-accounts/default/email")
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("unexpected status response from metadata service: %d",
			resp.StatusCode)
	}
	bod, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "unable to read metadata response")
	}
	result := strings.TrimSpace(string(bod))
	if result == "" {
		return "", errors.New("unexpected empty response from metadata service")
	}
	return result, nil
}

func getHTTPClient(ctx context.Context) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			IdleConnTimeout: 1 * time.Second,
		},
	}
}
