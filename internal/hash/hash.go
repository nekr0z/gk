// Package hash holds utilities for hash-handling.
package hash

func SliceToArray(b []byte) [32]byte {
	if len(b) == 0 {
		return [32]byte{}
	}

	if len(b) != 32 {
		panic("invalid hash length")
	}

	var arr [32]byte
	copy(arr[:], b)

	return arr
}
