package bugsnag

import (
	"context"
	"time"
)

// bcType represents the type of a breadcrumb
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
