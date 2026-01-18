package verifier

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyLocalFiles(t *testing.T) {
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.txt")

	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0600); err != nil {
		t.Fatal(err)
	}

	verifier := NewVerifier()
	stats, err := verifier.VerifyLocalFiles(testDir)

	if err != nil {
		t.Fatalf("VerifyLocalFiles failed: %v", err)
	}

	if stats.TotalFiles != 1 {
		t.Errorf("Expected 1 file, got %d", stats.TotalFiles)
	}

	if stats.VerifiedFiles != 1 {
		t.Errorf("Expected 1 verified file, got %d", stats.VerifiedFiles)
	}
}

func TestVerifyFiles(t *testing.T) {
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.txt")

	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0600); err != nil {
		t.Fatal(err)
	}

	verifier := NewVerifier()
	expectedFiles := []interface{}{
		map[string]interface{}{
			"uri":      "http://example.com/test.txt",
			"size":     float64(len(content)),
			"checksum": "adler32:08879620",
		},
	}

	stats, err := verifier.VerifyFiles(testDir, expectedFiles)

	if err != nil {
		t.Fatalf("VerifyFiles failed: %v", err)
	}

	if stats.TotalFiles != 1 {
		t.Errorf("Expected 1 file, got %d", stats.TotalFiles)
	}

	if stats.MissingFiles != 0 {
		t.Errorf("Expected 0 missing files, got %d", stats.MissingFiles)
	}
}

func TestParseChecksumMetadata(t *testing.T) {
	tests := []struct {
		name         string
		metadata     string
		expectedSum  string
		expectedSize int64
		expectError  bool
	}{
		{
			name:         "valid metadata",
			metadata:     "adler32:08879620\t12",
			expectedSum:  "adler32:08879620",
			expectedSize: 12,
			expectError:  false,
		},
		{
			name:         "invalid format",
			metadata:     "invalid",
			expectedSum:  "",
			expectedSize: 0,
			expectError:  true,
		},
		{
			name:         "invalid size",
			metadata:     "adler32:08879620\tinvalid",
			expectedSum:  "",
			expectedSize: 0,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sum, size, err := ParseChecksumMetadata(tt.metadata)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if sum != tt.expectedSum {
				t.Errorf("Expected sum %s, got %s", tt.expectedSum, sum)
			}

			if size != tt.expectedSize {
				t.Errorf("Expected size %d, got %d", tt.expectedSize, size)
			}
		})
	}
}
