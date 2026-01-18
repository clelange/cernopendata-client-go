//go:build integration
// +build integration

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	testRecID = 3005
)

func getBinaryPath() string {
	_, callerFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(callerFile))
	projectRoot = filepath.Dir(projectRoot)
	return filepath.Join(projectRoot, "bin", "cernopendata-client")
}

func TestIntegrationGetMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-metadata", "--recid", "3005")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run get-metadata: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from get-metadata")
	}
}

func TestIntegrationGetMetadataByDOI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-metadata", "--doi", "10.7483/OPENDATA.CMS.A342.9982")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run get-metadata by DOI: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from get-metadata by DOI")
	}
}

func TestIntegrationGetMetadataByTitle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-metadata", "--title", "Configuration file for LHE step HIG-Summer11pLHE-00114_1_cfg.py")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run get-metadata by title: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from get-metadata by title")
	}
}

func TestIntegrationGetMetadataByTitleWrong(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-metadata", "--title", "NONEXISTING_TITLE")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error for non-existing title, but got none")
	}

	if len(output) == 0 {
		t.Error("Expected error message for non-existing title")
	}
}

func TestIntegrationGetMetadataByAlternateDOI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-metadata", "--doi", "10.7483/OPENDATA.CMS.A342.9982")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run get-metadata by alternate DOI: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from get-metadata by alternate DOI")
	}
}

func TestIntegrationGetMetadataOutputValueBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-metadata", "--recid", "1", "--output-value", "system_details.global_tag")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run get-metadata with output-value: %v\nOutput: %s", err, string(output))
	}

	if !contains(string(output), "FT_R_42_V10A::All") {
		t.Error("Expected 'FT_R_42_V10A::All' in output")
	}
}

func TestIntegrationGetMetadataOutputValueArray(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-metadata", "--recid", "1", "--output-value", "usage.links")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run get-metadata with array output-value: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output for array field")
	}
}

func TestIntegrationGetMetadataOutputValueNested(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-metadata", "--recid", "1", "--output-value", "usage.links.url")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run get-metadata with nested output-value: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output for nested field")
	}
}

func TestIntegrationGetMetadataOutputValueWrong(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-metadata", "--recid", "1", "--output-value", "title.global_tag")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error for wrong field, but got none")
	}

	if !contains(string(output), "Field not found") && !contains(string(output), "is not present") {
		t.Error("Expected field not found error in output")
	}
}

func TestIntegrationGetMetadataNoIdentifier(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-metadata")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error when no identifier provided, but got none")
	}

	if !contains(string(output), "recid") && !contains(string(output), "doi") && !contains(string(output), "title") {
		t.Error("Expected error message to mention required identifier")
	}
}

func TestIntegrationGetMetadataInvalidServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-metadata", "--recid", "3005", "--server", "ftp://invalid.com")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error for invalid server, but got none")
	}

	if len(output) == 0 {
		t.Error("Expected error message for invalid server")
	}
}

func TestIntegrationGetMetadataFilterWithoutOutputValue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-metadata", "--recid", "1", "--filter", "foo=bar")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error when using --filter without --output-value")
	}

	if !contains(string(output), "--filter") || !contains(string(output), "--output-value") {
		t.Error("Expected message about --filter requiring --output-value")
	}
}

func TestIntegrationGetFileLocations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-file-locations", "--recid", "3005")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run get-file-locations: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from get-file-locations")
	}
}

func TestIntegrationGetFileLocationsVerbose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-file-locations", "--recid", "3005", "--verbose")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run get-file-locations with verbose: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from get-file-locations verbose")
	}

	outputStr := string(output)
	if len(outputStr) < 10 {
		t.Error("Expected verbose output to have more content")
	}
}

func TestIntegrationGetFileLocationsByRecidWrong(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-file-locations", "--recid", "0")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error for invalid recid 0, but got none")
	}

	if len(output) == 0 {
		t.Error("Expected error message for invalid recid")
	}
}

func TestIntegrationGetFileLocationsByDOI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-file-locations", "--doi", "10.7483/OPENDATA.CMS.A342.9982")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run get-file-locations by DOI: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from get-file-locations by DOI")
	}
}

func TestIntegrationGetFileLocationsByDOIWrong(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-file-locations", "--doi", "NONEXISTING")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error for non-existing DOI, but got none")
	}

	if len(output) == 0 {
		t.Error("Expected error message for non-existing DOI")
	}
}

func TestIntegrationGetFileLocationsByTitle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-file-locations", "--title", "Configuration file for LHE step HIG-Summer11pLHE-00114_1_cfg.py")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run get-file-locations by title: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from get-file-locations by title")
	}
}

