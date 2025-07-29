package mdadapter

import (
	"github.com/yuin/goldmark/ast"
)

var KindFileNode = ast.NewNodeKind("FileNode")
var KindFilesNode = ast.NewNodeKind("FilesNode")

type FileNode struct {
	ast.BaseInline
	Filename    string
	Description string
	AllFiles    bool
}

func (n *FileNode) Kind() ast.NodeKind {
	return KindFileNode
}

func (n *FileNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Filename":    n.Filename,
		"Description": n.Description,
	}, nil)
}

type FilesNode struct {
	ast.BaseBlock
}

func (n *FilesNode) Kind() ast.NodeKind {
	return KindFilesNode
}

func (n *FilesNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}
