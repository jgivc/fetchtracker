package mdadapter

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type FileNodeRenderer struct {
}

func NewFileDirectiveRenderer() renderer.NodeRenderer {
	return &FileNodeRenderer{}
}

func (r *FileNodeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindFileNode, r.renderFileDirective)
}

func (r *FileNodeRenderer) renderFileDirective(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	fileNode, ok := n.(*FileNode)
	if !ok {
		return ast.WalkStop, fmt.Errorf("unexpected node %T, expected *Filedirective", fileNode)
	}

	if fileNode.Error != nil {
		return ast.WalkStop, fmt.Errorf("cannot render file template: %w", fileNode.Error)
	}

	w.Write(fileNode.HTML)

	return ast.WalkContinue, nil
}

type FilesRenderer struct {
}

func NewFilesRenderer() *FilesRenderer {
	return &FilesRenderer{}
}

func (r *FilesRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindFilesNode, r.renderFiles)
}

func (r *FilesRenderer) renderFiles(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		return ast.WalkContinue, nil
	}

	filesNode, ok := n.(*FilesNode)
	if !ok {
		return ast.WalkStop, fmt.Errorf("unexpected node %T, expected *Filedirective", filesNode)
	}

	if filesNode.Error != nil {
		return ast.WalkStop, fmt.Errorf("cannot render files template: %w", filesNode.Error)
	}

	w.Write(filesNode.HTML)

	return ast.WalkContinue, nil
}
