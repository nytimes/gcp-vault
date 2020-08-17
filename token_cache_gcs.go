package gcpvault

import (
	"context"
)

type TokenCacheGCS struct {
	cfg Config
}

func (t TokenCacheGCS) GetToken(ctx context.Context) (*Token, error) {
	return nil, nil

}

func (t TokenCacheGCS) SaveToken(token Token) error {
	return nil
}
