package bugsnag

import "testing"

func TestConfigurationValidation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		expMsg string
		cfg    Configuration
	}{
		{
			name: "valid",
			cfg: Configuration{
				APIKey:           "b1234590abcabcabcabcddddddddabcd",
				EndpointNotify:   "http://localhost:8080",
				EndpointSessions: "https://sessions.bugsnag.com",
			},
		},
		{
			name: "API key contains non-hex chars",
			cfg: Configuration{
				APIKey:           "q1234590abcabcabcabcddddddddabcd",
				EndpointNotify:   "http://localhost:8080",
				EndpointSessions: "https://sessions.bugsnag.com",
			},
			expMsg: `API key must be 32 hex characters, but got "q1234590abcabcabcabcddddddddabcd"`,
		},
		{
			name: "API key too short",
			cfg: Configuration{
				APIKey:           "01234590",
				EndpointNotify:   "http://localhost:8080",
				EndpointSessions: "https://sessions.bugsnag.com",
			},
			expMsg: `API key must be 32 hex characters, but got "01234590"`,
		},
		{
			name: "API key too long",
			cfg: Configuration{
				APIKey:           "12345678901234567890123456789012345678901234567890",
				EndpointNotify:   "http://localhost:8080",
				EndpointSessions: "https://sessions.bugsnag.com",
			},
			expMsg: `API key must be 32 hex characters, but got "12345678901234567890123456789012345678901234567890"`,
		},
		{
			name: "notify endpoint not a url",
			cfg: Configuration{
				APIKey:           "b1234590abcabcabcabcddddddddabcd",
				EndpointNotify:   "fluff",
				EndpointSessions: "https://sessions.bugsnag.com",
			},
			expMsg: `notify endpoint be a valid URL, got "fluff"`,
		},
		{
			name: "notify endpoint not a url",
			cfg: Configuration{
				APIKey:           "b1234590abcabcabcabcddddddddabcd",
				EndpointNotify:   "https://notify.bugsnag.com",
				EndpointSessions: "fluff",
			},
			expMsg: `sessions endpoint be a valid URL, got "fluff"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.validate()
			if err == nil {
				if tc.expMsg != "" {
					t.Fatalf("expected error message '%s' but didn't get any errors", tc.expMsg)
				}
				return
			}
			if err.Error() != tc.expMsg {
				t.Errorf("expected error message '%s' but got '%s'", tc.expMsg, err.Error())
			}
		})
	}
}
