package utils

import (
	"testing"
)

func TestParseParameters(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    []string
		wantErr bool
	}{
		{"single parameter", []string{"file1.txt"}, []string{"file1.txt"}, false},
		{"multiple parameters", []string{"file1.txt", "file2.txt"}, []string{"file1.txt", "file2.txt"}, false},
		{"parameters with spaces", []string{"file1.txt , file2.txt"}, []string{"file1.txt ", " file2.txt"}, false},
		{"empty input", []string{}, []string{}, false},
		{"empty parameter", []string{"file1.txt", ""}, nil, true},
		{"whitespace only", []string{"   "}, nil, true},
		{"comma in input", []string{"file1.txt,file2.txt"}, []string{"file1.txt", "file2.txt"}, false},
		{"mixed separators", []string{"file1.txt file2.txt", "file3.txt"}, []string{"file1.txt file2.txt", "file3.txt"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseParameters(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseParameters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("ParseParameters() got %d items, want %d", len(got), len(tt.want))
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("ParseParameters()[%d] = %q, want %q", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestParseRanges(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    [][2]int
		wantErr bool
	}{
		{"single range", []string{"0-10"}, [][2]int{{0, 10}}, false},
		{"multiple ranges", []string{"1-2", "5-7"}, [][2]int{{1, 2}, {5, 7}}, false},
		{"comma separated ranges", []string{"1-2,5-7"}, [][2]int{{1, 2}, {5, 7}}, false},
		{"zero start", []string{"0-10"}, [][2]int{{0, 10}}, false},
		{"empty input", []string{}, [][2]int{}, false},
		{"no dash", []string{"1 10"}, nil, true},
		{"multiple dashes", []string{"1-2-3"}, nil, true},
		{"non-number start", []string{"a-10"}, nil, true},
		{"end before start", []string{"10-1"}, nil, true},
		{"negative start", []string{"-1-10"}, nil, true},
		{"same value", []string{"5-5"}, [][2]int{{5, 5}}, false},
		{"complex ranges", []string{"0-10", "20-30", "100-200"}, [][2]int{{0, 10}, {20, 30}, {100, 200}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRanges(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRanges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("ParseRanges() got %d items, want %d", len(got), len(tt.want))
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("ParseRanges()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}
