package helpers

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateDownloadVideoFactory(l core.Logger, fileDownloader legacy.IFileDownloader, videoFactory legacy.IVideoFactory) legacy.IVideoFactory {
	return &DownloadVideoFactory{l, fileDownloader, videoFactory}
}

type DownloadVideoFactory struct {
	l              core.Logger
	fileDownloader legacy.IFileDownloader
	videoFactory   legacy.IVideoFactory
}

// CreateVideo is a core.IVideoFactory interface implementation
func (factory *DownloadVideoFactory) CreateVideo(url string) (*legacy.Video, error) {
	filename := path.Base(url)
	if strings.Contains(filename, "?") {
		parts := strings.Split(url, "?")
		filename = path.Base(parts[0])
	}

	if !strings.HasSuffix(filename, ".mp4") {
		filename = filename + ".mp4"
	}

	videoPath := path.Join(os.TempDir(), filename)
	file, err := factory.fileDownloader.Download(url, videoPath)
	if err != nil {
		file.Dispose()
		return nil, fmt.Errorf("failed to download file for %s: %v", url, err)
	}
	factory.l.Info("file downloaded", "file_name", file.Name, "file_size", fmt.Sprintf("%0.2fMB", float64(file.Size)/1024/1024))
	return factory.videoFactory.CreateVideo(file.Path)
}
