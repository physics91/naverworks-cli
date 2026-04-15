package cmd

import (
	"fmt"

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
		resume, _ := cmd.Flags().GetBool("resume")

		fileName, fileSize, err := statFileForUpload(localPath)
		if err != nil {
			return err
		}

		uploadBody := map[string]interface{}{"fileName": fileName}
		if resume {
			uploadBody["resume"] = true
		}

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
		svc := api.NewDriveService(client)
		return runListCmd(cmd, nil, "files", func(c string, n int) (*api.Response, error) {
			return svc.ListTrashFiles(userID, c, n)
		})
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
		svc := api.NewDriveService(client)
		return runListCmd(cmd, nil, "revisions", func(c string, n int) (*api.Response, error) {
			return svc.ListRevisions(userID, args[0], c, n)
		})
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
		svc := api.NewDriveService(client)
		return runListCmd(cmd, nil, "folders", func(c string, n int) (*api.Response, error) {
			return svc.ListShareSubFolders(userID, args[0], c, n)
		})
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
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
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
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
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
	Use:   "upload <driveId>",
	Short: "공유 드라이브 파일 업로드",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewSharedDriveService(client)

		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			return fmt.Errorf("--file 플래그가 필요합니다")
		}

		fileName, fileSize, err := statFileForUpload(filePath)
		if err != nil {
			return err
		}

		uploadBody := map[string]interface{}{"fileName": fileName}
		folder, _ := cmd.Flags().GetString("folder")
		resume, _ := cmd.Flags().GetBool("resume")
		if resume {
			uploadBody["resume"] = true
		}

		var resp *api.Response
		if folder != "" {
			resp, err = svc.CreateUploadURLInFolder(args[0], folder, uploadBody, fileSize)
		} else {
			resp, err = svc.CreateUploadURL(args[0], uploadBody, fileSize)
		}
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

// ─── SharedDrive Task 5-8: 관리 + 파일 보완 ───

var driveSharedCreateDriveCmd = &cobra.Command{
	Use:   "create-drive",
	Short: "공유 드라이브 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).CreateDrive(body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedUpdateDriveCmd = &cobra.Command{
	Use:   "update-drive <driveId>",
	Short: "공유 드라이브 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).PatchDrive(args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedDeleteDriveCmd = &cobra.Command{
	Use:   "delete-drive <driveId>",
	Short: "공유 드라이브 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).DeleteDrive(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedMkdirCmd = &cobra.Command{
	Use:   "mkdir <driveId>",
	Short: "공유 드라이브 폴더 생성",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewSharedDriveService(client)
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		parent, _ := cmd.Flags().GetString("parent")
		var resp *api.Response
		if parent != "" {
			resp, err = svc.CreateSubFolder(args[0], parent, body)
		} else {
			resp, err = svc.CreateFolderInRoot(args[0], body)
		}
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedDeleteCmd = &cobra.Command{
	Use:   "delete <driveId> <fileId>",
	Short: "공유 드라이브 파일 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).DeleteFile(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── SharedDrive Task 5-9: 파일조작 ───

var driveSharedCopyCmd = &cobra.Command{
	Use:   "copy <driveId> <fileId>",
	Short: "공유 드라이브 파일 복사",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).CopyFile(args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedRenameCmd = &cobra.Command{
	Use:   "rename <driveId> <fileId>",
	Short: "공유 드라이브 파일 이름 변경",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).RenameFile(args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedMoveCmd = &cobra.Command{
	Use:   "move <driveId> <fileId>",
	Short: "공유 드라이브 파일 이동",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).MoveFile(args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedProtectCmd = &cobra.Command{
	Use:   "protect <driveId> <fileId>",
	Short: "공유 드라이브 파일 보호 설정",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).ProtectFile(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedUnprotectCmd = &cobra.Command{
	Use:   "unprotect <driveId> <fileId>",
	Short: "공유 드라이브 파일 보호 해제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).UnprotectFile(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedLockCmd = &cobra.Command{
	Use:   "lock <driveId> <fileId>",
	Short: "공유 드라이브 파일 잠금",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).LockFile(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedUnlockCmd = &cobra.Command{
	Use:   "unlock <driveId> <fileId>",
	Short: "공유 드라이브 파일 잠금 해제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).UnlockFile(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── SharedDrive Task 5-10: 리비전 + 휴지통 ───

var driveSharedRevisionCmd = &cobra.Command{
	Use:   "revision",
	Short: "공유 드라이브 파일 리비전 관리",
}

var driveSharedRevisionListCmd = &cobra.Command{
	Use:   "list <driveId> <fileId>",
	Short: "공유 드라이브 파일 리비전 목록 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewSharedDriveService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, nil, "revisions", func(c string, n int) (*api.Response, error) {
			return svc.ListRevisions(args[0], args[1], c, n)
		})
	},
}

var driveSharedRevisionGetCmd = &cobra.Command{
	Use:   "get <driveId> <fileId> <revisionId>",
	Short: "공유 드라이브 파일 리비전 상세 조회",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewSharedDriveService(client).GetRevision(args[0], args[1], args[2])
		})
	},
}

var driveSharedRevisionRestoreCmd = &cobra.Command{
	Use:   "restore <driveId> <fileId> <revisionId>",
	Short: "공유 드라이브 파일 리비전 복원",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).RestoreRevision(args[0], args[1], args[2])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedRevisionDownloadCmd = &cobra.Command{
	Use:   "download <driveId> <fileId> <revisionId>",
	Short: "공유 드라이브 파일 리비전 다운로드 URL 조회",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewSharedDriveService)
		if err != nil {
			return err
		}
		downloadURL, err := svc.GetRevisionDownloadURL(args[0], args[1], args[2])
		if err != nil {
			return err
		}
		printDownloadURL(downloadURL)
		return nil
	},
}

