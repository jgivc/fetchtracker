package fsadapter

import (
	"fmt"
	"html/template"

	"github.com/jgivc/fetchtracker/internal/entity"
)

const (
	buildIndexThreshold = 5
)

type fileResolver struct {
	files []*entity.File
	index map[string]int
}

func newFileResolver(files []*entity.File) *fileResolver {
	return &fileResolver{files: files}
}

func (r *fileResolver) GetFile(fileName string) (*entity.File, error) {
	if r.index == nil && len(r.files) > buildIndexThreshold {
		r.buildIndex()

		if idx, ok := r.index[fileName]; ok {
			return r.files[idx], nil
		}

		return nil, fmt.Errorf("cannot find file: %s", fileName)
	}

	for i := range r.files {
		file := r.files[i]
		if file.Name == fileName {
			return file, nil
		}
	}

	return nil, fmt.Errorf("cannot find file: %s", fileName)
}

func (r *fileResolver) buildIndex() {
	index := make(map[string]int)
	for i, file := range r.files {
		index[file.Name] = i
	}

	r.index = index
}

func (r *fileResolver) GetFiles() []*entity.File {
	return r.files
}

type templateResolver struct {
	fileTemplate  *template.Template
	filesTemplate *template.Template
}

func (r *templateResolver) GetFileTemplate() *template.Template {
	return r.fileTemplate
}

func (r *templateResolver) GetFilesTemplate() *template.Template {
	return r.filesTemplate
}

func newTemplateResolver(tmpl *template.Template) (*templateResolver, error) {
	r := &templateResolver{
		fileTemplate:  tmpl.Lookup(templateNameFile),
		filesTemplate: tmpl.Lookup(templateNameFiles),
	}

	if r.fileTemplate == nil {
		return nil, fmt.Errorf("cannot find template %s", templateNameFile)
	}

	if r.filesTemplate == nil {
		return nil, fmt.Errorf("cannot find template %s", templateNameFiles)
	}

	return r, nil
}
