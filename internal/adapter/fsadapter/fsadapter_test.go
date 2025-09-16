package fsadapter

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/jgivc/fetchtracker/internal/config"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
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
		{
			name:    "Scenario 3: Empty folder with description file",
			workDir: "one",
			files: map[string]string{
				cfg.DescFileName: `#Title
						Test test test`,
			},
			expectError: true,
		},
		{
			name:    "Scenario 4: Default index",
			workDir: "one",
			files: map[string]string{
				"test1.txt": "test1 content",
			},
			expectedGoldenFile: "scenario4.golden.html",
		},
		{
			name:    "Scenario 5: Custom index",
			workDir: "one",
			files: map[string]string{
				"test1.txt": "test1 content",
				"test2.txt": "test2 content",
				cfg.IndexPageFileName: `<html>
	<head>
		<title>{{ .Title }}</title>
		<link rel="stylesheet" href="{{ .URL }}/styles.css">
	</head>
	<body>
		<ul>
		{{ range .Files }}
			<li>{{ .SourcePath }}</li>
		{{ end }}
		</ul>
	</body>
</html>`,
			},
			expectedGoldenFile: "scenario5.golden.html",
		},
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
		{
			name:    "Scenario 7: Custom markdown template and default FILE/FILES template",
			workDir: "one",
			files: map[string]string{
				"test7.txt": "test7 content",
				"test8.txt": "test8 content",
				cfg.DescFileName: `# Share files
Test test test
## One file
Here is one file [[test7.txt]]

## All files
[[FILES]]
`,
				cfg.TemplateFileName: `<html>
	<head>
		<title>{{ .Title }}</title>
	</head>
	<body>
		{{ .ContentHTML }}
	</body>
</html>`,
			},
			expectedGoldenFile: "scenario7.golden.html",
		},
		{
			name:    "Scenario 8: Custom markdown template and FILE/FILES",
			workDir: "one",
			files: map[string]string{
				"test9.txt":  "test9 content",
				"test10.txt": "test10 content",
				cfg.DescFileName: `# Share files
Test test test
## One file
Here is one file [[test10.txt]]

## All files
[[FILES]]
`,
				cfg.TemplateFileName: `<html>
			<head>
				<title>{{ .Title }}</title>
			</head>
			<body>
				{{ .ContentHTML }}
			</body>
</html>

{{ define "FILE" }}
<a class="my-file-class">{{ .Name }}</a>
{{ end }}
{{ define "FILES" }}
<ul>
{{ range . }}
<li class="my-li"><a class="my-a">{{ .Name }}</a></li>
{{ end }}
</ul>
{{ end }}`,
			},
			expectedGoldenFile: "scenario8.golden.html",
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

func TestLocalization(t *testing.T) {
	// a := plural.Rule{}
	// plural.Rules[language.Russian] = a

	bundle := i18n.NewBundle(language.Russian)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile("./locales/ru.toml")
	bundle.MustLoadMessageFile("./locales/en.toml")
	fmt.Println(bundle.LanguageTags())

	tmpl, err := template.New("").Parse(`<!DOCTYPE html>
<html>
<body>
  <h2>{{ .FilesToDownload }}</h2>
  <p>{{ .TotalFiles }}</p>
</body>
</html>`)

	require.NoError(t, err)

	localizer := i18n.NewLocalizer(bundle, "ru")
	tf, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "TotalFiles",
			One:   "Total {{.PluralCount}} file.",
			Other: "Total {{.PluralCount}} files.",
		},
		PluralCount: 2,
		TemplateData: map[string]interface{}{
			"PluralCount": 2,
		},
	})
	require.NoError(t, err)

	ftd, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "FilesToDownload",
			Other: "Files to download 1",
		},
	})
	require.NoError(t, err)

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]interface{}{
		"TotalFiles":      tf,
		"FilesToDownload": ftd,
	})

	require.NoError(t, err)

	fmt.Println(buf.String())
}
