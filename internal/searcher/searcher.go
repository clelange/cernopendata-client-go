package searcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/clelange/cernopendata-client-go/internal/config"
)

type RecordResponse struct {
	Metadata RecordMetadata `json:"metadata"`
	ID       string         `json:"id"`
}

type RecordMetadata struct {
	Title         string        `json:"title"`
	RecID         int           `json:"recid,string"` // Parse as string, convert to int
	DOI           string        `json:"doi"`
	Files         []FileInfo    `json:"files"`
	FileIndices   []FileIndex   `json:"_file_indices"`
	SystemDetails SystemDetails `json:"system_details"`
	Usage         Usage         `json:"usage"`
}

type SystemDetails struct {
	ContainerImages []ContainerImage `json:"container_images"`
	GlobalTag       string           `json:"global_tag"`
	Release         string           `json:"release"`
}

type ContainerImage struct {
	Name     string `json:"name"`
	Registry string `json:"registry"`
}

type Usage struct {
	Description string      `json:"description"`
	Links       []UsageLink `json:"links"`
}

type UsageLink struct {
	Description string `json:"description"`
	URL         string `json:"url"`
}

type FileInfo struct {
	URI          string `json:"uri"`
	Size         int64  `json:"size"`
	Checksum     string `json:"checksum"`
	Availability string `json:"availability,omitempty"` // "online" or "on demand"
}

type FileIndex struct {
	Key      string          `json:"key"`
	Size     int64           `json:"size"`
	Checksum string          `json:"checksum"`
	Files    []InnerFileInfo `json:"files"`
}

type InnerFileInfo struct {
	URI          string `json:"uri"`
	Size         int64  `json:"size"`
	Checksum     string `json:"checksum"`
	Availability string `json:"availability"`
}

type SearchResponse struct {
	Hits         SearchHits             `json:"hits"`
	Aggregations map[string]Aggregation `json:"aggregations"`
}

// Aggregation represents a facet with its buckets of values
type Aggregation struct {
	Buckets []AggregationBucket `json:"buckets"`
}

// AggregationBucket represents a single facet value with its count
type AggregationBucket struct {
	Key      interface{} `json:"key"`
	DocCount int         `json:"doc_count"`
}

type SearchHits struct {
	Total int         `json:"total"`
	Hits  []SearchHit `json:"hits"`
}

type SearchHit struct {
	ID       string                 `json:"id"`
	Metadata map[string]interface{} `json:"metadata"`
}

type Client struct {
	server string
	client *http.Client
}

func NewClient(server string) *Client {
	return &Client{
		server: server,
		client: &http.Client{},
	}
}

func (c *Client) GetRecord(recid int) (*RecordResponse, error) {
	url := fmt.Sprintf("%s/api/records/%d", c.server, recid)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get record %d: %w", recid, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d for record %d", resp.StatusCode, recid)
	}

	var recordResp RecordResponse
	if err := json.NewDecoder(resp.Body).Decode(&recordResp); err != nil {
		return nil, fmt.Errorf("failed to decode record response: %w", err)
	}

	return &recordResp, nil
}

func (c *Client) GetRecordByDOI(doi string) (*RecordResponse, error) {
	return c.getRecordBySearch("doi", doi)
}

func (c *Client) GetRecordByTitle(title string) (*RecordResponse, error) {
	return c.getRecordBySearch("title", title)
}

func (c *Client) getRecordBySearch(field, value string) (*RecordResponse, error) {
	searchURL := fmt.Sprintf("%s/api/records?page=1&size=1&q=%s:%s",
		c.server,
		field,
		url.QueryEscape(fmt.Sprintf("\"%s\"", value)),
	)

	resp, err := c.client.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search for %s: %w", field, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d for search", resp.StatusCode)
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	if searchResp.Hits.Total == 0 {
		return nil, fmt.Errorf("no record found with %s: %s", field, value)
	}

	if searchResp.Hits.Total > 1 {
		return nil, fmt.Errorf("more than one record found with %s: %s", field, value)
	}

	return c.GetRecordByID(searchResp.Hits.Hits[0].ID)
}

func (c *Client) GetRecordByID(id string) (*RecordResponse, error) {
	recordID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid record ID: %w", err)
	}
	return c.GetRecord(recordID)
}

// convertURI transforms a URI based on protocol settings
func convertURI(uri, serverRoot, serverURI, protocol string) string {
	if !strings.HasPrefix(uri, serverRoot) {
		return uri
	}
	switch protocol {
	case "http":
		return strings.Replace(uri, serverRoot, serverURI+"/", 1)
	case "https":
		return strings.Replace(uri, serverRoot, config.ServerHTTPSURI+"/", 1)
	default:
		return uri
	}
}

