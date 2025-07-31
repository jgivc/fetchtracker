package mdadapter

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// type FileResolver interface {
// 	GetFile(fileName string) (*entity.File, error)
// 	GetFiles() []*entity.File
// }

type FilesExtension struct {
	// r         FileResolver
	// fileTmpl  *template.Template
	// filesTmpl *template.Template
}

func NewFilesExtension() goldmark.Extender {
	return &FilesExtension{}
}

func (e *FilesExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(NewFileDirectiveParser(), 199),
		),
		parser.WithBlockParsers(
			util.Prioritized(&FilesBlockParser{}, 199),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewFileDirectiveRenderer(), 199),
			util.Prioritized(NewFilesRenderer(), 199),
		),
	)
}
