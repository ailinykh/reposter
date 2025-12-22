package test_helpers

import "github.com/ailinykh/reposter/v2/internal/legacy/core"

func CreateMediaFactory() *FakeMediaFactory {
	return &FakeMediaFactory{[]core.URL{}}
}

type FakeMediaFactory struct {
	URLs []core.URL
}

func (factory *FakeMediaFactory) CreateMedia(url core.URL) ([]*core.Media, error) {
	factory.URLs = append(factory.URLs, url)
	return []*core.Media{{URL: url}}, nil
}
