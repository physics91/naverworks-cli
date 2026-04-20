package cmd

import (
	"fmt"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

var directoryCmd = &cobra.Command{
	Use:   "directory",
	Short: "디렉토리 관리 (사용자, 그룹, 조직)",
}

type directoryCall func(*api.DirectoryService) (*api.Response, error)
type directoryBodyCall func(*api.DirectoryService, []byte) (*api.Response, error)
type directoryIDCall func(*api.DirectoryService, string) (*api.Response, error)
type directoryIDBodyCall func(*api.DirectoryService, string, []byte) (*api.Response, error)
type directoryTwoIDCall func(*api.DirectoryService, string, string) (*api.Response, error)
type directoryTwoIDBodyCall func(*api.DirectoryService, string, string, []byte) (*api.Response, error)

// Keep these wrappers local so cmd/helpers.go does not grow a directory-only helper family.
func printDirectoryBody(resp *api.Response) {
	printBody(resp.Body)
}

func directoryRunE(call directoryCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		resp, err := call(svc)
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func directoryBodyRunE(call directoryBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
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

func directoryIDRunE(call directoryIDCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
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

func directoryIDBodyRunE(call directoryIDBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
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

func directoryTwoIDRunE(call directoryTwoIDCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
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

func directoryTwoIDBodyRunE(call directoryTwoIDBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
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
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
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
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
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
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
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
	RunE:  directoryBodyRunE((*api.DirectoryService).CreateUser, printDirectoryBody),
}

var dirUpdateUserCmd = &cobra.Command{
	Use:   "update-user <userId>",
	Short: "사용자 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).UpdateUser, printDirectoryBody),
}

var dirPatchUserCmd = &cobra.Command{
	Use:   "patch-user <userId>",
	Short: "사용자 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).PatchUser, printDirectoryBody),
}

var dirDeleteUserCmd = &cobra.Command{
	Use:   "delete-user <userId>",
	Short: "사용자 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).DeleteUser, printResponse),
}

var dirForceDeleteUserCmd = &cobra.Command{
	Use:   "force-delete-user <userId>",
	Short: "사용자 강제 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).ForceDeleteUser, printResponse),
}

var dirUndeleteUserCmd = &cobra.Command{
	Use:   "undelete-user <userId>",
	Short: "사용자 삭제 취소",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).UndeleteUser, printResponse),
}

var dirSuspendUserCmd = &cobra.Command{
	Use:   "suspend-user <userId>",
	Short: "사용자 정지",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).SuspendUser, printResponse),
}

var dirUnsuspendUserCmd = &cobra.Command{
	Use:   "unsuspend-user <userId>",
	Short: "사용자 정지 해제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).UnsuspendUser, printResponse),
}

var dirForceLogoutUserCmd = &cobra.Command{
	Use:   "force-logout-user <userId>",
	Short: "사용자 강제 로그아웃",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).ForceLogoutUser, printResponse),
}

var dirMoveUserCmd = &cobra.Command{
	Use:   "move-user <userId>",
	Short: "사용자 이동",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).MoveUser, printResponse),
}

var dirSetLeaveCmd = &cobra.Command{
	Use:   "set-leave <userId>",
	Short: "사용자 휴직 설정",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).SetLeaveOfAbsence, printResponse),
}

var dirClearLeaveCmd = &cobra.Command{
	Use:   "clear-leave <userId>",
	Short: "사용자 휴직 해제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).ClearLeaveOfAbsence, printResponse),
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
		fileName, fileSize, err := statFileForUpload(filePath)
		if err != nil {
			return err
		}

		uploadBody := map[string]interface{}{
			"fileName": fileName,
			"fileSize": fileSize,
		}
		resp, err := svc.CreateUserPhoto(args[0], uploadBody)
		if err != nil {
			return err
		}

		respBody, err := doUploadFromResponse(client, resp.Body, filePath)
		if err != nil {
			return err
		}
		printBody(respBody)
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
	RunE:  directoryIDRunE((*api.DirectoryService).DeleteUserPhoto, printResponse),
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
	RunE:  directoryTwoIDRunE((*api.DirectoryService).GetProfileStatus, printDirectoryBody),
}

