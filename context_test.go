package bugsnag

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kinbiko/jsonassert"
)

func TestContextWithMethods(t *testing.T) {
	n, err := New(Configuration{APIKey: "1234abcd1234abcd1234abcd1234abcd", AppVersion: "1.2.3", ReleaseStage: "test"})
	if err != nil {
		t.Fatal(err)
	}
	t.Run("WithBreadcrumb", func(t *testing.T) {
		asJSON := func(s interface{}) string {
			b, err := json.Marshal(s)
			if err != nil {
				return "<<ERROR>>"
			}
			return string(b)
		}

		expLatest := Breadcrumb{Name: "latest"}
		expNewest := Breadcrumb{Name: "newest", Type: BCTypeLog, Metadata: map[string]interface{}{"md": 1}}

		ctx := n.WithBreadcrumb(context.Background(), expLatest)
		ctx = n.WithBreadcrumb(ctx, expNewest)

		bcs := makeBreadcrumbs(ctx)

		if len(bcs) != 2 {
			t.Fatalf("expected 2 breadcrumbs but got %d", len(bcs))
		}

		ja := jsonassert.New(t)
		ja.Assertf(asJSON(bcs[0]), `{ "timestamp": "<<PRESENCE>>", "name": "newest", "type": "log", "metaData": { "md": 1 } }`)
		ja.Assertf(asJSON(bcs[1]), `{ "timestamp": "<<PRESENCE>>", "name": "latest", "type": "manual" }`)
	})

	t.Run("WithUser", func(t *testing.T) {
		exp := User{ID: "id", Name: "name", Email: "email"}
		if got := getAttachedContextData(n.WithUser(context.Background(), exp)).User; *got != exp {
			t.Errorf("expected that when I add '%+v' to the context what I get back ('%+v') should be equal", exp, got)
		}
	})

	t.Run("WithMetadata and WithMetadatum", func(t *testing.T) {
		ctx := n.WithMetadatum(context.Background(), "app", "id", "15011-2")
		ctx = n.WithMetadata(ctx, "device", map[string]interface{}{"model": "15023-2"})
		md := n.Metadata(ctx)
		if appID, exp := md["app"]["id"], "15011-2"; appID != exp {
			t.Errorf("expected app.id to be '%s' but was '%s'", exp, appID)
		}
		if deviceModel, exp := md["device"]["model"], "15023-2"; deviceModel != exp {
			t.Errorf("expected device.model to be '%s' but was '%s'", exp, deviceModel)
		}
	})
}

func TestCtxSerialization(t *testing.T) {
	n, err := New(Configuration{APIKey: "1234abcd1234abcd1234abcd1234abcd", AppVersion: "1.2.3", ReleaseStage: "test"})

	ctx := context.Background()
	ctx = n.WithBreadcrumb(ctx, Breadcrumb{Name: "log event", Type: BCTypeLog, Metadata: map[string]interface{}{"msg": "ruh roh"}})
	ctx = n.WithMetadata(ctx, "app", map[string]interface{}{"nick": "charmander"})
	ctx = n.WithMetadatum(ctx, "app", "types", []string{"fire"})
	ctx = n.WithUser(ctx, User{ID: "qwpeoiub", Name: "charlie", Email: "charlie@pokemon.example.com"})
	ctx = n.WithBugsnagContext(ctx, "/pokemon?type=fire")

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
		if err != nil {
			t.Fatal(err)
		}
		ctx = n.Deserialize(context.Background(), n.Serialize(ctx))
		b, _ := json.Marshal(getAttachedContextData(ctx))
		jsonassert.New(t).Assertf(string(b), `{
			"cx": "/pokemon?type=fire",
			"bc":[{"md":{"msg": "ruh roh"}, "ts":"<<PRESENCE>>", "na":"log event", "ty":4}],
			"us":{"email":"charlie@pokemon.example.com", "id":"qwpeoiub", "name":"charlie"},
			"md":{"app": {"nick": "charmander", "types": ["fire"]}}
		}`)
	})
}
