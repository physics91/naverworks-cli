package api

import (
	"fmt"
	"net/url"
)

type DirectoryService struct {
	client *Client
}

func NewDirectoryService(client *Client) *DirectoryService {
	return &DirectoryService{client: client}
}

// ─── User Read ───

func (s *DirectoryService) ListUsers(cursor string, count int) (*Response, error) {
	return s.client.Get("/users" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) GetUser(userID string) (*Response, error) {
	return s.client.Get("/users/" + url.PathEscape(userID))
}

// ─── Task 4-1: User CUD ───

func (s *DirectoryService) CreateUser(body []byte) (*Response, error) {
	return s.client.Post("/users", body)
}

func (s *DirectoryService) UpdateUser(userID string, body []byte) (*Response, error) {
	return s.client.Put("/users/"+url.PathEscape(userID), body)
}

func (s *DirectoryService) PatchUser(userID string, body []byte) (*Response, error) {
	return s.client.Patch("/users/"+url.PathEscape(userID), body)
}

func (s *DirectoryService) DeleteUser(userID string) (*Response, error) {
	return s.client.Delete("/users/" + url.PathEscape(userID))
}

func (s *DirectoryService) ForceDeleteUser(userID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/forcedelete", url.PathEscape(userID)))
}

func (s *DirectoryService) UndeleteUser(userID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/undelete", url.PathEscape(userID)), nil)
}

func (s *DirectoryService) SuspendUser(userID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/suspend", url.PathEscape(userID)), nil)
}

func (s *DirectoryService) UnsuspendUser(userID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/unsuspend", url.PathEscape(userID)), nil)
}

func (s *DirectoryService) ForceLogoutUser(userID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/force-logout", url.PathEscape(userID)), nil)
}

func (s *DirectoryService) MoveUser(userID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/move", url.PathEscape(userID)), body)
}

func (s *DirectoryService) SetLeaveOfAbsence(userID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/set-leave-of-absence", url.PathEscape(userID)), body)
}

func (s *DirectoryService) ClearLeaveOfAbsence(userID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/clear-leave-of-absence", url.PathEscape(userID)), nil)
}

// ─── Task 4-2: User Profile (Photo + Profile Status) ───

func (s *DirectoryService) CreateUserPhoto(userID string, body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/users/%s/photo", url.PathEscape(userID)), body)
}

func (s *DirectoryService) GetUserPhoto(userID string) (string, error) {
	return s.client.GetDownloadURL(fmt.Sprintf("/users/%s/photo", url.PathEscape(userID)))
}

func (s *DirectoryService) DeleteUserPhoto(userID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/photo", url.PathEscape(userID)))
}

func (s *DirectoryService) CreateProfileStatus(userID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/user-profile-statuses", url.PathEscape(userID)), body)
}

func (s *DirectoryService) ListProfileStatuses(userID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/user-profile-statuses", url.PathEscape(userID)) + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) GetProfileStatus(userID string, id string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/user-profile-statuses/%s", url.PathEscape(userID), url.PathEscape(id)))
}

func (s *DirectoryService) UpdateProfileStatus(userID string, id string, body []byte) (*Response, error) {
	return s.client.Put(fmt.Sprintf("/users/%s/user-profile-statuses/%s", url.PathEscape(userID), url.PathEscape(id)), body)
}

func (s *DirectoryService) PatchProfileStatus(userID string, id string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/users/%s/user-profile-statuses/%s", url.PathEscape(userID), url.PathEscape(id)), body)
}

func (s *DirectoryService) DeleteProfileStatus(userID string, id string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/user-profile-statuses/%s", url.PathEscape(userID), url.PathEscape(id)))
}

// ─── Task 4-3: Email + Invitations + Links ───

func (s *DirectoryService) AddAliasEmail(userID string, email string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/alias-emails/%s", url.PathEscape(userID), url.PathEscape(email)), nil)
}

func (s *DirectoryService) DeleteAliasEmail(userID string, email string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/alias-emails/%s", url.PathEscape(userID), url.PathEscape(email)))
}

func (s *DirectoryService) SendInvitationEmail(userID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/send-invitation-email", url.PathEscape(userID)), nil)
}

func (s *DirectoryService) SendInvitationEmailToAll() (*Response, error) {
	return s.client.Post("/users/send-invitation-email", nil)
}

