package output

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// TableRenderer is implemented by types that can produce their own ASCII
// table representation. Render dispatches to RenderTable when FormatTable
// is requested.
type TableRenderer interface {
	RenderTable() string
}

// Render writes data to w in the requested format.
//
// FormatJSON and FormatYAML serialize data directly using the standard
// encoding packages. FormatTable requires data to implement TableRenderer;
// if it does not, an error is returned.
//
// An error is returned for any format not defined in this package.
func Render(w io.Writer, format Format, data any) error {
	switch format {
	case FormatJSON:
		return renderJSON(w, data)
	case FormatYAML:
		return renderYAML(w, data)
	case FormatTable:
		return renderTable(w, data)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func renderJSON(w io.Writer, data any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func renderYAML(w io.Writer, data any) error {
	enc := yaml.NewEncoder(w)
	defer enc.Close()
	return enc.Encode(data)
}

func renderTable(w io.Writer, data any) error {
	renderer, ok := data.(TableRenderer)
	if !ok {
		return fmt.Errorf("data type %T does not support table rendering", data)
	}
	_, err := io.WriteString(w, renderer.RenderTable())
	return err
}
