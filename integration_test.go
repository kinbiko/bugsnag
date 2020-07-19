package bugsnag_test

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/kinbiko/bugsnag"

	"github.com/kinbiko/jsonassert"
)

func TestIntegration(t *testing.T) {
	reports := make(chan string, 1)

	ntfServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		reports <- string(payload)
	}))
	defer ntfServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer sessionServer.Close()

	ntf, _ := bugsnag.New(bugsnag.Configuration{
		EndpointNotify:   ntfServer.URL,
		EndpointSessions: sessionServer.URL,
		APIKey:           "abcd1234abcd1234abcd1234abcd1234",
		AppVersion:       "5.2.3",
		ReleaseStage:     "staging",
	})

	ctx := bugsnag.WithUser(context.Background(), bugsnag.User{
		ID:    "1234",
		Name:  "River Tam",
		Email: "river@serenity.space",
	})
	ctx = ntf.StartSession(ctx)

	ctx = bugsnag.WithBreadcrumb(ctx, bugsnag.Breadcrumb{
		Name:     "something happened",
		Type:     bugsnag.BCTypeProcess,
		Metadata: map[string]interface{}{"md": "foo"},
	})

	ctx = bugsnag.WithBreadcrumb(ctx, bugsnag.Breadcrumb{
		Name:     "something else happened",
		Type:     bugsnag.BCTypeRequest,
		Metadata: map[string]interface{}{"md": "bar"},
	})

	ctx = bugsnag.WithBugsnagContext(ctx, "User batch job")

	ctx = bugsnag.WithMetadata(ctx, "myTab", map[string]interface{}{"hello": 423})
	ctx = bugsnag.WithMetadatum(ctx, "myTab", "goodbye", "cruel world")

	err := ntf.Wrap(context.Background(), errors.New("oh ploppers"))
	err.Unhandled = true
	err.Panic = true
	ntf.Notify(ctx, err) // testing this synchronously in order to get more stack frames

	var payload string
	select {
	case rep := <-reports:
		payload = rep
	case <-time.Tick(time.Second):
		t.Fatal("waited 1 second for a report but none arrived")
	}

	hostname, _ := os.Hostname()
	// The inProject flag won't work, as the debug package doesn't identify test pacakges as Main modules
	jsonassert.New(t).Assertf(payload, `{
		"apiKey": "abcd1234abcd1234abcd1234abcd1234",
		"notifier": {
			"name": "Alternative Go Notifier",
			"version": "<<PRESENCE>>",
			"url": "https://github.com/kinbiko/bugsnag"
		},
		"events": [
			{
				"payloadVersion": "5",
				"severity": "error",
				"severityReason": { "type": "unhandledPanic" },
				"unhandled": true,
				"context": "User batch job",
				"app": { "version": "5.2.3", "releaseStage": "staging", "duration": "<<PRESENCE>>" },
				"device": { "hostname": "%s", "osName": "%s", "osVersion": "<<PRESENCE>>", "memStats": "<<PRESENCE>>", "goroutineCount": "<<PRESENCE>>", "runtimeVersions": { "go": "%s" } },
				"user": { "id": "1234", "name": "River Tam", "email": "river@serenity.space" },
				"metaData": {"myTab": {"goodbye": "cruel world", "hello": 423}},
				"session": { "id": "<<PRESENCE>>", "startedAt": "<<PRESENCE>>", "events": { "unhandled": 1 } },
				"breadcrumbs": [
					{ "metaData": {"md": "bar"}, "name": "something else happened", "timestamp": "<<PRESENCE>>", "type": "request" },
					{ "metaData": {"md": "foo"}, "name": "something happened", "timestamp": "<<PRESENCE>>", "type": "process" }
				],
				"exceptions": [
					{
						"errorClass": "*bugsnag.Error",
						"message": "oh ploppers",
						"stacktrace": [
							{"file":"<<PRESENCE>>","inProject":false,"lineNumber":67,"method":"github.com/kinbiko/bugsnag_test.TestIntegration"},
							{"file":"<<PRESENCE>>","inProject":false,"lineNumber":"<<PRESENCE>>","method":"<<PRESENCE>>"},
							{"file":"<<PRESENCE>>","inProject":false,"lineNumber":"<<PRESENCE>>","method":"<<PRESENCE>>"}
						]
					}, {
						"errorClass": "*errors.errorString",
						"message": "oh ploppers",
						"stacktrace": null
					}
				]
			}
		]
	}`, hostname, runtime.GOOS, runtime.Version())
}

