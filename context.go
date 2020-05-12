package bugsnag

import (
	"context"
	"errors"
	"time"
)

type ctxKey int

const (
	sessionKey ctxKey = iota + 1
	userKey
	breadcrumbKey
	contextKey
	metadataKey
)

type bcType int

const (
	// BCTypeManual is the type of breadcrumb representing user-defined,
	// manually added breadcrumbs
	BCTypeManual bcType = iota // listed first to make it the default
	// BCTypeNavigation is the type of breadcrumb representing changing screens
	// or content being displayed, with a defined destination and optionally a
	// previous location.
	BCTypeNavigation
	// BCTypeRequest is the type of breadcrumb representing sending and
	// receiving requests and responses.
	BCTypeRequest
	// BCTypeProcess is the type of breadcrumb representing performing an
	// intensive task or query.
	BCTypeProcess
	// BCTypeLog is the type of breadcrumb representing messages and severity
	// sent to a logging platform.
	BCTypeLog
	// BCTypeUser is the type of breadcrumb representing actions performed by
	// the user, like text input, button presses, or confirming/canceling an
	// alert dialog.
	BCTypeUser
	// BCTypeState is the type of breadcrumb representing changing the overall
	// state of an app, such as closing, pausing, or being moved to the
	// background, as well as device state changes like memory or battery
	// warnings and network connectivity changes.
	BCTypeState
	// BCTypeError is the type of breadcrumb representing an error which was
	// reported to Bugsnag encountered in the same session.
	BCTypeError
)

func (b bcType) val() string {
	return []string{
		"manual",
		"navigation",
		"request",
		"process",
		"log",
		"user",
		"state",
		"error",
	}[b]
}

// Breadcrumb represents user- and system-initiated events which led up
// to an error, providing additional context.
type Breadcrumb struct {
	// A short summary describing the breadcrumb, such as entering a new
	// application state
	Name string

	// Type is a category which describes the breadcrumb, from the list of
	// allowed values. Accepted values are of the form bugsnag.BCType*.
	Type bcType

	// Metadata contains any additional information about the breadcrumb, as
	// key/value pairs.
	Metadata map[string]interface{}

	timestamp time.Time
}

// WithBreadcrumb attaches a breadcrumb to the top of the stack of breadcrumbs
// stored in the given context.
func WithBreadcrumb(ctx context.Context, b Breadcrumb) context.Context {
	b.timestamp = time.Now().UTC()
	val := ctx.Value(breadcrumbKey)
	if val == nil {
		return context.WithValue(ctx, breadcrumbKey, []Breadcrumb{b})
	}
	if bcs, ok := val.([]Breadcrumb); ok {
		bcs = append([]Breadcrumb{b}, bcs...)
		return context.WithValue(ctx, breadcrumbKey, bcs)
	}
	return context.WithValue(ctx, breadcrumbKey, []Breadcrumb{b})
}

func makeBreadcrumbs(ctx context.Context) []*JSONBreadcrumb {
	val := ctx.Value(breadcrumbKey)
	if val == nil {
		return nil
	}

	bcs, ok := val.([]Breadcrumb)
	if !ok {
		return nil
	}

	payloads := make([]*JSONBreadcrumb, len(bcs))
	for i, bc := range bcs {
		payloads[i] = &JSONBreadcrumb{
			Timestamp: bc.timestamp.Format(time.RFC3339),
			Name:      bc.Name,
			Type:      bc.Type.val(),
			Metadata:  bc.Metadata,
		}
	}
	return payloads
}

// User information about the user affected by the error. These fields are
// optional but highly recommended. To display custom user data alongside these
// standard fields on the Bugsnag website, the custom data should be included
// in the metaData object in a user object.
type User struct {
	// ID is a unique identifier for a user affected by the event.
	// This could be any distinct identifier that makes sense for your app.
	ID string `json:"id,omitempty"`

	// Name is a human readable name of the user affected.
	Name string `json:"name,omitempty"`

	// Email is the user's email address, if known.
	Email string `json:"email,omitempty"`
}

