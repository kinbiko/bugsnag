package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kinbiko/jsonassert"
)

func TestRelease(t *testing.T) {
	testServer := func() (*httptest.Server, chan string) {
		reqs := make(chan string, 10)
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := ioutil.ReadAll(r.Body)
			reqs <- string(body)
			w.Write([]byte(`{"status": "ok"}`))
		})), reqs
	}

	ts, reqs := testServer()
	defer ts.Close()

	t.Run("small payload", func(t *testing.T) {
		cmd := fmt.Sprintf(`release
--api-key=1234abcd1234abcd1234abcd1234abcd
--app-version=2.5.2
--endpoint=%s`, ts.URL)
		err := run(strings.Split(cmd, "\n"), map[string]string{})
		if err != nil {
			t.Fatal(err)
		}

		var body string
		select {
		case body = <-reqs:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("no request received after half a second.")
		}

		jsonassert.New(t).Assertf(body, `
		{
			"apiKey": "1234abcd1234abcd1234abcd1234abcd",
			"appVersion": "2.5.2"
		}`)
	})

	t.Run("big payload", func(t *testing.T) {
		cmd := fmt.Sprintf(`release
--api-key=1234abcd1234abcd1234abcd1234abcd
--app-version=2.5.2
--release-stage=staging
--provider=github
--repository=https://github.com/kinbiko/bugsnag
--revision=11c61751225399433faa4805ec2011d1
--builder=kinbiko
--metadata=Ticket=JIRA-1234
--auto-assign-release=true
--app-version-code=1234
--app-bundle-version=5.2
--endpoint=%s`, ts.URL)
		err := run(strings.Split(cmd, "\n"), map[string]string{})
		if err != nil {
			t.Fatal(err)
		}

		var body string
		select {
		case body = <-reqs:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("no request received after half a second.")
		}

		jsonassert.New(t).Assertf(body, `
		{
			"apiKey": "1234abcd1234abcd1234abcd1234abcd",
			"appVersion": "2.5.2",
			"releaseStage": "staging",
			"sourceControl": {
				"provider": "github",
				"revision": "11c61751225399433faa4805ec2011d1",
				"repository": "https://github.com/kinbiko/bugsnag"
			},
			"builderName": "kinbiko",
			"autoAssignRelease": true,
			"appBundleVersion": "5.2",
			"appVersionCode": 1234,
			"metadata": {
				"Ticket": "JIRA-1234"
			}
		}`)
	})
}
