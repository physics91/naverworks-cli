package cmd

import (
	"encoding/json"
	"os"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/output"
	"github.com/spf13/cobra"
)

var directoryCmd = &cobra.Command{
	Use:   "directory",
	Short: "디렉토리 관리 (사용자, 그룹)",
}

var dirListUsersCmd = &cobra.Command{
	Use:   "list-users",
	Short: "사용자 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"userId", "userName", "email"}, "users")

		if all {
			var allUsers []json.RawMessage
			for {
				resp, err := dir.ListUsers(cursor, count)
				if err != nil {
					return err
				}
				var page struct {
					Users            []json.RawMessage `json:"users"`
					ResponseMetaData struct {
						NextCursor string `json:"nextCursor"`
					} `json:"responseMetaData"`
				}
				json.Unmarshal(resp.Body, &page)
				allUsers = append(allUsers, page.Users...)
				if page.ResponseMetaData.NextCursor == "" {
					break
				}
				cursor = page.ResponseMetaData.NextCursor
			}
			merged, _ := json.Marshal(map[string]interface{}{"users": allUsers})
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := dir.ListUsers(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var dirGetUserCmd = &cobra.Command{
	Use:   "get-user <userId>",
	Short: "사용자 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		resp, err := dir.GetUser(args[0])
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var dirListGroupsCmd = &cobra.Command{
	Use:   "list-groups",
	Short: "그룹 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"groupId", "groupName"}, "groups")

		if all {
			var allGroups []json.RawMessage
			for {
				resp, err := dir.ListGroups(cursor, count)
				if err != nil {
					return err
				}
				var page struct {
					Groups           []json.RawMessage `json:"groups"`
					ResponseMetaData struct {
						NextCursor string `json:"nextCursor"`
					} `json:"responseMetaData"`
				}
				json.Unmarshal(resp.Body, &page)
				allGroups = append(allGroups, page.Groups...)
				if page.ResponseMetaData.NextCursor == "" {
					break
				}
				cursor = page.ResponseMetaData.NextCursor
			}
			merged, _ := json.Marshal(map[string]interface{}{"groups": allGroups})
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := dir.ListGroups(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var dirGetGroupCmd = &cobra.Command{
	Use:   "get-group <groupId>",
	Short: "그룹 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		resp, err := dir.GetGroup(args[0])
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

func init() {
	for _, cmd := range []*cobra.Command{dirListUsersCmd, dirListGroupsCmd} {
		cmd.Flags().String("cursor", "", "페이지네이션 커서")
		cmd.Flags().Int("count", 0, "페이지 크기")
		cmd.Flags().Bool("all", false, "전체 페이지 자동 순회")
	}
	directoryCmd.AddCommand(dirListUsersCmd, dirGetUserCmd, dirListGroupsCmd, dirGetGroupCmd)
	rootCmd.AddCommand(directoryCmd)
}
