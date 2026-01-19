package ffmpeg

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
)

func GetInfo(filePath string) (*Info, error) {
	cmd := fmt.Sprintf(`ffprobe -v panic -of json -show_streams -show_format "%s"`, filePath)
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to read streams: %s", string(out))
	}

	var info Info
	err = json.Unmarshal(out, &info)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe json: %w", err)
	}

	return &info, nil
}

func Crop(srcPath string, width, height int) (string, error) {
	dirPath, err := os.MkdirTemp("", "ffmpeg*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	destPath := path.Join(dirPath, "cropped.jpg")
	cmd := fmt.Sprintf(`ffmpeg -v error -i %s -vf "crop=%d:%d:(in_w-out_w)/2:(in_h-out_h)/2" -y %s`, srcPath, width, height, destPath)
	_, err = exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to crop image: %w", err)
	}

	return destPath, nil
}
