package xdiff

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FlattenJSON decodes raw JSON bytes and flattens into dot-paths.
func FlattenJSON(data []byte) (map[string]any, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return Flatten(v), nil
}

// FlattenYAML decodes YAML bytes and flattens into dot-paths.
func FlattenYAML(data []byte) (map[string]any, error) {
	var v any
	if err := yaml.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return Flatten(toStringKeyMap(v)), nil
}

// LoadAndFlatten loads a file with extension-based decoding (json|yaml|yml).
func LoadAndFlatten(path string) (map[string]any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	switch ext := filepath.Ext(path); ext {
	case ".json":
		return FlattenJSON(b)
	case ".yaml", ".yml":
		return FlattenYAML(b)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// Flatten converts any nested maps/slices into dot/bracket paths → values.
func Flatten(v any) map[string]any {
	out := make(map[string]any)
	flattenInto(out, "", v)
	return out
}

func flattenInto(dst map[string]any, prefix string, v any) {
	switch val := v.(type) {
	case map[string]any:
		for k, vv := range val {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			flattenInto(dst, key, vv)
		}
	case []any:
		for i, vv := range val {
			key := fmt.Sprintf("%s[%d]", prefix, i)
			flattenInto(dst, key, vv)
		}
	default:
		if prefix == "" {
			// root scalar
			dst["$"] = val
			return
		}
		dst[prefix] = val
	}
}

// toStringKeyMap converts a decoded YAML structure into map[string]any form
// to enable uniform flattening.
func toStringKeyMap(v any) any {
	switch vt := v.(type) {
	case map[any]any:
		m := make(map[string]any, len(vt))
		for k, val := range vt {
			m[fmt.Sprint(k)] = toStringKeyMap(val)
		}
		return m
	case map[string]any:
		m := make(map[string]any, len(vt))
		for k, val := range vt {
			m[k] = toStringKeyMap(val)
		}
		return m
	case []any:
		arr := make([]any, len(vt))
		for i := range vt {
			arr[i] = toStringKeyMap(vt[i])
		}
		return arr
	default:
		return vt
	}
}
