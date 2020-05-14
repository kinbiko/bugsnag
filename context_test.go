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
	if got := getAttachedContextData(WithUser(context.Background(), exp)).User; *got != exp {
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

func TestCtxSerialization(t *testing.T) {
	ctx := context.Background()
	ctx = WithBreadcrumb(ctx, Breadcrumb{Name: "log event", Type: BCTypeLog, Metadata: map[string]interface{}{"msg": "ruh roh"}})
	ctx = WithMetadata(ctx, "app", map[string]interface{}{"nick": "charmander"})
	ctx = WithMetadatum(ctx, "app", "types", []string{"fire"})
	ctx = WithUser(ctx, User{ID: "qwpeoiub", Name: "charlie", Email: "charlie@pokemon.example.com"})
	ctx = WithBugsnagContext(ctx, "/pokemon?type=fire")

	t.Run("json serialization", func(t *testing.T) {
		b, _ := json.Marshal(getAttachedContextData(ctx))
		jsonassert.New(t).Assertf(string(b), `{
			"cx": "/pokemon?type=fire",
			"bc":[{"md":{"msg": "ruh roh"}, "ts":"<<PRESENCE>>", "na":"log event", "ty":4}],
			"us":{"email":"charlie@pokemon.example.com", "id":"qwpeoiub", "name":"charlie"},
			"md":{"app": {"nick": "charmander", "types": ["fire"]}}
		}`)
	})

	t.Run("serialize + deserialize yields the same data", func(t *testing.T) {
		ctx = Deserialize(context.Background(), Serialize(ctx))
		b, _ := json.Marshal(getAttachedContextData(ctx))
		jsonassert.New(t).Assertf(string(b), `{
			"cx": "/pokemon?type=fire",
			"bc":[{"md":{"msg": "ruh roh"}, "ts":"<<PRESENCE>>", "na":"log event", "ty":4}],
			"us":{"email":"charlie@pokemon.example.com", "id":"qwpeoiub", "name":"charlie"},
			"md":{"app": {"nick": "charmander", "types": ["fire"]}}
		}`)
	})
}
