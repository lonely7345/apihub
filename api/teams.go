package api

import (
	"encoding/json"
	"net/http"

	. "github.com/backstage/backstage/account"
	. "github.com/backstage/backstage/errors"
	"github.com/zenazn/goji/web"
)

type TeamsHandler struct {
	ApiHandler
}

func (handler *TeamsHandler) CreateTeam(c *web.C, w http.ResponseWriter, r *http.Request) *HTTPResponse {
	currentUser, err := handler.getCurrentUser(c)
	if err != nil {
		return ResponseError(c, http.StatusBadRequest, err.Error())
	}

	team := &Team{}
	err = handler.parseBody(r.Body, team)
	if err != nil {
		return ResponseError(c, http.StatusBadRequest, "The request was bad-formed.")
	}

	err = team.Save(currentUser)
	if err != nil {
		return ResponseError(c, http.StatusBadRequest, err.Error())
	}
	team, err = FindTeamByName(team.Name)
	if err != nil {
		return ResponseError(c, http.StatusBadRequest, err.Error())
	}
	payload, _ := json.Marshal(team)
	return &HTTPResponse{StatusCode: http.StatusCreated, Payload: string(payload)}
}

func (handler *TeamsHandler) DeleteTeam(c *web.C, w http.ResponseWriter, r *http.Request) *HTTPResponse {
	currentUser, err := handler.getCurrentUser(c)
	if err != nil {
		return ResponseError(c, http.StatusBadRequest, err.Error())
	}

	team, err := FindTeamByAlias(c.URLParams["alias"])
	if err != nil || team.Owner != currentUser.Email {
		return ResponseError(c, http.StatusForbidden, ErrOnlyOwnerHasPermission.Error())
	}
	err = team.Delete()
	if err != nil {
		return ResponseError(c, http.StatusBadRequest, err.Error())
	}

	team.Id = ""
	payload, _ := json.Marshal(team)
	return &HTTPResponse{StatusCode: http.StatusOK, Payload: string(payload)}
}

func (handler *TeamsHandler) GetUserTeams(c *web.C, w http.ResponseWriter, r *http.Request) *HTTPResponse {
	currentUser, err := handler.getCurrentUser(c)
	if err != nil {
		return ResponseError(c, http.StatusBadRequest, err.Error())
	}

	teams, _ := currentUser.GetTeams()
	payload, _ := json.Marshal(teams)
	return &HTTPResponse{StatusCode: http.StatusOK, Payload: string(payload)}
}

func (handler *TeamsHandler) GetTeamInfo(c *web.C, w http.ResponseWriter, r *http.Request) *HTTPResponse {
	currentUser, err := handler.getCurrentUser(c)
	if err != nil {
		return ResponseError(c, http.StatusBadRequest, err.Error())
	}

	team, err := FindTeamByAlias(c.URLParams["alias"])
	if err != nil {
		return ResponseError(c, http.StatusBadRequest, err.Error())
	}
	_, err = team.ContainsUser(currentUser)
	if err != nil {
		return ResponseError(c, http.StatusForbidden, err.Error())
	}

	result, _ := json.Marshal(team)
	return &HTTPResponse{StatusCode: http.StatusOK, Payload: string(result)}
}

func (handler *TeamsHandler) AddUsersToTeam(c *web.C, w http.ResponseWriter, r *http.Request) *HTTPResponse {
	currentUser, err := handler.getCurrentUser(c)
	if err != nil {
		return ResponseError(c, http.StatusBadRequest, err.Error())
	}

	team, err := FindTeamByAlias(c.URLParams["alias"])
	if err != nil {
		return ResponseError(c, http.StatusBadRequest, "Team not found.")
	}
	_, err = team.ContainsUser(currentUser)
	if err != nil {
		return ResponseError(c, http.StatusForbidden, err.Error())
	}

	var keys map[string]interface{}
	err = handler.parseBody(r.Body, &keys)
	if err != nil {
		return ResponseError(c, http.StatusBadRequest, err.Error())
	}
	if keys["users"] == nil {
		return ResponseError(c, http.StatusBadRequest, "The request was bad-formed.")
	}
	var users []string
	for _, v := range keys["users"].([]interface{}) {
		switch v.(type) {
		case string:
			user := v.(string)
			users = append(users, user)
		}
	}

	err = team.AddUsers(users)
	if err != nil {
		return ResponseError(c, http.StatusBadRequest, err.Error())
	}
	result, _ := json.Marshal(team)
	return &HTTPResponse{StatusCode: http.StatusCreated, Payload: string(result)}
}

func (handler *TeamsHandler) RemoveUsersFromTeam(c *web.C, w http.ResponseWriter, r *http.Request) *HTTPResponse {
	currentUser, err := handler.getCurrentUser(c)
	if err != nil {
		return ResponseError(c, http.StatusBadRequest, err.Error())
	}

	team, err := FindTeamByAlias(c.URLParams["alias"])
	if err != nil {
		return ResponseError(c, http.StatusBadRequest, "Team not found.")
	}

	_, err = team.ContainsUser(currentUser)
	if err != nil {
		return ResponseError(c, http.StatusForbidden, err.Error())
	}

	var keys map[string]interface{}
	err = handler.parseBody(r.Body, &keys)
	if err != nil {
		return ResponseError(c, http.StatusBadRequest, err.Error())
	}
	if keys["users"] == nil {
		return ResponseError(c, http.StatusBadRequest, "The request was bad-formed.")
	}
	var users []string
	for _, v := range keys["users"].([]interface{}) {
		switch v.(type) {
		case string:
			user := v.(string)
			users = append(users, user)
		}
	}

	err = team.RemoveUsers(users)
	if err != nil {
		return ResponseError(c, http.StatusForbidden, err.Error())
	}
	result, _ := json.Marshal(team)
	return &HTTPResponse{StatusCode: http.StatusOK, Payload: string(result)}
}
