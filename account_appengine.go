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

func getHTTPClient(ctx context.Context) *http.Client {
	return urlfetch.Client(ctx)
}
