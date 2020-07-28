package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/kinbiko/bugsnag"
)

type server struct{ *bugsnag.Notifier }

func main() {
	n, err := bugsnag.New(bugsnag.Configuration{
		APIKey:       os.Getenv("BUGSNAG_API_KEY"),
		AppVersion:   "1.2.3",
		ReleaseStage: "dev",
	})
	if err != nil {
		panic(err)
	}
	s := &server{Notifier: n}
	http.ListenAndServe(":8080", withRequestMetadata(withPanicReporting(s.Notifier, s.HandleCommentsGet())))
}

func (s *server) HandleCommentsGet() http.HandlerFunc {
	type res struct {
		Comment string `json:"comment"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(&res{Comment: "Nature must wait!"}); err != nil {
			s.Notify(r.Context(), err)
		}
	}
}

func withRequestMetadata(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bytes, _ := ioutil.ReadAll(r.Body)
		body := map[string]interface{}{}
		json.Unmarshal(bytes, &body)
		// NOTE: **you** are responsible for ensuring that you're not sending
		// sensitive information.
		request := map[string]interface{}{
			"body":    body,
			"headers": r.Header,
			"url":     r.URL,
			"method":  r.Method,
		}
		if r.Method == http.MethodGet {
			delete(request, "body")
		}
		ctx := bugsnag.WithMetadata(r.Context(), "request", request)

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withPanicReporting(n *bugsnag.Notifier, h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := n.StartSession(r.Context())
		defer func() {
			if rec := recover(); rec != nil {
				// Note: could additionally check if the panic reported is an
				// error, and then wrap here. It would add more value, but in
				// this case you're probably doing error handling in an
				// unidiomatic manner anyway.
				err := n.Wrap(ctx, fmt.Errorf("%v", rec), "panic in HTTP handler")
				err.Unhandled = true
				err.Panic = true
				// Note: passing in context.Background() here so that we're
				// guaranteed that the payload isn't dropped due to a
				// timeout/deadline on the HTTP request context.
				n.Notify(context.Background(), err)
			}
		}()

		h.ServeHTTP(w, r)
	})
}