func (c *Client) GetFilesList(record *RecordResponse, protocol string, expand bool) []FileInfo {
	var files []FileInfo

	serverRoot := config.ServerRootURI
	serverURI := c.server

	for _, file := range record.Metadata.Files {
		files = append(files, FileInfo{
			URI:          convertURI(file.URI, serverRoot, serverURI, protocol),
			Size:         file.Size,
			Checksum:     file.Checksum,
			Availability: "online", // Direct files are always online
		})
	}

	if expand {
		for _, index := range record.Metadata.FileIndices {
			for _, innerFile := range index.Files {
				files = append(files, FileInfo{
					URI:          convertURI(innerFile.URI, serverRoot, serverURI, protocol),
					Size:         innerFile.Size,
					Checksum:     innerFile.Checksum,
					Availability: innerFile.Availability,
				})
			}
		}
	} else {
		for _, index := range record.Metadata.FileIndices {
			var uri string
			if protocol == "xrootd" {
				uri = fmt.Sprintf("%s/record/%d/file_index/%s", serverRoot, record.Metadata.RecID, index.Key)
			} else {
				uri = fmt.Sprintf("%s/record/%d/file_index/%s", serverURI, record.Metadata.RecID, index.Key)
			}
			files = append(files, FileInfo{
				URI:      uri,
				Size:     index.Size,
				Checksum: "",
			})
		}
	}

	return files
}

// FilterFilesByAvailability filters files by their availability status.
// If availability is empty/nil, returns all files (default behavior).
// Returns (filtered files, has offline files warning).
func FilterFilesByAvailability(files []FileInfo, availability string) ([]FileInfo, bool) {
	hasOfflineFiles := false
	for _, f := range files {
		if f.Availability != "" && f.Availability != "online" {
			hasOfflineFiles = true
			break
		}
	}

	if availability == "" {
		return files, hasOfflineFiles
	}

	if availability == "online" {
		var filtered []FileInfo
		for _, f := range files {
			if f.Availability == "online" {
				filtered = append(filtered, f)
			}
		}
		return filtered, hasOfflineFiles
	}

	// For "all" or any other value, return all files
	return files, hasOfflineFiles
}

func GetRecid(server, doi string, title string, recid int) (int, error) {
	if recid > 0 {
		return recid, nil
	}

	client := NewClient(server)

	if doi != "" {
		record, err := client.GetRecordByDOI(doi)
		if err != nil {
			return 0, err
		}
		return record.Metadata.RecID, nil
	}

	if title != "" {
		record, err := client.GetRecordByTitle(title)
		if err != nil {
			return 0, err
		}
		return record.Metadata.RecID, nil
	}

	return 0, fmt.Errorf("please provide recid, doi, or title")
}

// SearchRecords searches for records using a query string and optional facets.
// page and size control pagination, sort controls ordering.
// Returns records with metadata included.
func (c *Client) SearchRecords(q string, facets map[string]string, page, size int, sort string) (*SearchResponse, error) {
	params := url.Values{}
	if q != "" {
		params.Set("q", q)
	}
	for key, value := range facets {
		params.Add("f", fmt.Sprintf("%s:%s", key, value))
	}
	params.Set("page", strconv.Itoa(page))
	params.Set("size", strconv.Itoa(size))
	if sort != "" {
		params.Set("sort", sort)
	}
	params.Set("skip_files", "1")

	searchURL := fmt.Sprintf("%s/api/records/?%s", c.server, params.Encode())

	resp, err := c.client.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search records: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d for search", resp.StatusCode)
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return &searchResp, nil
}

// SearchAllRecords fetches all matching records by paginating through results
// in batches of 50. Returns a combined SearchResponse with all hits.
func (c *Client) SearchAllRecords(q string, facets map[string]string, sort string) (*SearchResponse, error) {
	const batchSize = 50
	var allHits []SearchHit
	page := 1
	totalRecords := 0

	for {
		resp, err := c.SearchRecords(q, facets, page, batchSize, sort)
		if err != nil {
			return nil, err
		}

		if page == 1 {
			totalRecords = resp.Hits.Total
		}

		allHits = append(allHits, resp.Hits.Hits...)

		// Check if we've fetched all records
		if len(allHits) >= totalRecords {
			break
		}

		page++
	}

	return &SearchResponse{
		Hits: SearchHits{
			Total: totalRecords,
			Hits:  allHits,
		},
	}, nil
}

// GetFacets fetches available facets (aggregations) from the API.
// This makes a minimal search request to get the aggregation data.
func (c *Client) GetFacets() (map[string]Aggregation, error) {
	// Make a minimal search to get aggregations
	resp, err := c.SearchRecords("", nil, 1, 1, "")
	if err != nil {
		return nil, err
	}
	return resp.Aggregations, nil
}