func (s *DirectoryService) LinkAllUsersToWorks() (*Response, error) {
	return s.client.Post("/users/link-to-works", nil)
}

func (s *DirectoryService) LinkUserToWorks(userID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/link-to-works", url.PathEscape(userID)), nil)
}

func (s *DirectoryService) UnlinkUserToWorks(userID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/unlink-to-works", url.PathEscape(userID)), nil)
}

func (s *DirectoryService) LinkAllUsersToLine() (*Response, error) {
	return s.client.Post("/users/link-to-line", nil)
}

func (s *DirectoryService) LinkUserToLine(userID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/link-to-line", url.PathEscape(userID)), nil)
}

func (s *DirectoryService) UnlinkUserToLine(userID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/unlink-to-line", url.PathEscape(userID)), nil)
}

func (s *DirectoryService) GetLinkUrl(userID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/link-url", url.PathEscape(userID)))
}

func (s *DirectoryService) ResetLinkUrl(userID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/link-url/reset", url.PathEscape(userID)), nil)
}

// ─── Task 4-4: External Keys + Custom Properties ───

func (s *DirectoryService) UpsertUserExternalKeys(body []byte) (*Response, error) {
	return s.client.Post("/users/external-keys", body)
}

func (s *DirectoryService) ListUserExternalKeys(cursor string, count int) (*Response, error) {
	return s.client.Get("/users/external-keys" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) CreateUserCustomProperty(body []byte) (*Response, error) {
	return s.client.Post("/directory/users/custom-properties", body)
}

func (s *DirectoryService) ListUserCustomProperties(cursor string, count int) (*Response, error) {
	return s.client.Get("/directory/users/custom-properties" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) GetUserCustomProperty(id string) (*Response, error) {
	return s.client.Get("/directory/users/custom-properties/" + url.PathEscape(id))
}

func (s *DirectoryService) PatchUserCustomProperty(id string, body []byte) (*Response, error) {
	return s.client.Patch("/directory/users/custom-properties/"+url.PathEscape(id), body)
}

func (s *DirectoryService) DeleteUserCustomProperty(id string) (*Response, error) {
	return s.client.Delete("/directory/users/custom-properties/" + url.PathEscape(id))
}

// ─── Group Read ───

func (s *DirectoryService) ListGroups(cursor string, count int) (*Response, error) {
	return s.client.Get("/groups" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) GetGroup(groupID string) (*Response, error) {
	return s.client.Get("/groups/" + url.PathEscape(groupID))
}

// ─── Task 4-5: Group CUD + Members + Admins + External Keys ───

func (s *DirectoryService) CreateGroup(body []byte) (*Response, error) {
	return s.client.Post("/groups", body)
}

func (s *DirectoryService) UpdateGroup(groupID string, body []byte) (*Response, error) {
	return s.client.Put("/groups/"+url.PathEscape(groupID), body)
}

func (s *DirectoryService) PatchGroup(groupID string, body []byte) (*Response, error) {
	return s.client.Patch("/groups/"+url.PathEscape(groupID), body)
}

func (s *DirectoryService) DeleteGroup(groupID string) (*Response, error) {
	return s.client.Delete("/groups/" + url.PathEscape(groupID))
}

func (s *DirectoryService) ListGroupMembers(groupID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/groups/%s/members", url.PathEscape(groupID)) + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) AddGroupMembers(groupID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/members", url.PathEscape(groupID)), body)
}

func (s *DirectoryService) RemoveGroupMember(groupID string, memberID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/groups/%s/members/%s", url.PathEscape(groupID), url.PathEscape(memberID)))
}

func (s *DirectoryService) ListGroupAdministrators(groupID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/groups/%s/administrators", url.PathEscape(groupID)) + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) AddGroupAdministrator(groupID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/administrators", url.PathEscape(groupID)), body)
}

func (s *DirectoryService) RemoveGroupAdministrator(groupID string, userID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/groups/%s/administrators/%s", url.PathEscape(groupID), url.PathEscape(userID)))
}

func (s *DirectoryService) UpsertGroupExternalKeys(body []byte) (*Response, error) {
	return s.client.Post("/groups/external-keys", body)
}

func (s *DirectoryService) ListGroupExternalKeys(cursor string, count int) (*Response, error) {
	return s.client.Get("/groups/external-keys" + BuildPaginationQuery(cursor, count))
}

// ─── OrgUnit Read ───