func TestReportSerialization(t *testing.T) {
	payload, err := json.Marshal(&bugsnag.JSONErrorReport{
		APIKey: "hello",
		Notifier: &bugsnag.JSONNotifier{
			Name:    "My custom notifier",
			Version: "1.2.3",
			URL:     "https://github.com/kinbiko/bugsnag",
		},
		Events: []*bugsnag.JSONEvent{
			{

				PayloadVersion: "5",
				Context:        "UserController",
				Unhandled:      true,
				Severity:       "info",

				SeverityReason: &bugsnag.JSONSeverityReason{Type: "log"},

				Breadcrumbs: []*bugsnag.JSONBreadcrumb{
					{
						Timestamp: "2016-07-19T12:17:27-0700",
						Name:      "Error log",
						Type:      "log",
						Metadata:  map[string]interface{}{"message": "got a 500 from the server"},
					},
				},

				Request: &bugsnag.JSONRequest{
					ClientIP:   "127.0.0.1",
					HTTPMethod: "GET",
					URL:        "http://example.com/users/19/settings",
					Referer:    "http://example.com/fish?q=awesome",
					Headers: map[string]string{
						"Accept":          "*/*",
						"Accept-Encoding": "gzip, deflate, sdch, br",
						"Accept-Language": "en-US,en;q=0.8",
					},
				},

				User: &bugsnag.JSONUser{
					ID:    "5134",
					Name:  "Angus MacGyver",
					Email: "mac@phoenix.example.com",
				},

				App: &bugsnag.JSONApp{
					ID:           "61387",
					Version:      "5.2.3",
					ReleaseStage: "production",
					Type:         "gin",
					Duration:     1234,
				},

				Device: &bugsnag.JSONDevice{
					Hostname:        "web1.internal",
					OSName:          "android",
					OSVersion:       "8.0.1",
					RuntimeVersions: map[string]string{"go": "1.11.2"},
				},

				Session: &bugsnag.JSONSession{
					ID:        "67178",
					StartedAt: "2018-06-07T10:16:34.564Z",
					Events:    &bugsnag.JSONSessionEvents{Handled: 5, Unhandled: 2},
				},

				Metadata: map[string]map[string]interface{}{
					"whatever": {
						"doesntMatter": "as long as the structure is right",
					},
				},

				Exceptions: []*bugsnag.JSONException{
					{
						ErrorClass: "RandomError",
						Message:    "Something went terribly wrong",
						Stacktrace: []*bugsnag.JSONStackframe{
							{
								File:       "cool.go",
								LineNumber: 41,
								Method:     "coolFunc",
								InProject:  true,
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	jsonassert.New(t).Assertf(string(payload), `
		{
			"apiKey": "hello",
			"notifier": {
				"name": "My custom notifier",
				"version": "1.2.3",
				"url": "https://github.com/kinbiko/bugsnag"
			},
			"events": [
				{
					"payloadVersion": "5",
					"context": "UserController",
					"unhandled": true,
					"severity": "info",
					"severityReason": { "type": "log" },
					"breadcrumbs": [
						{
							"timestamp": "2016-07-19T12:17:27-0700",
							"name": "Error log",
							"type": "log",
							"metaData": {"message":"got a 500 from the server"}
						}
					],
					"request": {
						"clientIp": "127.0.0.1",
						"headers": {
							"Accept": "*/*",
							"Accept-Encoding": "gzip, deflate, sdch, br",
							"Accept-Language": "en-US,en;q=0.8"
						},
						"httpMethod": "GET",
						"url": "http://example.com/users/19/settings",
						"referer": "http://example.com/fish?q=awesome"
					},
					"user": {
						"id": "5134",
						"name": "Angus MacGyver",
						"email": "mac@phoenix.example.com"
					},
					"app": {
						"id": "61387",
						"version": "5.2.3",
						"releaseStage": "production",
						"type": "gin",
						"duration": 1234
					},
					"device": {
						"hostname": "web1.internal",
						"osName": "android",
						"osVersion": "8.0.1",
						"runtimeVersions": { "go": "1.11.2" }
					},
					"session": {
						"id": "67178",
						"startedAt": "2018-06-07T10:16:34.564Z",
						"events": { "handled": 5, "unhandled": 2 }
					},
					"metaData": {
						"whatever": {
							"doesntMatter": "as long as the structure is right"
						}
					},
					"exceptions": [
						{
							"errorClass": "RandomError",
							"message": "Something went terribly wrong",
							"stacktrace": [
								{
									"file": "cool.go",
									"lineNumber": 41,
									"method": "coolFunc",
									"inProject": true
								}
							]
						}
					]
				}
			]
		}`)
}
