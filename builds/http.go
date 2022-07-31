package builds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// DefaultPublisher returns a publisher for Payloads against the Bugsnag SaaS
// endpoint. If you need to target your own on-premise installation, please use
// NewPublisher.
func DefaultPublisher() *Publisher {
	return &Publisher{endpoint: "https://build.bugsnag.com/"}
}

// DefaultPublisher returns a publisher for Payloads against an on-premise
// installation of Bugsnag. If you're using the SaaS solution (app.bugsnag.com),
// please use the DefaultPublisher.
func NewPublisher(endpoint string) *Publisher {
	return &Publisher{endpoint: endpoint}
}

// Publisher is a type for sending build requests to the Bugsnag Build API, as defined here:
//https://bugsnagbuildapi.docs.apiary.io/
type Publisher struct {
	endpoint string
}

// Publish validates and sends the payload to Bugsnag's Build API.
func (p *Publisher) Publish(pl *JSONBuildRequest) error {
	if p.endpoint == "" {
		return fmt.Errorf("publisher created incorrectly; please use NewPublisher or DefaultPublisher to construct your builds.Publisher")
	}

	b, err := json.Marshal(pl)
	if err != nil {
		return fmt.Errorf("unable to marshal JSON: %w", err)
	}

	httpReq, err := http.NewRequest("POST", p.endpoint, bytes.NewBuffer(b))

	if err != nil {
		return fmt.Errorf("error when POST-ing payload to '%s': %w", p.endpoint, err)
	}
	httpReq.Header.Add("Content-Type", "application/json")

	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}

	got, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return err
	}

	type response struct {
		Status   string   `json:"status"`
		Warnings []string `json:"warnings"`
		Errors   []string `json:"errors"`
	}
	var res response
	if err := json.Unmarshal(got, &res); err != nil {
		return err
	}
	if res.Status == "error" {
		return fmt.Errorf("error when sending message: %s", res.Errors[0])
	}

	return nil
}
