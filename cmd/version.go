package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	versionCmd.Flags().BoolVar(&shortened, "short", false, "Print just the version number.")
	versionCmd.Flags().BoolVar(&jsonOut, "json", false, "Json output.")
	rootCmd.AddCommand(versionCmd)
}

var (
	// Versioning
	shortened = false
	jsonOut   = false
	version   = "dev"
	commit    = "none"
	date      = "unknown"

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the current version of cernopendata-client-go",
		Run: func(_ *cobra.Command, _ []string) {
			switch {
			case jsonOut:
				type versioning struct {
					Name    string
					Version string
					Commit  string
					Date    string
				}
				b, err := json.Marshal(versioning{
					Name:    "cernopendata-client-go",
					Version: version,
					Commit:  commit,
					Date:    date,
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to marshal json: %v", err)
					return
				}
				var out bytes.Buffer
				json.Indent(&out, b, "", "\t")
				out.WriteTo(os.Stdout)
			case shortened:
				fmt.Println(version)
			default:
				fmt.Println("cernopendata-client-go", version, "commit", commit, "built at", date)
			}
		},
	}
)