func (s *DirectoryService) ListOrgUnits(cursor string, count int) (*Response, error) {
	return s.client.Get("/orgunits" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) GetOrgUnit(orgUnitID string) (*Response, error) {
	return s.client.Get("/orgunits/" + url.PathEscape(orgUnitID))
}

// ─── Task 4-6: OrgUnit CUD + Members + AccessRestrict + External Keys ───

func (s *DirectoryService) CreateOrgUnit(body []byte) (*Response, error) {
	return s.client.Post("/orgunits", body)
}

func (s *DirectoryService) UpdateOrgUnit(orgUnitID string, body []byte) (*Response, error) {
	return s.client.Put("/orgunits/"+url.PathEscape(orgUnitID), body)
}

func (s *DirectoryService) PatchOrgUnit(orgUnitID string, body []byte) (*Response, error) {
	return s.client.Patch("/orgunits/"+url.PathEscape(orgUnitID), body)
}

func (s *DirectoryService) DeleteOrgUnit(orgUnitID string) (*Response, error) {
	return s.client.Delete("/orgunits/" + url.PathEscape(orgUnitID))
}

func (s *DirectoryService) MoveOrgUnit(orgUnitID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/orgunits/%s/move", url.PathEscape(orgUnitID)), body)
}

func (s *DirectoryService) ListOrgUnitMembers(orgUnitID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/orgunits/%s/members", url.PathEscape(orgUnitID)) + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) CreateOrgUnitAccessRestrict(orgUnitID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/orgunits/%s/orgunit-access-restrict", url.PathEscape(orgUnitID)), body)
}

func (s *DirectoryService) GetOrgUnitAccessRestrict(orgUnitID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/orgunits/%s/orgunit-access-restrict", url.PathEscape(orgUnitID)))
}

func (s *DirectoryService) UpdateOrgUnitAccessRestrict(orgUnitID string, body []byte) (*Response, error) {
	return s.client.Put(fmt.Sprintf("/orgunits/%s/orgunit-access-restrict", url.PathEscape(orgUnitID)), body)
}

func (s *DirectoryService) DeleteOrgUnitAccessRestrict(orgUnitID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/orgunits/%s/orgunit-access-restrict", url.PathEscape(orgUnitID)))
}

func (s *DirectoryService) UpsertOrgUnitExternalKeys(body []byte) (*Response, error) {
	return s.client.Post("/orgunits/external-keys", body)
}

func (s *DirectoryService) ListOrgUnitExternalKeys(cursor string, count int) (*Response, error) {
	return s.client.Get("/orgunits/external-keys" + BuildPaginationQuery(cursor, count))
}

// ─── Directory Metadata ───

