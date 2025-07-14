package fsadapter

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

type FSAdapterTestSuite struct {
	suite.Suite
	adapter *fsAdapter
	fs      afero.Fs
	log     *slog.Logger
}

func (s *FSAdapterTestSuite) SetupTest() {
	s.fs = afero.NewMemMapFs()
	s.log = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func (s *FSAdapterTestSuite) TestEmptyFolder() {
	dirName := "/testdir"
	a, err := NewFSAdapterWithFS(s.fs, dirName, "description.md", "", nil, s.log)
	s.NoError(err)
	a.fs.Mkdir(dirName, os.ModeDir)
	_, err = a.ToDownload(dirName)
	s.Error(err)
}

func (s *FSAdapterTestSuite) TestFolderWithOneFileWithoutDescription() {
	root := "/"
	dirName := "testdir"
	dirPath := filepath.Join(root, dirName)
	fileName := "file1.txt"
	filePath := filepath.Join(dirPath, fileName)
	fileContent := []byte("Test file content")

	s.fs.Mkdir(dirPath, os.ModeDir)
	afero.WriteFile(s.fs, filePath, fileContent, os.ModeAppend)

	a, err := NewFSAdapterWithFS(s.fs, dirName, "description.md", "", nil, s.log)
	s.NoError(err)
	d, err := a.ToDownload(dirPath)

	s.Require().NoError(err)

	s.NotNil(d)
	fmt.Println(d)
	s.Equal(dirPath, d.SourcePath)
	s.Equal(dirName, d.Title)
	s.Len(d.Files, 1)

	f := d.Files[0]
	s.Equal(fileName, f.Name)
	s.Equal(int64(len(fileContent)), f.Size)
	s.Equal(filePath, f.SourcePath)
	fmt.Println(f)
}

func TestFSAdapterTestSuite(t *testing.T) {
	suite.Run(t, new(FSAdapterTestSuite))
}
