package hotlink

import (
	"fmt"
	"os"

	"github.com/ailinykh/reposter/v3/pkg/ffmpeg"
	"github.com/ailinykh/reposter/v3/pkg/telegram"
	"github.com/ailinykh/reposter/v3/pkg/ytdlp"
)

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

func CroppedThumb(r *ytdlp.Response, v *ytdlp.LocalVideo) (*ytdlp.LocalFile, error) {
	info, err := ffmpeg.GetInfo(v.Thumb.Path)
	if err != nil {
		return nil, err
	}

	if len(info.Streams) < 1 {
		return nil, fmt.Errorf("no stream found at %s", v.Thumb.Path)
	}

	w := r.Width * info.Streams[0].Height / r.Height
	cropped, err := ffmpeg.Crop(v.Thumb.Path, w, info.Streams[0].Height)
	if err != nil {
		return nil, fmt.Errorf("failed to crop %s: %w", v.Thumb.Path, err)
	}

	f, err := os.Open(cropped)
	if err != nil {
		return nil, fmt.Errorf("failed to open cropped file: %w", err)
	}

	return &ytdlp.LocalFile{
		File: f,
		Path: cropped,
	}, nil
}
