package format

import (
	"bytes"
	"sort"
	"strings"

	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/token"
)

type Formatter struct {
	config          *Config
	out             bytes.Buffer
	indent          int
	comments        []*ast.Comment
	commentIdx      int
	file            *ast.File
	noInlineComment bool
}

func New(config *Config) *Formatter {
	if config == nil {
		config = DefaultConfig()
	}
	return &Formatter{config: config}
}

func (f *Formatter) Format(file *ast.File) string {
	f.out.Reset()
	f.indent = 0

	var flat []*ast.Comment
	for _, cg := range file.Comments {
		if cg != nil {
			flat = append(flat, cg.List...)
		}
	}
	f.comments = flat

	f.commentIdx = 0
	f.file = file

	if f.config.OrganizeImports {
		f.organizeImports(file)
	}

	for i, stmt := range file.Statements {
		f.printStatement(stmt)
		if i < len(file.Statements)-1 {
			nextStmt := file.Statements[i+1]
			if f.hasBlankLineBetween(stmt, nextStmt) {
				f.out.WriteString("\n\n")
			} else {
				// Group contiguous vars or imports, otherwise separate with a blank line
				sameGroup := false
				switch stmt.(type) {
				case *ast.VarStatement:
					_, sameGroup = nextStmt.(*ast.VarStatement)
				case *ast.ImportStatement:
					_, sameGroup = nextStmt.(*ast.ImportStatement)
				}

				if sameGroup {
					f.out.WriteString("\n")
				} else {
					f.out.WriteString("\n\n")
				}
			}
		}
	}

	f.printCommentsBefore(token.Position{Line: 999999})

	result := f.out.String()
	result = normalizeBlankLines(result)
	if result != "" && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
}

// normalizeBlankLines collapses runs of empty lines to a single blank line.
func normalizeBlankLines(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	emptyRun := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			emptyRun++
			if emptyRun == 1 {
				out = append(out, "")
			}
			continue
		}
		emptyRun = 0
		out = append(out, strings.TrimRight(line, " \t"))
	}
	return strings.Join(out, "\n")
}

func (f *Formatter) hasBlankLineBetween(stmt1, stmt2 ast.Statement) bool {
	if f.file == nil || f.file.StmtEndLine == nil {
		return false
	}
	end1, ok1 := f.file.StmtEndLine[stmt1]
	start2 := stmt2.Pos().Line
	if !ok1 {
		return false
	}
	for _, bl := range f.file.BlankLines {
		if bl > end1 && bl < start2 {
			return true
		}
	}
	return false
}

func (f *Formatter) organizeImports(file *ast.File) {
	var imports []*ast.ImportStatement
	var firstImportIdx = -1

	for i, stmt := range file.Statements {
		if imp, ok := stmt.(*ast.ImportStatement); ok {
			imports = append(imports, imp)
			if firstImportIdx == -1 {
				firstImportIdx = i
			}
		} else if firstImportIdx == -1 {
			continue
		} else {
			break
		}
	}

	if len(imports) == 0 {
		return
	}

	sort.Slice(imports, func(i, j int) bool {
		return imports[i].Path.Value < imports[j].Path.Value
	})

	var pkgStmt *ast.PackageStatement
	var rest []ast.Statement
	for _, s := range file.Statements {
		if p, ok := s.(*ast.PackageStatement); ok {
			pkgStmt = p
		} else if _, ok := s.(*ast.ImportStatement); !ok {
			rest = append(rest, s)
		}
	}

	newStatements := make([]ast.Statement, 0, len(file.Statements))
	if pkgStmt != nil {
		newStatements = append(newStatements, pkgStmt)
	}
	for _, imp := range imports {
		newStatements = append(newStatements, imp)
	}
	newStatements = append(newStatements, rest...)
	file.Statements = newStatements
}

func (f *Formatter) infixLength(expr ast.Expression) int {
	if ast.IsNil(expr) {
		return 0
	}
	switch e := expr.(type) {
	case *ast.InfixExpression:
		return f.infixLength(e.Left) + len(e.Operator) + 2 + f.infixLength(e.Right)
	case *ast.Identifier:
		return len(e.Value)
	case *ast.IntegerLiteral:
		return len(e.Token.Literal)
	case *ast.FloatLiteral:
		return len(e.Token.Literal)
	default:
		return len(expr.String())
	}
}

