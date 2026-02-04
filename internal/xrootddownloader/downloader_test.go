package xrootddownloader

import (
	"context"
	"testing"
)

func TestNewDownloader(t *testing.T) {
	d := NewDownloader()

	if d == nil {
		t.Fatal("NewDownloader returned nil")
	}

	if d.retryLimit != 10 {
		t.Errorf("Expected retryLimit 10, got %d", d.retryLimit)
	}

	if d.retrySleep != 5 {
		t.Errorf("Expected retrySleep 5, got %d", d.retrySleep)
	}

	if d.username != "gopher" {
		t.Errorf("Expected username 'gopher', got '%s'", d.username)
	}
}

func TestDownloadStats(t *testing.T) {
	stats := DownloadStats{
		TotalFiles:      10,
		TotalBytes:      1000000,
		DownloadedFiles: 8,
		DownloadedBytes: 800000,
		FailedFiles:     1,
		SkippedFiles:    1,
	}

	if stats.TotalFiles != 10 {
		t.Errorf("Expected TotalFiles 10, got %d", stats.TotalFiles)
	}

	if stats.DownloadedFiles != 8 {
		t.Errorf("Expected DownloadedFiles 8, got %d", stats.DownloadedFiles)
	}

	if stats.FailedFiles != 1 {
		t.Errorf("Expected FailedFiles 1, got %d", stats.FailedFiles)
	}

	if stats.SkippedFiles != 1 {
		t.Errorf("Expected SkippedFiles 1, got %d", stats.SkippedFiles)
	}
}

func TestFileDownloadResult(t *testing.T) {
	result := FileDownloadResult{
		URL:      "root://test/file.dat",
		Path:     "/tmp/file.dat",
		Size:     1024,
		Checksum: "abc123",
		Success:  true,
		Retries:  2,
	}

	if result.URL != "root://test/file.dat" {
		t.Errorf("Expected URL 'root://test/file.dat', got '%s'", result.URL)
	}

	if result.Path != "/tmp/file.dat" {
		t.Errorf("Expected Path '/tmp/file.dat', got '%s'", result.Path)
	}

	if result.Size != 1024 {
		t.Errorf("Expected Size 1024, got %d", result.Size)
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}

	if result.Retries != 2 {
		t.Errorf("Expected Retries 2, got %d", result.Retries)
	}
}

func TestClose(t *testing.T) {
	d := NewDownloader()

	err := d.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	err = d.Close()
	if err != nil {
		t.Errorf("Close after Close returned error: %v", err)
	}
}

func TestDownloadFilesDryRun(t *testing.T) {
	d := NewDownloader()
	d.dryRun = true
	d.verbose = true

	files := []interface{}{
		map[string]interface{}{
			"uri":      "root://test/file1.dat",
			"size":     float64(1000),
			"checksum": "abc123",
		},
		map[string]interface{}{
			"uri":      "root://test/file2.dat",
			"size":     float64(2000),
			"checksum": "def456",
		},
	}

	ctx := context.Background()
	stats := d.DownloadFiles(ctx, files, "/tmp/test", 3, 2, true, true, false)

	if stats.TotalFiles != 2 {
		t.Errorf("Expected TotalFiles 2, got %d", stats.TotalFiles)
	}

	if stats.DownloadedFiles != 2 {
		t.Errorf("Expected DownloadedFiles 2 (dry run), got %d", stats.DownloadedFiles)
	}

	if stats.FailedFiles != 0 {
		t.Errorf("Expected FailedFiles 0 (dry run), got %d", stats.FailedFiles)
	}

	if stats.SkippedFiles != 0 {
		t.Errorf("Expected SkippedFiles 0 (dry run), got %d", stats.SkippedFiles)
	}

	if stats.DownloadedBytes != 3000 {
		t.Errorf("Expected DownloadedBytes 3000, got %d", stats.DownloadedBytes)
	}
}

func TestDownloadFilesInvalidEntry(t *testing.T) {
	d := NewDownloader()

	files := []interface{}{
		map[string]interface{}{
			"uri":      "root://test/file1.dat",
			"size":     float64(1000),
			"checksum": "abc123",
		},
		"not a map",
		map[string]interface{}{
			"uri":      "root://test/file2.dat",
			"size":     float64(2000),
			"checksum": "def456",
		},
	}

	ctx := context.Background()
	stats := d.DownloadFiles(ctx, files, "/tmp/test", 3, 2, true, true, false)

	if stats.SkippedFiles != 1 {
		t.Errorf("Expected SkippedFiles 1, got %d", stats.SkippedFiles)
	}
}
