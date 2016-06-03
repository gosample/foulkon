package http

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/tecsisa/authorizr/api"
)

// Requests

type CreateGroupRequest struct {
	Name string `json:"Name, omitempty"`
	Path string `json:"Path, omitempty"`
}

type UpdateGroupRequest struct {
	Name string `json:"Name, omitempty"`
	Path string `json:"Path, omitempty"`
}

// Responses

type CreateGroupResponse struct {
	Group *api.Group
}

type UpdateGroupResponse struct {
	Group *api.Group
}

type GetGroupNameResponse struct {
	Group *api.Group
}

type GetGroupsResponse struct {
	Groups []api.GroupIdentity
}

type GetGroupMembersResponse struct {
	Members []string
}

type GetGroupPoliciesResponse struct {
	AttachedPolicies []api.PolicyIdentity
}

func (a *AuthHandler) handleCreateGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := a.core.Authenticator.RetrieveUserID(*r)
	// Decode request
	request := CreateGroupRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		a.core.Logger.Errorln(err)
		a.RespondBadRequest(r, &userID, w, &api.Error{Code: api.INVALID_PARAMETER_ERROR, Message: err.Error()})
		return
	}

	org := ps.ByName(ORG_NAME)
	// Call group API to create an group
	result, err := a.core.AuthApi.AddGroup(a.core.Authenticator.RetrieveUserID(*r), org, request.Name, request.Path)

	// Error handling
	if err != nil {
		a.core.Logger.Errorln(err)
		// Transform to API errors
		apiError := err.(*api.Error)
		switch apiError.Code {
		case api.GROUP_ALREADY_EXIST:
			a.RespondConflict(r, &userID, w, apiError)
		case api.INVALID_PARAMETER_ERROR:
			a.RespondBadRequest(r, &userID, w, apiError)
		case api.UNAUTHORIZED_RESOURCES_ERROR:
			a.RespondForbidden(r, &userID, w, apiError)
		default: // Unexpected API error
			a.RespondInternalServerError(r, &userID, w)
		}
		return
	}

	response := &CreateGroupResponse{
		Group: result,
	}

	// Write group to response
	a.RespondCreated(r, &userID, w, response)
}

func (a *AuthHandler) handleDeleteGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := a.core.Authenticator.RetrieveUserID(*r)
	// Retrieve group org and name from path
	org := ps.ByName(ORG_NAME)
	name := ps.ByName(GROUP_NAME)

	// Call user API to delete group
	err := a.core.AuthApi.RemoveGroup(a.core.Authenticator.RetrieveUserID(*r), org, name)

	// Check if there were errors
	if err != nil {
		a.core.Logger.Errorln(err)
		// Transform to API errors
		apiError := err.(*api.Error)
		switch apiError.Code {
		case api.GROUP_BY_ORG_AND_NAME_NOT_FOUND:
			a.RespondNotFound(r, &userID, w, apiError)
		case api.UNAUTHORIZED_RESOURCES_ERROR:
			a.RespondForbidden(r, &userID, w, apiError)
		default: // Unexpected API error
			a.RespondInternalServerError(r, &userID, w)
		}
		return
	} else { // a.Respond without content
		a.RespondNoContent(r, &userID, w)
	}
}

func (a *AuthHandler) handleGetGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := a.core.Authenticator.RetrieveUserID(*r)
	// Retrieve group org and name from path
	org := ps.ByName(ORG_NAME)
	name := ps.ByName(GROUP_NAME)

	// Call group API to retrieve group
	result, err := a.core.AuthApi.GetGroupByName(a.core.Authenticator.RetrieveUserID(*r), org, name)

	// Error handling
	if err != nil {
		a.core.Logger.Errorln(err)
		// Transform to API errors
		apiError := err.(*api.Error)
		switch apiError.Code {
		case api.GROUP_BY_ORG_AND_NAME_NOT_FOUND:
			a.RespondNotFound(r, &userID, w, apiError)
		case api.UNAUTHORIZED_RESOURCES_ERROR:
			a.RespondForbidden(r, &userID, w, apiError)
		default: // Unexpected API error
			a.RespondInternalServerError(r, &userID, w)
		}
		return
	}

	response := GetGroupNameResponse{
		Group: result,
	}

	// Write group to response
	a.RespondOk(r, &userID, w, response)
}

func (a *AuthHandler) handleListGroups(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := a.core.Authenticator.RetrieveUserID(*r)
	// Retrieve group org from path
	org := ps.ByName(ORG_NAME)

	// Retrieve query param if exist
	pathPrefix := r.URL.Query().Get("PathPrefix")

	// Call group API to retrieve groups
	result, err := a.core.AuthApi.GetListGroups(a.core.Authenticator.RetrieveUserID(*r), org, pathPrefix)
	if err != nil {
		a.core.Logger.Errorln(err)
		// Transform to API errors
		apiError := err.(*api.Error)
		switch apiError.Code {
		case api.UNAUTHORIZED_RESOURCES_ERROR:
			a.RespondForbidden(r, &userID, w, apiError)
		default: // Unexpected API error
			a.RespondInternalServerError(r, &userID, w)
		}
		return
	}

	// Create response
	response := &GetGroupsResponse{
		Groups: result,
	}

	// Return data
	a.RespondOk(r, &userID, w, response)

}

