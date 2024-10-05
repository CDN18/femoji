package download

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/CDN18/femoji-cli/internal/auth"
	"github.com/owu-one/gotosocial-sdk/models"
)

func Download(authClient *auth.Client, instance string, override bool) error {
	var emojis []*models.Emoji
	if instance == "DEFAULT" {
		emojiResp, err := authClient.Client.CustomEmojis.CustomEmojisGet(nil, authClient.Auth)
		if err != nil {
			return err
		}
		emojis = emojiResp.GetPayload()
	} else {
		endpoint := fmt.Sprintf("https://%s/api/v1/custom_emojis", instance)
		resp, err := http.Get(endpoint)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			slog.Error("failed to get custom emojis", "status", resp.StatusCode, "instance", instance)
			return nil
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(body, &emojis); err != nil {
			return err
		}
	}

	slog.Info("Emoji List Retrieved", "count", len(emojis))
	for _, emoji := range emojis {
		if emoji.Category == "" {
			emoji.Category = "uncategorized"
		}
		slog.Info("downloading emoji", "shortcode", emoji.Shortcode, "category", emoji.Category, "url", emoji.URL)
		dir := fmt.Sprintf("%s/%s", instance, emoji.Category)
		// create dir if not exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				slog.Error("failed to create directory", "error", err, "shortcode", emoji.Shortcode, "path", dir)
				continue
			}
		}
		// download emoji
		extension := filepath.Ext(emoji.URL)
		filePath := fmt.Sprintf("%s/%s%s", dir, emoji.Shortcode, extension)
		if _, err := os.Stat(filePath); err == nil && !override {
			slog.Info("skipping download as it already exists", "shortcode", emoji.Shortcode, "path", filePath)
			continue
		}
		resp, err := http.Get(emoji.URL)
		if err != nil {
			slog.Error("failed to download emoji", "error", err, "shortcode", emoji.Shortcode, "url", emoji.URL)
			continue
		}
		// check x-ratelimit-remaining and x-ratelimit-reset headers
		if resp.Header.Get("x-ratelimit-remaining") == "0" {
			slog.Error("rate limit exceeded")
			resetTime, err := time.Parse(time.RFC1123, resp.Header.Get("x-ratelimit-reset"))
			if err != nil {
				slog.Error("failed to parse x-ratelimit-reset header, using default of 300 seconds", "error", err)
				// default to 300 seconds
				resetTime = time.Now().Add(300 * time.Second)
			}
			slog.Info("rate limit will reset at", "time", resetTime)
			time.Sleep(time.Until(resetTime))
		}
		// follow 3xx redirects
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			redirectURL := resp.Header.Get("Location")
			resp, err = http.Get(redirectURL)
			if err != nil {
				slog.Error("failed to follow redirect", "error", err, "url", redirectURL)
				continue
			}
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			slog.Error("failed to download emoji", "status", resp.StatusCode, "shortcode", emoji.Shortcode, "url", emoji.URL)
			continue
		}
		file, err := os.Create(filePath)
		if err != nil {
			slog.Error("failed to create file", "error", err, "shortcode", emoji.Shortcode, "path", filePath)
			continue
		}
		defer file.Close()
		if _, err := io.Copy(file, resp.Body); err != nil {
			slog.Error("failed to write to file", "error", err, "shortcode", emoji.Shortcode, "path", filePath)
			continue
		}
		// check rate-limit, again
		if resp.Header.Get("x-ratelimit-remaining") == "0" {
			slog.Error("rate limit exceeded")
			resetTime, err := time.Parse(time.RFC1123, resp.Header.Get("x-ratelimit-reset"))
			if err != nil {
				slog.Error("failed to parse x-ratelimit-reset header, using default of 300 seconds", "error", err)
				// default to 300 seconds
				resetTime = time.Now().Add(300 * time.Second)
			}
			slog.Info("rate limit will reset at", "time", resetTime)
			time.Sleep(time.Until(resetTime))
		}
	}
	return nil
}
