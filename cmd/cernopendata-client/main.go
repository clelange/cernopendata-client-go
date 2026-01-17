package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cernopendata/cernopendata-client-go/internal/config"
	"github.com/cernopendata/cernopendata-client-go/internal/downloader"
	"github.com/cernopendata/cernopendata-client-go/internal/lister"
	"github.com/cernopendata/cernopendata-client-go/internal/metadater"
	"github.com/cernopendata/cernopendata-client-go/internal/printer"
	"github.com/cernopendata/cernopendata-client-go/internal/searcher"
	"github.com/cernopendata/cernopendata-client-go/internal/utils"
	"github.com/cernopendata/cernopendata-client-go/internal/verifier"
	"github.com/cernopendata/cernopendata-client-go/internal/version"
	"github.com/cernopendata/cernopendata-client-go/internal/xrootddownloader"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cernopendata-client",
		Short: "CLI for CERN Open Data portal",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				_ = cmd.Help()
				return nil
			}
			return fmt.Errorf("unknown command: %s", args[0])
		},
	}

	var completionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion script",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				printer.DisplayMessage(printer.Error, "Please specify bash or zsh")
				os.Exit(1)
			}

			shell := args[0]
			switch shell {
			case "bash":
				_ = rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				_ = rootCmd.GenZshCompletion(os.Stdout)
			default:
				printer.DisplayMessage(printer.Error, fmt.Sprintf("Unsupported shell: %s (supported: bash, zsh)", shell))
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(getMetadataCmd)
	rootCmd.AddCommand(getFileLocationsCmd)
	rootCmd.AddCommand(downloadFilesCmd)
	rootCmd.AddCommand(verifyFilesCmd)
	rootCmd.AddCommand(listDirectoryCmd)
	rootCmd.AddCommand(completionCmd)

	if err := rootCmd.Execute(); err != nil {
		printer.DisplayMessage(printer.Error, fmt.Sprintf("Error: %v", err))
		os.Exit(1)
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Return version",
	Run: func(cmd *cobra.Command, args []string) {
		printer.DisplayOutput(version.Version)
	},
}

var getMetadataCmd = &cobra.Command{
	Use:   "get-metadata",
	Short: "Get metadata content of a record",
	Run: func(cmd *cobra.Command, args []string) {
		recid, err := cmd.Flags().GetInt("recid")
		if err != nil {
			printer.DisplayMessage(printer.Error, fmt.Sprintf("Invalid recid: %v", err))
			os.Exit(1)
		}
		doi, _ := cmd.Flags().GetString("doi")
		title, _ := cmd.Flags().GetString("title")
		outputValue, _ := cmd.Flags().GetString("output-value")
		filterStr, _ := cmd.Flags().GetString("filter")
		outputFormat, _ := cmd.Flags().GetString("format")
		server, _ := cmd.Flags().GetString("server")

		if server == "" {
			server = config.ServerHTTPURI
		}

		if filterStr != "" && outputValue == "" {
			printer.DisplayMessage(printer.Error, "--filter can only be used with --output-value")
			os.Exit(1)
		}

		parsedRecid, err := searcher.GetRecid(server, doi, title, recid)
		if err != nil {
			printer.DisplayMessage(printer.Error, fmt.Sprintf("Failed to find record: %v", err))
			os.Exit(1)
		}

		client := searcher.NewClient(server)
		record, err := client.GetRecord(parsedRecid)
		if err != nil {
			printer.DisplayMessage(printer.Error, fmt.Sprintf("Failed to get metadata: %v", err))
			os.Exit(1)
		}

		var filters []string
		if filterStr != "" {
			filters = []string{filterStr}
		}

		if outputValue == "" {
			output, err := metadater.FormatOutput(record.Metadata, outputFormat)
			if err != nil {
				printer.DisplayMessage(printer.Error, fmt.Sprintf("Failed to format output: %v", err))
				os.Exit(1)
			}
			printer.DisplayOutput(output)
		} else {
			metadata, err := metadater.GetNestedField(record.Metadata, outputValue)
			if err != nil {
				printer.DisplayMessage(printer.Error, fmt.Sprintf("Field not found: %v", err))
				os.Exit(1)
			}

			if len(filters) > 0 {
				items, isArray := metadata.([]interface{})
				if !isArray {
					items = []interface{}{metadata}
				}
				filtered, err := metadater.FilterArray(items, filters)
				if err != nil {
					printer.DisplayMessage(printer.Error, fmt.Sprintf("Filter error: %v", err))
					os.Exit(1)
				}
				output, err := metadater.FormatOutput(filtered[0], outputFormat)
				if err != nil {
					printer.DisplayMessage(printer.Error, fmt.Sprintf("Failed to format output: %v", err))
					os.Exit(1)
				}
				printer.DisplayOutput(output)
			} else {
				output, err := metadater.FormatOutput(metadata, outputFormat)
				if err != nil {
					printer.DisplayMessage(printer.Error, fmt.Sprintf("Failed to format output: %v", err))
					os.Exit(1)
				}
				printer.DisplayOutput(output)
			}
		}
	},
}

var getFileLocationsCmd = &cobra.Command{
	Use:   "get-file-locations",
	Short: "Get a list of data file locations of a record",
	Run: func(cmd *cobra.Command, args []string) {
		recid, err := cmd.Flags().GetInt("recid")
		if err != nil {
			printer.DisplayMessage(printer.Error, fmt.Sprintf("Invalid recid: %v", err))
			os.Exit(1)
		}
		doi, _ := cmd.Flags().GetString("doi")
		title, _ := cmd.Flags().GetString("title")
		protocol, _ := cmd.Flags().GetString("protocol")
		expand, _ := cmd.Flags().GetBool("expand")
		verbose, _ := cmd.Flags().GetBool("verbose")
		server, _ := cmd.Flags().GetString("server")

		if server == "" {
			server = config.ServerHTTPURI
		}

		parsedRecid, err := searcher.GetRecid(server, doi, title, recid)
		if err != nil {
			printer.DisplayMessage(printer.Error, fmt.Sprintf("Failed to find record: %v", err))
			os.Exit(1)
		}

		client := searcher.NewClient(server)
		record, err := client.GetRecord(parsedRecid)
		if err != nil {
			printer.DisplayMessage(printer.Error, fmt.Sprintf("Failed to get record: %v", err))
			os.Exit(1)
		}

		files := client.GetFilesList(record, protocol, expand)
		for _, file := range files {
			if verbose {
				printer.DisplayOutput(fmt.Sprintf("%s\t%d\t%s", file.URI, file.Size, file.Checksum))
			} else {
				printer.DisplayOutput(file.URI)
			}
		}
	},
}

var downloadFilesCmd = &cobra.Command{
	Use:   "download-files",
	Short: "Download files from a record",
	Run: func(cmd *cobra.Command, args []string) {
		recid, err := cmd.Flags().GetInt("recid")
		if err != nil {
			printer.DisplayMessage(printer.Error, fmt.Sprintf("Invalid recid: %v", err))
			os.Exit(1)
		}
		doi, _ := cmd.Flags().GetString("doi")
		title, _ := cmd.Flags().GetString("title")
		outputDir, _ := cmd.Flags().GetString("output-dir")
		nameFilter, _ := cmd.Flags().GetString("name-filter")
		regexpFilter, _ := cmd.Flags().GetString("regexp")
		rangeFilter, _ := cmd.Flags().GetString("range-filter")
		expand, _ := cmd.Flags().GetBool("expand")
		start, _ := cmd.Flags().GetInt("start-index")
		end, _ := cmd.Flags().GetInt("end-index")
		retry, _ := cmd.Flags().GetInt("retry")
		retrySleep, _ := cmd.Flags().GetInt("retry-sleep")
		verbose, _ := cmd.Flags().GetBool("verbose")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verifyFlag, _ := cmd.Flags().GetBool("verify")
		downloadEngine, _ := cmd.Flags().GetString("download-engine")
		server, _ := cmd.Flags().GetString("server")

		if server == "" {
			server = config.ServerHTTPURI
		}

		parsedRecid, err := searcher.GetRecid(server, doi, title, recid)
		if err != nil {
			printer.DisplayMessage(printer.Error, fmt.Sprintf("Failed to find record: %v", err))
			os.Exit(1)
		}

		if outputDir == "" {
			outputDir = fmt.Sprintf("%d", parsedRecid)
		}

		client := searcher.NewClient(server)
		record, err := client.GetRecord(parsedRecid)
		if err != nil {
			printer.DisplayMessage(printer.Error, fmt.Sprintf("Failed to get record: %v", err))
			os.Exit(1)
		}

		protocol := "http"
		if downloadEngine == "xrootd" {
			protocol = "xrootd"
		}
		files := client.GetFilesList(record, protocol, expand)
		var fileList []interface{}
		for _, file := range files {
			fileList = append(fileList, map[string]interface{}{
				"uri":      file.URI,
				"size":     float64(file.Size),
				"checksum": file.Checksum,
			})
		}

		if nameFilter != "" {
			nameFilters := strings.Split(nameFilter, ",")
			for i, filter := range nameFilters {
				nameFilters[i] = strings.TrimSpace(filter)
			}
			fileList = downloader.FilterFilesByMultipleNames(fileList, nameFilters)
		}

		if regexpFilter != "" {
			fileList = downloader.FilterFilesByRegex(fileList, regexpFilter)
		}

		if rangeFilter != "" {
			ranges, err := utils.ParseRanges([]string{rangeFilter})
			if err != nil {
				printer.DisplayMessage(printer.Error, fmt.Sprintf("Invalid range filter: %v", err))
				os.Exit(1)
			}
			fileList = downloader.FilterFilesByMultipleRanges(fileList, ranges)
		} else if start >= 0 || end >= 0 {
			if start < 0 {
				start = 0
			}
			if end < 0 {
				end = len(fileList)
			}
			if start >= end {
				fileList = []interface{}{}
			}
			fileList = downloader.FilterFilesByRange(fileList, start, end)
		}

		if len(fileList) == 0 {
			printer.DisplayMessage(printer.Error, "No files matching filters")
			os.Exit(1)
		}

		var stats downloader.DownloadStats
		if downloadEngine == "xrootd" {
			xrdDownloader := xrootddownloader.NewDownloader()
			defer xrdDownloader.Close()
			xrdStats := xrdDownloader.DownloadFiles(cmd.Context(), fileList, outputDir, retry, retrySleep, verbose, dryRun)
			stats = downloader.DownloadStats{
				TotalFiles:      xrdStats.TotalFiles,
				TotalBytes:      xrdStats.TotalBytes,
				DownloadedFiles: xrdStats.DownloadedFiles,
				DownloadedBytes: xrdStats.DownloadedBytes,
				FailedFiles:     xrdStats.FailedFiles,
				SkippedFiles:    xrdStats.SkippedFiles,
			}
		} else {
			httpDownloader := downloader.NewDownloader()
			stats = httpDownloader.DownloadFiles(fileList, outputDir, retry, retrySleep, verbose, dryRun)
		}

		if verifyFlag {
			printer.DisplayMessage(printer.Info, "\nVerifying downloaded files...")
			v := verifier.NewVerifier()
			verifyStats, err := v.VerifyFiles(outputDir, fileList)
			if err != nil {
				printer.DisplayMessage(printer.Error, fmt.Sprintf("Verification failed: %v", err))
				os.Exit(1)
			}

			if verifyStats.SizeFailed > 0 || verifyStats.ChecksumFailed > 0 {
				printer.DisplayMessage(printer.Error, "Some files failed verification")
				os.Exit(1)
			}
		}

		if stats.FailedFiles == 0 {
			printer.DisplayMessage(printer.Info, "Success!")
		}

		if stats.FailedFiles > 0 {
			os.Exit(1)
		}
	},
}

var verifyFilesCmd = &cobra.Command{
	Use:   "verify-files",
	Short: "Verify local files against expected checksums and sizes",
	Run: func(cmd *cobra.Command, args []string) {
		recid, err := cmd.Flags().GetInt("recid")
		if err != nil {
			printer.DisplayMessage(printer.Error, fmt.Sprintf("Invalid recid: %v", err))
			os.Exit(1)
		}
		doi, _ := cmd.Flags().GetString("doi")
		title, _ := cmd.Flags().GetString("title")
		inputDir, _ := cmd.Flags().GetString("input-dir")
		nameFilter, _ := cmd.Flags().GetString("name-filter")
		regexpFilter, _ := cmd.Flags().GetString("regexp")
		server, _ := cmd.Flags().GetString("server")

		if server == "" {
			server = config.ServerHTTPURI
		}

		parsedRecid, err := searcher.GetRecid(server, doi, title, recid)
		if err != nil {
			printer.DisplayMessage(printer.Error, fmt.Sprintf("Failed to find record: %v", err))
			os.Exit(1)
		}

		if inputDir == "" {
			inputDir = fmt.Sprintf("%d", parsedRecid)
		}

		client := searcher.NewClient(server)
		record, err := client.GetRecord(parsedRecid)
		if err != nil {
			printer.DisplayMessage(printer.Error, fmt.Sprintf("Failed to get record: %v", err))
			os.Exit(1)
		}

		files := client.GetFilesList(record, "http", false)
		var fileList []interface{}
		for _, file := range files {
			fileList = append(fileList, map[string]interface{}{
				"uri":      file.URI,
				"size":     float64(file.Size),
				"checksum": file.Checksum,
			})
		}

		if nameFilter != "" {
			nameFilters := strings.Split(nameFilter, ",")
			for i, filter := range nameFilters {
				nameFilters[i] = strings.TrimSpace(filter)
			}
			fileList = downloader.FilterFilesByMultipleNames(fileList, nameFilters)
		}

		if regexpFilter != "" {
			fileList = downloader.FilterFilesByRegex(fileList, regexpFilter)
		}

		if len(fileList) == 0 {
			printer.DisplayMessage(printer.Error, "No files matching filters")
			os.Exit(1)
		}

		verifier := verifier.NewVerifier()
		stats, err := verifier.VerifyFiles(inputDir, fileList)
		if err != nil {
			printer.DisplayMessage(printer.Error, fmt.Sprintf("Verification failed: %v", err))
			os.Exit(1)
		}

		printer.DisplayMessage(printer.Info, fmt.Sprintf("Verifying number of files for record %d...", parsedRecid))
		printer.DisplayMessage(printer.Note, fmt.Sprintf("Expected %d, found %d", len(fileList), stats.VerifiedFiles+stats.SizeFailed+stats.ChecksumFailed+stats.MissingFiles))

		if len(fileList) != (stats.VerifiedFiles + stats.SizeFailed + stats.ChecksumFailed + stats.MissingFiles) {
			printer.DisplayMessage(printer.Error, "File count does not match.")
			os.Exit(1)
		}

		printer.DisplayMessage(printer.Info, fmt.Sprintf("\nVerification summary:"))
		printer.DisplayMessage(printer.Note, fmt.Sprintf("  Total files:     %d", stats.TotalFiles))
		printer.DisplayMessage(printer.Note, fmt.Sprintf("  Verified:        %d", stats.VerifiedFiles))
		printer.DisplayMessage(printer.Note, fmt.Sprintf("  Size errors:     %d", stats.SizeFailed))
		printer.DisplayMessage(printer.Note, fmt.Sprintf("  Checksum errors: %d", stats.ChecksumFailed))
		printer.DisplayMessage(printer.Note, fmt.Sprintf("  Missing files:   %d", stats.MissingFiles))

		if stats.SizeFailed > 0 || stats.ChecksumFailed > 0 || stats.MissingFiles > 0 {
			os.Exit(1)
		}

		printer.DisplayMessage(printer.Info, "Success!")
	},
}

var listDirectoryCmd = &cobra.Command{
	Use:   "list-directory",
	Short: "List directory contents on XRootD server",
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		recursive, _ := cmd.Flags().GetBool("recursive")
		timeout, _ := cmd.Flags().GetInt("timeout")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if path == "" {
			printer.DisplayMessage(printer.Error, "Path is required")
			os.Exit(1)
		}

		ctx := cmd.Context()
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
			defer cancel()
		}

		lister := lister.NewLister()

		if recursive {
			entries, err := lister.ListDirectoryRecursive(ctx, path)
			if err != nil {
				printer.DisplayMessage(printer.Error, fmt.Sprintf("Failed to list directory: %v", err))
				os.Exit(1)
			}
			for _, entry := range entries {
				if verbose {
					dirMarker := ""
					if entry.IsDir {
						dirMarker = "/"
					}
					printer.DisplayOutput(fmt.Sprintf("%s\t%d\t%s%s", entry.Name, entry.Size, entry.ModTime, dirMarker))
				} else {
					dirMarker := ""
					if entry.IsDir {
						dirMarker = "/"
					}
					printer.DisplayOutput(fmt.Sprintf("%s%s", entry.Name, dirMarker))
				}
			}
		} else {
			entries, err := lister.ListDirectory(ctx, path)
			if err != nil {
				printer.DisplayMessage(printer.Error, fmt.Sprintf("Failed to list directory: %v", err))
				os.Exit(1)
			}
			for _, entry := range entries {
				if verbose {
					dirMarker := ""
					if entry.IsDir {
						dirMarker = "/"
					}
					printer.DisplayOutput(fmt.Sprintf("%s\t%d\t%s%s", entry.Name, entry.Size, entry.ModTime, dirMarker))
				} else {
					dirMarker := ""
					if entry.IsDir {
						dirMarker = "/"
					}
					printer.DisplayOutput(fmt.Sprintf("%s%s", entry.Name, dirMarker))
				}
			}
		}
	},
}

