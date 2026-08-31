package cmd

import (
	"fmt"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/spf13/cobra"
)

const (
	channelFolderReadHelp  = "인증: 구성원 계정 Access Token 전용 (서비스 계정 사용 불가)\n최소 scope: file.read, group.folder.read"
	channelFolderWriteHelp = "인증: 구성원 계정 Access Token 전용 (서비스 계정 사용 불가)\n최소 scope: file, group.folder"
)

type channelFolderServiceRun func(*api.ChannelFolderService, *api.Client) error

func withChannelFolderService(run channelFolderServiceRun) error {
	client, _, token, err := newAPIClient()
	if err != nil {
		return err
	}
	if token.AuthMethod == auth.AuthMethodJWT {
		return fmt.Errorf("채널 폴더 API는 구성원 계정 Access Token만 지원합니다")
	}
	return run(api.NewChannelFolderService(client), client)
}

func newChannelFolderBodyCommand(use, short string, argCount int, call func(*api.ChannelFolderService, []string) (*api.Response, error)) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Long:  short + "\n\n" + channelFolderReadHelp,
		Args:  cobra.ExactArgs(argCount),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withChannelFolderService(func(svc *api.ChannelFolderService, _ *api.Client) error {
				resp, err := call(svc, args)
				if err != nil {
					return err
				}
				printBody(resp.Body)
				return nil
			})
		},
	}
}

func newChannelFolderWriteCommand(use, short string, argCount int, call func(*api.ChannelFolderService, []string) (*api.Response, error)) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Long:  short + "\n\n" + channelFolderWriteHelp,
		Args:  cobra.ExactArgs(argCount),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withChannelFolderService(func(svc *api.ChannelFolderService, _ *api.Client) error {
				resp, err := call(svc, args)
				if err != nil {
					return err
				}
				printResponse(resp)
				return nil
			})
		},
	}
}

func newChannelFolderJSONCommand(use, short string, argCount int, call func(*api.ChannelFolderService, []string, []byte) (*api.Response, error)) *cobra.Command {
	command := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  short + "\n\n" + channelFolderWriteHelp,
		Args:  cobra.ExactArgs(argCount),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONFlagRaw(cmd)
			if err != nil {
				return err
			}
			return withChannelFolderService(func(svc *api.ChannelFolderService, _ *api.Client) error {
				resp, err := call(svc, args, body)
				if err != nil {
					return err
				}
				printResponse(resp)
				return nil
			})
		},
	}
	command.Flags().String("json", "", "JSON 요청 본문 (- 이면 stdin)")
	return command
}

func newChannelFolderDownloadCommand(use, short string, argCount int, call func(*api.ChannelFolderService, []string) (string, error)) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Long:  short + "\n\n" + channelFolderReadHelp,
		Args:  cobra.ExactArgs(argCount),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withChannelFolderService(func(svc *api.ChannelFolderService, _ *api.Client) error {
				downloadURL, err := call(svc, args)
				if err != nil {
					return err
				}
				printDownloadURL(downloadURL)
				return nil
			})
		},
	}
}

