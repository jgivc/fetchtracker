package fsadapter

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
)

func TestGoldmark(t *testing.T) {
	src := []byte(`---
title: "Test page"
enabled: true
---
	
# Test test
test page
`)

	md := goldmark.New(
		goldmark.WithExtensions(
			&frontmatter.Extender{},
		),
	)

	var buf bytes.Buffer

	ctx := parser.NewContext()
	if err := md.Convert(src, &buf, parser.WithContext(ctx)); err != nil {
		t.Fatal(err)
	}

	fm := frontmatter.Get(ctx)

	var meta map[string]any
	if err := fm.Decode(&meta); err != nil {
		t.Fatal(err)
	}

	fmt.Println("Content:", buf.String())
	fmt.Println("fm:", meta)
}