func init() {
	getMetadataCmd.Flags().IntP("recid", "r", 0, "Get metadata by record ID")
	getMetadataCmd.Flags().StringP("doi", "d", "", "Get metadata by DOI")
	getMetadataCmd.Flags().StringP("title", "t", "", "Get metadata by title")
	getMetadataCmd.Flags().StringP("output-value", "v", "", "Get specific field value from metadata")
	getMetadataCmd.Flags().StringP("filter", "f", "", "Filter output by field=value")
	getMetadataCmd.Flags().StringP("format", "m", "pretty", "Output format (pretty|json)")
	getMetadataCmd.Flags().StringP("server", "s", "", "Server URI")

	getFileLocationsCmd.Flags().IntP("recid", "i", 0, "Get file locations by record ID")
	getFileLocationsCmd.Flags().StringP("doi", "D", "", "Get file locations by DOI")
	getFileLocationsCmd.Flags().StringP("title", "T", "", "Get file locations by title")
	getFileLocationsCmd.Flags().StringP("protocol", "p", "http", "Protocol to use (http|https)")
	getFileLocationsCmd.Flags().BoolP("expand", "e", false, "Expand file indices")
	getFileLocationsCmd.Flags().BoolP("verbose", "V", false, "Verbose output")
	getFileLocationsCmd.Flags().StringP("server", "S", "", "Server URI")

	downloadFilesCmd.Flags().IntP("recid", "R", 0, "Download files by record ID")
	downloadFilesCmd.Flags().StringP("doi", "d", "", "Download files by DOI")
	downloadFilesCmd.Flags().StringP("title", "t", "", "Download files by title")
	downloadFilesCmd.Flags().StringP("output-dir", "O", "", "Output directory")
	downloadFilesCmd.Flags().StringP("name-filter", "n", "", "Filter files by glob pattern")
	downloadFilesCmd.Flags().StringP("regexp", "e", "", "Filter files by regular expression")
	downloadFilesCmd.Flags().StringP("range-filter", "r", "", "Filter files by index ranges (e.g., '0-2,5-7')")
	downloadFilesCmd.Flags().BoolP("expand", "x", false, "Expand file indices")
	downloadFilesCmd.Flags().IntP("start-index", "a", -1, "Start index of files to download")
	downloadFilesCmd.Flags().IntP("end-index", "z", -1, "End index of files to download")
	downloadFilesCmd.Flags().IntP("retry", "y", 10, "Number of retry attempts")
	downloadFilesCmd.Flags().IntP("retry-sleep", "Y", 5, "Sleep time in seconds before retrying downloads")
	downloadFilesCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	downloadFilesCmd.Flags().BoolP("progress", "P", false, "Show progress (alias for verbose)")
	downloadFilesCmd.Flags().BoolP("dry-run", "N", false, "Dry run (don't actually download)")
	downloadFilesCmd.Flags().BoolP("verify", "V", false, "Verify downloaded files")
	downloadFilesCmd.Flags().String("download-engine", "", "Download engine to use (http|xrootd)")
	downloadFilesCmd.Flags().StringP("server", "s", "", "Server URI")

	verifyFilesCmd.Flags().IntP("recid", "r", 0, "Verify files by record ID")
	verifyFilesCmd.Flags().StringP("doi", "d", "", "Verify files by DOI")
	verifyFilesCmd.Flags().StringP("title", "t", "", "Verify files by title")
	verifyFilesCmd.Flags().StringP("input-dir", "i", "", "Input directory containing files to verify")
	verifyFilesCmd.Flags().StringP("name-filter", "n", "", "Filter files by glob pattern")
	verifyFilesCmd.Flags().StringP("regexp", "e", "", "Filter files by regular expression")
	verifyFilesCmd.Flags().StringP("server", "s", "", "Server URI")

	listDirectoryCmd.Flags().StringP("path", "p", "", "XRootD path to list")
	listDirectoryCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	listDirectoryCmd.Flags().BoolP("recursive", "r", false, "Recursive directory listing")
	listDirectoryCmd.Flags().IntP("timeout", "t", 0, "Timeout in seconds")
}
