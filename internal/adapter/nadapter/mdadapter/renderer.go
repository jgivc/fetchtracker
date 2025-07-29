package mdadapter

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

const (
	tmplNameFile  = "FILE"
	tmplNameFiles = "FILES"
)

type FileDirectiveRenderer struct {
	r    FileResolver
	tmpl *template.Template
}

func NewFileDirectiveRenderer(r FileResolver, tmpl *template.Template) renderer.NodeRenderer {
	return &FileDirectiveRenderer{r: r, tmpl: tmpl}
}

func (r *FileDirectiveRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindFileDirective, r.renderFileDirective)
}

func (r *FileDirectiveRenderer) renderFileDirective(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	directive, ok := n.(*FileDirective)
	if !ok {
		return ast.WalkStop, fmt.Errorf("unexpected node %T, expected *Filedirective", directive)
	}

	if directive.AllFiles {
		// tmpl := r.tmpl.Lookup(tmplNameFiles)
		// if tmpl == nil {
		// 	return ast.WalkStop, fmt.Errorf("template with name %s must be defined", tmplNameFiles)
		// }

		// if err := tmpl.Execute(w, r.r.GetFiles()); err != nil {
		// 	return ast.WalkStop, fmt.Errorf("cannot execute template: %w", err)
		// }

		data, err := r.renderTemplate(tmplNameFiles, r.r.GetFiles())
		if err != nil {
			return ast.WalkStop, err
		}

		w.Write(data)

		return ast.WalkContinue, nil
	}

	file, err := r.r.GetFile(directive.Filename)
	if err != nil {
		return ast.WalkStop, fmt.Errorf("cannot get file %s: %w", directive.Filename, err)
	}

	if directive.Description != "" {
		file.Description = directive.Description
	}

	// tmpl := r.tmpl.Lookup(tmplNameFile)
	// if tmpl == nil {
	// 	return ast.WalkStop, fmt.Errorf("template with name %s must be defined", tmplNameFile)
	// }

	// if err := tmpl.Execute(w, file); err != nil {
	// 	return ast.WalkStop, fmt.Errorf("cannot execute template: %w", err)
	// }
	//
	data, err := r.renderTemplate(tmplNameFile, file)
	if err != nil {
		return ast.WalkStop, err
	}

	w.Write(data)

	return ast.WalkContinue, nil
}

func (r *FileDirectiveRenderer) renderTemplate(tmplName string, data any) ([]byte, error) {
	tmpl := r.tmpl.Lookup(tmplNameFile)
	if tmpl == nil {
		return nil, fmt.Errorf("template with name %s must be defined", tmplName)
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return nil, fmt.Errorf("cannot execute template: %w", err)
	}

	return buf.Bytes(), nil
}
