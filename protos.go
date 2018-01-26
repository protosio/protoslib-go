package protos

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

var protosURL string

// DNSResource is a Protos resource for DNS
type DNSResource struct {
	Host  string `json:"host"`
	Value string `json:"value" hash:"-"`
	Type  string `json:"type"`
	TTL   int    `json:"ttl" hash:"-"`
}

// Resource represents a generalised Protos resource
type Resource struct {
	Type   string      `json:"type"`
	Record interface{} `json:"value"`
	Status string      `json:"status"`
	ID     string      `json:"id"`
}

// Protos client struct
type Protos struct {
	URL        string
	AppID      string
	HTTPclient *http.Client
}

func newDNSResource(m interface{}) DNSResource {
	v := m.(map[string]interface{})
	return DNSResource{
		Host:  v["host"].(string),
		Value: v["value"].(string),
		Type:  v["type"].(string),
		TTL:   int(v["ttl"].(float64)),
	}
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
	req.Header.Set("Content-Type", "application/json")

	_, err = p.makeRequest(req)
	if err != nil {
		return err
	}

	return nil
}

// SetStatusBatch takes a list of Resource and applies the same status to all of them
func (p Protos) SetStatusBatch(resources map[string]*Resource, rstatus string) error {
	for _, resource := range resources {
		err := p.SetResourceStatus(resource.ID, rstatus)
		if err != nil {
			return fmt.Errorf("Could not set status for resource %s: %v", resource.ID, err)
		}
	}
	return nil
}

// GetResources returns the resources of a specific provider
func (p Protos) GetResources() (map[string]*Resource, error) {

	resourcesReq, err := http.NewRequest("GET", p.URL+"internal/resource/provider", nil)
	resources := make(map[string]*Resource)

	payload, err := p.makeRequest(resourcesReq)
	if err != nil {
		return map[string]*Resource{}, err
	}

	err = json.Unmarshal(payload, &resources)
	if err != nil {
		return map[string]*Resource{}, err
	}

	for _, resource := range resources {
		if resource.Type == "dns" {
			resource.Record = newDNSResource(resource.Record)
		} else {
			return map[string]*Resource{}, errors.New("Unknown resource type: " + resource.Type + " for resource " + resource.ID)
		}
	}

	return resources, nil
}

// RegisterProvider allows an app to register as a provider for a specific resource type
func (p Protos) RegisterProvider(rtype string) error {
	req, err := http.NewRequest("POST", p.URL+"internal/provider/"+rtype, bytes.NewBuffer([]byte{}))
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
