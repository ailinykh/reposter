package helpers

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func MultipartFrom(fields map[string]io.Reader, writer *multipart.Writer) (err error) {
	for key, reader := range fields {
		var part io.Writer

		switch r := reader.(type) {
		case *os.File:
			baseName := filepath.Base(r.Name())
			part, err = writer.CreateFormFile(key, baseName)
		default:
			part, err = writer.CreateFormField(key)
		}

		if err != nil {
			return fmt.Errorf("failed to create writer: %w", err)
		}

		if _, err = io.Copy(part, reader); err != nil {
			return fmt.Errorf("failed to copy from reader: %w", err)
		}
	}

	return nil
}
