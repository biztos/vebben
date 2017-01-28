// utils.go -- misc functions used elsewhere.
// --------

package vebben

import (
	"golang.org/x/text/unicode/norm"
)

// GlyphLength returns the number of NFKD-normalized Unicode glyphs in a
// utf-8 enconded string.  For more information see here:
// http://unicode.org/reports/tr15/#Norm_Forms and here:
// http://stackoverflow.com/a/12668840
//
//   fmt.Println(GylphLength("műemlék")) // prints 7
//
// This is used in DecodeForm but is also handy for length-checking any
// international input.
//
// DEPRECATION WARNING
//
// It appears at first glance that one gets exactly the same result from this,
// which is presumably much faster:
//    len([]rune(s))
//
// If that turns out to be the case, GlyphLength will be deprecated at some
// point. (Per the stackoverflow link above it's not, but the evidence given
// might no longer hold.)
func GlyphLength(s string) int {

	var iter norm.Iter
	iter.InitString(norm.NFKD, s)
	count := 0
	for !iter.Done() {
		count += 1
		iter.Next()
	}
	return count

}
