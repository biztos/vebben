// vebben_test.go - test the template-accessible functions.
// ---------------

package vebben_test

import (
	// Standard:
	"testing"

	// Third-party:
	"github.com/stretchr/testify/assert"

	// Under test:
	"github.com/biztos/vebben"
)

func Test_New(t *testing.T) {

	assert := assert.New(t)

	fm := vebben.NewFuncMap()

	exp_funcs := []string{
		// products of our own tortured logic:
		"truncints",
		"truncintsto",
		"sortints",
		"sortintsasc",
		"sortintsdesc",
		"reverseints",
		"intrange",
		"pathdepth",
		"indent",

		// gtf freebies:
		"replace",
		"default",
		"length",
		"lower",
		"upper",
		"truncatechars",
		"urlencode",
		"wordcount",
		"divisibleby",
		"lengthis",
		"trim",
		"capfirst",
		"pluralize",
		"yesno",
		"rjust",
		"ljust",
		"center",
		"filesizeformat",
		"apnumber",
		"intcomma",
		"ordinal",
		"first",
		"last",
		"join",
		"slice",
		"random",
		"striptags",
	}

	for _, s := range exp_funcs {

		assert.NotNil(fm[s], "func defined for "+s)
	}
}

func Test_TruncInts(t *testing.T) {

	assert := assert.New(t)

	assert.Equal([]int{}, vebben.TruncInts([]int{}),
		"empty truncates to empty slice")

	assert.Equal([]int{}, vebben.TruncInts([]int{4}),
		"single truncates to empty slice")

	assert.Equal([]int{3}, vebben.TruncInts([]int{3, 2}),
		"two elements truncate to one")

	// This *should* be more efficient than dealing with array copies.
	i := []int{22, 33, 44}
	r := vebben.TruncInts(i)
	i[1] = 999
	assert.Equal(999, r[1], "slice is of same underlying array")

}

func Test_TruncIntsTo(t *testing.T) {

	assert := assert.New(t)

	assert.Equal([]int{}, vebben.TruncIntsTo(1, []int{}),
		"empty truncates to empty slice")

	assert.Equal([]int{4}, vebben.TruncIntsTo(1, []int{4}),
		"single truncates to single at 1")

	assert.Equal([]int{}, vebben.TruncIntsTo(0, []int{4}),
		"single truncates to nothing at 0")

	assert.Equal([]int{3}, vebben.TruncIntsTo(1, []int{3, 2}),
		"two elements truncate to one at 1")

	assert.Equal([]int{3, 2}, vebben.TruncIntsTo(2, []int{3, 2, 1}),
		"three elements truncate to two at 2")

	assert.Equal([]int{3, 2}, vebben.TruncIntsTo(-1, []int{3, 2, 1}),
		"three elements truncate to two at -1")

	assert.Equal([]int{3}, vebben.TruncIntsTo(-2, []int{3, 2, 1}),
		"three elements truncate to one at -2")

	assert.Equal([]int{3}, vebben.TruncIntsTo(-2, []int{3, 2, 1}),
		"three elements truncate to zero at -3")

	assert.Equal([]int{3}, vebben.TruncIntsTo(-2, []int{3, 2, 1}),
		"three elements truncate to zero at -4")

	// This *should* be more efficient than dealing with array copies.
	// i := []int{22, 33, 44}
	// r := vebben.TruncIntsTo(i, 2)
	// i[1] = 999
	// assert.Equal(999, r[1], "slice is of same underlying array")

}

func Test_ReverseInts(t *testing.T) {

	assert := assert.New(t)

	assert.Equal([]int{}, vebben.ReverseInts([]int{}),
		"empty stays empty")

	assert.Equal([]int{1, 2, 3, 4}, vebben.ReverseInts([]int{4, 3, 2, 1}),
		"ordered reverses")

	assert.Equal([]int{9, 1, 3, 1, 9}, vebben.ReverseInts([]int{9, 1, 3, 1, 9}),
		"result is not sorted")

	// Prove we copy.
	i := []int{22, 33, 44}
	r := vebben.ReverseInts(i)
	i[1] = 999
	assert.Equal(33, r[1], "return slice is a copy")

}

func Test_SortInts(t *testing.T) {

	assert := assert.New(t)

	assert.Equal([]int{}, vebben.SortInts([]int{}),
		"empty stays empty")

	assert.Equal([]int{1, 2, 3, 4}, vebben.SortInts([]int{4, 3, 2, 1}),
		"ordered reverses")

	assert.Equal([]int{1, 1, 3, 9, 9}, vebben.SortInts([]int{9, 1, 3, 1, 9}),
		"result is sorted")

	// Prove we copy.
	i := []int{22, 33, 44}
	r := vebben.SortInts(i)
	i[1] = 999
	assert.Equal(33, r[1], "return slice is a copy")

}

func Test_SortIntsDesc(t *testing.T) {

	assert := assert.New(t)

	assert.Equal([]int{}, vebben.SortIntsDesc([]int{}),
		"empty stays empty")

	assert.Equal([]int{4, 3, 2, 1}, vebben.SortIntsDesc([]int{1, 2, 3, 4}),
		"ordered reverses")

	assert.Equal([]int{9, 9, 3, 1, 1}, vebben.SortIntsDesc([]int{9, 1, 3, 1, 9}),
		"result is sorted")

	// Prove we copy.
	i := []int{22, 33, 44}
	r := vebben.SortIntsDesc(i)
	i[1] = 999
	assert.Equal(33, r[1], "return slice is a copy")

}

func Test_IntRange(t *testing.T) {

	assert := assert.New(t)

	assert.Equal([]int{4}, vebben.IntRange(4, 4),
		"IntRange(4,4)-> 4")

	assert.Equal([]int{1, 2, 3, 4}, vebben.IntRange(1, 4),
		"IntRange(1,4)-> 1 2 3 4")

	assert.Equal([]int{4, 3, 2, 1}, vebben.IntRange(4, 1),
		"IntRange(4,1) -> 4 3 2 1")

	assert.Equal([]int{-2, -1, 0, 1}, vebben.IntRange(-2, 1),
		"IntRange(-2, 1) -> -2 -1 0 1")

	assert.Equal([]int{0, -1, -2, -3}, vebben.IntRange(0, -3),
		"IntRange(0, -3) -> 0 -1 -2 -3")

}

func Test_PathDepth(t *testing.T) {

	assert := assert.New(t)

	assert.Equal(0, vebben.PathDepth("no slashes here"),
		"no slashes -> zero")

	assert.Equal(1, vebben.PathDepth("/standard"),
		"leading slashe -> one")

	assert.Equal(4, vebben.PathDepth("/fee/fi/foe/fum"),
		"multiple")
}

func Test_Indent(t *testing.T) {

	assert := assert.New(t)

	assert.Equal("", vebben.Indent(0),
		"zero -> empty")

	assert.Equal("    ", vebben.Indent(1),
		"one -> 4 sp.")

	assert.Equal("            ", vebben.Indent(3),
		"three -> 4 sp.")
}
