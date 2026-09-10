package agent

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

func (r *Report) RenderHTML() ([]byte, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.Table),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(r.Render()), &buf); err != nil {
		return nil, err
	}
	var page bytes.Buffer
	page.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>KerenScope Research Report</title>
<style>
  :root { color-scheme: light dark; }
  body { font-family: -apple-system, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; max-width: 880px; margin: 0 auto; padding: 40px 24px 80px; line-height: 1.65; color: #1a1c1f; background: #fafafa; }
  h1 { font-size: 1.9rem; border-bottom: 2px solid #2563eb; padding-bottom: 12px; }
  h2 { font-size: 1.3rem; margin-top: 2.2rem; }
  table { border-collapse: collapse; width: 100%; margin: 1rem 0; font-size: 0.92rem; }
  th, td { border: 1px solid #d8dbe0; padding: 8px 10px; text-align: left; }
  th { background: #eef1f6; }
  tr:nth-child(even) td { background: #f4f6f9; }
  code { background: #eef1f6; padding: 2px 5px; border-radius: 4px; font-size: 0.9em; }
  blockquote { border-left: 4px solid #2563eb; margin: 1rem 0; padding: 4px 16px; color: #444; background: #f0f4ff; }
  hr { border: none; border-top: 1px solid #d8dbe0; margin: 2.5rem 0; }
  .disclaimer { font-size: 0.85rem; color: #666; background: #fff6e6; border: 1px solid #f0d9a8; border-radius: 8px; padding: 12px 16px; }
  @media (prefers-color-scheme: dark) {
    body { color: #e6e8ea; background: #14161a; }
    th { background: #23262c; } th, td { border-color: #3a3e46; }
    tr:nth-child(even) td { background: #191c21; }
    code { background: #23262c; }
    blockquote { background: #1a2333; }
    .disclaimer { background: #2a2417; border-color: #4d4127; color: #c9c0a8; }
  }
</style>
</head>
<body>
`)
	page.WriteString(buf.String())
	page.WriteString(fmt.Sprintf(`
<footer class="disclaimer"><strong>Disclaimer:</strong> KerenScope is an information and analysis tool, not investment advice. Data: <a href="https://sectors.app">Sectors API</a>.</footer>
</body>
</html>`))
	return page.Bytes(), nil
}
