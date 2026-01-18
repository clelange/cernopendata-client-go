package lister

import (
	"context"
	"strings"
	"testing"
)

func TestNewLister(t *testing.T) {
	l := NewLister()
	if l == nil {
		t.Fatal("NewLister() returned nil")
	}
}

func TestLister_GetFileSize(t *testing.T) {
	ctx := context.Background()
	l := NewLister()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid file path",
			path:    "/eos/opendata/cms/validated-runs/Commissioning10/Commissioning10-May19ReReco_7TeV.json",
			wantErr: false,
		},
		{
			name:    "non-existent file",
			path:    "/eos/opendata/cms/nonexistent.txt",
			wantErr: true,
		},
		{
			name:    "path with root:// prefix",
			path:    "root://eospublic.cern.ch//eos/opendata/cms/validated-runs/Commissioning10/Commissioning10-May19ReReco_7TeV.json",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := l.GetFileSize(ctx, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFileSize() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && size < 0 {
				t.Errorf("GetFileSize() size = %d, want non-negative", size)
			}
		})
	}
}

func TestLister_DirectoryExists(t *testing.T) {
	ctx := context.Background()
	l := NewLister()

	tests := []struct {
		name    string
		path    string
		want    bool
		wantErr bool
	}{
		{
			name:    "existing directory",
			path:    "/eos/opendata/cms",
			want:    true,
			wantErr: false,
		},
		{
			name:    "non-existing directory",
			path:    "/eos/opendata/cms/nonexistent",
			want:    false,
			wantErr: false,
		},
		{
			name:    "path with root:// prefix",
			path:    "root://eospublic.cern.ch//eos/opendata/cms",
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := l.DirectoryExists(ctx, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("DirectoryExists() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && exists != tt.want {
				t.Errorf("DirectoryExists() = %v, want %v", exists, tt.want)
			}
		})
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
