package mdadapter

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/jgivc/fetchtracker/internal/entity"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

var source = `# AAA
[[aaa.txt]] dffdsfdsf sdfds f
Test text text test 321
[[bbb.txt|File bbb]] d
Test 123 eee ss kjhkjh

А тут еще

[[FILES]]

И вот еще текст
`

var tsrc = `{{ define "FILE" }}
<a href="#">{{ .Name }}</a>
{{ end }}

{{ define "FILES" }}
<ul>
{{ range . }}<li><a href="#">{{ .Name }}</a></li>
{{ end }}</ul>
{{ end }}`

//
// var tsrc = `{{ define "FILE" }}
// <a href="#">{{ .Name }}</a>
// {{ end}}
// {{ define "FILES" }}
// <ul>
// {{ range . }}
// <li><a href="#">{{ .Name }}</a></li>
// {{ end }}
// </ul>
// {{ end }}`

type MockFileResolver struct {
	mock.Mock
}

func (m *MockFileResolver) GetFile(fileName string) (*entity.File, error) {
	args := m.Called(fileName)

	var f *entity.File
	if args[0] != nil {
		if ff, ok := args.Get(0).(*entity.File); ok {
			f = ff
		}
	}

	return f, args.Error(1)
}

func (m *MockFileResolver) GetFiles() []*entity.File {
	args := m.Called()

	var f []*entity.File
	if args[0] != nil {
		if ff, ok := args.Get(0).([]*entity.File); ok {
			f = ff
		}
	}

	return f
}

func TestMDAdapter(t *testing.T) {
	m := new(MockFileResolver)
	m.On("GetFile", "aaa.txt").Return(&entity.File{
		Name: "aaa.txt",
	}, nil)
	m.On("GetFile", "bbb.txt").Return(&entity.File{
		Name: "bbb.txt",
	}, nil)

	m.On("GetFiles").Return([]*entity.File{
		{Name: "aaa.txt"},
		{Name: "bbb.txt"},
		{Name: "ccc.txt"},
	})

	tmpl, err := template.New("").Parse(tsrc)
	require.NoError(t, err)

	fTmpl := tmpl.Lookup("FILE")
	require.NotNil(t, fTmpl)

	fsTmpl := tmpl.Lookup("FILES")
	require.NotNil(t, fsTmpl)

	md := goldmark.New(
		goldmark.WithExtensions(
			NewFilesExtension(m, fTmpl, fsTmpl),
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

// func TestTemplate(t *testing.T) {

// 	data, err := os.ReadFile("templates/template.html")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	tt := template.New("")
// 	tt.Funcs(template.FuncMap{
// 		"file": func(name string, args ...string) template.HTML {
// 			tmpl := tt.Lookup("FILE")
// 			if tmpl == nil {
// 				return "No teplate FILE"
// 			}

// 			type C struct {
// 				Name  string
// 				Descr string
// 			}

// 			descr := name
// 			if len(args) > 0 {
// 				descr = args[0]
// 			}

// 			b := bytes.Buffer{}
// 			if err := tmpl.Execute(&b, &C{Name: name, Descr: descr}); err != nil {
// 				return "cannot execute template"
// 			}

// 			return template.HTML(b.String())
// 		},
// 	})
// 	_, err = tt.Parse(string(data))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if err := tt.Execute(os.Stdout, nil); err != nil {
// 		t.Fatal(err)
// 	}
// }
