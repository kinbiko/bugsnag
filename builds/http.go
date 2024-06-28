// Package builds is a package for sending build requests to the Bugsnag Build API, as defined here:
// https://bugsnagbuildapi.docs.apiary.io/
package builds

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const timeout = 10 * time.Second

// DefaultPublisher returns a publisher for requests against the Bugsnag SaaS
// endpoint. If you need to target your own on-premise installation, please use
// NewPublisher.
func DefaultPublisher() *Publisher {
	return &Publisher{endpoint: "https://build.bugsnag.com/"}
}

// NewPublisher returns a publisher for requests against an on-premise
// installation of Bugsnag. If you're using the SaaS solution (app.bugsnag.com),
// please use the DefaultPublisher.
func NewPublisher(endpoint string) *Publisher {
	return &Publisher{endpoint: endpoint}
}

// Publisher is a type for sending build requests to the Bugsnag Build API, as defined here:
// https://bugsnagbuildapi.docs.apiary.io/
type Publisher struct {
	endpoint string
}

// Publish sends the request to Bugsnag's Build API.
func (p *Publisher) Publish(req *JSONBuildRequest) error {
	if p.endpoint == "" {
		return errors.New("publisher created incorrectly; please use NewPublisher or DefaultPublisher to construct your builds.Publisher")
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("unable to marshal JSON: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("error when POST-ing request to '%s': %w", p.endpoint, err)
	}
	httpReq.Header.Add("Content-Type", "application/json")

	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("unable to make HTTP call: %w", err)
	}
	defer func() {
		_ = httpRes.Body.Close()
	}()

	got, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %w", err)
	}

	type response struct {
		Status   string   `json:"status"`
		Warnings []string `json:"warnings"`
		Errors   []string `json:"errors"`
	}
	var res response
	if err := json.Unmarshal(got, &res); err != nil {
		return fmt.Errorf("unable to read response body as JSON: %w", err)
	}
	if res.Status == "error" {
		return fmt.Errorf("error when sending message: %s", res.Errors[0])
	}

	return nil
}
