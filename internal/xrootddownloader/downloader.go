package xrootddownloader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/clelange/cernopendata-client-go/internal/printer"
	"go-hep.org/x/hep/xrootd"
	"go-hep.org/x/hep/xrootd/xrdio"
)

type DownloadStats struct {
	TotalFiles      int
	TotalBytes      int64
	DownloadedFiles int
	DownloadedBytes int64
	FailedFiles     int
	SkippedFiles    int
}

type FileDownloadResult struct {
	URL      string
	Path     string
	Size     int64
	Checksum string
	Success  bool
	Error    error
	Retries  int
}

type Downloader struct {
	client     *xrootd.Client
	retryLimit int
	retrySleep int
	verbose    bool
	dryRun     bool
	address    string
	username   string
}

func NewDownloader() *Downloader {
	return &Downloader{
		retryLimit: 10,
		retrySleep: 5,
		username:   "gopher",
	}
}

func (d *Downloader) DownloadFile(ctx context.Context, url, destPath string, resume bool) (*FileDownloadResult, error) {
	var existingSize int64

	if resume {
		if fi, err := os.Stat(destPath); err == nil {
			existingSize = fi.Size()
			printer.DisplayMessage(printer.Note, fmt.Sprintf("Resuming %s from %d bytes", destPath, existingSize))
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error checking file: %w", err)
		}
	}

	parsedURL, err := xrdio.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XRootD URL: %w", err)
	}

	if d.client == nil {
		client, err := xrootd.NewClient(ctx, parsedURL.Addr, d.username)
		if err != nil {
			return nil, fmt.Errorf("failed to create XRootD client: %w", err)
		}
		d.client = client
		d.address = parsedURL.Addr
	}

	fs := d.client.FS()

	var lastErr error
	var attempt int

	for attempt = 0; attempt < d.retryLimit; attempt++ {
		if attempt > 0 {
			printer.DisplayMessage(printer.Note, fmt.Sprintf("Retry attempt %d/%d after %ds...", attempt+1, d.retryLimit, d.retrySleep))
			time.Sleep(time.Duration(d.retrySleep) * time.Second)
		}

		var offset int64 = 0
		if resume && existingSize > 0 {
			offset = existingSize
		}

		file, err := fs.Open(ctx, parsedURL.Path, 0, 0)
		if err != nil {
			lastErr = err
			printer.DisplayMessage(printer.Note, fmt.Sprintf("Failed to open file: %v", err))
			continue
		}
		defer file.Close(ctx)

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}

		var localFile *os.File
		if resume && existingSize > 0 {
			localFile, err = os.OpenFile(destPath, os.O_APPEND|os.O_WRONLY, 0644)
		} else {
			localFile, err = os.Create(destPath)
		}

		if err != nil {
			file.Close(ctx)
			return nil, fmt.Errorf("failed to open local file: %w", err)
		}

		buf := make([]byte, 32*1024)
		var copied int64
		totalBytes := offset

	retryLoop:
		for {
			select {
			case <-ctx.Done():
				localFile.Close()
				return &FileDownloadResult{
					URL:     url,
					Path:    destPath,
					Success: false,
					Error:   ctx.Err(),
					Retries: attempt,
				}, ctx.Err()
			default:
			}

			n, err := file.ReadAt(buf, totalBytes)
			if n > 0 {
				if _, err := localFile.Write(buf[:n]); err != nil {
					localFile.Close()
					lastErr = err
					printer.DisplayMessage(printer.Note, fmt.Sprintf("Write error: %v", err))
					continue retryLoop
				}
				copied += int64(n)
				totalBytes += int64(n)
			}
			if err == io.EOF {
				break retryLoop
			}
			if err != nil {
				localFile.Close()
				lastErr = err
				printer.DisplayMessage(printer.Note, fmt.Sprintf("Read error: %v", err))
				continue retryLoop
			}
			if n == 0 {
				printer.DisplayMessage(printer.Note, "Read returned 0 bytes, assuming EOF")
				break retryLoop
			}
		}

		localFile.Close()

		if d.verbose {
			printer.DisplayMessage(printer.Note, fmt.Sprintf("Downloaded %d bytes to %s", copied, destPath))
		}

		return &FileDownloadResult{
			URL:     url,
			Path:    destPath,
			Success: true,
			Size:    existingSize + copied,
			Retries: attempt,
		}, nil
	}

	return &FileDownloadResult{
		URL:     url,
		Path:    destPath,
		Success: false,
		Error:   lastErr,
		Retries: attempt - 1,
	}, lastErr
}

func (d *Downloader) DownloadFiles(ctx context.Context, files []interface{}, baseDir string, retry int, retrySleep int, verbose bool, dryRun bool) DownloadStats {
	d.retryLimit = retry
	d.retrySleep = retrySleep
	d.verbose = verbose
	d.dryRun = dryRun

	stats := DownloadStats{}
	stats.TotalFiles = len(files)

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		printer.DisplayMessage(printer.Error, fmt.Sprintf("Failed to create directory %s: %v", baseDir, err))
		return stats
	}

	for i, file := range files {
		fileMap, ok := file.(map[string]interface{})
		if !ok {
			printer.DisplayMessage(printer.Note, fmt.Sprintf("Skipping invalid file entry %d", i))
			stats.SkippedFiles++
			continue
		}

		uri, _ := fileMap["uri"].(string)
		size, _ := fileMap["size"].(float64)
		checksum, _ := fileMap["checksum"].(string)

		stats.TotalBytes += int64(size)

		printer.DisplayMessage(printer.Info, fmt.Sprintf("Downloading file %d/%d: %s", i+1, stats.TotalFiles, filepath.Base(uri)))

		if dryRun {
			printer.DisplayMessage(printer.Note, fmt.Sprintf("Would download: %s (size: %d, checksum: %s)", uri, int64(size), checksum))
			stats.DownloadedFiles++
			stats.DownloadedBytes += int64(size)
			continue
		}

		destPath := filepath.Join(baseDir, filepath.Base(uri))

		if _, err := os.Stat(destPath); err == nil {
			printer.DisplayMessage(printer.Note, fmt.Sprintf("File already exists: %s", destPath))
			stats.SkippedFiles++
			continue
		}

		result, err := d.DownloadFile(ctx, uri, destPath, true)
		if err != nil {
			printer.DisplayMessage(printer.Error, fmt.Sprintf("Failed to download %s: %v", uri, err))
			stats.FailedFiles++
		} else if result.Success {
			stats.DownloadedFiles++
			stats.DownloadedBytes += result.Size
		}
	}

	printer.DisplayMessage(printer.Info, fmt.Sprintf("\nDownload summary:"))
	printer.DisplayMessage(printer.Note, fmt.Sprintf("  Total files:     %d", stats.TotalFiles))
	printer.DisplayMessage(printer.Note, fmt.Sprintf("  Downloaded:     %d", stats.DownloadedFiles))
	printer.DisplayMessage(printer.Note, fmt.Sprintf("  Skipped:        %d", stats.SkippedFiles))
	printer.DisplayMessage(printer.Note, fmt.Sprintf("  Failed:         %d", stats.FailedFiles))
	printer.DisplayMessage(printer.Note, fmt.Sprintf("  Total bytes:    %d", stats.DownloadedBytes))

	return stats
}

func (d *Downloader) Close() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}
