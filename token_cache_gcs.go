package gcpvault

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

type TokenCacheGCS struct {
	cfg Config
}

func (t TokenCacheGCS) GetToken(ctx context.Context) (*Token, error) {

	if t.cfg.CachedToken != "" {
		//TODO retrieve JSON from bucket and convert to Token object
		bucket := "bucket-name"
		object := "object-name"
		client, err := storage.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("storage.NewClient: %v", err)
		}
		defer client.Close()

		ctx, cancel := context.WithTimeout(ctx, time.Second*50)
		defer cancel()

		rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
		if err != nil {
			return nil, fmt.Errorf("Object(%q).NewReader: %v", object, err)
		}
		defer rc.Close()

		return &Token{Token: "AAA"}, nil
	}

	return nil, nil

}

func (t TokenCacheGCS) SaveToken(token string) error {

	if t.cfg.CachedToken != "" {
		//TODO create json file.
		bucket := "bucket-name"
		object := "object-name"
		//TODO change above
		ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			return fmt.Errorf("storage.NewClient: %v", err)
		}
		defer client.Close()

		// Open local file. //TODO convert to JSON
		f, err := os.Open("notes.txt")
		if err != nil {
			return fmt.Errorf("os.Open: %v", err)
		}
		defer f.Close()

		ctx, cancel := context.WithTimeout(ctx, time.Second*50)
		defer cancel()

		// Upload an object with storage.Writer.
		wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
		if _, err = io.Copy(wc, f); err != nil {
			return fmt.Errorf("io.Copy: %v", err)
		}
		if err := wc.Close(); err != nil {
			return fmt.Errorf("Writer.Close: %v", err)
		}
		//fmt.Fprintf(w, "Blob %v uploaded.\n", object) //todo
		return nil
	}

	return nil
}
