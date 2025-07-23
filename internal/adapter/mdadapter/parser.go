package mdadapter

import (
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// FileDirectiveParser парсит директивы файлов
type FileDirectiveParser struct{}

func NewFileDirectiveParser() parser.InlineParser {
	return &FileDirectiveParser{}
}

func (s *FileDirectiveParser) Trigger() []byte {
	// fmt.Println("sss")
	return []byte{'{', '{'}
}

func (s *FileDirectiveParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	// fmt.Println("sss1")

	re := regexp.MustCompile(`{{(?:\s+)?file:\s?([^\s]+)\s?}}`)
	if matches := re.FindSubmatch(line); matches != nil {
		// fmt.Println(string(matches[0]))
		block.Advance(len(matches[0]))
		return &FileDirective{
			Filename: string(matches[1]),
		}
	}

	// Паттерны для различных директив
	// patterns := map[string]*regexp.Regexp{
	// 	"file":            regexp.MustCompile(`^\{\{% file "([^"]+)"(?:\s+description="([^"]*)")?(?:\s+template="([^"]*)")? %\}\}`),
	// 	"files_pattern":   regexp.MustCompile(`^\{\{% files pattern="([^"]+)" %\}\}`),
	// 	"files_all":       regexp.MustCompile(`^\{\{% files:all %\}\}`),
	// 	"files_remaining": regexp.MustCompile(`^\{\{% files:remaining %\}\}`),
	// }

	// for directiveType, pattern := range patterns {
	// 	if matches := pattern.FindSubmatch(line); matches != nil {
	// 		node := &FileDirective{LinkType: directiveType}

	// 		switch directiveType {
	// 		case "file":
	// 			node.Filename = string(matches[1])
	// 			if len(matches) > 2 && len(matches[2]) > 0 {
	// 				node.Description = string(matches[2])
	// 			}
	// 			if len(matches) > 3 && len(matches[3]) > 0 {
	// 				node.Template = string(matches[3])
	// 			}
	// 		case "files_pattern":
	// 			node.Pattern = string(matches[1])
	// 		}

	// 		block.Advance(len(matches[0]))
	// 		return node
	// 	}
	// }

	return nil
}
