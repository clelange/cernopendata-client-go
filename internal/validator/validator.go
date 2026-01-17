package validator

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

var (
	ErrInvalidRecid     = errors.New("recid should be a positive integer")
	ErrInvalidServer    = errors.New("server should be a valid HTTP/HTTPS URI")
	ErrInvalidRange     = errors.New("range should have start and end index (i-j)")
	ErrInvalidDirectory = errors.New("directory is not a valid EOSPUBLIC path")
	ErrInvalidRetry     = errors.New("retry should be a positive integer")
	ErrEmptyInput       = errors.New("value cannot be empty")
)

func ValidateRecid(recid int) error {
	if recid <= 0 {
		return fmt.Errorf("%w: got %d", ErrInvalidRecid, recid)
	}
	return nil
}

func ValidateServer(server string) error {
	if server == "" {
		return fmt.Errorf("%w: server cannot be empty", ErrEmptyInput)
	}

	u, err := url.Parse(server)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidServer, err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%w: got scheme %q", ErrInvalidServer, u.Scheme)
	}

	return nil
}

func ValidateRange(rangeStr string, count int) (start, end int, err error) {
	if rangeStr == "" {
		return 0, 0, fmt.Errorf("%w: range cannot be empty", ErrEmptyInput)
	}

	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return 0, 0, ErrInvalidRange
	}

	start, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("%w: start is not a number: %v", ErrInvalidRange, err)
	}

	end, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("%w: end is not a number: %v", ErrInvalidRange, err)
	}

	if start <= 0 {
		return 0, 0, fmt.Errorf("%w: range should start from a positive integer, got %d", ErrInvalidRange, start)
	}

	if end > count {
		return 0, 0, fmt.Errorf("%w: range is too big, there are total %d files, got end %d", ErrInvalidRange, count, end)
	}

	if end < start {
		return 0, 0, fmt.Errorf("%w: start %d is greater than end %d", ErrInvalidRange, start, end)
	}

	return start, end, nil
}

func ValidateDirectory(directory string) error {
	if directory == "" {
		return fmt.Errorf("%w: directory cannot be empty", ErrEmptyInput)
	}

	if !strings.HasPrefix(directory, "/eos/opendata/") {
		return fmt.Errorf("%w: %q does not start with /eos/opendata/", ErrInvalidDirectory, directory)
	}

	return nil
}

func ValidateRetryLimit(retryLimit int) error {
	if retryLimit <= 0 {
		return fmt.Errorf("%w: retry limit should be a positive integer, got %d", ErrInvalidRetry, retryLimit)
	}
	return nil
}

func ValidateRetrySleep(retrySleep int) error {
	if retrySleep <= 0 {
		return fmt.Errorf("%w: retry sleep should be a positive integer, got %d", ErrInvalidRetry, retrySleep)
	}
	return nil
}
