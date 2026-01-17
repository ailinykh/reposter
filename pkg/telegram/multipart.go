package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

func (b *Bot) rawMultipart(m string, in map[string]any, out any) error {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	if err := makeMultipart(in, w); err != nil {
		return fmt.Errorf("failed to create muiltipart data: %w", err)
	}
	w.Close()

	url := b.endpoint + "/bot" + b.token + "/" + m
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("failed to create muiltipart request: %w", err)
	}
	req.Header.Add("Content-Type", w.FormDataContentType())

	res, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform muiltipart request: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	return json.Unmarshal(data, &out)
}

func makeMultipart(m map[string]any, w *multipart.Writer) (err error) {
	for key, value := range m {
		var reader io.Reader
		var part io.Writer

		switch v := value.(type) {
		case InputFileLocal:
			reader = v.Reader
			part, err = w.CreateFormFile(key, v.Name)
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
