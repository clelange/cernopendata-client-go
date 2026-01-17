package utils

import (
	"errors"
	"fmt"
	"strings"
)

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
