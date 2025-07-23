package mdadapter

// package main

// import (
// 	"bytes"
// 	"fmt"
// 	"regexp"
// 	"strings"

// 	"github.com/yuin/goldmark"
// 	"github.com/yuin/goldmark/ast"
// 	"github.com/yuin/goldmark/parser"
// 	"github.com/yuin/goldmark/renderer"
// 	"github.com/yuin/goldmark/renderer/html"
// 	"github.com/yuin/goldmark/text"
// 	"github.com/yuin/goldmark/util"
// )

// // FileDirective представляет директиву для файлов
// type FileDirective struct {
// 	ast.BaseInline
// 	Filename    string
// 	Description string
// 	Template    string
// 	Pattern     string
// 	Type        string // "file", "files", "files:remaining", "files:all"
// }

// // Dump реализует интерфейс ast.Node
// func (n *FileDirective) Dump(source []byte, level int) {
// 	ast.DumpHelper(n, source, level, map[string]string{
// 		"Type":        n.Type,
// 		"Filename":    n.Filename,
// 		"Description": n.Description,
// 		"Template":    n.Template,
// 		"Pattern":     n.Pattern,
// 	}, nil)
// }

// // Kind реализует интерфейс ast.Node
// func (n *FileDirective) Kind() ast.NodeKind {
// 	return KindFileDirective
// }

// var KindFileDirective = ast.NewNodeKind("FileDirective")

// // FileDirectiveParser парсит директивы файлов
// type FileDirectiveParser struct{}

// func NewFileDirectiveParser() parser.InlineParser {
// 	return &FileDirectiveParser{}
// }

// func (s *FileDirectiveParser) Trigger() []byte {
// 	return []byte{'{'}
// }

// func (s *FileDirectiveParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
// 	line, _ := block.PeekLine()

// 	// Паттерны для различных директив
// 	patterns := map[string]*regexp.Regexp{
// 		"file": regexp.MustCompile(`^\{\{% file "([^"]+)"(?:\s+description="([^"]*)")?(?:\s+template="([^"]*)")? %\}\}`),
// 		"files_pattern": regexp.MustCompile(`^\{\{% files pattern="([^"]+)" %\}\}`),
// 		"files_all": regexp.MustCompile(`^\{\{% files:all %\}\}`),
// 		"files_remaining": regexp.MustCompile(`^\{\{% files:remaining %\}\}`),
// 	}

// 	for directiveType, pattern := range patterns {
// 		if matches := pattern.FindSubmatch(line); matches != nil {
// 			node := &FileDirective{Type: directiveType}

// 			switch directiveType {
// 			case "file":
// 				node.Filename = string(matches[1])
// 				if len(matches) > 2 && len(matches[2]) > 0 {
// 					node.Description = string(matches[2])
// 				}
// 				if len(matches) > 3 && len(matches[3]) > 0 {
// 					node.Template = string(matches[3])
// 				}
// 			case "files_pattern":
// 				node.Pattern = string(matches[1])
// 			}

// 			block.Advance(len(matches[0]))
// 			return node
// 		}
// 	}

// 	return nil
// }

// // FileDirectiveRenderer рендерит директивы файлов
// type FileDirectiveRenderer struct {
// 	FileService FileService
// }

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

// 	switch directive.Type {
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

// // FilesExtension - расширение goldmark
// type FilesExtension struct {
// 	FileService FileService
// }

// func NewFilesExtension(fs FileService) goldmark.Extender {
// 	return &FilesExtension{FileService: fs}
// }

// func (e *FilesExtension) Extend(m goldmark.Markdown) {
// 	m.Parser().AddOptions(
// 		parser.WithInlineParsers(
// 			util.Prioritized(NewFileDirectiveParser(), 500),
// 		),
// 	)
// 	m.Renderer().AddOptions(
// 		renderer.WithNodeRenderers(
// 			util.Prioritized(NewFileDirectiveRenderer(e.FileService), 500),
// 		),
// 	)
// }

// // Пример использования
// func main() {
// 	// Реализация FileService
// 	fs := &MyFileService{
// 		baseURL: "https://example.com/download/",
// 		// ... другие поля
// 	}

// 	md := goldmark.New(
// 		goldmark.WithExtensions(
// 			NewFilesExtension(fs),
// 		),
// 		goldmark.WithRendererOptions(
// 			html.WithHardWraps(),
// 			html.WithXHTML(),
// 		),
// 	)

// 	source := `# Моя раздача

// Описание...

// ## Основные файлы
// {{% file "document.pdf" description="Главный документ" template="button" %}}
// {{% file "archive.zip" %}}

// ## Дополнительные файлы
// {{% files pattern="*.txt" %}}

// ## Все остальные
// {{% files:remaining %}}
// `

// 	var buf bytes.Buffer
// 	if err := md.Convert([]byte(source), &buf); err != nil {
// 		panic(err)
// 	}

// 	fmt.Println(buf.String())
// }

// // Пример реализации FileService
// type MyFileService struct {
// 	baseURL        string
// 	availableFiles []string
// 	processedFiles map[string]bool
// 	// redis client, etc.
// }

// func (fs *MyFileService) GetFileURL(filename string) string {
// 	return fs.baseURL + filename
// }

// func (fs *MyFileService) GetFileDownloads(filename string) int {
// 	// Запрос к Redis
// 	return 42
// }

// func (fs *MyFileService) GetAvailableFiles() []string {
// 	return fs.availableFiles
// }

// func (fs *MyFileService) GetFilesByPattern(pattern string) []string {
// 	// Реализация поиска по паттерну
// 	var matches []string
// 	// ... логика фильтрации
// 	return matches
// }

// func (fs *MyFileService) MarkFileAsProcessed(filename string) {
// 	if fs.processedFiles == nil {
// 		fs.processedFiles = make(map[string]bool)
// 	}
// 	fs.processedFiles[filename] = true
// }

// func (fs *MyFileService) GetRemainingFiles() []string {
// 	var remaining []string
// 	for _, file := range fs.availableFiles {
// 		if !fs.processedFiles[file] {
// 			remaining = append(remaining, file)
// 		}
// 	}
// 	return remaining
// }
