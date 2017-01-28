// forms.go --  form-handling code
// --------

package vebben

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// By default, trim space from form values.
var DecodeFormTrimSpace = true

// FormValueTimeLocation is the location (time zone) used for all form input.
var FormValueTimeLocation, _ = time.LoadLocation("CET")

// DateFormats holds the date formats we accept in forms (note: not times,
// just dates!)
var DateFormats = []string{
	"2006. 01. 02.",
	"2006. 01. 02",
	"2006. 1. 2.",
	"2006. 1. 2",
	"2006.01.02.",
	"2006.01.02",
	"2006.1.2.",
	"2006.1.2",
	"2006-01-02",
	"2006-1-2",
	"2006 01 02",
	"2006 1 2",
	"20060102",
	"02.01.2006",
	"2.1.2006",
	"01/02/2006",
	"1/2/2006",
	// etc as needed
}

// DateTimeFormats holds the datetime formats we accept in forms (note: all
// require times, and only to minute precision; this might change in a
// general library).
var DateTimeFormats = []string{
	"2006. 01. 02. 15:04",
	"2006. 01. 02 15:04",
	"2006. 1. 2. 15:04",
	"2006. 1. 2 15:04",
	"2006.01.02. 15:04",
	"2006.01.02 15:04",
	"2006.1.2. 15:04",
	"2006.1.2 15:04",
	"2006-01-02 15:04",
	"2006-1-2 15:04",
	"2006 01 02 15:04",
	"2006 1 2 15:04",
	"20060102150405",
	"02.01.2006 15:04",
	"2.1.2006 15:04",
	"01/02/2006 15:04",
	"1/2/2006 15:04",
	// etc as needed
}

// FormValuer is implemented by http.Request, and also in the unit tests for
// this package.  It may also be useful for overriding the standard,
// permissive form parsing behavior in net/http.
type FormValuer interface {
	FormValue(string) string
}

// MultiError is a type of error that contains a slice of errors.  In the
// standard Error method they are joined with a newline, but if cast to
// type the errors may be examined (or formatted) individually.
type MultiError struct {
	Errors []error
}

// Error implements the error interface for MultiError.
func (e *MultiError) Error() string {
	str := make([]string, len(e.Errors))
	for idx, err := range e.Errors {
		str[idx] = err.Error()
	}

	return strings.Join(str, "\n")
}

// AddFormSpecType adds or replaces FormSpec type t with converter function
// cf and optional default validator vf.  The converter must return a type
// that survives JSON marshaling and unmarshaling or runtime errors will
// occur in DecodeForm; its bool return value is the success or failure of
// the conversion.  Note that in many cases this is unnecessary, as the
// struct's final type will unmarshal from a simple string.
func AddFormSpecType(t string, cf func(string) (interface{}, bool),
	vf func(*FormSpec, interface{}) error) {

	formSpecTypeMap[t] = &formSpecType{
		converter: cf,
		validator: vf,
		custom:    true,
	}
}

type formSpecType struct {
	converter func(string) (interface{}, bool)
	validator func(*FormSpec, interface{}) error
	custom    bool
}

var formSpecTypeMap = map[string]*formSpecType{
	"bool":     &formSpecType{converter: boolConverter},
	"date":     &formSpecType{converter: dateConverter},
	"dateflex": &formSpecType{converter: dateFlexConverter},
	"datetime": &formSpecType{converter: dateTimeConverter},
	"float":    &formSpecType{converter: floatConverter, validator: floatValidator},
	"int":      &formSpecType{converter: intConverter, validator: intValidator},
	"int64":    &formSpecType{converter: int64Converter, validator: int64Validator},
	"string":   &formSpecType{converter: stringConverter, validator: stringValidator},
}

