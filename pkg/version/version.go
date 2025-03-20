package version

var (
	// Version is the current version of the application.
	// This is set during build using -ldflags.
	Version = "dev"

	// GitCommit is the git commit hash of the current version.
	// This is set during build using -ldflags.
	GitCommit = "none"

	// BuildTime is the time when the binary was built.
	// This is set during build using -ldflags.
	BuildTime = "unknown"
)
