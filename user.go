package protoslib

import (
	"bytes"
	"encoding/json"
	"net/http"

	auth "github.com/protosio/protos/auth"
)

// AuthUser authenticates a user and returns information about it
func (p Protos) AuthUser(username string, password string) (auth.UserInfo, error) {
	userInfo := auth.UserInfo{}
	url := p.createURL("user/auth")
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

// GetAdminUser retrieves the admin user of the Protos instance
func (p Protos) GetAdminUser() (string, error) {
	req, err := http.NewRequest("GET", p.createURL("info/adminuser"), nil)
	if err != nil {
		return "", err
	}
	user := struct{ Username string }{}

	payload, err := p.makeRequest(req)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(payload, &user)
	if err != nil {
		return "", err
	}

	return user.Username, nil
}
