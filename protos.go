package protoslib

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	resource "github.com/protosio/protos/resource"
)

var protosURL string

type httpErr struct {
	Error string `json:"error"`
}

// Protos client struct
type Protos struct {
	Host       string
	PathPrefix string
	AppID      string
	HTTPclient *http.Client
	Protocol   string
}

// Resources is a dictionary that stores resources, with the key being the resource id
type Resources map[string]*resource.Resource

func decodeError(resp *http.Response) (string, error) {
	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	httperr := httpErr{}
	err = json.Unmarshal(payload, &httperr)
	if err != nil {
		return "", fmt.Errorf("Failed to decode error message from Protos: %s", err.Error())
	}
	return httperr.Error, nil
}

// UnmarshalJSON is a custom json decode for resources
func (rscs Resources) UnmarshalJSON(b []byte) error {
	var resources map[string]json.RawMessage
	err := json.Unmarshal(b, &resources)
	if err != nil {
		return err
	}
	for key, value := range resources {
		rsc := resource.Resource{}
		err := json.Unmarshal(value, &rsc)
		if err != nil {
			return err
		}
		rscs[key] = &rsc
	}
	return nil
}

// makeRequest prepares and sends a request to the protos backend
func (p Protos) makeRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("Appid", p.AppID)

	resp, err := p.HTTPclient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	if resp.StatusCode != 200 {
		httperr := httpErr{}
		err := json.Unmarshal(payload, &httperr)
		if err != nil {
			return []byte{}, fmt.Errorf("Failed to decode error message from Protos: %s", err.Error())
		}
		return []byte{}, errors.New(httperr.Error)
	}

	return payload, nil
}

// NewClient returns a client that can be used to interact with Protos
func NewClient(host string, appid string) Protos {
	return Protos{
		Host:       host,
		PathPrefix: "api/v1/i",
		AppID:      appid,
		HTTPclient: &http.Client{},
		Protocol:   "http",
	}
}
