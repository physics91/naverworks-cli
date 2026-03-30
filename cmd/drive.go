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

// ─── MyDrive File Operations ───

var driveCopyCmd = &cobra.Command{
	Use:   "copy <fileId>",
	Short: "파일 복사",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).CopyFile(userID, args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveRenameCmd = &cobra.Command{
	Use:   "rename <fileId>",
	Short: "파일 이름 변경",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).RenameFile(userID, args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveMoveCmd = &cobra.Command{
	Use:   "move <fileId>",
	Short: "파일 이동",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).MoveFile(userID, args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveProtectCmd = &cobra.Command{
	Use:   "protect <fileId>",
	Short: "파일 보호 설정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).ProtectFile(userID, args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveUnprotectCmd = &cobra.Command{
	Use:   "unprotect <fileId>",
	Short: "파일 보호 해제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).UnprotectFile(userID, args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveLockCmd = &cobra.Command{
	Use:   "lock <fileId>",
	Short: "파일 잠금",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).LockFile(userID, args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveUnlockCmd = &cobra.Command{
	Use:   "unlock <fileId>",
	Short: "파일 잠금 해제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).UnlockFile(userID, args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── MyDrive Revisions ───

var driveRevisionCmd = &cobra.Command{
	Use:   "revision",
	Short: "파일 리비전 관리",
}

var driveRevisionListCmd = &cobra.Command{
	Use:   "list <fileId>",
	Short: "리비전 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		resp, err := api.NewDriveService(client).ListRevisions(userID, args[0], cursor, count)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveRevisionGetCmd = &cobra.Command{
	Use:   "get <fileId> <revisionId>",
	Short: "리비전 상세 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).GetRevision(userID, args[0], args[1])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveRevisionRestoreCmd = &cobra.Command{
	Use:   "restore <fileId> <revisionId>",
	Short: "리비전 복원",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).RestoreRevision(userID, args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveRevisionDownloadCmd = &cobra.Command{
	Use:   "download <fileId> <revisionId>",
	Short: "리비전 다운로드 URL 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		downloadURL, err := api.NewDriveService(client).GetRevisionDownloadURL(userID, args[0], args[1])
		if err != nil {
			return err
		}
		printDownloadURL(downloadURL)
		return nil
	},
}

// ─── MyDrive Trash (additional) ───

var driveTrashDeleteCmd = &cobra.Command{
	Use:   "trash-delete <fileId>",
	Short: "휴지통 파일 영구 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).DeleteTrashFile(userID, args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── MyDrive Link ───

var driveLinkCmd = &cobra.Command{
	Use:   "link",
	Short: "파일 링크 관리",
}

var driveLinkSettingCmd = &cobra.Command{
	Use:   "link-setting",
	Short: "링크 설정 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).GetLinkSetting(userID)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveLinkGetCmd = &cobra.Command{
	Use:   "get <fileId>",
	Short: "파일 링크 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).GetLink(userID, args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveLinkCreateCmd = &cobra.Command{
	Use:   "create <fileId>",
	Short: "파일 링크 생성",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).CreateLink(userID, args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveLinkUpdateCmd = &cobra.Command{
	Use:   "update <fileId>",
	Short: "파일 링크 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).PatchLink(userID, args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveLinkDeleteCmd = &cobra.Command{
	Use:   "delete <fileId>",
	Short: "파일 링크 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).DeleteLink(userID, args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── MyDrive Share ───

var driveShareCmd = &cobra.Command{
	Use:   "share",
	Short: "파일 공유 관리",
}

var driveShareGetCmd = &cobra.Command{
	Use:   "get <fileId>",
	Short: "파일 공유 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).GetShare(userID, args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveShareCreateCmd = &cobra.Command{
	Use:   "create <fileId>",
	Short: "파일 공유 생성",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).CreateShare(userID, args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveShareUpdateCmd = &cobra.Command{
	Use:   "update <fileId>",
	Short: "파일 공유 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).PatchShare(userID, args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveShareDeleteCmd = &cobra.Command{
	Use:   "delete <fileId>",
	Short: "파일 공유 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewDriveService(client).DeleteShare(userID, args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveShareListSubFoldersCmd = &cobra.Command{
	Use:   "list-sub-folders <fileId>",
	Short: "공유 하위 폴더 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		resp, err := api.NewDriveService(client).ListShareSubFolders(userID, args[0], cursor, count)
		if err != nil {
			return err
		}
		printBody(resp.Body)
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
		svc, err := newSvc(api.NewSharedDriveService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, nil, "sharedrives", svc.ListDrives)
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
		svc, err := newSvc(api.NewSharedDriveService)
		if err != nil {
			return err
		}
		return listFilesWithFolder(cmd, args[0], svc)
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
		svc, err := newSvc(api.NewSharedDriveService)
		if err != nil {
			return err
		}
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
		svc, err := newSvc(api.NewGroupFolderService)
		if err != nil {
			return err
		}
		return listFilesWithFolder(cmd, args[0], svc)
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
		svc := api.NewDriveService(client)
		return runListCmd(cmd, nil, "sharedfolders", func(c string, n int) (*api.Response, error) {
			return svc.ListSharedFolders(userID, c, n)
		})
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
		svc := api.NewDriveService(client)
		return runListCmd(cmd, nil, "files", func(c string, n int) (*api.Response, error) {
			return svc.ListSharedFolderFiles(userID, args[0], c, n)
		})
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
		// Task 5-1: File operations
		driveCopyCmd, driveRenameCmd, driveMoveCmd,
		driveProtectCmd, driveUnprotectCmd, driveLockCmd, driveUnlockCmd,
		// Task 5-2: Revisions
		driveRevisionListCmd, driveRevisionGetCmd, driveRevisionRestoreCmd, driveRevisionDownloadCmd,
		// Task 5-3: Trash delete, link-setting, link, share
		driveTrashDeleteCmd, driveLinkSettingCmd,
		driveLinkGetCmd, driveLinkCreateCmd, driveLinkUpdateCmd, driveLinkDeleteCmd,
		driveShareGetCmd, driveShareCreateCmd, driveShareUpdateCmd, driveShareDeleteCmd,
		driveShareListSubFoldersCmd,
	} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}

	// --json flag for commands that need JSON body
	for _, c := range []*cobra.Command{
		driveCopyCmd, driveRenameCmd, driveMoveCmd,
		driveLinkCreateCmd, driveLinkUpdateCmd,
		driveShareCreateCmd, driveShareUpdateCmd,
	} {
		c.Flags().String("json", "", "JSON 요청 본문 (- 이면 stdin)")
	}

	// Pagination flags for all list commands
	for _, c := range []*cobra.Command{
		driveListCmd, driveTrashListCmd,
		driveSharedListDrivesCmd, driveSharedListCmd,
		driveGroupListCmd,
		driveSharedFolderListCmd, driveSharedFolderFilesCmd,
		driveRevisionListCmd, driveShareListSubFoldersCmd,
	} {
		c.Flags().String("cursor", "", "페이지네이션 커서")
		c.Flags().Int("count", 0, "페이지 크기")
	}

	// Folder listing flag for file-browsing list commands
	for _, c := range []*cobra.Command{driveListCmd, driveSharedListCmd, driveGroupListCmd} {
		c.Flags().String("folder", "", "폴더 ID (하위 파일 조회)")
	}

	// SharedFolder also needs --user-id
	for _, c := range []*cobra.Command{driveSharedFolderListCmd, driveSharedFolderFilesCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}

	driveUploadCmd.Flags().String("folder", "", "업로드 대상 폴더 ID")
	driveMkdirCmd.Flags().String("name", "", "폴더 이름 (필수)")
	driveMkdirCmd.Flags().String("parent", "", "상위 폴더 ID")

	driveCmd.AddCommand(driveInfoCmd, driveListCmd, driveGetCmd, driveDownloadCmd,
		driveUploadCmd, driveMkdirCmd, driveDeleteCmd, driveTrashListCmd, driveTrashRestoreCmd)

	// Task 5-1: File operations
	driveCmd.AddCommand(driveCopyCmd, driveRenameCmd, driveMoveCmd,
		driveProtectCmd, driveUnprotectCmd, driveLockCmd, driveUnlockCmd)

	// Task 5-2: Revisions
	driveRevisionCmd.AddCommand(driveRevisionListCmd, driveRevisionGetCmd,
		driveRevisionRestoreCmd, driveRevisionDownloadCmd)
	driveCmd.AddCommand(driveRevisionCmd)

	// Task 5-3: Trash delete + Link + Share
	driveCmd.AddCommand(driveTrashDeleteCmd, driveLinkSettingCmd)

	driveLinkCmd.AddCommand(driveLinkGetCmd, driveLinkCreateCmd, driveLinkUpdateCmd, driveLinkDeleteCmd)
	driveCmd.AddCommand(driveLinkCmd)

	driveShareCmd.AddCommand(driveShareGetCmd, driveShareCreateCmd, driveShareUpdateCmd,
		driveShareDeleteCmd, driveShareListSubFoldersCmd)
	driveCmd.AddCommand(driveShareCmd)

	driveSharedCmd.AddCommand(driveSharedListDrivesCmd, driveSharedGetDriveCmd,
		driveSharedListCmd, driveSharedGetCmd, driveSharedDownloadCmd, driveSharedUploadCmd)
	driveCmd.AddCommand(driveSharedCmd)

	driveGroupCmd.AddCommand(driveGroupGetFolderCmd, driveGroupListCmd, driveGroupGetCmd)
	driveCmd.AddCommand(driveGroupCmd)

	driveSharedFolderCmd.AddCommand(driveSharedFolderListCmd, driveSharedFolderFilesCmd)
	driveCmd.AddCommand(driveSharedFolderCmd)

	rootCmd.AddCommand(driveCmd)
}
