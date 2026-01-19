package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		latest   string
		expected int
	}{
		{
			name:     "dev vs release",
			current:  "dev",
			latest:   "v1.0.0",
			expected: -1,
		},
		{
			name:     "release vs dev",
			current:  "v1.0.0",
			latest:   "dev",
			expected: 1,
		},
		{
			name:     "equal versions",
			current:  "v1.0.0",
			latest:   "v1.0.0",
			expected: 0,
		},
		{
			name:     "older vs newer",
			current:  "v1.0.0",
			latest:   "v1.1.0",
			expected: -1,
		},
		{
			name:     "newer vs older",
			current:  "v1.1.0",
			latest:   "v1.0.0",
			expected: 1,
		},
		{
			name:     "major version difference",
			current:  "v1.0.0",
			latest:   "v2.0.0",
			expected: -1,
		},
		{
			name:     "patch version difference",
			current:  "v1.0.0",
			latest:   "v1.0.1",
			expected: -1,
		},
		{
			name:     "without v prefix",
			current:  "1.0.0",
			latest:   "1.1.0",
			expected: -1,
		},
		{
			name:     "with pre-release suffix",
			current:  "1.0.0-rc1",
			latest:   "1.0.0",
			expected: 0,
		},
		{
			name:     "dev prefix",
			current:  "dev-abc123",
			latest:   "v1.0.0",
			expected: -1,
		},
		{
			name:     "multi-part version",
			current:  "v1.2.3.4",
			latest:   "v1.2.3.5",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareVersions(tt.current, tt.latest)
			if result != tt.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, expected %d", tt.current, tt.latest, result, tt.expected)
			}
		})
	}
}

func TestFetchChecksums(t *testing.T) {
	testData := "a1b2c3d4  file1.txt\ne5f6g7h8  file2.bin\n"
	checksums, err := parseChecksums(testData)
	if err != nil {
		t.Fatalf("Failed to parse checksums: %v", err)
	}

	expectedChecksums := map[string]string{
		"file1.txt": "a1b2c3d4",
		"file2.bin": "e5f6g7h8",
	}

	if len(checksums) != len(expectedChecksums) {
		t.Errorf("Expected %d checksums, got %d", len(expectedChecksums), len(checksums))
	}

	for file, expected := range expectedChecksums {
		if actual, ok := checksums[file]; !ok || actual != expected {
			t.Errorf("Checksum for %s: expected %s, got %s", file, expected, actual)
		}
	}
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("test data")
	hash := sha256.Sum256(data)
	expectedChecksum := hex.EncodeToString(hash[:])

	tests := []struct {
		name             string
		data             []byte
		expectedChecksum string
		wantErr          bool
	}{
		{
			name:             "valid checksum",
			data:             data,
			expectedChecksum: expectedChecksum,
			wantErr:          false,
		},
		{
			name:             "invalid checksum",
			data:             data,
			expectedChecksum: "0000000000000000000000000000000000000000000000000000000000000000",
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyChecksum(tt.data, tt.expectedChecksum)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyChecksum() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsHomebrewInstall(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Apple Silicon Homebrew",
			path:     "/opt/homebrew/bin/cernopendata-client",
			expected: true,
		},
		{
			name:     "Intel Mac Homebrew",
			path:     "/usr/local/Cellar/cernopendata-client-go/1.0.0/bin/cernopendata-client",
			expected: true,
		},
		{
			name:     "Linux Homebrew",
			path:     "/home/linuxbrew/.linuxbrew/bin/cernopendata-client",
			expected: true,
		},
		{
			name:     "User-local Linux Homebrew",
			path:     "/.linuxbrew/bin/cernopendata-client",
			expected: true,
		},
		{
			name:     "usr/local/bin (not Homebrew)",
			path:     "/usr/local/bin/cernopendata-client",
			expected: false,
		},
		{
			name:     "home/bin (user directory)",
			path:     "/home/user/bin/cernopendata-client",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHomebrewPath(tt.path)
			if result != tt.expected {
				t.Errorf("isHomebrewPath(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestGetAssetForCurrentPlatform(t *testing.T) {
	tests := []struct {
		name        string
		release     ReleaseInfo
		wantBinary  string
		wantCheck   string
		wantErr     bool
		errContains string
	}{
		{
			name: "matching binary",
			release: ReleaseInfo{
				Assets: []ReleaseAsset{
					{Name: fmt.Sprintf("cernopendata-client-%s-%s", runtime.GOOS, runtime.GOARCH), BrowserDownloadURL: "https://example.com/binary"},
					{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt"},
				},
			},
			wantBinary: "https://example.com/binary",
			wantCheck:  "https://example.com/checksums.txt",
			wantErr:    false,
		},
		{
			name: "no matching binary",
			release: ReleaseInfo{
				Assets: []ReleaseAsset{
					{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt"},
				},
			},
			wantBinary:  "",
			wantCheck:   "https://example.com/checksums.txt",
			wantErr:     true,
			errContains: "no binary found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binaryURL, checksumURL, err := GetAssetForCurrentPlatform(&tt.release)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAssetForCurrentPlatform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errContains, err.Error())
				}
			}
			if binaryURL != tt.wantBinary {
				t.Errorf("GetAssetForCurrentPlatform() binaryURL = %v, want %v", binaryURL, tt.wantBinary)
			}
			if !tt.wantErr && checksumURL != tt.wantCheck {
				t.Errorf("GetAssetForCurrentPlatform() checksumURL = %v, want %v", checksumURL, tt.wantCheck)
			}
		})
	}
}

func parseChecksums(data string) (map[string]string, error) {
	checksums := make(map[string]string)
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			checksums[parts[1]] = parts[0]
		}
	}
	return checksums, nil
}

func isHomebrewPath(path string) bool {
	homebrewPaths := []string{
		"/opt/homebrew/",
		"/usr/local/Cellar/",
		"/home/linuxbrew/",
		"/.linuxbrew/",
	}

	for _, prefix := range homebrewPaths {
		if strings.Contains(path, prefix) {
			return true
		}
	}

	return false
}