func (f *Formatter) getStmtEndLine(node ast.Node) int {
	if ast.IsNil(node) {
		return 0
	}
	switch n := node.(type) {
	case *ast.BlockStatement:
		if n.Rbrace.Position.Line > 0 {
			return n.Rbrace.Position.Line
		}
		if len(n.Statements) > 0 {
			return f.getStmtEndLine(n.Statements[len(n.Statements)-1])
		}
		return n.Pos().Line
	case *ast.ImportStatement:
		return f.getStmtEndLine(n.Path)
	case *ast.PackageStatement:
		return n.Name.Pos().Line
	case *ast.FunctionStatement:
		if n.Body != nil {
			return f.getStmtEndLine(n.Body)
		}
		if n.ReturnType != nil {
			return f.getStmtEndLine(n.ReturnType)
		}
		if len(n.Parameters) > 0 {
			return f.getStmtEndLine(n.Parameters[len(n.Parameters)-1])
		}
		return n.Name.Pos().Line
	case *ast.TypeStatement:
		return f.getStmtEndLine(n.Value)
	case *ast.ReturnStatement:
		if n.ReturnValue == nil {
			return n.Pos().Line
		}
		return f.getStmtEndLine(n.ReturnValue)
	case *ast.VarStatement:
		if n.Value != nil {
			return f.getStmtEndLine(n.Value)
		}
		if n.Type != nil {
			return f.getStmtEndLine(n.Type)
		}
		return n.Name.Pos().Line
	case *ast.ExpressionStatement:
		return f.getStmtEndLine(n.Expression)
	case *ast.AssignmentStatement:
		if n.Value != nil {
			return f.getStmtEndLine(n.Value)
		}
		return f.getStmtEndLine(n.Left)
	case *ast.InfixExpression:
		return f.getStmtEndLine(n.Right)
	case *ast.CallExpression:
		if len(n.Arguments) > 0 {
			return f.getStmtEndLine(n.Arguments[len(n.Arguments)-1])
		}
		return f.getStmtEndLine(n.Function)
	case *ast.IfExpression:
		if n.Alternative != nil {
			return f.getStmtEndLine(n.Alternative)
		}
		return f.getStmtEndLine(n.Consequence)
	}

	if f.file != nil && f.file.StmtEndLine != nil {
		if stmt, ok := node.(ast.Statement); ok {
			if end, ok := f.file.StmtEndLine[stmt]; ok {
				return end
			}
		}
	}
	return node.Pos().Line
}

func (f *Formatter) hasCommentsInside(startLine, endLine int) bool {
	for _, c := range f.comments {
		if c.Pos().Line > startLine && c.Pos().Line < endLine {
			return true
		}
	}
	return false
}

func (f *Formatter) printInlineComment(line int) {
	for f.commentIdx < len(f.comments) {
		c := f.comments[f.commentIdx]
		if c.Pos().Line == line {
			f.out.WriteString(" ")
			f.out.WriteString(c.String())
			f.commentIdx++
		} else {
			break
		}
	}
}

func (f *Formatter) printAttributes(attrs []ast.Attribute) {
	if len(attrs) == 0 {
		return
	}
	for _, attr := range attrs {
		f.writeIndent()
		f.out.WriteString("[")
		f.out.WriteString(attr.Name)
		if len(attr.Args) > 0 {
			f.out.WriteString("(")
			for i, arg := range attr.Args {
				f.out.WriteString("\"")
				f.out.WriteString(arg)
				f.out.WriteString("\"")
				if i < len(attr.Args)-1 {
					f.out.WriteString(", ")
				}
			}
			f.out.WriteString(")")
		}
		f.out.WriteString("]\n")
	}
}

func (f *Formatter) currentLineLength() int {
	parts := strings.Split(f.out.String(), "\n")
	if len(parts) == 0 {
		return 0
	}
	return len(parts[len(parts)-1])
}

