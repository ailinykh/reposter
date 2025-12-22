package infrastructure

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

// CreateFileDownloader is a default FileDownloader factory
func CreateFileDownloader(l core.Logger) *FileDownloader {
	return &FileDownloader{l}
}

// FileDownloader is a default implementation for core.IFileDownloader
type FileDownloader struct {
	l core.Logger
}

// Download is a core.IFileDownloader interface implementation
func (downloader *FileDownloader) Download(url legacy.URL, filepath string) (*legacy.File, error) {
	name := path.Base(filepath)
	downloader.l.Info("downloading", "url", url, "file_path", strings.ReplaceAll(filepath, os.TempDir(), "$TMPDIR/"))
	// Get the data
	client := http.DefaultClient
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:59.0) Gecko/20100101 Firefox/59.0")
	req.Header.Set("Referer", url)
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get url: %v", err)
	}
	defer res.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file at %s: %v", filepath, err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %v", err)
	}

	// Retreive file size
	stat, err := os.Stat(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to get stat for %s: %v", filepath, err)
	}

	return &legacy.File{Name: name, Path: filepath, Size: stat.Size()}, err
}
