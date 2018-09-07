package protoslib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	resource "github.com/protosio/protos/resource"
)

// RegisterProvider allows an app to register as a provider for a specific resource type
func (p Protos) RegisterProvider(rtype string) error {
	req, err := http.NewRequest("POST", p.URL+"provider/"+rtype, bytes.NewBuffer([]byte{}))
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
	req, err := http.NewRequest("DELETE", p.URL+"provider/"+rtype, bytes.NewBuffer([]byte{}))
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

// UpdateResourceValue updates a resource's value based on the ID and new value provided
func (p Protos) UpdateResourceValue(resourceID string, newValue resource.Type) error {
	payloadJSON, err := json.Marshal(newValue)
	if err != nil {
		return err
	}

	url := p.URL + "resource/" + resourceID
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

	url := p.URL + "resource/" + resourceID
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
	resourcesReq, err := http.NewRequest("GET", p.URL+"resource/provider", nil)
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