var dirProfileStatusCreateCmd = &cobra.Command{
	Use:   "create <userId>",
	Short: "프로필 상태 생성",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).CreateProfileStatus, printDirectoryBody),
}

var dirProfileStatusUpdateCmd = &cobra.Command{
	Use:   "update <userId> <id>",
	Short: "프로필 상태 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(2),
	RunE:  directoryTwoIDBodyRunE((*api.DirectoryService).UpdateProfileStatus, printDirectoryBody),
}

var dirProfileStatusPatchCmd = &cobra.Command{
	Use:   "patch <userId> <id>",
	Short: "프로필 상태 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(2),
	RunE:  directoryTwoIDBodyRunE((*api.DirectoryService).PatchProfileStatus, printDirectoryBody),
}

var dirProfileStatusDeleteCmd = &cobra.Command{
	Use:   "delete <userId> <id>",
	Short: "프로필 상태 삭제",
	Args:  cobra.ExactArgs(2),
	RunE:  directoryTwoIDRunE((*api.DirectoryService).DeleteProfileStatus, printResponse),
}

// ─── Task 4-3: Email + Invitations + Links ───

var dirAddAliasEmailCmd = &cobra.Command{
	Use:   "add-alias-email <userId> <email>",
	Short: "별칭 이메일 추가",
	Args:  cobra.ExactArgs(2),
	RunE:  directoryTwoIDRunE((*api.DirectoryService).AddAliasEmail, printResponse),
}

var dirDeleteAliasEmailCmd = &cobra.Command{
	Use:   "delete-alias-email <userId> <email>",
	Short: "별칭 이메일 삭제",
	Args:  cobra.ExactArgs(2),
	RunE:  directoryTwoIDRunE((*api.DirectoryService).DeleteAliasEmail, printResponse),
}

var dirSendInvitationCmd = &cobra.Command{
	Use:   "send-invitation <userId>",
	Short: "초대 이메일 발송",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).SendInvitationEmail, printResponse),
}

var dirSendInvitationAllCmd = &cobra.Command{
	Use:   "send-invitation-all",
	Short: "전체 사용자 초대 이메일 발송",
	RunE:  directoryRunE((*api.DirectoryService).SendInvitationEmailToAll, printResponse),
}

var dirLinkToWorksCmd = &cobra.Command{
	Use:   "link-to-works <userId>",
	Short: "사용자 WORKS 연동",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).LinkUserToWorks, printResponse),
}

var dirLinkAllToWorksCmd = &cobra.Command{
	Use:   "link-all-to-works",
	Short: "전체 사용자 WORKS 연동",
	RunE:  directoryRunE((*api.DirectoryService).LinkAllUsersToWorks, printResponse),
}

var dirUnlinkToWorksCmd = &cobra.Command{
	Use:   "unlink-to-works <userId>",
	Short: "사용자 WORKS 연동 해제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).UnlinkUserToWorks, printResponse),
}

var dirLinkToLineCmd = &cobra.Command{
	Use:   "link-to-line <userId>",
	Short: "사용자 LINE 연동",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).LinkUserToLine, printResponse),
}

var dirLinkAllToLineCmd = &cobra.Command{
	Use:   "link-all-to-line",
	Short: "전체 사용자 LINE 연동",
	RunE:  directoryRunE((*api.DirectoryService).LinkAllUsersToLine, printResponse),
}

var dirUnlinkToLineCmd = &cobra.Command{
	Use:   "unlink-to-line <userId>",
	Short: "사용자 LINE 연동 해제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).UnlinkUserToLine, printResponse),
}

