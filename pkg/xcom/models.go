package xcom

type Tweet struct {
	Core struct {
		UserResults struct {
			Result struct {
				Core struct {
					Name       string `json:"name"`
					ScreenName string `json:"screen_name"`
				} `json:"core"`
			} `json:"result"`
		} `json:"user_results"`
	} `json:"core"`
	NoteTweet *NoteTweet `json:"note_tweet,omitempty"`
	Legacy    struct {
		Entities struct {
			Media []struct {
				MediaUrlHttps string     `json:"media_url_https"`
				Type          string     `json:"type"`
				VideoInfo     *VideoInfo `json:"video_info,omitempty"`
			} `json:"media"`
		} `json:"entities"`
		FullText string `json:"full_text"`
	} `json:"legacy"`
	TypeName string `json:"__typename"`
	Reason   string `json:"reason"`
	RestID   string `json:"rest_id"`
}

type NoteTweet struct {
	NoteTweetResults struct {
		Result struct {
			Text string `json:"text"`
		} `json:"result"`
	} `json:"note_tweet_results"`
}

type VideoInfo struct {
	Duration int64               `json:"duration_millis"`
	Variants []*VideoInfoVariant `json:"variants"`
}

func (info *VideoInfo) Best() *VideoInfoVariant {
	variant := info.Variants[0]
	for _, v := range info.Variants {
		if v.ContentType == "video/mp4" && v.Bitrate > variant.Bitrate {
			variant = v
		}
	}
	return variant
}

type VideoInfoVariant struct {
	Bitrate     int    `json:"bitrate"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
}
