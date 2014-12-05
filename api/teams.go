package api

import (
	"encoding/json"
	"net/http"

	. "github.com/albertoleal/backstage/account"
	"github.com/albertoleal/backstage/errors"
	"github.com/zenazn/goji/web"
)

type TeamsHandler struct {
	ApiHandler
}

func (handler *TeamsHandler) CreateTeam(c *web.C, w http.ResponseWriter, r *http.Request) *HTTPResponse {
	var response *HTTPResponse
	owner, err := handler.getCurrentUser(c)
	if err != nil {
		response = &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: err.Error()}
		AddRequestError(c, response)
		return response
	}

	team := &Team{}
	err = handler.parseBody(r.Body, team)
	if err != nil {
		response = &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: "The request was bad-formed."}
		AddRequestError(c, response)
		return response
	}

	err = team.Save(owner)
	if err != nil {
		e := err.(*errors.ValidationError)
		erro := &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: e.Message}
		AddRequestError(c, erro)
		return erro
	}
	team, err = FindTeamByName(team.Name)
	if err != nil {
		e := err.(*errors.ValidationError)
		erro := &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: e.Message}
		AddRequestError(c, erro)
		return erro
	}
	payload, _ := json.Marshal(team)
	response = &HTTPResponse{StatusCode: http.StatusCreated, Payload: string(payload)}
	return response
}

func (handler *TeamsHandler) DeleteTeam(c *web.C, w http.ResponseWriter, r *http.Request) *HTTPResponse {
	var response *HTTPResponse
	currentUser, err := handler.getCurrentUser(c)
	if err != nil {
		response = &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: err.Error()}
		AddRequestError(c, response)
		return response
	}
	team, err := FindTeamById(c.URLParams["id"])
	if err != nil || team.Owner != currentUser.Username {
		response = &HTTPResponse{StatusCode: http.StatusForbidden, Payload: "Team not found or you're not the owner."}
		AddRequestError(c, response)
		return response
	}
	err = team.Delete()
	if err != nil {
		response = &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: "It was not possible to delete your team."}
		AddRequestError(c, response)
		return response
	}
	team.Id = ""
	payload, _ := json.Marshal(team)
	response = &HTTPResponse{StatusCode: http.StatusOK, Payload: string(payload)}
	return response
}

func (handler *TeamsHandler) GetUserTeams(c *web.C, w http.ResponseWriter, r *http.Request) *HTTPResponse {
	var response *HTTPResponse
	currentUser, err := handler.getCurrentUser(c)
	if err != nil {
		response = &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: err.Error()}
		AddRequestError(c, response)
		return response
	}
	teams, _ := currentUser.GetTeams()
	payload, _ := json.Marshal(teams)
	response = &HTTPResponse{StatusCode: http.StatusOK, Payload: string(payload)}
	return response
}

func (handler *TeamsHandler) GetTeamInfo(c *web.C, w http.ResponseWriter, r *http.Request) *HTTPResponse {
	var response *HTTPResponse
	currentUser, err := handler.getCurrentUser(c)
	if err != nil {
		response = &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: err.Error()}
		AddRequestError(c, response)
		return response
	}
	team, err := FindTeamById(c.URLParams["id"])
	if err != nil {
		erro := &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: "Team not found."}
		AddRequestError(c, erro)
		return erro
	}
	_, ok := team.ContainsUser(currentUser)
	if !ok {
		erro := &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: "You do not belong to this team!"}
		AddRequestError(c, erro)
		return erro
	}
	result, _ := json.Marshal(team)
	response = &HTTPResponse{StatusCode: http.StatusOK, Payload: string(result)}
	return response
}

func (handler *TeamsHandler) AddUsersToTeam(c *web.C, w http.ResponseWriter, r *http.Request) *HTTPResponse {
	var response *HTTPResponse
	currentUser, err := handler.getCurrentUser(c)
	if err != nil {
		response = &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: err.Error()}
		AddRequestError(c, response)
		return response
	}
	team, err := FindTeamById(c.URLParams["id"])
	if err != nil {
		erro := &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: "Team not found."}
		AddRequestError(c, erro)
		return erro
	}

	_, ok := team.ContainsUser(currentUser)
	if !ok {
		erro := &HTTPResponse{StatusCode: http.StatusForbidden, Payload: "You do not belong to this team!"}
		AddRequestError(c, erro)
		return erro
	}

	var keys map[string]interface{}
	err = handler.parseBody(r.Body, &keys)
	if err != nil {
		response = &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: err.Error()}
		AddRequestError(c, response)
		return response
	}
	if keys["users"] == nil {
		erro := &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: "The request was bad-formed."}
		AddRequestError(c, erro)
		return erro
	}

	var users []string
	for _, v := range keys["users"].([]interface{}) {
		switch v.(type) {
		case string:
			user := v.(string)
			users = append(users, user)
		}
	}
	team.AddUsers(users)
	result, _ := json.Marshal(team)
	response = &HTTPResponse{StatusCode: http.StatusCreated, Payload: string(result)}
	return response
}

func (handler *TeamsHandler) RemoveUsersFromTeam(c *web.C, w http.ResponseWriter, r *http.Request) *HTTPResponse {
	var response *HTTPResponse
	currentUser, err := handler.getCurrentUser(c)
	if err != nil {
		response = &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: err.Error()}
		AddRequestError(c, response)
		return response
	}
	team, err := FindTeamById(c.URLParams["id"])
	if err != nil {
		erro := &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: "Team not found."}
		AddRequestError(c, erro)
		return erro
	}

	_, ok := team.ContainsUser(currentUser)
	if !ok {
		erro := &HTTPResponse{StatusCode: http.StatusForbidden, Payload: "You do not belong to this team!"}
		AddRequestError(c, erro)
		return erro
	}

	var keys map[string]interface{}
	err = handler.parseBody(r.Body, &keys)
	if err != nil {
		response = &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: err.Error()}
		AddRequestError(c, response)
		return response
	}
	if keys["users"] == nil {
		erro := &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: "The request was bad-formed."}
		AddRequestError(c, erro)
		return erro
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
		erro := &HTTPResponse{StatusCode: http.StatusBadRequest, Payload: err.Error()}
		AddRequestError(c, erro)
		return erro
	}
	result, _ := json.Marshal(team)
	response = &HTTPResponse{StatusCode: http.StatusOK, Payload: string(result)}
	return response
}