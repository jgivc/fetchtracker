package mdadapter

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"

	"github.com/jgivc/fetchtracker/internal/entity"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

const (
	TemplateResolverKey parser.ContextKey = iota
	FileResolverKey

	fileNameMinLength = 5
	filesLength       = 5
)

var (
	startSeq = []byte{'[', '['}
	endSeq   = []byte{']', ']'}
	descSeq  = []byte{'|'}

	wordFiles = []byte("FILES")

	ErrNoTemplateResolverError = errors.New("no template resolver")
	ErrNoFileResolverError     = errors.New("no file resolver")
	ErrNoTemplateError         = errors.New("cannot get template")
	ErrInvalidSyntax           = errors.New("invalid syntax")
)

type FileResolver interface {
	GetFile(fileName string) (*entity.File, error)
	GetFiles() []*entity.File
}

type TemplateResolver interface {
	GetFileTemplate() *template.Template
	GetFilesTemplate() *template.Template
}

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
		node := &FileNode{Error: ErrInvalidSyntax}

		return node
	}

	block.Advance(end + len(endSeq))
	filename, description := s.parseFileReference(line)

	return s.makeNode(filename, description, pc)
}

func (s *FileNodeParser) parseFileReference(line []byte) (string, string) {
	if idx := bytes.Index(line, descSeq); idx > 0 {
		data := bytes.Split(line, descSeq)
		if len(data) >= 2 {
			filename := string(bytes.TrimSpace(data[0]))
			description := string(bytes.TrimSpace(data[1]))
			return filename, description
		}
	}

	return string(bytes.TrimSpace(line)), ""
}

func (s *FileNodeParser) makeNode(filename, description string, pc parser.Context) *FileNode {
	node := &FileNode{Filename: filename}

	tr, ok := pc.Get(TemplateResolverKey).(TemplateResolver)
	if !ok {
		node.Error = ErrNoTemplateResolverError

		return node
	}

	fr, ok := pc.Get(FileResolverKey).(FileResolver)
	if !ok {
		node.Error = ErrNoFileResolverError

		return node
	}

	file, err := fr.GetFile(filename)
	if err != nil {
		node.Error = fmt.Errorf("cannot get file %s: %w", filename, err)

		return node
	}

	if description != "" {
		fileCopy := *file
		fileCopy.Description = description
		file = &fileCopy
	}

	tmpl := tr.GetFileTemplate()
	if tmpl == nil {
		node.Error = ErrNoTemplateError

		return node
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, file); err != nil {
		node.Error = fmt.Errorf("cannot execute template: %w", err)

		return node
	}

	node.HTML = buf.Bytes()

	return node
}

type FilesBlockParser struct{}

func (p *FilesBlockParser) Trigger() []byte {
	return []byte{'[', '['}
}

func (p *FilesBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	b, seg := reader.PeekLine()
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

		return p.makeNode(pc), parser.NoChildren
	}

	return nil, parser.NoChildren
}

func (p *FilesBlockParser) makeNode(pc parser.Context) *FilesNode {
	node := &FilesNode{}

	tr, ok := pc.Get(TemplateResolverKey).(TemplateResolver)
	if !ok {
		fmt.Println(222)
		node.Error = ErrNoTemplateResolverError

		return node
	}

	fr, ok := pc.Get(FileResolverKey).(FileResolver)
	if !ok {
		node.Error = ErrNoFileResolverError

		return node
	}

	files := fr.GetFiles()

	tmpl := tr.GetFilesTemplate()
	if tmpl == nil {
		node.Error = ErrNoTemplateError

		return node
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, files); err != nil {
		node.Error = fmt.Errorf("cannot execute files template: %w", err)
		return node
	}

	node.HTML = buf.Bytes()

	return node
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
