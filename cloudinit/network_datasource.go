package cloudinit

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// NetworkDataSourceConfig represents the configuration for a network data source.
type NetworkDataSourceConfig struct {
	URL string
}

// GenerateMetadata generates the metadata for the network data source.
func (n *NetworkDataSourceConfig) GenerateMetadata() ([]byte, error) {
	if n.URL == "" {
		return nil, fmt.Errorf("URL is not configured for NetworkDataSourceConfig")
	}
	metadataURL := n.URL + "/meta-data"

	resp, err := http.Get(metadataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata from %s: %w", metadataURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code (%d) from %s", resp.StatusCode, metadataURL)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata response body from %s: %w", metadataURL, err)
	}

	return body, nil
}

// GenerateUserData generates the user-data for the network data source.
func (n *NetworkDataSourceConfig) GenerateUserData() ([]byte, error) {
	if n.URL == "" {
		return nil, fmt.Errorf("URL is not configured for NetworkDataSourceConfig")
	}
	userDataURL := n.URL + "/user-data"

	resp, err := http.Get(userDataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch userdata from %s: %w", userDataURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code (%d) from %s", resp.StatusCode, userDataURL)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read userdata response body from %s: %w", userDataURL, err)
	}

	return body, nil
}

// GetVolumeName returns the volume name for the network data source.
// For network sources, this often defaults to "config-2" or is derived from metadata.
func (n *NetworkDataSourceConfig) GetVolumeName() string {
	return "config-2"
}

// GetFilePaths returns the file paths for the network data source.
// This can be an empty map if no specific files are written to disk by default.
func (n *NetworkDataSourceConfig) GetFilePaths() map[string]string {
	return make(map[string]string)
}

// NewNetworkDataSourceConfig creates a new NetworkDataSourceConfig with the given URL.
func NewNetworkDataSourceConfig(url string) (*NetworkDataSourceConfig, error) {
	if url == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}
	return &NetworkDataSourceConfig{URL: url}, nil
}