// FormSpec defines a single specification item for validating a form
// value corresponding to Key.  It is used by DecodeForm.
//
// Valid Type values include:
//
//   "string"       // plain string
//   "int"          // int, max 32 bits large
//   "int64"        // int64
//   "float"        // float64
//   "bool"         // bool: input must be "true" or "false" if required
//   "date"         // date, without time part; see below.
//   "datetime"     // date, with time part; see below.
//   "dateflex"     // date, with or without time part; see below.
//
// This list can be extended using the AddFormSpecType function.
//
// The Limit describes a validation check, and may be left as an empty
// string.  Limits include:
//
//   "123"          // length (strings, int, int64)
//   "1-10"         // allowed range of value (numeric) or length (string)
//   "a,b,c"        // list of simple string values accepted
//   "1,3,5"        // list of simple numeric values accepted
//   "re:^\w\d+$"   // regular expression (strings only)
//
// Note that the Limit is only processed during the Init phase.  If Init is
// not called, the Validator should enforce any custom limits.
//
// The optional Name is used in formatting error messages that may be shown to
// the user, e.g. "<Name> is out of range."
//
// Dates are valid in any format listed under DateFormats; DateTimes use
// those in DateTimeFormat; DateFlex use both.
//
// The Validator function is called with the FormSpec itself and the
// type-converted value (cf. Convert). Standard Validator functions are set
// by Init if no Validator exists when it is called.
type FormSpec struct {
	Key       string
	Type      string
	Required  bool
	Limit     string
	Name      string
	Validator func(*FormSpec, interface{}) error

	// Helpers for standard validators:
	limitLength     int
	limitRangeInt   []int64
	limitRangeFloat []float64
	limitListString []string
	limitListInt    []int64
	limitRegexp     *regexp.Regexp
}

// Init validates the FormSpec and prepares it for use.  This should
// usually not be called directly; use OptionalFormSpec and
// RequiredFormSpec instead wherever possible. Since bad specs indicate
// programmer error, failures result in panic.
func (fs *FormSpec) Init() {

	t := formSpecTypeMap[fs.Type]
	if t == nil {
		panic("Unsupported FormSpec type: " + fs.Type)
	}
	// Parse the Limit only for standard types.
	if !t.custom {
		fs.initLimit()
	}
	if fs.Validator == nil {
		fs.Validator = t.validator
	}

}

// Convert converts raw to the type indicated in the FormSpec's Type property,
// returning an error if it can not be converted.  If there is no error then
// the returned value is safe to pass to a standard Validator function.
func (fs *FormSpec) Convert(raw string) (interface{}, error) {
	val, ok := formSpecTypeMap[fs.Type].converter(raw)
	if !ok {
		// TODO: consider bubbling up errors for things like int out of range.
		return nil, fmt.Errorf("%s could not be converted to %s.",
			fs.Name, fs.Type)
	}
	return val, nil

}

// Copy returns a copy of the FormSpec with a new Key and Name.  The Key
// must not be an empty or whitespace-only string (it is whitespace-trimmed);
// the Name defaults to the Key.  Init is not called on the new object, as
// the source state is preserved.
//
// Use this to more efficiently create many functionally identical spec
// items, e.g. required-string validators.
func (fs *FormSpec) Copy(key, name string) *FormSpec {
	key = strings.TrimSpace(key)
	if key == "" {
		panic("Empty FormSpec key.")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = key
	}
	return &FormSpec{
		Key:       key,
		Type:      fs.Type,
		Required:  fs.Required,
		Limit:     fs.Limit,
		Name:      name,
		Validator: fs.Validator,

		// And:
		limitLength:     fs.limitLength,
		limitRangeInt:   fs.limitRangeInt,
		limitRangeFloat: fs.limitRangeFloat,
		limitListString: fs.limitListString,
		limitListInt:    fs.limitListInt,
		limitRegexp:     fs.limitRegexp,
	}

}

var formSpecLimitMatchLength = regexp.MustCompile("^[1-9][0-9]*$")
var formSpecLimitMatchRangeInt = regexp.MustCompile("^([0-9]+)-([0-9]+)$")
var formSpecLimitMatchRangeFloat = regexp.MustCompile("^([0-9]*[.][0-9]+)-([0-9]*[.][0-9]+)$")

