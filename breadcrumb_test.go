package bugsnag

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kinbiko/jsonassert"
)

func TestBreadcrumbs(t *testing.T) {
	asJSON := func(s interface{}) string {
		b, err := json.Marshal(s)
		if err != nil {
			return "<<ERROR>>"
		}
		return string(b)
	}

	expLatest := Breadcrumb{Name: "latest"}
	expNewest := Breadcrumb{Name: "newest", Type: BCTypeLog, Metadata: map[string]interface{}{"md": 1}}

	ctx := context.Background()
	ctx = WithBreadcrumb(ctx, expLatest)
	ctx = WithBreadcrumb(ctx, expNewest)

	bcs := makeBreadcrumbs(ctx)

	if len(bcs) != 2 {
		t.Fatalf("expected 2 breadcrumbs but got %d", len(bcs))
	}

	ja := jsonassert.New(t)
	ja.Assertf(asJSON(bcs[0]), `{ "timestamp": "<<PRESENCE>>", "name": "newest", "type": "log", "metaData": { "md": 1 } }`)
	ja.Assertf(asJSON(bcs[1]), `{ "timestamp": "<<PRESENCE>>", "name": "latest", "type": "manual" }`)
}
