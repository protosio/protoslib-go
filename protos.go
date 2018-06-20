package protoslib

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	resource "github.com/nustiueudinastea/protos/resource"
)

var protosURL string

// Protos client struct
type Protos struct {
	URL        string
	AppID      string
	HTTPclient *http.Client
}

// Resources is a dictionary that stores resources, with the key being the resource id
type Resources map[string]*resource.Resource

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
		return []byte{}, errors.New(string(payload))
	}

	return payload, nil
}

// NewClient returns a client that can be used to interact with Protos
func NewClient(url string, appid string) Protos {
	return Protos{
		URL:        url + "/api/v1/",
		AppID:      appid,
		HTTPclient: &http.Client{},
	}
}
