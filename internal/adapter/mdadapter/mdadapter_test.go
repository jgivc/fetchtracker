package mdadapter

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

var source = `# AAA
{{ file: aaa.txt }} dffdsfdsf sdfds f

`

func TestMDAdapter(t *testing.T) {
	md := goldmark.New(
		goldmark.WithExtensions(
			NewFilesExtension(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(source), &buf); err != nil {
		panic(err)
	}

	fmt.Println(buf.String())
}

func TestTemplate(t *testing.T) {
	// tmpl, err := template.ParseFiles("templates/template.html")
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// ftmpl := tmpl.Lookup("FILE")
	// if ftmpl == nil {
	// 	fmt.Println("No template FILE found")
	// }

	// fstmpl := tmpl.Lookup("FILES")
	// if fstmpl == nil {
	// 	fmt.Println("No template FILES found")
	// }

	// funeFunc := func(name string) string {
	// 	var b bytes.Buffer

	// 	type C struct {
	// 		Name string
	// 	}

	// 	ftmpl.Execute(&b, &C{Name: name})

	// 	return b.String()
	// }
	//
	data, err := os.ReadFile("templates/template.html")
	if err != nil {
		t.Fatal(err)
	}

	tt := template.New("")
	tt.Funcs(template.FuncMap{
		"file": func(name string, args ...string) template.HTML {
			tmpl := tt.Lookup("FILE")
			if tmpl == nil {
				return "No teplate FILE"
			}

			type C struct {
				Name  string
				Descr string
			}

			descr := name
			if len(args) > 0 {
				descr = args[0]
			}

			b := bytes.Buffer{}
			if err := tmpl.Execute(&b, &C{Name: name, Descr: descr}); err != nil {
				return "cannot execute template"
			}

			return template.HTML(b.String())
		},
	})
	_, err = tt.Parse(string(data))
	if err != nil {
		t.Fatal(err)
	}

	if err := tt.Execute(os.Stdout, nil); err != nil {
		t.Fatal(err)
	}
	// fmt.Println(tt1.Templates())

	// b := &bytes.Buffer{}

	// tt, err := template.New("").Funcs(template.FuncMap{
	// 	"file": func(name string) string {
	// 		return "AAAA"
	// 	},
	// }).ParseFiles("templates/template1.html")
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// fmt.Println("a", tt.Name())
	// for _, t := range tt.Templates() {
	// 	fmt.Println("a1", t.Name())
	// }

	// tt.ExecuteTemplate(b, "template1.html", "qqq")
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// fmt.Print(b.String())
}
