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
