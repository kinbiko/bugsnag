package builds_test

import (
	"testing"
)

func TestValidate(t *testing.T) {
	t.Run("API key", func(t *testing.T) {
		r := makeBigValidReq()
		r.APIKey = ""
		mustContain(t, r.Validate(), "APIKey", "32 hex")

		r = makeBigValidReq()
		r.APIKey = "1234567890"
		mustContain(t, r.Validate(), "APIKey", "32 hex")

		r = makeBigValidReq()
		r.APIKey = "123456789012345678901234567890ZZ"
		mustContain(t, r.Validate(), "APIKey", "32 hex")
	})

	t.Run("App version", func(t *testing.T) {
		r := makeBigValidReq()
		r.AppVersion = ""
		mustContain(t, r.Validate(), "AppVersion", "present")
	})

	t.Run("Source control", func(t *testing.T) {
		r := makeBigValidReq()
		r.SourceControl.Repository = ""
		mustContain(t, r.Validate(), "SourceControl.Repository", "present when")

		r = makeBigValidReq()
		r.SourceControl.Revision = ""
		mustContain(t, r.Validate(), "SourceControl.Revision", "present when")

		t.Run("Provider", func(t *testing.T) {
			r := makeBigValidReq()
			r.SourceControl.Provider = "sourceforge"
			mustContain(t, r.Validate(), "SourceControl.Provider", "unset", "automatic", "one of", "server")
		})
	})

	t.Run("perfectly valid", func(t *testing.T) {
		t.Run("all values populated", func(t *testing.T) {
			if err := makeBigValidReq().Validate(); err != nil {
				t.Errorf("unexpected error: %v", err)
			}

		})
		t.Run("only bare-minimum populated", func(t *testing.T) {
			if err := makeSmallValidReq().Validate(); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	})
}
