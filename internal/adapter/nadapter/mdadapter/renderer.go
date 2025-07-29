package mdadapter

import (
	"fmt"
	"text/template"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type FileNodeRenderer struct {
	r    FileResolver
	tmpl *template.Template
}

func NewFileDirectiveRenderer(r FileResolver, tmpl *template.Template) renderer.NodeRenderer {
	return &FileNodeRenderer{r: r, tmpl: tmpl}
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

	file, err := r.r.GetFile(fileNode.Filename)
	if err != nil {
		return ast.WalkStop, fmt.Errorf("cannot get file %s: %w", fileNode.Filename, err)
	}

	if fileNode.Description != "" {
		file.Description = fileNode.Description
	}

	if err := r.tmpl.Execute(w, file); err != nil {
		return ast.WalkStop, fmt.Errorf("cannot execute template: %w", err)
	}

	return ast.WalkContinue, nil
}

// FilesRenderer рендерер для FilesNode
type FilesRenderer struct {
	r    FileResolver
	tmpl *template.Template
}

// RegisterFuncs регистрирует функции рендеринга
func (r *FilesRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindFilesNode, r.renderFiles)
}

// renderFiles рендерит список файлов
func (r *FilesRenderer) renderFiles(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		if err := r.tmpl.Execute(w, r.r.GetFiles()); err != nil {
			return ast.WalkStop, fmt.Errorf("cannot execute template: %w", err)
		}
		// fmt.Printf("IIIIIIIIIIIIII: %T\n", node)
		// // Здесь вы можете получить список файлов из контекста или другого источника
		// files := []struct {
		// 	ID   string
		// 	Name string
		// }{
		// 	{"1", "document.pdf"},
		// 	{"2", "config.json"},
		// 	{"3", "readme.txt"},
		// }

		// _, _ = w.WriteString("<ul class=\"files-list\">\n")

		// for _, file := range files {
		// 	_, _ = w.WriteString("  <li>\n")
		// 	_, _ = w.WriteString("    <form method=\"POST\" action=\"/file/" + file.ID + "/\" style=\"display: inline;\">\n")
		// 	_, _ = w.WriteString("      <button type=\"submit\" class=\"btn btn-outline-primary btn-sm\">\n")
		// 	_, _ = w.WriteString("        <i class=\"bi bi-download me-1\"></i>\n")
		// 	_, _ = w.WriteString("        " + file.Name + "\n")
		// 	_, _ = w.WriteString("      </button>\n")
		// 	_, _ = w.WriteString("    </form>\n")
		// 	_, _ = w.WriteString("  </li>\n")
		// }

		// _, _ = w.WriteString("</ul>\n")
	}

	return ast.WalkContinue, nil
}
