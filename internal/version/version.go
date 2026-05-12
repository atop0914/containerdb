// Package version provides version information for ContainerDB.
package version

import "fmt"

// These are set at build time via ldflags.
var (
	// Version is the current release version.
	Version = "1.0.0"
	// GitCommit is the git commit hash.
	GitCommit = "unknown"
	// BuildDate is the date the binary was built.
	BuildDate = "unknown"
)

// Info returns a formatted version string.
func Info() string {
	return fmt.Sprintf("containerdb %s (commit: %s, built: %s)", Version, GitCommit, BuildDate)
}
