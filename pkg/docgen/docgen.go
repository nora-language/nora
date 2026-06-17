package docgen

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/nora-language/nora/pkg/lexer"
	"github.com/nora-language/nora/pkg/parser"
	"github.com/nora-language/nora/pkg/parser/ast"
)

type DocGenerator struct {
	PackageName string
	Functions   []*ast.FunctionStatement
	Types       []*ast.TypeStatement
	Vars        []*ast.VarStatement
}

func NewDocGenerator(files []*ast.File) *DocGenerator {
	gen := &DocGenerator{}

	for _, file := range files {
		for _, stmt := range file.Statements {
			switch s := stmt.(type) {
			case *ast.PackageStatement:
				if gen.PackageName == "" && s.Name != nil {
					gen.PackageName = s.Name.Value
				}
			case *ast.FunctionStatement:
				if s.Doc != nil && len(s.Doc.List) > 0 {
					gen.Functions = append(gen.Functions, s)
				}
			case *ast.TypeStatement:
				if s.Doc != nil && len(s.Doc.List) > 0 {
					gen.Types = append(gen.Types, s)
				}
			case *ast.VarStatement:
				if s.Doc != nil && len(s.Doc.List) > 0 {
					gen.Vars = append(gen.Vars, s)
				}
			}
		}
	}

	if gen.PackageName == "" {
		gen.PackageName = "main"
	}

	return gen
}

func (g *DocGenerator) GenerateMarkdown() string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("# Package `%s`\n\n", g.PackageName))

	if len(g.Types) > 0 {
		out.WriteString("## Types\n\n")
		for _, t := range g.Types {
			out.WriteString(fmt.Sprintf("### `%s`\n\n", t.String()))
			out.WriteString(t.Doc.Text() + "\n\n")
		}
	}

	if len(g.Vars) > 0 {
		out.WriteString("## Variables\n\n")
		for _, v := range g.Vars {
			out.WriteString(fmt.Sprintf("### `%s`\n\n", v.String()))
			out.WriteString(v.Doc.Text() + "\n\n")
		}
	}

	if len(g.Functions) > 0 {
		out.WriteString("## Functions\n\n")
		for _, fn := range g.Functions {
			out.WriteString(fmt.Sprintf("### `%s`\n\n", fn.String()))
			out.WriteString(fn.Doc.Text() + "\n\n")
		}
	}

	return out.String()
}

func (g *DocGenerator) GenerateHTML() string {
	tmpl := `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Package {{.PackageName}} Documentation</title>
<style>
	body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; max-width: 800px; margin: 40px auto; padding: 0 20px; color: #333; }
	pre { background: #f4f4f4; padding: 10px; border-radius: 5px; overflow-x: auto; font-family: monospace; }
	h1, h2, h3 { color: #222; }
	code { background: #f4f4f4; padding: 2px 5px; border-radius: 3px; font-family: monospace; }
	.doc-block { margin-bottom: 30px; border-bottom: 1px solid #eaeaea; padding-bottom: 20px; }
	.signature { font-weight: bold; font-size: 1.1em; margin-bottom: 10px; }
	.comment { color: #555; white-space: pre-wrap; }
</style>
</head>
<body>
	<h1>Package <code>{{.PackageName}}</code></h1>

	{{if .Types}}
	<h2>Types</h2>
	{{range .Types}}
	<div class="doc-block">
		<div class="signature"><pre>{{.String}}</pre></div>
		<div class="comment">{{.Doc.Text}}</div>
	</div>
	{{end}}
	{{end}}

	{{if .Vars}}
	<h2>Variables</h2>
	{{range .Vars}}
	<div class="doc-block">
		<div class="signature"><pre>{{.String}}</pre></div>
		<div class="comment">{{.Doc.Text}}</div>
	</div>
	{{end}}
	{{end}}

	{{if .Functions}}
	<h2>Functions</h2>
	{{range .Functions}}
	<div class="doc-block">
		<div class="signature"><pre>{{.String}}</pre></div>
		<div class="comment">{{.Doc.Text}}</div>
	</div>
	{{end}}
	{{end}}
</body>
</html>`

	t, err := template.New("doc").Parse(tmpl)
	if err != nil {
		return err.Error()
	}

	var out bytes.Buffer
	if err := t.Execute(&out, g); err != nil {
		return err.Error()
	}

	return out.String()
}

func ParsePackage(path string) ([]*ast.File, error) {
	files := []*ast.File{}

	info, err := os.Stat(path)
	if err != nil {
		if !strings.HasSuffix(path, ".nr") {
			path += ".nr"
			info, err = os.Stat(path)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("error accessing input path '%s': %v", path, err)
	}

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".nr") {
				fullPath := filepath.Join(path, entry.Name())
				input, err := os.ReadFile(fullPath)
				if err != nil {
					continue
				}
				l := lexer.New(string(input), fullPath)
				p := parser.New(l)
				files = append(files, p.Parse(fullPath))
			}
		}
	} else {
		input, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		l := lexer.New(string(input), path)
		p := parser.New(l)
		files = append(files, p.Parse(path))
	}

	return files, nil
}

func Serve(g *DocGenerator, port string) {
	html := g.GenerateHTML()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	fmt.Printf("Serving documentation for package '%s' at http://localhost:%s\n", g.PackageName, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("HTTP Server failed: %v\n", err)
	}
}
