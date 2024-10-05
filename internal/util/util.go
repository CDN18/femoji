package util

import (
	"io"
	"os"
	"strings"

	"github.com/owu-one/gotosocial-sdk/client/admin"
	"github.com/owu-one/gotosocial-sdk/models"
)

func IsImage(name string) bool {
	suffixes := []string{".png", ".gif"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

func FilterEmojisByCategory(emojis *admin.EmojisGetOK, category string) ([]*models.AdminEmoji, error) {
	var filtered []*models.AdminEmoji
	for _, emoji := range emojis.Payload {
		if emoji.Category == category {
			filtered = append(filtered, emoji)
		}
	}
	return filtered, nil
}

func OpenFile(path string) io.Reader {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	return file
}
