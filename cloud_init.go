package cloudinit

// User holds the configuration for a user.
//
//nolint:tagliatelle // This format is required by the cloud-init metadata.
type User struct {
	// Name is the username.
	Name string `yaml:"name"`
	// Groups is a comma separated list of groups the user should be added to.
	Groups string `yaml:"groups"`
	// Shell is the shell the user should use.
	Shell string `yaml:"shell,omitempty"`
	// Sudo is a list of sudo rules for the user.
	Sudo string `yaml:"sudo,omitempty"`
	// AuthorizedKeys is a list of SSH public keys for the user.
	AuthorizedKeys []string `yaml:"ssh_authorized_keys,omitempty"`
	// Password is the password hash for the user.
	Password string `yaml:"passwd,omitempty"`
	// LockPassword is a flag to lock the password.
	LockPassword bool `yaml:"lock_passwd"`
	// EnableSSHPasswordAuth if true the user can login using password over SSH.
	// Otherwise only Public Key authentication is enabled
	EnableSSHPasswordAuth bool `yaml:"ssh_pwauth,omitempty"`
}

// PasswordChange holds the configuration for password change on first boot.
type PasswordChange struct {
	// Expire is a flag to force password change on first boot.
	Expire bool `yaml:"expire"`
	// List is a list of users to force password change.
	List []string `yaml:"list"`
}

// CloudConfig holds the configuration for cloud-init.
//
//nolint:tagliatelle // This format is required by the cloud-init metadata.
type CloudConfig struct {
	// Users is a list of users to create.
	Users []*User `yaml:"users,omitempty"`
	// PasswordChange is the configuration for password change on first boot.
	PasswordChange PasswordChange `yaml:"chpasswd,omitempty"`
	// PackageUpdate is a flag to update the package list.
	PackageUpdate bool `yaml:"package_update,omitempty"`
	// PackageUpgrade is a flag to upgrade the packages.
	PackageUpgrade bool `yaml:"package_upgrade,omitempty"`
	// RunCommands is a list of commands to run.
	RunCommands []string `yaml:"runcmd,omitempty"`
	// TODO: write_files
	// Packages is a list of packages to install.
	Packages []string `yaml:"packages,omitempty"`
	// Timezone is the timezone to set.
	Timezone string `yaml:"timezone,omitempty"`
	// EnableSSHPasswordAuth is a flag to enable SSH password authentication for the root user.
	EnableSSHPasswordAuth bool `yaml:"ssh_pwauth,omitempty"`
	// Hostname is the hostname to set.
	Hostname string `yaml:"hostname,omitempty"`
	// Locale is the locale to set.
	Locale string `yaml:"locale,omitempty"`
	// Mounts is a list of mounts to configure.
	Mounts [][]string `yaml:"mounts,omitempty"`
	// DisableRoot is a flag to disable root login.
	DisableRoot bool `yaml:"disable_root,omitempty"`
	// Growpart is the configuration for growpart.
	Growpart *GrowpartConfig `yaml:"growpart,omitempty"`
}

// Metadata holds the metadata for cloud-init.
//
//nolint:tagliatelle // This format is required by the cloud-init metadata.
type Metadata struct {
	// InstanceID is the instance ID.
	InstanceID string `yaml:"instance-id"`
	// LocalHostname is the hostname of the instance.
	LocalHostname string `yaml:"local-hostname"`
}

type GrowpartConfig struct {
	Mode    string   `yaml:"mode,omitempty"`
	Devices []string `yaml:"devices,omitempty"`
}