var dirGetLinkUrlCmd = &cobra.Command{
	Use:   "get-link-url <userId>",
	Short: "사용자 연동 URL 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).GetLinkUrl, printDirectoryBody),
}

var dirResetLinkUrlCmd = &cobra.Command{
	Use:   "reset-link-url <userId>",
	Short: "사용자 연동 URL 재설정",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).ResetLinkUrl, printResponse),
}

// ─── Task 4-4: External Keys + Custom Properties ───

var dirUpsertExternalKeysCmd = &cobra.Command{
	Use:   "upsert-external-keys",
	Short: "사용자 외부 키 업서트",
	RunE:  directoryBodyRunE((*api.DirectoryService).UpsertUserExternalKeys, printDirectoryBody),
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
	RunE:  directoryIDRunE((*api.DirectoryService).GetUserCustomProperty, printDirectoryBody),
}

var dirUserCustomPropertyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "사용자 커스텀 속성 생성",
	RunE:  directoryBodyRunE((*api.DirectoryService).CreateUserCustomProperty, printDirectoryBody),
}

var dirUserCustomPropertyUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "사용자 커스텀 속성 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).PatchUserCustomProperty, printDirectoryBody),
}

var dirUserCustomPropertyDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "사용자 커스텀 속성 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).DeleteUserCustomProperty, printResponse),
}

// ─── Task 4-5: Group CUD + Members + Admins + External Keys ───

var dirCreateGroupCmd = &cobra.Command{
	Use:   "create-group",
	Short: "그룹 생성",
	RunE:  directoryBodyRunE((*api.DirectoryService).CreateGroup, printDirectoryBody),
}

var dirUpdateGroupCmd = &cobra.Command{
	Use:   "update-group <groupId>",
	Short: "그룹 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).UpdateGroup, printDirectoryBody),
}

var dirPatchGroupCmd = &cobra.Command{
	Use:   "patch-group <groupId>",
	Short: "그룹 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).PatchGroup, printDirectoryBody),
}

var dirDeleteGroupCmd = &cobra.Command{
	Use:   "delete-group <groupId>",
	Short: "그룹 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).DeleteGroup, printResponse),
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
	RunE:  directoryIDBodyRunE((*api.DirectoryService).AddGroupMembers, printDirectoryBody),
}

var dirRemoveGroupMemberCmd = &cobra.Command{
	Use:   "remove-group-member <groupId> <memberId>",
	Short: "그룹 멤버 삭제",
	Args:  cobra.ExactArgs(2),
	RunE:  directoryTwoIDRunE((*api.DirectoryService).RemoveGroupMember, printResponse),
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
	RunE:  directoryIDBodyRunE((*api.DirectoryService).AddGroupAdministrator, printDirectoryBody),
}

var dirRemoveGroupAdminCmd = &cobra.Command{
	Use:   "remove-group-admin <groupId> <userId>",
	Short: "그룹 관리자 삭제",
	Args:  cobra.ExactArgs(2),
	RunE:  directoryTwoIDRunE((*api.DirectoryService).RemoveGroupAdministrator, printResponse),
}

var dirUpsertGroupExternalKeysCmd = &cobra.Command{
	Use:   "upsert-group-external-keys",
	Short: "그룹 외부 키 업서트",
	RunE:  directoryBodyRunE((*api.DirectoryService).UpsertGroupExternalKeys, printDirectoryBody),
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
	RunE:  directoryBodyRunE((*api.DirectoryService).CreateOrgUnit, printDirectoryBody),
}

var dirUpdateOrgUnitCmd = &cobra.Command{
	Use:   "update-orgunit <orgUnitId>",
	Short: "조직 단위 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).UpdateOrgUnit, printDirectoryBody),
}

var dirPatchOrgUnitCmd = &cobra.Command{
	Use:   "patch-orgunit <orgUnitId>",
	Short: "조직 단위 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).PatchOrgUnit, printDirectoryBody),
}

