package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func GetAssetForCurrentPlatform(release *ReleaseInfo) (binaryURL, checksumURL string, err error) {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	assetName := fmt.Sprintf("cernopendata-client-%s-%s", osName, archName)

	for _, asset := range release.Assets {
		if asset.Name == assetName {
			binaryURL = asset.BrowserDownloadURL
		}
		if asset.Name == "checksums.txt" {
			checksumURL = asset.BrowserDownloadURL
		}
	}

	if binaryURL == "" {
		return "", "", fmt.Errorf("no binary found for %s/%s", osName, archName)
	}

	return binaryURL, checksumURL, nil
}

func DownloadBinary(url string, progress func(downloaded, total int64)) ([]byte, error) {
	resp, err := http.Get(url) // #nosec G107
	if err != nil {
		return nil, fmt.Errorf("failed to download binary: %w", err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	var data []byte
	if progress != nil && resp.ContentLength > 0 {
		data = make([]byte, 0, resp.ContentLength)
		buf := make([]byte, 32*1024)
		var downloaded int64
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				data = append(data, buf[:n]...)
				downloaded += int64(n)
				progress(downloaded, resp.ContentLength)
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("failed to read response: %w", err)
			}
		}
	} else {
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read binary: %w", err)
		}
	}

	return data, nil
}

func FetchChecksums(url string) (map[string]string, error) {
	resp, err := http.Get(url) // #nosec G107
	if err != nil {
		return nil, fmt.Errorf("failed to download checksums: %w", err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("checksums download failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read checksums: %w", err)
	}

	checksums := make(map[string]string)
	for _, line := range strings.Split(string(body), "\n") {
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

func VerifyChecksum(data []byte, expectedChecksum string) error {
	hash := sha256.Sum256(data)
	actual := hex.EncodeToString(hash[:])

	if actual != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actual)
	}

	return nil
}

func ReplaceBinary(newBinary []byte) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	info, err := os.Stat(execPath)
	if err != nil {
		return fmt.Errorf("failed to stat executable: %w", err)
	}

	dir := filepath.Dir(execPath)
	tmpFile, err := os.CreateTemp(dir, "cernopendata-client-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	defer func() {
		if tmpPath != "" {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(newBinary); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to write new binary: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, info.Mode()); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	if err := os.Rename(tmpPath, execPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	tmpPath = ""

	return nil
}

func IsHomebrewInstall() bool {
	execPath, err := os.Executable()
	if err != nil {
		return false
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return false
	}

	homebrewPaths := []string{
		"/opt/homebrew/",
		"/usr/local/Cellar/",
		"/home/linuxbrew/",
		"/.linuxbrew/",
	}

	for _, prefix := range homebrewPaths {
		if strings.Contains(execPath, prefix) {
			return true
		}
	}

	return false
}
