package main

import (
	"fmt"
	"os"
	"path/filepath"
	"synkronus/internal/domain/storage"
	"synkronus/internal/flags"

	"github.com/spf13/cobra"
)

func newUploadObjectCmd() *cobra.Command {
	var provider string
	var bucket string
	var objectKey string
	var contentType string
	var metadata map[string]string

	cmd := &cobra.Command{
		Use:   "upload [file]",
		Short: "Upload a local file as a storage object",
		Long:  `Uploads a local file to a storage bucket. If --key is omitted, the object key is derived from the filename.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			filePath := args[0]

			info, err := os.Stat(filePath)
			if err != nil {
				return fmt.Errorf("cannot access file %q: %w", filePath, err)
			}
			if info.IsDir() {
				return fmt.Errorf("path is a directory, not a file: %s", filePath)
			}

			if objectKey == "" {
				objectKey = filepath.Base(filePath)
			}

			f, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("opening file %q: %w", filePath, err)
			}
			defer f.Close()

			opts := storage.UploadObjectOptions{
				BucketName:  bucket,
				ObjectKey:   objectKey,
				ContentType: contentType,
				Metadata:    metadata,
			}

			if err := app.StorageService.UploadObject(cmd.Context(), opts, provider, f); err != nil {
				return err
			}

			fmt.Printf("Object '%s' uploaded successfully to bucket '%s' on provider %s.\n", objectKey, bucket, provider)
			return nil
		},
	}

	cmd.Flags().StringVarP(&provider, flags.Provider, flags.ProviderShort, "", "The provider where the bucket resides (required)")
	cmd.MarkFlagRequired(flags.Provider)
	cmd.Flags().StringVarP(&bucket, flags.Bucket, flags.BucketShort, "", "The name of the target bucket (required)")
	cmd.MarkFlagRequired(flags.Bucket)
	cmd.Flags().StringVar(&objectKey, flags.ObjectKey, "", "Object key (defaults to filename if omitted)")
	cmd.Flags().StringVar(&contentType, flags.ContentType, "", "Content-Type MIME type (auto-detected if omitted)")
	cmd.Flags().StringToStringVar(&metadata, flags.Metadata, nil, "User-defined metadata as key=value pairs")

	return cmd
}
