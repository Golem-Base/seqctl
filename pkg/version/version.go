package version

var (
	// Version is the version of the application
	Version = "v0.1.0"

	// GitCommit is the git commit that was compiled
	GitCommit = ""

	// GitDate is the date of the git commit that was compiled
	GitDate = ""
)

// FullVersionInfo returns a formatted string with version information
func FullVersionInfo() string {
	return Version + "-" + GitCommit + "-" + GitDate
}

// Info returns a formatted string with version information
func Info() string {
	return Version
}
