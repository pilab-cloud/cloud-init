package cloudinit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/kdomanski/iso9660"
)

// GCEMetadata represents Google Compute Engine instance metadata
// For more information see: https://cloud.google.com/compute/docs/metadata/overview
type GCEMetadata struct {
	// Instance contains instance-specific metadata
	Instance struct {
		// ID is the unique identifier for the instance
		ID string `json:"id"`

		// Name is the instance name
		Name string `json:"name"`

		// Hostname is the instance hostname
		Hostname string `json:"hostname"`

		// Zone is the GCP zone where the instance is running
		Zone string `json:"zone"`

		// MachineType is the GCE machine type
		MachineType string `json:"machineType"`

		// Tags are the network tags associated with the instance
		Tags []string `json:"tags,omitempty"`

		// Labels are the instance labels
		Labels map[string]string `json:"labels,omitempty"`

		// ServiceAccounts contains the service account information
		ServiceAccounts []GCEServiceAccount `json:"serviceAccounts,omitempty"`

		// NetworkInterfaces contains the network configuration
		NetworkInterfaces []GCENetworkInterface `json:"networkInterfaces,omitempty"`

		// CreatedAt is the instance creation timestamp
		CreatedAt time.Time `json:"createdAt"`
	} `json:"instance"`

	// Project contains project-specific metadata
	Project struct {
		// ProjectID is the GCP project identifier
		ProjectID string `json:"projectId"`

		// ProjectNumber is the numeric project identifier
		ProjectNumber string `json:"projectNumber"`
	} `json:"project"`
}

// GCEServiceAccount represents a GCP service account configuration
type GCEServiceAccount struct {
	// Email is the service account email address
	Email string `json:"email"`

	// Scopes are the OAuth scopes granted to the service account
	Scopes []string `json:"scopes"`
}

// GCENetworkInterface represents a GCE network interface configuration
type GCENetworkInterface struct {
	// Network is the VPC network name
	Network string `json:"network"`

	// Subnetwork is the subnet name
	Subnetwork string `json:"subnetwork"`

	// NetworkIP is the internal IP address
	NetworkIP string `json:"networkIP"`

	// AccessConfigs contains external access configurations
	AccessConfigs []GCEAccessConfig `json:"accessConfigs,omitempty"`
}

// GCEAccessConfig represents external access configuration for a network interface
type GCEAccessConfig struct {
	// Type is the access configuration type
	Type string `json:"type"`

	// Name is the access configuration name
	Name string `json:"name"`
}

// New constructor for GCE
func NewGCEConfig() *Config {
	c := NewConfig()
	c.dataSourceType = "gce"
	c.gceMetadata = &GCEMetadata{}
	return c
}

// GCE-specific methods
func (c *Config) SetGCEMetadata(instanceName, zone, projectID string) {
	if c.gceMetadata == nil {
		c.gceMetadata = &GCEMetadata{}
	}
	c.gceMetadata.Instance.Name = instanceName
	c.gceMetadata.Instance.Zone = zone
	c.gceMetadata.Project.ProjectID = projectID
	c.gceMetadata.Instance.CreatedAt = time.Now().UTC()
}

func (c *Config) AddGCELabel(key, value string) {
	if c.gceMetadata.Instance.Labels == nil {
		c.gceMetadata.Instance.Labels = make(map[string]string)
	}
	c.gceMetadata.Instance.Labels[key] = value
}

func (c *Config) writeGCEISO(w io.Writer) error {
	writer, err := iso9660.NewWriter()
	if err != nil {
		return fmt.Errorf("failed to create ISO writer: %w", err)
	}
	defer writer.Cleanup()

	// GCE metadata structure
	metadataJSON, err := json.Marshal(c.gceMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal GCE metadata: %w", err)
	}

	// GCE expects files in a specific structure
	if err := writer.AddFile(bytes.NewReader(metadataJSON),
		"computeMetadata/v1/instance/attributes.json"); err != nil {
		return fmt.Errorf("failed to add GCE metadata: %w", err)
	}

	if err := writer.AddFile(bytes.NewReader(c.GenerateConfigContent()),
		"user-data"); err != nil {
		return fmt.Errorf("failed to add user-data: %w", err)
	}

	// Add network configuration if present
	if len(c.networkInterfaces) > 0 {
		networkConfig := c.generateGCENetworkConfig()
		if err := writer.AddFile(bytes.NewReader(networkConfig),
			"network-config"); err != nil {
			return fmt.Errorf("failed to add network-config: %w", err)
		}
	}

	if err := writer.WriteTo(w, "google-compute-engine"); err != nil {
		return fmt.Errorf("failed to write GCE ISO image: %w", err)
	}

	return nil
}

func (c *Config) generateGCENetworkConfig() []byte {
	// GCE network configuration format
	type gceNetwork struct {
		Version int `json:"version"`
		Config  []struct {
			Type       string `json:"type"`
			Name       string `json:"name"`
			MacAddress string `json:"mac_address,omitempty"`
			Subnets    []struct {
				Type    string   `json:"type"`
				Address string   `json:"address,omitempty"`
				Gateway string   `json:"gateway,omitempty"`
				DNS     []string `json:"dns_nameservers,omitempty"`
			} `json:"subnets,omitempty"`
		} `json:"config"`
	}

	net := gceNetwork{
		Version: 1,
		Config: make([]struct {
			Type       string `json:"type"`
			Name       string `json:"name"`
			MacAddress string `json:"mac_address,omitempty"`
			Subnets    []struct {
				Type    string   `json:"type"`
				Address string   `json:"address,omitempty"`
				Gateway string   `json:"gateway,omitempty"`
				DNS     []string `json:"dns_nameservers,omitempty"`
			} `json:"subnets,omitempty"`
		}, 0),
	}

	// Convert our network interfaces to GCE format
	for mac, iface := range c.networkInterfaces {
		netConfig := struct {
			Type       string `json:"type"`
			Name       string `json:"name"`
			MacAddress string `json:"mac_address,omitempty"`
			Subnets    []struct {
				Type    string   `json:"type"`
				Address string   `json:"address,omitempty"`
				Gateway string   `json:"gateway,omitempty"`
				DNS     []string `json:"dns_nameservers,omitempty"`
			} `json:"subnets,omitempty"`
		}{
			Type:       "physical",
			Name:       "eth0",
			MacAddress: mac,
		}

		subnet := struct {
			Type    string   `json:"type"`
			Address string   `json:"address,omitempty"`
			Gateway string   `json:"gateway,omitempty"`
			DNS     []string `json:"dns_nameservers,omitempty"`
		}{
			Type:    "static",
			Address: iface.Address,
			Gateway: iface.Gateway,
			DNS:     iface.Nameservers,
		}

		netConfig.Subnets = append(netConfig.Subnets, subnet)
		net.Config = append(net.Config, netConfig)
	}

	data, _ := json.Marshal(net)
	return data
}
