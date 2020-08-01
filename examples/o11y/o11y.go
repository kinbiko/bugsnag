package o11y

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/kinbiko/bugsnag"
	"github.com/sirupsen/logrus"
)

// Olly is easier to write and talk about than O11y.
type Olly struct {
	// We embed the Bugsnag notifier here so that calls to o.Wrap won't include
	// an intermediate stackframe in this wrapper.
	// This prevents awkward groupings, as the top in-project stakframe won't
	// consistently be this intermediate stackframe.
	*bugsnag.Notifier
	// Embedding the others just to get a clean interfaces in the app.
	*statsd.Client
	*logrus.Logger
}

type Config struct{ BugsnagAPIKey, AppVersion, ReleaseStage, DatadogAgentAddr string }

func NewO11y(cfg *Config) *Olly {
	l := logrus.New()
	return &Olly{
		Notifier: makeBugsnagNotifier(l, cfg.BugsnagAPIKey, cfg.AppVersion, cfg.ReleaseStage),
		Client:   makeDatadogClient(cfg.DatadogAgentAddr),
		Logger:   l,
	}
}

// Log logs the given message at INFO severity, including log metadata
// provided in the context, formatting the message if appropriate.
func (o *Olly) Log(ctx context.Context, msg string, args ...interface{}) context.Context {
	md := metadata(ctx)
	delete(md, "request.headers")
	message := fmt.Sprintf(msg, args...)
	o.WithFields(logrus.Fields(md)).Infof(message)
	md["message"] = message
	return bugsnag.WithBreadcrumb(ctx, bugsnag.Breadcrumb{Name: "Info log message", Type: bugsnag.BCTypeLog, Metadata: md})
}

func makeBugsnagNotifier(l *logrus.Logger, apiKey, appVersion, releaseStage string) *bugsnag.Notifier {
	n, err := bugsnag.New(bugsnag.Configuration{
		APIKey:       apiKey,
		AppVersion:   appVersion,
		ReleaseStage: releaseStage,
		ErrorReportSanitizer: func(ctx context.Context, r *bugsnag.JSONErrorReport) context.Context {
			// Log whenever we report an exception.
			l.Error(r.Events[0].Exceptions[0].Message)
			return context.Background()
		},
	})
	if err != nil {
		panic(err)
	}
	return n
}

func makeDatadogClient(addr string) *statsd.Client {
	c, err := statsd.New(addr)
	if err != nil {
		panic(err)
	}
	return c
}

func metadata(ctx context.Context) map[string]interface{} {
	var (
		bmd = bugsnag.Metadata(ctx)
		md  = map[string]interface{}{}
	)

	for t := range bmd {
		for k, v := range bmd[t] {
			md[t+"."+k] = v
		}
	}
	return md
}
