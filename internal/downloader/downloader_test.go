package downloader

import (
	"testing"
)

func TestNewDownloader(t *testing.T) {
	d := NewDownloader()

	if d == nil {
		t.Fatal("NewDownloader() returned nil")
	}

	if d.client == nil {
		t.Error("client is nil")
	}

	if d.retryLimit != 10 {
		t.Errorf("retryLimit = %d, want 10", d.retryLimit)
	}

	if d.retrySleep != 5 {
		t.Errorf("retrySleep = %d, want 5", d.retrySleep)
	}
}

func TestFilterFiles(t *testing.T) {
	files := []interface{}{
		map[string]interface{}{"uri": "/path/file1.txt"},
		map[string]interface{}{"uri": "/path/file2.csv"},
		map[string]interface{}{"uri": "/path/file3.log"},
	}

	tests := []struct {
		name     string
		filter   string
		expected int
	}{
		{"no filter", "", 3},
		{"glob filter", "*.txt", 1},
		{"no match", "*.json", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterFiles(files, tt.filter)
			if len(result) != tt.expected {
				t.Errorf("FilterFiles() = %d, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestFilterFilesByRange(t *testing.T) {
	files := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	tests := []struct {
		name     string
		start    int
		end      int
		expected int
	}{
		{"all files", 0, 10, 10},
		{"first three", 0, 3, 3},
		{"middle", 4, 7, 3},
		{"last two", 8, 10, 2},
		{"out of range", 0, 20, 10},
		{"invalid range", 5, 3, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterFilesByRange(files, tt.start, tt.end)
			if len(result) != tt.expected {
				t.Errorf("FilterFilesByRange() = %d, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestDownloadStats(t *testing.T) {
	stats := DownloadStats{
		TotalFiles:      10,
		TotalBytes:      10000,
		DownloadedFiles: 8,
		DownloadedBytes: 8000,
		FailedFiles:     1,
		SkippedFiles:    1,
	}

	if stats.TotalFiles != 10 {
		t.Errorf("TotalFiles = %d, want 10", stats.TotalFiles)
	}

	if stats.DownloadedFiles != 8 {
		t.Errorf("DownloadedFiles = %d, want 8", stats.DownloadedFiles)
	}

	downloadedRatio := float64(stats.DownloadedFiles) / float64(stats.TotalFiles)
	if downloadedRatio < 0.7 || downloadedRatio > 0.9 {
		t.Errorf("DownloadedFiles ratio = %.2f, want ~0.8", downloadedRatio)
	}
}

func TestFileDownloadResult(t *testing.T) {
	result := &FileDownloadResult{
		URL:     "http://example.com/file.txt",
		Path:    "/tmp/file.txt",
		Size:    1024,
		Success: true,
		Retries: 2,
	}

	if result.URL != "http://example.com/file.txt" {
		t.Errorf("URL = %q, want %q", result.URL, "http://example.com/file.txt")
	}

	if result.Size != 1024 {
		t.Errorf("Size = %d, want 1024", result.Size)
	}

	if result.Retries != 2 {
		t.Errorf("Retries = %d, want 2", result.Retries)
	}
}