func (f *Formatter) printStatement(stmt ast.Statement) {
	if ast.IsNil(stmt) {
		return
	}

	f.printCommentsBefore(stmt.Pos())

	switch s := stmt.(type) {
	case *ast.PackageStatement:
		f.writeIndent()
		f.out.WriteString("package ")
		f.out.WriteString(s.Name.Value)

	case *ast.ImportStatement:
		f.writeIndent()
		f.out.WriteString("import ")
		if s.Alias != nil {
			f.out.WriteString(s.Alias.Value)
			f.out.WriteString(" ")
		}
		f.out.WriteString("\"")
		f.out.WriteString(s.Path.Value)
		f.out.WriteString("\"")

	case *ast.FunctionStatement:
		f.printFunction(s)

	case *ast.VarStatement:
		f.writeIndent()
		if s.IsPublic {
			f.out.WriteString("pub ")
		}
		f.out.WriteString("var ")
		f.out.WriteString(s.Name.Value)
		if s.Type != nil {
			f.out.WriteString(": ")
			f.printTypeNode(s.Type)
		}
		if s.Value != nil {
			f.out.WriteString(" = ")
			f.printExpression(s.Value)
		}

	case *ast.ExpressionStatement:
		f.writeIndent()
		f.printExpression(s.Expression)

	case *ast.ReturnStatement:
		f.writeIndent()
		f.out.WriteString("return")
		if s.ReturnValue != nil {
			f.out.WriteString(" ")
			f.printExpression(s.ReturnValue)
		}

	case *ast.BlockStatement:
		f.writeIndent()
		f.printBlock(s)

	case *ast.WhileStatement:
		f.writeIndent()
		f.out.WriteString("while ")
		f.printExpression(s.Condition)
		f.out.WriteString(" ")
		f.printBlock(s.Body)

	case *ast.ForStatement:
		f.printFor(s)

	case *ast.TypeStatement:
		f.printType(s)

	case *ast.AssignmentStatement:
		f.writeIndent()
		if s.Token.Type == token.INC {
			f.printExpression(s.Left)
			f.out.WriteString("++")
		} else if s.Token.Type == token.DEC {
			f.printExpression(s.Left)
			f.out.WriteString("--")
		} else {
			f.printExpression(s.Left)
			f.out.WriteString(" = ")
			f.printExpression(s.Value)
		}

	case *ast.BranchStatement:
		f.writeIndent()
		f.out.WriteString(s.Token.Literal)

	case *ast.BreakStatement:
		f.writeIndent()
		f.out.WriteString(s.Token.Literal)

	case *ast.ContinueStatement:
		f.writeIndent()
		f.out.WriteString(s.Token.Literal)

	case *ast.DeferStatement:
		f.writeIndent()
		f.out.WriteString("defer ")
		f.printExpression(s.Call)

	case *ast.SelectStatement:
		f.printSelect(s)

	case *ast.PinStatement:
		f.writeIndent()
		f.out.WriteString("pin ")
		for i, t := range s.Targets {
			f.out.WriteString(t.Value)
			if i < len(s.Targets)-1 {
				f.out.WriteString(", ")
			}
		}

	default:
		f.writeIndent()
		f.out.WriteString(stmt.String())
	}

	if !f.noInlineComment {
		f.printInlineComment(f.getStmtEndLine(stmt))
	}
}

