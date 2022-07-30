package builds

// Payload defines the complete payload to Bugsnag's build API as defined here:
// https://bugsnagbuildapi.docs.apiary.io/#introduction/matching-error-events-and-sessions-to-builds
type Payload struct {
	// The notifier API key of the project.
	APIKey string `json:"apiKey"` // E.g. "1234abcd1234abcd1234abcd1234abcd"

	// The version number of the application.
	// This is used to identify the particular version of the application that
	// the build relates to.
	AppVersion string `json:"appVersion"` // E.g. "1.5.2"

	//The release stage (eg, production, staging) that is being released (if
	//applicable). Normally the fact that a build has been released to a
	//release stage is detected automatically when an error event or session is
	//received for the build. However if you would like to manually notify
	//Bugsnag of the build being released you can specify the stage that the
	//build was released to.
	ReleaseStage string `json:"releaseStage,omitempty"` // E.g. "staging"

	// The name of the entity that triggered the build. Could be a user,
	// system, etc.
	BuilderName string `json:"builderName,omitempty"` // E.g. "River Tam"

	// Key value pairs containing any custom build information that provides
	// useful metadata about the build. e.g. build configuration parameters,
	// versions of dependencies, reason for the build etc.
	Metadata map[string]string `json:"metadata,omitempty"` // E.g. map[string]string{"buildServer": "build1", "buildReason": "Releasing JIRA-1234"}

	// Information about the source control of the code. This can be used to
	// link errors to the source code (for supported source control tools)
	SourceControl *SourceControl `json:"sourceControl,omitempty"`

	// The version code of the application (Android only).
	// For Android apps if no code is provided Bugsnag will associate the build
	// information with the most recent build for the app version.
	AppVersionCode int `json:"appVersionCode,omitempty"` // E.g. 1234

	// The bundle version/build number of the application (iOS/macOS/tvOS only).
	// For iOS/macOS/tvOS apps if no bundle version is provided we will
	// associate the build information with the most recent build for the app
	// version.
	AppBundleVersion string `json:"appBundleVersion,omitempty"` // E.g. "1.2.3"

	// Flag indicating whether to automatically associate this build with any
	// new error events and sessions that are received for the releaseStage
	// until a subsequent build notification is received for the release stage
	// If this is set to true and no releaseStage is provided the build will be
	// applied to production. Automatically assigning builds to new error
	// events is generally discouraged as it can result events from previous
	// builds being incorrectly recorded against a new build while builds are
	// being rolled out.
	AutoAssignRelease bool `json:"autoAssignRelease,omitempty"`
}

// Information about the source control of the code. This can be used to
// link errors to the source code (for supported source control tools)
type SourceControl struct {
	// If the provider can be inferred from the repository then it is not
	// required.
	Provider Provider `json:"provider,omitempty"`

	// Repository represents the URL of the repository containing the source
	// code being deployed.
	Repository string `json:"repository"`

	// Revision is the source control SHA-1 hash for the code that has been
	// built (short or long hash).
	Revision string `json:"revision"`
}

// Provider is the name of the source control provider that contains the
// source code for the build.
type Provider string

// ToProvider converts a string into a Provider, if valid, otherwise returns "".
func ToProvider(s string) Provider {
	return map[string]Provider{
		"github":            ProviderGitHub,
		"github-enterprise": ProviderGitHubEnterprise,
		"bitbucket":         ProviderBitbucket,
		"bitbucket-server":  ProviderBitbucketServer,
		"gitlab":            ProviderGitLab,
		"gitlab-onpremise":  ProviderGitLabOnPremise,
	}[s]
}

const (
	// ProviderGitHub indicates that your source code is stored on GitHub's
	// SaaS platform.
	ProviderGitHub = "github"
	// ProviderGitHubEnterprise indicates that your source code is stored on
	// GitHub's enterprise platform.
	ProviderGitHubEnterprise = "github-enterprise"
	// ProviderBitbucket indicates that your source code is stored on
	// Bitbucket's SaaS platform. Git and Mercurial are supported for
	// Bitbucket.
	ProviderBitbucket = "bitbucket"
	// ProviderBitbucketServer indicates that your source code is stored on
	// Bitbucket's Server platform. Formerly known as Stash.
	ProviderBitbucketServer = "bitbucket-server"
	// ProviderGitLab indicates that your source code is stored on
	// GitLab's SaaS platform.
	ProviderGitLab = "gitlab"
	// ProviderGitLabOnPremise indicates that your source code is stored on
	// GitLab's GitLab CE or GitLab Enterprise platform.
	ProviderGitLabOnPremise = "gitlab-onpremise"
)
