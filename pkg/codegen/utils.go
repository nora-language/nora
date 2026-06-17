package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/semantic"
)

var formatStr = fmt.Sprintf

// emit writes a formatted string to the current generator buffer
func (g *Generator) emit(format string, args ...interface{}) {
	if g.EnableDebug && g.currentFile != "" {
		// If the buffer currently ends with a newline (i.e. starting a new line in C),
		// inject the #line directive immediately to ensure all generated C lines for this statement map back to the same source line.
		if g.buf.Len() == 0 || g.buf.Bytes()[g.buf.Len()-1] == '\n' {
			g.buf.WriteString(fmt.Sprintf("#line %d \"%s\"\n", g.currentLine, g.currentFile))
		}
	}

	if len(args) == 0 {
		g.buf.WriteString(format)
	} else {
		g.buf.WriteString(formatStr(format, args...))
	}
	g.buf.WriteString("\n")
}

// emitLine adds a #line preprocessor directive for debugging
func (g *Generator) emitLine(node ast.Node) {
	if !g.EnableDebug || node == nil {
		return
	}
	pos := node.Pos()
	g.currentLine = pos.Line
	g.currentFile = normalizeDebugPath(pos.Filename)

	// Ensure we are at the start of a line for the preprocessor directive
	if g.buf.Len() > 0 {
		lastByte := g.buf.Bytes()[g.buf.Len()-1]
		if lastByte != '\n' {
			g.buf.WriteByte('\n')
		}
	}

	g.buf.WriteString(fmt.Sprintf("#line %d \"%s\"\n", g.currentLine, g.currentFile))
}

func (g *Generator) findSymbolByName(name string) *semantic.Symbol {
	if g.CurrentFunc != nil && g.CurrentFunc.DefNode != nil {
		if fnStmt, ok := g.CurrentFunc.DefNode.(*ast.FunctionStatement); ok {
			for _, p := range fnStmt.Parameters {
				if p.Name != nil && p.Name.Value == name {
					if sym := g.SemanticInfo.Defs[p.Name]; sym != nil {
						return sym
					}
				}
			}
		}
	}
	if g.CurrentLambda != nil {
		for _, p := range g.CurrentLambda.Parameters {
			if p.Name != nil && p.Name.Value == name {
				if sym := g.SemanticInfo.Defs[p.Name]; sym != nil {
					return sym
				}
			}
		}
	}

	for _, sym := range g.SemanticInfo.Uses {
		if sym != nil && sym.Name == name {
			return sym
		}
	}
	for _, sym := range g.SemanticInfo.Defs {
		if sym != nil && sym.Name == name {
			return sym
		}
	}
	return nil
}

func normalizeDebugPath(path string) string {
	cwd, _ := os.Getwd()
	if absPath, err := filepath.Abs(path); err == nil {
		if relPath, err := filepath.Rel(cwd, absPath); err == nil {
			// Convert to relative path using forward slashes (e.g., src/variables.nr)
			return strings.ReplaceAll(relPath, "\\", "/")
		}
	}
	return strings.ReplaceAll(path, "\\", "/")
}

