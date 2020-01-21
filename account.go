package gcpvault

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func getDefaultServiceAccountEmail(ctx context.Context, cfg Config) (string, error) {
	result, err := callMetadataService(ctx, cfg)
	if err != nil {
		return "", errors.Wrap(err, "unable to retrieve default service account email")
	}
	return result, nil
}

func callMetadataService(ctx context.Context, cfg Config) (string, error) {
	c := getHTTPClient(ctx, cfg)
	if cfg.MetadataAddress == "" {
		cfg.MetadataAddress = "http://metadata"
	}
	r, err := http.NewRequest(http.MethodGet,
		cfg.MetadataAddress+
			"/computeMetadata/v1/instance/service-accounts/default/email", nil)
	if err != nil {
		return "", errors.Wrap(err, "unable create metadata request")
	}
	r.Header.Add("Metadata-Flavor", "Google")

	resp, err := c.Do(r)
	if err != nil {
		return "", errors.Wrap(err, "error connecting to metadata server")
	}
	defer resp.Body.Close()

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

func getHTTPClient(ctx context.Context, cfg Config) *http.Client {
	if cfg.HTTPClient != nil {
		return cfg.HTTPClient
	}
	return &http.Client{
		Transport: &http.Transport{
			IdleConnTimeout: 1 * time.Second,
		},
	}
}
