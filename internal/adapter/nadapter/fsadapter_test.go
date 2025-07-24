package nadapter

import (
	"fmt"
	"testing"
	"text/template"
)

func TestTmpl(t *testing.T) {

	a := "aaa"
	b := "bbb"

	tt, err := template.New(a).Parse(a)
	if err != nil {
		t.Fatal(err)
	}

	printTemplate(tt)

	tt, err = tt.New(b).Parse(b)
	if err != nil {
		t.Fatal(err)
	}

	printTemplate(tt)
}

func printTemplate(tt *template.Template) {
	for i, tmpl := range tt.Templates() {
		fmt.Println(i, tmpl.Name())
	}
}
