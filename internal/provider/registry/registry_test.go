package registry

import (
	"context"
	"log/slog"
	"synkronus/internal/config"
	"testing"
)

func newTestRegistry() *Registry[string] {
	return NewRegistry[string]("test")
}

func dummyReg(val string) Registration[string] {
	return Registration[string]{
		ConfigCheck: func(cfg *config.Config) bool { return true },
		Initializer: func(ctx context.Context, cfg *config.Config, logger *slog.Logger) (string, error) {
			return val, nil
		},
	}
}

func TestRegister_And_GetSupported(t *testing.T) {
	r := newTestRegistry()
	r.Register("beta", dummyReg("b"))
	r.Register("alpha", dummyReg("a"))

	supported := r.GetSupported()
	if len(supported) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(supported))
	}
	if supported[0] != "alpha" || supported[1] != "beta" {
		t.Errorf("expected sorted [alpha beta], got %v", supported)
	}
}

func TestRegister_Duplicate_Panics(t *testing.T) {
	r := newTestRegistry()
	r.Register("dup", dummyReg("first"))

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	r.Register("dup", dummyReg("second"))
}

func TestIsSupported(t *testing.T) {
	r := newTestRegistry()
	r.Register("gcp", dummyReg("g"))

	if !r.IsSupported("gcp") {
		t.Error("expected gcp to be supported")
	}
	if !r.IsSupported("GCP") {
		t.Error("expected case-insensitive match")
	}
	if r.IsSupported("aws") {
		t.Error("expected aws to be unsupported")
	}
}

func TestGet(t *testing.T) {
	r := newTestRegistry()
	r.Register("gcp", dummyReg("g"))

	reg, ok := r.Get("gcp")
	if !ok {
		t.Fatal("expected gcp to be found")
	}
	val, err := reg.Initializer(context.Background(), nil, slog.Default())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "g" {
		t.Errorf("expected 'g', got %q", val)
	}

	_, ok = r.Get("missing")
	if ok {
		t.Error("expected missing provider to not be found")
	}
}

func TestGetAll(t *testing.T) {
	r := newTestRegistry()
	r.Register("a", dummyReg("1"))
	r.Register("b", dummyReg("2"))

	all := r.GetAll()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}

	// Verify it's a copy by modifying it
	delete(all, "a")
	if len(r.GetAll()) != 2 {
		t.Error("GetAll should return a copy, not the underlying map")
	}
}
