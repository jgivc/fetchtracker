package mdadapter

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// FileDirectiveRenderer рендерит директивы файлов
type FileDirectiveRenderer struct {
	// FileService FileService
}

func NewFileDirectiveRenderer() renderer.NodeRenderer {
	return &FileDirectiveRenderer{}
}

func (r *FileDirectiveRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindFileDirective, r.renderFileDirective)
}

func (r *FileDirectiveRenderer) renderFileDirective(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	directive := n.(*FileDirective)

	w.WriteString(fmt.Sprintf(`<a class="aabb" href="#">%s</a>`, directive.Filename))

	return ast.WalkContinue, nil
}

// type FileService interface {
// 	GetFileURL(filename string) string
// 	GetFileDownloads(filename string) int
// 	GetAvailableFiles() []string
// 	GetFilesByPattern(pattern string) []string
// 	MarkFileAsProcessed(filename string)
// 	GetRemainingFiles() []string
// }

// func NewFileDirectiveRenderer(fs FileService) renderer.NodeRenderer {
// 	return &FileDirectiveRenderer{FileService: fs}
// }

// func (r *FileDirectiveRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
// 	reg.Register(KindFileDirective, r.renderFileDirective)
// }

// func (r *FileDirectiveRenderer) renderFileDirective(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
// 	if !entering {
// 		return ast.WalkContinue, nil
// 	}

// 	directive := n.(*FileDirective)

// 	switch directive.LinkType {
// 	case "file":
// 		r.renderSingleFile(w, directive)
// 	case "files_pattern":
// 		r.renderFilesByPattern(w, directive)
// 	case "files_all":
// 		r.renderAllFiles(w)
// 	case "files_remaining":
// 		r.renderRemainingFiles(w)
// 	}

// 	return ast.WalkContinue, nil
// }

// func (r *FileDirectiveRenderer) renderSingleFile(w util.BufWriter, directive *FileDirective) {
// 	url := r.FileService.GetFileURL(directive.Filename)
// 	downloads := r.FileService.GetFileDownloads(directive.Filename)
// 	description := directive.Description
// 	if description == "" {
// 		description = directive.Filename
// 	}

// 	// Отмечаем файл как обработанный
// 	r.FileService.MarkFileAsProcessed(directive.Filename)

// 	template := directive.Template
// 	if template == "" {
// 		template = "default"
// 	}

// 	html := r.renderFileTemplate(template, map[string]interface{}{
// 		"URL":         url,
// 		"Filename":    directive.Filename,
// 		"Description": description,
// 		"Downloads":   downloads,
// 	})

// 	w.WriteString(html)
// }

// func (r *FileDirectiveRenderer) renderFilesByPattern(w util.BufWriter, directive *FileDirective) {
// 	files := r.FileService.GetFilesByPattern(directive.Pattern)

// 	w.WriteString(`<div class="files-group">`)
// 	for _, filename := range files {
// 		r.FileService.MarkFileAsProcessed(filename)
// 		url := r.FileService.GetFileURL(filename)
// 		downloads := r.FileService.GetFileDownloads(filename)

// 		html := r.renderFileTemplate("default", map[string]interface{}{
// 			"URL":         url,
// 			"Filename":    filename,
// 			"Description": filename,
// 			"Downloads":   downloads,
// 		})
// 		w.WriteString(html)
// 		w.WriteString(`<br>`)
// 	}
// 	w.WriteString(`</div>`)
// }

// func (r *FileDirectiveRenderer) renderAllFiles(w util.BufWriter) {
// 	files := r.FileService.GetAvailableFiles()

// 	w.WriteString(`<div class="all-files">`)
// 	for _, filename := range files {
// 		r.FileService.MarkFileAsProcessed(filename)
// 		url := r.FileService.GetFileURL(filename)
// 		downloads := r.FileService.GetFileDownloads(filename)

// 		html := r.renderFileTemplate("default", map[string]interface{}{
// 			"URL":         url,
// 			"Filename":    filename,
// 			"Description": filename,
// 			"Downloads":   downloads,
// 		})
// 		w.WriteString(html)
// 		w.WriteString(`<br>`)
// 	}
// 	w.WriteString(`</div>`)
// }

// func (r *FileDirectiveRenderer) renderRemainingFiles(w util.BufWriter) {
// 	files := r.FileService.GetRemainingFiles()

// 	if len(files) == 0 {
// 		return
// 	}

// 	w.WriteString(`<div class="remaining-files">`)
// 	w.WriteString(`<h3>Дополнительные файлы:</h3>`)
// 	for _, filename := range files {
// 		url := r.FileService.GetFileURL(filename)
// 		downloads := r.FileService.GetFileDownloads(filename)

// 		html := r.renderFileTemplate("default", map[string]interface{}{
// 			"URL":         url,
// 			"Filename":    filename,
// 			"Description": filename,
// 			"Downloads":   downloads,
// 		})
// 		w.WriteString(html)
// 		w.WriteString(`<br>`)
// 	}
// 	w.WriteString(`</div>`)
// }

// func (r *FileDirectiveRenderer) renderFileTemplate(templateName string, data map[string]interface{}) string {
// 	templates := map[string]string{
// 		"default": `<a href="{{.URL}}" class="file-link" data-file="{{.Filename}}">{{.Description}}</a> <span class="download-counter">({{.Downloads}})</span>`,
// 		"button":  `<button onclick="downloadFile('{{.URL}}', '{{.Filename}}')" class="download-btn">Скачать {{.Description}}</button> <span class="download-counter">({{.Downloads}})</span>`,
// 		"card": `<div class="file-card">
// 			<h4>{{.Description}}</h4>
// 			<a href="{{.URL}}" class="file-download" data-file="{{.Filename}}">Скачать</a>
// 			<span class="download-counter">Загружено: {{.Downloads}} раз</span>
// 		</div>`,
// 	}

// 	template, ok := templates[templateName]
// 	if !ok {
// 		template = templates["default"]
// 	}

// 	// Простая замена переменных (можно заменить на text/template)
// 	result := template
// 	for key, value := range data {
// 		placeholder := fmt.Sprintf("{{.%s}}", key)
// 		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
// 	}

// 	return result
// }