func (f *Formatter) printExpression(expr ast.Expression) {
	if ast.IsNil(expr) {
		return
	}

	printedNewline := f.printCommentsBefore(expr.Pos())
	if printedNewline {
		f.writeIndent()
	}

	switch e := expr.(type) {
	case *ast.Identifier:
		f.out.WriteString(e.Value)
	case *ast.IntegerLiteral:
		f.out.WriteString(e.Token.Literal)
	case *ast.FloatLiteral:
		f.out.WriteString(e.Token.Literal)
	case *ast.ChanType, *ast.FunctionType:
		if typeNode, ok := expr.(ast.TypeNode); ok {
			f.printTypeNode(typeNode)
		} else {
			f.out.WriteString(expr.String())
		}
	case *ast.RangeExpression:
		f.printExpression(e.Start)
		f.out.WriteString("..")
		f.printExpression(e.End)
	case *ast.LambdaExpression:
		f.out.WriteString("fn(")
		for i, p := range e.Parameters {
			f.printParameter(p)
			if i < len(e.Parameters)-1 {
				f.out.WriteString(", ")
			}
		}
		f.out.WriteString(")")
		if e.ReturnType != nil {
			f.out.WriteString(" ")
			f.printTypeNode(e.ReturnType)
		}
		f.out.WriteString(" ")
		f.printBlock(e.Body)
	case *ast.StringLiteral:
		f.out.WriteString("\"")
		if e.Token.Literal != "" {
			f.out.WriteString(e.Token.Literal)
		} else {
			f.out.WriteString(e.Value)
		}
		f.out.WriteString("\"")
	case *ast.InterpolatedString:
		f.out.WriteString("\"")
		for _, part := range e.Parts {
			if sl, ok := part.(*ast.StringLiteral); ok {
				f.out.WriteString(sl.Value)
			} else {
				f.out.WriteString("${")
				f.printExpression(part)
				f.out.WriteString("}")
			}
		}
		f.out.WriteString("\"")
	case *ast.Boolean:
		f.out.WriteString(e.Token.Literal)
	case *ast.NoneLiteral:
		f.out.WriteString("none")
	case *ast.InfixExpression:
		f.printExpression(e.Left)
		if (e.Operator == "&&" || e.Operator == "||") && f.currentLineLength() > 50 {
			f.out.WriteString(" ")
			f.out.WriteString(e.Operator)
			f.out.WriteString("\n")
			f.writeIndent()
			f.out.WriteString("    ")
			f.printExpression(e.Right)
		} else {
			f.out.WriteString(" ")
			f.out.WriteString(e.Operator)
			f.out.WriteString(" ")
			f.printExpression(e.Right)
		}
	case *ast.IfExpression:
		f.printIf(e)
	case *ast.GroupedExpression:
		f.out.WriteString("(")
		f.printExpression(e.Expression)
		f.out.WriteString(")")
	case *ast.PrefixExpression:
		f.out.WriteString(e.Operator)
		f.printExpression(e.Right)
	case *ast.CallExpression:
		f.printExpression(e.Function)
		if len(e.TypeArguments) > 0 {
			f.out.WriteString("[")
			for i, ta := range e.TypeArguments {
				f.printTypeNode(ta)
				if i < len(e.TypeArguments)-1 {
					f.out.WriteString(", ")
				}
			}
			f.out.WriteString("]")
		}
		f.out.WriteString("(")
		f.indent++
		for i, arg := range e.Arguments {
			f.printExpression(arg)
			if i < len(e.Arguments)-1 {
				f.out.WriteString(", ")
			}
		}
		f.indent--
		f.out.WriteString(")")
	case *ast.ArgumentsExpression:
		if e.Name != nil {
			f.printExpression(e.Name)
			f.out.WriteString(": ")
		}
		f.printExpression(e.Value)
	case *ast.SelectorExpression:
		f.printExpression(e.Left)
		f.out.WriteString(".")
		f.printExpression(e.Field)
	case *ast.IndexExpression:
		f.printExpression(e.Left)
		f.out.WriteString("[")
		for i, idx := range e.Indices {
			f.printExpression(idx)
			if i < len(e.Indices)-1 {
				f.out.WriteString(", ")
			}
		}
		f.out.WriteString("]")
	case *ast.MatchExpression:
		f.printMatch(e)
	case *ast.BlockStatement:
		f.printBlock(e)
		f.printInlineComment(f.getStmtEndLine(e))
	case *ast.StructLiteral:
		f.printStructLiteral(e, false)
		f.printInlineComment(f.getStmtEndLine(e))
	case *ast.SumTypeLiteral:
		f.printSumTypeLiteral(e)
		f.printInlineComment(f.getStmtEndLine(e))
	case *ast.InterfaceLiteral:
		f.printInterfaceLiteral(e)
		f.printInlineComment(f.getStmtEndLine(e))
	case *ast.AllocExpression:
		f.out.WriteString("alloc ")
		f.printExpression(e.Value)
	case *ast.TryExpression:
		f.printExpression(e.Value)
		f.out.WriteString("?")
	case *ast.SpawnExpression:
		f.out.WriteString("spawn ")
		if e.Call != nil {
			f.printExpression(e.Call)
		} else if e.Body != nil {
			f.printBlock(e.Body)
			f.printInlineComment(f.getStmtEndLine(e.Body))
		}
	case *ast.SendExpression:
		f.printExpression(e.Left)
		f.out.WriteString(" <- ")
		f.printExpression(e.Right)
	case *ast.ReceiveExpression:
		f.out.WriteString("<-")
		f.printExpression(e.Value)
	case *ast.AssignmentStatement:
		if e.Token.Type == token.INC {
			f.printExpression(e.Left)
			f.out.WriteString("++")
		} else if e.Token.Type == token.DEC {
			f.printExpression(e.Left)
			f.out.WriteString("--")
		} else {
			f.printExpression(e.Left)
			f.out.WriteString(" = ")
			f.printExpression(e.Value)
		}
	case *ast.ParallelExpression:
		f.out.WriteString("parallel ")
		f.printBlock(e.Body)
		f.printInlineComment(f.getStmtEndLine(e.Body))
	case *ast.ArrayLiteral:
		f.out.WriteString("[")
		for i, el := range e.Elements {
			f.printExpression(el)
			if i < len(e.Elements)-1 {
				f.out.WriteString(", ")
			}
		}
		f.out.WriteString("]")
	case *ast.MapLiteral:
		f.out.WriteString("{")
		var keys []ast.Expression
		for k := range e.Pairs {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Pos().Offset < keys[j].Pos().Offset
		})
		for i, k := range keys {
			f.printExpression(k)
			f.out.WriteString(": ")
			f.printExpression(e.Pairs[k])
			if i < len(keys)-1 {
				f.out.WriteString(", ")
			}
		}
		f.out.WriteString("}")
	case *ast.RuneLiteral:
		f.out.WriteString(e.Token.Literal)
	case *ast.SliceExpression:
		if e.Start != nil {
			f.printExpression(e.Start)
		}
		f.out.WriteString(":")
		if e.End != nil {
			f.printExpression(e.End)
		}
	case *ast.ImaginaryLiteral:
		f.out.WriteString(e.Token.Literal)
	case *ast.StructInstantiation:
		f.printExpression(e.Type)
		f.out.WriteString(" {")
		var keys []string
		for k := range e.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			f.out.WriteString(k)
			f.out.WriteString(": ")
			f.printExpression(e.Fields[k])
			if i < len(keys)-1 {
				f.out.WriteString(", ")
			}
		}
		f.out.WriteString("}")
	default:
		f.out.WriteString(expr.String())
	}
}

