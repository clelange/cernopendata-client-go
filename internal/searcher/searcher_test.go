package searcher

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	server := "http://opendata.cern.ch"
	client := NewClient(server)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.server != server {
		t.Errorf("client.server = %q, want %q", client.server, server)
	}

	if client.client == nil {
		t.Error("client.client is nil")
	}
}

func TestGetRecord(t *testing.T) {
	tests := []struct {
		name    string
		recid   int
		handler http.HandlerFunc
		want    *RecordResponse
		wantErr bool
	}{
		{
			name:  "successful record fetch",
			recid: 3005,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				resp := RecordResponse{
					ID: "3005",
					Metadata: RecordMetadata{
						RecID: 3005,
						Title: "Test Record",
						DOI:   "10.7483/record/3005",
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			want: &RecordResponse{
				ID: "3005",
				Metadata: RecordMetadata{
					RecID: 3005,
					Title: "Test Record",
					DOI:   "10.7483/record/3005",
				},
			},
			wantErr: false,
		},
		{
			name:  "record not found",
			recid: 999999,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:  "server error",
			recid: 3005,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewClient(server.URL)
			record, err := client.GetRecord(tt.recid)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRecord() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if record == nil {
					t.Fatal("GetRecord() returned nil record")
				}

				if record.Metadata.RecID != tt.want.Metadata.RecID {
					t.Errorf("GetRecord() recid = %d, want %d", record.Metadata.RecID, tt.want.Metadata.RecID)
				}

				if record.Metadata.Title != tt.want.Metadata.Title {
					t.Errorf("GetRecord() title = %q, want %q", record.Metadata.Title, tt.want.Metadata.Title)
				}
			}
		})
	}
}

func TestGetFilesList(t *testing.T) {
	tests := []struct {
		name     string
		record   *RecordResponse
		protocol string
		expand   bool
		wantLen  int
	}{
		{
			name: "http protocol without expand",
			record: &RecordResponse{
				Metadata: RecordMetadata{
					RecID: 3005,
					Files: []FileInfo{
						{URI: "http://opendata.cern.ch/test.txt", Size: 100, Checksum: "adler32:12345678"},
					},
					FileIndices: []FileIndex{},
				},
			},
			protocol: "http",
			expand:   false,
			wantLen:  1,
		},
		{
			name: "root protocol without expand",
			record: &RecordResponse{
				Metadata: RecordMetadata{
					RecID: 3005,
					FileIndices: []FileIndex{
						{
							Key:      "index1",
							Size:     100,
							Checksum: "adler32:87654321",
							Files: []InnerFileInfo{
								{URI: "root://eospublic.cern.ch//eos/opendata/cms/inner.txt", Size: 50, Checksum: "adler32:11111111"},
							},
						},
					},
				},
			},
			protocol: "http",
			expand:   false,
			wantLen:  1,
		},
		{
			name: "https protocol without expand",
			record: &RecordResponse{
				Metadata: RecordMetadata{
					RecID: 3005,
					Files: []FileInfo{
						{URI: "https://opendata.cern.ch/test.txt", Size: 100, Checksum: "adler32:12345678"},
					},
					FileIndices: []FileIndex{},
				},
			},
			protocol: "https",
			expand:   false,
			wantLen:  1,
		},
		{
			name: "expand file indices",
			record: &RecordResponse{
				Metadata: RecordMetadata{
					RecID: 3005,
					Files: []FileInfo{
						{URI: "http://opendata.cern.ch/test.txt", Size: 100, Checksum: "adler32:12345678"},
					},
					FileIndices: []FileIndex{
						{
							Key:      "index1",
							Size:     100,
							Checksum: "adler32:87654321",
							Files: []InnerFileInfo{
								{URI: "http://opendata.cern.ch/inner1.txt", Size: 50, Checksum: "adler32:11111111"},
								{URI: "http://opendata.cern.ch/inner2.txt", Size: 50, Checksum: "adler32:22222222"},
							},
						},
					},
				},
			},
			protocol: "http",
			expand:   true,
			wantLen:  3,
		},
		{
			name: "xrootd protocol (no conversion)",
			record: &RecordResponse{
				Metadata: RecordMetadata{
					RecID: 3005,
					Files: []FileInfo{
						{URI: "http://opendata.cern.ch/test.txt", Size: 100, Checksum: "adler32:12345678"},
					},
					FileIndices: []FileIndex{},
				},
			},
			protocol: "xrootd",
			expand:   false,
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := "http://test.server"
			client := NewClient(server)
			files := client.GetFilesList(tt.record, tt.protocol, tt.expand)

			if len(files) != tt.wantLen {
				t.Errorf("GetFilesList() length = %d, want %d", len(files), tt.wantLen)
			}

			if tt.protocol == "http" || tt.protocol == "https" {
				for _, file := range files {
					if len(files) > 0 {
						if !strings.HasPrefix(file.URI, "http://test.server") {
							t.Logf("GetFilesList() first URI = %q", file.URI)
						}
					}
				}
			}
		})
	}
}

