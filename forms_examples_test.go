// forms_examples_test.go
// ----------------------

package vebben_test

import (
	"fmt"
	"net/http"

	"github.com/biztos/vebben"
)

func ExampleDecodeForm() {

	type ExampleFoo struct {
		Foo string `json:"foo"`
		Bar int    `json:"bar"`
	}

	specs := []*vebben.FormSpec{
		vebben.RequiredFormSpec("foo", "string", "6", "Foo"),
		vebben.RequiredFormSpec("bar", "int", "0-100", "Bar Percentage"),
	}

	target := &ExampleFoo{}

	// A validation failure:
	r, _ := http.NewRequest("GET", "/example?foo=bar", nil)
	err := vebben.DecodeForm(r, specs, target)
	if err != nil {
		fmt.Println(err.Error())
	}

	// And a success:
	r2, _ := http.NewRequest("GET", "/example?foo=Bärfuß&bar=23", nil)
	err = vebben.DecodeForm(r2, specs, target)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(target.Foo, target.Bar)

	// Output:
	// Foo has the wrong length
	// Bar Percentage is required
	// Bärfuß 23
}
