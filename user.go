package protoslib

import (
	"bytes"
	"encoding/json"
	"net/http"

	auth "github.com/nustiueudinastea/protos/auth"
)

// AuthUser authenticates a user and returns information about it
func (p Protos) AuthUser(username string, password string) (auth.UserInfo, error) {
	userInfo := auth.UserInfo{}
	url := p.URL + "internal/user/auth"
	login := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: username,
		Password: password,
	}
	loginJSON, err := json.Marshal(login)
	if err != nil {
		return userInfo, err
	}

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(loginJSON))
	if err != nil {
		return userInfo, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := p.makeRequest(req)
	if err != nil {
		return userInfo, err
	}

	err = json.Unmarshal(res, &userInfo)
	if err != nil {
		return userInfo, err
	}
	return userInfo, nil
}