func (s *DirectoryService) ListLevels(cursor string, count int) (*Response, error) {
	return s.client.Get("/directory/levels" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) ListPositions(cursor string, count int) (*Response, error) {
	return s.client.Get("/directory/positions" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) ListUserTypes(cursor string, count int) (*Response, error) {
	return s.client.Get("/directory/user-types" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) ListEmploymentTypes(cursor string, count int) (*Response, error) {
	return s.client.Get("/directory/employment-types" + BuildPaginationQuery(cursor, count))
}

// ─── Task 4-7: Positions CRUD + External Keys ───

func (s *DirectoryService) CreatePosition(body []byte) (*Response, error) {
	return s.client.Post("/directory/positions", body)
}

func (s *DirectoryService) GetPosition(positionID string) (*Response, error) {
	return s.client.Get("/directory/positions/" + url.PathEscape(positionID))
}

func (s *DirectoryService) UpdatePosition(positionID string, body []byte) (*Response, error) {
	return s.client.Put("/directory/positions/"+url.PathEscape(positionID), body)
}

func (s *DirectoryService) PatchPosition(positionID string, body []byte) (*Response, error) {
	return s.client.Patch("/directory/positions/"+url.PathEscape(positionID), body)
}

func (s *DirectoryService) DeletePosition(positionID string) (*Response, error) {
	return s.client.Delete("/directory/positions/" + url.PathEscape(positionID))
}

func (s *DirectoryService) EnablePositions() (*Response, error) {
	return s.client.Post("/directory/positions/enable", nil)
}

func (s *DirectoryService) DisablePositions() (*Response, error) {
	return s.client.Post("/directory/positions/disable", nil)
}

func (s *DirectoryService) UpsertPositionExternalKeys(body []byte) (*Response, error) {
	return s.client.Post("/directory/positions/external-keys", body)
}

func (s *DirectoryService) ListPositionExternalKeys() (*Response, error) {
	return s.client.Get("/directory/positions/external-keys")
}

// ─── Task 4-8: Levels CRUD + External Keys ───

func (s *DirectoryService) CreateLevel(body []byte) (*Response, error) {
	return s.client.Post("/directory/levels", body)
}

func (s *DirectoryService) GetLevel(levelID string) (*Response, error) {
	return s.client.Get("/directory/levels/" + url.PathEscape(levelID))
}

func (s *DirectoryService) UpdateLevel(levelID string, body []byte) (*Response, error) {
	return s.client.Put("/directory/levels/"+url.PathEscape(levelID), body)
}

func (s *DirectoryService) PatchLevel(levelID string, body []byte) (*Response, error) {
	return s.client.Patch("/directory/levels/"+url.PathEscape(levelID), body)
}

func (s *DirectoryService) DeleteLevel(levelID string) (*Response, error) {
	return s.client.Delete("/directory/levels/" + url.PathEscape(levelID))
}

func (s *DirectoryService) EnableLevels() (*Response, error) {
	return s.client.Post("/directory/levels/enable", nil)
}

func (s *DirectoryService) DisableLevels() (*Response, error) {
	return s.client.Post("/directory/levels/disable", nil)
}

func (s *DirectoryService) UpsertLevelExternalKeys(body []byte) (*Response, error) {
	return s.client.Post("/directory/levels/external-keys", body)
}

func (s *DirectoryService) ListLevelExternalKeys() (*Response, error) {
	return s.client.Get("/directory/levels/external-keys")
}

// ─── Task 4-9: Employment Types CRUD + External Keys + Access Restrict ───

func (s *DirectoryService) CreateEmploymentType(body []byte) (*Response, error) {
	return s.client.Post("/directory/employment-types", body)
}

func (s *DirectoryService) GetEmploymentType(id string) (*Response, error) {
	return s.client.Get("/directory/employment-types/" + url.PathEscape(id))
}

func (s *DirectoryService) UpdateEmploymentType(id string, body []byte) (*Response, error) {
	return s.client.Put("/directory/employment-types/"+url.PathEscape(id), body)
}

func (s *DirectoryService) PatchEmploymentType(id string, body []byte) (*Response, error) {
	return s.client.Patch("/directory/employment-types/"+url.PathEscape(id), body)
}

func (s *DirectoryService) DeleteEmploymentType(id string) (*Response, error) {
	return s.client.Delete("/directory/employment-types/" + url.PathEscape(id))
}

func (s *DirectoryService) EnableEmploymentTypes() (*Response, error) {
	return s.client.Post("/directory/employment-types/enable", nil)
}

func (s *DirectoryService) DisableEmploymentTypes() (*Response, error) {
	return s.client.Post("/directory/employment-types/disable", nil)
}

func (s *DirectoryService) UpsertEmploymentTypeExternalKeys(body []byte) (*Response, error) {
	return s.client.Post("/directory/employment-types/external-keys", body)
}

func (s *DirectoryService) ListEmploymentTypeExternalKeys() (*Response, error) {
	return s.client.Get("/directory/employment-types/external-keys")
}

func (s *DirectoryService) CreateEmploymentTypeAccessRestrict(id string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/directory/employment-types/%s/orgunit-access-restrict", url.PathEscape(id)), body)
}

func (s *DirectoryService) GetEmploymentTypeAccessRestrict(id string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/directory/employment-types/%s/orgunit-access-restrict", url.PathEscape(id)))
}

func (s *DirectoryService) UpdateEmploymentTypeAccessRestrict(id string, body []byte) (*Response, error) {
	return s.client.Put(fmt.Sprintf("/directory/employment-types/%s/orgunit-access-restrict", url.PathEscape(id)), body)
}

func (s *DirectoryService) DeleteEmploymentTypeAccessRestrict(id string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/directory/employment-types/%s/orgunit-access-restrict", url.PathEscape(id)))
}

// ─── Task 4-10: User Types CRUD + External Keys + Access Restrict ───

func (s *DirectoryService) CreateUserType(body []byte) (*Response, error) {
	return s.client.Post("/directory/user-types", body)
}

func (s *DirectoryService) GetUserType(id string) (*Response, error) {
	return s.client.Get("/directory/user-types/" + url.PathEscape(id))
}

func (s *DirectoryService) UpdateUserType(id string, body []byte) (*Response, error) {
	return s.client.Put("/directory/user-types/"+url.PathEscape(id), body)
}

func (s *DirectoryService) PatchUserType(id string, body []byte) (*Response, error) {
	return s.client.Patch("/directory/user-types/"+url.PathEscape(id), body)
}

func (s *DirectoryService) DeleteUserType(id string) (*Response, error) {
	return s.client.Delete("/directory/user-types/" + url.PathEscape(id))
}

func (s *DirectoryService) EnableUserTypes() (*Response, error) {
	return s.client.Post("/directory/user-types/enable", nil)
}

func (s *DirectoryService) DisableUserTypes() (*Response, error) {
	return s.client.Post("/directory/user-types/disable", nil)
}

func (s *DirectoryService) UpsertUserTypeExternalKeys(body []byte) (*Response, error) {
	return s.client.Post("/directory/user-types/external-keys", body)
}

func (s *DirectoryService) ListUserTypeExternalKeys() (*Response, error) {
	return s.client.Get("/directory/user-types/external-keys")
}

func (s *DirectoryService) CreateUserTypeAccessRestrict(id string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/directory/user-types/%s/orgunit-access-restrict", url.PathEscape(id)), body)
}

func (s *DirectoryService) GetUserTypeAccessRestrict(id string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/directory/user-types/%s/orgunit-access-restrict", url.PathEscape(id)))
}

func (s *DirectoryService) UpdateUserTypeAccessRestrict(id string, body []byte) (*Response, error) {
	return s.client.Put(fmt.Sprintf("/directory/user-types/%s/orgunit-access-restrict", url.PathEscape(id)), body)
}

func (s *DirectoryService) DeleteUserTypeAccessRestrict(id string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/directory/user-types/%s/orgunit-access-restrict", url.PathEscape(id)))
}

// ─── Task 4-11: Profile Statuses CRUD ───

func (s *DirectoryService) CreateDirectoryProfileStatus(body []byte) (*Response, error) {
	return s.client.Post("/directory/profile-statuses", body)
}

func (s *DirectoryService) ListDirectoryProfileStatuses(cursor string, count int) (*Response, error) {
	return s.client.Get("/directory/profile-statuses" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) GetDirectoryProfileStatus(id string) (*Response, error) {
	return s.client.Get("/directory/profile-statuses/" + url.PathEscape(id))
}

func (s *DirectoryService) UpdateDirectoryProfileStatus(id string, body []byte) (*Response, error) {
	return s.client.Put("/directory/profile-statuses/"+url.PathEscape(id), body)
}

func (s *DirectoryService) PatchDirectoryProfileStatus(id string, body []byte) (*Response, error) {
	return s.client.Patch("/directory/profile-statuses/"+url.PathEscape(id), body)
}

func (s *DirectoryService) DeleteDirectoryProfileStatus(id string) (*Response, error) {
	return s.client.Delete("/directory/profile-statuses/" + url.PathEscape(id))
}

func (s *DirectoryService) EnableDirectoryProfileStatuses() (*Response, error) {
	return s.client.Post("/directory/profile-statuses/enable", nil)
}

func (s *DirectoryService) DisableDirectoryProfileStatuses() (*Response, error) {
	return s.client.Post("/directory/profile-statuses/disable", nil)
}

// ─── Task 4-12: Custom Fields CRUD ───

func (s *DirectoryService) CreateCustomField(body []byte) (*Response, error) {
	return s.client.Post("/directory/custom-fields", body)
}

func (s *DirectoryService) ListCustomFields(cursor string, count int) (*Response, error) {
	return s.client.Get("/directory/custom-fields" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) GetCustomField(id string) (*Response, error) {
	return s.client.Get("/directory/custom-fields/" + url.PathEscape(id))
}

func (s *DirectoryService) PatchCustomField(id string, body []byte) (*Response, error) {
	return s.client.Patch("/directory/custom-fields/"+url.PathEscape(id), body)
}

func (s *DirectoryService) DeleteCustomField(id string) (*Response, error) {
	return s.client.Delete("/directory/custom-fields/" + url.PathEscape(id))
}
