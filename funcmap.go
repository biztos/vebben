// funcmap.go - functions for Templates, and a map thereof.
// ----------
// NOTE: this is now part of the vebben package (misc. web code).

package vebben

import (
	"html/template"
	"sort"
	"strings"

	// Third-party:
	"github.com/leekchan/gtf"
)

// NewFuncMap returns a function map containing all the available functions,
// mapped to names as shown below.  This should be included in
// a template, usually the master template, as:
//  fm := vebben.NewFuncMap()
//  tmpl, err := template.New("").Funcs(fm).Parse(MyTemplateData)
//
// Functions Defined Here
//
// These correspond to the exported function names.  By convention(?) the
// map keys, i.e. the function names called in the template, are lowercased.
//
//     truncints
//     truncintsto
//     sortints
//     sortintsasc
//     sortintsdesc
//     reverseints
//     intrange
//     pathdepth
//     indent
//
// Functions From Kyoung-chan Lee's Gtf
//
// See https://github.com/leekchan/gtf for documentation.
//
//     replace
//     default
//     length
//     lower
//     upper
//     truncatechars
//     urlencode
//     wordcount
//     divisibleby
//     lengthis
//     trim
//     capfirst
//     pluralize
//     yesno
//     rjust
//     ljust
//     center
//     filesizeformat
//     apnumber
//     intcomma
//     ordinal
//     first
//     last
//     join
//     slice
//     random
//     striptags
func NewFuncMap() template.FuncMap {

	fm := template.FuncMap{

		// General-Purpose Custom Functions:
		"truncints":    TruncInts,
		"truncintsto":  TruncIntsTo,
		"sortints":     SortInts,
		"sortintsasc":  SortInts, // for the literal-minded
		"sortintsdesc": SortIntsDesc,
		"reverseints":  ReverseInts,
		"intrange":     IntRange,
		"pathdepth":    PathDepth,
		"indent":       Indent,

		// Golang Standard Functions:
		// TODO (maybe)

	}

	// Thank You Kyoung-chan Lee!
	gtf.Inject(fm)

	return fm
}

// TruncInts returns a slice of the input slice with its final element
// removed.
func TruncInts(i []int) []int {
	if len(i) < 2 {
		return []int{}
	}
	return i[:len(i)-1]
}

// TruncIntsTo returns a slice of the input slice with a maximum of t
// elements.  If t is negative, truncates to len(i) - t.
// Note the argument order: this is to facilitate piping in the template:
//   {{ $r := .SomeIntSlice | truncintsto 5 }}
func TruncIntsTo(t int, i []int) []int {

	if t < 0 {
		t += len(i)
	}
	if len(i) == 0 || t <= 0 {
		return []int{}
	}
	if len(i) <= t {
		return i
	}
	return i[:t]

}

// ReverseInts returns a copy of the input slice in reverse order, i.e. the
// opposite of the given order.
func ReverseInts(ii []int) []int {
	length := len(ii)
	res := make([]int, length)
	for idx, i := range ii {
		res[length-idx-1] = i
	}
	return res
}

// SortInts returns a sorted (Ascending) copy of a slice of integers.
func SortInts(ii []int) []int {
	res := make([]int, len(ii))
	copy(res, ii)
	sort.Ints(res)
	return res
}

// SortIntsDesc returns a reverse-sorted (Descending) copy of a slice of
// integers.
func SortIntsDesc(ii []int) []int {
	res := make([]int, len(ii))
	copy(res, ii)
	sort.Sort(sort.Reverse(sort.IntSlice(res)))
	return res
}

// IntRange returns an integer slice containing the range from a to b
// inclusive.
func IntRange(a, b int) []int {
	if a == b {
		return []int{a}
	}
	r := []int{}
	if b > a {
		for i := a; i <= b; i++ {
			r = append(r, i)
		}
	} else {
		for i := a; i >= b; i-- {
			r = append(r, i)
		}
	}
	return r
}

// PathDepth returns the depth of a cleaned URL path (or similar string),
// i.e. the number of slashes it contains.
func PathDepth(p string) int {
	return strings.Count(string(p), "/")
}

// Indent returns a string with four spaces repeated up to the given depth.
func Indent(depth int) string {
	return strings.Repeat("    ", depth)
}