// WithUser attaches the given User data to the given context, such that it can
// later be provided to the Notify method, and have this data show up in your
// dashboard.
func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userKey, &user)
}

// WithBugsnagContext applies the given bContext as the "Context" for the errors that
// show up in your Bugsnag dashboard. The naming here is unfortunate, but to be
// fair, Bugsnag had this nomenclature before Go did...
func WithBugsnagContext(ctx context.Context, bContext string) context.Context {
	return context.WithValue(ctx, contextKey, bContext)
}

// WithMetadatum attaches the given key and value under the provided tab in the
// Bugsnag dashboard. You may use the following tab names to add data to
// existing/common tabs in the dashboard with the same name:
//   "user", "app", "device", "request"
func WithMetadatum(ctx context.Context, tab, key string, value interface{}) context.Context {
	m := initializeMetadataTab(ctx, tab)
	m[tab][key] = value
	return WithMetadata(ctx, tab, m[tab])
}

// WithMetadata attaches the given data under the provided tab in the
// Bugsnag dashboard. You may use the following tab names to add data to
// existing/common tabs in the dashboard with the same name:
//   "user", "app", "device", "request"
func WithMetadata(ctx context.Context, tab string, data map[string]interface{}) context.Context {
	m := initializeMetadataTab(ctx, tab)
	m[tab] = data
	return context.WithValue(ctx, metadataKey, m)
}

func initializeMetadataTab(ctx context.Context, tab string) map[string]map[string]interface{} {
	m := Metadata(ctx)
	if m == nil {
		m = map[string]map[string]interface{}{}
	}

	if m[tab] == nil {
		m[tab] = map[string]interface{}{}
	}
	return m
}

// Metadata pulls out all the metadata known by this package as a
// map[tab]map[key]value from the given context.
func Metadata(ctx context.Context) map[string]map[string]interface{} {
	if m, ok := ctx.Value(metadataKey).(map[string]map[string]interface{}); ok {
		return m
	}
	return nil
}

func extractInnermostCtx(ctx context.Context, err error, unhandled bool) *ctxData {
	data := &ctxData{
		bContext:    makeContext(ctx),
		breadcrumbs: makeBreadcrumbs(ctx),
		user:        makeUser(ctx),
		session:     makeJSONSession(ctx, unhandled),
		metadata:    Metadata(ctx),
	}
	var e error = err
	for {
		if berr, ok := e.(*Error); ok {
			ctx = berr.ctx
			if ctx != nil {
				data.updateFromCtx(ctx, unhandled)
			}
		}
		e = errors.Unwrap(e)
		if e == nil {
			break
		}
	}

	if data.bContext == "" {
		data.bContext = err.Error()
	}
	return data
}

func makeUser(ctx context.Context) *User {
	u := ctx.Value(userKey)
	if u == nil {
		return nil
	}
	user, ok := u.(*User)
	if !ok {
		return nil
	}
	return user
}

func (data *ctxData) updateFromCtx(ctx context.Context, unhandled bool) {
	if dataBContext := makeContext(ctx); dataBContext != "" {
		data.bContext = dataBContext
	}
	if dataBreadcrumbs := makeBreadcrumbs(ctx); dataBreadcrumbs != nil {
		data.breadcrumbs = dataBreadcrumbs
	}
	if dataUser := makeUser(ctx); dataUser != nil {
		data.user = dataUser
	}
	if dataSession := makeJSONSession(ctx, unhandled); dataSession != nil {
		data.session = dataSession
	}

	dataMetadata := Metadata(ctx)
	if dataMetadata == nil {
		return
	}
	if data.metadata == nil {
		data.metadata = map[string]map[string]interface{}{}
	}
	for tab, kvps := range dataMetadata {
		for k, v := range kvps {
			data.metadata[tab][k] = v
		}
	}
}

func makeContext(ctx context.Context) string {
	if v, ok := ctx.Value(contextKey).(string); ok {
		return v
	}
	return ""
}
