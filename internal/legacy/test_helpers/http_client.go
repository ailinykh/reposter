package test_helpers

import (
	"fmt"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateHttpClient() *FakeHttpClient {
	return &FakeHttpClient{make(map[string]string)}
}

type FakeHttpClient struct {
	ContentTypeForURL map[core.URL]string
}

func (client *FakeHttpClient) GetContentType(url core.URL) (string, error) {
	if contentType, ok := client.ContentTypeForURL[url]; ok {
		return contentType, nil
	}
	return "", fmt.Errorf("content type not found for %s", url)
}

func (client *FakeHttpClient) GetContent(core.URL) (string, error) {
	return "", nil
}

func (client *FakeHttpClient) GetRedirectLocation(url core.URL) (core.URL, error) {
	return "", nil
}

func (client *FakeHttpClient) SetHeader(string, string) {}
