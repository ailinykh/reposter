package test_helpers

import (
	"os"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateImageDownloader() *FakeImageDownloader {
	return &FakeImageDownloader{[]string{}, nil}
}

type FakeImageDownloader struct {
	Downloaded []string
	Err        error
}

// Upload is a core.IImageDownloader interface implementation
func (fid *FakeImageDownloader) Download(image *core.Image) (*os.File, error) {
	fid.Downloaded = append(fid.Downloaded, image.FileURL)
	file, _ := os.Open(image.File.Path)
	return file, fid.Err
}
