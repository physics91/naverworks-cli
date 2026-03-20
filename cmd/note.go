package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/output"
	"github.com/spf13/cobra"
)

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "노트 관리",
}

var noteCreateCmd = &cobra.Command{
	Use:   "create <groupId>",
	Short: "노트 생성",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewNoteService(client)

		resp, err := svc.CreateNote(args[0])
		if err != nil {
			return err
		}
		if len(resp.Body) == 0 || strings.TrimSpace(string(resp.Body)) == "" {
			fmt.Println("{}")
		} else {
			output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		}
		return nil
	},
}

var noteDeleteCmd = &cobra.Command{
	Use:   "delete <groupId>",
	Short: "노트 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewNoteService(client)

		resp, err := svc.DeleteNote(args[0])
		if err != nil {
			return err
		}
		if len(resp.Body) == 0 || strings.TrimSpace(string(resp.Body)) == "" {
			fmt.Println("{}")
		} else {
			output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		}
		return nil
	},
}

var noteListPostsCmd = &cobra.Command{
	Use:   "list-posts <groupId>",
	Short: "노트 게시글 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewNoteService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"postId", "title"}, "posts")

		if all {
			var allItems []json.RawMessage
			for {
				resp, err := svc.ListPosts(args[0], cursor, count)
				if err != nil {
					return err
				}
				var page struct {
					Posts            []json.RawMessage `json:"posts"`
					ResponseMetaData struct {
						NextCursor string `json:"nextCursor"`
					} `json:"responseMetaData"`
				}
				json.Unmarshal(resp.Body, &page)
				allItems = append(allItems, page.Posts...)
				if page.ResponseMetaData.NextCursor == "" {
					break
				}
				cursor = page.ResponseMetaData.NextCursor
			}
			merged, _ := json.Marshal(map[string]interface{}{"posts": allItems})
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := svc.ListPosts(args[0], cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var noteGetPostCmd = &cobra.Command{
	Use:   "get-post <groupId> <postId>",
	Short: "노트 게시글 상세 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewNoteService(client)

		resp, err := svc.GetPost(args[0], args[1])
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var noteCreatePostCmd = &cobra.Command{
	Use:   "create-post <groupId>",
	Short: "노트 게시글 생성",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewNoteService(client)

		title, _ := cmd.Flags().GetString("title")
		body, _ := cmd.Flags().GetString("body")
		if title == "" {
			return fmt.Errorf("--title은 필수입니다")
		}

		post := map[string]interface{}{"title": title}
		if body != "" {
			post["body"] = body
		}

		resp, err := svc.CreatePost(args[0], post)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var noteUpdatePostCmd = &cobra.Command{
	Use:   "update-post <groupId> <postId>",
	Short: "노트 게시글 수정",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewNoteService(client)

		post := map[string]interface{}{}
		if title, _ := cmd.Flags().GetString("title"); title != "" {
			post["title"] = title
		}
		if body, _ := cmd.Flags().GetString("body"); body != "" {
			post["body"] = body
		}

		resp, err := svc.UpdatePost(args[0], args[1], post)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var noteDeletePostCmd = &cobra.Command{
	Use:   "delete-post <groupId> <postId>",
	Short: "노트 게시글 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewNoteService(client)

		resp, err := svc.DeletePost(args[0], args[1])
		if err != nil {
			return err
		}
		if len(resp.Body) == 0 || strings.TrimSpace(string(resp.Body)) == "" {
			fmt.Println("{}")
		} else {
			output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		}
		return nil
	},
}

func init() {
	noteListPostsCmd.Flags().String("cursor", "", "페이지네이션 커서")
	noteListPostsCmd.Flags().Int("count", 0, "페이지 크기")
	noteListPostsCmd.Flags().Bool("all", false, "전체 페이지 자동 순회")

	noteCreatePostCmd.Flags().String("title", "", "게시글 제목 (필수)")
	noteCreatePostCmd.Flags().String("body", "", "게시글 본문")

	noteUpdatePostCmd.Flags().String("title", "", "게시글 제목")
	noteUpdatePostCmd.Flags().String("body", "", "게시글 본문")

	noteCmd.AddCommand(noteCreateCmd, noteDeleteCmd, noteListPostsCmd, noteGetPostCmd,
		noteCreatePostCmd, noteUpdatePostCmd, noteDeletePostCmd)
	rootCmd.AddCommand(noteCmd)
}
