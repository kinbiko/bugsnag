package builds

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Validate checks that the given JSONBuildRequest is valid.
// If it is not valid, an error is returned.
func (r *JSONBuildRequest) Validate() error {
	if re := regexp.MustCompile("^[0-9a-f]{32}$"); !re.MatchString(r.APIKey) {
		return fmt.Errorf(`APIKey must be 32 hex characters, but got "%s"`, r.APIKey)
	}

	if strings.TrimSpace(r.AppVersion) == "" {
		return errors.New(`AppVersion must be present`)
	}

	if r.SourceControl == nil {
		return nil
	}

	if r.SourceControl.Repository == "" {
		return errors.New(`SourceControl.Repository must be present when SourceControl is set`)
	}
	if r.SourceControl.Revision == "" {
		return errors.New(`SourceControl.Revision must be present when SourceControl is set`)
	}
	if provider := r.SourceControl.Provider; provider != "" {
		if err := validateSourceControlProvider(provider); err != nil {
			return err
		}
	}
	return nil
}

func validateSourceControlProvider(provider string) error {
	validProviders := []string{"github", "github-enterprise", "bitbucket", "bitbucket-server", "gitlab", "gitlab-onpremise"}
	found := false
	for _, p := range validProviders {
		found = found || p == provider
	}
	if !found {
		return fmt.Errorf(`SourceControl.Provider must be unset (for automatic inference) or one of %v`, validProviders)
	}
	return nil
}