func (f *Formatter) printTypeNode(t ast.TypeNode) {
	if ast.IsNil(t) {
		return
	}
	switch tn := t.(type) {
	case *ast.Identifier:
		f.out.WriteString(tn.Value)
	case *ast.GroupedExpression:
		f.out.WriteString("(")
		if typeNode, ok := tn.Expression.(ast.TypeNode); ok {
			f.printTypeNode(typeNode)
		} else {
			f.printExpression(tn.Expression)
		}
		f.out.WriteString(")")
	case *ast.IndexExpression:
		if leftType, ok := tn.Left.(ast.TypeNode); ok {
			f.printTypeNode(leftType)
		} else {
			f.printExpression(tn.Left)
		}
		f.out.WriteString("[")
		for i, idx := range tn.Indices {
			if idxType, ok := idx.(ast.TypeNode); ok {
				f.printTypeNode(idxType)
			} else {
				f.printExpression(idx)
			}
			if i < len(tn.Indices)-1 {
				f.out.WriteString(", ")
			}
		}
		f.out.WriteString("]")
	case *ast.PrefixExpression:
		f.out.WriteString(tn.Operator)
		if rightType, ok := tn.Right.(ast.TypeNode); ok {
			f.printTypeNode(rightType)
		} else {
			f.printExpression(tn.Right)
		}
	case *ast.ChanType:
		f.out.WriteString("chan[")
		f.printTypeNode(tn.Value)
		f.out.WriteString("]")
	case *ast.FunctionType:
		f.out.WriteString("fn(")
		for i, p := range tn.Parameters {
			f.printTypeNode(p)
			if i < len(tn.Parameters)-1 {
				f.out.WriteString(", ")
			}
		}
		f.out.WriteString(")")
		if tn.ReturnType != nil {
			f.out.WriteString(" ")
			f.printTypeNode(tn.ReturnType)
		}
	default:
		if e, ok := t.(ast.Expression); ok {
			f.printExpression(e)
		} else {
			f.out.WriteString(t.String())
		}
	}
}

