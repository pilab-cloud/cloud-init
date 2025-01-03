# CloudInit GO

[![Go Report Card](https://goreportcard.com/badge/go.pilab.hu/cloud/cloud-init)](https://goreportcard.com/report/go.pilab.hu/cloud/cloud-init)
[![GoDoc](https://godoc.org/go.pilab.hu/cloud/cloud-init?status.svg)](https://godoc.org/go.pilab.hu/cloud/cloud-init)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/pilab-hu/cloud-init)](https://go.dev/)

Cloud-Init utilities to create cloud-init ISO to configure cloud images.

## Overview

This project provides a set of utilities for generating cloud-init configurations and ISO images. Cloud-init is a widely used tool for initializing cloud instances, allowing users to customize their cloud environments at boot time.

## Features

- Generate cloud-init configuration files in YAML format.
- Create ISO images containing cloud-init configurations.
- Support for user management, including password and SSH key configuration.
- Ability to configure network settings and run custom commands on first boot.
- File writing support for custom configurations
- Storage configuration capabilities
- Network interface management

## Installation

To install the CloudInit GO package, use the following command:

```bash
go get go.pilab.hu/cloud/cloud-init
```

## Usage

```go
import cloudinit "go.pilab.hu/cloud/cloud-init"

// Create a new configuration
config := cloudinit.NewConfig()

// Add a user
config.AddUser(cloudinit.User{
    Name:           "admin",
    Groups:         "sudo",
    Shell:          "/bin/bash",
    Sudo:           "ALL=(ALL) NOPASSWD:ALL",
    AuthorizedKeys: []string{"ssh-rsa AAAAB..."},
})

// Write configuration to ISO
file, _ := os.Create("cloud-init.iso")
defer file.Close()
config.WriteISO(file)
```

## Contributing

Contributions are welcome! Here's how you can help:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please make sure to update tests as appropriate and follow the existing coding style.

### Development Requirements

- Go 1.23 or higher
- Make (for building and testing)
- Git

### Running Tests

```bash
go test ./... -v
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Thanks to all contributors who have helped shape this project
- Built with Go and ❤️
