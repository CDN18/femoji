package upload

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/CDN18/femoji-cli/internal/auth"
	"github.com/CDN18/femoji-cli/internal/util"
	"github.com/go-openapi/runtime"
	"github.com/owu-one/gotosocial-sdk/client/admin"
	"github.com/owu-one/gotosocial-sdk/models"
)

func Upload(authClient *auth.Client, path, category string, override bool) error {
	slog.Info("Started uploading emojis", "path", path, "category", category, "override", override)
	// get emojis data from current instance
	emojis, err := authClient.Client.Admin.EmojisGet(
		&admin.EmojisGetParams{
			Filter: util.Ptr("domain:local"),
			Limit:  util.Ptr(int64(0)),
		},
		// nil,
		admin.ClientOption(
			func(op *runtime.ClientOperation) {
				op.AuthInfo = authClient.Auth
				// op.PathPattern = "/api/v1/admin/custom_emojis?filter=domain:local&limit=0"
			},
		),
	)
	if err != nil {
		slog.Error("Error getting emojis", "error", err)
		return err
	}
	// filter emojis by category
	currEmojis, err := util.FilterAdminEmojisByCategory(emojis, category)
	if err != nil {
		slog.Error("Error filtering emojis", "error", err)
		return err
	}

	files, err := os.ReadDir(path)
	if err != nil {
		slog.Error("Error reading directory", "error", err)
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			slog.Info("Skipping", file.Name(), "as it is a directory")
			continue
		}
		// check if file is image
		if !util.IsImage(file.Name()) {
			slog.Info("Skipping", file.Name(), "as it is not an image")
			continue
		}
		// check if filename equals to any emoji shortcode
		var exist bool
		var existingEmoji *models.AdminEmoji
		for _, emoji := range currEmojis {
			if emoji.Shortcode == strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())) {
				exist = true
				existingEmoji = emoji
				slog.Info("Emoji already exists, will override if --override flag is set", "shortcode", file.Name())
				break
			}
		}
		if exist && override {
			slog.Info("Overriding existing emoji", "shortcode", file.Name())
			// override emoji
			_, err := authClient.Client.Admin.EmojiUpdate(
				&admin.EmojiUpdateParams{
					Type:  "modify",
					ID:    existingEmoji.ID,
					Image: runtime.NamedReader(file.Name(), util.OpenFile(path+"/"+file.Name())),
				},
				authClient.Auth,
				func(op *runtime.ClientOperation) {
					op.ConsumesMediaTypes = []string{"multipart/form-data"}
				},
			)
			if err != nil {
				slog.Error("Error overriding", "file", file.Name(), "error", err)
				// continue to next file
				continue
			} else {
				slog.Info("Skipping", file.Name(), "as it already exists, to override set --override flag")
				continue
			}
		}
		// upload emoji
		slog.Info("Uploading emoji", "shortcode", strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())))
		_, err := authClient.Client.Admin.EmojiCreate(
			&admin.EmojiCreateParams{
				Category:  util.Ptr(category),
				Image:     runtime.NamedReader(file.Name(), util.OpenFile(path+"/"+file.Name())),
				Shortcode: strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())),
			},
			authClient.Auth,
			func(op *runtime.ClientOperation) {
				op.ConsumesMediaTypes = []string{"multipart/form-data"}
			},
		)
		if err != nil {
			slog.Error("Error uploading", "file", file.Name(), "error", err)
			// continue to next file
			continue
		}
	}
	return nil
}