func (f *Formatter) printBlock(block *ast.BlockStatement) {
	if block == nil {
		f.out.WriteString("{}")
		return
	}
	startLine := block.Pos().Line
	endLine := f.getStmtEndLine(block)
	if len(block.Statements) == 0 && !f.hasCommentsInside(startLine, endLine) {
		f.out.WriteString("{}")
		return
	}
	f.out.WriteString("{")
	f.printInlineComment(block.Pos().Line)
	f.out.WriteString("\n")
	f.indent++
	for i, stmt := range block.Statements {
		f.printStatement(stmt)
		if i < len(block.Statements)-1 {
			if f.hasBlankLineBetween(stmt, block.Statements[i+1]) {
				f.out.WriteString("\n\n")
			} else {
				f.out.WriteString("\n")
			}
		} else {
			f.out.WriteString("\n")
		}
	}
	f.printCommentsBefore(token.Position{Line: endLine})
	f.indent--
	f.writeIndent()
	f.out.WriteString("}")
}

func (f *Formatter) printMatch(me *ast.MatchExpression) {
	f.out.WriteString("match ")
	f.printExpression(me.Target)
	f.out.WriteString(" {")
	f.printInlineComment(me.Pos().Line)
	f.out.WriteString("\n")
	f.indent++
	for _, c := range me.Cases {
		f.printMatchCase(c)
		f.out.WriteString("\n")
	}
	f.printCommentsBefore(token.Position{Line: f.getStmtEndLine(me)})
	f.indent--
	f.writeIndent()
	f.out.WriteString("}")
}

func (f *Formatter) printMatchCase(mc *ast.MatchCase) {
	f.writeIndent()
	f.printExpression(mc.Pattern)
	f.out.WriteString(" => ")
	if mc.Body != nil {
		f.printBlock(mc.Body)
	}
}

func (f *Formatter) printSelect(ss *ast.SelectStatement) {
	f.writeIndent()
	f.out.WriteString("select {")
	f.printInlineComment(ss.Pos().Line)
	f.out.WriteString("\n")
	f.indent++
	for _, c := range ss.Cases {
		f.printSelectCase(c)
		f.out.WriteString("\n")
	}
	f.printCommentsBefore(token.Position{Line: f.getStmtEndLine(ss)})
	f.indent--
	f.writeIndent()
	f.out.WriteString("}")
}

func (f *Formatter) printSelectCase(sc *ast.SelectCase) {
	f.writeIndent()
	if sc.Condition == nil {
		f.out.WriteString("default:")
		f.printInlineComment(sc.Pos().Line)
	} else {
		f.out.WriteString("case ")
		oldIndent := f.indent
		f.indent = 0
		f.noInlineComment = true
		f.printStatement(sc.Condition)
		f.noInlineComment = false
		f.indent = oldIndent
		f.out.WriteString(":")
		f.printInlineComment(f.getStmtEndLine(sc.Condition))
	}
	if sc.Body != nil && len(sc.Body.Statements) > 0 {
		f.out.WriteString("\n")
		f.indent++
		for i, stmt := range sc.Body.Statements {
			f.printStatement(stmt)
			if i < len(sc.Body.Statements)-1 {
				f.out.WriteString("\n")
			}
		}
		f.indent--
	}
}

func (f *Formatter) printStructLiteral(sl *ast.StructLiteral, _ bool) {
	if sl.Name != nil {
		f.printExpression(sl.Name)
		f.out.WriteString(" {")
		f.printInlineComment(sl.Pos().Line)
		f.out.WriteString("\n")
	} else {
		f.out.WriteString("struct {")
		f.printInlineComment(sl.Pos().Line)
		f.out.WriteString("\n")
	}
	f.indent++
	for i, field := range sl.Fields {
		f.writeIndent()
		f.out.WriteString(field.Name.Value)
		f.out.WriteString(": ")
		if field.Type != nil {
			f.printTypeNode(field.Type)
		} else if field.Value != nil {
			f.printExpression(field.Value)
		}
		if i < len(sl.Fields)-1 {
			f.out.WriteString(",")
		}
		f.printInlineComment(f.getStmtEndLine(field))
		f.out.WriteString("\n")
	}
	f.printCommentsBefore(token.Position{Line: f.getStmtEndLine(sl)})
	f.indent--
	f.writeIndent()
	f.out.WriteString("}")
}

