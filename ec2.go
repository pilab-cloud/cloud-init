package cloudinit

// EC2Metadata represents the EC2-specific metadata structure
// For more information see: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-categories.html
type EC2Metadata struct {
	// InstanceID is the EC2 instance identifier
	InstanceID string `json:"instance-id"`

	// LocalHostname is the local hostname of the instance
	LocalHostname string `json:"local-hostname"`

	// PublicHostname is the public DNS hostname of the instance
	PublicHostname string `json:"public-hostname,omitempty"`

	// PublicIPv4 is the public IPv4 address of the instance
	PublicIPv4 string `json:"public-ipv4,omitempty"`

	// LocalIPv4 is the private IPv4 address of the instance
	LocalIPv4 string `json:"local-ipv4,omitempty"`

	// AvailabilityZone is the AZ where the instance is running
	AvailabilityZone string `json:"availability-zone,omitempty"`

	// InstanceType is the EC2 instance type
	InstanceType string `json:"instance-type,omitempty"`

	// Tags are the instance tags
	Tags map[string]string `json:"tags,omitempty"`
}

// EC2Config implements the DataSourceConfig interface for EC2
type EC2Config struct {
	*Config
	metadata EC2Metadata
}

// NewEC2Config creates a new EC2-specific configuration
func NewEC2Config() *Config {
	c := NewConfig()
	c.dataSourceType = string(DataSourceEC2)
	return c
}
