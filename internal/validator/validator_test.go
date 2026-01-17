package validator

import (
	"testing"
)

func TestValidateRecid(t *testing.T) {
	tests := []struct {
		name    string
		recid   int
		wantErr bool
	}{
		{"valid recid", 3005, false},
		{"valid recid large", 999999, false},
		{"invalid recid zero", 0, true},
		{"invalid recid negative", -1, true},
		{"invalid recid negative large", -100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRecid(tt.recid)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRecid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateServer(t *testing.T) {
	tests := []struct {
		name    string
		server  string
		wantErr bool
	}{
		{"valid http", "http://opendata.cern.ch", false},
		{"valid https", "https://opendata.cern.ch", false},
		{"valid http with port", "http://opendata.cern.ch:80", false},
		{"valid https with port", "https://opendata.cern.ch:443", false},
		{"invalid ftp", "ftp://opendata.cern.ch", true},
		{"invalid root", "root://eospublic.cern.ch/", true},
		{"invalid no scheme", "opendata.cern.ch", true},
		{"invalid empty", "", true},
		{"invalid malformed", "://opendata.cern.ch", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServer(tt.server)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRange(t *testing.T) {
	tests := []struct {
		name      string
		rangeStr  string
		count     int
		wantErr   bool
		wantStart int
		wantEnd   int
	}{
		{"valid range", "1-10", 10, false, 1, 10},
		{"valid range at boundary", "1-100", 100, false, 1, 100},
		{"valid range single", "5-5", 10, false, 5, 5},
		{"invalid empty", "", 10, true, 0, 0},
		{"invalid no dash", "1 10", 10, true, 0, 0},
		{"invalid multiple dashes", "1-2-3", 10, true, 0, 0},
		{"invalid start not number", "a-10", 10, true, 0, 0},
		{"invalid end not number", "1-a", 10, true, 0, 0},
		{"invalid start zero", "0-10", 10, true, 0, 0},
		{"invalid start negative", "-1-10", 10, true, 0, 0},
		{"invalid range too big", "1-100", 50, true, 0, 0},
		{"invalid end before start", "10-1", 10, true, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := ValidateRange(tt.rangeStr, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRange() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if start != tt.wantStart {
					t.Errorf("ValidateRange() start = %d, want %d", start, tt.wantStart)
				}
				if end != tt.wantEnd {
					t.Errorf("ValidateRange() end = %d, want %d", end, tt.wantEnd)
				}
			}
		})
	}
}

func TestValidateDirectory(t *testing.T) {
	tests := []struct {
		name      string
		directory string
		wantErr   bool
	}{
		{"valid cms path", "/eos/opendata/cms", false},
		{"valid atlas path", "/eos/opendata/atlas", false},
		{"valid deep path", "/eos/opendata/cms/Run2010B/BTau/AOD", false},
		{"invalid missing eos", "/opendata/cms", true},
		{"invalid missing opendata", "/eos/cms", true},
		{"invalid empty", "", true},
		{"invalid relative path", "eos/opendata/cms", true},
		{"invalid http path", "http://eospublic.cern.ch//eos/opendata/cms", true},
		{"invalid root path", "root://eospublic.cern.ch//eos/opendata/cms", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDirectory(tt.directory)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDirectory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRetryLimit(t *testing.T) {
	tests := []struct {
		name       string
		retryLimit int
		wantErr    bool
	}{
		{"valid retry", 10, false},
		{"valid retry large", 100, false},
		{"invalid retry zero", 0, true},
		{"invalid retry negative", -1, true},
		{"invalid retry negative large", -10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRetryLimit(tt.retryLimit)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRetryLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRetrySleep(t *testing.T) {
	tests := []struct {
		name       string
		retrySleep int
		wantErr    bool
	}{
		{"valid sleep", 5, false},
		{"valid sleep large", 60, false},
		{"invalid sleep zero", 0, true},
		{"invalid sleep negative", -1, true},
		{"invalid sleep negative large", -10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRetrySleep(tt.retrySleep)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRetrySleep() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
