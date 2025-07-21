package version

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/nekr0z/gk/internal/i18n"
)

func TestString(t *testing.T) {
	buildDate = "2025-07-05"
	buildVersion = "1.2.3"

	want := "1.2.3 built on 2025-07-05"
	got := String()
	assert.Equal(t, want, got)
}

func TestString_DefaultValues(t *testing.T) {
	buildVersion = "N/A"
	buildDate = "N/A"
	assert.Equal(t, "N/A built on N/A", String())
}

func TestString_SpecialCharacters(t *testing.T) {
	buildVersion = "v1.0.0+beta"
	buildDate = "2025-07-06T12:00:00Z"
	assert.Equal(t, "v1.0.0+beta built on 2025-07-06T12:00:00Z", String())
}

func TestLocalizedString(t *testing.T) {
	buildDate = "2025-07-05"
	buildVersion = "1.2.3"

	cmd := &cobra.Command{
		Use: "cmd",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}

	os.Setenv("LANGUAGE", "en")
	loc := i18n.NewLocalizer(cmd)

	assert.Equal(t, String(), LocalizedString(loc))

	os.Setenv("LANGUAGE", "ru")
	loc = i18n.NewLocalizer(cmd)

	assert.NotEqual(t, String(), LocalizedString(loc))
}