func newDriveChannelCommand() *cobra.Command {
	channelCmd := &cobra.Command{
		Use:   "channel",
		Short: "채널 폴더 관리",
		Long:  "채널 폴더 관리\n\n" + channelFolderReadHelp + "\n쓰기 scope: file, group.folder",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "채널 폴더 목록 조회",
		Long:  "채널 폴더 목록 조회\n\n" + channelFolderReadHelp,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return withChannelFolderService(func(svc *api.ChannelFolderService, _ *api.Client) error {
				return runListCmd(cmd, []string{"channelFolderId", "name", "createdTime"}, "channelFolders", func(_ string, _ int) (*api.Response, error) {
					return svc.ListChannelFolders()
				})
			})
		},
	}
	getCmd := newChannelFolderBodyCommand("get <channelFolderId>", "채널 폴더 속성 조회", 1, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.GetChannelFolder(args[0])
	})
	filesCmd := &cobra.Command{
		Use:   "files <channelFolderId>",
		Short: "채널 폴더 파일 목록 조회",
		Long:  "채널 폴더 파일 목록 조회\n\n" + channelFolderReadHelp,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			folder, _ := cmd.Flags().GetString("folder")
			return withChannelFolderService(func(svc *api.ChannelFolderService, _ *api.Client) error {
				return runListCmd(cmd, []string{"fileId", "fileName", "fileType", "modifiedTime"}, "files", func(cursor string, count int) (*api.Response, error) {
					if folder != "" {
						return svc.ListFolderChildren(args[0], folder, cursor, count)
					}
					return svc.ListFiles(args[0], cursor, count)
				})
			})
		},
	}
	getFileCmd := newChannelFolderBodyCommand("get-file <channelFolderId> <fileId>", "채널 폴더 파일 속성 조회", 2, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.GetFile(args[0], args[1])
	})
	uploadCmd := &cobra.Command{
		Use:   "upload <channelFolderId>",
		Short: "채널 폴더 파일 업로드",
		Long:  "채널 폴더 파일 업로드\n\n" + channelFolderWriteHelp,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			localPath, _ := cmd.Flags().GetString("file")
			if localPath == "" {
				return fmt.Errorf("--file 플래그가 필요합니다")
			}
			fileName, fileSize, err := statFileForUpload(localPath)
			if err != nil {
				return err
			}
			folder, _ := cmd.Flags().GetString("folder")
			resume, _ := cmd.Flags().GetBool("resume")
			body := map[string]interface{}{"fileName": fileName}
			if resume {
				body["resume"] = true
			}
			return withChannelFolderService(func(svc *api.ChannelFolderService, client *api.Client) error {
				var resp *api.Response
				var err error
				if folder == "" {
					resp, err = svc.CreateRootUploadURL(args[0], body, fileSize)
				} else {
					resp, err = svc.CreateUploadURL(args[0], folder, body, fileSize)
				}
				if err != nil {
					return err
				}
				responseBody, err := doUploadFromResponse(client, resp.Body, localPath)
				if err != nil {
					return err
				}
				printBody(responseBody)
				return nil
			})
		},
	}
	downloadCmd := newChannelFolderDownloadCommand("download <channelFolderId> <fileId>", "채널 폴더 파일 다운로드 URL 조회", 2, func(svc *api.ChannelFolderService, args []string) (string, error) {
		return svc.GetDownloadURL(args[0], args[1])
	})
	mkdirCmd := &cobra.Command{
		Use:   "mkdir <channelFolderId>",
		Short: "채널 폴더에 폴더 생성",
		Long:  "채널 폴더에 폴더 생성\n\n" + channelFolderWriteHelp,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readJSONFlagRaw(cmd)
			if err != nil {
				return err
			}
			parent, _ := cmd.Flags().GetString("parent")
			return withChannelFolderService(func(svc *api.ChannelFolderService, _ *api.Client) error {
				var resp *api.Response
				var err error
				if parent == "" {
					resp, err = svc.CreateFolderInRoot(args[0], body)
				} else {
					resp, err = svc.CreateSubFolder(args[0], parent, body)
				}
				if err != nil {
					return err
				}
				printResponse(resp)
				return nil
			})
		},
	}
	deleteCmd := newChannelFolderWriteCommand("delete <channelFolderId> <fileId>", "채널 폴더 파일 삭제", 2, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.DeleteFile(args[0], args[1])
	})
	copyCmd := newChannelFolderJSONCommand("copy <channelFolderId> <fileId>", "채널 폴더 파일 복사", 2, func(svc *api.ChannelFolderService, args []string, body []byte) (*api.Response, error) {
		return svc.CopyFile(args[0], args[1], body)
	})
	renameCmd := newChannelFolderJSONCommand("rename <channelFolderId> <fileId>", "채널 폴더 파일 이름 변경", 2, func(svc *api.ChannelFolderService, args []string, body []byte) (*api.Response, error) {
		return svc.RenameFile(args[0], args[1], body)
	})
	moveCmd := newChannelFolderJSONCommand("move <channelFolderId> <fileId>", "채널 폴더 파일 이동", 2, func(svc *api.ChannelFolderService, args []string, body []byte) (*api.Response, error) {
		return svc.MoveFile(args[0], args[1], body)
	})
	protectCmd := newChannelFolderWriteCommand("protect <channelFolderId> <fileId>", "채널 폴더 파일 중요 표시", 2, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.ProtectFile(args[0], args[1])
	})
	unprotectCmd := newChannelFolderWriteCommand("unprotect <channelFolderId> <fileId>", "채널 폴더 파일 중요 표시 해제", 2, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.UnprotectFile(args[0], args[1])
	})
	lockCmd := newChannelFolderWriteCommand("lock <channelFolderId> <fileId>", "채널 폴더 파일 잠금", 2, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.LockFile(args[0], args[1])
	})
	unlockCmd := newChannelFolderWriteCommand("unlock <channelFolderId> <fileId>", "채널 폴더 파일 잠금 해제", 2, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.UnlockFile(args[0], args[1])
	})

	revisionCmd := &cobra.Command{Use: "revision", Short: "채널 폴더 파일 버전 관리"}
	revisionListCmd := &cobra.Command{
		Use:   "list <channelFolderId> <fileId>",
		Short: "채널 폴더 파일 버전 목록 조회",
		Long:  "채널 폴더 파일 버전 목록 조회\n\n" + channelFolderReadHelp,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withChannelFolderService(func(svc *api.ChannelFolderService, _ *api.Client) error {
				return runListCmd(cmd, []string{"revisionId", "fileSize", "modifiedTime", "updateUser"}, "revisions", func(cursor string, count int) (*api.Response, error) {
					return svc.ListRevisions(args[0], args[1], cursor, count)
				})
			})
		},
	}
	revisionGetCmd := newChannelFolderBodyCommand("get <channelFolderId> <fileId> <revisionId>", "채널 폴더 파일 버전 속성 조회", 3, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.GetRevision(args[0], args[1], args[2])
	})
	revisionRestoreCmd := newChannelFolderWriteCommand("restore <channelFolderId> <fileId> <revisionId>", "채널 폴더 파일 버전 복원", 3, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.RestoreRevision(args[0], args[1], args[2])
	})
	revisionDownloadCmd := newChannelFolderDownloadCommand("download <channelFolderId> <fileId> <revisionId>", "채널 폴더 파일 버전 다운로드 URL 조회", 3, func(svc *api.ChannelFolderService, args []string) (string, error) {
		return svc.GetRevisionDownloadURL(args[0], args[1], args[2])
	})
	revisionCmd.AddCommand(revisionListCmd, revisionGetCmd, revisionRestoreCmd, revisionDownloadCmd)

	linkSettingCmd := newChannelFolderBodyCommand("link-setting <channelFolderId>", "채널 폴더 링크 설정 조회", 1, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.GetLinkSetting(args[0])
	})
	linkCmd := &cobra.Command{Use: "link", Short: "채널 폴더 파일 링크 관리"}
	linkGetCmd := newChannelFolderBodyCommand("get <channelFolderId> <fileId>", "채널 폴더 파일 링크 조회", 2, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.GetLink(args[0], args[1])
	})
	linkCreateCmd := newChannelFolderJSONCommand("create <channelFolderId> <fileId>", "채널 폴더 파일 링크 생성", 2, func(svc *api.ChannelFolderService, args []string, body []byte) (*api.Response, error) {
		return svc.CreateLink(args[0], args[1], body)
	})
	linkUpdateCmd := newChannelFolderJSONCommand("update <channelFolderId> <fileId>", "채널 폴더 파일 링크 수정", 2, func(svc *api.ChannelFolderService, args []string, body []byte) (*api.Response, error) {
		return svc.PatchLink(args[0], args[1], body)
	})
	linkDeleteCmd := newChannelFolderWriteCommand("delete <channelFolderId> <fileId>", "채널 폴더 파일 링크 삭제", 2, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.DeleteLink(args[0], args[1])
	})
	linkCmd.AddCommand(linkGetCmd, linkCreateCmd, linkUpdateCmd, linkDeleteCmd)

	trashListCmd := &cobra.Command{
		Use:   "trash-list <channelFolderId>",
		Short: "채널 폴더 휴지통 목록 조회",
		Long:  "채널 폴더 휴지통 목록 조회\n\n" + channelFolderReadHelp,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withChannelFolderService(func(svc *api.ChannelFolderService, _ *api.Client) error {
				return runListCmd(cmd, []string{"trashFileId", "fileName", "deletedTime", "deleteUser"}, "trashFiles", func(cursor string, count int) (*api.Response, error) {
					return svc.ListTrashFiles(args[0], cursor, count)
				})
			})
		},
	}
	trashRestoreCmd := newChannelFolderWriteCommand("trash-restore <channelFolderId> <trashFileId>", "채널 폴더 휴지통 파일 복원", 2, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.RestoreTrashFile(args[0], args[1])
	})
	trashDeleteCmd := newChannelFolderWriteCommand("trash-delete <channelFolderId> <trashFileId>", "채널 폴더 휴지통 파일 영구 삭제", 2, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.DeleteTrashFile(args[0], args[1])
	})

	permissionCmd := &cobra.Command{Use: "permission", Short: "채널 폴더 접근 권한 관리"}
	permissionListCmd := &cobra.Command{
		Use:   "list <channelFolderId> <fileId>",
		Short: "채널 폴더 접근 권한 목록 조회",
		Long:  "채널 폴더 접근 권한 목록 조회\n\n" + channelFolderReadHelp,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withChannelFolderService(func(svc *api.ChannelFolderService, _ *api.Client) error {
				return runListCmd(cmd, []string{"permissionId", "userId", "userType", "type"}, "permissions", func(_ string, _ int) (*api.Response, error) {
					return svc.ListPermissions(args[0], args[1])
				})
			})
		},
	}
	permissionCreateCmd := newChannelFolderJSONCommand("create <channelFolderId> <fileId>", "채널 폴더 접근 권한 생성", 2, func(svc *api.ChannelFolderService, args []string, body []byte) (*api.Response, error) {
		return svc.CreatePermission(args[0], args[1], body)
	})
	permissionGetCmd := newChannelFolderBodyCommand("get <channelFolderId> <fileId> <permissionId>", "채널 폴더 접근 권한 조회", 3, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.GetPermission(args[0], args[1], args[2])
	})
	permissionUpdateCmd := newChannelFolderJSONCommand("update <channelFolderId> <fileId> <permissionId>", "채널 폴더 접근 권한 수정", 3, func(svc *api.ChannelFolderService, args []string, body []byte) (*api.Response, error) {
		return svc.PatchPermission(args[0], args[1], args[2], body)
	})
	permissionDeleteCmd := newChannelFolderWriteCommand("delete <channelFolderId> <fileId> <permissionId>", "채널 폴더 접근 권한 해제", 3, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.DeletePermission(args[0], args[1], args[2])
	})
	permissionDeleteAllCmd := newChannelFolderWriteCommand("delete-all <channelFolderId> <fileId>", "채널 폴더 접근 권한 전체 해제", 2, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.DeleteAllPermissions(args[0], args[1])
	})
	permissionEnableCmd := newChannelFolderWriteCommand("enable <channelFolderId> <fileId>", "채널 폴더 접근 권한 허용", 2, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.EnablePermissions(args[0], args[1])
	})
	permissionDisableCmd := newChannelFolderWriteCommand("disable <channelFolderId> <fileId>", "채널 폴더 접근 권한 미허용", 2, func(svc *api.ChannelFolderService, args []string) (*api.Response, error) {
		return svc.DisablePermissions(args[0], args[1])
	})
	permissionCmd.AddCommand(permissionListCmd, permissionCreateCmd, permissionGetCmd, permissionUpdateCmd,
		permissionDeleteCmd, permissionDeleteAllCmd, permissionEnableCmd, permissionDisableCmd)

	addListFlags(filesCmd, revisionListCmd, trashListCmd)
	filesCmd.Flags().String("folder", "", "폴더 ID (하위 파일 조회)")
	uploadCmd.Flags().String("file", "", "업로드할 파일 경로 (필수)")
	uploadCmd.Flags().String("folder", "", "업로드 대상 폴더 ID")
	uploadCmd.Flags().Bool("resume", false, "서버 resumable 업로드 세션 요청 및 이어올리기")
	mkdirCmd.Flags().String("json", "", "JSON 요청 본문 (- 이면 stdin)")
	mkdirCmd.Flags().String("parent", "", "상위 폴더 ID")

	channelCmd.AddCommand(listCmd, getCmd, filesCmd, getFileCmd, uploadCmd, downloadCmd, mkdirCmd, deleteCmd,
		copyCmd, renameCmd, moveCmd, protectCmd, unprotectCmd, lockCmd, unlockCmd,
		revisionCmd, linkSettingCmd, linkCmd, trashListCmd, trashRestoreCmd, trashDeleteCmd, permissionCmd)
	return channelCmd
}

var driveChannelCmd = newDriveChannelCommand()

func init() {
	driveCmd.AddCommand(driveChannelCmd)
}
