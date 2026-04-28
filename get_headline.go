package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/chromedp/chromedp"
)

type HeadlineItem struct {
	ID       string `json:"id"`
	Slug     string `json:"slug"`
	Headline string `json:"headline"`
	URL      string `json:"url"`
}

func getHeadline(browserCtx context.Context, cfg Config) error {
	var jsonText string
	log.Printf("opening headline url: %s", cfg.HeadlineURL)
	if err := chromedp.Run(withTimeout(browserCtx, cfg),
		chromedp.Navigate(cfg.HeadlineURL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Text("body", &jsonText, chromedp.ByQuery),
	); err != nil {
		return fmt.Errorf("get headline JSON: %w", err)
	}

	var items []HeadlineItem
	if err := json.Unmarshal([]byte(jsonText), &items); err != nil {
		return fmt.Errorf("parse headline JSON: %w", err)
	}

	if err := ensureParentDir(cfg.HeadlinePath); err != nil {
		return err
	}
	if err := os.WriteFile(cfg.HeadlinePath, []byte(jsonText), 0o644); err != nil {
		return err
	}
	log.Printf("saved headline JSON: %s", cfg.HeadlinePath)
	return nil
}
