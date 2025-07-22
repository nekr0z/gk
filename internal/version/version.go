// Package version displays the version of the software.
package version

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

// LocalizedString returns the string representation of the version.
func LocalizedString(localizer *i18n.Localizer) string {
	return localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "version",
			Other: "{{.Version}} built on {{.Date}}",
		},
		TemplateData: map[string]interface{}{
			"Version": buildVersion,
			"Date":    buildDate,
		},
	})
}

// String returns the string representation of the version.
func String() string {
	return fmt.Sprintf("%s built on %s", buildVersion, buildDate)
}
