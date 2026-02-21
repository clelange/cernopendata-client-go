package metadater

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func ExtractNestedField(data any, path string) (any, error) {
	if path == "" {
		return data, nil
	}

	fields := strings.Split(path, ".")
	current := data

	for i, field := range fields {
		val := reflect.ValueOf(current)
		if val.Kind() == reflect.Map {
			mapVal, ok := val.Interface().(map[string]any)
			if !ok {
				return nil, fmt.Errorf("expected map[string]interface{}, got %T", val.Interface())
			}
			if v, ok := mapVal[field]; ok {
				// If this is the last field, return the value directly
				if i == len(fields)-1 {
					return v, nil
				}
				// If we have more fields and the value is a slice, process it immediately
				if val := reflect.ValueOf(v); val.Kind() == reflect.Slice && i < len(fields)-1 {
					sliceVal := v.([]any)
					var results []any
					for _, item := range sliceVal {
						itemMap, ok := item.(map[string]any)
						if !ok {
							return nil, fmt.Errorf("array item is not a map")
						}
						// Recursive call to extract nested fields from array items
						remainingPath := strings.Join(fields[i+1:], ".")
						nested, err := ExtractNestedField(itemMap, remainingPath)
						if err == nil && nested != nil {
							results = append(results, nested)
						}
					}
					return results, nil
				}
				current = v
			} else {
				return nil, fmt.Errorf("field %s not found", field)
			}
		} else if val.Kind() == reflect.Slice {
			// We have an array - return it directly
			// This shouldn't happen if we're still in the loop with remaining fields
			return current, nil
		} else {
			return nil, fmt.Errorf("cannot access field %s on non-map type", field)
		}
	}

	return current, nil
}

func GetNestedField(record any, path string) (any, error) {
	if path == "" {
		return record, nil
	}

	// Record is now a map, delegate to ExtractNestedField
	if recordMap, ok := record.(map[string]any); ok {
		return ExtractNestedField(recordMap, path)
	}

	// Fallback: try to convert to map (backward compatibility)
	var recordMap map[string]any
	jsonBytes, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record: %w", err)
	}
	err = json.Unmarshal(jsonBytes, &recordMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal record: %w", err)
	}

	return ExtractNestedField(recordMap, path)
}

func FilterArray(items []any, filters []string) ([]any, error) {
	if len(items) == 0 {
		return items, nil
	}

	if len(filters) == 0 {
		return items, nil
	}

	var result []any

	filterMap := make(map[string]string)
	for _, f := range filters {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid filter format: %s", f)
		}
		filterMap[parts[0]] = parts[1]
	}

	for _, item := range items {
		itemMap, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("item is not a map")
		}

		match := true
		for field, value := range filterMap {
			if itemValue, exists := itemMap[field]; exists {
				itemStr := fmt.Sprintf("%v", itemValue)
				if itemStr != value {
					match = false
					break
				}
			}
		}

		if match {
			result = append(result, item)
		}
	}

	return result, nil
}

func FormatOutput(data any, format string) (string, error) {
	switch format {
	case "json":
		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", err
		}
		return string(jsonBytes), nil
	case "pretty":
		return fmt.Sprintf("%+v\n", data), nil
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}
}
