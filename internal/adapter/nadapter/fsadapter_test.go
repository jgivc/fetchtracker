package nadapter

import (
	"flag"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/jgivc/fetchtracker/internal/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

func TestFCAdapter(t *testing.T) {
	appCFG := &config.Config{}
	appCFG.SetDefaults()
	appCFG.IndexerConfig.WorkDir = "/test"
	cfg := appCFG.FSAdapterConfig()

	// log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	testCases := []struct {
		name               string
		workDir            string
		files              map[string]string
		expectError        bool
		expectedGoldenFile string
		logEnabled         bool
	}{
		{
			name:        "Scenario 1: Empty folder",
			expectError: true,
		},
		{
			name:    "Scenario 2: Empty folder with index.html",
			workDir: "one",
			files: map[string]string{
				cfg.IndexPageFileName: "",
			},
			expectError: true,
		},
		// 		{
		// 			name:    "Scenario 3: Empty folder with index.html",
		// 			workDir: "one",
		// 			files: map[string]string{
		// 				cfg.DescFileName: `#Title
		// 				Test test test`,
		// 			},
		// 			expectError: true,
		// 		},
		// 		{
		// 			name:    "Scenario 4: Default index",
		// 			workDir: "one",
		// 			files: map[string]string{
		// 				"test1.txt": "test1 content",
		// 			},
		// 			expectedGoldenFile: "scenario4.golden.html",
		// 		},
		// 		{
		// 			name:    "Scenario 5: Custom index",
		// 			workDir: "one",
		// 			files: map[string]string{
		// 				"test1.txt": "test1 content",
		// 				"test2.txt": "test2 content",
		// 				cfg.IndexPageFileName: `<html>
		// 	<head>
		// 		<title>{{ .Title }}</title>
		// 		<link rel="stylesheet" href="{{ .URL }}/styles.css">
		// 	</head>
		// 	<body>
		// 		<ul>
		// 		{{ range .Files }}
		// 			<li>{{ .SourcePath }}</li>
		// 		{{ end }}
		// 		</ul>
		// 	</body>
		// </html>`,
		// 			},
		// 			expectedGoldenFile: "scenario5.golden.html",
		// 		},
		{
			name:    "Scenario 6: Default markdown template",
			workDir: "one",
			files: map[string]string{
				"test4.txt": "test4 content",
				"test5.txt": "test5 content",
				"test6.txt": "test6 content",
				cfg.DescFileName: `# Tilte

Test file content

## Files
[[FILES]]

## File

Here is one file [[test4.txt]] and text further...
Here is another file [[test5.txt|Test 5 file]] with description and text further...

`,
			},
			expectedGoldenFile: "scenario6.golden.html",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			var workdir string

			if tc.workDir != "" {
				workdir = filepath.Join(cfg.WorkDir, tc.workDir)
				err := fs.MkdirAll(workdir, os.ModeDir)
				require.NoError(t, err)
			}

			for path, content := range tc.files {
				err := afero.WriteFile(fs, filepath.Join(workdir, path), []byte(content), os.ModeAppend)
				require.NoError(t, err)
			}

			logW := io.Discard
			if tc.logEnabled {
				logW = os.Stderr
			}
			log := slog.New(slog.NewTextHandler(logW, &slog.HandlerOptions{}))

			adapter, err := NewFSAdapterWithFS(fs, cfg, log)
			require.NoError(t, err)

			download, err := adapter.ToDownload(workdir)
			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			goldenFilePath := filepath.Join("testdata", tc.expectedGoldenFile)

			if *update {
				t.Log("updating golden file:", goldenFilePath)
				err := os.WriteFile(goldenFilePath, []byte(download.PageContent), 0644)
				require.NoError(t, err, "failed to update golden file")
			}

			expectedHTML, err := os.ReadFile(goldenFilePath)
			require.NoError(t, err, "failed to read golden file")

			require.Equal(t, string(expectedHTML), download.PageContent)
		})
	}
}

// type FSAdapterTestSuite struct {
// 	suite.Suite
// 	adapter *fsAdapter
// 	fs      afero.Fs
// 	cfg     *config.Config
// 	log     *slog.Logger
// }

// type testFile struct {
// 	Path    string
// 	Content string
// }

// func (s *FSAdapterTestSuite) makeFS(files ...*testFile) {
// 	dirMap := make(map[string]struct{})
// 	for _, file := range files {
// 		dirMap[filepath.Dir(file.Path)] = struct{}{}
// 	}

// 	for dir := range dirMap {
// 		err := s.fs.MkdirAll(dir, os.ModeDir)
// 		s.Require().NoError(err)
// 	}

// 	for _, file := range files {
// 		err := afero.WriteFile(s.fs, file.Path, []byte(file.Content), os.ModeAppend)
// 		s.Require().NoError(err)
// 	}
// }

// func (s *FSAdapterTestSuite) SetupTest() {
// 	s.fs = afero.NewMemMapFs()
// 	s.log = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
// 	s.cfg = &config.Config{}
// 	s.cfg.SetDefaults()
// }

// func (s *FSAdapterTestSuite) TestEmptyFolder() {
// 	workDir := "/testdir"
// 	a, err := NewFSAdapterWithFS(s.fs, &config.FSAdapterConfig{
// 		WorkDir: workDir,
// 	}, s.log)
// 	s.NoError(err)
// 	a.fs.Mkdir(workDir, os.ModeDir)
// 	_, err = a.ToDownload(workDir)
// 	s.Error(err)
// }

// func (s *FSAdapterTestSuite) TestIndexFile() {
// 	workDir := "/testdir"
// 	s.makeFS(
// 		&testFile{workDir + "/test.txt", "Test text"},
// 	)

// 	s.cfg.IndexerConfig.WorkDir = workDir
// 	a, err := NewFSAdapterWithFS(s.fs, s.cfg.FSAdapterConfig(), s.log)
// 	s.NoError(err)
// 	d, err := a.ToDownload(workDir)
// 	s.NoError(err)
// 	s.NotNil(d)

// 	fmt.Println(1, d.PageContent)

// 	if *update {
// 		s.T().Log("updating golden file ########################################")
// 	}
// }

// func TestFSAdapterTestSuite(t *testing.T) {
// 	suite.Run(t, new(FSAdapterTestSuite))
// }
