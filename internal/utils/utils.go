package utils

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ParsedQuery represents a parsed search query from a URL or query string.
type ParsedQuery struct {
	Q      string
	Facets map[string]string
	Page   *int
	Size   *int
	Sort   string
}

// ParseQueryFromURL parses a full URL or query string from the CERN Open Data portal.
// It handles both full URLs (https://opendata.cern.ch/search?...) and query strings (q=Higgs&f=...).
// Facets are parsed from the 'f' parameter in format 'key:value'.
func ParseQueryFromURL(input string) (*ParsedQuery, error) {
	if input == "" {
		return &ParsedQuery{
			Facets: make(map[string]string),
		}, nil
	}

	var queryString string

	// Check if it's a full URL or just a query string
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		parsedURL, err := url.Parse(input)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}
		queryString = parsedURL.RawQuery
	} else {
		// Assume it's a query string
		queryString = input
	}

	values, err := url.ParseQuery(queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query string: %w", err)
	}

	result := &ParsedQuery{
		Facets: make(map[string]string),
	}

	// Parse 'q' parameter
	if q := values.Get("q"); q != "" {
		result.Q = q
	}

	// Parse 'f' parameters (facets in format key:value)
	for _, f := range values["f"] {
		parts := strings.SplitN(f, ":", 2)
		if len(parts) == 2 {
			result.Facets[parts[0]] = parts[1]
		}
	}

	// Parse 'page' or 'p' parameter
	if p := values.Get("page"); p != "" {
		if pageNum, err := strconv.Atoi(p); err == nil {
			result.Page = &pageNum
		}
	} else if p := values.Get("p"); p != "" {
		if pageNum, err := strconv.Atoi(p); err == nil {
			result.Page = &pageNum
		}
	}

	// Parse 'size' or 's' parameter
	if s := values.Get("size"); s != "" {
		if sizeNum, err := strconv.Atoi(s); err == nil {
			result.Size = &sizeNum
		}
	} else if s := values.Get("s"); s != "" {
		if sizeNum, err := strconv.Atoi(s); err == nil {
			result.Size = &sizeNum
		}
	}

	// Parse 'sort' parameter
	if sort := values.Get("sort"); sort != "" {
		result.Sort = sort
	}

	return result, nil
}

func ParseParameters(input []string) ([]string, error) {
	if len(input) == 0 {
		return []string{}, nil
	}

	joined := strings.Join(input, ",")
	result := strings.Split(joined, ",")

	for _, item := range result {
		if strings.TrimSpace(item) == "" {
			return nil, errors.New("empty parameter found")
		}
	}

	return result, nil
}

func ParseRanges(input []string) ([][2]int, error) {
	if len(input) == 0 {
		return [][2]int{}, nil
	}

	joined := strings.Join(input, ",")
	rangeStrings := strings.Split(joined, ",")

	var ranges [][2]int

	for _, rangeStr := range rangeStrings {
		parts := strings.Split(rangeStr, "-")
		if len(parts) != 2 {
			return nil, errors.New("invalid range format, expected 'i-j'")
		}

		var start, end int
		_, err := fmt.Sscanf(rangeStr, "%d-%d", &start, &end)
		if err != nil {
			return nil, errors.New("invalid range format: " + rangeStr)
		}

		if start < 0 {
			return nil, errors.New("range start must be non-negative")
		}

		if end < start {
			return nil, errors.New("range end must be >= start")
		}

		ranges = append(ranges, [2]int{start, end})
	}

	return ranges, nil
}
