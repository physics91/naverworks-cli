package cmd

import (
	"github.com/physics91/naverworks-cli/internal/api"
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
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewNoteService(client)

		resp, err := svc.CreateNote(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var noteDeleteCmd = &cobra.Command{
	Use:   "delete <groupId>",
	Short: "노트 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewNoteService(client)

		resp, err := svc.DeleteNote(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var noteListPostsCmd = &cobra.Command{
	Use:   "list-posts <groupId>",
	Short: "노트 게시글 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewNoteService(client)
		return runListCmd(cmd, []string{"postId", "title"}, "posts", func(c string, n int) (*api.Response, error) {
			return svc.ListPosts(args[0], c, n)
		})
	},
}

var noteGetPostCmd = &cobra.Command{
	Use:   "get-post <groupId> <postId>",
	Short: "노트 게시글 상세 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewNoteService(client).GetPost(args[0], args[1])
		})
	},
}

var noteCreatePostCmd = &cobra.Command{
	Use:   "create-post <groupId>",
	Short: "노트 게시글 생성",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewNoteService(client)

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

var noteUpdatePostCmd = &cobra.Command{
	Use:   "update-post <groupId> <postId>",
	Short: "노트 게시글 수정",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewNoteService(client)

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

var noteDeletePostCmd = &cobra.Command{
	Use:   "delete-post <groupId> <postId>",
	Short: "노트 게시글 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewNoteService(client)

		resp, err := svc.DeletePost(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

func init() {
	addListFlags(noteListPostsCmd)

	noteCreatePostCmd.Flags().String("title", "", "게시글 제목 (필수)")
	noteCreatePostCmd.Flags().String("body", "", "게시글 본문")

	noteUpdatePostCmd.Flags().String("title", "", "게시글 제목")
	noteUpdatePostCmd.Flags().String("body", "", "게시글 본문")

	noteCmd.AddCommand(noteCreateCmd, noteDeleteCmd, noteListPostsCmd, noteGetPostCmd,
		noteCreatePostCmd, noteUpdatePostCmd, noteDeletePostCmd)
	rootCmd.AddCommand(noteCmd)
}
