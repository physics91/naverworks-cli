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
