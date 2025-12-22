package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateHttpClient() *HttpClient {
	return &HttpClient{map[string]string{"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36"}}
}

type HttpClient struct {
	headers map[string]string
}

func (c *HttpClient) GetRedirectLocation(url core.URL) (core.URL, error) {
	client := http.DefaultClient
	req, _ := http.NewRequest("HEAD", url, nil)

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	return res.Request.URL.String(), nil
}

// GetContentType is a core.IHttpClient interface implementation
func (c *HttpClient) GetContentType(url core.URL) (string, error) {
	client := http.DefaultClient
	req, _ := http.NewRequest("HEAD", url, nil)

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if header, ok := res.Header["Content-Type"]; ok {
		return header[0], nil
	}
	return "", fmt.Errorf("content-type not found")
}

// GetContent is a core.IHttpClient interface implementation
func (c *HttpClient) GetContent(url core.URL) (string, error) {
	client := http.DefaultClient
	req, _ := http.NewRequest("GET", url, nil)

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// SetHeader remembers all passed values and applies it to every request
func (c *HttpClient) SetHeader(key string, value string) {
	c.headers[key] = value
}
