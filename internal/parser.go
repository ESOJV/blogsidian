package internal

import (
	"bytes"
	"strings"

	goldmarkkatex "github.com/FurqanSoftware/goldmark-katex"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"gopkg.in/yaml.v3"
)

// ParsePost parses a markdown file with frontmatter into a Post.
// The content is converted to HTML.
func ParsePost(data []byte) (*Post, error) {
	splitFile := strings.SplitN(string(data), "---", 3)

	frontMatter := splitFile[1]
	postContents := splitFile[2]

	p := Post{}
	err := yaml.Unmarshal([]byte(frontMatter), &p)
	if err != nil {
		return nil, err
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			&goldmarkkatex.Extender{},
		),
	)

	var buf bytes.Buffer
	err = md.Convert([]byte(postContents), &buf)
	if err != nil {
		return nil, err
	}

	p.Content = buf.String()

	return &p, nil
}
