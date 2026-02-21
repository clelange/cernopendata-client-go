package metadater

import (
	"reflect"
	"testing"
)

func TestExtractNestedField(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		path     string
		expected any
		wantErr  bool
	}{
		{
			name:     "simple field",
			data:     map[string]any{"title": "Test Record"},
			path:     "title",
			expected: "Test Record",
			wantErr:  false,
		},
		{
			name:     "nested field",
			data:     map[string]any{"system": map[string]any{"tag": "v1"}},
			path:     "system.tag",
			expected: "v1",
			wantErr:  false,
		},
		{
			name:    "missing field",
			data:    map[string]any{"title": "Test"},
			path:    "missing",
			wantErr: true,
		},
		{
			name:     "empty path",
			data:     map[string]any{"title": "Test"},
			path:     "",
			expected: map[string]any{"title": "Test"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractNestedField(tt.data, tt.path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ExtractNestedField() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFilterArray(t *testing.T) {
	items := []any{
		map[string]any{"name": "Alice", "age": 30},
		map[string]any{"name": "Bob", "age": 25},
		map[string]any{"name": "Charlie", "age": 30},
	}

	tests := []struct {
		name     string
		filters  []string
		expected []any
		wantErr  bool
	}{
		{
			name:     "filter by name",
			filters:  []string{"name=Bob"},
			expected: []any{items[1]},
			wantErr:  false,
		},
		{
			name:     "filter by age",
			filters:  []string{"age=30"},
			expected: []any{items[0], items[2]},
			wantErr:  false,
		},
		{
			name:     "no filters",
			filters:  []string{},
			expected: items,
			wantErr:  false,
		},
		{
			name:    "invalid filter format",
			filters: []string{"invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FilterArray(items, tt.filters)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("FilterArray() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatOutput(t *testing.T) {
	data := map[string]any{"name": "Test", "value": 123}

	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{
			name:    "json format",
			format:  "json",
			wantErr: false,
		},
		{
			name:    "pretty format",
			format:  "pretty",
			wantErr: false,
		},
		{
			name:    "invalid format",
			format:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatOutput(data, tt.format)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == "" {
				t.Errorf("FormatOutput() returned empty string")
			}
		})
	}
}
