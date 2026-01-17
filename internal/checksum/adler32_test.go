package checksum

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateChecksum(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "empty file",
			content:  "",
			expected: "adler32:00000001",
		},
		{
			name:     "simple text",
			content:  "test",
			expected: "adler32:045d01c1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.txt")

			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			result, err := CalculateChecksum(testFile)
			if err != nil {
				t.Errorf("CalculateChecksum() error = %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("CalculateChecksum() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetFileSize(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	testContent := []byte("test content for size check")
	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	size, err := GetFileSize(testFile)
	if err != nil {
		t.Errorf("GetFileSize() error = %v", err)
		return
	}

	expected := int64(len(testContent))
	if size != expected {
		t.Errorf("GetFileSize() = %d, want %d", size, expected)
	}
}
