package mdadapter

import (
	"github.com/yuin/goldmark/ast"
)

var KindFileNode = ast.NewNodeKind("FileNode")
var KindFilesNode = ast.NewNodeKind("FilesNode")

type FileNode struct {
	ast.BaseInline
	HTML     []byte
	Error    error
	Filename string
}

func (n *FileNode) Kind() ast.NodeKind {
	return KindFileNode
}

func (n *FileNode) Dump(source []byte, level int) {
	attrs := map[string]string{
		"Filename": n.Filename,
		"HTML":     string(n.HTML),
	}
	if n.Error != nil {
		attrs["Error"] = n.Error.Error()
	}
	ast.DumpHelper(n, source, level, attrs, nil)
}

type FilesNode struct {
	ast.BaseBlock
	HTML  []byte
	Error error
}

func (n *FilesNode) Kind() ast.NodeKind {
	return KindFilesNode
}

func (n *FilesNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"HTML":  string(n.HTML),
		"Error": n.Error.Error(),
	}, nil)
}
