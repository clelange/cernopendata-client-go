package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.PersistentFlags().StringVarP(&protocol, "protocol", "p", "http", "Protocol to be used (http or root)")
	listCmd.Flags().BoolVar(&jsonOut, "json", false, "JSON output.")

}

var (
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "Get a list of data file locations of a record",
		Long: `This command will print a list of data file locations
(URIs) associated with the record ID provided.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := validateProtocolChoice(); err != nil {
				er(err)
			}
			if err := verifyUniqueID(); err != nil {
				er(err)
			}
			recordJSON, err := getRecordJSON()
			if err != nil {
				er(err)
			}
			if jsonOut {
				b, err := json.MarshalIndent(recordJSON, "", "\t")
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to marshal json: %v", err)
					return
				}
				fmt.Println(string(b))
				return
			}
			filesList, err := getFilesList(recordJSON)
			if err != nil {
				er(err)
			}
			printList(filesList)
		},
	}
)

func printList(filesList []string) {

	for i := range filesList {
		fmt.Println(filesList[i])
	}
}
