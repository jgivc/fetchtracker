package mdadapter

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// FilesExtension - расширение goldmark
type FilesExtension struct {
	// FileService FileService
}

func NewFilesExtension() goldmark.Extender {
	return &FilesExtension{}
}

// func NewFilesExtension(fs FileService) goldmark.Extender {
// 	return &FilesExtension{FileService: fs}
// }

func (e *FilesExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(NewFileDirectiveParser(), 500),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewFileDirectiveRenderer(), 500),
		),
	)
}

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
