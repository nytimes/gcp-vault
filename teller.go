package gaevault

import (
	"context"
	"sync"
)

// TellFunc will be provided by users to receive secrets from the Teller.
type TellFunc func(context.Context, map[string]interface{}) error

// Teller is a mechanism to lazily fetch secrets from Vault.
type Teller struct {
	tellFunc TellFunc
	tellOnce sync.Once

	sPath, iamRole string
}

// NewTeller will return a Teller instance to fetch secrets just one time.
// iamRole is the name of the Vault role given to your service account when configuring
// GCP and Vault. secretPath is the path of the secrets we wish to fetch from Vault
// with our IAM role.
func NewTeller(iamRole, secretPath string, tellFunc TellFunc) *Teller {
	return &Teller{iamRole: iamRole, sPath: secretPath}
}

// Tell will get secrets from Vault and call the given TellFunc with them. If they have
// already been fetched, Vault will not be contacted again.
// Users will likely want to put this method call in their service middleware and enable
// warm up requests in hopes of fetching the secrets before exposing the service to
// users.
func (t *Teller) Tell(ctx context.Context) error {
	var err error
	t.tellOnce.Do(func() {
		var secrets map[string]interface{}
		secrets, err = GetSecrets(ctx, t.iamRole, t.sPath)
		if err != nil {
			return
		}
		err = t.tellFunc(ctx, secrets)
	})
	return err
}
