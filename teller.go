package gaevault

import (
	"context"
	"sync"
)

// TellFunc will be provided by users to receive secrets from the Teller.
type TellFunc func(context.Context, map[string]interface{}) error

// Teller is a mechanism to lazily fetch secrets from Vault.
type Teller struct {
	secrets map[string]interface{}
	secOnce sync.Once

	tellFunc TellFunc

	sPath, iamRole string
}

// NewTeller will return a Teller instance to fetch secrets just one time.
// iamRole is the name of the Vault role given to your service account when configuring
// GCP and Vault. secretPath is the path of the secrets we wish to fetch from Vault
// with our IAM role.
//
// Under the hood this is using the Vault API client to log in, so make sure you inject
// the appropriate 'VAULT_*' environment variables like VAULT_ADDR.
//
// If running in a local development environment (via 'goapp test' or dev_appserver.py)
// this will look for a VAULT_LOCAL_TOKEN environment variable, which should contain
// the oken obtained after logging into Vault via the CLI tool.
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
	t.secOnce.Do(func() {
		t.secrets, err = GetSecrets(ctx, t.iamRole, t.sPath)
		if err != nil {
			return
		}
	})
	return t.tellFunc(ctx, t.secrets)
}

// SetTellFunc allows users to reuse a Teller. This is meant to be used for
// reusing a Teller across a test suite.
func (t *Teller) SetTellFunc(f TellFunc) {
	t.tellFunc = f
}
