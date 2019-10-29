// +build appengine

package gcpvault

import (
	"context"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func getDefaultServiceAccountEmail(ctx context.Context, cfg Config) (string, error) {
	return appengine.ServiceAccount(ctx)
}

func getHTTPClient(ctx context.Context, _ Config) *http.Client {
	return urlfetch.Client(ctx)
}
