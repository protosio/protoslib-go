package protoslib

import (
	"bytes"
	"encoding/json"
	"net/http"

	resource "github.com/nustiueudinastea/protos/resource"
)

// CreateResource creates a Protos resource
func (p Protos) CreateResource(rsc resource.Resource) (*resource.Resource, error) {
	rscJSON, err := json.Marshal(rsc)
	if err != nil {
		return nil, err
	}

	url := p.URL + "resource"
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

// GetResource retrieves a resources based on the provided ID
func (p Protos) GetResource(resourceID string) (*resource.Resource, error) {
	url := p.URL + "resource/" + resourceID
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := p.makeRequest(req)
	if err != nil {
		return nil, err
	}

	rsc := resource.Resource{}
	err = json.Unmarshal(res, &rsc)
	if err != nil {
		return nil, err
	}
	return &rsc, nil
}

// DeleteResource deletes a resource based on the provided id
func (p Protos) DeleteResource(resourceID string) error {
	url := p.URL + "resource/" + resourceID
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
