package gcpvault

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"cloud.google.com/go/storage"
)

type TokenCacheGCS struct {
	cfg *Config
}

func (t TokenCacheGCS) GetToken(ctx context.Context) (*Token, error) {

	if t.cfg.TokenCache != nil {
		bucket := t.cfg.TokenCacheStorageGCS
		object := t.cfg.TokenCacheKeyName
		client, err := storage.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("error creating new storage client: %v", err)
		}
		defer client.Close()

		rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
		if err != nil {
			// swallowing the error here since we may not have cached a token yet
			return nil, nil
		}
		defer rc.Close()

		data, err := ioutil.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("error reading: %v", err)
		}
		var token Token
		err = json.Unmarshal(data, &token)
		if err != nil {
			return nil, err
		}
		return &token, nil
	}

	return nil, nil

}

func (t TokenCacheGCS) SaveToken(ctx context.Context, token Token) error {

	if t.cfg.TokenCache != nil {

		bucket := t.cfg.TokenCacheStorageGCS
		object := t.cfg.TokenCacheKeyName

		client, err := storage.NewClient(ctx)
		if err != nil {
			return fmt.Errorf("error creating new storage client: %v", err)
		}
		defer client.Close()

		// Upload an object with storage.Writer.
		wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
		wc.ContentType = "application/json"
		payload, err := json.Marshal(&token)
		if err != nil {
			return fmt.Errorf("error mashaling: %v", err)
		}
		if _, err := wc.Write(payload); err != nil {
			return fmt.Errorf("error writing: %v", err)
		}
		if err := wc.Close(); err != nil {
			return fmt.Errorf("error closing: %v", err)
		}
		return nil
	}

	return nil
}
