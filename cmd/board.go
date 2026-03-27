package cmd

import (
	"github.com/physics91/naverworks-cli/internal/api"
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
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"boardId", "boardName"}, "boards", svc.ListBoards)
	},
}

var boardGetCmd = &cobra.Command{
	Use:   "get <boardId>",
	Short: "게시판 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewBoardService(client).GetBoard(args[0])
		})
	},
}

var boardListPostsCmd = &cobra.Command{
	Use:   "list-posts <boardId>",
	Short: "게시글 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"postId", "title"}, "posts", func(c string, n int) (*api.Response, error) {
			return svc.ListPosts(args[0], c, n)
		})
	},
}

var boardGetPostCmd = &cobra.Command{
	Use:   "get-post <boardId> <postId>",
	Short: "게시글 상세 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewBoardService(client).GetPost(args[0], args[1])
		})
	},
}

var boardCreatePostCmd = &cobra.Command{
	Use:   "create-post <boardId>",
	Short: "게시글 생성",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		post, err := requireTitleBodyPost(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreatePost(args[0], post)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var boardUpdatePostCmd = &cobra.Command{
	Use:   "update-post <boardId> <postId>",
	Short: "게시글 수정",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		post, err := requireTitleBodyPost(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.UpdatePost(args[0], args[1], post)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var boardDeletePostCmd = &cobra.Command{
	Use:   "delete-post <boardId> <postId>",
	Short: "게시글 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		resp, err := svc.DeletePost(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var boardListCommentsCmd = &cobra.Command{
	Use:   "list-comments <boardId> <postId>",
	Short: "댓글 목록 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"commentId", "content"}, "comments", func(c string, n int) (*api.Response, error) {
			return svc.ListComments(args[0], args[1], c, n)
		})
	},
}

func init() {
	addListFlags(boardListCmd, boardListPostsCmd, boardListCommentsCmd)

	boardCreatePostCmd.Flags().String("title", "", "게시글 제목 (필수)")
	boardCreatePostCmd.Flags().String("body", "", "게시글 본문")

	boardUpdatePostCmd.Flags().String("title", "", "게시글 제목")
	boardUpdatePostCmd.Flags().String("body", "", "게시글 본문")

	boardCmd.AddCommand(boardListCmd, boardGetCmd, boardListPostsCmd, boardGetPostCmd,
		boardCreatePostCmd, boardUpdatePostCmd, boardDeletePostCmd, boardListCommentsCmd)
	rootCmd.AddCommand(boardCmd)
}
