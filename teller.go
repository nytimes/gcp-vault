package gaevault

import (
	"context"
	"sync"
)

// Teller is a mechanism to cache secrets from Vault.
type Teller struct {
	secrets map[string]interface{}
	once    sync.Once

	sPath, iamRole string
}

// NewTeller will return a Teller instance to fetch secrets just one time.
func NewTeller(iamRole, secretPath string) *Teller {
	return &Teller{iamRole: iamRole, sPath: secretPath}
}

// Tell will get user secrets. If they have already been fetched, Vault will not be
// contacted again.
func (t *Teller) Tell(ctx context.Context) (map[string]interface{}, error) {
	var err error
	t.once.Do(func() {
		t.secrets, err = GetSecrets(ctx, t.iamRole, t.sPath)
	})
	return t.secrets, err
}
