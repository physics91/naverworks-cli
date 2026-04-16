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

type boardIDCall func(*api.BoardService, string) (*api.Response, error)
type boardTwoIDCall func(*api.BoardService, string, string) (*api.Response, error)
type boardThreeIDCall func(*api.BoardService, string, string, string) (*api.Response, error)
type boardFourIDCall func(*api.BoardService, string, string, string, string) (*api.Response, error)
type boardListCall func(*api.BoardService, string, int) (*api.Response, error)
type boardIDListCall func(*api.BoardService, string, string, int) (*api.Response, error)
type boardTwoIDListCall func(*api.BoardService, string, string, string, int) (*api.Response, error)
type boardBodyReader func(*cobra.Command) (map[string]interface{}, error)
type boardBodyCall func(*api.BoardService, map[string]interface{}) (*api.Response, error)
type boardIDBodyCall func(*api.BoardService, string, map[string]interface{}) (*api.Response, error)
type boardTwoIDBodyCall func(*api.BoardService, string, string, map[string]interface{}) (*api.Response, error)
type boardThreeIDBodyCall func(*api.BoardService, string, string, string, map[string]interface{}) (*api.Response, error)

// Keep these wrappers local so cmd/helpers.go does not grow a board-only helper family.
func printBoardBody(resp *api.Response) {
	printBody(resp.Body)
}

func boardIDRunE(call boardIDCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		resp, err := call(svc, args[0])
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func boardTwoIDRunE(call boardTwoIDCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		resp, err := call(svc, args[0], args[1])
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func boardThreeIDRunE(call boardThreeIDCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		resp, err := call(svc, args[0], args[1], args[2])
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func boardFourIDRunE(call boardFourIDCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		resp, err := call(svc, args[0], args[1], args[2], args[3])
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func boardListRunListE(columns []string, itemKey string, call boardListCall) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, columns, itemKey, func(cursor string, count int) (*api.Response, error) {
			return call(svc, cursor, count)
		})
	}
}

func boardIDRunListE(columns []string, itemKey string, call boardIDListCall) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, columns, itemKey, func(cursor string, count int) (*api.Response, error) {
			return call(svc, args[0], cursor, count)
		})
	}
}

func boardTwoIDRunListE(columns []string, itemKey string, call boardTwoIDListCall) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, columns, itemKey, func(cursor string, count int) (*api.Response, error) {
			return call(svc, args[0], args[1], cursor, count)
		})
	}
}

