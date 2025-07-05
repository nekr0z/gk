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
