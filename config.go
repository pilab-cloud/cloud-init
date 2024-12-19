package cloudinit

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"strings"

	"github.com/kdomanski/iso9660"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v2"
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
	writer, err := iso9660.NewWriter()
	if err != nil {
		return fmt.Errorf("failed to create writer: %w", err)
	}
	defer writer.Cleanup()

	// TODO: create proper error handling
	_ = writer.AddFile(bytes.NewReader(c.GenerateMetadataContent()), "meta-data")
	_ = writer.AddFile(bytes.NewReader(c.GenerateConfigContent()), "user-data")

	if len(c.networkInterfaces) > 0 {
		_ = writer.AddFile(bytes.NewReader(c.GenerateNetworkConfigContent()), "network-config")
	}

	err = writer.WriteTo(w, VolumeName)
	if err != nil {
		return fmt.Errorf("failed to write ISO image: %w", err)
	}

	return nil
}

func (c *Config) GenerateNetworkConfigContent() []byte {
	ncf := new(NetworkConfigFile)

	ncf.Network.Version = 1

	i := 0
	for k, v := range c.networkInterfaces {
		ncf.Network.Config = append(ncf.Network.Config, NetworkConfig{
			Type:       NetworkConfigTypePhysical,
			Name:       fmt.Sprintf("interface%d", i),
			MACAddress: k,
			Subnets: []Subnet{{
				Type:        SubnetTypeStatic,
				Address:     v.Address,
				Gateway:     v.Gateway,
				Nameservers: v.Nameservers,
				DNSSearch:   nil,
			}},
		})

		i++
	}

	out, _ := yaml.Marshal(ncf)
	return out
}
