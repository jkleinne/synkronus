package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

// ProviderClient is the constraint for provider clients that can be closed.
type ProviderClient interface {
	io.Closer
}

// concurrentFanOut runs listFn concurrently across all providers, collecting
// results and errors. Partial results are returned alongside any errors.
func concurrentFanOut[C ProviderClient, T any](
	ctx context.Context,
	providerNames []string,
	getClient func(ctx context.Context, name string) (C, error),
	listFn func(ctx context.Context, client C) ([]T, error),
	logger *slog.Logger,
) ([]T, error) {
	var allResults []T
	var errs []error
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, providerName := range providerNames {
		wg.Add(1)
		go func(providerName string) {
			defer wg.Done()

			client, err := getClient(ctx, providerName)
			if err != nil {
				logger.Error("Failed to initialize provider client", "provider", providerName, "error", err)
				mu.Lock()
				errs = append(errs, fmt.Errorf("provider %s: %w", providerName, err))
				mu.Unlock()
				return
			}
			defer client.Close()

			results, err := listFn(ctx, client)
			if err != nil {
				logger.Error("Failed to list from provider", "provider", providerName, "error", err)
				mu.Lock()
				errs = append(errs, fmt.Errorf("provider %s: %w", providerName, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			allResults = append(allResults, results...)
			mu.Unlock()

			logger.Debug("Successfully fetched results", "provider", providerName, "count", len(results))
		}(providerName)
	}

	wg.Wait()

	return allResults, errors.Join(errs...)
}
