package web

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/chromedp/chromedp"
)

const EntryURL = "https://talk.worksmobile.com/"

type Browser struct {
	SessionDir string
	Profile    string
	ExecPath   string
}

// Login keeps credential entry and additional authentication in the browser.
// It neither reads passwords nor converts the web session into an API token.
func (b Browser) Login(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	ctx, cancelTimeout := context.WithTimeout(ctx, 5*time.Minute)
	defer cancelTimeout()
	executable, err := browserExecutable(b.ExecPath)
	if err != nil {
		return err
	}
	if err := prepareSession(b.SessionDir, b.Profile); err != nil {
		return err
	}
	opts := b.options(executable, false)
	allocator, cancelAllocator := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAllocator()
	page, cancelPage := chromedp.NewContext(allocator)
	defer cancelPage()
	if err := chromedp.Run(page, chromedp.Navigate(EntryURL)); err != nil {
		return browserStartError(executable, err)
	}
	// Poll's JavaScript promise does not survive login redirects. Evaluate each
	// time in the current document instead and tolerate navigation transitions.
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	readyCount := 0
	for {
		var ready bool
		err := chromedp.Run(page, chromedp.Evaluate(authenticatedPage, &ready))
		if err == nil && ready {
			readyCount++
			if readyCount >= 3 {
				break
			}
		} else {
			readyCount = 0
		}
		select {
		case <-page.Done():
			return fmt.Errorf("브라우저 로그인을 완료하지 못했습니다. naverworks auth login --method browser로 다시 로그인하세요")
		case <-ticker.C:
		}
	}
	// Close through the protocol so Chrome flushes persistent cookie storage.
	if err := chromedp.Cancel(page); err != nil {
		return fmt.Errorf("브라우저 세션 저장 실패: %w", err)
	}
	return nil
}

const authenticatedPage = `location.origin === 'https://talk.worksmobile.com' && document.readyState === 'complete' && !!window.userContactNo && !!document.querySelector('#wrap.works_talk')`

func (b Browser) options(executable string, headless bool) []chromedp.ExecAllocatorOption {
	// chromedp's test defaults use a basic password store and mock keychain and
	// disable several browser protections. Keep Chrome's normal security defaults
	// for a browser which will hold a real user's credentials.
	opts := []chromedp.ExecAllocatorOption{chromedp.NoFirstRun, chromedp.NoDefaultBrowserCheck,
		chromedp.UserDataDir(b.SessionDir), chromedp.Flag("headless", headless),
		chromedp.Flag("lang", "ko-KR"), chromedp.Flag("remote-debugging-address", "127.0.0.1"),
		chromedp.Flag("enable-automation", true), chromedp.Flag("disable-sync", true),
		chromedp.Flag("no-sandbox", false)}
	if executable != "" {
		opts = append(opts, chromedp.ExecPath(executable))
	}
	return opts
}

func browserStartError(executable string, err error) error {
	if executable == "" {
		executable = "자동 검색"
	}
	return fmt.Errorf("브라우저 시작 실패 (%s): Chrome/Chromium 설치와 NW_BROWSER_PATH를 확인하세요: %w", executable, err)
}
