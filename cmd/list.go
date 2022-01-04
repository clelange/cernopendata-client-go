package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.PersistentFlags().StringVarP(&protocol, "protocol", "p", "http", "Protocol to be used (http or root)")
	listCmd.Flags().BoolVar(&jsonOut, "json", false, "Json output.")

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
				b, err := json.Marshal(recordJSON)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to marshal json: %v", err)
					return
				}
				var out bytes.Buffer
				json.Indent(&out, b, "", "\t")
				out.WriteTo(os.Stdout)
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
