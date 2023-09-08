package gcs

import (
	"context"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

func Backup(ctx context.Context, input string, dest string, credsFile string) error {
	fmt.Println("Uploading to GCS at", dest)

	file, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("opening %s: %w", input, err)
	}

	srv, err := storage.NewClient(ctx, option.WithCredentialsFile(credsFile))
	if err != nil {
		return fmt.Errorf("creating GCS client: %w", err)
	}

	wc := srv.Bucket(dest).Object("backup.tar.gz").NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("writer.Close: %w", err)
	}
	fmt.Println("Copied to GCS under name:", wc.Name)
	return nil
}