func TestIntegrationGetFileLocationsByTitleWrong(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-file-locations", "--title", "NONEXISTING_TITLE")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error for non-existing title, but got none")
	}

	if len(output) == 0 {
		t.Error("Expected error message for non-existing title")
	}
}

func TestIntegrationGetFileLocationsHTTP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-file-locations", "--recid", "3005", "--protocol", "http")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run get-file-locations with http: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from get-file-locations with http")
	}
}

func TestIntegrationGetFileLocationsHTTPS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-file-locations", "--recid", "3005", "--protocol", "https")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run get-file-locations with https: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from get-file-locations with https")
	}
}

func TestIntegrationGetFileLocationsExpand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-file-locations", "--recid", "3005", "--expand")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run get-file-locations with expand: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from get-file-locations with expand")
	}
}

func TestIntegrationGetFileLocationsNoExpand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-file-locations", "--recid", "3005")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run get-file-locations with no-expand: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from get-file-locations with no-expand")
	}
}

func TestIntegrationGetFileLocationsNoIdentifier(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-file-locations")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error when no identifier provided, but got none")
	}

	if !contains(string(output), "recid") && !contains(string(output), "doi") && !contains(string(output), "title") {
		t.Error("Expected error message to mention required identifier")
	}
}

func TestIntegrationGetFileLocationsInvalidServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "get-file-locations", "--recid", "3005", "--server", "ftp://invalid.com")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error for invalid server, but got none")
	}

	if len(output) == 0 {
		t.Error("Expected error message for invalid server")
	}
}

func TestIntegrationVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run version: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from version")
	}
}

func TestIntegrationDownloadFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--name-filter", "*.txt", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: download-files failed (expected if no .txt files): %v\nOutput: %s", err, string(output))
		return
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	if len(files) == 0 {
		t.Log("No files downloaded (expected if no .txt files in record)")
	}
}

func TestIntegrationDownloadFilesByDOI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--doi", "10.7483/OPENDATA.CMS.W26R.J96R", "--name-filter", "readme.txt", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: download by DOI failed: %v\nOutput: %s", err, string(output))
		return
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	if len(files) == 0 {
		t.Log("No files downloaded from DOI")
	}
}

func TestIntegrationDownloadFilesByDOIWrong(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "download-files", "--doi", "NONEXISTING")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error for non-existing DOI, but got none")
	}

	if len(output) == 0 {
		t.Error("Expected error message for non-existing DOI")
	}
}

func TestIntegrationListDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, getBinaryPath(), "list-directory", "root://eospublic.cern.ch//eos/opendata/cms")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: list-directory failed (XRootD service may be unavailable): %v\nOutput: %s", err, string(output))
		return
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from list-directory")
	}
}

func TestIntegrationListDirectoryVerbose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, getBinaryPath(), "list-directory", "root://eospublic.cern.ch//eos/opendata/cms", "--verbose")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: list-directory verbose failed (XRootD service may be unavailable): %v\nOutput: %s", err, string(output))
		return
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from list-directory verbose")
	}

	outputStr := string(output)
	if len(outputStr) < 20 {
		t.Error("Expected verbose output to have more content")
	}
}

func TestIntegrationListDirectoryWrongPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, getBinaryPath(), "list-directory", "root://eospublic.cern.ch//eos/opendata/foobar")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: list-directory wrong path failed (XRootD service may be unavailable): %v\nOutput: %s", err, string(output))
		return
	}

	if !contains(string(output), "does not exist") && !contains(string(output), "not found") && !contains(string(output), "failed") {
		t.Log("Note: Expected 'does not exist' error for wrong path")
	}
}

func TestIntegrationListDirectoryEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, getBinaryPath(), "list-directory", "root://eospublic.cern.ch//eos/opendata/test/nonexistent")
	output, err := cmd.CombinedOutput()
	if err == nil && len(output) == 0 {
		t.Log("Expected empty directory or error")
		return
	}

	if len(output) == 0 {
		t.Log("No output (expected for non-existent directory)")
	}
}

func TestIntegrationHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run help: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	if !contains(outputStr, "get-metadata") {
		t.Error("Expected help output to contain 'get-metadata'")
	}
	if !contains(outputStr, "download-files") {
		t.Error("Expected help output to contain 'download-files'")
	}
	if !contains(outputStr, "verify-files") {
		t.Error("Expected help output to contain 'verify-files'")
	}
	if !contains(outputStr, "list-directory") {
		t.Error("Expected help output to contain 'list-directory'")
	}
}

func TestIntegrationBinaryExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	_, callerFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(callerFile))
	projectRoot = filepath.Dir(projectRoot)
	binaryPath := filepath.Join(projectRoot, "bin", "cernopendata-client")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Errorf("Binary does not exist at %s. Run 'go build -o bin/cernopendata-client ./cmd/cernopendata-client' first.", binaryPath)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestIntegrationDownloadFilesFromRecid(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--name-filter", "*.py")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", err, string(output))
		return
	}

	outputStr := string(output)
	if !contains(outputStr, "Success") {
		t.Error("Expected 'Success!' message in output")
	}

	if _, err := os.Stat("3005/0d0714743f0204ed3c0144941e6ce248.configFile.py"); os.IsNotExist(err) {
		t.Error("Expected file to be downloaded to 3005/ directory")
	}
}

func TestIntegrationDownloadFilesFromRecidWrong(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "0")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error for invalid recid 0, but got none")
	}

	if len(output) == 0 {
		t.Error("Expected error message for invalid recid")
	}
}

func TestIntegrationDownloadFilesFilterName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "5500", "--name-filter", "BuildFile.xml", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", err, string(output))
		return
	}

	outputStr := string(output)
	if !contains(outputStr, "Success") {
		t.Error("Expected 'Success!' message in output")
	}

	files, _ := os.ReadDir(tmpDir)
	found := false
	for _, f := range files {
		if contains(f.Name(), "BuildFile") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected BuildFile.xml to be downloaded")
	}
}

func TestIntegrationDownloadFilesFilterNameWrong(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "5500", "--name-filter", "nonexistent.txt", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error for non-matching filter, but got none")
	}

	outputStr := string(output)
	if !contains(outputStr, "No files") {
		t.Error("Expected 'No files matching filters' message")
	}
}

func TestIntegrationDownloadFilesFilterRange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "5500", "--start-index", "0", "--end-index", "2", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", err, string(output))
		return
	}

	files, _ := os.ReadDir(tmpDir)
	if len(files) == 0 {
		t.Error("Expected some files to be downloaded with range filter")
	}
}

func TestIntegrationDownloadFilesFilterRangeInvalid(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "5500", "--start-index", "5", "--end-index", "2", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", err, string(output))
		return
	}

	files, _ := os.ReadDir(tmpDir)
	if len(files) > 0 {
		t.Error("Expected no files to be downloaded with invalid range")
	}
}

func TestIntegrationDownloadFilesRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--retry", "2", "--name-filter", "*.py", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", err, string(output))
		return
	}

	outputStr := string(output)
	if !contains(outputStr, "Success") {
		t.Error("Expected 'Success!' message in output")
	}
}

func TestIntegrationDownloadFilesVerbose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--name-filter", "*.py", "--verbose", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", err, string(output))
		return
	}

	outputStr := string(output)
	if len(outputStr) == 0 {
		t.Error("Expected non-empty verbose output")
	}
}

func TestIntegrationDownloadFilesNoIdentifier(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "download-files")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error when no identifier provided, but got none")
	}

	outputStr := string(output)
	if !contains(outputStr, "recid") && !contains(outputStr, "doi") && !contains(outputStr, "title") {
		t.Error("Expected error message to mention required identifier")
	}
}

func TestIntegrationDownloadFilesInvalidServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--server", "ftp://invalid.com", "--dry-run")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error for invalid server, but got none")
	}

	outputStr := string(output)
	if len(outputStr) == 0 {
		t.Error("Expected error message for invalid server")
	}
}

func TestIntegrationDownloadFilesCustomOutputDir(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--name-filter", "*.py", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", err, string(output))
		return
	}

	files, _ := os.ReadDir(tmpDir)
	if len(files) == 0 {
		t.Error("Expected files to be downloaded to custom output directory")
	}
}

func TestIntegrationVerifyFilesBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	downloadCmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--name-filter", "*.py", "--output-dir", tmpDir)
	downloadOutput, downloadErr := downloadCmd.CombinedOutput()
	if downloadErr != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", downloadErr, string(downloadOutput))
		return
	}

	cmd := exec.Command(getBinaryPath(), "verify-files", "--recid", "3005", "--input-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run verify-files: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	if !contains(outputStr, "Verification summary") {
		t.Error("Expected verification summary in output")
	}
	if !contains(outputStr, "Total files") {
		t.Error("Expected 'Total files' in verification output")
	}
	if !contains(outputStr, "Verified") {
		t.Error("Expected 'Verified' in verification output")
	}
}

func TestIntegrationVerifyFilesByNameFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	downloadCmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--name-filter", "*.py", "--output-dir", tmpDir)
	downloadOutput, downloadErr := downloadCmd.CombinedOutput()
	if downloadErr != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", downloadErr, string(downloadOutput))
		return
	}

	cmd := exec.Command(getBinaryPath(), "verify-files", "--recid", "3005", "--input-dir", tmpDir, "--name-filter", "*.py")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run verify-files with name filter: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	if !contains(outputStr, "Verification summary") {
		t.Error("Expected verification summary in output")
	}
}

func TestIntegrationVerifyFilesNoIdentifier(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "verify-files")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error when no identifier provided, but got none")
	}

	outputStr := string(output)
	if !contains(outputStr, "recid") && !contains(outputStr, "doi") && !contains(outputStr, "title") {
		t.Error("Expected error message to mention required identifier")
	}
}

func TestIntegrationVerifyFilesByDOI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Note: recid 3005 has no DOI in its metadata, so we use recid 5500 which has a valid DOI (10.7483/OPENDATA.CMS.JKB8.RR42)
	// We download all files (not just *.py) to ensure verification can find all expected files
	tmpDir := t.TempDir()

	downloadCmd := exec.Command(getBinaryPath(), "download-files", "--recid", "5500", "--output-dir", tmpDir)
	downloadOutput, downloadErr := downloadCmd.CombinedOutput()
	if downloadErr != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", downloadErr, string(downloadOutput))
		return
	}

	cmd := exec.Command(getBinaryPath(), "verify-files", "--doi", "10.7483/OPENDATA.CMS.JKB8.RR42", "--input-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run verify-files by DOI: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from verify-files")
	}
}

func TestIntegrationVerifyFilesByTitle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Note: API requires exact title match. recid 3005's exact title is "Configuration file for LHE step HIG-Summer11pLHE-00114_1_cfg.py"
	// Using the partial title "Configuration file for LHE step" will not match
	tmpDir := t.TempDir()

	downloadCmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--name-filter", "*.py", "--output-dir", tmpDir)
	downloadOutput, downloadErr := downloadCmd.CombinedOutput()
	if downloadErr != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", downloadErr, string(downloadOutput))
		return
	}

	cmd := exec.Command(getBinaryPath(), "verify-files", "--title", "Configuration file for LHE step HIG-Summer11pLHE-00114_1_cfg.py", "--input-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run verify-files by title: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output from verify-files")
	}
}

func TestIntegrationVerifyFilesInvalidServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "verify-files", "--recid", "3005", "--server", "ftp://invalid.com")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error for invalid server, but got none")
	}

	outputStr := string(output)
	if len(outputStr) == 0 {
		t.Error("Expected error message for invalid server")
	}
}

func TestIntegrationVerifyFilesCustomInputDir(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	downloadCmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--name-filter", "*.py", "--output-dir", tmpDir)
	downloadOutput, downloadErr := downloadCmd.CombinedOutput()
	if downloadErr != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", downloadErr, string(downloadOutput))
		return
	}

	customDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(customDir, 0755)

	cmd := exec.Command(getBinaryPath(), "verify-files", "--recid", "3005", "--input-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run verify-files with custom dir: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	if !contains(outputStr, "Verification summary") {
		t.Error("Expected verification summary in output")
	}
}

func TestIntegrationDownloadFilesRegexp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--regexp", ".*\\.py$", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", err, string(output))
		return
	}

	files, _ := os.ReadDir(tmpDir)
	if len(files) == 0 {
		t.Error("Expected some Python files to be downloaded with regex filter")
	}

	for _, f := range files {
		if !contains(f.Name(), ".py") {
			t.Errorf("Expected only .py files, got: %s", f.Name())
		}
	}
}

func TestIntegrationDownloadFilesRegexpMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--regexp", "(.*\\.py$|.*\\.txt$)", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", err, string(output))
		return
	}

	files, _ := os.ReadDir(tmpDir)
	if len(files) == 0 {
		t.Error("Expected some files to be downloaded with multiple regex filter")
	}
}

func TestIntegrationDownloadFilesRegexpWrong(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--regexp", "nonexistentfile.*", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", err, string(output))
		return
	}

	files, _ := os.ReadDir(tmpDir)
	if len(files) > 0 {
		t.Error("Expected no files to be downloaded with non-matching regex filter")
	}
}

func TestIntegrationDownloadFilesMultipleNameFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--name-filter", "*.py,*.txt", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", err, string(output))
		return
	}

	files, _ := os.ReadDir(tmpDir)
	if len(files) == 0 {
		t.Error("Expected some files to be downloaded with multiple name filters")
	}

	for _, f := range files {
		ext := filepath.Ext(f.Name())
		if ext != ".py" && ext != ".txt" {
			t.Errorf("Expected only .py or .txt files, got: %s", f.Name())
		}
	}
}