var driveSharedTrashListCmd = &cobra.Command{
	Use:   "trash-list <driveId>",
	Short: "공유 드라이브 휴지통 파일 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewSharedDriveService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, nil, "files", func(c string, n int) (*api.Response, error) {
			return svc.ListTrashFiles(args[0], c, n)
		})
	},
}

var driveSharedTrashRestoreCmd = &cobra.Command{
	Use:   "trash-restore <driveId> <fileId>",
	Short: "공유 드라이브 휴지통 파일 복원",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).RestoreTrashFile(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedTrashDeleteCmd = &cobra.Command{
	Use:   "trash-delete <driveId> <fileId>",
	Short: "공유 드라이브 휴지통 파일 영구 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).DeleteTrashFile(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── SharedDrive Task 5-11: 링크 + 권한 ───

var driveSharedLinkSettingCmd = &cobra.Command{
	Use:   "link-setting <driveId>",
	Short: "공유 드라이브 링크 설정 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewSharedDriveService(client).GetLinkSetting(args[0])
		})
	},
}

var driveSharedLinkCmd = &cobra.Command{
	Use:   "link",
	Short: "공유 드라이브 파일 링크 관리",
}

var driveSharedLinkGetCmd = &cobra.Command{
	Use:   "get <driveId> <fileId>",
	Short: "공유 드라이브 파일 링크 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewSharedDriveService(client).GetLink(args[0], args[1])
		})
	},
}

var driveSharedLinkCreateCmd = &cobra.Command{
	Use:   "create <driveId> <fileId>",
	Short: "공유 드라이브 파일 링크 생성",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).CreateLink(args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedLinkUpdateCmd = &cobra.Command{
	Use:   "update <driveId> <fileId>",
	Short: "공유 드라이브 파일 링크 수정",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).PatchLink(args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedLinkDeleteCmd = &cobra.Command{
	Use:   "delete <driveId> <fileId>",
	Short: "공유 드라이브 파일 링크 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).DeleteLink(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// Drive-level permissions

var driveSharedPermissionCmd = &cobra.Command{
	Use:   "permission",
	Short: "공유 드라이브 권한 관리",
}

var driveSharedPermissionListCmd = &cobra.Command{
	Use:   "list <driveId>",
	Short: "공유 드라이브 권한 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewSharedDriveService(client).ListPermissions(args[0])
		})
	},
}

var driveSharedPermissionCreateCmd = &cobra.Command{
	Use:   "create <driveId>",
	Short: "공유 드라이브 권한 생성",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).CreatePermission(args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedPermissionGetCmd = &cobra.Command{
	Use:   "get <driveId> <permissionId>",
	Short: "공유 드라이브 권한 상세 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewSharedDriveService(client).GetPermission(args[0], args[1])
		})
	},
}

var driveSharedPermissionUpdateCmd = &cobra.Command{
	Use:   "update <driveId> <permissionId>",
	Short: "공유 드라이브 권한 수정",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).PatchPermission(args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedPermissionDeleteCmd = &cobra.Command{
	Use:   "delete <driveId> <permissionId>",
	Short: "공유 드라이브 권한 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).DeletePermission(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedPermissionDeleteAllCmd = &cobra.Command{
	Use:   "delete-all <driveId>",
	Short: "공유 드라이브 권한 전체 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).DeleteAllPermissions(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedPermissionEnableCmd = &cobra.Command{
	Use:   "enable <driveId>",
	Short: "공유 드라이브 권한 활성화",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).EnablePermissions(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedPermissionDisableCmd = &cobra.Command{
	Use:   "disable <driveId>",
	Short: "공유 드라이브 권한 비활성화",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).DisablePermissions(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// File-level permissions

var driveSharedFilePermissionCmd = &cobra.Command{
	Use:   "file-permission",
	Short: "공유 드라이브 파일 권한 관리",
}

var driveSharedFilePermissionListCmd = &cobra.Command{
	Use:   "list <driveId> <fileId>",
	Short: "공유 드라이브 파일 권한 목록 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewSharedDriveService(client).ListFilePermissions(args[0], args[1])
		})
	},
}

