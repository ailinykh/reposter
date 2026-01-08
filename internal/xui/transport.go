package xui

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/ailinykh/reposter/v3/pkg/helpers"
)

func NewAuthTransport(login, password string) http.RoundTripper {
	return &AuthTransport{
		T:        http.DefaultTransport,
		login:    login,
		password: password,
	}
}

type AuthTransport struct {
	T        http.RoundTripper
	cookie   *http.Cookie
	login    string
	password string
}

func (at *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if at.cookie == nil {
		err := at.Authorize(req)
		if err != nil {
			return nil, fmt.Errorf("failed to authorize: %w", err)
		}
	}
	req.AddCookie(at.cookie)
	resp, err := at.T.RoundTrip(req)

	if err != nil {
		return resp, err
	}

	if len(resp.Header["Content-Type"]) > 0 {
		if resp.Header["Content-Type"][0] != "application/json; charset=utf-8" {
			at.cookie = nil
			return at.RoundTrip(req)
		}
	}

	return resp, err
}

func (at *AuthTransport) Authorize(req *http.Request) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	params := map[string]io.Reader{
		"username": strings.NewReader(at.login),
		"password": strings.NewReader(at.password),
	}

	err := helpers.MultipartFrom(params, writer)
	if err != nil {
		return fmt.Errorf("failed to create muiltipart/form data: %w", err)
	}
	writer.Close()

	apiIndex := strings.Index(req.URL.String(), "/xui/API")
	if apiIndex == -1 {
		return fmt.Errorf("failed to find /xui/API in url: %s", req.URL.String())
	}
	loginUrl := fmt.Sprintf("%s/login", req.URL.String()[:apiIndex])

	r, _ := http.NewRequest("POST", loginUrl, body)
	r.Header.Add("Content-Type", writer.FormDataContentType())
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	for _, cookie := range res.Cookies() {
		if cookie.Name == "x-ui" {
			at.cookie = cookie
			return nil
		}
	}
	return fmt.Errorf("x-ui cookie not found")
}
