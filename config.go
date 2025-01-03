package cloudinit

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/kdomanski/iso9660"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

const VolumeName = "cidata"

type Interface struct {
	Address     string
	Gateway     string
	Nameservers []string
}

// Config represents a cloud-init configuration.
type Config struct {
	fqdn         string
	rootPassword string

	networkInterfaces map[string]Interface
	users             []User
	enableGuestAgent  bool
	dataSourceType    string
	ec2Meta           *EC2Metadata
}

func NewConfig() *Config {
	// Generate a random ID
	rb := make([]byte, 10)
	_, _ = rand.Read(rb)

	return &Config{
		fqdn:              fmt.Sprintf("vps-%x.pilab.cloud", rb),
		rootPassword:      "",
		users:             make([]User, 0),
		networkInterfaces: make(map[string]Interface),
		enableGuestAgent:  false,
	}
}

func NewEC2Config() *Config {
	c := NewConfig()
	c.dataSourceType = "ec2"
	c.ec2Meta = &EC2Metadata{}
	return c
}

func (c *Config) SetStaticInterfaceAddress(mac, addr, gateway string, ns ...string) {
	c.networkInterfaces[mac] = Interface{
		Address:     addr,
		Gateway:     gateway,
		Nameservers: ns,
	}
}

// EncryptPassword is a helper function to create a bcrypt password, for the /etc/shadow file.
// This is used when the users are defined in the config, with plaintext password.
func EncryptPassword(password string) string {
	// Generate a salt and hash the password using bcrypt.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "" //, fmt.Errorf("failed to hash password: %w", err)
	}

	// Return the hash as a string suitable for /etc/shadow.
	return fmt.Sprintf("%s", hash)
}

// AddUser adds a user to the cloud-init configuration. If the password is not
// already hashed, it will be hashed with MD5 hash.
func (c *Config) AddUser(user User) {
	if !strings.HasPrefix(user.Password, "$") {
		user.Password = EncryptPassword(user.Password)
	}

	c.users = append(c.users, user)
}

func (c *Config) SetRootPassword(password string) {
	c.rootPassword = password
}

func (c *Config) SetFQDN(fqdn string) {
	c.fqdn = fqdn
}

func (c *Config) EnableGuestAgent() {
	c.enableGuestAgent = true
}

func (c *Config) GenerateMetadataContent() []byte {
	hostAndDomain := strings.SplitN(c.fqdn, ".", 2)

	m := Metadata{
		InstanceID:    hostAndDomain[0],
		LocalHostname: c.fqdn,
	}

	buf := new(bytes.Buffer)
	_ = yaml.NewEncoder(buf).Encode(m)

	return buf.Bytes()
}

func (c *Config) GenerateConfigContent() []byte {
	cc := new(CloudConfig)

	cc.Users = make([]*User, len(c.users))
	for i, user := range c.users {
		cc.Users[i] = &user
	}

	if c.rootPassword != "" {
		cc.PasswordChange = PasswordChange{
			Expire: false,
			List:   []string{"root:" + c.rootPassword},
		}
	}

	buf := new(bytes.Buffer)

	// Write header
	buf.WriteString("#cloud-config\n")

	if c.enableGuestAgent {
		cc.PackageUpdate = true
		cc.Packages = append(cc.Packages, "qemu-guest-agent")
		cc.RunCommands = append(cc.RunCommands, "systemctl enable qemu-guest-agent --now")
	}

	// Write the rest of the data
	_ = yaml.NewEncoder(buf).Encode(cc)

	return buf.Bytes()
}

// WriteISO writes the cloud-init configuration to an ISO image.
func (c *Config) WriteISO(w io.Writer) error {
	switch c.dataSourceType {
	case "ec2":
		return c.writeEC2ISO(w)
	case "gce":
		return c.writeGCEISO(w)
	default:
		return c.writeNoCloudISO(w)
	}
}

func (c *Config) writeEC2ISO(w io.Writer) error {
	writer, err := iso9660.NewWriter()
	if err != nil {
		return fmt.Errorf("failed to create ISO writer: %w", err)
	}
	defer writer.Cleanup()

	// EC2 expects metadata in JSON format
	metadataJSON, err := json.Marshal(c.ec2Meta)
	if err != nil {
		return fmt.Errorf("failed to marshal EC2 metadata: %w", err)
	}

	// EC2 specific paths
	if err := writer.AddFile(bytes.NewReader(metadataJSON), "ec2/latest/meta-data.json"); err != nil {
		return fmt.Errorf("failed to add EC2 metadata: %w", err)
	}

	if err := writer.AddFile(bytes.NewReader(c.GenerateConfigContent()), "ec2/latest/user-data"); err != nil {
		return fmt.Errorf("failed to add user-data: %w", err)
	}

	if len(c.networkInterfaces) > 0 {
		networkData := c.generateEC2NetworkConfig()
		if err := writer.AddFile(bytes.NewReader(networkData), "ec2/latest/network-data.json"); err != nil {
			return fmt.Errorf("failed to add network-data: %w", err)
		}
	}

	if err := writer.WriteTo(w, "ec2-seed"); err != nil {
		return fmt.Errorf("failed to write EC2 ISO image: %w", err)
	}

	return nil
}

func (c *Config) generateEC2NetworkConfig() []byte {
	// Convert our network config to EC2 format
	type ec2Network struct {
		Interfaces []struct {
			MACAddress string   `json:"mac"`
			IPAddress  string   `json:"ip"`
			Gateway    string   `json:"gateway"`
			DNS        []string `json:"dns"`
		} `json:"interfaces"`
	}

	net := ec2Network{
		Interfaces: make([]struct {
			MACAddress string   `json:"mac"`
			IPAddress  string   `json:"ip"`
			Gateway    string   `json:"gateway"`
			DNS        []string `json:"dns"`
		}, 0),
	}

	for mac, iface := range c.networkInterfaces {
		net.Interfaces = append(net.Interfaces, struct {
			MACAddress string   `json:"mac"`
			IPAddress  string   `json:"ip"`
			Gateway    string   `json:"gateway"`
			DNS        []string `json:"dns"`
		}{
			MACAddress: mac,
			IPAddress:  iface.Address,
			Gateway:    iface.Gateway,
			DNS:        iface.Nameservers,
		})
	}

	data, _ := json.Marshal(net)
	return data
}

func (c *Config) SetEC2Metadata(instanceID, az string, tags map[string]string) {
	if c.ec2Meta == nil {
		c.ec2Meta = &EC2Metadata{}
	}
	c.ec2Meta.InstanceID = instanceID
	c.ec2Meta.AvailabilityZone = az
	c.ec2Meta.Tags = tags
}
