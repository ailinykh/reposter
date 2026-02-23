package telegram

import "strings"

func NormalizeCaption(caption string) string {
	if len(caption) > 1024 {
		caption = caption[:1024]
	}
	return strings.ToValidUTF8(caption, "")
}