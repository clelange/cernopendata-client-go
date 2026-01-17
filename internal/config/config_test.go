package config

import "testing"

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		constant string
	}{
		{
			name:     "ServerHTTPURI",
			got:      ServerHTTPURI,
			constant: "http://opendata.cern.ch",
		},
		{
			name:     "ServerHTTPSURI",
			got:      ServerHTTPSURI,
			constant: "https://opendata.cern.ch",
		},
		{
			name:     "ServerRootURI",
			got:      ServerRootURI,
			constant: "root://eospublic.cern.ch//",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.constant {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.constant)
			}
		})
	}
}
