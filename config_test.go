package cloudinit_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	cloudinit "go.pilab.hu/cloud/cloud-init/v1"
)

func TestCloudInitConfig_AddUser(t *testing.T) {
	// TestCloudInitConfig_AddUser tests the AddUser function.
	t.Log("TestCloudInitConfig_AddUser")

	// Create a new CloudInitConfig
	c := cloudinit.NewConfig()

	// Add a user
	c.AddUser(cloudinit.User{
		Name:           "test",
		Groups:         "sudo",
		Shell:          "/bin/bash",
		Sudo:           "",
		AuthorizedKeys: nil,
		Password:       "test123",
		LockPassword:   false,
	})

	bb := c.GenerateConfigContent()
	assert.NotEmpty(t, bb)

	t.Log("TestCloudInitConfig_AddUser: success", string(bb))
}

func TestEncryptPassword(t *testing.T) {
	// TestEncryptPassword tests the EncryptPassword function.
	t.Log("TestEncryptPassword")

	// Encode a password
	p := cloudinit.EncryptPassword("test123")
	assert.NotEmpty(t, p)

	t.Log("TestEncryptPassword: success", p)
}
