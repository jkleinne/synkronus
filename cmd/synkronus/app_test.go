package main

import (
	"context"
	"testing"
)

func TestAppFromContext_MissingContainer(t *testing.T) {
	ctx := context.Background()
	_, err := appFromContext(ctx)
	if err == nil {
		t.Fatal("expected error for missing container")
	}
}

func TestToContext_RoundTrip(t *testing.T) {
	app := &appContainer{}
	ctx := app.ToContext(context.Background())
	got, err := appFromContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != app {
		t.Error("expected same pointer back from context")
	}
}
