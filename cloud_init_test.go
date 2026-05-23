package cloudinit_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cloudinit "go.pilab.hu/cloud/cloud-init"

	"github.com/kdomanski/iso9660"

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

const userDataFileContent = `#cloud-config
users:
    - name: nev3rkn0wn
      groups: sudo,wheel,admin,forestgump
      shell: /bin/bash
      sudo: ALL=(ALL) NOPASSWD:ALL
      ssh_authorized_keys:
        - ssh-rsa AAAAB3NzaC1yc2EA... nn0wn@pm.me
      lock_passwd: false
chpasswd:
    expire: false
    list:
        - root:rootpassword
`

const metaDataFileContent = `instance-id: testinstance
local-hostname: testinstance.newdomain.com
`

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

	fh, err := os.Open("cloud-init.iso")
	if err != nil {
		t.Fatalf("could not open cloud-init.iso: %v", err)
	}
	defer fh.Close()
	image, err := iso9660.OpenImage(fh)
	if err != nil {
		t.Fatalf("could not open iso: %v", err)
	}
	rd, err := image.RootDir()
	if err != nil {
		t.Fatalf("could not open root directory: %v", err)
	}

	files, err := rd.GetChildren()
	if err != nil {
		t.Fatalf("could not get root dir files: %v", err)
	}

	for _, file := range files {
		t.Logf("file %s", file.Name())
		bs, err := io.ReadAll(file.Reader())
		if err != nil {
			t.Fatalf("could not open %s: %v", file.Name(), err)
		}
		content := string(bs)

		if file.Name() == "user-data" {
			lines := strings.Split(content, "\n")
			var contentWithoutPassword string
			for _, line := range lines {
				if strings.Contains(line, "passwd: $") {
					continue
				}
				if contentWithoutPassword == "" {
					contentWithoutPassword = line
				} else {
					contentWithoutPassword = fmt.Sprintf("%s\n%s", contentWithoutPassword, line)
				}
			}
			if contentWithoutPassword != userDataFileContent {
				t.Fatalf("expected:\n%s\n----\ngot:\n%s", userDataFileContent, contentWithoutPassword)
			}
		} else if file.Name() == "meta-data" {
			t.Logf("file content: %s", string(bs))

			if content != metaDataFileContent {
				t.Fatalf("expected:\n%s\n----\ngot:\n%s", metaDataFileContent, content)
			}
		}
	}

	t.Log("Test successful")
}

func TestCloudInitConfigWriteFiles(t *testing.T) {
	c := cloudinit.NewConfig()

	content := c.GenerateConfigContent()
	assert.Contains(t, string(content), "write_files:")
	assert.Contains(t, string(content), "/etc/myapp/config.json")
}

func TestCloudInitStorageConfig(t *testing.T) {
	c := cloudinit.NewConfig()

	// c.ConfigureStorage([]string{"/", "/dev/vda1"})

	content := c.GenerateConfigContent()
	assert.Contains(t, string(content), "growpart:")
	assert.Contains(t, string(content), "/dev/vda1")
}