var driveSharedFilePermissionCreateCmd = &cobra.Command{
	Use:   "create <driveId> <fileId>",
	Short: "공유 드라이브 파일 권한 생성",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).CreateFilePermission(args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedFilePermissionGetCmd = &cobra.Command{
	Use:   "get <driveId> <fileId> <permissionId>",
	Short: "공유 드라이브 파일 권한 상세 조회",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewSharedDriveService(client).GetFilePermission(args[0], args[1], args[2])
		})
	},
}

var driveSharedFilePermissionUpdateCmd = &cobra.Command{
	Use:   "update <driveId> <fileId> <permissionId>",
	Short: "공유 드라이브 파일 권한 수정",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).PatchFilePermission(args[0], args[1], args[2], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedFilePermissionDeleteCmd = &cobra.Command{
	Use:   "delete <driveId> <fileId> <permissionId>",
	Short: "공유 드라이브 파일 권한 삭제",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).DeleteFilePermission(args[0], args[1], args[2])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedFilePermissionDeleteAllCmd = &cobra.Command{
	Use:   "delete-all <driveId> <fileId>",
	Short: "공유 드라이브 파일 권한 전체 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).DeleteAllFilePermissions(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedFilePermissionEnableCmd = &cobra.Command{
	Use:   "enable <driveId> <fileId>",
	Short: "공유 드라이브 파일 권한 활성화",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).EnableFilePermissions(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSharedFilePermissionDisableCmd = &cobra.Command{
	Use:   "disable <driveId> <fileId>",
	Short: "공유 드라이브 파일 권한 비활성화",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewSharedDriveService(client).DisableFilePermissions(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
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
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
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
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewGroupFolderService(client).GetFile(args[0], args[1])
		})
	},
}

// ─── GroupFolder Task 5-4: 관리 + 파일 보완 ───

var driveGroupCreateFolderCmd = &cobra.Command{
	Use:   "create-folder <groupId>",
	Short: "그룹 폴더 생성",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewGroupFolderService(client).CreateFolder(args[0])
		})
	},
}

var driveGroupDeleteFolderCmd = &cobra.Command{
	Use:   "delete-folder <groupId>",
	Short: "그룹 폴더 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).DeleteFolder(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupMkdirCmd = &cobra.Command{
	Use:   "mkdir <groupId>",
	Short: "그룹 폴더 내 폴더 생성",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewGroupFolderService(client)
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		parent, _ := cmd.Flags().GetString("parent")
		var resp *api.Response
		if parent != "" {
			resp, err = svc.CreateSubFolder(args[0], parent, body)
		} else {
			resp, err = svc.CreateFolderInRoot(args[0], body)
		}
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupDeleteCmd = &cobra.Command{
	Use:   "delete <groupId> <fileId>",
	Short: "그룹 파일 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).DeleteFile(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupUploadCmd = &cobra.Command{
	Use:   "upload <groupId> [--folder <fileId>] --file <path>",
	Short: "그룹 파일 업로드 URL 생성 (--folder 미지정 시 루트 업로드)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewGroupFolderService(client)

		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			return fmt.Errorf("--file 플래그가 필요합니다")
		}

		folderID, _ := cmd.Flags().GetString("folder")
		resume, _ := cmd.Flags().GetBool("resume")

		fileName, fileSize, err := statFileForUpload(filePath)
		if err != nil {
			return err
		}

		uploadBody := map[string]interface{}{"fileName": fileName}
		if resume {
			uploadBody["resume"] = true
		}
		var resp *api.Response
		if folderID != "" {
			resp, err = svc.CreateUploadURL(args[0], folderID, uploadBody, fileSize)
		} else {
			resp, err = svc.CreateRootUploadURL(args[0], uploadBody, fileSize)
		}
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

var driveGroupDownloadCmd = &cobra.Command{
	Use:   "download <groupId> <fileId>",
	Short: "그룹 파일 다운로드 URL 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewGroupFolderService)
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

// ─── GroupFolder Task 5-5: 파일조작 ───

var driveGroupCopyCmd = &cobra.Command{
	Use:   "copy <groupId> <fileId>",
	Short: "그룹 파일 복사",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).CopyFile(args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupRenameCmd = &cobra.Command{
	Use:   "rename <groupId> <fileId>",
	Short: "그룹 파일 이름 변경",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).RenameFile(args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupMoveCmd = &cobra.Command{
	Use:   "move <groupId> <fileId>",
	Short: "그룹 파일 이동",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).MoveFile(args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupProtectCmd = &cobra.Command{
	Use:   "protect <groupId> <fileId>",
	Short: "그룹 파일 보호 설정",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).ProtectFile(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupUnprotectCmd = &cobra.Command{
	Use:   "unprotect <groupId> <fileId>",
	Short: "그룹 파일 보호 해제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).UnprotectFile(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupLockCmd = &cobra.Command{
	Use:   "lock <groupId> <fileId>",
	Short: "그룹 파일 잠금",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).LockFile(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupUnlockCmd = &cobra.Command{
	Use:   "unlock <groupId> <fileId>",
	Short: "그룹 파일 잠금 해제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).UnlockFile(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── GroupFolder Task 5-6: 리비전 + 휴지통 ───

var driveGroupRevisionCmd = &cobra.Command{
	Use:   "revision",
	Short: "그룹 파일 리비전 관리",
}

var driveGroupRevisionListCmd = &cobra.Command{
	Use:   "list <groupId> <fileId>",
	Short: "그룹 파일 리비전 목록 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewGroupFolderService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, nil, "revisions", func(c string, n int) (*api.Response, error) {
			return svc.ListRevisions(args[0], args[1], c, n)
		})
	},
}

var driveGroupRevisionGetCmd = &cobra.Command{
	Use:   "get <groupId> <fileId> <revisionId>",
	Short: "그룹 파일 리비전 상세 조회",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewGroupFolderService(client).GetRevision(args[0], args[1], args[2])
		})
	},
}

var driveGroupRevisionRestoreCmd = &cobra.Command{
	Use:   "restore <groupId> <fileId> <revisionId>",
	Short: "그룹 파일 리비전 복원",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).RestoreRevision(args[0], args[1], args[2])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupRevisionDownloadCmd = &cobra.Command{
	Use:   "download <groupId> <fileId> <revisionId>",
	Short: "그룹 파일 리비전 다운로드 URL 조회",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewGroupFolderService)
		if err != nil {
			return err
		}
		downloadURL, err := svc.GetRevisionDownloadURL(args[0], args[1], args[2])
		if err != nil {
			return err
		}
		printDownloadURL(downloadURL)
		return nil
	},
}

