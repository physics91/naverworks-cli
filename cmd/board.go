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

var boardCmd = &cobra.Command{
	Use:   "board",
	Short: "게시판 관리",
}

var boardListCmd = &cobra.Command{
	Use:   "list",
	Short: "게시판 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewBoardService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"boardId", "boardName"}, "boards")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return svc.ListBoards(c, count)
			}, "boards")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"boards": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := svc.ListBoards(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var boardGetCmd = &cobra.Command{
	Use:   "get <boardId>",
	Short: "게시판 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewBoardService(client)

		resp, err := svc.GetBoard(args[0])
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var boardListPostsCmd = &cobra.Command{
	Use:   "list-posts <boardId>",
	Short: "게시글 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewBoardService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"postId", "title"}, "posts")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return svc.ListPosts(args[0], c, count)
			}, "posts")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"posts": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
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

var boardGetPostCmd = &cobra.Command{
	Use:   "get-post <boardId> <postId>",
	Short: "게시글 상세 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewBoardService(client)

		resp, err := svc.GetPost(args[0], args[1])
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var boardCreatePostCmd = &cobra.Command{
	Use:   "create-post <boardId>",
	Short: "게시글 생성",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewBoardService(client)

		title, _ := cmd.Flags().GetString("title")
		body, _ := cmd.Flags().GetString("body")
		if title == "" {
			return fmt.Errorf("--title은 필수입니다")
		}
		if body == "" {
			return fmt.Errorf("--body는 필수입니다")
		}

		post := map[string]interface{}{"title": title, "body": body}

		resp, err := svc.CreatePost(args[0], post)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var boardUpdatePostCmd = &cobra.Command{
	Use:   "update-post <boardId> <postId>",
	Short: "게시글 수정",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewBoardService(client)

		title, _ := cmd.Flags().GetString("title")
		body, _ := cmd.Flags().GetString("body")
		if title == "" {
			return fmt.Errorf("--title은 필수입니다")
		}
		if body == "" {
			return fmt.Errorf("--body는 필수입니다")
		}

		post := map[string]interface{}{"title": title, "body": body}

		resp, err := svc.UpdatePost(args[0], args[1], post)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var boardDeletePostCmd = &cobra.Command{
	Use:   "delete-post <boardId> <postId>",
	Short: "게시글 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewBoardService(client)

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

var boardListCommentsCmd = &cobra.Command{
	Use:   "list-comments <boardId> <postId>",
	Short: "댓글 목록 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewBoardService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"commentId", "content"}, "comments")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return svc.ListComments(args[0], args[1], c, count)
			}, "comments")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"comments": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := svc.ListComments(args[0], args[1], cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

func init() {
	for _, c := range []*cobra.Command{boardListCmd, boardListPostsCmd, boardListCommentsCmd} {
		c.Flags().String("cursor", "", "페이지네이션 커서")
		c.Flags().Int("count", 0, "페이지 크기")
		c.Flags().Bool("all", false, "전체 페이지 자동 순회")
	}

	boardCreatePostCmd.Flags().String("title", "", "게시글 제목 (필수)")
	boardCreatePostCmd.Flags().String("body", "", "게시글 본문")

	boardUpdatePostCmd.Flags().String("title", "", "게시글 제목")
	boardUpdatePostCmd.Flags().String("body", "", "게시글 본문")

	boardCmd.AddCommand(boardListCmd, boardGetCmd, boardListPostsCmd, boardGetPostCmd,
		boardCreatePostCmd, boardUpdatePostCmd, boardDeletePostCmd, boardListCommentsCmd)
	rootCmd.AddCommand(boardCmd)
}
