package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	// Used for flags.
	recordID string
	server   string
	noExpand bool

	rootCmd = &cobra.Command{
		Use:   "cernopendata-client-go",
		Short: "A commandline tool to interact with the CERN Open Data portal",
		Long: `The cernopendata-client-go is a tool to retrieve information
from the CERN Open Data portal and to download individual
files and complete records.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {

	rootCmd.PersistentFlags().StringVarP(&recordID, "record", "r", "", "record ID to list (required)")
	rootCmd.PersistentFlags().BoolVar(&noExpand, "no-expand", false, "expand files indices")
	rootCmd.MarkFlagRequired("record")
	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "http://opendata.cern.ch", "CERN Open Data server to query")
	err := doc.GenMarkdownTree(rootCmd, "docs")
	if err != nil {
		fmt.Errorf("error generating docs", err)
	}

}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}

func verifyRecordID() error {
	// Verify that record provided exists

	// Check if recordID provided is a positive integer
	i, err := strconv.Atoi(recordID)
	if (err != nil) || (i < 1) {
		return fmt.Errorf("recordID needs to be positive integer number")
	}

	queryURL := fmt.Sprintf("%s/record/%s", server, recordID)
	resp, err := http.Get(queryURL)
	if err != nil {
		return fmt.Errorf("could not contact server at %s", server)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not retrieve record %s: %s", recordID, resp.Status)
	}
	return nil
}

func getRecordJSON() (map[string]interface{}, error) {
	// Get the API for the given recordID

	queryURL := fmt.Sprintf("%s/api/records/%s", server, recordID)
	resp, err := http.Get(queryURL)
	if err != nil {
		return nil, fmt.Errorf("could not contact server at %s", server)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not retrieve record API %s: %s", recordID, resp.Status)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiInterface interface{}
	err = json.Unmarshal(body, &apiInterface)

	apiResponse := apiInterface.(map[string]interface{})

	_, hasMetadata := apiResponse["metadata"]
	if hasMetadata {
		metadata := apiResponse["metadata"].(map[string]interface{})
		_, hasChecksumFiles := (metadata["_files"])
		if hasChecksumFiles {
			delete(metadata, "_files")
		}
		_, hasFiles := (metadata["files"])
		if hasFiles {
			filesList := metadata["files"].([]interface{})
			for idx := range filesList {
				fileEntry := filesList[idx].(map[string]interface{})
				_, hasBucket := fileEntry["bucket"]
				if hasBucket {
					delete(fileEntry, "bucket")
				}
				_, hasVersionID := fileEntry["version_id"]
				if hasVersionID {
					delete(fileEntry, "version_id")
				}
			}
		}
	}

	return apiResponse, nil

}

func getFilesList(recordJSON map[string]interface{}) ([]string, error) {

	_, hasMetadata := recordJSON["metadata"]
	filesSlice := make([]string, 0)
	if hasMetadata {
		metadata := recordJSON["metadata"].(map[string]interface{})
		_, hasFiles := (metadata["files"])
		if hasFiles {
			filesList := metadata["files"].([]interface{})
			for idx := range filesList {
				fileEntry := filesList[idx].(map[string]interface{})
				_, hasURI := (fileEntry["uri"])
				if hasURI {
					filesSlice = append(filesSlice, fileEntry["uri"].(string))
				}
			}
		}
	}

	if !noExpand {
		filesSliceExpanded := make([]string, 0)
		for idx := range filesSlice {
			if strings.HasSuffix(filesSlice[idx], "_file_index.txt") {
				fileURL := strings.Replace(filesSlice[idx], "root://eospublic.cern.ch/", server, 1)
				resp, err := http.Get(fileURL)
				if err != nil {
					return nil, fmt.Errorf("could not contact server at %s", server)
				}
				if resp.StatusCode != http.StatusOK {
					return nil, fmt.Errorf("could not retrieve record API %s: %s", recordID, resp.Status)
				}
				defer resp.Body.Close()

				scanner := bufio.NewScanner(resp.Body)
				for i := 0; scanner.Scan(); i++ {
					filesSliceExpanded = append(filesSliceExpanded, scanner.Text())
				}
			} else if strings.HasSuffix(filesSlice[idx], "_file_index.json") {
			} else {
				filesSliceExpanded = append(filesSliceExpanded, filesSlice[idx])
			}
		}
		filesSlice = filesSliceExpanded
	}
	return filesSlice, nil
}