var driveGroupTrashListCmd = &cobra.Command{
	Use:   "trash-list <groupId>",
	Short: "그룹 휴지통 파일 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewGroupFolderService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, nil, "files", func(c string, n int) (*api.Response, error) {
			return svc.ListTrashFiles(args[0], c, n)
		})
	},
}

var driveGroupTrashRestoreCmd = &cobra.Command{
	Use:   "trash-restore <groupId> <fileId>",
	Short: "그룹 휴지통 파일 복원",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).RestoreTrashFile(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupTrashDeleteCmd = &cobra.Command{
	Use:   "trash-delete <groupId> <fileId>",
	Short: "그룹 휴지통 파일 영구 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).DeleteTrashFile(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── GroupFolder Task 5-7: 링크 + 권한 ───

var driveGroupLinkSettingCmd = &cobra.Command{
	Use:   "link-setting <groupId>",
	Short: "그룹 폴더 링크 설정 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewGroupFolderService(client).GetLinkSetting(args[0])
		})
	},
}

var driveGroupLinkCmd = &cobra.Command{
	Use:   "link",
	Short: "그룹 파일 링크 관리",
}

var driveGroupLinkGetCmd = &cobra.Command{
	Use:   "get <groupId> <fileId>",
	Short: "그룹 파일 링크 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewGroupFolderService(client).GetLink(args[0], args[1])
		})
	},
}

var driveGroupLinkCreateCmd = &cobra.Command{
	Use:   "create <groupId> <fileId>",
	Short: "그룹 파일 링크 생성",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).CreateLink(args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupLinkUpdateCmd = &cobra.Command{
	Use:   "update <groupId> <fileId>",
	Short: "그룹 파일 링크 수정",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).PatchLink(args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupLinkDeleteCmd = &cobra.Command{
	Use:   "delete <groupId> <fileId>",
	Short: "그룹 파일 링크 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).DeleteLink(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupPermissionCmd = &cobra.Command{
	Use:   "permission",
	Short: "그룹 파일 권한 관리",
}

var driveGroupPermissionListCmd = &cobra.Command{
	Use:   "list <groupId> <fileId>",
	Short: "그룹 파일 권한 목록 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewGroupFolderService(client).ListPermissions(args[0], args[1])
		})
	},
}

var driveGroupPermissionCreateCmd = &cobra.Command{
	Use:   "create <groupId> <fileId>",
	Short: "그룹 파일 권한 생성",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).CreatePermission(args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupPermissionGetCmd = &cobra.Command{
	Use:   "get <groupId> <fileId> <permissionId>",
	Short: "그룹 파일 권한 상세 조회",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewGroupFolderService(client).GetPermission(args[0], args[1], args[2])
		})
	},
}