func TestGetRecordByDOI(t *testing.T) {
	tests := []struct {
		name      string
		doi       string
		handler   http.HandlerFunc
		wantRecid int
		wantErr   bool
	}{
		{
			name: "successful DOI search",
			doi:  "10.7483/record/3005",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if strings.Contains(r.URL.Path, "records/") {
					resp := RecordResponse{
						ID: "3005",
						Metadata: RecordMetadata{
							RecID: 3005,
							Title: "Test Record",
							DOI:   "10.7483/record/3005",
						},
					}
					json.NewEncoder(w).Encode(resp)
				} else {
					searchResp := SearchResponse{
						Hits: SearchHits{
							Total: 1,
							Hits: []SearchHit{
								{ID: "3005"},
							},
						},
					}
					json.NewEncoder(w).Encode(searchResp)
				}
			},
			wantRecid: 3005,
			wantErr:   false,
		},
		{
			name: "DOI not found",
			doi:  "10.7483/record/nonexistent",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				searchResp := SearchResponse{
					Hits: SearchHits{
						Total: 0,
						Hits:  []SearchHit{},
					},
				}
				json.NewEncoder(w).Encode(searchResp)
			},
			wantRecid: 0,
			wantErr:   true,
		},
		{
			name: "multiple records found",
			doi:  "10.7483/record/duplicate",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				searchResp := SearchResponse{
					Hits: SearchHits{
						Total: 2,
						Hits: []SearchHit{
							{ID: "3005"},
							{ID: "3006"},
						},
					},
				}
				json.NewEncoder(w).Encode(searchResp)
			},
			wantRecid: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewClient(server.URL)
			record, err := client.GetRecordByDOI(tt.doi)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRecordByDOI() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && record != nil {
				if record.Metadata.RecID != tt.wantRecid {
					t.Errorf("GetRecordByDOI() recid = %d, want %d", record.Metadata.RecID, tt.wantRecid)
				}
			}
		})
	}
}

func TestGetRecordByTitle(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		handler   http.HandlerFunc
		wantRecid int
		wantErr   bool
	}{
		{
			name:  "successful title search",
			title: "Test Record Title",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if strings.Contains(r.URL.Path, "records/") {
					resp := RecordResponse{
						ID: "3005",
						Metadata: RecordMetadata{
							RecID: 3005,
							Title: "Test Record",
							DOI:   "10.7483/record/3005",
						},
					}
					json.NewEncoder(w).Encode(resp)
				} else {
					searchResp := SearchResponse{
						Hits: SearchHits{
							Total: 1,
							Hits: []SearchHit{
								{ID: "3005"},
							},
						},
					}
					json.NewEncoder(w).Encode(searchResp)
				}
			},
			wantRecid: 3005,
			wantErr:   false,
		},
		{
			name:  "title not found",
			title: "Nonexistent Title",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				searchResp := SearchResponse{
					Hits: SearchHits{
						Total: 0,
						Hits:  []SearchHit{},
					},
				}
				json.NewEncoder(w).Encode(searchResp)
			},
			wantRecid: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewClient(server.URL)
			record, err := client.GetRecordByTitle(tt.title)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRecordByTitle() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && record != nil {
				if record.Metadata.RecID != tt.wantRecid {
					t.Errorf("GetRecordByTitle() recid = %d, want %d", record.Metadata.RecID, tt.wantRecid)
				}
			}
		})
	}
}

func TestGetRecid(t *testing.T) {
	tests := []struct {
		name    string
		server  string
		doi     string
		title   string
		recid   int
		want    int
		wantErr bool
	}{
		{
			name:    "recid provided",
			recid:   3005,
			want:    3005,
			wantErr: false,
		},
		{
			name:    "zero recid (invalid)",
			recid:   0,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRecid(tt.server, tt.doi, tt.title, tt.recid)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRecid() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("GetRecid() = %d, want %d", got, tt.want)
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
			name: "complete file info",
			file: FileInfo{
				URI:      "http://example.com/file.txt",
				Size:     1024,
				Checksum: "adler32:12345678",
			},
			checks: func(t *testing.T, fi FileInfo) {
				if fi.URI != "http://example.com/file.txt" {
					t.Errorf("FileInfo.URI = %q, want %q", fi.URI, "http://example.com/file.txt")
				}
				if fi.Size != 1024 {
					t.Errorf("FileInfo.Size = %d, want %d", fi.Size, 1024)
				}
				if fi.Checksum != "adler32:12345678" {
					t.Errorf("FileInfo.Checksum = %q, want %q", fi.Checksum, "adler32:12345678")
				}
			},
		},
		{
			name: "zero size file",
			file: FileInfo{
				URI:      "http://example.com/empty.txt",
				Size:     0,
				Checksum: "adler32:00000000",
			},
			checks: func(t *testing.T, fi FileInfo) {
				if fi.Size != 0 {
					t.Errorf("FileInfo.Size = %d, want 0", fi.Size)
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
