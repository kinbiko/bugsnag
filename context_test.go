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

func TestUserAndContext(t *testing.T) {
	exp := User{ID: "id", Name: "name", Email: "email"}
	if got := makeUser(WithUser(context.Background(), exp)); *got != exp {
		t.Errorf("expected that when I add '%+v' to the context what I get back ('%+v') should be equal", exp, got)
	}
}

func TestMetadata(t *testing.T) {
	ctx := WithMetadatum(context.Background(), "app", "id", "15011-2")
	ctx = WithMetadata(ctx, "device", map[string]interface{}{"model": "15023-2"})
	md := Metadata(ctx)
	if appID, exp := md["app"]["id"], "15011-2"; appID != exp {
		t.Errorf("expected app.id to be '%s' but was '%s'", exp, appID)
	}
	if deviceModel, exp := md["device"]["model"], "15023-2"; deviceModel != exp {
		t.Errorf("expected device.model to be '%s' but was '%s'", exp, deviceModel)
	}
}
