package mdadapter

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

const (
	fileNameMinLength = 5
)

var (
	startSeq = []byte{'[', '['}
	endSeq   = []byte{']', ']'}
	descSeq  = []byte{'|'}
	allFiles = []byte("FILES")
)

/*
 * Wiki link
 * [[filenzme.txt]]
 * [[filename.txt|Description]]
 * [[FILES]] - all files
 */
type FileDirectiveParser struct{}

func NewFileDirectiveParser() parser.InlineParser {
	return &FileDirectiveParser{}
}

func (s *FileDirectiveParser) Trigger() []byte {
	return startSeq
}

func (s *FileDirectiveParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	b, _ := block.PeekLine()
	end := bytes.Index(b, endSeq)
	if end < 0 {
		return nil
	}

	line := bytes.TrimSpace(b[len(startSeq):end])
	if len(b) < fileNameMinLength {
		return nil
	}

	block.Advance(end + 2)

	if idx := bytes.Index(line, descSeq); idx > 0 {
		data := bytes.Split(line, descSeq)
		if len(data) > 1 {
			return &FileDirective{
				Filename:    string(bytes.TrimSpace(data[0])),
				Description: string(bytes.TrimSpace(data[0])),
			}
		}
	}

	if bytes.Equal(line, allFiles) {
		return &FileDirective{
			AllFiles: true,
		}
	}

	return &FileDirective{
		Filename: string(bytes.TrimSpace(line)),
	}
}
