package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

var driveCmd = &cobra.Command{
	Use:   "drive",
	Short: "드라이브 관리",
}

// ─── MyDrive ───

var driveInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "내 드라이브 정보 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewDriveService(client)
		resp, err := svc.GetDriveInfo(userID)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveListCmd = &cobra.Command{
	Use:   "list",
	Short: "내 드라이브 파일 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		return listFilesWithFolder(cmd, userID, api.NewDriveService(client))
	},
}

var driveGetCmd = &cobra.Command{
	Use:   "get <fileId>",
	Short: "파일 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).GetFile(userID, args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveDownloadCmd = &cobra.Command{
	Use:   "download <fileId>",
	Short: "파일 다운로드 URL 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		downloadURL, err := api.NewDriveService(client).GetDownloadURL(userID, args[0])
		if err != nil {
			return err
		}
		printDownloadURL(downloadURL)
		return nil
	},
}

var driveUploadCmd = &cobra.Command{
	Use:   "upload <localPath>",
	Short: "파일 업로드",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewDriveService(client)

		localPath := args[0]
		folder, _ := cmd.Flags().GetString("folder")

		fileName, fileSize, err := statFileForUpload(localPath)
		if err != nil {
			return err
		}

		uploadBody := map[string]interface{}{"fileName": fileName}

		var resp *api.Response
		if folder != "" {
			resp, err = svc.CreateUploadURLInFolder(userID, folder, uploadBody, fileSize)
		} else {
			resp, err = svc.CreateUploadURL(userID, uploadBody, fileSize)
		}
		if err != nil {
			return err
		}

		if err := doUploadFromResponse(client, resp.Body, localPath); err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveMkdirCmd = &cobra.Command{
	Use:   "mkdir",
	Short: "폴더 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewDriveService(client)

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name은 필수입니다")
		}
		parent, _ := cmd.Flags().GetString("parent")

		var resp *api.Response
		if parent != "" {
			resp, err = svc.CreateFolderInParent(userID, parent, name)
		} else {
			resp, err = svc.CreateFolder(userID, name)
		}
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveDeleteCmd = &cobra.Command{
	Use:   "delete <fileId>",
	Short: "파일 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).DeleteFile(userID, args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveTrashListCmd = &cobra.Command{
	Use:   "trash-list",
	Short: "휴지통 파일 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		resp, err := api.NewDriveService(client).ListTrashFiles(userID, cursor, count)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveTrashRestoreCmd = &cobra.Command{
	Use:   "trash-restore <fileId>",
	Short: "휴지통 파일 복원",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).RestoreTrashFile(userID, args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── SharedDrive ───

var driveSharedCmd = &cobra.Command{
	Use:   "shared",
	Short: "공유 드라이브 관리",
}

var driveSharedListDrivesCmd = &cobra.Command{
	Use:   "list-drives",
	Short: "공유 드라이브 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewSharedDriveService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")

		resp, err := svc.ListDrives(cursor, count)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveSharedGetDriveCmd = &cobra.Command{
	Use:   "get-drive <driveId>",
	Short: "공유 드라이브 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewSharedDriveService(client).GetDrive(args[0])
		})
	},
}

var driveSharedListCmd = &cobra.Command{
	Use:   "list <driveId>",
	Short: "공유 드라이브 파일 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		return listFilesWithFolder(cmd, args[0], api.NewSharedDriveService(client))
	},
}

var driveSharedGetCmd = &cobra.Command{
	Use:   "get <driveId> <fileId>",
	Short: "공유 드라이브 파일 상세 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewSharedDriveService(client).GetFile(args[0], args[1])
		})
	},
}

var driveSharedDownloadCmd = &cobra.Command{
	Use:   "download <driveId> <fileId>",
	Short: "공유 드라이브 파일 다운로드 URL 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewSharedDriveService(client)

		downloadURL, err := svc.GetDownloadURL(args[0], args[1])
		if err != nil {
			return err
		}
		printDownloadURL(downloadURL)
		return nil
	},
}

var driveSharedUploadCmd = &cobra.Command{
	Use:   "upload <driveId> <localPath>",
	Short: "공유 드라이브 파일 업로드",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewSharedDriveService(client)

		localPath := args[1]

		fileName, fileSize, err := statFileForUpload(localPath)
		if err != nil {
			return err
		}

		uploadBody := map[string]interface{}{"fileName": fileName}

		resp, err := svc.CreateUploadURL(args[0], uploadBody, fileSize)
		if err != nil {
			return err
		}

		if err := doUploadFromResponse(client, resp.Body, localPath); err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

// ─── GroupFolder ───

var driveGroupCmd = &cobra.Command{
	Use:   "group",
	Short: "그룹 폴더 관리",
}

var driveGroupGetFolderCmd = &cobra.Command{
	Use:   "get-folder <groupId>",
	Short: "그룹 폴더 정보 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewGroupFolderService(client).GetFolder(args[0])
		})
	},
}

var driveGroupListCmd = &cobra.Command{
	Use:   "list <groupId>",
	Short: "그룹 폴더 파일 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		return listFilesWithFolder(cmd, args[0], api.NewGroupFolderService(client))
	},
}

