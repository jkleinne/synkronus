package config

import "fmt"

// FlattenSettings converts a nested config map into a flat map with dot-notation keys.
// For example: {"gcp": {"project": "x"}} becomes {"gcp.project": "x"}.
func FlattenSettings(settings map[string]any) map[string]string {
	result := make(map[string]string)
	flattenRecursive(settings, "", result)
	return result
}

func flattenRecursive(m map[string]any, prefix string, result map[string]string) {
	for key, val := range m {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		switch v := val.(type) {
		case map[string]any:
			flattenRecursive(v, fullKey, result)
		default:
			if fullKey != "" {
				result[fullKey] = fmt.Sprintf("%v", v)
			}
		}
	}
}
