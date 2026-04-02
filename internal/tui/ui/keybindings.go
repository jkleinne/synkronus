// File: internal/tui/ui/keybindings.go
package ui

import "strings"

// BindingContext identifies the TUI screen where a keybinding is active.
type BindingContext string

const (
	ContextStorageList    BindingContext = "Storage"
	ContextBucketDetail   BindingContext = "Bucket Detail"
	ContextObjectList     BindingContext = "Objects"
	ContextObjectDetail   BindingContext = "Object Detail"
	ContextSqlList        BindingContext = "SQL"
	ContextInstanceDetail BindingContext = "Instance Detail"
	ContextConfigList     BindingContext = "Config"
	ContextConfigEdit     BindingContext = "Config Edit"
	ContextModal          BindingContext = "Modal"
	ContextCreateForm     BindingContext = "Create Bucket"
)

// Binding describes a single keybinding shown in help and hint bars.
type Binding struct {
	Key         string
	Description string
	Context     BindingContext
}

// bindings is the single source of truth for all user-facing keybinding labels.
var bindings = []Binding{
	// Storage list
	{Key: "Enter", Description: "Describe", Context: ContextStorageList},
	{Key: "o", Description: "Objects", Context: ContextStorageList},
	{Key: "c", Description: "Create", Context: ContextStorageList},
	{Key: "d", Description: "Delete", Context: ContextStorageList},
	{Key: "j/k", Description: "Navigate", Context: ContextStorageList},
	{Key: "r", Description: "Refresh", Context: ContextStorageList},
	{Key: "Tab", Description: "Next Tab", Context: ContextStorageList},
	{Key: "h", Description: "Help", Context: ContextStorageList},
	{Key: "q", Description: "Quit", Context: ContextStorageList},

	// SQL list
	{Key: "Enter", Description: "Describe", Context: ContextSqlList},
	{Key: "j/k", Description: "Navigate", Context: ContextSqlList},
	{Key: "r", Description: "Refresh", Context: ContextSqlList},
	{Key: "Tab", Description: "Next Tab", Context: ContextSqlList},
	{Key: "h", Description: "Help", Context: ContextSqlList},
	{Key: "q", Description: "Quit", Context: ContextSqlList},

	// Config list
	{Key: "Enter", Description: "Edit", Context: ContextConfigList},
	{Key: "a", Description: "Add", Context: ContextConfigList},
	{Key: "d", Description: "Delete", Context: ContextConfigList},
	{Key: "j/k", Description: "Navigate", Context: ContextConfigList},
	{Key: "r", Description: "Refresh", Context: ContextConfigList},
	{Key: "Tab", Description: "Next Tab", Context: ContextConfigList},
	{Key: "h", Description: "Help", Context: ContextConfigList},
	{Key: "q", Description: "Quit", Context: ContextConfigList},

	// Bucket detail
	{Key: "o", Description: "Objects", Context: ContextBucketDetail},
	{Key: "d", Description: "Delete", Context: ContextBucketDetail},
	{Key: "j/k", Description: "Scroll", Context: ContextBucketDetail},
	{Key: "Esc", Description: "Back", Context: ContextBucketDetail},
	{Key: "q", Description: "Quit", Context: ContextBucketDetail},

	// Object list
	{Key: "Enter", Description: "Describe", Context: ContextObjectList},
	{Key: "j/k", Description: "Navigate", Context: ContextObjectList},
	{Key: "Esc", Description: "Back", Context: ContextObjectList},
	{Key: "q", Description: "Quit", Context: ContextObjectList},

	// Object detail
	{Key: "j/k", Description: "Scroll", Context: ContextObjectDetail},
	{Key: "Esc", Description: "Back", Context: ContextObjectDetail},
	{Key: "q", Description: "Quit", Context: ContextObjectDetail},

	// Instance detail
	{Key: "j/k", Description: "Scroll", Context: ContextInstanceDetail},
	{Key: "Esc", Description: "Back", Context: ContextInstanceDetail},
	{Key: "q", Description: "Quit", Context: ContextInstanceDetail},

	// Config edit
	{Key: "Enter", Description: "Save", Context: ContextConfigEdit},
	{Key: "Esc", Description: "Cancel", Context: ContextConfigEdit},

	// Modal
	{Key: "Enter", Description: "Confirm", Context: ContextModal},
	{Key: "Esc", Description: "Cancel", Context: ContextModal},

	// Create bucket form
	{Key: "Tab", Description: "Next Field", Context: ContextCreateForm},
	{Key: "Enter", Description: "Create", Context: ContextCreateForm},
	{Key: "Esc", Description: "Cancel", Context: ContextCreateForm},
}

// BindingsForContext returns all bindings matching the given context.
func BindingsForContext(ctx BindingContext) []Binding {
	var result []Binding
	for _, b := range bindings {
		if b.Context == ctx {
			result = append(result, b)
		}
	}
	return result
}

// FormatHints builds a centered hint bar from all bindings in a context.
func FormatHints(ctx BindingContext) string {
	ctxBindings := BindingsForContext(ctx)
	parts := make([]string, 0, len(ctxBindings))
	for _, b := range ctxBindings {
		key := HintKeyStyle.Render(b.Key)
		desc := HintDescStyle.Render(b.Description)
		parts = append(parts, key+" "+desc)
	}
	return strings.Join(parts, HintDescStyle.Render("  "))
}