var driveGroupGetCmd = &cobra.Command{
	Use:   "get <groupId> <fileId>",
	Short: "그룹 폴더 파일 상세 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewGroupFolderService(client).GetFile(args[0], args[1])
		})
	},
}

// ─── SharedFolder ───

var driveSharedFolderCmd = &cobra.Command{
	Use:   "shared-folder",
	Short: "공유 폴더 관리",
}

var driveSharedFolderListCmd = &cobra.Command{
	Use:   "list",
	Short: "공유 폴더 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		resp, err := api.NewDriveService(client).ListSharedFolders(userID, cursor, count)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveSharedFolderFilesCmd = &cobra.Command{
	Use:   "files <sharedFolderId>",
	Short: "공유 폴더 파일 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		resp, err := api.NewDriveService(client).ListSharedFolderFiles(userID, args[0], cursor, count)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

type driveLister interface {
	ListFiles(id, cursor string, count int) (*api.Response, error)
	ListFolderChildren(id, folder, cursor string, count int) (*api.Response, error)
}

func listFilesWithFolder(cmd *cobra.Command, id string, svc driveLister) error {
	cursor, _ := cmd.Flags().GetString("cursor")
	count, _ := cmd.Flags().GetInt("count")
	folder, _ := cmd.Flags().GetString("folder")
	var resp *api.Response
	var err error
	if folder != "" {
		resp, err = svc.ListFolderChildren(id, folder, cursor, count)
	} else {
		resp, err = svc.ListFiles(id, cursor, count)
	}
	if err != nil {
		return err
	}
	printBody(resp.Body)
	return nil
}

func statFileForUpload(localPath string) (fileName string, fileSize int64, err error) {
	stat, err := os.Stat(localPath)
	if err != nil {
		return "", 0, fmt.Errorf("파일 정보 조회 실패: %w", err)
	}
	return filepath.Base(localPath), stat.Size(), nil
}

func doUploadFromResponse(client *api.Client, respBody []byte, localPath string) error {
	var result struct {
		UploadURL string `json:"uploadUrl"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("업로드 URL 파싱 실패: %w", err)
	}
	if result.UploadURL == "" {
		return fmt.Errorf("업로드 URL을 받지 못했습니다")
	}
	return client.UploadFile(result.UploadURL, localPath)
}

func init() {
	// MyDrive commands with --user-id
	for _, c := range []*cobra.Command{
		driveInfoCmd, driveListCmd, driveGetCmd, driveDownloadCmd,
		driveUploadCmd, driveMkdirCmd, driveDeleteCmd,
		driveTrashListCmd, driveTrashRestoreCmd,
	} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}

	// Pagination flags
	for _, c := range []*cobra.Command{driveListCmd, driveTrashListCmd} {
		c.Flags().String("cursor", "", "페이지네이션 커서")
		c.Flags().Int("count", 0, "페이지 크기")
	}

	driveListCmd.Flags().String("folder", "", "폴더 ID (하위 파일 조회)")
	driveUploadCmd.Flags().String("folder", "", "업로드 대상 폴더 ID")
	driveMkdirCmd.Flags().String("name", "", "폴더 이름 (필수)")
	driveMkdirCmd.Flags().String("parent", "", "상위 폴더 ID")

	driveCmd.AddCommand(driveInfoCmd, driveListCmd, driveGetCmd, driveDownloadCmd,
		driveUploadCmd, driveMkdirCmd, driveDeleteCmd, driveTrashListCmd, driveTrashRestoreCmd)

	// SharedDrive
	for _, c := range []*cobra.Command{driveSharedListDrivesCmd, driveSharedListCmd} {
		c.Flags().String("cursor", "", "페이지네이션 커서")
		c.Flags().Int("count", 0, "페이지 크기")
	}
	driveSharedListCmd.Flags().String("folder", "", "폴더 ID (하위 파일 조회)")

	driveSharedCmd.AddCommand(driveSharedListDrivesCmd, driveSharedGetDriveCmd,
		driveSharedListCmd, driveSharedGetCmd, driveSharedDownloadCmd, driveSharedUploadCmd)
	driveCmd.AddCommand(driveSharedCmd)

	// GroupFolder
	for _, c := range []*cobra.Command{driveGroupListCmd} {
		c.Flags().String("cursor", "", "페이지네이션 커서")
		c.Flags().Int("count", 0, "페이지 크기")
		c.Flags().String("folder", "", "폴더 ID (하위 파일 조회)")
	}

	driveGroupCmd.AddCommand(driveGroupGetFolderCmd, driveGroupListCmd, driveGroupGetCmd)
	driveCmd.AddCommand(driveGroupCmd)

	// SharedFolder
	for _, c := range []*cobra.Command{driveSharedFolderListCmd, driveSharedFolderFilesCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
		c.Flags().String("cursor", "", "페이지네이션 커서")
		c.Flags().Int("count", 0, "페이지 크기")
	}

	driveSharedFolderCmd.AddCommand(driveSharedFolderListCmd, driveSharedFolderFilesCmd)
	driveCmd.AddCommand(driveSharedFolderCmd)

	rootCmd.AddCommand(driveCmd)
}
