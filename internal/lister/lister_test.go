package lister

import (
	"strings"
	"testing"
)

func TestNewLister(t *testing.T) {
	l := NewLister()
	if l == nil {
		t.Fatal("NewLister() returned nil")
	}
}

func TestFileInfo(t *testing.T) {
	tests := []struct {
		name   string
		file   FileInfo
		checks func(t *testing.T, fi FileInfo)
	}{
		{
			name: "file entry",
			file: FileInfo{
				Name:    "test.txt",
				Size:    1024,
				IsDir:   false,
				ModTime: "2025-01-17 00:00:00",
			},
			checks: func(t *testing.T, fi FileInfo) {
				if fi.Name != "test.txt" {
					t.Errorf("FileInfo.Name = %q, want %q", fi.Name, "test.txt")
				}
				if fi.Size != 1024 {
					t.Errorf("FileInfo.Size = %d, want %d", fi.Size, 1024)
				}
				if fi.IsDir {
					t.Errorf("FileInfo.IsDir = true, want false")
				}
				if fi.ModTime == "" {
					t.Error("FileInfo.ModTime is empty")
				}
			},
		},
		{
			name: "directory entry",
			file: FileInfo{
				Name:    "testdir",
				Size:    0,
				IsDir:   true,
				ModTime: "2025-01-17 00:00:00",
			},
			checks: func(t *testing.T, fi FileInfo) {
				if fi.Name != "testdir" {
					t.Errorf("FileInfo.Name = %q, want %q", fi.Name, "testdir")
				}
				if !fi.IsDir {
					t.Errorf("FileInfo.IsDir = false, want true")
				}
			},
		},
		{
			name: "zero-size file",
			file: FileInfo{
				Name:    "empty.txt",
				Size:    0,
				IsDir:   false,
				ModTime: "2025-01-17 00:00:00",
			},
			checks: func(t *testing.T, fi FileInfo) {
				if fi.Size != 0 {
					t.Errorf("FileInfo.Size = %d, want 0", fi.Size)
				}
			},
		},
		{
			name: "large file",
			file: FileInfo{
				Name:    "large.bin",
				Size:    1024 * 1024 * 1024,
				IsDir:   false,
				ModTime: "2025-01-17 00:00:00",
			},
			checks: func(t *testing.T, fi FileInfo) {
				if fi.Size != 1024*1024*1024 {
					t.Errorf("FileInfo.Size = %d, want %d", fi.Size, 1024*1024*1024)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.checks != nil {
				tt.checks(t, tt.file)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "path without protocol",
			input:    "/eos/opendata/cms",
			contains: "root://eospublic.cern.ch//eos/opendata/cms",
		},
		{
			name:     "path with root:// protocol",
			input:    "root://eospublic.cern.ch//eos/opendata/cms",
			contains: "root://eospublic.cern.ch//eos/opendata/cms",
		},
		{
			name:     "path with root:// and single slash",
			input:    "root://eospublic.cern.ch/eos/opendata/cms",
			contains: "root://eospublic.cern.ch/eos/opendata/cms",
		},
		{
			name:     "path without leading slash",
			input:    "eos/opendata/cms",
			contains: "root://eospublic.cern.ch//eos/opendata/cms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("normalizePath() = %v, want to contain %v", result, tt.contains)
			}
		})
	}
}
