package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/ailinykh/reposter/v3/pkg/helpers"
)

func (b *Bot) SendVideoMultipart(m map[string]any) (*Message, error) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	if err := helpers.CreateMultipart(m, w); err != nil {
		return nil, fmt.Errorf("failed to create muiltipart data: %w", err)
	}
	w.Close()

	url := b.endpoint + "/bot" + b.token + "/sendVideo"
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create muiltipart request: %w", err)
	}
	req.Header.Add("Content-Type", w.FormDataContentType())

	res, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform muiltipart request: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var i struct {
		Result *Message `json:"result"`
	}
	if err := json.Unmarshal(data, &i); err != nil {
		return nil, fmt.Errorf("failed to unmarshall message: %w", err)
	}
	return i.Result, nil
}
