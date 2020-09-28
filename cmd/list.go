package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var (
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "Get a list of data file locations of a record",
		Long: `This command will print a list of data file locations
(URIs) associated with the record ID provided.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := verifyRecordID()
			if err != nil {
				er(err)
			}
			recordJSON, err := getRecordJSON()
			if err != nil {
				er(err)
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
