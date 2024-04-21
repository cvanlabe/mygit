package main

import (
	"crypto/sha1"
	"fmt"
)

// findNullByteIndex goes and find the first location in a byte-array
// where we find a null-byte '0'.
//
// If it can't be found, we say it's at the end.
func findNullByteIndex(data []byte) int {
	for i, v := range data {
		if v == 0 {
			return i
		}
	}
	return len(data)
}

// sha1Hash returns a sha1 hash of a byte-array as a byte-array
func sha1Hash(text []byte) []byte {
	hash := fmt.Sprintf("%x", sha1.Sum(text))

	return []byte(hash)
}
