// Package cloudinit provides cloud-init configuration generation for various cloud providers
package cloudinit

// DataSourceType represents the type of cloud-init data source
type DataSourceType string

const (
	// DataSourceNoCloud represents the NoCloud data source type
	// Used for local ISO and network-based provisioning
	DataSourceNoCloud DataSourceType = "nocloud"

	// DataSourceEC2 represents the Amazon EC2 data source type
	DataSourceEC2 DataSourceType = "ec2"

	// DataSourceGCE represents the Google Compute Engine data source type
	DataSourceGCE DataSourceType = "gce"

	// DataSourceConfigDrive represents the OpenStack ConfigDrive data source type
	DataSourceConfigDrive DataSourceType = "configdrive"
)

// DataSourceConfig is the interface that all data source configurations must implement
type DataSourceConfig interface {
	// GenerateMetadata generates the metadata content for the data source
	GenerateMetadata() ([]byte, error)

	// GetVolumeName returns the ISO volume name for the data source
	GetVolumeName() string

	// GetFilePaths returns the required file paths for the data source
	GetFilePaths() map[string]string
}
