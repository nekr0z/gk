// Package i18ninit provides i18n initialization.
package i18n

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/Xuanwo/go-locale"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

//go:embed *.yaml
var translations embed.FS

func NewLocalizer(cmd *cobra.Command) *i18n.Localizer {
	bundle := prepareI18n()

	lang, err := locale.Detect()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Failed to detect locale: %s.", err)
	}

	return i18n.NewLocalizer(bundle, lang.String())
}

func prepareI18n() *i18n.Bundle {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	bundle.MustAddMessages(language.English, messages...)
	if files, err := translations.ReadDir("."); err == nil {
		for _, file := range files {
			if ok, err := filepath.Match("*.yaml", file.Name()); ok && err == nil {
				bundle.LoadMessageFileFS(translations, file.Name())
			}
		}
	}
	return bundle
}
