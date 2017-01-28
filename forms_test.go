// forms_test.go
// -------------
//
// TODO (tricky stuff):
// * prove that Copy keeps all the cached limit stuff.
// * the limit parser

package vebben_test

import (
	// Standard:
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	// Helpers:
	"github.com/biztos/testig"
	"github.com/stretchr/testify/assert"

	// Under test:
	"github.com/biztos/vebben"
)

type SimpleType struct {
	Foo string
}

type LessSimpleType struct {
	OptBool     bool      `json:"opt_bool"`
	ReqBool     bool      `json:"req_bool"`
	OptInt      int       `json:"opt_int"`
	ReqInt      int       `json:"req_int"`
	OptInt64    int64     `json:"opt_int64"`
	ReqInt64    int64     `json:"req_int64"`
	OptFloat    float64   `json:"opt_float"`
	ReqFloat    float64   `json:"req_float"`
	OptString   string    `json:"opt_string"`
	ReqString   string    `json:"req_string"`
	OptDate     time.Time `json:"opt_date"`
	ReqDate     time.Time `json:"req_date"`
	OptDateTime time.Time `json:"opt_datetime"`
	ReqDateTime time.Time `json:"req_datetime"`
	OptDateFlex time.Time `json:"opt_dateflex"`
	ReqDateFlex time.Time `json:"req_dateflex"`
}

type TestFormValuer struct {
	vmap map[string]string
}

func (f *TestFormValuer) FormValue(k string) string {
	return f.vmap[k]
}

func Test_MultiError(t *testing.T) {

	assert := assert.New(t)

	err := vebben.MultiError{[]error{
		errors.New("first"),
		errors.New("second"),
	}}

	assert.Equal("first\nsecond", err.Error(),
		"Error() joins Errors w/newline")
}

func Test_DecodeForm_NoSpecs(t *testing.T) {

	assert := assert.New(t)

	f := &TestFormValuer{map[string]string{}}
	target := &SimpleType{}
	err := vebben.DecodeForm(f, []*vebben.FormSpec{}, target)
	assert.Nil(err, "no error")
	assert.Equal("", target.Foo, "nothing set")
}

func Test_FormSpec_Init(t *testing.T) {

	assert := assert.New(t)

	spec := &vebben.FormSpec{
		Key:   "foo",
		Name:  "Foo",
		Type:  "string",
		Limit: "x=20",
	}
	noPanic := func() { spec.Init() }
	assert.NotPanics(noPanic, "no panic on Init")

}

func Test_FormSpec_Init_BadTypePanics(t *testing.T) {

	spec := &vebben.FormSpec{
		Key:   "foo",
		Name:  "Foo",
		Type:  "NoSuchType",
		Limit: "x=20",
	}
	willPanic := func() { spec.Init() }

	testig.AssertPanicsWith(t, willPanic,
		"Unsupported FormSpec type: NoSuchType",
		"got expected panic")

}

func Test_RequiredFormSpec(t *testing.T) {

	assert := assert.New(t)

	// No limit, no name:
	spec1 := vebben.RequiredFormSpec("foo", "string")
	assert.True(spec1.Required)
	assert.Equal(spec1.Key, "foo")
	assert.Equal(spec1.Type, "string")
	assert.Equal(spec1.Limit, "")
	assert.Equal(spec1.Name, "foo")
	assert.NotNil(spec1.Validator)

	// Limit, no name:
	spec2 := vebben.RequiredFormSpec("foo", "string", "8")
	assert.True(spec2.Required)
	assert.Equal(spec2.Key, "foo")
	assert.Equal(spec2.Type, "string")
	assert.Equal(spec2.Limit, "8")
	assert.Equal(spec2.Name, "foo")
	assert.NotNil(spec2.Validator)

	// Limit and name:
	spec3 := vebben.RequiredFormSpec("foo", "string", "8", "The Foo")
	assert.True(spec3.Required)
	assert.Equal(spec3.Key, "foo")
	assert.Equal(spec3.Type, "string")
	assert.Equal(spec3.Limit, "8")
	assert.Equal(spec3.Name, "The Foo")
	assert.NotNil(spec3.Validator)

}

func Test_OptionalFormSpec(t *testing.T) {

	assert := assert.New(t)

	// No limit, no name:
	spec1 := vebben.OptionalFormSpec("foo", "string")
	assert.False(spec1.Required)
	assert.Equal(spec1.Key, "foo")
	assert.Equal(spec1.Type, "string")
	assert.Equal(spec1.Limit, "")
	assert.Equal(spec1.Name, "foo")
	assert.NotNil(spec1.Validator)

	// Limit, no name:
	spec2 := vebben.OptionalFormSpec("foo", "string", "8")
	assert.False(spec2.Required)
	assert.Equal(spec2.Key, "foo")
	assert.Equal(spec2.Type, "string")
	assert.Equal(spec2.Limit, "8")
	assert.Equal(spec2.Name, "foo")
	assert.NotNil(spec2.Validator)

	// Limit and name:
	spec3 := vebben.OptionalFormSpec("foo", "string", "8", "The Foo")
	assert.False(spec3.Required)
	assert.Equal(spec3.Key, "foo")
	assert.Equal(spec3.Type, "string")
	assert.Equal(spec3.Limit, "8")
	assert.Equal(spec3.Name, "The Foo")
	assert.NotNil(spec3.Validator)

}

