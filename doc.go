// Package bugsnag exposes an alternative implementation of the Bugsnag Go notifier.
//
// Create a new *bugsnag.Notifier:
//	n, err := bugsnag.New(bugsnag.Configuration{
//		APIKey:       "<<YOUR API KEY HERE>>",
//		AppVersion:   "1.3.5", // Some semver
//		ReleaseStage: "production",
//	})
//	if err != nil {
//		panic(err)
//	}
//	defer n.Close() // Close once you know there are no further calls to n.Notify or n.StartSession.
// Note: There are other configuration options you may set, which enable some
// advanced and very powerful features of the Notifier. See
// bugsnag.Configuration for more information.
//
// After creating the notifier, you  can then report any errors that appear in
// your application:
//	n.Notify(ctx, err)
// You can attach a lot of useful data to context.Context instances that get
// reflected in your dashboard. See the various With*(ctx, ...) funcs for more
// details.
// You can also attach a stacktrace to your errors by calling
//	err = Wrap(ctx, err)
// to wrap another error.
// Another benefit of using Wrap is that the diagnostic data attached to the
// ctx given in the call to Wrap is preserved in the returned err. Meaning that
// you can call n.Notify in a single location without worrying about losing the
// location of the original error or the diagnostic data that was available at
// that point in time.
// If you are reporting a panic or errors you were unable to handle, then you
// may set the Panic and Unhandled fields of the *bugsnag.Error returned from
// Wrap.
//
// To start using session tracking to enable your stability score calculation
// you should call n.StartSession on a per-request basis.
//	ctx = n.StartSession(ctx)
package bugsnag
