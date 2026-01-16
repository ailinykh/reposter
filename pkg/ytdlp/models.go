package ytdlp

import "fmt"

type Format struct {
	Ext        string `json:"ext"`
	Filesize   int64  `json:"filesize"`
	Format     string `json:"format"`
	FormatID   string `json:"format_id"`
	FormatNote string `json:"format_note"`
	Height     int64  `json:"height"`
	Width      int64  `json:"width"`
	ACodec     string `json:"acodec"`
	VCodec     string `json:"vcodec"`
}

type Response struct {
	*Format
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Duration    int64     `json:"duration"`
	Extractor   string    `json:"extractor"`
	Filesize    int64     `json:"filesize_approx"`
	Formats     []*Format `json:"formats"`
MediaType   string    `json:"media_type"`
	OriginalUrl string    `json:"original_url"`
	WebpageUrl  string    `json:"webpage_url"`
}

func (r *Response) FormatByID(id string) (*Format, error) {
	for _, f := range r.Formats {
		if f.FormatID == id {
			return f, nil
		}
	}
	return nil, fmt.Errorf("format with id %s not found", id)
}

func (r *Response) SuitableFormats(size int64) (vf *Format, af *Format, err error) {
	if af, err = r.FormatByID("140"); err != nil {
		return nil, nil, err
	}

	if af.Filesize > size {
		return nil, nil, fmt.Errorf("audio already exceeds desired size")
	}

	for i := range r.Formats {
		f := r.Formats[len(r.Formats)-i-1]
		if f.Filesize+af.Filesize < size && f.Ext == "mp4" {
			return f, af, nil
		}
	}

	return nil, nil, fmt.Errorf("no suitable formats found for size %d", size)
}
