package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

var directoryCmd = &cobra.Command{
	Use:   "directory",
	Short: "디렉토리 관리 (사용자, 그룹, 조직)",
}

// ─── Existing Read Commands ───

var dirListUsersCmd = &cobra.Command{
	Use:   "list-users",
	Short: "사용자 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"userId", "userName", "email"}, "users", dir.ListUsers)
	},
}

var dirGetUserCmd = &cobra.Command{
	Use:   "get-user <userId>",
	Short: "사용자 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewDirectoryService(client).GetUser(args[0])
		})
	},
}

var dirListGroupsCmd = &cobra.Command{
	Use:   "list-groups",
	Short: "그룹 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"groupId", "groupName"}, "groups", dir.ListGroups)
	},
}

var dirGetGroupCmd = &cobra.Command{
	Use:   "get-group <groupId>",
	Short: "그룹 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewDirectoryService(client).GetGroup(args[0])
		})
	},
}

var dirListOrgUnitsCmd = &cobra.Command{
	Use:   "list-orgunits",
	Short: "조직 단위 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"orgUnitId", "orgUnitName"}, "orgUnits", dir.ListOrgUnits)
	},
}

var dirGetOrgUnitCmd = &cobra.Command{
	Use:   "get-orgunit <orgUnitId>",
	Short: "조직 단위 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewDirectoryService(client).GetOrgUnit(args[0])
		})
	},
}

var dirListLevelsCmd = &cobra.Command{
	Use:   "list-levels",
	Short: "직급 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"levelId", "levelName"}, "levels", dir.ListLevels)
	},
}

var dirListPositionsCmd = &cobra.Command{
	Use:   "list-positions",
	Short: "직책 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"positionId", "positionName"}, "positions", dir.ListPositions)
	},
}

var dirListUserTypesCmd = &cobra.Command{
	Use:   "list-user-types",
	Short: "사용자 유형 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"userTypeId", "userTypeName"}, "userTypes", dir.ListUserTypes)
	},
}

var dirListEmploymentTypesCmd = &cobra.Command{
	Use:   "list-employment-types",
	Short: "고용 유형 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"employmentTypeId", "employmentTypeName"}, "employmentTypes", dir.ListEmploymentTypes)
	},
}

// ─── Task 4-1: User CUD ───

