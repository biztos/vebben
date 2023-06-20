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
	fmt.Println("listening on 8080 for POST such as:")
	fmt.Println("curl -i http://localhost:8080/ -d size=2 -d variant=foop -d strength=3.4")
	fmt.Println(http.ListenAndServe(":8080", nil))
}
