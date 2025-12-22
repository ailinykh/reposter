package test_helpers

import (
	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateFileDownloader() *FakeFileDownloader {
	return &FakeFileDownloader{make(map[string]string), nil}
}

type FakeFileDownloader struct {
	DownloadedFiles map[string]string
	Err             error
}

// Download is a core.IFileDownloader interface implementation
func (ffd *FakeFileDownloader) Download(url core.URL, filepath string) (*core.File, error) {
	ffd.DownloadedFiles[url] = filepath
	return &core.File{Path: filepath}, ffd.Err
}
