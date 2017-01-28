// utils_test.go
// -------------

package vebben_test

import (
	// Standard:
	"testing"

	// Helpers:
	"github.com/stretchr/testify/assert"

	// Under test:
	"github.com/biztos/vebben"
)

func Test_GlyphLength(t *testing.T) {

	assert := assert.New(t)

	lengths := map[string]int{
		"műemlék": 7,
		"foo":     3,
		"啤酒":      2,
	}
	for s, length := range lengths {
		assert.Equal(length, vebben.GlyphLength(s),
			"correct length for %s: %d", s, length)

	}
}
