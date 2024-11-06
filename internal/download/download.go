package download

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/CDN18/femoji-cli/internal/auth"
	"github.com/CDN18/femoji-cli/internal/util"
	"github.com/owu-one/gotosocial-sdk/models"
)

type MisskeyResponse struct {
	Emojis []MisskeyEmoji `json:"emojis"`
}

type MisskeyEmoji struct {
	Aliases   []string `json:"aliases"`
	Name      string   `json:"name"`
	Category  *string  `json:"category"`
	URL       string   `json:"url"`
	LocalOnly bool     `json:"localOnly"`
	Sensitive bool     `json:"isSensitive"`
	RoleIds   []string `json:"roleIdsThatCanBeUsedThisEmojiAsReaction"`
}

var mastodonLike = []string{"mastodon", "gotosocial", "pleroma", "akkoma", "hometown"}
var misskeyLike = []string{"misskey", "firefish", "iceshrimp", "sharkey", "catodon", "foundkey"}

func downloadWorker(id int, jobs <-chan *models.Emoji, wg *sync.WaitGroup, instance string, override bool) {
	defer wg.Done()

	for emoji := range jobs {
		if emoji.Category == "" {
			emoji.Category = "uncategorized"
		}

		dir := fmt.Sprintf("%s/%s", instance, emoji.Category)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				slog.Error("failed to create directory", "worker", id, "error", err, "shortcode", emoji.Shortcode, "path", dir)
				continue
			}
		}

		extension := filepath.Ext(emoji.URL)
		filePath := fmt.Sprintf("%s/%s%s", dir, emoji.Shortcode, extension)
		if _, err := os.Stat(filePath); err == nil && !override {
			slog.Info("skipping download as it already exists", "worker", id, "shortcode", emoji.Shortcode, "path", filePath)
			continue
		}

		resp, err := http.Get(emoji.URL)
		if err != nil {
			slog.Error("failed to download emoji", "worker", id, "error", err, "shortcode", emoji.Shortcode, "url", emoji.URL)
			continue
		}

		if resp.Header.Get("x-ratelimit-remaining") == "0" {
			slog.Error("rate limit exceeded", "worker", id)
			resetTime, err := time.Parse(time.RFC1123, resp.Header.Get("x-ratelimit-reset"))
			if err != nil {
				resetTime = time.Now().Add(300 * time.Second)
			}
			slog.Info("rate limit will reset at", "worker", id, "time", resetTime)
			time.Sleep(time.Until(resetTime))
		}

		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			redirectURL := resp.Header.Get("Location")
			resp, err = http.Get(redirectURL)
			if err != nil {
				slog.Error("failed to follow redirect", "worker", id, "error", err, "url", redirectURL)
				continue
			}
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			slog.Error("failed to download emoji", "worker", id, "status", resp.StatusCode, "shortcode", emoji.Shortcode, "url", emoji.URL)
			continue
		}

		file, err := os.Create(filePath)
		if err != nil {
			resp.Body.Close()
			slog.Error("failed to create file", "worker", id, "error", err, "shortcode", emoji.Shortcode, "path", filePath)
			continue
		}

		_, err = io.Copy(file, resp.Body)
		resp.Body.Close()
		file.Close()

		if err != nil {
			slog.Error("failed to write to file", "worker", id, "error", err, "shortcode", emoji.Shortcode, "path", filePath)
			continue
		}

		slog.Info("downloaded emoji", "worker", id, "shortcode", emoji.Shortcode)
	}
}

func Download(authClient *auth.Client, instance string, category string, override bool, instanceType string, threadCount int) error {
	if instance != "DEFAULT" {
		if instanceType == "mastodon" {
			nodeinfo, err := util.GetNodeInfo(instance)
			if err != nil {
				return err
			}

			isMastodon := false
			isMisskey := false

			for _, name := range mastodonLike {
				if nodeinfo.Software.Name == name {
					isMastodon = true
					break
				}
			}

			for _, name := range misskeyLike {
				if nodeinfo.Software.Name == name {
					isMisskey = true
					break
				}
			}

			if isMisskey {
				instanceType = "misskey"
			} else if !isMastodon {
				return fmt.Errorf("unknown instance type: %s", nodeinfo.Software.Name)
			}
		} else if instanceType != "misskey" {
			return fmt.Errorf("invalid instance type: %s", instanceType)
		}
	}

	var emojis []*models.Emoji
	if instance == "DEFAULT" {
		emojiResp, err := authClient.Client.CustomEmojis.CustomEmojisGet(nil, authClient.Auth)
		if err != nil {
			return err
		}
		emojis = emojiResp.GetPayload()
	} else if instanceType == "mastodon" {
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
	} else {
		endpoint := fmt.Sprintf("https://%s/api/emojis", instance)
		resp, err := http.Get(endpoint)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			slog.Error("failed to get misskey emojis", "status", resp.StatusCode, "instance", instance)
			return nil
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var misskeyResp MisskeyResponse
		if err := json.Unmarshal(body, &misskeyResp); err != nil {
			return err
		}

		for _, me := range misskeyResp.Emojis {
			category := "uncategorized"
			if me.Category != nil {
				category = *me.Category
			}

			emojis = append(emojis, &models.Emoji{
				Category:  category,
				Shortcode: me.Name,
				URL:       me.URL,
			})
		}
	}

	if category != "*" {
		emojis, _ = util.FilterEmojisByCategory(emojis, category)
	}

	totalCount := len(emojis)
	slog.Info("Emoji List Retrieved", "count", totalCount)

	if threadCount <= 0 {
		threadCount = runtime.NumCPU()
	}

	if threadCount > 1 {
		slog.Info("Starting multi-threaded download", "threads", threadCount)

		jobs := make(chan *models.Emoji, totalCount)
		var wg sync.WaitGroup

		for i := 0; i < threadCount; i++ {
			wg.Add(1)
			go downloadWorker(i+1, jobs, &wg, instance, override)
		}

		for _, emoji := range emojis {
			jobs <- emoji
		}
		close(jobs)

		wg.Wait()
	} else {
		for _, emoji := range emojis {
			if emoji.Category == "" {
				emoji.Category = "uncategorized"
			}
			slog.Info(fmt.Sprintf("downloading emoji %d/%d", 1, totalCount), "shortcode", emoji.Shortcode, "category", emoji.Category, "url", emoji.URL)
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
	}
	slog.Info(fmt.Sprintf("Completed! Downloaded %d emojis", totalCount))
	return nil
}
