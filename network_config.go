package cloudinit

type NetworkConfigFile struct {
	Network Network `yaml:"network"`
}

type Network struct {
	Version int             `yaml:"version"`
	Config  []NetworkConfig `yaml:"config"`
}

type NetworkConfigType string

const (
	NetworkConfigTypePhysical   NetworkConfigType = "physical"
	NetworkConfigTypeNameserver NetworkConfigType = "nameserver"
)

type NetworkConfig struct {
	Type       NetworkConfigType `yaml:"type"`
	Name       string            `yaml:"name"`
	MACAddress string            `yaml:"mac_address"`
	Subnets    []Subnet          `yaml:"subnets"`
}

type SubnetType string

const (
	SubnetTypeDHCP   SubnetType = "dhcp"
	SubnetTypeStatic SubnetType = "static"
)

type Subnet struct {
	// Type can be static or dhcp
	Type SubnetType `yaml:"type"`
	// Address is a network address in CIDR format
	Address string `yaml:"address"`
	// Gateway address.
	Gateway     string   `yaml:"gateway"`
	Nameservers []string `yaml:"dns_nameservers"`
	DNSSearch   []string `yaml:"dns_search"`
}