func Test_FormSpec_Copy_RequiresKey(t *testing.T) {

	fs := &vebben.FormSpec{}

	willPanic := func() { fs.Copy("", "") }

	testig.AssertPanicsWith(t, willPanic,
		"Empty FormSpec key",
		"got expected panic")

}

func Test_FormSpec_Copy_NameDefaults(t *testing.T) {

	assert := assert.New(t)

	fs := &vebben.FormSpec{}
	fs2 := fs.Copy("second", "")
	assert.Equal("second", fs2.Key, "Key is from new input")
	assert.Equal("second", fs2.Name, "Name is the key")

}

func Test_FormSpec_Copy(t *testing.T) {

	assert := assert.New(t)

	vf := func(f *vebben.FormSpec, v interface{}) error { return nil }
	fs := &vebben.FormSpec{
		Key:       "first",
		Name:      "First",
		Limit:     "some-limit",
		Validator: vf,
	}

	fs2 := fs.Copy("second", "The Second")
	assert.Equal("second", fs2.Key, "Key is from new input")
	assert.Equal("The Second", fs2.Name, "Name is from new input")
	assert.Equal("some-limit", fs2.Limit, "Limit is from the source")

	// Hmm, having some trouble comparing func types here.
	s1 := fmt.Sprintf("%#v", fs.Validator)
	s2 := fmt.Sprintf("%#v", fs2.Validator)
	assert.EqualValues(s1, s2, "Validator func is from the source")

}

func Test_DecodeForm_MissingRequired(t *testing.T) {

	assert := assert.New(t)

	f := &TestFormValuer{map[string]string{"nope": "that"}}
	specs := []*vebben.FormSpec{
		vebben.RequiredFormSpec("foo", "string"),
		vebben.RequiredFormSpec("bar", "int"),
	}

	err := vebben.DecodeForm(f, specs, &SimpleType{})
	if assert.Error(err, "error returned") {
		if assert.IsType(&vebben.MultiError{}, err, "error has correct type") {
			assert.Equal("foo is required\nbar is required",
				err.Error(), "stringified error as expected")
		}
	}

}

func Test_DecodeForm_Success(t *testing.T) {

	assert := assert.New(t)

	f := &TestFormValuer{map[string]string{
		"req_bool":     "true",
		"req_int":      "23456",
		"req_int64":    "20002147483647",
		"req_float":    "2.2345",
		"req_string":   "  another string\n\n\n",
		"req_date":     "2017.02.25",
		"req_datetime": "2017.02.26 10:30",
		"req_dateflex": "2017.02.27 16:30",

		"ignoreme": "anything",
	}}
	specs := []*vebben.FormSpec{
		vebben.OptionalFormSpec("opt_bool", "bool"),
		vebben.RequiredFormSpec("req_bool", "bool"),
		vebben.OptionalFormSpec("opt_int", "int"),
		vebben.RequiredFormSpec("req_int", "int"),
		vebben.OptionalFormSpec("opt_int64", "int64"),
		vebben.RequiredFormSpec("req_int64", "int64"),
		vebben.OptionalFormSpec("opt_float", "float"),
		vebben.RequiredFormSpec("req_float", "float"),
		vebben.OptionalFormSpec("opt_string", "string"),
		vebben.RequiredFormSpec("req_string", "string"),
		vebben.OptionalFormSpec("opt_date", "date"),
		vebben.RequiredFormSpec("req_date", "date"),
		vebben.OptionalFormSpec("opt_datetime", "datetime"),
		vebben.RequiredFormSpec("req_datetime", "datetime"),
		vebben.OptionalFormSpec("opt_dateflex", "dateflex"),
		vebben.RequiredFormSpec("req_dateflex", "dateflex"),
	}

	loc := vebben.FormValueTimeLocation
	d, _ := time.ParseInLocation("2006.01.02.", "2017.02.25.", loc)
	dt, _ := time.ParseInLocation("2006.1.2. 15:04", "2017.02.26. 10:30", loc)
	df, _ := time.ParseInLocation("2006.1.2. 15:04", "2017.02.27. 16:30", loc)
	exp := &LessSimpleType{
		ReqBool:     true,
		ReqInt:      23456,
		ReqInt64:    20002147483647,
		ReqFloat:    2.2345,
		ReqString:   "another string",
		ReqDate:     d,
		ReqDateTime: dt,
		ReqDateFlex: df,
	}

	target := &LessSimpleType{}
	err := vebben.DecodeForm(f, specs, target)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Unfortunately we have trouble comparing the time properties, since
	// they contain locations which are pointers which are different.
	// (This might be a bug in assert; shouldn't EqualValues recurse?)
	// So back to the JSON comparison hack.
	expB, _ := json.Marshal(exp)
	targetB, _ := json.Marshal(target)
	assert.Equal(string(expB), string(targetB),
		"target struct filled to expectation, at least per the JSON")

}
