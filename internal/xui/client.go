package xui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ailinykh/reposter/v3/pkg/telegram"
	"github.com/google/uuid"
)

const inboundId = 4

func NewClient(l *slog.Logger, baseUrl, login, password string) *Client {
	return &Client{
		l:       l,
		baseUrl: strings.TrimSuffix(baseUrl, "/"),
		httpClient: &http.Client{
			Transport: NewAuthTransport(login, password),
		},
	}
}

type Client struct {
	l          *slog.Logger
	baseUrl    string
	httpClient *http.Client
}

func (c *Client) GetKeys(ctx context.Context, chatId int64) ([]*VpnKey, error) {
	urlString := fmt.Sprintf("%s/xui/API/inbounds/get/%d", c.baseUrl, inboundId)

	var inboundResp InboundResponse
	if err := c.do(ctx, "GET", urlString, nil, &inboundResp); err != nil {
		return nil, fmt.Errorf("failed to parse inboundResponse: %w", err)
	}

	var inboundStreamSettings InboundStreamSettings
	if err := json.Unmarshal([]byte(inboundResp.Obj.StreamSettings), &inboundStreamSettings); err != nil {
		return nil, fmt.Errorf("failed to parse streamSettings: %w", err)
	}

	var inboundSettings InboundSettings
	if err := json.Unmarshal([]byte(inboundResp.Obj.Settings), &inboundSettings); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	baseUrl, err := url.Parse(c.baseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse baseUrl: %w", err)
	}

	var keys []*VpnKey
	for _, client := range inboundSettings.Clients {
		parts := strings.SplitN(client.Email, "|", 3)
		if len(parts) == 3 && strconv.FormatInt(chatId, 10) == parts[0] {
			key := fmt.Sprintf(
				"%s://%s@%s:%d?type=%s&security=%s&pbk=%s&fp=%s&sni=%s&sid=&spx=%s#%s",
				inboundResp.Obj.Protocol,
				client.ID,
				baseUrl.Hostname(),
				inboundResp.Obj.Port,
				inboundStreamSettings.Network,
				inboundStreamSettings.Security,
				inboundStreamSettings.RealitySettings.Settings.PublikKey,
				inboundStreamSettings.RealitySettings.Settings.Fingerprint,
				inboundStreamSettings.RealitySettings.ServerNames[0],
				inboundStreamSettings.RealitySettings.Settings.SpiderX,
				// https://go.dev/play/p/pOfrn-Wsq5
				(&url.URL{Path: parts[2]}).String(),
			)

			keys = append(keys, &VpnKey{
				ID:     client.ID,
				ChatID: chatId,
				Title:  parts[2],
				Key:    key,
			})
		}
	}

	return keys, nil
}

func (c *Client) CreateKey(ctx context.Context, keyName string, chatId int64, user *telegram.User) (*VpnKey, error) {
	settings := InboundSettings{
		Clients: []InboundClient{{
			ID:     uuid.NewString(),
			Flow:   "",
			Email:  fmt.Sprintf("%d|%s|%s", chatId, user.DisplayName(), keyName),
			Enable: true,
			TgId:   user.Username,
		}},
	}

	settingsData, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}

	req := CreateClientRequest{
		ID:       inboundId,
		Settings: string(settingsData),
	}
	var res CreateClientResponse
	urlString := c.baseUrl + "/xui/API/inbounds/addClient"
	if err := c.do(ctx, "POST", urlString, req, &res); err != nil {
		return nil, fmt.Errorf("failed to add client: %w", err)
	}

	if !res.Success {
		return nil, fmt.Errorf("create client error: %s", res.Msg)
	}

	keys, err := c.GetKeys(ctx, chatId)
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %s", res.Msg)
	}

	if len(keys) < 1 {
		return nil, fmt.Errorf("expected at least one key, got %d", len(keys))
	}

	c.l.Info("key created", "user", user.DisplayName(), "name", keyName, "link", keys[len(keys)-1].Key)
	return keys[len(keys)-1], nil
}

func (c *Client) DeleteKey(ctx context.Context, key *VpnKey) error {
	urlString := fmt.Sprintf("%s/xui/API/inbounds/%d/delClient/%s", c.baseUrl, inboundId, key.ID)
	var out struct{}
	return c.do(ctx, "POST", urlString, nil, &out)
}

func (c *Client) do(ctx context.Context, method, url string, in, out any) error {
	var body io.Reader
	if in != nil {
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(&in); err != nil {
			return fmt.Errorf("failed to pack data %w", err)
		}
		body = io.NopCloser(buf)
	}

	request, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if method == "POST" {
		request.Header.Add("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	c.l.Debug("unmarshalling", "data", data)
	return json.Unmarshal(data, out)
}