func (a *AuthHandler) handleUpdateGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := a.core.Authenticator.RetrieveUserID(*r)
	// Decode request
	request := UpdateGroupRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		a.core.Logger.Errorln(err)
		a.RespondBadRequest(r, &userID, w, &api.Error{Code: api.INVALID_PARAMETER_ERROR, Message: err.Error()})
		return
	}

	// Retrieve group, org from path
	org := ps.ByName(ORG_NAME)
	groupName := ps.ByName(GROUP_NAME)

	// Call group API to update group
	result, err := a.core.AuthApi.UpdateGroup(a.core.Authenticator.RetrieveUserID(*r), org, groupName, request.Name, request.Path)

	// Check errors
	if err != nil {
		a.core.Logger.Errorln(err)
		// Transform to API errors
		apiError := err.(*api.Error)
		switch apiError.Code {
		case api.GROUP_BY_ORG_AND_NAME_NOT_FOUND:
			a.RespondNotFound(r, &userID, w, apiError)
		case api.GROUP_ALREADY_EXIST:
			a.RespondConflict(r, &userID, w, apiError)
		case api.UNAUTHORIZED_RESOURCES_ERROR:
			a.RespondForbidden(r, &userID, w, apiError)
		case api.INVALID_PARAMETER_ERROR:
			a.RespondBadRequest(r, &userID, w, apiError)
		default:
			a.RespondInternalServerError(r, &userID, w)
		}
		return
	}

	// Create response
	response := &UpdateGroupResponse{
		Group: result,
	}

	// Write group to response
	a.RespondOk(r, &userID, w, response)
}

func (a *AuthHandler) handleListMembers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := a.core.Authenticator.RetrieveUserID(*r)
	// Retrieve group, org
	org := ps.ByName(ORG_NAME)
	group := ps.ByName(GROUP_NAME)

	// Call group API to list members
	result, err := a.core.AuthApi.ListMembers(a.core.Authenticator.RetrieveUserID(*r), org, group)

	// Check errors
	if err != nil {
		a.core.Logger.Errorln(err)
		// Transform to API errors
		apiError := err.(*api.Error)
		switch apiError.Code {
		case api.GROUP_BY_ORG_AND_NAME_NOT_FOUND:
			a.RespondNotFound(r, &userID, w, apiError)
		case api.UNAUTHORIZED_RESOURCES_ERROR:
			a.RespondForbidden(r, &userID, w, apiError)
		default:
			a.RespondInternalServerError(r, &userID, w)
		}
		return
	}

	// Create response
	response := &GetGroupMembersResponse{
		Members: result,
	}

	// Write GroupMembers to response
	a.RespondOk(r, &userID, w, response)

}

func (a *AuthHandler) handleAddMember(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := a.core.Authenticator.RetrieveUserID(*r)
	// Retrieve group, org and user from path
	org := ps.ByName(ORG_NAME)
	user := ps.ByName(USER_ID)
	group := ps.ByName(GROUP_NAME)

	// Call group API to create an group
	err := a.core.AuthApi.AddMember(a.core.Authenticator.RetrieveUserID(*r), user, group, org)
	// Error handling
	if err != nil {
		a.core.Logger.Errorln(err)
		// Transform to API errors
		apiError := err.(*api.Error)
		switch apiError.Code {
		case api.GROUP_BY_ORG_AND_NAME_NOT_FOUND, api.USER_BY_EXTERNAL_ID_NOT_FOUND:
			a.RespondNotFound(r, &userID, w, apiError)
		case api.UNAUTHORIZED_RESOURCES_ERROR:
			a.RespondForbidden(r, &userID, w, apiError)
		case api.USER_IS_ALREADY_A_MEMBER_OF_GROUP:
			a.RespondConflict(r, &userID, w, apiError)
		default:
			a.RespondInternalServerError(r, &userID, w)
		}
		return
	} else { // a.Respond without content
		a.RespondNoContent(r, &userID, w)
	}
}

func (a *AuthHandler) handleRemoveMember(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := a.core.Authenticator.RetrieveUserID(*r)
	// Retrieve group, org and user from path
	org := ps.ByName(ORG_NAME)
	user := ps.ByName(USER_ID)
	group := ps.ByName(GROUP_NAME)

	// Call group API to create an group
	err := a.core.AuthApi.RemoveMember(a.core.Authenticator.RetrieveUserID(*r), user, group, org)
	// Error handling
	if err != nil {
		a.core.Logger.Errorln(err)
		// Transform to API errors
		apiError := err.(*api.Error)
		switch apiError.Code {
		case api.GROUP_BY_ORG_AND_NAME_NOT_FOUND, api.USER_BY_EXTERNAL_ID_NOT_FOUND, api.USER_IS_NOT_A_MEMBER_OF_GROUP:
			a.RespondNotFound(r, &userID, w, apiError)
		case api.UNAUTHORIZED_RESOURCES_ERROR:
			a.RespondForbidden(r, &userID, w, apiError)
		default:
			a.RespondInternalServerError(r, &userID, w)
		}
		return
	} else { // a.Respond without content
		a.RespondNoContent(r, &userID, w)
	}

}

