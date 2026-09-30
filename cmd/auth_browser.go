package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/physics91/naverworks-cli/internal/config"
	"github.com/physics91/naverworks-cli/internal/web"
	"github.com/spf13/cobra"
)

func activeBrowser() (web.Browser, error) {
	_, profile, err := loadActiveConfig()
	if err != nil {
		return web.Browser{}, err
	}
	configPath, err := config.DefaultPathOrError()
	if err != nil {
		return web.Browser{}, err
	}
	dir, err := web.SessionDir(configPath, profile)
	if err != nil {
		return web.Browser{}, err
	}
	return web.Browser{SessionDir: dir, Profile: profile}, nil
}

func loginBrowser(cmd *cobra.Command) error {
	browser, err := activeBrowser()
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.ErrOrStderr(), "전용 브라우저에서 네이버웍스에 로그인하세요. 비밀번호와 추가 인증은 브라우저에서 직접 입력합니다.")
	if err := browser.Login(cmd.Context()); err != nil {
		return err
	}
	body, err := json.Marshal(map[string]string{"auth_method": "browser", "profile": browser.Profile})
	if err != nil {
		return err
	}
	printBody(body)
	return nil
}

func browserAuthRequested(cmd *cobra.Command) (bool, error) {
	method, _ := cmd.Flags().GetString("method")
	if method != "api" && method != "browser" {
		return false, fmt.Errorf("유효하지 않은 인증 방식: %s (api|browser)", method)
	}
	if method == "browser" && previewRequested() {
		return false, fmt.Errorf("브라우저 인증 명령에는 미리보기 옵션을 사용할 수 없습니다")
	}
	return method == "browser", nil
}

func statusBrowser(cmd *cobra.Command) error {
	browser, err := activeBrowser()
	if err != nil {
		return err
	}
	client, err := browser.Open(cmd.Context())
	if err != nil {
		return err
	}
	defer client.Close()
	body, err := json.Marshal(map[string]any{"auth_method": "browser", "profile": browser.Profile, "authenticated": true, "headless": true})
	if err != nil {
		return err
	}
	printBody(body)
	return nil
}

func logoutBrowser() error {
	browser, err := activeBrowser()
	if err != nil {
		return err
	}
	if err := browser.Logout(); err != nil {
		return err
	}
	body, err := json.Marshal(map[string]any{"auth_method": "browser", "profile": browser.Profile, "local_session_deleted": true})
	if err != nil {
		return err
	}
	printBody(body)
	return nil
}
