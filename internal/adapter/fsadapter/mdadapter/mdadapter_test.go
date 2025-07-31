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
	"github.com/yuin/goldmark/parser"
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

type MockTemplateResolver struct {
	mock.Mock
}

func (m *MockTemplateResolver) GetFileTemplate() *template.Template {
	args := m.Called()

	var tmpl *template.Template
	if args[0] != nil {
		if t, ok := args.Get(0).(*template.Template); ok {
			tmpl = t
		}
	}

	return tmpl
}

func (m *MockTemplateResolver) GetFilesTemplate() *template.Template {
	args := m.Called()

	var tmpl *template.Template
	if args[0] != nil {
		if t, ok := args.Get(0).(*template.Template); ok {
			tmpl = t
		}
	}

	return tmpl
}

func TestMDAdapter(t *testing.T) {
	mfr := new(MockFileResolver)
	mfr.On("GetFile", "aaa.txt").Return(&entity.File{
		Name: "aaa.txt",
	}, nil)
	mfr.On("GetFile", "bbb.txt").Return(&entity.File{
		Name: "bbb.txt",
	}, nil)

	mfr.On("GetFiles").Return([]*entity.File{
		{Name: "aaa.txt"},
		{Name: "bbb.txt"},
		{Name: "ccc.txt"},
	})

	tmpl, err := template.New("").Parse(tsrc)
	require.NoError(t, err)

	fTmpl := tmpl.Lookup("FILE")
	require.NotNil(t, fTmpl)

	mtr := new(MockTemplateResolver)
	mtr.On("GetFileTemplate").Return(fTmpl)

	fsTmpl := tmpl.Lookup("FILES")
	require.NotNil(t, fsTmpl)

	mtr.On("GetFilesTemplate").Return(fsTmpl)

	md := goldmark.New(
		goldmark.WithExtensions(
			NewFilesExtension(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	pc := parser.NewContext()
	pc.Set(FileResolverKey, mfr)
	pc.Set(TemplateResolverKey, mtr)

	var buf bytes.Buffer
	if err := md.Convert([]byte(source), &buf, parser.WithContext(pc)); err != nil {
		panic(err)
	}

	fmt.Println(buf.String())
}
