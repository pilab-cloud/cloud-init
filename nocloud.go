package cloudinit

import (
	"bytes"
	"fmt"
	"io"

	"github.com/kdomanski/iso9660"
)

func writeFile(writer *iso9660.ImageWriter, isoFilename string, isoFilecontent []byte) error {
	buf := bytes.NewBuffer(isoFilecontent)

	return writer.AddFile(buf, isoFilename)
}

func (c *Config) writeNoCloudISO(w io.Writer) error {
	writer, err := iso9660.NewWriter()
	if err != nil {
		return fmt.Errorf("failed to create writer: %w", err)
	}
	defer writer.Cleanup()

	metaDataFileContent := c.GenerateMetadataContent()
	err = writeFile(writer, "meta-data", metaDataFileContent)
	if err != nil {
		return fmt.Errorf("failed to write meta-data: %w", err)
	}

	userDataFileContent := c.GenerateConfigContent()
	err = writeFile(writer, "user-data", userDataFileContent)
	if err != nil {
		return fmt.Errorf("failed to write user-data: %w", err)
	}

	err = writer.WriteTo(w, "cidata")
	if err != nil {
		return fmt.Errorf("failed to write iso file: %w", err)
	}

	return nil
}
