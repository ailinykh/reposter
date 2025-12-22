package image_uploader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ailinykh/reposter/v2/internal/core"
	"github.com/ailinykh/reposter/v2/internal/helpers"
)

func NewPostimages(l core.Logger) Uploader {
	return &Postimages{
		l: l,
	}
}

type Postimages struct {
	l core.Logger
}

func (p *Postimages) Upload(file *os.File) (*url.URL, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// timestamp with 16-digits after period (1732443159.6623318195343018)
	session := fmt.Sprintf("%.016f", float64(time.Now().UnixNano())/10e8)

	params := map[string]io.Reader{
		"numfiles":       strings.NewReader("1"),
		"upload_session": strings.NewReader(session),
		"file":           file,
	}

	err := helpers.MultipartFrom(params, writer)
	if err != nil {
		return nil, fmt.Errorf("failed to create muiltipart/form data: %s", err)
	}
	writer.Close()

	p.l.Info("uploading", "file", file.Name())

	r, _ := http.NewRequest("POST", "https://postimages.org/json", body)
	r.Header.Add("Content-Type", writer.FormDataContentType())
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		p.l.Error(err)
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body from response %d: %s", res.StatusCode, err)
	}

	p.l.Debug(string(data))

	type imageResp struct {
		Url string `json:"url"`
	}

	var image imageResp
	err = json.Unmarshal(data, &image)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json '%s': %s", data, err)
	}

	return p.parseHotlink(image.Url)
}

func (p *Postimages) parseHotlink(link string) (*url.URL, error) {
	res, err := http.DefaultClient.Get(link)
	if err != nil {
		return nil, fmt.Errorf("failet to GET url %s: %s", link, err)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body from response %d: %s", res.StatusCode, err)
	}
	defer res.Body.Close()

	// <input id="code_direct" type="text" value="https://i.postimg.cc/wTFZ3q5L/uTwd.jpg" autocomplete="off" readonly="">
	t := regexp.MustCompile(`<input.*?id="code_direct".*?value="(.*?)".*?>`)
	matches := t.FindStringSubmatch(string(data))

	if len(matches) < 2 {
		return nil, fmt.Errorf("hotlink not found in HTML")
	}

	url, err := url.Parse(matches[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse url from %s: %s", matches[1], err)
	}

	return url, nil
}
