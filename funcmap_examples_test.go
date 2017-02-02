// vebben_examples_test.go - examples for template helper funcs
// ------------------------

package vebben_test

import (
	"html/template"
	"os"

	"github.com/biztos/vebben"
)

func Example() {

	dot := map[string]interface{}{
		"Title": "Hello World!",
		"Id":    "examples",
		"Prices": []int{
			15,
			11800002,
			3582,
			999231,
			10012,
		},
	}
	var tsrc = `<h1 id="{{ .Id | capfirst }}">{{ .Title }}</h1>
<h2>Top Three Prices:</h2>
<ol>
{{- $p := sortintsdesc .Prices | truncintsto 3 }}{{ range $p }}
<li>${{ intcomma . }}</li>
{{- end }}
</ol>
`

	tmpl, err := template.New("test").Funcs(vebben.NewFuncMap()).Parse(tsrc)
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(os.Stdout, dot); err != nil {
		panic(err)
	}

	// Output:
	// <h1 id="Examples">Hello World!</h1>
	// <h2>Top Three Prices:</h2>
	// <ol>
	// <li>$11,800,002</li>
	// <li>$999,231</li>
	// <li>$10,012</li>
	// </ol>
}
