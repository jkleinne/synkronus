package service

import "context"

// withClientResult acquires a provider client, calls fn, and ensures the
// client is closed afterward. Used by service methods that return a value.
// Error wrapping for the specific operation belongs in fn, not here.
func withClientResult[C ProviderClient, R any](
	ctx context.Context,
	getClient func(ctx context.Context, name string) (C, error),
	providerName string,
	fn func(client C) (R, error),
) (R, error) {
	client, err := getClient(ctx, providerName)
	if err != nil {
		var zero R
		return zero, err
	}
	defer client.Close()
	return fn(client)
}
