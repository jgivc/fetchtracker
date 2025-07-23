package mdadapter

import (
	"github.com/yuin/goldmark/ast"
)

var KindFileDirective = ast.NewNodeKind("FileDirective")

// FileDirective представляет директиву для файлов
type FileDirective struct {
	ast.BaseInline
	Filename string
	// Description string
	// Template    string
	// Pattern     string
	// LinkType    string // "file", "files", "files:remaining", "files:all"
}

// Kind реализует интерфейс ast.Node
func (n *FileDirective) Kind() ast.NodeKind {
	return KindFileDirective
}

// Dump реализует интерфейс ast.Node
func (n *FileDirective) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		// "LinkType":    n.LinkType,
		"Filename": n.Filename,
		// "Description": n.Description,
		// "Template":    n.Template,
		// "Pattern":     n.Pattern,
	}, nil)
}
