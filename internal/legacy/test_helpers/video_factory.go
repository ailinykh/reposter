package test_helpers

import (
	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateVideoFactory() *FakeVideoFactory {
	return &FakeVideoFactory{[]string{}, nil}
}

type FakeVideoFactory struct {
	CreatedVideos []string
	Err           error
}

// CreateVideo is a core.IVideoFactory interface implementation
func (fvf *FakeVideoFactory) CreateVideo(path string) (*core.Video, error) {
	fvf.CreatedVideos = append(fvf.CreatedVideos, path)
	return &core.Video{File: core.File{Path: path}}, fvf.Err
}
