package version

// These variables are populated by ldflags at build time.
var (
	// Version is the semantic version number.
	Version = "dev"
	
	// Commit is the git commit hash.
	Commit = "none"
	
	// Date is the build date.
	Date = "unknown"
	
	// BuildInfo returns a formatted version string with all build information.
	BuildInfo = func() string {
		return "aid version " + Version
	}
)