var dirDeleteOrgUnitCmd = &cobra.Command{
	Use:   "delete-orgunit <orgUnitId>",
	Short: "조직 단위 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).DeleteOrgUnit, printResponse),
}

var dirMoveOrgUnitCmd = &cobra.Command{
	Use:   "move-orgunit <orgUnitId>",
	Short: "조직 단위 이동",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).MoveOrgUnit, printResponse),
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
	RunE:  directoryIDBodyRunE((*api.DirectoryService).CreateOrgUnitAccessRestrict, printDirectoryBody),
}

var dirOrgUnitAccessRestrictGetCmd = &cobra.Command{
	Use:   "get <orgUnitId>",
	Short: "조직 단위 접근 제한 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).GetOrgUnitAccessRestrict, printDirectoryBody),
}

var dirOrgUnitAccessRestrictUpdateCmd = &cobra.Command{
	Use:   "update <orgUnitId>",
	Short: "조직 단위 접근 제한 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).UpdateOrgUnitAccessRestrict, printDirectoryBody),
}

var dirOrgUnitAccessRestrictDeleteCmd = &cobra.Command{
	Use:   "delete <orgUnitId>",
	Short: "조직 단위 접근 제한 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).DeleteOrgUnitAccessRestrict, printResponse),
}

var dirUpsertOrgUnitExternalKeysCmd = &cobra.Command{
	Use:   "upsert-orgunit-external-keys",
	Short: "조직 단위 외부 키 업서트",
	RunE:  directoryBodyRunE((*api.DirectoryService).UpsertOrgUnitExternalKeys, printDirectoryBody),
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

// ─── Task 4-7: Positions CRUD + External Keys ───

var dirGetPositionCmd = &cobra.Command{
	Use:   "get-position <positionId>",
	Short: "직책 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).GetPosition, printDirectoryBody),
}

var dirCreatePositionCmd = &cobra.Command{
	Use:   "create-position",
	Short: "직책 생성",
	RunE:  directoryBodyRunE((*api.DirectoryService).CreatePosition, printDirectoryBody),
}

var dirUpdatePositionCmd = &cobra.Command{
	Use:   "update-position <positionId>",
	Short: "직책 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).UpdatePosition, printDirectoryBody),
}

var dirPatchPositionCmd = &cobra.Command{
	Use:   "patch-position <positionId>",
	Short: "직책 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).PatchPosition, printDirectoryBody),
}

var dirDeletePositionCmd = &cobra.Command{
	Use:   "delete-position <positionId>",
	Short: "직책 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).DeletePosition, printResponse),
}

var dirEnablePositionsCmd = &cobra.Command{
	Use:   "enable-positions",
	Short: "직책 활성화",
	RunE:  directoryRunE((*api.DirectoryService).EnablePositions, printResponse),
}

var dirDisablePositionsCmd = &cobra.Command{
	Use:   "disable-positions",
	Short: "직책 비활성화",
	RunE:  directoryRunE((*api.DirectoryService).DisablePositions, printResponse),
}

var dirUpsertPositionExternalKeysCmd = &cobra.Command{
	Use:   "upsert-position-external-keys",
	Short: "직책 외부 키 업서트",
	RunE:  directoryBodyRunE((*api.DirectoryService).UpsertPositionExternalKeys, printDirectoryBody),
}

var dirListPositionExternalKeysCmd = &cobra.Command{
	Use:   "list-position-external-keys",
	Short: "직책 외부 키 목록 조회",
	RunE:  directoryRunE((*api.DirectoryService).ListPositionExternalKeys, printDirectoryBody),
}

// ─── Task 4-8: Levels CRUD + External Keys ───

var dirGetLevelCmd = &cobra.Command{
	Use:   "get-level <levelId>",
	Short: "직급 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).GetLevel, printDirectoryBody),
}

var dirCreateLevelCmd = &cobra.Command{
	Use:   "create-level",
	Short: "직급 생성",
	RunE:  directoryBodyRunE((*api.DirectoryService).CreateLevel, printDirectoryBody),
}

var dirUpdateLevelCmd = &cobra.Command{
	Use:   "update-level <levelId>",
	Short: "직급 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).UpdateLevel, printDirectoryBody),
}

var dirPatchLevelCmd = &cobra.Command{
	Use:   "patch-level <levelId>",
	Short: "직급 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).PatchLevel, printDirectoryBody),
}

var dirDeleteLevelCmd = &cobra.Command{
	Use:   "delete-level <levelId>",
	Short: "직급 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).DeleteLevel, printResponse),
}

var dirEnableLevelsCmd = &cobra.Command{
	Use:   "enable-levels",
	Short: "직급 활성화",
	RunE:  directoryRunE((*api.DirectoryService).EnableLevels, printResponse),
}

var dirDisableLevelsCmd = &cobra.Command{
	Use:   "disable-levels",
	Short: "직급 비활성화",
	RunE:  directoryRunE((*api.DirectoryService).DisableLevels, printResponse),
}

var dirUpsertLevelExternalKeysCmd = &cobra.Command{
	Use:   "upsert-level-external-keys",
	Short: "직급 외부 키 업서트",
	RunE:  directoryBodyRunE((*api.DirectoryService).UpsertLevelExternalKeys, printDirectoryBody),
}

var dirListLevelExternalKeysCmd = &cobra.Command{
	Use:   "list-level-external-keys",
	Short: "직급 외부 키 목록 조회",
	RunE:  directoryRunE((*api.DirectoryService).ListLevelExternalKeys, printDirectoryBody),
}

// ─── Task 4-9: Employment Types CRUD + External Keys + Access Restrict ───

var dirGetEmploymentTypeCmd = &cobra.Command{
	Use:   "get-employment-type <id>",
	Short: "고용 유형 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).GetEmploymentType, printDirectoryBody),
}

var dirCreateEmploymentTypeCmd = &cobra.Command{
	Use:   "create-employment-type",
	Short: "고용 유형 생성",
	RunE:  directoryBodyRunE((*api.DirectoryService).CreateEmploymentType, printDirectoryBody),
}

var dirUpdateEmploymentTypeCmd = &cobra.Command{
	Use:   "update-employment-type <id>",
	Short: "고용 유형 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).UpdateEmploymentType, printDirectoryBody),
}

var dirPatchEmploymentTypeCmd = &cobra.Command{
	Use:   "patch-employment-type <id>",
	Short: "고용 유형 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).PatchEmploymentType, printDirectoryBody),
}

var dirDeleteEmploymentTypeCmd = &cobra.Command{
	Use:   "delete-employment-type <id>",
	Short: "고용 유형 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).DeleteEmploymentType, printResponse),
}

var dirEnableEmploymentTypesCmd = &cobra.Command{
	Use:   "enable-employment-types",
	Short: "고용 유형 활성화",
	RunE:  directoryRunE((*api.DirectoryService).EnableEmploymentTypes, printResponse),
}

var dirDisableEmploymentTypesCmd = &cobra.Command{
	Use:   "disable-employment-types",
	Short: "고용 유형 비활성화",
	RunE:  directoryRunE((*api.DirectoryService).DisableEmploymentTypes, printResponse),
}

var dirUpsertEmploymentTypeExternalKeysCmd = &cobra.Command{
	Use:   "upsert-employment-type-external-keys",
	Short: "고용 유형 외부 키 업서트",
	RunE:  directoryBodyRunE((*api.DirectoryService).UpsertEmploymentTypeExternalKeys, printDirectoryBody),
}

var dirListEmploymentTypeExternalKeysCmd = &cobra.Command{
	Use:   "list-employment-type-external-keys",
	Short: "고용 유형 외부 키 목록 조회",
	RunE:  directoryRunE((*api.DirectoryService).ListEmploymentTypeExternalKeys, printDirectoryBody),
}

// ─── Employment Type Access Restrict Subcommand Group ───

var dirEmploymentTypeAccessRestrictCmd = &cobra.Command{
	Use:   "employment-type-access-restrict",
	Short: "고용 유형 접근 제한 관리",
}

var dirEmploymentTypeAccessRestrictCreateCmd = &cobra.Command{
	Use:   "create <id>",
	Short: "고용 유형 접근 제한 생성",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).CreateEmploymentTypeAccessRestrict, printDirectoryBody),
}

var dirEmploymentTypeAccessRestrictGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "고용 유형 접근 제한 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).GetEmploymentTypeAccessRestrict, printDirectoryBody),
}

var dirEmploymentTypeAccessRestrictUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "고용 유형 접근 제한 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).UpdateEmploymentTypeAccessRestrict, printDirectoryBody),
}

var dirEmploymentTypeAccessRestrictDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "고용 유형 접근 제한 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).DeleteEmploymentTypeAccessRestrict, printResponse),
}

// ─── Task 4-10: User Types CRUD + External Keys + Access Restrict ───

var dirGetUserTypeCmd = &cobra.Command{
	Use:   "get-user-type <id>",
	Short: "사용자 유형 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).GetUserType, printDirectoryBody),
}

var dirCreateUserTypeCmd = &cobra.Command{
	Use:   "create-user-type",
	Short: "사용자 유형 생성",
	RunE:  directoryBodyRunE((*api.DirectoryService).CreateUserType, printDirectoryBody),
}

var dirUpdateUserTypeCmd = &cobra.Command{
	Use:   "update-user-type <id>",
	Short: "사용자 유형 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).UpdateUserType, printDirectoryBody),
}

var dirPatchUserTypeCmd = &cobra.Command{
	Use:   "patch-user-type <id>",
	Short: "사용자 유형 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).PatchUserType, printDirectoryBody),
}

