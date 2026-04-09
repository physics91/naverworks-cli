package cmd

import (
	"fmt"

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
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewBoardService(client).GetBoard(args[0])
		})
	},
}

var boardCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "게시판 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateBoard(body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var boardUpdateCmd = &cobra.Command{
	Use:   "update <boardId>",
	Short: "게시판 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.UpdateBoard(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var boardDeleteCmd = &cobra.Command{
	Use:   "delete <boardId>",
	Short: "게시판 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		resp, err := svc.DeleteBoard(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── Posts ───

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
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
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

// ─── Post Readers ───

var boardListReadersCmd = &cobra.Command{
	Use:   "list-readers <boardId> <postId>",
	Short: "게시글 읽은 사람 목록 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"userId", "userName"}, "readers", func(c string, n int) (*api.Response, error) {
			return svc.ListPostReaders(args[0], args[1], c, n)
		})
	},
}

// ─── Aggregation Queries ───

var boardListRecentCmd = &cobra.Command{
	Use:   "list-recent",
	Short: "최근 게시글 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"postId", "title"}, "posts", svc.ListRecentPosts)
	},
}

var boardListMyCmd = &cobra.Command{
	Use:   "list-my",
	Short: "내 게시글 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"postId", "title"}, "posts", svc.ListMyPosts)
	},
}

var boardListMustCmd = &cobra.Command{
	Use:   "list-must",
	Short: "필독 게시글 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"postId", "title"}, "posts", svc.ListMustPosts)
	},
}

// ─── Post Attachments ───

var boardCreateAttachmentCmd = &cobra.Command{
	Use:   "create-attachment <boardId> <postId>",
	Short: "게시글 첨부파일 업로드",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewBoardService(client)

		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			return fmt.Errorf("--file 플래그가 필요합니다")
		}

		fileName, fileSize, err := statFileForUpload(filePath)
		if err != nil {
			return err
		}

		uploadBody := map[string]interface{}{"fileName": fileName, "fileSize": fileSize}
		resp, err := svc.CreatePostAttachment(args[0], args[1], uploadBody)
		if err != nil {
			return err
		}

		if err := doUploadFromResponse(client, resp.Body, filePath); err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var boardListAttachmentsCmd = &cobra.Command{
	Use:   "list-attachments <boardId> <postId>",
	Short: "게시글 첨부파일 목록 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewBoardService(client).ListPostAttachments(args[0], args[1])
		})
	},
}

var boardGetAttachmentCmd = &cobra.Command{
	Use:   "get-attachment <boardId> <postId> <attachmentId>",
	Short: "게시글 첨부파일 상세 조회",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewBoardService(client).GetPostAttachment(args[0], args[1], args[2])
		})
	},
}

var boardDeleteAttachmentCmd = &cobra.Command{
	Use:   "delete-attachment <boardId> <postId> <attachmentId>",
	Short: "게시글 첨부파일 삭제",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		resp, err := svc.DeletePostAttachment(args[0], args[1], args[2])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── Comments ───

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

var boardCreateCommentCmd = &cobra.Command{
	Use:   "create-comment <boardId> <postId>",
	Short: "댓글 생성",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateComment(args[0], args[1], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var boardGetCommentCmd = &cobra.Command{
	Use:   "get-comment <boardId> <postId> <commentId>",
	Short: "댓글 상세 조회",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewBoardService(client).GetComment(args[0], args[1], args[2])
		})
	},
}

