package ui

import (
	"strings"
	"testing"
)

func TestBindingsForContext(t *testing.T) {
	tests := []struct {
		ctx      BindingContext
		wantKeys []string
	}{
		{ContextStorageList, []string{"Enter", "o", "c", "d", "j/k", "r", "Tab", "h", "q"}},
		{ContextSqlList, []string{"Enter", "j/k", "r", "Tab", "h", "q"}},
		{ContextConfigList, []string{"Enter", "a", "d", "j/k", "r", "Tab", "h", "q"}},
		{ContextBucketDetail, []string{"o", "d", "j/k", "Esc", "q"}},
		{ContextModal, []string{"Enter", "Esc"}},
	}
	for _, tt := range tests {
		bindings := BindingsForContext(tt.ctx)
		keys := make(map[string]bool)
		for _, b := range bindings {
			keys[b.Key] = true
		}
		for _, wantKey := range tt.wantKeys {
			if !keys[wantKey] {
				t.Errorf("context %s: missing key %q", tt.ctx, wantKey)
			}
		}
	}
}

func TestFormatHintsNonEmpty(t *testing.T) {
	result := FormatHints(ContextStorageList)
	if result == "" {
		t.Error("expected non-empty hint string for ContextStorageList")
	}
	if !strings.Contains(result, "Enter") && !strings.Contains(result, "q") {
		t.Error("hint string should reference keys")
	}
}
