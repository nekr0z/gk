package hash_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nekr0z/gk/internal/hash"
)

func TestToHashArray(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var expected [32]byte
		result := hash.SliceToArray(nil)
		assert.Equal(t, expected, result)
	})

	t.Run("valid 32 bytes", func(t *testing.T) {
		b := make([]byte, 32)
		for i := range b {
			b[i] = byte(i)
		}
		var expected [32]byte
		copy(expected[:], b)
		result := hash.SliceToArray(b)
		assert.Equal(t, expected, result)
	})

	t.Run("invalid length", func(t *testing.T) {
		b := make([]byte, 31)
		assert.Panics(t, func() {
			hash.SliceToArray(b)
		})
	})
}