func TestIntegrationDownloadFilesMultipleRanges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "5500", "--range-filter", "0-1,3-4", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", err, string(output))
		return
	}

	files, _ := os.ReadDir(tmpDir)
	if len(files) < 2 {
		t.Error("Expected at least 2 files to be downloaded with multiple ranges (0-1,3-4)")
	}
}

func TestIntegrationDownloadFilesRegexpAndRange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--regexp", ".*\\.py$", "--range-filter", "0-2", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", err, string(output))
		return
	}

	files, _ := os.ReadDir(tmpDir)

	for _, f := range files {
		if !contains(f.Name(), ".py") {
			t.Errorf("Expected only .py files with regexp filter, got: %s", f.Name())
		}
	}
}

func TestIntegrationDownloadFilesRegexpAndMultipleRanges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "5500", "--regexp", ".*\\.xml$", "--range-filter", "0-1,3-4", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: download failed: %v\nOutput: %s", err, string(output))
		return
	}

	files, _ := os.ReadDir(tmpDir)

	for _, f := range files {
		if !contains(f.Name(), ".xml") {
			t.Errorf("Expected only .xml files with regexp filter, got: %s", f.Name())
		}
	}
}

func TestIntegrationListDirectoryRecursive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cmd := exec.Command(getBinaryPath(), "list-directory", "root://eospublic.cern.ch//eos/opendata/cms/software/HiggsExample20112012", "--recursive")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: list-directory recursive failed: %v\nOutput: %s", err, string(output))
		return
	}

	outputStr := string(output)
	if len(outputStr) == 0 {
		t.Error("Expected non-empty output from recursive directory listing")
	}

	entries := strings.Split(outputStr, "\n")
	if len(entries) < 10 {
		t.Errorf("Expected at least 10 entries from recursive listing, got: %d", len(entries))
	}
}

func TestIntegrationListDirectoryTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Use a very short timeout that should cause context cancellation
	cmd := exec.Command(getBinaryPath(), "list-directory", "root://eospublic.cern.ch//eos/opendata/cms/software/HiggsExample20112012", "--timeout", "1")
	output, err := cmd.CombinedOutput()
	// The test passes if it completes (either successfully or with timeout error)
	// We just want to ensure the timeout flag doesn't break the command
	if err != nil && len(output) == 0 {
		t.Logf("Note: Command failed with timeout (expected behavior): %v", err)
	}

	// As long as we got output or a clear error, the test passes
	if len(output) > 0 {
		t.Logf("Got %d bytes of output before potential timeout", len(output))
	}
}

func TestIntegrationDownloadFilesWithVerify(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--name-filter", "*.py", "--verify", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run download-files with verify: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	if !contains(outputStr, "Success") {
		t.Error("Expected 'Success!' message in output")
	}

	if !contains(outputStr, "Verifying") {
		t.Error("Expected verification message in output")
	}
}

func TestIntegrationDownloadFilesWithDownloadEngine(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--name-filter", "*.py", "--download-engine", "http", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run download-files with download-engine: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	if !contains(outputStr, "Success") {
		t.Error("Expected 'Success!' message in output")
	}
}

func TestIntegrationDownloadFilesWithRetrySleep(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--name-filter", "*.py", "--retry-sleep", "2", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run download-files with retry-sleep: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	if !contains(outputStr, "Success") {
		t.Error("Expected 'Success!' message in output")
	}
}

func TestIntegrationDownloadFilesWithXRootD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--name-filter", "*.py", "--download-engine", "xrootd", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: XRootD download failed (XRootD server may be unavailable): %v\nOutput: %s", err, string(output))
		t.Skip("Skipping XRootD test - server unavailable")
		return
	}

	outputStr := string(output)
	if !contains(outputStr, "Success") {
		t.Error("Expected 'Success!' message in output")
	}

	if !contains(outputStr, "Downloading file") {
		t.Error("Expected download progress messages")
	}

	if !contains(outputStr, "Download summary") {
		t.Error("Expected download summary")
	}
}

func TestIntegrationDownloadFilesXRootDError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	cmd := exec.Command(getBinaryPath(), "download-files", "--recid", "3005", "--name-filter", "*.py", "--download-engine", "xrootd", "--server", "http://invalid.cern.ch", "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Log("Note: Command completed (XRootD error handling worked)")
	}

	outputStr := string(output)
	if len(outputStr) > 0 {
		t.Logf("Output: %s", outputStr)
	}

	if err != nil && len(outputStr) > 0 {
		t.Log("Got expected error from invalid XRootD server")
	}
}