func (fs *FormSpec) initLimit() {

	val := fs.Limit
	if val == "" {
		return
	}

	// Regexp limit:
	if strings.HasPrefix(val, "re:") {
		re, err := regexp.Compile(strings.TrimPrefix(val, "re:"))
		if err != nil {
			panic("Error compiling limit regexp: " + err.Error())
		}
		fs.limitRegexp = re
		return
	}

	// Length limit:
	if formSpecLimitMatchLength.MatchString(val) {

		// Only useful for strings and int-ies.
		if fs.Type != "string" && fs.Type != "int" && fs.Type != "int64" {
			panic("Length limit does not apply to " + fs.Type)
		}
		i, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			panic("Error parsing int for length limit: " + err.Error())
		}
		fs.limitLength = int(i)
		return
	}

	// Range limit (numeric or for strings, length) as integers:
	if m := formSpecLimitMatchRangeInt.FindStringSubmatch((val)); len(m) == 3 {
		bits := 32
		if fs.Type == "int64" || fs.Type == "float" {
			bits = 64
		}
		lower, err := strconv.ParseInt(m[1], 10, bits)
		if err != nil {
			panic("Error parsing int for range limit: " + err.Error())
		}
		upper, err := strconv.ParseInt(m[2], 10, bits)
		if err != nil {
			panic("Error parsing int for range limit: " + err.Error())
		}
		if upper < lower {
			panic("Bad range limit: upper < lower.")
		}
		fs.limitRangeInt = []int64{lower, upper}
		return
	}

	// Range limit as floats (only for float values):
	if m := formSpecLimitMatchRangeFloat.FindStringSubmatch((val)); len(m) == 3 {
		bits := 64 // always, for now.
		if fs.Type != "float" {
			panic("Float limit requires float type, not " + fs.Type)
		}
		lower, err := strconv.ParseFloat(m[1], bits)
		if err != nil {
			panic("Error parsing float for range limit: " + err.Error())
		}
		upper, err := strconv.ParseFloat(m[2], bits)
		if err != nil {
			panic("Error parsing float for range limit: " + err.Error())
		}
		if upper < lower {
			panic("Bad range limit: upper < lower.")
		}
		fs.limitRangeFloat = []float64{lower, upper}
		return
	}

	// Set of strings limit:
	if vals := strings.Split(val, ","); len(vals) > 0 {
		switch fs.Type {
		case "string":
			fs.limitListString = vals
		case "int", "int64":
			ints := make([]int64, len(vals))
			for idx, s := range vals {
				i, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					panic("Bad integer in list: " + err.Error())
				}
				ints[idx] = i
			}
		default:
			panic("Value list not compatible with type " + fs.Type)

		}
		return
	}

	// etc. as new ones come online.

	panic("Unknown limit: " + val)
}

// NewFormSpec returns a pointer to an initialized FormSpec that is ready for
// use and whose Required property is set to r.  The final two arguments,
// limit and name, may be omitted. If the spec is not understood, the
// function panics.
func NewFormSpec(r bool, k, t string, limitAndName ...string) *FormSpec {

	k = strings.TrimSpace(k)
	if k == "" {
		panic("key may not be empty")
	}
	if len(limitAndName) > 2 {
		panic("too many args")
	}
	l := ""
	if len(limitAndName) > 0 {
		l = limitAndName[0]
	}
	n := k
	if len(limitAndName) == 2 {
		n = limitAndName[1]
	}

	f := &FormSpec{
		Key:      k,
		Type:     t,
		Required: r,
		Limit:    l,
		Name:     n,
	}
	f.Init()

	return f
}

// OptionalFormSpec returns a pointer to an initialized FormSpec that is ready
// for use, and whose Required property is false, with key k, type t and the
// provided limit and name. The latter two may be omitted.
func OptionalFormSpec(k, t string, limitAndName ...string) *FormSpec {
	return NewFormSpec(false, k, t, limitAndName...)
}

// RequiredFormSpec returns a pointer to a validated FormSpec that is ready
// for use, and whose Required property is true, with key k, type t and the
// provided limit and name. The latter two may be omitted. If the spec is
// not understood, the function panics.
func RequiredFormSpec(k, t string, limitAndName ...string) *FormSpec {
	return NewFormSpec(true, k, t, limitAndName...)
}

