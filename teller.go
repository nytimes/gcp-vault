package gaevault

import (
	"context"
	"sync"
)

// Teller is a mechanism to cache secrets from Vault.
type Teller struct {
	secrets map[string]interface{}
	once    sync.Once

	k KMSInfo
	v VaultInfo
}

// NewTeller will return a Teller instance to fetch secrets just one time.
func NewTeller(k KMSInfo, v VaultInfo) *Teller {
	return &Teller{k: k, v: v}
}

// Tell will get user secrets. If they have already been fetched, Vault will not be
// contacted again.
func (t *Teller) Tell(ctx context.Context) (map[string]interface{}, error) {
	var err error
	t.once.Do(func() {
		t.secrets, err = GetSecrets(ctx, t.k, t.v)
	})
	return t.secrets, err
}
