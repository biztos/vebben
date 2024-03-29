# vebben

A jarful of golang web glue.

[![GoDoc][b1]][doc] [![Coverage Status][b3]][cov]


[b1]: https://pkg.go.dev/badge/github.com/biztos/vebben
[doc]: https://pkg.go.dev/github.com/biztos/vebben
[b2]: https://travis-ci.org/biztos/vebben.svg?branch=master
[ci]: https://travis-ci.org/biztos/vebben
[b3]: https://coveralls.io/repos/github/biztos/vebben/badge.svg
[cov]: https://coveralls.io/github/biztos/vebben


Thank you for stopping by.  I hope you find something useful here, even if
it's only the counter-example of my preoccupations. :-)

## WARNING! ALPHA SOFTWARE!


This package has not been tested much, and the author hasn't actually used it
in production in years. Web templating has come a long way since `vebben` was
written, and you should probably use something else!

Like all software, it probably contains bugs, and like all new software it 
probably contains a lot of them. 🪲🪲🪲

---

I don't know how much this will encompass in the end, but my goal is to
include in the `vebben` package all web-dev helpers I write in the course of
developing a couple of different systems: a bloggy thing that's still on the
eternal drawing board but will eventually be open-source; and a
salon-management tool in Google App Engine that will remain private.

Right now, it's just a form processor I built because I couldn't make the
ones I found do what I needed; and a collection of utility functions for
standard [templates][tmpl] that leans *heavily* on the work of
[Kyoung-chan Lee][leekchan].

And it may change at any time.  If you use it (I'd be flattered) be sure to
vendor it in, at least until I reach a `1.0` sort of release.

[tmpl]: https://pkg.go.dev/html/template
[leekchan]: http://leekchan.com

## FormSpec Example


```go
// vebben simple demo
package main

import (
    "fmt"
    "net/http"

    "github.com/biztos/vebben"
)

type Flubber struct {
    Variant  string  `json:"variant"`
    Size     int     `json:"size"`
    Strength float64 `json:"strength"`
}

var FlubberSpecs = []*vebben.FormSpec{
    vebben.RequiredFormSpec("variant", "string", "4", "The 4-letter variant"),
    vebben.RequiredFormSpec("size", "int", "1-4", "The size (1-4)"),
    vebben.OptionalFormSpec("strength", "float", "", "Flubber strength"),
}

func main() {

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

        if r.Method == "POST" {
            f := &Flubber{}
            err := vebben.DecodeForm(r, FlubberSpecs, f)
            if err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }
            fmt.Fprintf(w, "Flubber %s: size %d, strength %0.2f\n",
                f.Variant, f.Size, f.Strength)
        } else {
            fmt.Fprintf(w, "POST to describe some Flubber.")
        }

    })
    fmt.Println(http.ListenAndServe(":8080", nil))
}
```

And then, with the above code running, from the comfort of your favorite
UNIX shell try this:

```bash
curl -i http://localhost:8080/ -d variant=FLUB -d size=4 -d strength=99.4522
```

...and variations thereof.

## FuncMap Example

```go

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

```

## Acknowledgements

The following packages have helped tremendously, and have even wittier names
than this package.  All kudos to their authors!

* Testify by Mat Ryer and Tyler Bunnel
    * https://github.com/stretchr/testify
* GTF by Kyoung-chan Lee
    * https://github.com/leekchan/gtf

## About the Name

While it's arguably true that `vebben` would be a great name for a Norwegian
Black Metal band, the only such band I really listen to is [Mayhem][m], and I
only started listening to them because of their Hungarian connection.

Which brings us back to *vebben.*  Say it loud and say it proud:

> *Mindenféle szar program van fönt a vebben...*


[m]: https://www.thetruemayhem.com
