package tpladapter

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

	_ "embed"

	"github.com/jgivc/fetchtracker/internal/entity"
)

const (
	templateNameFile  = "FILE"
	templateNameFiles = "FILES"

	funcNameFile  = "file"
	funcNameFiles = "files"
)

//go:embed template.html
var defaultTemplate string

type tplAdapter struct {
	tpl      *template.Template
	download *entity.Download
}

func NewTplAdapter(templateFileName string) (*tplAdapter, error) {
	a := &tplAdapter{}
	tpl := template.New("").Funcs(template.FuncMap{
		funcNameFile:  a.renderFile,
		funcNameFiles: a.renderFiles,
	})

	src := defaultTemplate
	if templateFileName != "" {
		data, err := os.ReadFile(templateFileName)
		if err != nil {
			return nil, fmt.Errorf("cannot read template: %w", err)
		}

		src = string(data)
	}

	if _, err := tpl.Parse(src); err != nil {
		return nil, fmt.Errorf("cannot parse template: %w", err)
	}

	a.tpl = tpl

	return a, nil
}

func (a *tplAdapter) Parse(download *entity.Download) (string, error) {
	a.download = download

	buf := bytes.Buffer{}
	if err := a.tpl.Execute(&buf, download); err != nil {
		return "", fmt.Errorf("cannot execute template: %w", err)
	}

	return buf.String(), nil
}

func (a *tplAdapter) lookupFile(name string) *entity.File {
	if a.download == nil || len(a.download.Files) < 1 {
		return nil
	}

	for _, file := range a.download.Files {
		if file.Name == name {
			return file
		}
	}

	return nil
}

func (a *tplAdapter) renderFile(fileName string, args ...string) (template.HTML, error) {
	tpl := a.tpl.Lookup(templateNameFile)
	if tpl == nil {
		return "", fmt.Errorf("template %s must be defined", templateNameFile)
	}

	file := a.lookupFile(fileName)
	if file == nil {
		return "", fmt.Errorf("cannot find file: %s", fileName)
	}

	if len(args) > 0 {
		file.Description = args[0]
	} else {
		file.Description = fileName
	}

	buf := bytes.Buffer{}
	if err := tpl.Execute(&buf, file); err != nil {
		return "", fmt.Errorf("cannot execute template %s: %w", templateNameFile, err)
	}

	return template.HTML(buf.String()), nil
}

func (a *tplAdapter) renderFiles() (template.HTML, error) {
	tpl := a.tpl.Lookup(templateNameFiles)
	if tpl == nil {
		return "", fmt.Errorf("template %s must be defined", templateNameFiles)
	}

	buf := bytes.Buffer{}
	if err := tpl.Execute(&buf, a.download.Files); err != nil {
		return "", fmt.Errorf("cannot execute template %s: %w", templateNameFile, err)
	}

	return template.HTML(buf.String()), nil
}
