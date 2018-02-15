package protos

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

// GetDomain retrieves the domain name of the Protos instance
func (p Protos) GetDomain() (string, error) {
	resourcesReq, err := http.NewRequest("GET", p.URL+"internal/info/domain", nil)
	if err != nil {
		return "", err
	}
	domain := struct{ Domain string }{Domain: ""}

	payload, err := p.makeRequest(resourcesReq)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(payload, &domain)
	if err != nil {
		return "", err
	}

	return domain.Domain, nil
}

// UpdateResourceValue updates a resource's value based on the ID and new value provided
func (p Protos) UpdateResourceValue(resourceID string, newValue resource.Type) error {
	payloadJSON, err := json.Marshal(newValue)
	if err != nil {
		return err
	}

	url := p.URL + "internal/resource/" + resourceID
	req, err := http.NewRequest("UPDATE", url, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	_, err = p.makeRequest(req)
	if err != nil {
		return err
	}

	return nil
}

// SetResourceStatus takes a resource ID and sets a new status
func (p Protos) SetResourceStatus(resourceID string, rstatus string) error {

	statusJSON, err := json.Marshal(&struct {
		Status string `json:"status"`
	}{
		Status: rstatus,
	})
	if err != nil {
		return err
	}

	url := p.URL + "internal/resource/" + resourceID
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(statusJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	_, err = p.makeRequest(req)
	if err != nil {
		return err
	}

	return nil
}

// SetStatusBatch takes a list of Resource and applies the same status to all of them
func (p Protos) SetStatusBatch(resources map[string]*resource.Resource, rstatus string) error {
	for _, resource := range resources {
		err := p.SetResourceStatus(resource.ID, rstatus)
		if err != nil {
			return fmt.Errorf("Could not set status for resource %s: %v", resource.ID, err)
		}
	}
	return nil
}

// GetResources returns the resources of a specific provider
func (p Protos) GetResources() (map[string]*resource.Resource, error) {

	resources := Resources{}
	resourcesReq, err := http.NewRequest("GET", p.URL+"internal/resource/provider", nil)
	if err != nil {
		return resources, err
	}

	payload, err := p.makeRequest(resourcesReq)
	if err != nil {
		return resources, err
	}

	err = json.Unmarshal(payload, &resources)
	if err != nil {
		return resources, err
	}

	return resources, nil
}

// CreateResource creates a Protos resource
func (p Protos) CreateResource(rsc resource.Resource) (*resource.Resource, error) {
	rscJSON, err := json.Marshal(rsc)
	if err != nil {
		return nil, err
	}

	url := p.URL + "internal/resource"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(rscJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := p.makeRequest(req)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res, &rsc)
	if err != nil {
		return nil, err
	}
	return &rsc, nil
}

// DeleteResource deletes a resource based on the provided id
func (p Protos) DeleteResource(resourceID string) error {
	url := p.URL + "internal/resource/" + resourceID
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	_, err = p.makeRequest(req)
	if err != nil {
		return err
	}
	return nil
}

// RegisterProvider allows an app to register as a provider for a specific resource type
func (p Protos) RegisterProvider(rtype string) error {
	req, err := http.NewRequest("POST", p.URL+"internal/provider/"+rtype, bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	_, err = p.makeRequest(req)
	if err != nil {
		return err
	}

	return nil
}

// DeregisterProvider allows an app to register as a provider for a specific resource type
func (p Protos) DeregisterProvider(rtype string) error {
	req, err := http.NewRequest("DELETE", p.URL+"internal/provider/"+rtype, bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	_, err = p.makeRequest(req)
	if err != nil {
		return err
	}

	return nil
}

// NewClient returns a client that can be used to interact with Protos
func NewClient(url string, appid string) Protos {
	return Protos{
		URL:        url,
		AppID:      appid,
		HTTPclient: &http.Client{},
	}
}