var dirCreateUserCmd = &cobra.Command{
	Use:   "create-user",
	Short: "사용자 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateUser(body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirUpdateUserCmd = &cobra.Command{
	Use:   "update-user <userId>",
	Short: "사용자 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.UpdateUser(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirPatchUserCmd = &cobra.Command{
	Use:   "patch-user <userId>",
	Short: "사용자 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchUser(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirDeleteUserCmd = &cobra.Command{
	Use:   "delete-user <userId>",
	Short: "사용자 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.DeleteUser(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirForceDeleteUserCmd = &cobra.Command{
	Use:   "force-delete-user <userId>",
	Short: "사용자 강제 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.ForceDeleteUser(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirUndeleteUserCmd = &cobra.Command{
	Use:   "undelete-user <userId>",
	Short: "사용자 삭제 취소",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.UndeleteUser(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirSuspendUserCmd = &cobra.Command{
	Use:   "suspend-user <userId>",
	Short: "사용자 정지",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.SuspendUser(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirUnsuspendUserCmd = &cobra.Command{
	Use:   "unsuspend-user <userId>",
	Short: "사용자 정지 해제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.UnsuspendUser(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirForceLogoutUserCmd = &cobra.Command{
	Use:   "force-logout-user <userId>",
	Short: "사용자 강제 로그아웃",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.ForceLogoutUser(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirMoveUserCmd = &cobra.Command{
	Use:   "move-user <userId>",
	Short: "사용자 이동",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.MoveUser(args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirSetLeaveCmd = &cobra.Command{
	Use:   "set-leave <userId>",
	Short: "사용자 휴직 설정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.SetLeaveOfAbsence(args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirClearLeaveCmd = &cobra.Command{
	Use:   "clear-leave <userId>",
	Short: "사용자 휴직 해제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.ClearLeaveOfAbsence(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── Task 4-2: User Profile ───

var dirUploadPhotoCmd = &cobra.Command{
	Use:   "upload-photo <userId>",
	Short: "사용자 사진 업로드",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewDirectoryService(client)

		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			return fmt.Errorf("--file 플래그가 필요합니다")
		}
		stat, statErr := os.Stat(filePath)
		if statErr != nil {
			return fmt.Errorf("파일 정보 조회 실패: %w", statErr)
		}
		fileName := filepath.Base(filePath)
		fileSize := stat.Size()

		uploadBody := map[string]interface{}{
			"fileName": fileName,
			"fileSize": fileSize,
		}
		resp, err := svc.CreateUserPhoto(args[0], uploadBody)
		if err != nil {
			return err
		}

		var result struct {
			UploadURL string `json:"uploadUrl"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			return fmt.Errorf("업로드 URL 파싱 실패: %w", err)
		}
		if result.UploadURL == "" {
			return fmt.Errorf("업로드 URL을 받지 못했습니다")
		}
		if err := client.UploadFile(result.UploadURL, filePath); err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirGetPhotoCmd = &cobra.Command{
	Use:   "get-photo <userId>",
	Short: "사용자 사진 다운로드 URL 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		downloadURL, err := svc.GetUserPhoto(args[0])
		if err != nil {
			return err
		}
		printDownloadURL(downloadURL)
		return nil
	},
}

var dirDeletePhotoCmd = &cobra.Command{
	Use:   "delete-photo <userId>",
	Short: "사용자 사진 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.DeleteUserPhoto(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── Profile Status Subcommand Group ───

var dirProfileStatusCmd = &cobra.Command{
	Use:   "profile-status",
	Short: "사용자 프로필 상태 관리",
}

var dirProfileStatusListCmd = &cobra.Command{
	Use:   "list <userId>",
	Short: "프로필 상태 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, nil, "userProfileStatuses", func(cursor string, count int) (*api.Response, error) {
			return svc.ListProfileStatuses(args[0], cursor, count)
		})
	},
}

var dirProfileStatusGetCmd = &cobra.Command{
	Use:   "get <userId> <id>",
	Short: "프로필 상태 상세 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewDirectoryService(client).GetProfileStatus(args[0], args[1])
		})
	},
}

var dirProfileStatusCreateCmd = &cobra.Command{
	Use:   "create <userId>",
	Short: "프로필 상태 생성",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateProfileStatus(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirProfileStatusUpdateCmd = &cobra.Command{
	Use:   "update <userId> <id>",
	Short: "프로필 상태 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.UpdateProfileStatus(args[0], args[1], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirProfileStatusPatchCmd = &cobra.Command{
	Use:   "patch <userId> <id>",
	Short: "프로필 상태 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchProfileStatus(args[0], args[1], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirProfileStatusDeleteCmd = &cobra.Command{
	Use:   "delete <userId> <id>",
	Short: "프로필 상태 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.DeleteProfileStatus(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── Task 4-3: Email + Invitations + Links ───

var dirAddAliasEmailCmd = &cobra.Command{
	Use:   "add-alias-email <userId> <email>",
	Short: "별칭 이메일 추가",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.AddAliasEmail(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirDeleteAliasEmailCmd = &cobra.Command{
	Use:   "delete-alias-email <userId> <email>",
	Short: "별칭 이메일 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.DeleteAliasEmail(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirSendInvitationCmd = &cobra.Command{
	Use:   "send-invitation <userId>",
	Short: "초대 이메일 발송",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.SendInvitationEmail(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirSendInvitationAllCmd = &cobra.Command{
	Use:   "send-invitation-all",
	Short: "전체 사용자 초대 이메일 발송",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.SendInvitationEmailToAll()
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirLinkToWorksCmd = &cobra.Command{
	Use:   "link-to-works <userId>",
	Short: "사용자 WORKS 연동",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.LinkUserToWorks(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirLinkAllToWorksCmd = &cobra.Command{
	Use:   "link-all-to-works",
	Short: "전체 사용자 WORKS 연동",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.LinkAllUsersToWorks()
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirUnlinkToWorksCmd = &cobra.Command{
	Use:   "unlink-to-works <userId>",
	Short: "사용자 WORKS 연동 해제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.UnlinkUserToWorks(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirLinkToLineCmd = &cobra.Command{
	Use:   "link-to-line <userId>",
	Short: "사용자 LINE 연동",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.LinkUserToLine(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirLinkAllToLineCmd = &cobra.Command{
	Use:   "link-all-to-line",
	Short: "전체 사용자 LINE 연동",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.LinkAllUsersToLine()
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirUnlinkToLineCmd = &cobra.Command{
	Use:   "unlink-to-line <userId>",
	Short: "사용자 LINE 연동 해제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.UnlinkUserToLine(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirGetLinkUrlCmd = &cobra.Command{
	Use:   "get-link-url <userId>",
	Short: "사용자 연동 URL 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewDirectoryService(client).GetLinkUrl(args[0])
		})
	},
}

var dirResetLinkUrlCmd = &cobra.Command{
	Use:   "reset-link-url <userId>",
	Short: "사용자 연동 URL 재설정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.ResetLinkUrl(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── Task 4-4: External Keys + Custom Properties ───

var dirUpsertExternalKeysCmd = &cobra.Command{
	Use:   "upsert-external-keys",
	Short: "사용자 외부 키 업서트",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.UpsertUserExternalKeys(body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirListExternalKeysCmd = &cobra.Command{
	Use:   "list-external-keys",
	Short: "사용자 외부 키 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, nil, "externalKeys", svc.ListUserExternalKeys)
	},
}

// ─── User Custom Property Subcommand Group ───

var dirUserCustomPropertyCmd = &cobra.Command{
	Use:   "user-custom-property",
	Short: "사용자 커스텀 속성 관리",
}

var dirUserCustomPropertyListCmd = &cobra.Command{
	Use:   "list",
	Short: "사용자 커스텀 속성 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, nil, "customProperties", svc.ListUserCustomProperties)
	},
}

var dirUserCustomPropertyGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "사용자 커스텀 속성 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewDirectoryService(client).GetUserCustomProperty(args[0])
		})
	},
}

var dirUserCustomPropertyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "사용자 커스텀 속성 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateUserCustomProperty(body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirUserCustomPropertyUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "사용자 커스텀 속성 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchUserCustomProperty(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirUserCustomPropertyDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "사용자 커스텀 속성 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.DeleteUserCustomProperty(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── Task 4-5: Group CUD + Members + Admins + External Keys ───

var dirCreateGroupCmd = &cobra.Command{
	Use:   "create-group",
	Short: "그룹 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateGroup(body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirUpdateGroupCmd = &cobra.Command{
	Use:   "update-group <groupId>",
	Short: "그룹 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.UpdateGroup(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirPatchGroupCmd = &cobra.Command{
	Use:   "patch-group <groupId>",
	Short: "그룹 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchGroup(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirDeleteGroupCmd = &cobra.Command{
	Use:   "delete-group <groupId>",
	Short: "그룹 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.DeleteGroup(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirListGroupMembersCmd = &cobra.Command{
	Use:   "list-group-members <groupId>",
	Short: "그룹 멤버 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, nil, "members", func(cursor string, count int) (*api.Response, error) {
			return svc.ListGroupMembers(args[0], cursor, count)
		})
	},
}

var dirAddGroupMembersCmd = &cobra.Command{
	Use:   "add-group-members <groupId>",
	Short: "그룹 멤버 추가",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.AddGroupMembers(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirRemoveGroupMemberCmd = &cobra.Command{
	Use:   "remove-group-member <groupId> <memberId>",
	Short: "그룹 멤버 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.RemoveGroupMember(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirListGroupAdminsCmd = &cobra.Command{
	Use:   "list-group-admins <groupId>",
	Short: "그룹 관리자 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, nil, "administrators", func(cursor string, count int) (*api.Response, error) {
			return svc.ListGroupAdministrators(args[0], cursor, count)
		})
	},
}

var dirAddGroupAdminCmd = &cobra.Command{
	Use:   "add-group-admin <groupId>",
	Short: "그룹 관리자 추가",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.AddGroupAdministrator(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirRemoveGroupAdminCmd = &cobra.Command{
	Use:   "remove-group-admin <groupId> <userId>",
	Short: "그룹 관리자 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.RemoveGroupAdministrator(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirUpsertGroupExternalKeysCmd = &cobra.Command{
	Use:   "upsert-group-external-keys",
	Short: "그룹 외부 키 업서트",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.UpsertGroupExternalKeys(body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirListGroupExternalKeysCmd = &cobra.Command{
	Use:   "list-group-external-keys",
	Short: "그룹 외부 키 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, nil, "externalKeys", svc.ListGroupExternalKeys)
	},
}

// ─── Task 4-6: OrgUnit CUD + Members + AccessRestrict + External Keys ───

var dirCreateOrgUnitCmd = &cobra.Command{
	Use:   "create-orgunit",
	Short: "조직 단위 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateOrgUnit(body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirUpdateOrgUnitCmd = &cobra.Command{
	Use:   "update-orgunit <orgUnitId>",
	Short: "조직 단위 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.UpdateOrgUnit(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirPatchOrgUnitCmd = &cobra.Command{
	Use:   "patch-orgunit <orgUnitId>",
	Short: "조직 단위 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchOrgUnit(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirDeleteOrgUnitCmd = &cobra.Command{
	Use:   "delete-orgunit <orgUnitId>",
	Short: "조직 단위 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.DeleteOrgUnit(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirMoveOrgUnitCmd = &cobra.Command{
	Use:   "move-orgunit <orgUnitId>",
	Short: "조직 단위 이동",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.MoveOrgUnit(args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirListOrgUnitMembersCmd = &cobra.Command{
	Use:   "list-orgunit-members <orgUnitId>",
	Short: "조직 단위 멤버 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, nil, "members", func(cursor string, count int) (*api.Response, error) {
			return svc.ListOrgUnitMembers(args[0], cursor, count)
		})
	},
}

// ─── OrgUnit Access Restrict Subcommand Group ───

var dirOrgUnitAccessRestrictCmd = &cobra.Command{
	Use:   "orgunit-access-restrict",
	Short: "조직 단위 접근 제한 관리",
}

var dirOrgUnitAccessRestrictCreateCmd = &cobra.Command{
	Use:   "create <orgUnitId>",
	Short: "조직 단위 접근 제한 생성",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateOrgUnitAccessRestrict(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirOrgUnitAccessRestrictGetCmd = &cobra.Command{
	Use:   "get <orgUnitId>",
	Short: "조직 단위 접근 제한 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewDirectoryService(client).GetOrgUnitAccessRestrict(args[0])
		})
	},
}

var dirOrgUnitAccessRestrictUpdateCmd = &cobra.Command{
	Use:   "update <orgUnitId>",
	Short: "조직 단위 접근 제한 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.UpdateOrgUnitAccessRestrict(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirOrgUnitAccessRestrictDeleteCmd = &cobra.Command{
	Use:   "delete <orgUnitId>",
	Short: "조직 단위 접근 제한 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := svc.DeleteOrgUnitAccessRestrict(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var dirUpsertOrgUnitExternalKeysCmd = &cobra.Command{
	Use:   "upsert-orgunit-external-keys",
	Short: "조직 단위 외부 키 업서트",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.UpsertOrgUnitExternalKeys(body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var dirListOrgUnitExternalKeysCmd = &cobra.Command{
	Use:   "list-orgunit-external-keys",
	Short: "조직 단위 외부 키 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, nil, "externalKeys", svc.ListOrgUnitExternalKeys)
	},
}

func init() {
	// ── List flags ──
	addListFlags(
		dirListUsersCmd, dirListGroupsCmd, dirListOrgUnitsCmd,
		dirListLevelsCmd, dirListPositionsCmd, dirListUserTypesCmd, dirListEmploymentTypesCmd,
		dirListExternalKeysCmd, dirListGroupExternalKeysCmd, dirListOrgUnitExternalKeysCmd,
		dirUserCustomPropertyListCmd,
		dirProfileStatusListCmd,
		dirListGroupMembersCmd, dirListGroupAdminsCmd,
		dirListOrgUnitMembersCmd,
	)

	// ── JSON flags ──
	for _, c := range []*cobra.Command{
		dirCreateUserCmd, dirUpdateUserCmd, dirPatchUserCmd,
		dirMoveUserCmd, dirSetLeaveCmd,
		dirProfileStatusCreateCmd, dirProfileStatusUpdateCmd, dirProfileStatusPatchCmd,
		dirUpsertExternalKeysCmd,
		dirUserCustomPropertyCreateCmd, dirUserCustomPropertyUpdateCmd,
		dirCreateGroupCmd, dirUpdateGroupCmd, dirPatchGroupCmd,
		dirAddGroupMembersCmd, dirAddGroupAdminCmd,
		dirUpsertGroupExternalKeysCmd,
		dirCreateOrgUnitCmd, dirUpdateOrgUnitCmd, dirPatchOrgUnitCmd,
		dirMoveOrgUnitCmd,
		dirOrgUnitAccessRestrictCreateCmd, dirOrgUnitAccessRestrictUpdateCmd,
		dirUpsertOrgUnitExternalKeysCmd,
	} {
		c.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")
	}

	// ── File flag ──
	dirUploadPhotoCmd.Flags().String("file", "", "사진 파일 경로 (필수)")

	// ── Profile Status subcommand group ──
	dirProfileStatusCmd.AddCommand(
		dirProfileStatusListCmd, dirProfileStatusGetCmd,
		dirProfileStatusCreateCmd, dirProfileStatusUpdateCmd, dirProfileStatusPatchCmd,
		dirProfileStatusDeleteCmd,
	)

	// ── User Custom Property subcommand group ──
	dirUserCustomPropertyCmd.AddCommand(
		dirUserCustomPropertyListCmd, dirUserCustomPropertyGetCmd,
		dirUserCustomPropertyCreateCmd, dirUserCustomPropertyUpdateCmd,
		dirUserCustomPropertyDeleteCmd,
	)

	// ── OrgUnit Access Restrict subcommand group ──
	dirOrgUnitAccessRestrictCmd.AddCommand(
		dirOrgUnitAccessRestrictCreateCmd, dirOrgUnitAccessRestrictGetCmd,
		dirOrgUnitAccessRestrictUpdateCmd, dirOrgUnitAccessRestrictDeleteCmd,
	)

	// ── Register all subcommands ──
	directoryCmd.AddCommand(
		// Existing read commands
		dirListUsersCmd, dirGetUserCmd,
		dirListGroupsCmd, dirGetGroupCmd,
		dirListOrgUnitsCmd, dirGetOrgUnitCmd,
		dirListLevelsCmd, dirListPositionsCmd, dirListUserTypesCmd, dirListEmploymentTypesCmd,
		// Task 4-1: User CUD
		dirCreateUserCmd, dirUpdateUserCmd, dirPatchUserCmd,
		dirDeleteUserCmd, dirForceDeleteUserCmd, dirUndeleteUserCmd,
		dirSuspendUserCmd, dirUnsuspendUserCmd, dirForceLogoutUserCmd,
		dirMoveUserCmd, dirSetLeaveCmd, dirClearLeaveCmd,
		// Task 4-2: User Profile
		dirUploadPhotoCmd, dirGetPhotoCmd, dirDeletePhotoCmd,
		dirProfileStatusCmd,
		// Task 4-3: Email + Invitations + Links
		dirAddAliasEmailCmd, dirDeleteAliasEmailCmd,
		dirSendInvitationCmd, dirSendInvitationAllCmd,
		dirLinkToWorksCmd, dirLinkAllToWorksCmd, dirUnlinkToWorksCmd,
		dirLinkToLineCmd, dirLinkAllToLineCmd, dirUnlinkToLineCmd,
		dirGetLinkUrlCmd, dirResetLinkUrlCmd,
		// Task 4-4: External Keys + Custom Properties
		dirUpsertExternalKeysCmd, dirListExternalKeysCmd,
		dirUserCustomPropertyCmd,
		// Task 4-5: Group CUD + Members + Admins + External Keys
		dirCreateGroupCmd, dirUpdateGroupCmd, dirPatchGroupCmd, dirDeleteGroupCmd,
		dirListGroupMembersCmd, dirAddGroupMembersCmd, dirRemoveGroupMemberCmd,
		dirListGroupAdminsCmd, dirAddGroupAdminCmd, dirRemoveGroupAdminCmd,
		dirUpsertGroupExternalKeysCmd, dirListGroupExternalKeysCmd,
		// Task 4-6: OrgUnit CUD + Members + AccessRestrict + External Keys
		dirCreateOrgUnitCmd, dirUpdateOrgUnitCmd, dirPatchOrgUnitCmd, dirDeleteOrgUnitCmd,
		dirMoveOrgUnitCmd, dirListOrgUnitMembersCmd,
		dirOrgUnitAccessRestrictCmd,
		dirUpsertOrgUnitExternalKeysCmd, dirListOrgUnitExternalKeysCmd,
	)

	rootCmd.AddCommand(directoryCmd)
}
