package protoslib

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

const (
	// EnvVarAppID is the name of the environment variable that holds the unique application ID that is used
	// identifying every request.
	EnvVarAppID = "APPID"
)

// AppInfo contains information about the requesting app
type AppInfo struct {
	Name string
}

// GetAppID retrieves the app ID from an evironment variable
func GetAppID() (string, error) {
	appID := os.Getenv(EnvVarAppID)
	if appID == "" {
		return "", errors.New("APPID environment variable is not set")
	}
	return appID, nil
}

// GetDomain retrieves the domain name of the Protos instance
func (p Protos) GetDomain() (string, error) {
	req, err := http.NewRequest("GET", p.URL+"info/domain", nil)
	if err != nil {
		return "", err
	}
	domain := struct{ Domain string }{Domain: ""}

	payload, err := p.makeRequest(req)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(payload, &domain)
	if err != nil {
		return "", err
	}

	return domain.Domain, nil
}

// GetAppInfo retrieves information about the requesting application, from the internal Protos API
func (p Protos) GetAppInfo() (AppInfo, error) {
	appInfo := AppInfo{}
	req, err := http.NewRequest("GET", p.URL+"info/app", nil)
	if err != nil {
		return appInfo, err
	}
	payload, err := p.makeRequest(req)
	if err != nil {
		return appInfo, err
	}

	err = json.Unmarshal(payload, &appInfo)
	if err != nil {
		return appInfo, err
	}
	return appInfo, nil
}
