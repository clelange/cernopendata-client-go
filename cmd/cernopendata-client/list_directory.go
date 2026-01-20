package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/clelange/cernopendata-client-go/internal/config"
	"github.com/clelange/cernopendata-client-go/internal/lister"
	"github.com/clelange/cernopendata-client-go/internal/printer"
)

// printEntries formats and prints directory entries based on verbose mode
func printEntries(entries []lister.FileInfo, verbose bool) {
	for _, entry := range entries {
		dirMarker := ""
		if entry.IsDir {
			dirMarker = "/"
		}
		if verbose {
			printer.DisplayOutput(fmt.Sprintf("%s\t%d\t%s%s", entry.Name, entry.Size, entry.ModTime, dirMarker))
		} else {
			printer.DisplayOutput(fmt.Sprintf("%s%s", entry.Name, dirMarker))
		}
	}
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
			printEntries(entries, verbose)
		} else {
			entries, err := lister.ListDirectory(ctx, path)
			if err != nil {
				printer.DisplayMessage(printer.Error, fmt.Sprintf("Failed to list directory: %v", err))
				os.Exit(1)
			}
			printEntries(entries, verbose)
		}
	},
}

func init() {
	listDirectoryCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	listDirectoryCmd.Flags().BoolP("recursive", "r", false, "Iterate recursively in the given directory path")
	listDirectoryCmd.Flags().IntP("timeout", "t", config.ListDirectoryTimeout, "Timeout in seconds after which to exit running the command")
}
