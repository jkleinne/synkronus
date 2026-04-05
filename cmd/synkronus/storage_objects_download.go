package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"synkronus/internal/flags"
	"synkronus/internal/provider/storage/shared"

	"github.com/spf13/cobra"
)

func newDownloadObjectCmd() *cobra.Command {
	var provider string
	var bucket string
	var outputPath string

	cmd := &cobra.Command{
		Use:   "download [object-key]",
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
				return err
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

			return shared.WriteToFile(destPath, reader)
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
	basename, err := shared.ObjectBasename(objectKey)
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
