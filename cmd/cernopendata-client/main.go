package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/clelange/cernopendata-client-go/internal/config"
	"github.com/clelange/cernopendata-client-go/internal/downloader"
	"github.com/clelange/cernopendata-client-go/internal/lister"
	"github.com/clelange/cernopendata-client-go/internal/metadater"
	"github.com/clelange/cernopendata-client-go/internal/printer"
	"github.com/clelange/cernopendata-client-go/internal/searcher"
	"github.com/clelange/cernopendata-client-go/internal/utils"
	"github.com/clelange/cernopendata-client-go/internal/verifier"
	"github.com/clelange/cernopendata-client-go/internal/version"
	"github.com/clelange/cernopendata-client-go/internal/xrootddownloader"
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
	Long: `Get metadata content of a record.

Select a CERN Open Data bibliographic record by a record ID, a
DOI, or a title and return its metadata in the JSON format.

Examples:

     $ cernopendata-client get-metadata --recid 1

     $ cernopendata-client get-metadata --recid 1 --output-value title

     $ cernopendata-client get-metadata --recid 329 --output-value authors.orcid --filter name="Rousseau, David"`,
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
	Long: `Get a list of data file locations of a record.

Select a CERN Open Data bibliographic record by a record ID, a DOI, or a
title and return the list of data file locations belonging to this record.

Examples:

     $ cernopendata-client get-file-locations --recid 5500

     $ cernopendata-client get-file-locations --recid 5500 --protocol xrootd

     $ cernopendata-client get-file-locations --recid 5500 --verbose`,
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
		noExpand, _ := cmd.Flags().GetBool("no-expand")
		verbose, _ := cmd.Flags().GetBool("verbose")
		server, _ := cmd.Flags().GetString("server")

		if cmd.Flags().Changed("expand") && cmd.Flags().Changed("no-expand") {
			printer.DisplayMessage(printer.Error, "Cannot specify both --expand and --no-expand")
			os.Exit(1)
		}

		if noExpand {
			expand = false
		}

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
	Long: `Download data files belonging to a record.

Select a CERN Open Data bibliographic record by a record ID, a
DOI, or a title and download data files belonging to this record.

Examples:

     $ cernopendata-client download-files --recid 5500

     $ cernopendata-client download-files --recid 5500 --filter-name BuildFile.xml

     $ cernopendata-client download-files --recid 5500 --filter-regexp py$

     $ cernopendata-client download-files --recid 5500 --filter-range 1-4

     $ cernopendata-client download-files --recid 5500 --filter-range 1-2,5-7`,
	Run: func(cmd *cobra.Command, args []string) {
		recid, err := cmd.Flags().GetInt("recid")
		if err != nil {
			printer.DisplayMessage(printer.Error, fmt.Sprintf("Invalid recid: %v", err))
			os.Exit(1)
		}
		doi, _ := cmd.Flags().GetString("doi")
		title, _ := cmd.Flags().GetString("title")
		outputDir, _ := cmd.Flags().GetString("output-dir")
		filterName, _ := cmd.Flags().GetString("filter-name")
		filterRegexp, _ := cmd.Flags().GetString("filter-regexp")
		filterRange, _ := cmd.Flags().GetString("filter-range")
		expand, _ := cmd.Flags().GetBool("expand")
		noExpand, _ := cmd.Flags().GetBool("no-expand")
		retryLimit, _ := cmd.Flags().GetInt("retry-limit")
		retrySleep, _ := cmd.Flags().GetInt("retry-sleep")
		verbose, _ := cmd.Flags().GetBool("verbose")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verifyFlag, _ := cmd.Flags().GetBool("verify")
		downloadEngine, _ := cmd.Flags().GetString("download-engine")
		protocol, _ := cmd.Flags().GetString("protocol")
		server, _ := cmd.Flags().GetString("server")

		if cmd.Flags().Changed("expand") && cmd.Flags().Changed("no-expand") {
			printer.DisplayMessage(printer.Error, "Cannot specify both --expand and --no-expand")
			os.Exit(1)
		}

		if noExpand {
			expand = false
		}

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

		if protocol == "" {
			if downloadEngine == "xrootd" {
				protocol = "xrootd"
			} else {
				protocol = "http"
			}
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

		if filterName != "" {
			nameFilters := strings.Split(filterName, ",")
			for i, filter := range nameFilters {
				nameFilters[i] = strings.TrimSpace(filter)
			}
			fileList = downloader.FilterFilesByMultipleNames(fileList, nameFilters)
		}

		if filterRegexp != "" {
			fileList = downloader.FilterFilesByRegex(fileList, filterRegexp)
		}

		if filterRange != "" {
			ranges, err := utils.ParseRanges([]string{filterRange})
			if err != nil {
				printer.DisplayMessage(printer.Error, fmt.Sprintf("Invalid range filter: %v", err))
				os.Exit(1)
			}
			fileList = downloader.FilterFilesByMultipleRanges(fileList, ranges)
		}

		if len(fileList) == 0 {
			printer.DisplayMessage(printer.Error, "No files matching filters")
			os.Exit(1)
		}

		var stats downloader.DownloadStats
		if downloadEngine == "xrootd" {
			xrdDownloader := xrootddownloader.NewDownloader()
			defer xrdDownloader.Close()
			xrdStats := xrdDownloader.DownloadFiles(cmd.Context(), fileList, outputDir, retryLimit, retrySleep, verbose, dryRun)
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
			stats = httpDownloader.DownloadFiles(fileList, outputDir, retryLimit, retrySleep, verbose, dryRun)
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
	Long: `Verify downloaded data file integrity.

Select a CERN Open Data bibliographic record by a record ID, a
DOI, or a title and verify integrity of downloaded data files
belonging to this record.

Examples:

     $ cernopendata-client verify-files --recid 5500`,
	Run: func(cmd *cobra.Command, args []string) {
		recid, err := cmd.Flags().GetInt("recid")
		if err != nil {
			printer.DisplayMessage(printer.Error, fmt.Sprintf("Invalid recid: %v", err))
			os.Exit(1)
		}
		doi, _ := cmd.Flags().GetString("doi")
		title, _ := cmd.Flags().GetString("title")
		inputDir, _ := cmd.Flags().GetString("input-dir")
		filterName, _ := cmd.Flags().GetString("filter-name")
		filterRegexp, _ := cmd.Flags().GetString("filter-regexp")
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

		if filterName != "" {
			nameFilters := strings.Split(filterName, ",")
			for i, filter := range nameFilters {
				nameFilters[i] = strings.TrimSpace(filter)
			}
			fileList = downloader.FilterFilesByMultipleNames(fileList, nameFilters)
		}

		if filterRegexp != "" {
			fileList = downloader.FilterFilesByRegex(fileList, filterRegexp)
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
	Use:   "list-directory [path]",
	Short: "List contents of a EOSPUBLIC Open Data directory.",
	Long: `List contents of a EOSPUBLIC Open Data directory.

Returns the list of files and subdirectories of a given EOSPUBLIC directory.

Examples:

     $ cernopendata-client list-directory /eos/opendata/cms/validated-runs/Commissioning10

     $ cernopendata-client list-directory /eos/opendata/cms/Run2010B/BTau/AOD --recursive

     $ cernopendata-client list-directory /eos/opendata/cms/Run2010B --recursive --timeout 10`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		recursive, _ := cmd.Flags().GetBool("recursive")
		timeout, _ := cmd.Flags().GetInt("timeout")
		verbose, _ := cmd.Flags().GetBool("verbose")

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
	getMetadataCmd.Flags().IntP("recid", "r", 0, "Record ID (exact match)")
	getMetadataCmd.Flags().StringP("doi", "d", "", "Digital Object Identifier (exact match)")
	getMetadataCmd.Flags().StringP("title", "t", "", "Record title (exact match, no wildcards)")
	getMetadataCmd.Flags().StringP("output-value", "v", "", "Output value of only desired metadata field [example=title]")
	getMetadataCmd.Flags().StringP("filter", "f", "", "Filter only certain output values matching filtering criteria. [Use --filter some_field_name=some_value]")
	getMetadataCmd.Flags().StringP("format", "m", "pretty", "Output format (pretty|json)")
	getMetadataCmd.Flags().StringP("server", "s", "", "Which CERN Open Data server to query? [default=http://opendata.cern.ch]")

	getFileLocationsCmd.Flags().IntP("recid", "i", 0, "Record ID (exact match)")
	getFileLocationsCmd.Flags().StringP("doi", "D", "", "Digital Object Identifier (exact match)")
	getFileLocationsCmd.Flags().StringP("title", "T", "", "Record title (exact match, no wildcards)")
	getFileLocationsCmd.Flags().StringP("protocol", "p", "http", "Protocol to be used in links [http,xrootd]")
	getFileLocationsCmd.Flags().BoolP("expand", "e", true, "Expand file indexes?")
	getFileLocationsCmd.Flags().Bool("no-expand", false, "Don't expand file indexes")
	getFileLocationsCmd.Flags().BoolP("verbose", "V", false, "Output also the file size (in the second column) and the file checksum (in the third column)")
	getFileLocationsCmd.Flags().StringP("server", "S", "", "Which CERN Open Data server to query? [default=http://opendata.cern.ch]")

	downloadFilesCmd.Flags().IntP("recid", "R", 0, "Record ID (exact match)")
	downloadFilesCmd.Flags().StringP("doi", "d", "", "Digital Object Identifier (exact match)")
	downloadFilesCmd.Flags().StringP("title", "t", "", "Record title (exact match, no wildcards)")
	downloadFilesCmd.Flags().StringP("output-dir", "O", "", "Output directory")
	downloadFilesCmd.Flags().StringP("filter-name", "n", "", "Download files matching exactly the file name")
	downloadFilesCmd.Flags().StringP("filter-regexp", "e", "", "Download files matching the regular expression")
	downloadFilesCmd.Flags().StringP("filter-range", "r", "", "Download files from a specified list range (i-j)")
	downloadFilesCmd.Flags().BoolP("expand", "x", true, "Expand file indexes?")
	downloadFilesCmd.Flags().Bool("no-expand", false, "Don't expand file indexes")
	downloadFilesCmd.Flags().IntP("retry-limit", "y", 10, "Number of retries when downloading a file")
	downloadFilesCmd.Flags().IntP("retry-sleep", "Y", 5, "Sleep time in seconds before retrying downloads")
	downloadFilesCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	downloadFilesCmd.Flags().BoolP("progress", "P", false, "Show progress (alias for verbose)")
	downloadFilesCmd.Flags().BoolP("dry-run", "N", false, "Dry run (don't actually download)")
	downloadFilesCmd.Flags().BoolP("verify", "V", false, "Verify downloaded files")
	downloadFilesCmd.Flags().String("download-engine", "", "Download engine to use (http|xrootd)")
	downloadFilesCmd.Flags().StringP("protocol", "p", "", "Protocol to be used in links [http,xrootd]")
	downloadFilesCmd.Flags().StringP("server", "s", "", "Which CERN Open Data server to query? [default=http://opendata.cern.ch]")

	verifyFilesCmd.Flags().IntP("recid", "r", 0, "Record ID (exact match)")
	verifyFilesCmd.Flags().StringP("doi", "d", "", "Digital Object Identifier (exact match)")
	verifyFilesCmd.Flags().StringP("title", "t", "", "Record title (exact match, no wildcards)")
	verifyFilesCmd.Flags().StringP("input-dir", "i", "", "Input directory containing files to verify")
	verifyFilesCmd.Flags().StringP("filter-name", "n", "", "Verify files matching exactly the file name")
	verifyFilesCmd.Flags().StringP("filter-regexp", "e", "", "Verify files matching the regular expression")
	verifyFilesCmd.Flags().StringP("server", "s", "", "Which CERN Open Data server to query? [default=http://opendata.cern.ch]")

	listDirectoryCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	listDirectoryCmd.Flags().BoolP("recursive", "r", false, "Iterate recursively in the given directory path")
	listDirectoryCmd.Flags().IntP("timeout", "t", config.ListDirectoryTimeout, "Timeout in seconds after which to exit running the command")
}
