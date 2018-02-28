package protoslib

import (
	"errors"
	"os"
)

const (
	// EnvVarAppID is the name of the environment variable that holds the unique application ID that is used
	// identifying every request.
	EnvVarAppID = "APPID"
)

// GetAppID retrieves the app ID from an evironment variable
func GetAppID() (string, error) {
	appID := os.Getenv(EnvVarAppID)
	if appID == "" {
		return "", errors.New("APPID environment variable is not set")
	}
	return appID, nil
}