// DecodeForm populates the target structure from the values of a submitted
// form (or any other FormValuer). On failure, returns an error which may be
// cast as a MultiError for formatting.
//
//   Q: Why this and not one of the introspection-based libraries?
//   A: None of those examined yet would work without major changes:
//   * gorilla schema doesn't handle times
//   * ajg/form is close but doesn't do times at locations, nor any validation
//   * vala (with form) would almost work but is stubbornly non-idiomatic
//
// Values are whitespace-trimmed before any processing occurs, unless
// DecodeFormTrimSpace is set to false.
//
// Missing form fields are treated as the zero value unless they are required.
// Unhandled fields are ignored.  Bad spec entries result in a panic.
//
// Optional empty fields are converted to the zero value for the type.
//
// Yes, this is messy, but whatchagonnado?
func DecodeForm(f FormValuer, specs []*FormSpec, target interface{}) error {

	errors := []error{}
	values := map[string]interface{}{}

	for _, spec := range specs {
		input := f.FormValue(spec.Key)
		if DecodeFormTrimSpace {
			input = strings.TrimSpace(input)
		}
		if spec.Required && input == "" {
			errors = append(errors, fmt.Errorf("%s is required.", spec.Name))
			continue
		}
		// Convert and validate!
		val, err := spec.Convert(input)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if spec.Validator != nil {
			if err := spec.Validator(spec, val); err != nil {
				errors = append(errors, err)
				continue
			}
		}

		// All passed for this item.
		values[spec.Key] = val

	}

	if len(errors) > 0 {
		return &MultiError{errors}
	}

	// Hmm, there must be a nice generic way to do this round-trip...
	jsonB, err := json.Marshal(values)
	if err != nil {
		panic("Could not marshal values to JSON: " + err.Error())
	}
	if err := json.Unmarshal(jsonB, target); err != nil {
		panic("Could not unmarshal JSON string: " + err.Error())
	}

	return nil
}

func stringValidator(fs *FormSpec, v interface{}) error {

	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("%s (%T) is not a string.", fs.Name, v)
	}

	slen := GlyphLength(s)
	if fs.limitLength > 0 && slen != fs.limitLength {
		return fmt.Errorf("%s has the wrong length.", fs.Name)
	}
	if len(fs.limitRangeInt) == 2 {
		if int64(slen) < fs.limitRangeInt[0] {
			return fmt.Errorf("%s is too short.", fs.Name)
		}
		if int64(slen) > fs.limitRangeInt[1] {
			return fmt.Errorf("%s is too long.", fs.Name)
		}
	}
	if fs.limitRegexp != nil && !fs.limitRegexp.MatchString(s) {
		return fmt.Errorf("%s has the wrong format.", fs.Name)
	}
	if len(fs.limitListString) > 0 {
		have := false
		for _, item := range fs.limitListString {
			if s == item {
				have = true
				break
			}
		}
		if !have {
			return fmt.Errorf("%s has the wrong value.", fs.Name)
		}
	}

	return nil
}

func intValidator(fs *FormSpec, v interface{}) error {

	i, ok := v.(int)
	if !ok && fs.Type == "int" {
		return fmt.Errorf("%s (%T) is not an integer.", fs.Name, v)
	}
	// Everything else is the same for int and int64.
	return int64Validator(fs, int64(i))

	// can't len(int) so we cheat...
	if fs.limitLength > 0 && len(fmt.Sprintf("%d", i)) != fs.limitLength {
		return fmt.Errorf("%s has the wrong length.", fs.Name)
	}
	if len(fs.limitRangeInt) == 2 {
		if int64(i) < fs.limitRangeInt[0] {
			return fmt.Errorf("%s is too short.", fs.Name)
		}
		if int64(i) > fs.limitRangeInt[1] {
			return fmt.Errorf("%s is too long.", fs.Name)
		}
	}

	// TODO: consider applying regexep to stringified numbers.  Why not?
	//       OTOH, why? (Zip codes? Phone numbers? Order numbers?)
	if len(fs.limitListInt) > 0 {
		have := false
		for _, item := range fs.limitListInt {
			if int64(i) == item {
				have = true
				break
			}
		}
		if !have {
			return fmt.Errorf("%s has the wrong value.", fs.Name)
		}
	}

	return nil
}