var driveGroupPermissionUpdateCmd = &cobra.Command{
	Use:   "update <groupId> <fileId> <permissionId>",
	Short: "그룹 파일 권한 수정",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).PatchPermission(args[0], args[1], args[2], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupPermissionDeleteCmd = &cobra.Command{
	Use:   "delete <groupId> <fileId> <permissionId>",
	Short: "그룹 파일 권한 삭제",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).DeletePermission(args[0], args[1], args[2])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveGroupPermissionDeleteAllCmd = &cobra.Command{
	Use:   "delete-all <groupId> <fileId>",
	Short: "그룹 파일 권한 전체 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := api.NewGroupFolderService(client).DeleteAllPermissions(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
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

// ─── SharedFolder Task 5-12: 관리 + 파일 ───

var driveSFGetCmd = &cobra.Command{
	Use:   "get <sharedFolderId>",
	Short: "공유 폴더 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).GetFolder(userID, args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveSFLeaveCmd = &cobra.Command{
	Use:   "leave <sharedFolderId>",
	Short: "공유 폴더 나가기",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).LeaveFolder(userID, args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSFListMembersCmd = &cobra.Command{
	Use:   "list-members <sharedFolderId>",
	Short: "공유 폴더 멤버 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).ListMembers(userID, args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveSFListFilesCmd = &cobra.Command{
	Use:   "list-files <sharedFolderId>",
	Short: "공유 폴더 루트 파일 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewSharedFolderService(client)
		folder, _ := cmd.Flags().GetString("folder")
		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		var resp *api.Response
		if folder != "" {
			resp, err = svc.ListFolderChildren(userID, args[0], folder, cursor, count)
		} else {
			resp, err = svc.ListFiles(userID, args[0], cursor, count)
		}
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveSFGetFileCmd = &cobra.Command{
	Use:   "get-file <sharedFolderId> <fileId>",
	Short: "공유 폴더 파일 상세 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).GetFile(userID, args[0], args[1])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveSFMkdirCmd = &cobra.Command{
	Use:   "mkdir <sharedFolderId>",
	Short: "공유 폴더 내 폴더 생성",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewSharedFolderService(client)
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		parent, _ := cmd.Flags().GetString("parent")
		var resp *api.Response
		if parent != "" {
			resp, err = svc.CreateSubFolder(userID, args[0], parent, body)
		} else {
			resp, err = svc.CreateFolderInRoot(userID, args[0], body)
		}
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSFDeleteCmd = &cobra.Command{
	Use:   "delete <sharedFolderId> <fileId>",
	Short: "공유 폴더 파일 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).DeleteFile(userID, args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSFUploadCmd = &cobra.Command{
	Use:   "upload <sharedFolderId>",
	Short: "공유 폴더 파일 업로드",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewSharedFolderService(client)

		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			return fmt.Errorf("--file 플래그가 필요합니다")
		}

		fileName, fileSize, err := statFileForUpload(filePath)
		if err != nil {
			return err
		}

		uploadBody := map[string]interface{}{"fileName": fileName}
		folder, _ := cmd.Flags().GetString("folder")
		resume, _ := cmd.Flags().GetBool("resume")
		if resume {
			uploadBody["resume"] = true
		}

		var resp *api.Response
		if folder != "" {
			resp, err = svc.CreateUploadURL(userID, args[0], folder, uploadBody, fileSize)
		} else {
			resp, err = svc.CreateUploadURLInRoot(userID, args[0], uploadBody, fileSize)
		}
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

var driveSFDownloadCmd = &cobra.Command{
	Use:   "download <sharedFolderId> <fileId>",
	Short: "공유 폴더 파일 다운로드 URL 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		downloadURL, err := api.NewSharedFolderService(client).GetDownloadURL(userID, args[0], args[1])
		if err != nil {
			return err
		}
		printDownloadURL(downloadURL)
		return nil
	},
}

// ─── SharedFolder Task 5-13: 파일조작 + 리비전 ───

var driveSFCopyCmd = &cobra.Command{
	Use:   "copy <sharedFolderId> <fileId>",
	Short: "공유 폴더 파일 복사",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).CopyFile(userID, args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSFRenameCmd = &cobra.Command{
	Use:   "rename <sharedFolderId> <fileId>",
	Short: "공유 폴더 파일 이름 변경",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).RenameFile(userID, args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSFMoveCmd = &cobra.Command{
	Use:   "move <sharedFolderId> <fileId>",
	Short: "공유 폴더 파일 이동",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).MoveFile(userID, args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSFProtectCmd = &cobra.Command{
	Use:   "protect <sharedFolderId> <fileId>",
	Short: "공유 폴더 파일 보호 설정",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).ProtectFile(userID, args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSFUnprotectCmd = &cobra.Command{
	Use:   "unprotect <sharedFolderId> <fileId>",
	Short: "공유 폴더 파일 보호 해제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).UnprotectFile(userID, args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSFLockCmd = &cobra.Command{
	Use:   "lock <sharedFolderId> <fileId>",
	Short: "공유 폴더 파일 잠금",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).LockFile(userID, args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSFUnlockCmd = &cobra.Command{
	Use:   "unlock <sharedFolderId> <fileId>",
	Short: "공유 폴더 파일 잠금 해제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).UnlockFile(userID, args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSFRevisionCmd = &cobra.Command{
	Use:   "revision",
	Short: "공유 폴더 파일 리비전 관리",
}

var driveSFRevisionListCmd = &cobra.Command{
	Use:   "list <sharedFolderId> <fileId>",
	Short: "공유 폴더 파일 리비전 목록 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewSharedFolderService(client)
		return runListCmd(cmd, nil, "revisions", func(c string, n int) (*api.Response, error) {
			return svc.ListRevisions(userID, args[0], args[1], c, n)
		})
	},
}

var driveSFRevisionGetCmd = &cobra.Command{
	Use:   "get <sharedFolderId> <fileId> <revisionId>",
	Short: "공유 폴더 파일 리비전 상세 조회",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).GetRevision(userID, args[0], args[1], args[2])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveSFRevisionRestoreCmd = &cobra.Command{
	Use:   "restore <sharedFolderId> <fileId> <revisionId>",
	Short: "공유 폴더 파일 리비전 복원",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).RestoreRevision(userID, args[0], args[1], args[2])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSFRevisionDownloadCmd = &cobra.Command{
	Use:   "download <sharedFolderId> <fileId> <revisionId>",
	Short: "공유 폴더 파일 리비전 다운로드 URL 조회",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		downloadURL, err := api.NewSharedFolderService(client).GetRevisionDownloadURL(userID, args[0], args[1], args[2])
		if err != nil {
			return err
		}
		printDownloadURL(downloadURL)
		return nil
	},
}

// ─── SharedFolder Task 5-14: 링크 ───

var driveSFLinkSettingCmd = &cobra.Command{
	Use:   "link-setting <sharedFolderId>",
	Short: "공유 폴더 링크 설정 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).GetLinkSetting(userID, args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveSFLinkCmd = &cobra.Command{
	Use:   "link",
	Short: "공유 폴더 파일 링크 관리",
}

var driveSFLinkGetCmd = &cobra.Command{
	Use:   "get <sharedFolderId> <fileId>",
	Short: "공유 폴더 파일 링크 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).GetLink(userID, args[0], args[1])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var driveSFLinkCreateCmd = &cobra.Command{
	Use:   "create <sharedFolderId> <fileId>",
	Short: "공유 폴더 파일 링크 생성",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).CreateLink(userID, args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSFLinkUpdateCmd = &cobra.Command{
	Use:   "update <sharedFolderId> <fileId>",
	Short: "공유 폴더 파일 링크 수정",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).PatchLink(userID, args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var driveSFLinkDeleteCmd = &cobra.Command{
	Use:   "delete <sharedFolderId> <fileId>",
	Short: "공유 폴더 파일 링크 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewSharedFolderService(client).DeleteLink(userID, args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
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

	// --json flag for MyDrive commands that need JSON body
	for _, c := range []*cobra.Command{
		driveCopyCmd, driveRenameCmd, driveMoveCmd,
		driveLinkCreateCmd, driveLinkUpdateCmd,
		driveShareCreateCmd, driveShareUpdateCmd,
	} {
		c.Flags().String("json", "", "JSON 요청 본문 (- 이면 stdin)")
	}

	// --json flag for SharedDrive commands
	for _, c := range []*cobra.Command{
		driveSharedCreateDriveCmd, driveSharedUpdateDriveCmd, driveSharedMkdirCmd,
		driveSharedCopyCmd, driveSharedRenameCmd, driveSharedMoveCmd,
		driveSharedLinkCreateCmd, driveSharedLinkUpdateCmd,
		driveSharedPermissionCreateCmd, driveSharedPermissionUpdateCmd,
		driveSharedFilePermissionCreateCmd, driveSharedFilePermissionUpdateCmd,
	} {
		c.Flags().String("json", "", "JSON 요청 본문 (- 이면 stdin)")
	}

	// --json flag for GroupFolder commands
	for _, c := range []*cobra.Command{
		driveGroupMkdirCmd,
		driveGroupCopyCmd, driveGroupRenameCmd, driveGroupMoveCmd,
		driveGroupLinkCreateCmd, driveGroupLinkUpdateCmd,
		driveGroupPermissionCreateCmd, driveGroupPermissionUpdateCmd,
	} {
		c.Flags().String("json", "", "JSON 요청 본문 (- 이면 stdin)")
	}

	// Pagination flags for list commands that use runListCmd (cursor + count + all)
	addListFlags(
		driveSharedListDrivesCmd,
		driveTrashListCmd,
		driveRevisionListCmd, driveShareListSubFoldersCmd,
		driveGroupRevisionListCmd, driveGroupTrashListCmd,
		driveSharedRevisionListCmd, driveSharedTrashListCmd,
		driveSFRevisionListCmd,
	)

	// Pagination flags for list commands with manual branching (cursor + count only)
	for _, c := range []*cobra.Command{
		driveListCmd,
		driveSharedListCmd,
		driveGroupListCmd,
		driveSharedFolderListCmd, driveSharedFolderFilesCmd,
		driveSFListFilesCmd,
	} {
		c.Flags().String("cursor", "", "페이지네이션 커서")
		c.Flags().Int("count", 0, "페이지 크기")
	}

	// Folder listing flag for file-browsing list commands
	for _, c := range []*cobra.Command{driveListCmd, driveSharedListCmd, driveGroupListCmd} {
		c.Flags().String("folder", "", "폴더 ID (하위 파일 조회)")
	}

	// SharedFolder also needs --user-id (existing + new)
	for _, c := range []*cobra.Command{
		driveSharedFolderListCmd, driveSharedFolderFilesCmd,
		// Task 5-12
		driveSFGetCmd, driveSFLeaveCmd, driveSFListMembersCmd, driveSFListFilesCmd,
		driveSFGetFileCmd, driveSFMkdirCmd, driveSFDeleteCmd,
		driveSFUploadCmd, driveSFDownloadCmd,
		// Task 5-13
		driveSFCopyCmd, driveSFRenameCmd, driveSFMoveCmd,
		driveSFProtectCmd, driveSFUnprotectCmd, driveSFLockCmd, driveSFUnlockCmd,
		driveSFRevisionListCmd, driveSFRevisionGetCmd, driveSFRevisionRestoreCmd, driveSFRevisionDownloadCmd,
		// Task 5-14
		driveSFLinkSettingCmd,
		driveSFLinkGetCmd, driveSFLinkCreateCmd, driveSFLinkUpdateCmd, driveSFLinkDeleteCmd,
	} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}

	// --json flag for SharedFolder commands
	for _, c := range []*cobra.Command{
		driveSFMkdirCmd,
		driveSFCopyCmd, driveSFRenameCmd, driveSFMoveCmd,
		driveSFLinkCreateCmd, driveSFLinkUpdateCmd,
	} {
		c.Flags().String("json", "", "JSON 요청 본문 (- 이면 stdin)")
	}

	// SharedFolder mkdir/upload flags
	driveSFMkdirCmd.Flags().String("parent", "", "상위 폴더 ID")
	driveSFUploadCmd.Flags().String("file", "", "업로드할 파일 경로 (필수)")
	driveSFUploadCmd.Flags().String("folder", "", "업로드 대상 폴더 ID")
	driveSFListFilesCmd.Flags().String("folder", "", "폴더 ID (하위 파일 조회)")

	driveUploadCmd.Flags().String("folder", "", "업로드 대상 폴더 ID")
	driveMkdirCmd.Flags().String("name", "", "폴더 이름 (필수)")
	driveMkdirCmd.Flags().String("parent", "", "상위 폴더 ID")

	// SharedDrive mkdir/upload flags
	driveSharedMkdirCmd.Flags().String("parent", "", "상위 폴더 ID")
	driveSharedUploadCmd.Flags().String("file", "", "업로드할 파일 경로 (필수)")
	driveSharedUploadCmd.Flags().String("folder", "", "업로드 대상 폴더 ID")

	// GroupFolder mkdir/upload flags
	driveGroupMkdirCmd.Flags().String("parent", "", "상위 폴더 ID")
	driveGroupUploadCmd.Flags().String("file", "", "업로드할 파일 경로 (필수)")
	driveGroupUploadCmd.Flags().String("folder", "", "업로드 대상 폴더 ID (미지정 시 루트)")

	// --resume flag for all upload commands (Core v4.2 #4).
	// NOTE: requests a resumable upload session from the server by sending
	// "resume": true in the body. The offset-based re-transmission path
	// (Content-Range PUT from the server-returned offset) is not yet
	// implemented in the CLI; this flag only forwards the v4.2 body field.
	for _, c := range []*cobra.Command{driveUploadCmd, driveSharedUploadCmd, driveGroupUploadCmd, driveSFUploadCmd} {
		c.Flags().Bool("resume", false, "서버 resumable 업로드 세션 요청 (실제 재전송은 미구현)")
	}

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

	// Task 5-8: SharedDrive 관리 + 파일 보완
	driveSharedCmd.AddCommand(driveSharedCreateDriveCmd, driveSharedUpdateDriveCmd, driveSharedDeleteDriveCmd,
		driveSharedMkdirCmd, driveSharedDeleteCmd)

	// Task 5-9: SharedDrive 파일조작
	driveSharedCmd.AddCommand(driveSharedCopyCmd, driveSharedRenameCmd, driveSharedMoveCmd,
		driveSharedProtectCmd, driveSharedUnprotectCmd, driveSharedLockCmd, driveSharedUnlockCmd)

	// Task 5-10: SharedDrive 리비전 + 휴지통
	driveSharedRevisionCmd.AddCommand(driveSharedRevisionListCmd, driveSharedRevisionGetCmd,
		driveSharedRevisionRestoreCmd, driveSharedRevisionDownloadCmd)
	driveSharedCmd.AddCommand(driveSharedRevisionCmd)
	driveSharedCmd.AddCommand(driveSharedTrashListCmd, driveSharedTrashRestoreCmd, driveSharedTrashDeleteCmd)

	// Task 5-11: SharedDrive 링크 + 권한
	driveSharedCmd.AddCommand(driveSharedLinkSettingCmd)
	driveSharedLinkCmd.AddCommand(driveSharedLinkGetCmd, driveSharedLinkCreateCmd,
		driveSharedLinkUpdateCmd, driveSharedLinkDeleteCmd)
	driveSharedCmd.AddCommand(driveSharedLinkCmd)

	driveSharedPermissionCmd.AddCommand(driveSharedPermissionListCmd, driveSharedPermissionCreateCmd,
		driveSharedPermissionGetCmd, driveSharedPermissionUpdateCmd,
		driveSharedPermissionDeleteCmd, driveSharedPermissionDeleteAllCmd,
		driveSharedPermissionEnableCmd, driveSharedPermissionDisableCmd)
	driveSharedCmd.AddCommand(driveSharedPermissionCmd)

	driveSharedFilePermissionCmd.AddCommand(driveSharedFilePermissionListCmd, driveSharedFilePermissionCreateCmd,
		driveSharedFilePermissionGetCmd, driveSharedFilePermissionUpdateCmd,
		driveSharedFilePermissionDeleteCmd, driveSharedFilePermissionDeleteAllCmd,
		driveSharedFilePermissionEnableCmd, driveSharedFilePermissionDisableCmd)
	driveSharedCmd.AddCommand(driveSharedFilePermissionCmd)

	driveCmd.AddCommand(driveSharedCmd)

	// GroupFolder: existing + Task 5-4 ~ 5-7
	driveGroupCmd.AddCommand(driveGroupGetFolderCmd, driveGroupListCmd, driveGroupGetCmd)
	driveGroupCmd.AddCommand(driveGroupCreateFolderCmd, driveGroupDeleteFolderCmd,
		driveGroupMkdirCmd, driveGroupDeleteCmd, driveGroupUploadCmd, driveGroupDownloadCmd)
	driveGroupCmd.AddCommand(driveGroupCopyCmd, driveGroupRenameCmd, driveGroupMoveCmd,
		driveGroupProtectCmd, driveGroupUnprotectCmd, driveGroupLockCmd, driveGroupUnlockCmd)

	driveGroupRevisionCmd.AddCommand(driveGroupRevisionListCmd, driveGroupRevisionGetCmd,
		driveGroupRevisionRestoreCmd, driveGroupRevisionDownloadCmd)
	driveGroupCmd.AddCommand(driveGroupRevisionCmd)
	driveGroupCmd.AddCommand(driveGroupTrashListCmd, driveGroupTrashRestoreCmd, driveGroupTrashDeleteCmd)

	driveGroupCmd.AddCommand(driveGroupLinkSettingCmd)
	driveGroupLinkCmd.AddCommand(driveGroupLinkGetCmd, driveGroupLinkCreateCmd,
		driveGroupLinkUpdateCmd, driveGroupLinkDeleteCmd)
	driveGroupCmd.AddCommand(driveGroupLinkCmd)

	driveGroupPermissionCmd.AddCommand(driveGroupPermissionListCmd, driveGroupPermissionCreateCmd,
		driveGroupPermissionGetCmd, driveGroupPermissionUpdateCmd,
		driveGroupPermissionDeleteCmd, driveGroupPermissionDeleteAllCmd)
	driveGroupCmd.AddCommand(driveGroupPermissionCmd)

	driveCmd.AddCommand(driveGroupCmd)

	driveSharedFolderCmd.AddCommand(driveSharedFolderListCmd, driveSharedFolderFilesCmd)

	// Task 5-12: SharedFolder 관리 + 파일
	driveSharedFolderCmd.AddCommand(driveSFGetCmd, driveSFLeaveCmd, driveSFListMembersCmd,
		driveSFListFilesCmd, driveSFGetFileCmd, driveSFMkdirCmd, driveSFDeleteCmd,
		driveSFUploadCmd, driveSFDownloadCmd)

	// Task 5-13: SharedFolder 파일조작 + 리비전
	driveSharedFolderCmd.AddCommand(driveSFCopyCmd, driveSFRenameCmd, driveSFMoveCmd,
		driveSFProtectCmd, driveSFUnprotectCmd, driveSFLockCmd, driveSFUnlockCmd)
	driveSFRevisionCmd.AddCommand(driveSFRevisionListCmd, driveSFRevisionGetCmd,
		driveSFRevisionRestoreCmd, driveSFRevisionDownloadCmd)
	driveSharedFolderCmd.AddCommand(driveSFRevisionCmd)

	// Task 5-14: SharedFolder 링크
	driveSharedFolderCmd.AddCommand(driveSFLinkSettingCmd)
	driveSFLinkCmd.AddCommand(driveSFLinkGetCmd, driveSFLinkCreateCmd,
		driveSFLinkUpdateCmd, driveSFLinkDeleteCmd)
	driveSharedFolderCmd.AddCommand(driveSFLinkCmd)

	driveCmd.AddCommand(driveSharedFolderCmd)

	rootCmd.AddCommand(driveCmd)
}
