package infrastructure

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

// CreateFfmpegConverter is a basic FfmpegConverter factory
func CreateFfmpegConverter(l core.Logger) *FfmpegConverter {
	return &FfmpegConverter{l}
}

// FfmpegConverter implements core.IVideoConverter and core.IVideoFactory using ffmpeg
type FfmpegConverter struct {
	l core.Logger
}

// Convert is a core.IVideoConverter interface implementation
func (c *FfmpegConverter) Convert(vf *legacy.Video, bitrate int) (*legacy.Video, error) {
	path := path.Join(os.TempDir(), vf.Name+"_converted.mp4")
	cmd := fmt.Sprintf(`ffmpeg -v error -y -i "%s" -pix_fmt yuv420p -vf "scale=trunc(iw/2)*2:trunc(ih/2)*2" "%s"`, vf.Path, path)
	if bitrate > 0 {
		cmd = fmt.Sprintf(`ffmpeg -v error -y -i "%s" -c:v libx264 -preset medium -b:v %dk -pass 1 -b:a 128k -f mp4 /dev/null && ffmpeg -v error -y -i "%s" -c:v libx264 -preset medium -b:v %dk -pass 2 -b:a 128k "%s"`, vf.Path, bitrate/1024, vf.Path, bitrate/1024, path)
	}
	c.l.Info(strings.ReplaceAll(cmd, os.TempDir(), "$TMPDIR/"))
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		os.Remove(path)
		c.l.Error(err)
		return nil, fmt.Errorf(string(out))
	}

	return c.CreateVideo(path)
}

// GetCodec is a core.IVideoConverter interface implementation
func (c *FfmpegConverter) GetCodec(path string) string {
	ffprobe, err := c.getFFProbe(path)
	if err != nil {
		c.l.Error(err)
		return "unknown"
	}

	stream, err := ffprobe.getVideoStream()
	if err != nil {
		c.l.Error(err)
		return "unknown"
	}

	return stream.CodecName
}

// CreateMedia is a core.IMediaFactory interface implementation
func (c *FfmpegConverter) CreateMedia(url string) ([]*legacy.Media, error) {
	ffprobe, err := c.getFFProbe(url)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	stream, err := ffprobe.getVideoStream()
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	size, err := strconv.Atoi(ffprobe.Format.Size)
	if err != nil {
		c.l.Error("failed to convert size", "error", err)
		size = 0
	}

	if ffprobe.Format.FormatName == "image2" {
		return []*legacy.Media{{ResourceURL: url, URL: url, Codec: stream.CodecName, Size: size, Type: legacy.TPhoto}}, nil
	}

	return []*legacy.Media{{ResourceURL: url, URL: url, Codec: stream.CodecName, Size: size, Type: legacy.TVideo}}, nil
}

// CreateVideo is a core.IVideoSplitter interface implementation
func (c *FfmpegConverter) Split(video *legacy.Video, limit int) ([]*legacy.Video, error) {
	duration, n := 0, 0
	var videos = []*legacy.Video{}
	for duration < video.Duration {
		path := fmt.Sprintf("%s[%02d].mp4", video.File.Path, n)
		cmd := fmt.Sprintf(`ffmpeg -v error -y -i %s -ss %d -fs %dM -map_metadata 0 -c copy %s`, video.File.Path, duration, limit, path)
		c.l.Info(strings.ReplaceAll(cmd, os.TempDir(), "$TMPDIR/"))
		out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
		if err != nil {
			c.l.Error(err)
			os.Remove(path)
			return nil, fmt.Errorf(string(out))
		}

		file, err := c.CreateVideo(path)
		if err != nil {
			c.l.Error(err)
			os.Remove(path)
			if err.Error() == "file is too short" {
				// the last piece might be shorter than a second - https://youtu.be/1MLRCczBKn8
				// of have a black screen - https://youtu.be/TQ2szA18aEc
				duration += 10
				continue
			} else {
				return nil, err
			}
		}
		// defer file.Dispose()

		videos = append(videos, file)
		duration += file.Duration
		n++
	}
	return videos, nil
}

// CreateVideo is a core.IVideoFactory interface implementation
func (c *FfmpegConverter) CreateVideo(path string) (*legacy.Video, error) {
	c.l.Info("create video", "file_path", strings.ReplaceAll(path, os.TempDir(), "$TMPDIR/"))
	ffprobe, err := c.getFFProbe(path)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	duration, err := strconv.ParseFloat(ffprobe.Format.Duration, 32)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	var min = 0.3
	if duration < min {
		return nil, fmt.Errorf("file is too short: expected at least %f seconds duration, got %f", min, duration)
	}

	stream, err := ffprobe.getVideoStream()
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	scale := "320:-1"
	if stream.Width < stream.Height {
		scale = "-1:320"
	}

	thumb, err := c.createThumb(path, scale)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	bitrate, _ := strconv.Atoi(stream.BitRate) // empty for .gif

	stat, err := os.Stat(path)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	return &legacy.Video{
		File:     legacy.File{Name: stat.Name(), Path: path, Size: stat.Size()},
		Width:    stream.Width,
		Height:   stream.Height,
		Bitrate:  bitrate,
		Duration: int(duration),
		Codec:    stream.CodecName,
		Thumb:    thumb}, nil
}

func (c *FfmpegConverter) getFFProbe(file string) (*ffpResponse, error) {
	cmd := fmt.Sprintf(`ffprobe -v panic -of json -show_streams -show_format "%s"`, file)
	c.l.Info(strings.ReplaceAll(cmd, os.TempDir(), "$TMPDIR/"))
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		c.l.Error(err)
		return nil, fmt.Errorf(string(out))
	}

	var resp ffpResponse
	err = json.Unmarshal(out, &resp)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	return &resp, nil
}

func (c *FfmpegConverter) createThumb(videoPath string, scale string) (*legacy.Image, error) {
	thumbPath := videoPath + ".jpg"

	cmd := fmt.Sprintf(`ffmpeg -v error -y -i "%s" -ss 00:00:01.000 -vframes 1 -filter:v scale="%s" "%s"`, videoPath, scale, thumbPath)
	c.l.Info(strings.ReplaceAll(cmd, os.TempDir(), "$TMPDIR/"))
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		c.l.Error(err)
		return nil, fmt.Errorf(string(out))
	}

	ffprobe, err := c.getFFProbe(thumbPath)
	if err != nil {
		c.l.Error(err)
		return nil, err
	}

	return &legacy.Image{
		File:   legacy.File{Name: path.Base(thumbPath), Path: thumbPath},
		Width:  ffprobe.Streams[0].Width,
		Height: ffprobe.Streams[0].Height,
	}, nil
}
