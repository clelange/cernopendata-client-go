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

func TestParseQueryFromURL(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantQ      string
		wantFacets map[string]string
		wantPage   *int
		wantSize   *int
		wantSort   string
		wantErr    bool
	}{
		{
			name:       "empty input",
			input:      "",
			wantQ:      "",
			wantFacets: map[string]string{},
			wantErr:    false,
		},
		{
			name:       "simple query",
			input:      "q=Higgs",
			wantQ:      "Higgs",
			wantFacets: map[string]string{},
			wantErr:    false,
		},
		{
			name:       "query with facet",
			input:      "q=muon&f=experiment:CMS",
			wantQ:      "muon",
			wantFacets: map[string]string{"experiment": "CMS"},
			wantErr:    false,
		},
		{
			name:       "multiple facets",
			input:      "q=test&f=experiment:CMS&f=type:Dataset",
			wantQ:      "test",
			wantFacets: map[string]string{"experiment": "CMS", "type": "Dataset"},
			wantErr:    false,
		},
		{
			name:       "full URL",
			input:      "https://opendata.cern.ch/search?q=Higgs&f=experiment:ATLAS",
			wantQ:      "Higgs",
			wantFacets: map[string]string{"experiment": "ATLAS"},
			wantErr:    false,
		},
		{
			name:       "URL encoded",
			input:      "q=heavy%20ion&f=experiment%3ACMS",
			wantQ:      "heavy ion",
			wantFacets: map[string]string{"experiment": "CMS"},
			wantErr:    false,
		},
		{
			name:       "with page parameter",
			input:      "q=test&page=5",
			wantQ:      "test",
			wantFacets: map[string]string{},
			wantPage:   intPtr(5),
			wantErr:    false,
		},
		{
			name:       "with size parameter",
			input:      "q=test&size=20",
			wantQ:      "test",
			wantFacets: map[string]string{},
			wantSize:   intPtr(20),
			wantErr:    false,
		},
		{
			name:       "with sort parameter",
			input:      "q=test&sort=mostrecent",
			wantQ:      "test",
			wantFacets: map[string]string{},
			wantSort:   "mostrecent",
			wantErr:    false,
		},
		{
			name:       "with p parameter (alternate page)",
			input:      "q=test&p=3",
			wantQ:      "test",
			wantFacets: map[string]string{},
			wantPage:   intPtr(3),
			wantErr:    false,
		},
		{
			name:       "with s parameter (alternate size)",
			input:      "q=test&s=50",
			wantQ:      "test",
			wantFacets: map[string]string{},
			wantSize:   intPtr(50),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQueryFromURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQueryFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.Q != tt.wantQ {
				t.Errorf("ParseQueryFromURL().Q = %q, want %q", got.Q, tt.wantQ)
			}

			if len(got.Facets) != len(tt.wantFacets) {
				t.Errorf("ParseQueryFromURL().Facets has %d items, want %d", len(got.Facets), len(tt.wantFacets))
			}
			for k, v := range tt.wantFacets {
				if got.Facets[k] != v {
					t.Errorf("ParseQueryFromURL().Facets[%q] = %q, want %q", k, got.Facets[k], v)
				}
			}

			if tt.wantPage != nil {
				if got.Page == nil || *got.Page != *tt.wantPage {
					t.Errorf("ParseQueryFromURL().Page = %v, want %v", got.Page, *tt.wantPage)
				}
			}

			if tt.wantSize != nil {
				if got.Size == nil || *got.Size != *tt.wantSize {
					t.Errorf("ParseQueryFromURL().Size = %v, want %v", got.Size, *tt.wantSize)
				}
			}

			if got.Sort != tt.wantSort {
				t.Errorf("ParseQueryFromURL().Sort = %q, want %q", got.Sort, tt.wantSort)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