var dirDeleteUserTypeCmd = &cobra.Command{
	Use:   "delete-user-type <id>",
	Short: "사용자 유형 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).DeleteUserType, printResponse),
}

var dirEnableUserTypesCmd = &cobra.Command{
	Use:   "enable-user-types",
	Short: "사용자 유형 활성화",
	RunE:  directoryRunE((*api.DirectoryService).EnableUserTypes, printResponse),
}

var dirDisableUserTypesCmd = &cobra.Command{
	Use:   "disable-user-types",
	Short: "사용자 유형 비활성화",
	RunE:  directoryRunE((*api.DirectoryService).DisableUserTypes, printResponse),
}

var dirUpsertUserTypeExternalKeysCmd = &cobra.Command{
	Use:   "upsert-user-type-external-keys",
	Short: "사용자 유형 외부 키 업서트",
	RunE:  directoryBodyRunE((*api.DirectoryService).UpsertUserTypeExternalKeys, printDirectoryBody),
}

var dirListUserTypeExternalKeysCmd = &cobra.Command{
	Use:   "list-user-type-external-keys",
	Short: "사용자 유형 외부 키 목록 조회",
	RunE:  directoryRunE((*api.DirectoryService).ListUserTypeExternalKeys, printDirectoryBody),
}

// ─── User Type Access Restrict Subcommand Group ───

var dirUserTypeAccessRestrictCmd = &cobra.Command{
	Use:   "user-type-access-restrict",
	Short: "사용자 유형 접근 제한 관리",
}

var dirUserTypeAccessRestrictCreateCmd = &cobra.Command{
	Use:   "create <id>",
	Short: "사용자 유형 접근 제한 생성",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).CreateUserTypeAccessRestrict, printDirectoryBody),
}

var dirUserTypeAccessRestrictGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "사용자 유형 접근 제한 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).GetUserTypeAccessRestrict, printDirectoryBody),
}

var dirUserTypeAccessRestrictUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "사용자 유형 접근 제한 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).UpdateUserTypeAccessRestrict, printDirectoryBody),
}

var dirUserTypeAccessRestrictDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "사용자 유형 접근 제한 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).DeleteUserTypeAccessRestrict, printResponse),
}

// ─── Task 4-11: Profile Statuses Definition CRUD ───

var dirProfileStatusDefCmd = &cobra.Command{
	Use:   "profile-status-def",
	Short: "프로필 상태 정의 관리",
}

var dirProfileStatusDefListCmd = &cobra.Command{
	Use:   "list",
	Short: "프로필 상태 정의 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, nil, "profileStatuses", svc.ListDirectoryProfileStatuses)
	},
}

var dirProfileStatusDefGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "프로필 상태 정의 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).GetDirectoryProfileStatus, printDirectoryBody),
}

var dirProfileStatusDefCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "프로필 상태 정의 생성",
	RunE:  directoryBodyRunE((*api.DirectoryService).CreateDirectoryProfileStatus, printDirectoryBody),
}

var dirProfileStatusDefUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "프로필 상태 정의 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).UpdateDirectoryProfileStatus, printDirectoryBody),
}

var dirProfileStatusDefPatchCmd = &cobra.Command{
	Use:   "patch <id>",
	Short: "프로필 상태 정의 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).PatchDirectoryProfileStatus, printDirectoryBody),
}

var dirProfileStatusDefDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "프로필 상태 정의 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).DeleteDirectoryProfileStatus, printResponse),
}

var dirProfileStatusDefEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "프로필 상태 정의 활성화",
	RunE:  directoryRunE((*api.DirectoryService).EnableDirectoryProfileStatuses, printResponse),
}

var dirProfileStatusDefDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "프로필 상태 정의 비활성화",
	RunE:  directoryRunE((*api.DirectoryService).DisableDirectoryProfileStatuses, printResponse),
}

// ─── Task 4-12: Custom Fields CRUD ───

var dirCustomFieldCmd = &cobra.Command{
	Use:   "custom-field",
	Short: "커스텀 필드 관리",
}

var dirCustomFieldListCmd = &cobra.Command{
	Use:   "list",
	Short: "커스텀 필드 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewDirectoryService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, nil, "customFields", svc.ListCustomFields)
	},
}

var dirCustomFieldGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "커스텀 필드 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).GetCustomField, printDirectoryBody),
}

var dirCustomFieldCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "커스텀 필드 생성",
	RunE:  directoryBodyRunE((*api.DirectoryService).CreateCustomField, printDirectoryBody),
}

var dirCustomFieldUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "커스텀 필드 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDBodyRunE((*api.DirectoryService).PatchCustomField, printDirectoryBody),
}

var dirCustomFieldDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "커스텀 필드 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  directoryIDRunE((*api.DirectoryService).DeleteCustomField, printResponse),
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
		dirProfileStatusDefListCmd,
		dirCustomFieldListCmd,
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
		// Task 4-7: Positions
		dirCreatePositionCmd, dirUpdatePositionCmd, dirPatchPositionCmd,
		dirUpsertPositionExternalKeysCmd,
		// Task 4-8: Levels
		dirCreateLevelCmd, dirUpdateLevelCmd, dirPatchLevelCmd,
		dirUpsertLevelExternalKeysCmd,
		// Task 4-9: Employment Types
		dirCreateEmploymentTypeCmd, dirUpdateEmploymentTypeCmd, dirPatchEmploymentTypeCmd,
		dirUpsertEmploymentTypeExternalKeysCmd,
		dirEmploymentTypeAccessRestrictCreateCmd, dirEmploymentTypeAccessRestrictUpdateCmd,
		// Task 4-10: User Types
		dirCreateUserTypeCmd, dirUpdateUserTypeCmd, dirPatchUserTypeCmd,
		dirUpsertUserTypeExternalKeysCmd,
		dirUserTypeAccessRestrictCreateCmd, dirUserTypeAccessRestrictUpdateCmd,
		// Task 4-11: Profile Statuses Def
		dirProfileStatusDefCreateCmd, dirProfileStatusDefUpdateCmd, dirProfileStatusDefPatchCmd,
		// Task 4-12: Custom Fields
		dirCustomFieldCreateCmd, dirCustomFieldUpdateCmd,
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

	// ── Employment Type Access Restrict subcommand group ──
	dirEmploymentTypeAccessRestrictCmd.AddCommand(
		dirEmploymentTypeAccessRestrictCreateCmd, dirEmploymentTypeAccessRestrictGetCmd,
		dirEmploymentTypeAccessRestrictUpdateCmd, dirEmploymentTypeAccessRestrictDeleteCmd,
	)

	// ── User Type Access Restrict subcommand group ──
	dirUserTypeAccessRestrictCmd.AddCommand(
		dirUserTypeAccessRestrictCreateCmd, dirUserTypeAccessRestrictGetCmd,
		dirUserTypeAccessRestrictUpdateCmd, dirUserTypeAccessRestrictDeleteCmd,
	)

	// ── Profile Status Definition subcommand group ──
	dirProfileStatusDefCmd.AddCommand(
		dirProfileStatusDefListCmd, dirProfileStatusDefGetCmd,
		dirProfileStatusDefCreateCmd, dirProfileStatusDefUpdateCmd, dirProfileStatusDefPatchCmd,
		dirProfileStatusDefDeleteCmd,
		dirProfileStatusDefEnableCmd, dirProfileStatusDefDisableCmd,
	)

	// ── Custom Field subcommand group ──
	dirCustomFieldCmd.AddCommand(
		dirCustomFieldListCmd, dirCustomFieldGetCmd,
		dirCustomFieldCreateCmd, dirCustomFieldUpdateCmd,
		dirCustomFieldDeleteCmd,
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
		// Task 4-7: Positions CRUD + External Keys
		dirGetPositionCmd, dirCreatePositionCmd, dirUpdatePositionCmd, dirPatchPositionCmd,
		dirDeletePositionCmd, dirEnablePositionsCmd, dirDisablePositionsCmd,
		dirUpsertPositionExternalKeysCmd, dirListPositionExternalKeysCmd,
		// Task 4-8: Levels CRUD + External Keys
		dirGetLevelCmd, dirCreateLevelCmd, dirUpdateLevelCmd, dirPatchLevelCmd,
		dirDeleteLevelCmd, dirEnableLevelsCmd, dirDisableLevelsCmd,
		dirUpsertLevelExternalKeysCmd, dirListLevelExternalKeysCmd,
		// Task 4-9: Employment Types CRUD + External Keys + Access Restrict
		dirGetEmploymentTypeCmd, dirCreateEmploymentTypeCmd, dirUpdateEmploymentTypeCmd, dirPatchEmploymentTypeCmd,
		dirDeleteEmploymentTypeCmd, dirEnableEmploymentTypesCmd, dirDisableEmploymentTypesCmd,
		dirUpsertEmploymentTypeExternalKeysCmd, dirListEmploymentTypeExternalKeysCmd,
		dirEmploymentTypeAccessRestrictCmd,
		// Task 4-10: User Types CRUD + External Keys + Access Restrict
		dirGetUserTypeCmd, dirCreateUserTypeCmd, dirUpdateUserTypeCmd, dirPatchUserTypeCmd,
		dirDeleteUserTypeCmd, dirEnableUserTypesCmd, dirDisableUserTypesCmd,
		dirUpsertUserTypeExternalKeysCmd, dirListUserTypeExternalKeysCmd,
		dirUserTypeAccessRestrictCmd,
		// Task 4-11: Profile Statuses Definition
		dirProfileStatusDefCmd,
		// Task 4-12: Custom Fields
		dirCustomFieldCmd,
	)

	rootCmd.AddCommand(directoryCmd)
}
