package api

import (
	"fmt"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateTikTokMediaFactory(api YoutubeApi) core.IMediaFactory {
	return &TikTokMediaFactory{api}
}

type TikTokMediaFactory struct {
	api YoutubeApi
}

func (factory *TikTokMediaFactory) CreateMedia(url string) ([]*core.Media, error) {
	item, err := factory.api.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get content for %s: %v", url, err)
	}

	media := &core.Media{
		URL:         url,
		ResourceURL: item.Url,
		Title:       fmt.Sprintf("%s (@%s)", item.Creator, item.Uploader),
		Description: fmt.Sprintf(`<a href="https://tiktok.com/@%s">(@%s)'s</a> short video with %s - %s.`, item.Uploader, item.Creator, item.Track, item.Artist),
	}

	return []*core.Media{media}, nil
}