func (a *AuthHandler) handleAttachGroupPolicy(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := a.core.Authenticator.RetrieveUserID(*r)
	// Retrieve group, org and policy from path
	org := ps.ByName(ORG_NAME)
	groupName := ps.ByName(GROUP_NAME)
	policyName := ps.ByName(POLICY_NAME)

	// Call group API to attach policy to group
	err := a.core.AuthApi.AttachPolicyToGroup(a.core.Authenticator.RetrieveUserID(*r), org, groupName, policyName)

	// Error handling
	if err != nil {
		a.core.Logger.Errorln(err)
		// Transform to API errors
		apiError := err.(*api.Error)
		switch apiError.Code {
		case api.GROUP_BY_ORG_AND_NAME_NOT_FOUND, api.POLICY_BY_ORG_AND_NAME_NOT_FOUND:
			a.RespondNotFound(r, &userID, w, apiError)
		case api.UNAUTHORIZED_RESOURCES_ERROR:
			a.RespondForbidden(r, &userID, w, apiError)
		case api.POLICY_IS_ALREADY_ATTACHED_TO_GROUP:
			a.RespondConflict(r, &userID, w, apiError)
		default: // Unexpected API error
			a.RespondInternalServerError(r, &userID, w)
		}
		return

	} else { // a.Respond without content
		a.RespondNoContent(r, &userID, w)
	}

}

func (a *AuthHandler) handleDetachGroupPolicy(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := a.core.Authenticator.RetrieveUserID(*r)
	// Retrieve group, org and policy from path
	org := ps.ByName(ORG_NAME)
	groupName := ps.ByName(GROUP_NAME)
	policyName := ps.ByName(POLICY_NAME)

	// Call group API to detach policy to group
	err := a.core.AuthApi.DetachPolicyToGroup(a.core.Authenticator.RetrieveUserID(*r), org, groupName, policyName)

	// Error handling
	if err != nil {
		a.core.Logger.Errorln(err)
		// Transform to API errors
		apiError := err.(*api.Error)
		switch apiError.Code {
		case api.GROUP_BY_ORG_AND_NAME_NOT_FOUND, api.POLICY_BY_ORG_AND_NAME_NOT_FOUND, api.POLICY_IS_NOT_ATTACHED_TO_GROUP:
			a.RespondNotFound(r, &userID, w, apiError)
		case api.UNAUTHORIZED_RESOURCES_ERROR:
			a.RespondForbidden(r, &userID, w, apiError)
		default: // Unexpected API error
			a.RespondInternalServerError(r, &userID, w)
		}
		return

	} else { // a.Respond without content
		a.RespondNoContent(r, &userID, w)
	}
}

func (a *AuthHandler) handleListAttachedGroupPolicies(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := a.core.Authenticator.RetrieveUserID(*r)
	// Retrieve group, org from path
	org := ps.ByName(ORG_NAME)
	groupName := ps.ByName(GROUP_NAME)

	// Call group API to retrieve attached policies
	result, err := a.core.AuthApi.ListAttachedGroupPolicies(a.core.Authenticator.RetrieveUserID(*r), org, groupName)
	if err != nil {
		a.core.Logger.Errorln(err)
		// Transform to API errors
		apiError := err.(*api.Error)
		switch apiError.Code {
		case api.GROUP_BY_ORG_AND_NAME_NOT_FOUND:
			a.RespondNotFound(r, &userID, w, apiError)
		case api.UNAUTHORIZED_RESOURCES_ERROR:
			a.RespondForbidden(r, &userID, w, apiError)
		default:
			a.RespondInternalServerError(r, &userID, w)
		}
		return
	}

	// Create response
	response := &GetGroupPoliciesResponse{
		AttachedPolicies: result,
	}

	// Return data
	a.RespondOk(r, &userID, w, response)

}

func (a *AuthHandler) handleListAllGroups(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := a.core.Authenticator.RetrieveUserID(*r)
	// get Org and PathPrefix from request, so the query can be filtered
	org := r.URL.Query().Get("Org")
	pathPrefix := r.URL.Query().Get("PathPrefix")

	// Call group API to retrieve groups
	result, err := a.core.AuthApi.GetListGroups(a.core.Authenticator.RetrieveUserID(*r), org, pathPrefix)
	if err != nil {
		a.core.Logger.Errorln(err)
		// Transform to API errors
		apiError := err.(*api.Error)
		switch apiError.Code {
		case api.UNAUTHORIZED_RESOURCES_ERROR:
			a.RespondForbidden(r, &userID, w, apiError)
		default:
			a.RespondInternalServerError(r, &userID, w)
		}
		return
	}

	// Create response
	response := &GetGroupsResponse{
		Groups: result,
	}

	// Return data
	a.RespondOk(r, &userID, w, response)

}