func int64Validator(fs *FormSpec, v interface{}) error {

	i, ok := v.(int64)
	if !ok && fs.Type == "int" {
		return fmt.Errorf("%s (%T) is not a 64-bit integer.", fs.Name, v)
	}

	// can't len(int) so we cheat...
	if fs.limitLength > 0 && len(fmt.Sprintf("%d", i)) != fs.limitLength {
		return fmt.Errorf("%s has the wrong length.", fs.Name)
	}
	if len(fs.limitRangeInt) == 2 {
		if i < fs.limitRangeInt[0] {
			return fmt.Errorf("%s is too low.", fs.Name)
		}
		if i > fs.limitRangeInt[1] {
			return fmt.Errorf("%s is too high.", fs.Name)
		}
	}

	// TODO: consider applying regexep to stringified numbers.  Why not?
	//       OTOH, why? (Zip codes? Phone numbers? Order numbers?)
	if len(fs.limitListInt) > 0 {
		have := false
		for _, item := range fs.limitListInt {
			if i == item {
				have = true
				break
			}
		}
		if !have {
			return fmt.Errorf("%s has the wrong value.", fs.Name)
		}
	}

	return nil
}

func floatValidator(fs *FormSpec, v interface{}) error {

	f, ok := v.(float64)
	if !ok {
		return fmt.Errorf(
			"%s is not a 64-bit floating point number, but a %T.",
			fs.Name, v)
	}

	// "int" range limit works for floats too, except perhaps at the extremes
	// (wait for the bug on that one...)
	// TODO: rethink the whole range limit idea... maybe stricter typing?
	if len(fs.limitRangeFloat) == 2 {
		if f < fs.limitRangeFloat[0] {
			return fmt.Errorf("%s is too low.", fs.Name)
		}
		if f > fs.limitRangeFloat[1] {
			return fmt.Errorf("%s is too high.", fs.Name)
		}
	}

	return nil
}

func boolConverter(raw string) (interface{}, bool) {
	if raw == "true" {
		return true, true
	}
	if raw == "false" || raw == "" {
		return false, true
	}
	return nil, false
}

func stringConverter(raw string) (interface{}, bool) { return raw, true }

func intConverter(raw string) (interface{}, bool) {
	if raw == "" {
		return int(0), true
	}
	i64, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return nil, false
	}
	return int(i64), true
}

func int64Converter(raw string) (interface{}, bool) {
	if raw == "" {
		return int64(0), true
	}
	i64, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, false
	}
	return i64, true
}

func floatConverter(raw string) (interface{}, bool) {
	if raw == "" {
		return float64(0), true
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, false
	}
	return f, true
}

func dateConverter(raw string) (interface{}, bool) {
	if raw == "" {
		return time.Time{}, true
	}
	for _, layout := range DateFormats {
		d, err := time.ParseInLocation(layout, raw, FormValueTimeLocation)
		if err == nil {
			return d, true
		}
	}
	return nil, false
}

func dateFlexConverter(raw string) (interface{}, bool) {
	if raw == "" {
		return time.Time{}, true
	}
	for _, layout := range DateFormats {
		d, err := time.ParseInLocation(layout, raw, FormValueTimeLocation)
		if err == nil {
			return d, true
		}
	}
	for _, layout := range DateTimeFormats {
		d, err := time.ParseInLocation(layout, raw, FormValueTimeLocation)
		if err == nil {
			return d, true
		}
	}
	return nil, false
}

func dateTimeConverter(raw string) (interface{}, bool) {
	if raw == "" {
		return time.Time{}, true
	}
	for _, layout := range DateTimeFormats {
		d, err := time.ParseInLocation(layout, raw, FormValueTimeLocation)
		if err == nil {
			return d, true
		}
	}
	return nil, false
}