func boardBodyRunE(readBody boardBodyReader, call boardBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		body, err := readBody(cmd)
		if err != nil {
			return err
		}
		resp, err := call(svc, body)
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func boardIDBodyRunE(readBody boardBodyReader, call boardIDBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		body, err := readBody(cmd)
		if err != nil {
			return err
		}
		resp, err := call(svc, args[0], body)
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func boardTwoIDBodyRunE(readBody boardBodyReader, call boardTwoIDBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		body, err := readBody(cmd)
		if err != nil {
			return err
		}
		resp, err := call(svc, args[0], args[1], body)
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func boardThreeIDBodyRunE(readBody boardBodyReader, call boardThreeIDBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBoardService)
		if err != nil {
			return err
		}
		body, err := readBody(cmd)
		if err != nil {
			return err
		}
		resp, err := call(svc, args[0], args[1], args[2], body)
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

var boardListCmd = &cobra.Command{
	Use:   "list",
	Short: "게시판 목록 조회",
	RunE:  boardListRunListE([]string{"boardId", "boardName"}, "boards", (*api.BoardService).ListBoards),
}

var boardGetCmd = &cobra.Command{
	Use:   "get <boardId>",
	Short: "게시판 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  boardIDRunE((*api.BoardService).GetBoard, printBoardBody),
}

var boardCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "게시판 생성",
	RunE:  boardBodyRunE(readJSONFlag, (*api.BoardService).CreateBoard, printBoardBody),
}

var boardUpdateCmd = &cobra.Command{
	Use:   "update <boardId>",
	Short: "게시판 수정",
	Args:  cobra.ExactArgs(1),
	RunE:  boardIDBodyRunE(readJSONFlag, (*api.BoardService).UpdateBoard, printBoardBody),
}

var boardDeleteCmd = &cobra.Command{
	Use:   "delete <boardId>",
	Short: "게시판 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  boardIDRunE((*api.BoardService).DeleteBoard, printResponse),
}

// ─── Posts ───

var boardListPostsCmd = &cobra.Command{
	Use:   "list-posts <boardId>",
	Short: "게시글 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  boardIDRunListE([]string{"postId", "title"}, "posts", (*api.BoardService).ListPosts),
}

var boardGetPostCmd = &cobra.Command{
	Use:   "get-post <boardId> <postId>",
	Short: "게시글 상세 조회",
	Args:  cobra.ExactArgs(2),
	RunE:  boardTwoIDRunE((*api.BoardService).GetPost, printBoardBody),
}

var boardCreatePostCmd = &cobra.Command{
	Use:   "create-post <boardId>",
	Short: "게시글 생성",
	Args:  cobra.ExactArgs(1),
	RunE:  boardIDBodyRunE(requireTitleBodyPost, (*api.BoardService).CreatePost, printBoardBody),
}

var boardUpdatePostCmd = &cobra.Command{
	Use:   "update-post <boardId> <postId>",
	Short: "게시글 수정",
	Args:  cobra.ExactArgs(2),
	RunE:  boardTwoIDBodyRunE(requireTitleBodyPost, (*api.BoardService).UpdatePost, printBoardBody),
}

var boardDeletePostCmd = &cobra.Command{
	Use:   "delete-post <boardId> <postId>",
	Short: "게시글 삭제",
	Args:  cobra.ExactArgs(2),
	RunE:  boardTwoIDRunE((*api.BoardService).DeletePost, printResponse),
}

// ─── Post Readers ───

var boardListReadersCmd = &cobra.Command{
	Use:   "list-readers <boardId> <postId>",
	Short: "게시글 읽은 사람 목록 조회",
	Args:  cobra.ExactArgs(2),
	RunE:  boardTwoIDRunListE([]string{"userId", "userName"}, "readers", (*api.BoardService).ListPostReaders),
}

// ─── Aggregation Queries ───

var boardListRecentCmd = &cobra.Command{
	Use:   "list-recent",
	Short: "최근 게시글 목록 조회",
	RunE:  boardListRunListE([]string{"postId", "title"}, "posts", (*api.BoardService).ListRecentPosts),
}

var boardListMyCmd = &cobra.Command{
	Use:   "list-my",
	Short: "내 게시글 목록 조회",
	RunE:  boardListRunListE([]string{"postId", "title"}, "posts", (*api.BoardService).ListMyPosts),
}

var boardListMustCmd = &cobra.Command{
	Use:   "list-must",
	Short: "필독 게시글 목록 조회",
	RunE:  boardListRunListE([]string{"postId", "title"}, "posts", (*api.BoardService).ListMustPosts),
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
	RunE:  boardTwoIDRunE((*api.BoardService).ListPostAttachments, printBoardBody),
}

var boardGetAttachmentCmd = &cobra.Command{
	Use:   "get-attachment <boardId> <postId> <attachmentId>",
	Short: "게시글 첨부파일 상세 조회",
	Args:  cobra.ExactArgs(3),
	RunE:  boardThreeIDRunE((*api.BoardService).GetPostAttachment, printBoardBody),
}

var boardDeleteAttachmentCmd = &cobra.Command{
	Use:   "delete-attachment <boardId> <postId> <attachmentId>",
	Short: "게시글 첨부파일 삭제",
	Args:  cobra.ExactArgs(3),
	RunE:  boardThreeIDRunE((*api.BoardService).DeletePostAttachment, printResponse),
}

// ─── Comments ───

var boardListCommentsCmd = &cobra.Command{
	Use:   "list-comments <boardId> <postId>",
	Short: "댓글 목록 조회",
	Args:  cobra.ExactArgs(2),
	RunE:  boardTwoIDRunListE([]string{"commentId", "content"}, "comments", (*api.BoardService).ListComments),
}

var boardCreateCommentCmd = &cobra.Command{
	Use:   "create-comment <boardId> <postId>",
	Short: "댓글 생성",
	Args:  cobra.ExactArgs(2),
	RunE:  boardTwoIDBodyRunE(readJSONFlag, (*api.BoardService).CreateComment, printBoardBody),
}

var boardGetCommentCmd = &cobra.Command{
	Use:   "get-comment <boardId> <postId> <commentId>",
	Short: "댓글 상세 조회",
	Args:  cobra.ExactArgs(3),
	RunE:  boardThreeIDRunE((*api.BoardService).GetComment, printBoardBody),
}

var boardUpdateCommentCmd = &cobra.Command{
	Use:   "update-comment <boardId> <postId> <commentId>",
	Short: "댓글 수정",
	Args:  cobra.ExactArgs(3),
	RunE:  boardThreeIDBodyRunE(readJSONFlag, (*api.BoardService).UpdateComment, printBoardBody),
}

var boardDeleteCommentCmd = &cobra.Command{
	Use:   "delete-comment <boardId> <postId> <commentId>",
	Short: "댓글 삭제",
	Args:  cobra.ExactArgs(3),
	RunE:  boardThreeIDRunE((*api.BoardService).DeleteComment, printResponse),
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
	RunE:  boardThreeIDRunE((*api.BoardService).ListCommentAttachments, printBoardBody),
}

var boardGetCommentAttachmentCmd = &cobra.Command{
	Use:   "get-comment-attachment <boardId> <postId> <commentId> <attachmentId>",
	Short: "댓글 첨부파일 상세 조회",
	Args:  cobra.ExactArgs(4),
	RunE:  boardFourIDRunE((*api.BoardService).GetCommentAttachment, printBoardBody),
}

var boardDeleteCommentAttachmentCmd = &cobra.Command{
	Use:   "delete-comment-attachment <boardId> <postId> <commentId> <attachmentId>",
	Short: "댓글 첨부파일 삭제",
	Args:  cobra.ExactArgs(4),
	RunE:  boardFourIDRunE((*api.BoardService).DeleteCommentAttachment, printResponse),
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
