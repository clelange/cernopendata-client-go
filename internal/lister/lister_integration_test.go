//go:build integration

package lister

import (
	"context"
	"testing"
)

// These tests require access to real XRootD servers at CERN.
// Run with: go test -tags=integration ./internal/lister/...

func TestLister_GetFileSize_Integration(t *testing.T) {
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

func TestLister_DirectoryExists_Integration(t *testing.T) {
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
