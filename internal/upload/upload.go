package upload

import (
	"log/slog"
	"os"

	"github.com/CDN18/femoji-cli/internal/auth"
	"github.com/CDN18/femoji-cli/internal/util"
	"github.com/go-openapi/runtime"
	"github.com/owu-one/gotosocial-sdk/client/admin"
)

func Upload(authClient *auth.Client, path, category string) error {
	// get emojis data from current instance
	emojis, err := authClient.Client.Admin.EmojisGet(
		&admin.EmojisGetParams{
			Filter: util.Ptr("domain:local"),
			Limit:  util.Ptr(int64(0)),
		},
	)
	if err != nil {
		return err
	}
	// filter emojis by category
	currEmojis, err := util.FilterEmojisByCategory(emojis, category)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		// check if file is image
		if !util.IsImage(file.Name()) {
			slog.Info("Skipping", file.Name(), "as it is not an image")
			continue
		}
		// check if filename equals to any emoji shortcode
		var exist bool
		for _, emoji := range currEmojis {
			if emoji.Shortcode == file.Name() {
				exist = true
				break
			}
		}
		if exist {
			slog.Info("Skipping", file.Name(), "as it already exists")
			continue
		}
		// upload emoji
		_, err := authClient.Client.Admin.EmojiCreate(
			&admin.EmojiCreateParams{
				Category:  util.Ptr(category),
				Image:     runtime.NamedReader(file.Name(), util.OpenFile(path+"/"+file.Name())),
				Shortcode: file.Name(),
			},
			authClient.Auth,
			func(op *runtime.ClientOperation) {
				op.ConsumesMediaTypes = []string{"application/x-www-form-urlencoded"}
			},
		)
		if err != nil {
			slog.Error("Error uploading", "file", file.Name(), "error", err)
			return err
		}
	}
	return nil
}