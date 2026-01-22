package xcom

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func NewAuthTransport() http.RoundTripper {
	return &AuthTransport{
		T:           http.DefaultTransport,
		BearerToken: "AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA",
	}
}

type AuthTransport struct {
	T           http.RoundTripper
	BearerToken string
	GuestToken  string
}

func (at *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if at.GuestToken == "" {
		if err := at.Authorize(req); err != nil {
			return nil, fmt.Errorf("failed to authorize: %w", err)
		}
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", at.BearerToken))
	req.Header.Set("X-Guest-Token", at.GuestToken)
	resp, err := at.T.RoundTrip(req)

	if err != nil {
		return resp, err
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json; charset=utf-8" && contentType != "application/json" {
		at.GuestToken = ""
		return at.RoundTrip(req)
	}

	return resp, err
}

func (at *AuthTransport) Authorize(req *http.Request) error {
	req, _ = http.NewRequestWithContext(req.Context(), "POST", "https://api.x.com/1.1/guest/activate.json", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", at.BearerToken))
	res, err := at.T.RoundTrip(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var tokenResponse struct {
		Token   string `json:"guest_token,omitempty"`
		Code    int    `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
	}

	if err := json.NewDecoder(res.Body).Decode(&tokenResponse); err != nil {
		return err
	}
	at.GuestToken = tokenResponse.Token
	return nil
}
