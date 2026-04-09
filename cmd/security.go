package cmd

import (
	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "보안 설정 관리",
}

var secGetExternalBrowserCmd = &cobra.Command{
	Use:   "get-external-browser",
	Short: "외부 브라우저 설정 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewSecurityService(client).GetExternalBrowser()
		})
	},
}

var secEnableExternalBrowserCmd = &cobra.Command{
	Use:   "enable-external-browser",
	Short: "외부 브라우저 허용",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewSecurityService)
		if err != nil {
			return err
		}
		resp, err := svc.EnableExternalBrowser()
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var secDisableExternalBrowserCmd = &cobra.Command{
	Use:   "disable-external-browser",
	Short: "외부 브라우저 차단",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewSecurityService)
		if err != nil {
			return err
		}
		resp, err := svc.DisableExternalBrowser()
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

func init() {
	securityCmd.AddCommand(secGetExternalBrowserCmd, secEnableExternalBrowserCmd, secDisableExternalBrowserCmd)
	rootCmd.AddCommand(securityCmd)
}