var boardUpdateCommentCmd = &cobra.Command{
	Use:   "update-comment <boardId> <postId> <commentId>",
	Short: "댓글 수정",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.UpdateComment(args[0], args[1], args[2], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var boardDeleteCommentCmd = &cobra.Command{
	Use:   "delete-comment <boardId> <postId> <commentId>",
	Short: "댓글 삭제",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		resp, err := svc.DeleteComment(args[0], args[1], args[2])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── Comment Attachments ───

var boardCreateCommentAttachmentCmd = &cobra.Command{
	Use:   "create-comment-attachment <boardId> <postId> <commentId>",
	Short: "댓글 첨부파일 업로드",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewBoardService(client)

		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			return fmt.Errorf("--file 플래그가 필요합니다")
		}

		fileName, fileSize, err := statFileForUpload(filePath)
		if err != nil {
			return err
		}

		uploadBody := map[string]interface{}{"fileName": fileName, "fileSize": fileSize}
		resp, err := svc.CreateCommentAttachment(args[0], args[1], args[2], uploadBody)
		if err != nil {
			return err
		}

		if err := doUploadFromResponse(client, resp.Body, filePath); err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var boardListCommentAttachmentsCmd = &cobra.Command{
	Use:   "list-comment-attachments <boardId> <postId> <commentId>",
	Short: "댓글 첨부파일 목록 조회",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewBoardService(client).ListCommentAttachments(args[0], args[1], args[2])
		})
	},
}

var boardGetCommentAttachmentCmd = &cobra.Command{
	Use:   "get-comment-attachment <boardId> <postId> <commentId> <attachmentId>",
	Short: "댓글 첨부파일 상세 조회",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewBoardService(client).GetCommentAttachment(args[0], args[1], args[2], args[3])
		})
	},
}

var boardDeleteCommentAttachmentCmd = &cobra.Command{
	Use:   "delete-comment-attachment <boardId> <postId> <commentId> <attachmentId>",
	Short: "댓글 첨부파일 삭제",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		resp, err := svc.DeleteCommentAttachment(args[0], args[1], args[2], args[3])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

func init() {
	// Pagination flags
	addListFlags(boardListCmd, boardListPostsCmd, boardListCommentsCmd,
		boardListReadersCmd, boardListRecentCmd, boardListMyCmd, boardListMustCmd)

	// Post flags
	boardCreatePostCmd.Flags().String("title", "", "게시글 제목 (필수)")
	boardCreatePostCmd.Flags().String("body", "", "게시글 본문")
	boardUpdatePostCmd.Flags().String("title", "", "게시글 제목")
	boardUpdatePostCmd.Flags().String("body", "", "게시글 본문")

	// JSON flags
	boardCreateCmd.Flags().String("json", "", "게시판 생성 JSON")
	boardUpdateCmd.Flags().String("json", "", "게시판 수정 JSON")
	boardCreateCommentCmd.Flags().String("json", "", "댓글 생성 JSON")
	boardUpdateCommentCmd.Flags().String("json", "", "댓글 수정 JSON")

	// File flags
	boardCreateAttachmentCmd.Flags().String("file", "", "업로드 파일 경로")
	boardCreateCommentAttachmentCmd.Flags().String("file", "", "업로드 파일 경로")

	boardCmd.AddCommand(
		// Board CRUD
		boardListCmd, boardGetCmd, boardCreateCmd, boardUpdateCmd, boardDeleteCmd,
		// Posts
		boardListPostsCmd, boardGetPostCmd, boardCreatePostCmd, boardUpdatePostCmd, boardDeletePostCmd,
		// Post readers
		boardListReadersCmd,
		// Aggregation queries
		boardListRecentCmd, boardListMyCmd, boardListMustCmd,
		// Post attachments
		boardCreateAttachmentCmd, boardListAttachmentsCmd, boardGetAttachmentCmd, boardDeleteAttachmentCmd,
		// Comments
		boardListCommentsCmd, boardCreateCommentCmd, boardGetCommentCmd, boardUpdateCommentCmd, boardDeleteCommentCmd,
		// Comment attachments
		boardCreateCommentAttachmentCmd, boardListCommentAttachmentsCmd, boardGetCommentAttachmentCmd, boardDeleteCommentAttachmentCmd,
	)
	rootCmd.AddCommand(boardCmd)
}
