package bugsnag

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

type ctxKey int

const (
	sessionKey ctxKey = iota + 1
	ctxDataKey
)

// Serialize extracts all the diagnostic data tracked within the given ctx to a
// []byte that can later be deserialized using Deserialize. Useful for passing
// diagnostic to downstream services in header values.
func (n *Notifier) Serialize(ctx context.Context) []byte {
	b, err := json.Marshal(getAttachedContextData(ctx))
	if err != nil {
		n.cfg.InternalErrorCallback(err)
		return nil
	}
	return []byte(base64.StdEncoding.EncodeToString(b))
}

// Deserialize extracts diagnostic data that has previously been serialized
// with Serialize and attaches it to the given ctx. Intended to be called in
// server middleware to attach diagnostic data identified from upstream
// services. As a result, any existing diagnostic data (session data from
// StartSession not inclusive) will be wiped.
// Note: If the upstream service attaches sensitive data this service should
// not report (e.g. user info), then this too will be propagated in this
// context, and you will have to use the ErrorReportSanitizer to remove this
// data.
func (n *Notifier) Deserialize(ctx context.Context, data []byte) context.Context {
	jsonData, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		n.cfg.InternalErrorCallback(err)
		return ctx
	}
	cd := &ctxData{} // nolint:exhaustivestruct // we're about to fill the data dynamically
	if err := json.Unmarshal(jsonData, cd); err != nil {
		n.cfg.InternalErrorCallback(err)
		return ctx
	}
	return context.WithValue(ctx, ctxDataKey, cd)
}

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

type ctxData struct {
	BContext    string                            `json:"cx"`
	Breadcrumbs []Breadcrumb                      `json:"bc"`
	User        *User                             `json:"us"`
	Metadata    map[string]map[string]interface{} `json:"md"`
}

// Breadcrumb represents user- and system-initiated events which led up
// to an error, providing additional context.
type Breadcrumb struct {
	// A short summary describing the breadcrumb, such as entering a new
	// application state
	Name string `json:"na"`

	// Type is a category which describes the breadcrumb, from the list of
	// allowed values. Accepted values are of the form bugsnag.BCType*.
	Type bcType `json:"ty"`

	// Metadata contains any additional information about the breadcrumb, as
	// key/value pairs.
	Metadata map[string]interface{} `json:"md"`

	// Timestamp is set automatically to the current timestamp if not set.
	Timestamp time.Time `json:"ts"`
}

// WithBreadcrumb attaches a breadcrumb to the top of the stack of breadcrumbs
// stored in the given context.
func (n *Notifier) WithBreadcrumb(ctx context.Context, breadcrumb Breadcrumb) context.Context {
	if ctx == nil {
		return nil
	}
	// This function currently uses no features of the Notifier type, however
	// we're attaching it to the Notifier to ensure that we can use
	// Notifier-only functionalities in the future AND so that users need only
	// import the bugsnag package in a single location in their app.
	if breadcrumb.Timestamp.IsZero() {
		breadcrumb.Timestamp = time.Now().UTC()
	}
	cd := getAttachedContextData(ctx)
	cd.Breadcrumbs = append(cd.Breadcrumbs, breadcrumb)
	return context.WithValue(ctx, ctxDataKey, cd)
}

