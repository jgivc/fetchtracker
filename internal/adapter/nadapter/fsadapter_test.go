package nadapter

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/jgivc/fetchtracker/internal/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

type FSAdapterTestSuite struct {
	suite.Suite
	adapter *fsAdapter
	fs      afero.Fs
	cfg     *config.Config
	log     *slog.Logger
}

type testFile struct {
	Path    string
	Content string
}

// type testFS struct {
// 	Dirs  []string
// 	Files []*testFile
// }

func (s *FSAdapterTestSuite) makeFS(files ...*testFile) {
	dirMap := make(map[string]struct{})
	for _, file := range files {
		dirMap[filepath.Dir(file.Path)] = struct{}{}
	}

	for dir := range dirMap {
		s.fs.MkdirAll(dir, os.ModeDir)
	}

	for _, file := range files {
		afero.WriteFile(s.fs, file.Path, []byte(file.Content), os.ModeAppend)
	}
}

func (s *FSAdapterTestSuite) SetupTest() {
	s.fs = afero.NewMemMapFs()
	s.log = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	s.cfg = &config.Config{}
	s.cfg.SetDefaults()
}

func (s *FSAdapterTestSuite) TestEmptyFolder() {
	workDir := "/testdir"
	a, err := NewFSAdapterWithFS(s.fs, &config.FSAdapterConfig{
		WorkDir: workDir,
	}, s.log)
	s.NoError(err)
	a.fs.Mkdir(workDir, os.ModeDir)
	_, err = a.ToDownload(workDir)
	s.Error(err)
}

func (s *FSAdapterTestSuite) TestIndexFile() {
	workDir := "/testdir"
	s.makeFS(
		&testFile{workDir + "/test.txt", "Test text"},
	)

	s.cfg.IndexerConfig.WorkDir = workDir
	a, err := NewFSAdapterWithFS(s.fs, s.cfg.FSAdapterConfig(), s.log)
	s.NoError(err)
	d, err := a.ToDownload(workDir)
	s.NoError(err)
	s.NotNil(d)

	fmt.Println(1, d.PageContent)
}

func TestFSAdapterTestSuite(t *testing.T) {
	suite.Run(t, new(FSAdapterTestSuite))
}
