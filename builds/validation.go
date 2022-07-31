package builds

import (
	"fmt"
	"regexp"
	"strings"
)

func (r *JSONBuildRequest) Validate() error {
	if re := regexp.MustCompile("^[0-9a-f]{32}$"); !re.MatchString(r.APIKey) {
		return fmt.Errorf(`APIKey must be 32 hex characters, but got "%s"`, r.APIKey)
	}

	if strings.TrimSpace(r.AppVersion) == "" {
		return fmt.Errorf(`AppVersion must be present`)
	}

	if r.SourceControl != nil {
		if r.SourceControl.Repository == "" {
			return fmt.Errorf(`SourceControl.Repository must be present when SourceControl is set`)
		}
		if r.SourceControl.Revision == "" {
			return fmt.Errorf(`SourceControl.Revision must be present when SourceControl is set`)
		}
		if r.SourceControl.Provider != "" {
			validProviders := []string{"github", "github-enterprise", "bitbucket", "bitbucket-server", "gitlab", "gitlab-onpremise"}
			found := false
			for _, p := range validProviders {
				found = found || p == r.SourceControl.Provider
			}
			if !found {
				return fmt.Errorf(`SourceControl.Provider must be unset (for automatic inference) or one of %v`, validProviders)
			}
		}
	}
	return nil
}
