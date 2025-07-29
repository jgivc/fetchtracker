package mdadapter

import (
	"github.com/yuin/goldmark/ast"
)

var KindFileDirective = ast.NewNodeKind("FileDirective")

type FileDirective struct {
	ast.BaseInline
	Filename    string
	Description string
	AllFiles    bool
}

func (n *FileDirective) Kind() ast.NodeKind {
	return KindFileDirective
}

func (n *FileDirective) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Filename":    n.Filename,
		"Description": n.Description,
	}, nil)
}
