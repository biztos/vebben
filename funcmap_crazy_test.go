// vebben_crazy_test.go - crazy complex vebben tests.
// ---------------------

package vebben_test

import (

	// Standard:
	"bytes"
	"html/template"
	"testing"

	// Third-party:
	"github.com/stretchr/testify/assert"

	// Under test:
	"github.com/biztos/vebben"
)

type someDot struct {
	Title    string
	Urls     []string
	Register int
}

func (d *someDot) SetRegister(v int) int {
	old := d.Register
	d.Register = v
	return old
}

func Test_CrazyNestedList(t *testing.T) {

	assert := assert.New(t)

	// Some data, perhaps useful:
	dot := &someDot{
		Title: "Hello World",
		Urls: []string{
			"/foo",
			"/foo/bar",
			"/foo/bar/baz",
			"/foo/bar/boo",
			"/foo/bar/baz/bat",
			"/zoo",
		},
	}

	// Let's nest flat depth-sorted lists!
	var tsrc = `<h1>{{ .Title }}</h1>
<h2>URLS:</h2>
{{ if .Urls -}}
<ul>
    {{-  $d := . -}}
    {{ $toss := $d.SetRegister 1 -}}
    {{ range .Urls -}}
        {{ $cur := pathdepth . -}}
        {{ $old := $d.SetRegister $cur -}}
        {{ if gt $cur $old -}}
            {{ $r := intrange $old $cur | truncints -}}
            {{ range $r -}}
                {{ "\n" }}{{ indent . }}<li><ul>
            {{- end -}}
        {{- else if lt $cur $old -}}
            {{ $r := intrange $cur $old | truncints | reverseints -}}
            {{ range $r -}}
                {{ "\n" }}{{ indent .  }}</ul></li>
            {{- end -}}
        {{- end -}}
        {{ "\n" }}{{ indent $cur }}<li>{{ . }}</li>
    {{- end -}}
    {{ $r := intrange 1 $d.Register | truncints | reverseints -}}
    {{ range $r -}}
        {{ "\n" }}{{ indent . }}</ul></li>
    {{- end }}
</ul>
{{- end }}`
	tmpl, err := template.New("test").Funcs(vebben.NewFuncMap()).Parse(tsrc)
	if err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, dot); err != nil {
		t.Fatal(err)
	}

	exp := `<h1>Hello World</h1>
<h2>URLS:</h2>
<ul>
    <li>/foo</li>
    <li><ul>
        <li>/foo/bar</li>
        <li><ul>
            <li>/foo/bar/baz</li>
            <li>/foo/bar/boo</li>
            <li><ul>
                <li>/foo/bar/baz/bat</li>
            </ul></li>
        </ul></li>
    </ul></li>
    <li>/zoo</li>
</ul>`

	assert.Equal(exp, buf.String(), "output as expected")
}
