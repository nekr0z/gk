// Package version displays the version of the software.
package version

import "fmt"

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

// String returns the string representation of the version.
func String() string {
	return fmt.Sprintf("%s built on %s", buildVersion, buildDate)
}
