package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	// Used for flags.
	recordID string
	doi      string
	server   string
	protocol string
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

	rootCmd.PersistentFlags().StringVarP(&recordID, "record", "r", "", "record ID to list")
	rootCmd.PersistentFlags().StringVarP(&doi, "doi", "d", "", "Digital Object Identifier")
	rootCmd.PersistentFlags().BoolVar(&noExpand, "no-expand", false, "expand files indices")
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

func validateProtocolChoice() error {

	validProtocols := []string{"http", "root"}
	for i := range validProtocols {
		if validProtocols[i] == protocol {
			return nil
		}
	}
	return fmt.Errorf("Invalid protocol option \"%s\", choose either of %s", protocol, strings.Join(validProtocols, ", "))

}

type apiRecords struct {
	Hits struct {
		Hits []struct {
			Created time.Time `json:"created"`
			ID      int       `json:"id"`
			Links   struct {
				Self string `json:"self"`
			} `json:"links"`
			Metadata struct {
				Schema   string `json:"$schema"`
				Abstract struct {
					Description string `json:"description"`
					Links       []struct {
						Recid string `json:"recid"`
					} `json:"links"`
				} `json:"abstract"`
				ControlNumber   string   `json:"control_number"`
				DateCreated     []string `json:"date_created"`
				DatePublished   string   `json:"date_published"`
				DateReprocessed string   `json:"date_reprocessed"`
				Doi             string   `json:"doi"`
				Experiment      string   `json:"experiment"`
				License         struct {
					Attribution string `json:"attribution"`
				} `json:"license"`
				Methodology struct {
					Description string `json:"description"`
				} `json:"methodology"`
				Publisher string `json:"publisher"`
				Recid     string `json:"recid"`
				Title     string `json:"title"`
			} `json:"metadata"`
			Updated time.Time `json:"updated"`
		} `json:"hits"`
		Total int `json:"total"`
	} `json:"hits"`
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
}

func getRecordIDFromDoi() error {
	params := url.Values{}
	params.Add("page", "1")
	params.Add("size", "1")
	params.Add("q", fmt.Sprintf(`doi:"%s"`, doi))

	queryURL := fmt.Sprintf("%s/api/records?%s", server, params.Encode())
	resp, err := http.Get(queryURL)
	if err != nil {
		return fmt.Errorf("could not contact server at %s", server)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not retrieve doi %s: %s", doi, resp.Status)
	}
	defer resp.Body.Close()

	r := &apiRecords{}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("failed to parse json from response: %v", err)
	}

	if r.Hits.Total != 1 {
		return fmt.Errorf("failed to get record id from doi")
	}
	recordID = r.Hits.Hits[0].Metadata.Recid
	i, err := strconv.Atoi(recordID)
	if (err != nil) || (i < 1) {
		return fmt.Errorf("recordID needs to be positive integer number")
	}

	return nil
}

func verifyUniqueID() error {
	switch {
	case recordID != "" && doi != "":
		return fmt.Errorf("can not combine recordID with doi")
	case recordID != "":
		i, err := strconv.Atoi(recordID)
		if (err != nil) || (i < 1) {
			return fmt.Errorf("recordID needs to be positive integer number")
		}
	case doi != "":
		if err := getRecordIDFromDoi(); err != nil {
			return fmt.Errorf("failed to get recordID from doi: %v", err)
		}
	default:
		return fmt.Errorf("either recordID or doi is required")
	}
	queryURL := fmt.Sprintf("%s/record/%s", server, recordID)
	resp, err := http.Get(queryURL)
	if err != nil {
		return fmt.Errorf("could not contact server at %s", server)
	}
	defer resp.Body.Close()
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

	if protocol == "http" {
		filesSliceHttp := make([]string, 0)
		for i := range filesSlice {
			filesSliceHttp = append(filesSliceHttp, strings.Replace(filesSlice[i], "root://eospublic.cern.ch/", server, 1))
		}
		filesSlice = filesSliceHttp
	}

	return filesSlice, nil
}
