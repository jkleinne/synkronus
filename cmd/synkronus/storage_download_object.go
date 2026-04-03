package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"synkronus/internal/flags"

	"github.com/spf13/cobra"
)

func newDownloadObjectCmd() *cobra.Command {
	var provider string
	var bucket string
	var outputPath string

	cmd := &cobra.Command{
		Use:   "download-object [object-key]",
		Short: "Download a storage object to a local file or stdout",
		Long:  `Downloads an object from a storage bucket. If --output-path is specified, writes to that file or directory. Otherwise, streams the object content to stdout for piping.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := appFromContext(cmd.Context())
			if err != nil {
				return err
			}

			objectKey := args[0]

			reader, err := app.StorageService.DownloadObject(cmd.Context(), bucket, objectKey, provider)
			if err != nil {
				return fmt.Errorf("error downloading object '%s' from bucket '%s' on %s: %w", objectKey, bucket, provider, err)
			}
			defer reader.Close()

			if outputPath == "" {
				_, err = io.Copy(os.Stdout, reader)
				if err != nil {
					return fmt.Errorf("error writing to stdout: %w", err)
				}
				return nil
			}

			destPath, err := resolveOutputPath(outputPath, objectKey)
			if err != nil {
				return err
			}

			return writeToFile(destPath, reader)
		},
	}

	cmd.Flags().StringVarP(&provider, flags.Provider, flags.ProviderShort, "", "The provider where the object resides (required)")
	cmd.MarkFlagRequired(flags.Provider)
	cmd.Flags().StringVarP(&bucket, flags.Bucket, flags.BucketShort, "", "The name of the bucket containing the object (required)")
	cmd.MarkFlagRequired(flags.Bucket)
	cmd.Flags().StringVar(&outputPath, flags.OutputPath, "", "File or directory path to write to (omit for stdout)")

	return cmd
}

// resolveOutputPath determines the final file path for the downloaded object.
// If outputPath is an existing directory (or ends with a path separator), the
// object's basename is appended. Otherwise, outputPath is used as-is.
func resolveOutputPath(outputPath string, objectKey string) (string, error) {
	basename, err := objectBasename(objectKey)
	if err != nil {
		return "", err
	}

	info, statErr := os.Stat(outputPath)
	if statErr == nil && info.IsDir() {
		return filepath.Join(outputPath, basename), nil
	}
	if strings.HasSuffix(outputPath, string(filepath.Separator)) {
		return filepath.Join(outputPath, basename), nil
	}

	return outputPath, nil
}

// objectBasename extracts a safe filename from an object key.
func objectBasename(objectKey string) (string, error) {
	if strings.HasSuffix(objectKey, "/") {
		return "", fmt.Errorf("cannot download directory marker object '%s'", objectKey)
	}
	base := filepath.Base(objectKey)
	if base == "." || base == "" {
		return "", fmt.Errorf("cannot derive filename from object key '%s'", objectKey)
	}
	return base, nil
}

// writeToFile creates the destination file, copies the reader content into it,
// and removes the file if the copy fails to avoid leaving partial data on disk.
func writeToFile(path string, src io.Reader) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file '%s': %w", path, err)
	}

	_, copyErr := io.Copy(f, src)
	closeErr := f.Close()

	if copyErr != nil {
		os.Remove(path)
		return fmt.Errorf("error writing to '%s': %w", path, copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("error closing '%s': %w", path, closeErr)
	}
	return nil
}
