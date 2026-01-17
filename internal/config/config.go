package config

const (
	ServerHTTPURI  = "http://opendata.cern.ch"
	ServerHTTPSURI = "https://opendata.cern.ch"
	ServerRootURI  = "root://eospublic.cern.ch//"

	ListDirectoryTimeout = 60

	DownloadRetryLimit = 10
	DownloadRetrySleep = 5

	DownloadErrorPageSize     = 3846
	DownloadErrorPageChecksum = "adler32:a82d5324"
)