func makeBreadcrumbs(ctx context.Context) []*JSONBreadcrumb {
	bcs := getAttachedContextData(ctx).Breadcrumbs
	if bcs == nil {
		return nil
	}

	payloads := make([]*JSONBreadcrumb, len(bcs))
	for i, bc := range bcs {
		payloads[len(bcs)-1-i] = &JSONBreadcrumb{
			Timestamp: bc.Timestamp.Format(time.RFC3339),
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
func (n *Notifier) WithUser(ctx context.Context, user User) context.Context {
	if ctx == nil {
		return nil
	}
	// This function currently uses no features of the Notifier type, however
	// we're attaching it to the Notifier to ensure that we can use
	// Notifier-only functionalities in the future AND so that users need only
	// import the bugsnag package in a single location in their app.
	cd := getAttachedContextData(ctx)
	cd.User = &user
	return context.WithValue(ctx, ctxDataKey, cd)
}

// WithBugsnagContext applies the given bContext as the "Context" for the errors that
// show up in your Bugsnag dashboard. The naming here is unfortunate, but to be
// fair, Bugsnag had this nomenclature before Go did...
func (n *Notifier) WithBugsnagContext(ctx context.Context, bContext string) context.Context {
	if ctx == nil {
		return nil
	}
	// This function currently uses no features of the Notifier type, however
	// we're attaching it to the Notifier to ensure that we can use
	// Notifier-only functionalities in the future AND so that users need only
	// import the bugsnag package in a single location in their app.
	cd := getAttachedContextData(ctx)
	cd.BContext = bContext
	return context.WithValue(ctx, ctxDataKey, cd)
}

// WithMetadatum attaches the given key and value under the provided tab in the
// Bugsnag dashboard. You may use the following tab names to add data to
// existing/common tabs in the dashboard with the same name:
//   "user", "app", "device", "request"
func (n *Notifier) WithMetadatum(ctx context.Context, tab, key string, value interface{}) context.Context {
	if ctx == nil {
		return nil
	}
	m := initializeMetadataTab(ctx, tab)
	m[tab][key] = value
	return n.WithMetadata(ctx, tab, m[tab])
}

// WithMetadata attaches the given data under the provided tab in the
// Bugsnag dashboard. You may use the following tab names to add data to
// existing/common tabs in the dashboard with the same name:
//   "user", "app", "device", "request"
func (n *Notifier) WithMetadata(ctx context.Context, tab string, data map[string]interface{}) context.Context {
	if ctx == nil {
		return nil
	}
	// This function currently uses no features of the Notifier type, however
	// we're attaching it to the Notifier to ensure that we can use
	// Notifier-only functionalities in the future AND so that users need only
	// import the bugsnag package in a single location in their app.
	m := initializeMetadataTab(ctx, tab)
	m[tab] = data
	cd := getAttachedContextData(ctx)
	cd.Metadata = m
	return context.WithValue(ctx, ctxDataKey, cd)
}

// Metadata pulls out all the metadata known by this package as a
// map[tab]map[key]value from the given context.
func (n *Notifier) Metadata(ctx context.Context) map[string]map[string]interface{} {
	// This function currently uses no features of the Notifier type, however
	// we're attaching it to the Notifier to ensure that we can use
	// Notifier-only functionalities in the future AND so that users need only
	// import the bugsnag package in a single location in their app.
	return getAttachedContextData(ctx).Metadata
}

func initializeMetadataTab(ctx context.Context, tab string) map[string]map[string]interface{} {
	metadata := getAttachedContextData(ctx).Metadata
	if metadata == nil {
		metadata = map[string]map[string]interface{}{}
	}

	if metadata[tab] == nil {
		metadata[tab] = map[string]interface{}{}
	}
	return metadata
}

type jsonCtxData struct {
	bContext    string
	breadcrumbs []*JSONBreadcrumb
	user        *JSONUser
	session     *JSONSession
	metadata    map[string]map[string]interface{}
}

func extractAugmentedContextData(ctx context.Context, err error, unhandled bool) (*jsonCtxData, context.Context) {
	data := &jsonCtxData{
		bContext:    getAttachedContextData(ctx).BContext,
		breadcrumbs: makeBreadcrumbs(ctx),
		user:        getAttachedContextData(ctx).User,
		session:     makeJSONSession(ctx, unhandled),
		metadata:    getAttachedContextData(ctx).Metadata,
	}
	lowestCtx := ctx
	lowestErr := err
	for {
		if berr, ok := lowestErr.(*Error); ok {
			ctx = berr.ctx
			if ctx != nil {
				data.updateFromCtx(ctx, unhandled)
				lowestCtx = ctx
			}
		}
		lowestErr = errors.Unwrap(lowestErr)
		if lowestErr == nil {
			break
		}
	}

	if data.bContext == "" {
		data.bContext = err.Error()
	}
	return data, lowestCtx
}

func (data *jsonCtxData) updateFromCtx(ctx context.Context, unhandled bool) {
	if dataBContext := getAttachedContextData(ctx).BContext; dataBContext != "" {
		data.bContext = dataBContext
	}
	if dataBreadcrumbs := makeBreadcrumbs(ctx); dataBreadcrumbs != nil {
		data.breadcrumbs = dataBreadcrumbs
	}
	if dataUser := getAttachedContextData(ctx).User; dataUser != nil {
		data.user = dataUser
	}
	if dataSession := makeJSONSession(ctx, unhandled); dataSession != nil {
		data.session = dataSession
	}

	dataMetadata := getAttachedContextData(ctx).Metadata
	if dataMetadata == nil {
		return
	}
	if data.metadata == nil {
		data.metadata = map[string]map[string]interface{}{}
	}
	for tab, kvps := range dataMetadata {
		if data.metadata[tab] == nil {
			data.metadata[tab] = map[string]interface{}{}
		}
		for k, v := range kvps {
			data.metadata[tab][k] = v
		}
	}
}

func getAttachedContextData(ctx context.Context) *ctxData {
	if val := ctx.Value(ctxDataKey); val != nil {
		return val.(*ctxData) // nolint:forcetypeassert // This is safe. We own the key => we own the type
	}
	return &ctxData{} // nolint:exhaustivestruct // this saves lots of nil checks elsewhere
}
