package helpers

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func CreateMultipart(m map[string]any, w *multipart.Writer) (err error) {
	for key, value := range m {
		var reader io.Reader
		var part io.Writer

		switch v := value.(type) {
		case *os.File:
			baseName := filepath.Base(v.Name())
			reader = v
			part, err = w.CreateFormFile(key, baseName)
		case string:
			part, err = w.CreateFormField(key)
			reader = strings.NewReader(v)
		case int:
			part, err = w.CreateFormField(key)
			reader = strings.NewReader(strconv.Itoa(v))
		case int64:
			part, err = w.CreateFormField(key)
			reader = strings.NewReader(strconv.FormatInt(v, 10))
		case float64:
			part, err = w.CreateFormField(key)
			reader = strings.NewReader(fmt.Sprintf("%.6g", v))
		case bool:
			part, err = w.CreateFormField(key)
			reader = strings.NewReader(fmt.Sprintf("%v", v))
		default:
			return fmt.Errorf("unsupported muiltipart/form parameter %s", v)
		}

		if err != nil {
			return fmt.Errorf("failed to create part: %w", err)
		}

		if _, err = io.Copy(part, reader); err != nil {
			return fmt.Errorf("failed to copy from reader: %w", err)
		}
	}

	return nil
}
