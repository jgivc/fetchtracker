package mdadapter

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

const (
	fileNameMinLength = 5
	filesLength       = 5
)

var (
	startSeq = []byte{'[', '['}
	endSeq   = []byte{']', ']'}
	descSeq  = []byte{'|'}

	wordFiles = []byte("FILES")
)

/*
 * Wiki link
 * [[filenzme.txt]]
 * [[filename.txt|Description]]
 */
type FileNodeParser struct{}

func NewFileDirectiveParser() parser.InlineParser {
	return &FileNodeParser{}
}

func (s *FileNodeParser) Trigger() []byte {
	return startSeq
}

func (s *FileNodeParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	b, _ := block.PeekLine()
	end := bytes.Index(b, endSeq)
	if end < 0 {
		return nil
	}

	line := bytes.TrimSpace(b[len(startSeq):end])
	if len(b) < fileNameMinLength {
		return nil
	}

	if bytes.Contains(b, []byte("FILES")) {
		fmt.Println("UUUUUUUUUUUUU")
	}

	fmt.Println(123, string(line))

	block.Advance(end + 2)

	if idx := bytes.Index(line, descSeq); idx > 0 {
		data := bytes.Split(line, descSeq)
		if len(data) > 1 {
			return &FileNode{
				Filename:    string(bytes.TrimSpace(data[0])),
				Description: string(bytes.TrimSpace(data[0])),
			}
		}
	}

	return &FileNode{
		Filename: string(bytes.TrimSpace(line)),
	}
}

type FilesBlockParser struct{}

// Trigger определяет символы, которые активируют парсер
func (p *FilesBlockParser) Trigger() []byte {
	return []byte{'[', '['}
}

// Open проверяет, может ли строка быть началом блока [[FILES]]
func (p *FilesBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	b, seg := reader.PeekLine()
	// fmt.Println("LINE:", string(line))
	end := bytes.Index(b, endSeq)
	if end < 0 {
		return nil, parser.NoChildren
	}

	line := bytes.TrimSpace(b[len(startSeq):end])
	if len(line) < filesLength {
		return nil, parser.NoChildren
	}

	if bytes.Equal(bytes.TrimSpace(line), wordFiles) {
		reader.Advance(seg.Len())

		return &FilesNode{}, parser.NoChildren
	}

	return nil, parser.NoChildren
}

// Continue проверяет, продолжается ли блок (в нашем случае это однострочный блок)
func (p *FilesBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

// Close завершает парсинг блока
func (p *FilesBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	// Ничего не делаем, блок уже готов
}

// CanInterruptParagraph определяет, может ли блок прервать параграф
func (p *FilesBlockParser) CanInterruptParagraph() bool {
	return true
}

// CanAcceptIndentedLine определяет, может ли блок принимать отступы
func (p *FilesBlockParser) CanAcceptIndentedLine() bool {
	return false
}
