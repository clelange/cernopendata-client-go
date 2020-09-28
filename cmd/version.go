package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	versionCmd.Flags().BoolVar(&shortened, "short", false, "Print just the version number.")
	rootCmd.AddCommand(versionCmd)
}

var (
	// Versioning
	shortened = false
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	output    = "json"

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the current version of cernopendata-client-go",
		Run: func(_ *cobra.Command, _ []string) {
			if shortened {
				fmt.Println(version)
			} else {
				fmt.Println("cernopendata-client-go", version, "commit", commit, "built at", date)
			}
			return
		},
	}
)
