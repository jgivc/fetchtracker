package mdadapter

import (
	"text/template"

	"github.com/jgivc/fetchtracker/internal/entity"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type FileResolver interface {
	GetFile(fileName string) (*entity.File, error)
	GetFiles() []*entity.File
}

type FilesExtension struct {
	r    FileResolver
	tmpl *template.Template
}

func NewFilesExtension(r FileResolver, tmpl *template.Template) goldmark.Extender {
	return &FilesExtension{r: r, tmpl: tmpl}
}

func (e *FilesExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(NewFileDirectiveParser(), 199),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewFileDirectiveRenderer(e.r, e.tmpl), 199),
		),
	)
}
