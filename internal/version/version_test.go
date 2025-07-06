package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
