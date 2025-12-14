package api

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/internal/core"
	legacy "github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
)

func CreateYoutubeMediaFactory(l core.Logger, api YoutubeApi, fd legacy.IFileDownloader) *YoutubeMediaFactory {
	return &YoutubeMediaFactory{l, api, fd}
}

type YoutubeMediaFactory struct {
	l   core.Logger
	api YoutubeApi
	fd  legacy.IFileDownloader
}

// CreateMedia is a core.IMediaFactory interface implementation
func (y *YoutubeMediaFactory) CreateMedia(url string) ([]*legacy.Media, error) {
	resp, err := y.api.Get(url)
	if err != nil {
		y.l.Error(err)
		return nil, err
	}

	video, audio, err := y.getFormats(resp)
	if err != nil {
		y.l.Error(err)
		return nil, err
	}

	return []*legacy.Media{
		{
			URL:         video.Url,
			Title:       resp.Title,
			Description: resp.Description,
			Duration:    int(resp.Duration),
			Codec:       video.Vcodec,
			Size:        int(video.Filesize),
			Type:        legacy.TVideo,
		},
		{
			URL:         audio.Url,
			Title:       resp.Title,
			Description: resp.Description,
			Duration:    int(resp.Duration),
			Codec:       audio.Acodec,
			Size:        int(audio.Filesize),
			Type:        legacy.TAudio,
		},
	}, nil
}

func (y *YoutubeMediaFactory) getFormats(resp *YtDlpResponse) (*YtDlpFormat, *YtDlpFormat, error) {
	audio, err := y.getPreferredAudioFormat(resp)
	if err != nil {
		y.l.Error(err)
		return nil, nil, err
	}

	video, err := y.getPreferredVideoFormat(resp)
	if err != nil {
		y.l.Error(err)
		return nil, nil, err
	}
	return video, audio, nil
}

// CreateVideo is a core.IVideoFactory interface implementation
func (y *YoutubeMediaFactory) CreateVideo(id string) (*legacy.Video, error) {
	resp, err := y.api.Get(id)
	if err != nil {
		y.l.Error(err)
		return nil, err
	}

	video, audio, err := y.getFormats(resp)
	if err != nil {
		y.l.Error(err)
		return nil, err
	}

	name := fmt.Sprintf("youtube[%s][%s][%s].mp4", resp.Id, video.FormatNote, audio.FormatNote)
	videoPath := path.Join(os.TempDir(), name)

	cmd := fmt.Sprintf("yt-dlp -f %s+%s https://youtu.be/%s -o %s", video.FormatId, audio.FormatId, resp.Id, videoPath)
	y.l.Info(strings.ReplaceAll(cmd, os.TempDir(), "$TMPDIR/"))
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		y.l.Error(err)
		return nil, fmt.Errorf("%s", out)
	}

	thumb, err := y.makeThumb(resp, video)
	if err != nil {
		return nil, err
	}

	return &legacy.Video{
		File:     legacy.File{Name: name, Path: videoPath, Size: video.Filesize + audio.Filesize},
		Width:    video.Width,
		Height:   video.Height,
		Bitrate:  0,
		Duration: int(resp.Duration),
		Codec:    video.Vcodec,
		Thumb:    thumb,
	}, nil
}

func (y *YoutubeMediaFactory) makeThumb(resp *YtDlpResponse, vf *YtDlpFormat) (*legacy.Image, error) {
	filename := fmt.Sprintf("youtube[%s][%s].jpg", resp.Id, vf.FormatId)
	originalThumbPath := path.Join(os.TempDir(), filename+"-original")
	thumbPath := path.Join(os.TempDir(), filename)
	file, err := y.fd.Download(resp.Thumbnail, originalThumbPath)
	if err != nil {
		y.l.Error(err)
		return nil, err
	}
	defer file.Dispose()

	scale := "320:-1"
	if vf.Width < vf.Height {
		scale = "-1:320"
	}

	cmd := fmt.Sprintf(`ffmpeg -v error -y -i "%s" -vf scale=%s "%s"`, originalThumbPath, scale, thumbPath)
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		y.l.Error(err)
		y.l.Error(string(out))
		return nil, err
	}

	stat, err := os.Stat(thumbPath)
	if err != nil {
		y.l.Error(err)
		return nil, err
	}

	return &legacy.Image{
		File:   legacy.File{Name: filename, Path: thumbPath, Size: stat.Size()},
		Width:  vf.Width,
		Height: vf.Height,
	}, nil
}

func (y *YoutubeMediaFactory) getPreferredAudioFormat(resp *YtDlpResponse) (*YtDlpFormat, error) {
	for _, f := range resp.Formats {
		if f.FormatId == "140" {
			return f, nil
		}
	}

	return nil, fmt.Errorf("140 not found for %s", resp.Id)
}

func (y *YoutubeMediaFactory) getPreferredVideoFormat(resp *YtDlpResponse) (*YtDlpFormat, error) {
	n := -1
	for i, f := range resp.Formats {
		if f.Filesize > 0 && f.Filesize < 50_000_000 && strings.HasPrefix(f.Vcodec, "avc1") && (n == -1 || resp.Formats[n].Filesize < f.Filesize) {
			n = i
		}
	}

	if n < 0 {
		// the smallest `mp4` video format
		for _, f := range resp.Formats {
			if f.FormatId == "134" {
				return f, nil
			}
		}
		return nil, fmt.Errorf("appropriate video format not found")
	}

	return resp.Formats[n], nil
}
