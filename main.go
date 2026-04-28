package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

const defaultHeadlineURL = "https://www.bloomberg.com/lineup-next/api/stories?types=ARTICLE%2CFEATURE%2CINTERACTIVE%2CLETTER%2CEXPLAINERS&locale=ja&limit=5"

type Config struct {
	HeadlineURL  string
	HeadlinePath string
	ArticleDir   string
	Headless     bool
	BrowserExec  string
	KeepOpen     bool
	Timeout      time.Duration
}

func main() {
	cfg := parseFlags()
	browserCtx, cleanup, err := newBrowser(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	if err := getHeadline(browserCtx, cfg); err != nil {
		log.Fatal(err)
	}
	if err := getArticles(browserCtx, cfg); err != nil {
		log.Fatal(err)
	}

	if cfg.KeepOpen {
		fmt.Println("Chrome is open. Press Enter to close.")
		_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
	}
}

func parseFlags() Config {
	headlineURL := flag.String("headline-url", defaultHeadlineURL, "headline API URL")
	headlinePath := flag.String("headline-out", "headline.json", "path to save headline JSON")
	articleDir := flag.String("article-dir", "articles", "directory to save article JSON files")
	headless := flag.Bool("headless", false, "run Chrome in headless mode")
	browserExec := flag.String("browser-exec", "", "Chrome executable path")
	keepOpen := flag.Bool("keep-open", false, "keep Chrome open until Enter is pressed")
	timeoutSec := flag.Int("timeout-sec", 60, "browser operation timeout in seconds")
	flag.Parse()

	return Config{
		HeadlineURL:  *headlineURL,
		HeadlinePath: *headlinePath,
		ArticleDir:   *articleDir,
		Headless:     *headless,
		BrowserExec:  *browserExec,
		KeepOpen:     *keepOpen,
		Timeout:      time.Duration(*timeoutSec) * time.Second,
	}
}

func newBrowser(cfg Config) (context.Context, func(), error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", cfg.Headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("lang", "ja-JP"),
		chromedp.Flag("start-maximized", true),
	)
	if cfg.BrowserExec != "" {
		opts = append(opts, chromedp.ExecPath(cfg.BrowserExec))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	cleanup := func() {
		cancelBrowser()
		cancelAlloc()
	}

	if err := chromedp.Run(withTimeout(browserCtx, cfg)); err != nil {
		cleanup()
		return nil, nil, err
	}
	return browserCtx, cleanup, nil
}

func withTimeout(parent context.Context, cfg Config) context.Context {
	ctx, _ := context.WithTimeout(parent, cfg.Timeout)
	return ctx
}

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
