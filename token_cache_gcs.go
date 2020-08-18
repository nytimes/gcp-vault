package gcpvault

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"cloud.google.com/go/storage"
)

type TokenCacheGCS struct {
	cfg *Config
}

func (t TokenCacheGCS) GetToken(ctx context.Context) (*Token, error) {

	if t.cfg.TokenCache != nil {
		bucket := t.cfg.TokenCacheStorageGCS
		object := "token-cache"
		client, err := storage.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("storage.NewClient: %v", err)
		}
		defer client.Close()

		ctx, cancel := context.WithTimeout(ctx, time.Second*50)
		defer cancel()

		rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
		if err != nil {
			// swallowing the error here since we may not have cached a token yet
			return nil, nil
		}
		defer rc.Close()

		data, err := ioutil.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("ioutil.ReadAll: %v", err)
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

func (t TokenCacheGCS) SaveToken(token Token) error {

	if t.cfg.TokenCache != nil {

		bucket := t.cfg.TokenCacheStorageGCS
		object := "token-cache"

		ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			return fmt.Errorf("storage.NewClient: %v", err)
		}
		defer client.Close()

		ctx, cancel := context.WithTimeout(ctx, time.Second*50)
		defer cancel()

		// Upload an object with storage.Writer.
		wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
		wc.ContentType = "application/json"
		payload, err := json.Marshal(&token)
		if err != nil {
			return fmt.Errorf("json.Marshal: %v", err)
		}
		if _, err := wc.Write(payload); err != nil {
			return fmt.Errorf("wc.Write: %v", err)
		}
		if err := wc.Close(); err != nil {
			return fmt.Errorf("Writer.Close: %v", err)
		}
		return nil
	}

	return nil
}