func (f *Formatter) printSumTypeLiteral(sl *ast.SumTypeLiteral) {
	f.out.WriteString("enum {")
	f.printInlineComment(sl.Pos().Line)
	f.out.WriteString("\n")
	f.indent++
	for i, v := range sl.Variants {
		f.printVariant(v)
		if i < len(sl.Variants)-1 {
			f.printInlineComment(f.getStmtEndLine(v))
			f.out.WriteString(",\n")
		} else {
			f.printInlineComment(f.getStmtEndLine(v))
			f.out.WriteString("\n")
		}
	}
	f.printCommentsBefore(token.Position{Line: f.getStmtEndLine(sl)})
	f.indent--
	f.writeIndent()
	f.out.WriteString("}")
}

func (f *Formatter) printVariant(v *ast.VariantDefinition) {
	f.writeIndent()
	f.out.WriteString(v.Name.Value)
	if len(v.Fields) > 0 {
		f.out.WriteString("(")
		for i, field := range v.Fields {
			f.out.WriteString(field.Name.Value)
			f.out.WriteString(": ")
			if field.Type != nil {
				f.printTypeNode(field.Type)
			} else if field.Value != nil {
				f.printExpression(field.Value)
			}
			if i < len(v.Fields)-1 {
				f.out.WriteString(", ")
			}
		}
		f.out.WriteString(")")
	}
}

func (f *Formatter) printInterfaceLiteral(il *ast.InterfaceLiteral) {
	f.out.WriteString("interface {")
	f.printInlineComment(il.Pos().Line)
	f.out.WriteString("\n")
	f.indent++
	for _, emb := range il.Embedded {
		f.writeIndent()
		f.printExpression(emb)
		f.printInlineComment(f.getStmtEndLine(emb))
		f.out.WriteString("\n")
	}
	for _, m := range il.Methods {
		f.writeIndent()
		f.printFunctionSignature(m)
		f.printInlineComment(f.getStmtEndLine(m))
		f.out.WriteString("\n")
	}
	f.printCommentsBefore(token.Position{Line: f.getStmtEndLine(il)})
	f.indent--
	f.writeIndent()
	f.out.WriteString("}")
}

func (f *Formatter) printFunction(fn *ast.FunctionStatement) {
	f.printAttributes(fn.Attributes)
	f.writeIndent()
	if fn.IsPublic {
		f.out.WriteString("pub ")
	}
	if fn.IsExport {
		f.out.WriteString("export ")
	}
	if fn.IsExtern {
		f.out.WriteString("extern ")
	}
	f.out.WriteString("fn ")
	if fn.Receiver != nil {
		f.out.WriteString("(")
		f.printParameter(fn.Receiver)
		f.out.WriteString(") ")
	}
	f.out.WriteString(fn.Name.Value)

	if len(fn.TypeParameters) > 0 {
		f.out.WriteString("[")
		for i, tp := range fn.TypeParameters {
			f.out.WriteString(tp.Name.Value)
			if tp.Constraint != nil {
				f.out.WriteString(": ")
				f.printTypeNode(tp.Constraint)
			}
			if i < len(fn.TypeParameters)-1 {
				f.out.WriteString(", ")
			}
		}
		f.out.WriteString("]")
	}

	f.out.WriteString("(")
	for i, p := range fn.Parameters {
		f.printParameter(p)
		if i < len(fn.Parameters)-1 {
			f.out.WriteString(", ")
		}
	}
	f.out.WriteString(")")

	if fn.ReturnType != nil {
		f.out.WriteString(" ")
		f.printTypeNode(fn.ReturnType)
	}

	if fn.Body != nil {
		f.out.WriteString(" ")
		f.printBlock(fn.Body)
	}
}

