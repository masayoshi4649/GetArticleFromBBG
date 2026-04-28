package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chromedp/chromedp"
)

func getArticles(browserCtx context.Context, cfg Config) error {
	data, err := os.ReadFile(cfg.HeadlinePath)
	if err != nil {
		return fmt.Errorf("read headline file: %w", err)
	}

	var items []HeadlineItem
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("parse headline file: %w", err)
	}

	if err := os.MkdirAll(cfg.ArticleDir, 0o755); err != nil {
		return err
	}

	for index, item := range items {
		slug := strings.TrimSpace(item.Slug)
		if slug == "" {
			continue
		}
		url := "https://www.bloomberg.com/article/api/daper/stories?slug=" + slug
		var jsonText string
		log.Printf("[%d/%d] opening article url: %s", index+1, len(items), url)
		if err := chromedp.Run(withTimeout(browserCtx, cfg),
			chromedp.Navigate(url),
			chromedp.WaitVisible("body", chromedp.ByQuery),
			chromedp.Text("body", &jsonText, chromedp.ByQuery),
		); err != nil {
			return fmt.Errorf("get article JSON slug=%s: %w", slug, err)
		}

		if !json.Valid([]byte(jsonText)) {
			return fmt.Errorf("article JSON is invalid slug=%s", slug)
		}

		path := filepath.Join(cfg.ArticleDir, filepath.FromSlash(slug)+".json")
		if err := ensureParentDir(path); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(jsonText), 0o644); err != nil {
			return err
		}
		log.Printf("saved article JSON: %s", path)
	}
	return nil
}
