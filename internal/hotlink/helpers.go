package hotlink

import "github.com/ailinykh/reposter/v3/pkg/telegram"

func MediaFromVideos(videos []*telegram.Video, caption string) []telegram.InputMedia {
	medias := []telegram.InputMedia{}
	for i, v := range videos {
		video := telegram.InputMediaVideo{
			Type:                  "video",
			Media:                 v.FileID,
			ParseMode:             telegram.ParseModeHTML,
			ShowCaptionAboveMedia: true,
		}
		if (len(videos) - 1) == i {
			video.Caption = caption
		}
		medias = append(medias, video)
	}
	return medias
}

func VideoFromMessages(messages []*telegram.Message) []*telegram.Video {
	var videos = []*telegram.Video{}
	for _, m := range messages {
		if m.Video != nil {
			videos = append(videos, m.Video)
		}
	}
	return videos
}