func (f *Formatter) printFunctionSignature(fn *ast.FunctionStatement) {
	f.out.WriteString("fn ")
	if fn.Receiver != nil {
		f.out.WriteString("(")
		f.printParameter(fn.Receiver)
		f.out.WriteString(") ")
	}
	f.out.WriteString(fn.Name.Value)
	if len(fn.TypeParameters) > 0 {
		f.out.WriteString("[")
		for i, tp := range fn.TypeParameters {
			f.out.WriteString(tp.Name.Value)
			if i < len(fn.TypeParameters)-1 {
				f.out.WriteString(", ")
			}
		}
		f.out.WriteString("]")
	}
	f.out.WriteString("(")
	for i, p := range fn.Parameters {
		f.printParameter(p)
		if i < len(fn.Parameters)-1 {
			f.out.WriteString(", ")
		}
	}
	f.out.WriteString(")")
	if fn.ReturnType != nil {
		f.out.WriteString(" ")
		f.printTypeNode(fn.ReturnType)
	}
}

func (f *Formatter) printParameter(p *ast.Parameter) {
	if p == nil {
		return
	}
	if p.Name.Value == "..." {
		f.out.WriteString("...")
		return
	}
	f.out.WriteString(p.Name.Value)
	f.out.WriteString(": ")
	if p.IsVariadic {
		f.out.WriteString("...")
		if p.Type != nil {
			f.out.WriteString(" ")
			f.printTypeNode(p.Type)
		}
		return
	}
	if p.Type != nil {
		switch p.LeaseKind {
		case ast.LeaseWrite:
			if _, ok := p.Type.(*ast.PrefixExpression); !ok {
				f.out.WriteString("&")
			}
		case ast.LeaseMove:
			if _, ok := p.Type.(*ast.PrefixExpression); !ok {
				f.out.WriteString("@")
			}
		}
		f.printTypeNode(p.Type)
	}
}

func (f *Formatter) printIf(expr *ast.IfExpression) {
	f.out.WriteString("if ")
	f.printExpression(expr.Condition)
	f.out.WriteString(" ")
	f.printBlock(expr.Consequence)
	if expr.Alternative != nil {
		f.out.WriteString(" else ")
		if altIf, ok := expr.Alternative.(*ast.IfExpression); ok {
			f.printIf(altIf)
		} else if altBlock, ok := expr.Alternative.(*ast.BlockStatement); ok {
			f.printBlock(altBlock)
		} else {
			f.printExpression(expr.Alternative)
		}
	}
}

func (f *Formatter) printFor(s *ast.ForStatement) {
	f.writeIndent()
	f.out.WriteString("for ")
	if s.Value != nil {
		if s.Key != nil {
			f.out.WriteString(s.Key.Value)
			f.out.WriteString(", ")
		}
		f.out.WriteString(s.Value.Value)
		f.out.WriteString(" in ")
		f.printExpression(s.Iterable)
	}
	f.out.WriteString(" ")
	f.printBlock(s.Body)
}

func (f *Formatter) printType(s *ast.TypeStatement) {
	f.printAttributes(s.Attributes)
	f.writeIndent()
	if s.IsPublic {
		f.out.WriteString("pub ")
	}
	f.out.WriteString("type ")
	f.out.WriteString(s.Name.Value)
	if len(s.TypeParameters) > 0 {
		f.out.WriteString("[")
		for i, tp := range s.TypeParameters {
			f.out.WriteString(tp.Name.Value)
			if tp.Constraint != nil {
				f.out.WriteString(": ")
				f.printTypeNode(tp.Constraint)
			}
			if i < len(s.TypeParameters)-1 {
				f.out.WriteString(", ")
			}
		}
		f.out.WriteString("]")
	}
	f.out.WriteString(" = ")
	f.printExpression(s.Value)
}

func (f *Formatter) printCommentsBefore(pos token.Position) bool {
	printed := false
	for f.commentIdx < len(f.comments) {
		c := f.comments[f.commentIdx]
		if c.Pos().Line < pos.Line {
			f.writeIndent()
			f.out.WriteString(c.String())
			f.out.WriteString("\n")
			f.commentIdx++
			printed = true
		} else {
			break
		}
	}
	return printed
}

func (f *Formatter) writeIndent() {
	if f.config.UseTabs {
		f.out.WriteString(strings.Repeat("\t", f.indent))
	} else {
		f.out.WriteString(strings.Repeat(" ", f.indent*f.config.IndentSize))
	}
}
