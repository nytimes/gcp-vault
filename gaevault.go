package gaevault

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func GetSecrets(ctx context.Context, iamRole, secretPath string) (map[string]interface{}, error) {
	if appengine.IsDevAppServer() {
		return getLocalSecrets(ctx, secretPath)
	}

	// create signed JWT with our service account
	jwt, err := newJWT(ctx, iamRole)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create JWT")
	}

	// init vault client
	vcfg := api.DefaultConfig()
	vcfg.HttpClient = urlfetch.Client(ctx)
	vClient, err := api.NewClient(vcfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init vault client")
	}

	// 'login' to vault
	_, err = vClient.Logical().Write("auth/gcp/login", map[string]interface{}{
		"role": iamRole, "jwt": jwt,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to login to vault")
	}

	// fetch secrets
	secrets, err := vClient.Logical().Read(secretPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get secrets")
	}

	return secrets.Data, nil
}

func getLocalSecrets(ctx context.Context, secretPath string) (map[string]interface{}, error) {
	// init vault client
	vcfg := api.DefaultConfig()
	vcfg.HttpClient = urlfetch.Client(ctx)
	vClient, err := api.NewClient(vcfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init vault client")
	}

	vClient.SetToken(os.Getenv("VAULT_LOCAL_TOKEN"))

	// fetch secrets
	secrets, err := vClient.Logical().Read(secretPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get secrets")
	}

	return secrets.Data, nil
}

// created JWT should match https://www.vaultproject.io/docs/auth/gcp.html#the-iam-authentication-token
// and be signed with GAE service acct
func newJWT(ctx context.Context, iamRole string) (string, error) {
	svc, err := appengine.ServiceAccount(ctx)
	if err != nil {
		return "", errors.Wrap(err, "unable to find service account")
	}

	h := map[string]string{"alg": "RS256", "typ": "JWT"}
	p := map[string]string{
		"sub": svc,
		"aud": "vault/" + iamRole,
		"exp": strconv.FormatInt(time.Now().UTC().Add(15*time.Minute).Unix(), 10),
	}
	encode := func(i map[string]string) (string, error) {
		b, err := json.Marshal(i)
		if err != nil {
			return "", err
		}
		return base64.RawURLEncoding.EncodeToString(b), nil
	}
	header, err := encode(h)
	if err != nil {
		return "", err
	}
	payload, err := encode(p)
	if err != nil {
		return "", err
	}

	ss := fmt.Sprintf("%s.%s", header, payload)
	_, sig, err := appengine.SignBytes(ctx, []byte(ss))
	if err != nil {
		return "", errors.Wrap(err, "unable to sign JWT")
	}
	return fmt.Sprintf("%s.%s", ss, base64.RawURLEncoding.EncodeToString(sig)), nil
}
