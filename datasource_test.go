package cloudinit_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudinit "go.pilab.hu/cloud/cloud-init"
	"gopkg.in/yaml.v2"
)

func TestGCEConfig(t *testing.T) {
	t.Run("basic configuration", func(t *testing.T) {
		c := cloudinit.NewGCEConfig()
		c.SetGCEMetadata("test-instance", "us-central1-a", "test-project")

		// Verify metadata
		content := c.GenerateMetadataContent()
		var metadata map[string]interface{}
		err := json.Unmarshal(content, &metadata)
		require.NoError(t, err)

		instance := metadata["instance"].(map[string]interface{})
		assert.Equal(t, "test-instance", instance["name"])
		assert.Equal(t, "us-central1-a", instance["zone"])

		project := metadata["project"].(map[string]interface{})
		assert.Equal(t, "test-project", project["projectId"])
	})

	t.Run("with labels and network", func(t *testing.T) {
		c := cloudinit.NewGCEConfig()
		c.SetGCEMetadata("test-instance", "us-central1-a", "test-project")
		c.AddGCELabel("env", "test")
		c.AddGCELabel("team", "dev")

		c.SetStaticInterfaceAddress(
			"42:01:0a:8a:00:2b",
			"10.0.0.10/24",
			"10.0.0.1",
			"8.8.8.8",
		)

		// Write ISO and verify
		f, err := os.CreateTemp("", "gce-test-*.iso")
		require.NoError(t, err)
		defer os.Remove(f.Name())

		err = c.WriteISO(f)
		require.NoError(t, err)

		assert.FileExists(t, f.Name())
	})
}

func TestEC2Config(t *testing.T) {
	t.Run("basic configuration", func(t *testing.T) {
		c := cloudinit.NewEC2Config()
		c.SetEC2Metadata("i-1234567890", "us-east-1a", map[string]string{
			"Name": "test-instance",
		})

		// Verify metadata
		content := c.GenerateMetadataContent()
		var metadata map[string]interface{}
		err := json.Unmarshal(content, &metadata)
		require.NoError(t, err)

		assert.Equal(t, "i-1234567890", metadata["instance-id"])
		assert.Equal(t, "us-east-1a", metadata["availability-zone"])
	})

	t.Run("with network configuration", func(t *testing.T) {
		c := cloudinit.NewEC2Config()
		c.SetEC2Metadata("i-1234567890", "us-east-1a", nil)

		c.SetStaticInterfaceAddress(
			"0e:49:61:0f:c3:11",
			"172.31.16.100/20",
			"172.31.16.1",
			"169.254.169.253",
		)

		// Write ISO and verify
		f, err := os.CreateTemp("", "ec2-test-*.iso")
		require.NoError(t, err)
		defer os.Remove(f.Name())

		err = c.WriteISO(f)
		require.NoError(t, err)

		assert.FileExists(t, f.Name())
	})
}

func TestDataSourceCompatibility(t *testing.T) {
	testCases := []struct {
		name     string
		newFunc  func() *cloudinit.Config
		metadata map[string]string
	}{
		{
			name:    "NoCloud",
			newFunc: cloudinit.NewConfig,
			metadata: map[string]string{
				"instance-id": "test-nocloud",
				"hostname":    "test.local",
			},
		},
		{
			name:    "EC2",
			newFunc: cloudinit.NewEC2Config,
			metadata: map[string]string{
				"instance-id":       "i-test123",
				"availability-zone": "us-east-1a",
			},
		},
		{
			name:    "GCE",
			newFunc: cloudinit.NewGCEConfig,
			metadata: map[string]string{
				"name": "test-instance",
				"zone": "us-central1-a",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.newFunc()

			// Add common configuration
			c.AddUser(cloudinit.User{
				Name:   "test-user",
				Groups: "sudo",
				Shell:  "/bin/bash",
			})

			c.SetStaticInterfaceAddress(
				"00:11:22:33:44:55",
				"192.168.1.100/24",
				"192.168.1.1",
				"8.8.8.8",
			)

			// Write ISO
			f, err := os.CreateTemp("", "test-*.iso")
			require.NoError(t, err)
			defer os.Remove(f.Name())

			err = c.WriteISO(f)
			require.NoError(t, err)

			assert.FileExists(t, f.Name())
		})
	}
}

func TestNetworkConfiguration(t *testing.T) {
	testCases := []struct {
		name     string
		config   func() *cloudinit.Config
		networks []struct {
			mac     string
			ip      string
			gateway string
			dns     []string
		}
	}{
		{
			name:   "Single Interface",
			config: cloudinit.NewConfig,
			networks: []struct {
				mac     string
				ip      string
				gateway string
				dns     []string
			}{
				{
					mac:     "00:11:22:33:44:55",
					ip:      "192.168.1.100/24",
					gateway: "192.168.1.1",
					dns:     []string{"8.8.8.8", "8.8.4.4"},
				},
			},
		},
		{
			name:   "Multiple Interfaces",
			config: cloudinit.NewConfig,
			networks: []struct {
				mac     string
				ip      string
				gateway string
				dns     []string
			}{
				{
					mac:     "00:11:22:33:44:55",
					ip:      "192.168.1.100/24",
					gateway: "192.168.1.1",
					dns:     []string{"8.8.8.8"},
				},
				{
					mac:     "00:11:22:33:44:66",
					ip:      "10.0.0.100/24",
					gateway: "10.0.0.1",
					dns:     []string{"10.0.0.2"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.config()

			for _, net := range tc.networks {
				c.SetStaticInterfaceAddress(
					net.mac,
					net.ip,
					net.gateway,
					net.dns...,
				)
			}

			content := c.GenerateConfigContent()
			assert.NotEmpty(t, content)

			// Verify network configuration
			var config map[string]interface{}
			err := yaml.Unmarshal(content, &config)
			require.NoError(t, err)

			network := config["network"].(map[string]interface{})
			assert.Equal(t, 1, network["version"])
			assert.Len(t, network["config"], len(tc.networks))
		})
	}
}
