package marvinexample

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	gcpvault "github.com/NYTimes/gcp-vault"
	"github.com/pkg/errors"
)

func (s *service) initSecrets(ctx context.Context) error {
	var err error
	// only attempt to call vault 1 time at startup
	s.secretsOnce.Do(func() {
		envconfig.Process("", s.vaultConfig)

		// fetch the secrets
		var secrets map[string]interface{}
		secrets, err = gcpvault.GetSecrets(ctx, s.vaultConfig)
		if err != nil {
			return
		}

		// make sure the secret is there
		mySecret, ok := secrets["my-secret"]
		if !ok {
			err = errors.New("my-secret did not exist in vault")
			return
		}

		// make sure the secret has the right type
		s.mySecret, ok = mySecret.(string)
		if !ok {
			err = errors.New("my-secret was not a string type")
			return
		}
	})
	return errors.Wrap(err, "unable to init secrets")
}
