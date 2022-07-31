package builds_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kinbiko/bugsnag/builds"
	"github.com/kinbiko/jsonassert"
)

func TestNonConstructedPublisher(t *testing.T) {
	pub := builds.Publisher{}
	err := pub.Publish(&builds.JSONBuildRequest{})
	mustContain(t, err, "NewPublisher", "DefaultPublisher")
}

func TestPublishingBuilds(t *testing.T) {
	ts, reqs := testServer()
	defer ts.Close()
	p := builds.NewPublisher(ts.URL)
	t.Run("large payload", func(t *testing.T) {
		err := p.Publish(makeBigValidReq())
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
			"appVersion": "1.5.2",
			"builderName": "River Tam",
			"sourceControl": {
				"provider": "github",
				"repository": "https://github.com/kinbiko/bugsnag",
				"revision": "9fc0b224985fc09d1ced97e51a0e8f166f1d190a"
			},
			"metadata": {
				"Tickets": "JIRA-1234, JIRA-4321"
			},
			"appVersionCode": 33,
			"appBundleVersion": "42.3",
			"autoAssignRelease": true,
			"releaseStage": "staging"
		}`)
	})

	t.Run("small payload", func(t *testing.T) {
		err := p.Publish(makeSmallValidReq())
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
			"appVersion": "1.5.2"
		}`)
	})

}

func testServer() (*httptest.Server, chan string) {
	reqs := make(chan string, 10)
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		reqs <- string(body)
		w.Write([]byte(`{"status": "ok"}`))
	})), reqs
}
