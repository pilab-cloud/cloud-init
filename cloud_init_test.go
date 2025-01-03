package cloudinit_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	cloudinit "go.pilab.hu/cloud/cloud-init"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

var TestUser = cloudinit.User{
	Name:           "nev3rkn0wn",
	Groups:         "sudo,wheel,admin,forestgump",
	Shell:          "/bin/bash",
	Sudo:           "ALL=(ALL) NOPASSWD:ALL",
	AuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EA... nn0wn@pm.me"},
	Password:       "ultrasecretpassword",
	LockPassword:   false,
}

func TestConfigMarshal(t *testing.T) {
	t.Cleanup(func() {
		t.Log("Cleaning up")

		_ = os.Remove("cloud-config.yaml")
		_ = os.Remove("meta-data")
	})

	cc := cloudinit.CloudConfig{
		Users: []*cloudinit.User{
			&TestUser,
		},
		PasswordChange: cloudinit.PasswordChange{
			Expire: true,
			List: []string{
				"test:asd",
			},
		},
	}

	bb := new(bytes.Buffer)
	_, _ = bb.WriteString("#cloud-config\n")

	_ = yaml.NewEncoder(bb).Encode(cc)

	t.Log("Config", bb.String())

	// TODO: write more tests, not just length tests
	require.Equal(t, 305, bb.Len())
	bb.Reset()

	bb = new(bytes.Buffer)
	_ = yaml.NewEncoder(bb).Encode(cloudinit.Metadata{
		InstanceID:    "testinstance",
		LocalHostname: "testinstance.testdomain",
	})
	require.Equal(t, 66, bb.Len())
}

func TestNewCloudInitConfig(t *testing.T) {
	t.Cleanup(func() {
		t.Log("Cleaning up")

		_ = os.Remove("cloud-init.iso")
	})

	c := cloudinit.NewConfig()
	c.SetRootPassword("rootpassword")
	c.SetFQDN("testinstance.newdomain.com")

	c.AddUser(TestUser)

	c.SetStaticInterfaceAddress("c2:da:53:50:4d:61", "195.199.213.137/27", "195.199.213.254", "8.8.8.8", "8.8.4.4")

	c.SetEC2Metadata("i-01234567890abcdef0", "us-east-1a", map[string]string{"project": "my-project", "env": "prod"})

	_ = c.GenerateMetadataContent()
	_ = c.GenerateConfigContent()

	t.Log("Creating cloud-init.iso")
	f, _ := os.Create("cloud-init.iso")
	assert.NoError(t, c.WriteISO(f))

	assert.FileExists(t, "cloud-init.iso")

	t.Log("Test successful")
}

func TestCloudInitConfigWriteFiles(t *testing.T) {
	c := cloudinit.NewConfig()

	c.AddFile("/etc/myapp/config.json", `{"key": "value"}`, "0644")
	c.AddFile("/etc/myapp/secret", "secret-content", "0600")

	content := c.GenerateConfigContent()
	assert.Contains(t, string(content), "write_files:")
	assert.Contains(t, string(content), "/etc/myapp/config.json")
}

func TestCloudInitStorageConfig(t *testing.T) {
	c := cloudinit.NewConfig()

	c.ConfigureStorage([]string{"/", "/dev/vda1"})

	content := c.GenerateConfigContent()
	assert.Contains(t, string(content), "growpart:")
	assert.Contains(t, string(content), "/dev/vda1")
}
