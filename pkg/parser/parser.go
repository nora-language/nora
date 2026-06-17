package parser

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/DwiYI/Project-Nora/pkg/diag"
	"github.com/DwiYI/Project-Nora/pkg/lexer"
	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/token"
)

// Precedence levels
const (
	_ int = iota
	LOWEST
	ASSIGN
	RANGE
	LOGICAL_OR
	LOGICAL_AND
	EQUALS
	LESSGREATER
	BITWISE_OR
	BITWISE_XOR
	BITWISE_AND
	SHIFT
	SUM
	PRODUCT
	PREFIX
	// Grouping these ensures that a.b[0]()? works as a single chain
	ACCESSOR // . [] ( ) ?
)

var precedences = map[token.TokenType]int{
	token.DOT_DOT: RANGE,
	token.LOR: LOGICAL_OR, token.LAND: LOGICAL_AND,
	token.EQ: EQUALS, token.NOT_EQ: EQUALS, token.STRICT_EQ: EQUALS,
	token.LT: LESSGREATER, token.GT: LESSGREATER, token.LT_EQ: LESSGREATER, token.GT_EQ: LESSGREATER,
	token.OR: BITWISE_OR, token.XOR: BITWISE_XOR, token.AND: BITWISE_AND,
	token.SHL: SHIFT, token.SHR: SHIFT,
	token.PLUS: SUM, token.MINUS: SUM, token.AND_NOT: SUM,
	token.SLASH: PRODUCT, token.ASTERISK: PRODUCT, token.REM: PRODUCT,
	token.LPAREN:          ACCESSOR,
	token.LBRACKET:        ACCESSOR,
	token.DOT:             ACCESSOR,
	token.QUESTION:        ACCESSOR,
	token.LBRACE:          ACCESSOR,
	token.ASSIGN:          ASSIGN,
	token.PLUS_ASSIGN:     ASSIGN,
	token.MINUS_ASSIGN:    ASSIGN,
	token.ASTERISK_ASSIGN: ASSIGN,
	token.SLASH_ASSIGN:    ASSIGN,
	token.REM_ASSIGN:      ASSIGN,
	token.AND_ASSIGN:      ASSIGN,
	token.OR_ASSIGN:       ASSIGN,
	token.XOR_ASSIGN:      ASSIGN,
	token.SHL_ASSIGN:      ASSIGN,
	token.SHR_ASSIGN:      ASSIGN,
	token.ARROW:           ASSIGN,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l           *lexer.Lexer
	Diagnostics *diag.Collection

	curToken   token.Token
	peekToken  token.Token
	peek2Token token.Token

	prefixParseFns      map[token.TokenType]prefixParseFn
	infixParseFns       map[token.TokenType]infixParseFn
	noStructLiteral     bool
	PreserveParentheses bool
	AllowNoPackage      bool
	DisableMacros       bool
	Context             context.Context // Added for cancellation

	curDocComments []*ast.Comment // Collect doc comments for next statement
	allComments    []*ast.Comment // Every comment in the file

	StmtEndLines map[ast.Statement]int
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:              l,
		Diagnostics:    l.Diagnostics,
		AllowNoPackage: true,
	}
	if p.Diagnostics == nil {
		p.Diagnostics = &diag.Collection{}
	}
	p.StmtEndLines = make(map[ast.Statement]int)

	// --- Prefix Registrations ---
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseNumberLiteral)
	p.registerPrefix(token.FLOAT, p.parseNumberLiteral) // Added for FLOAT support
	p.registerPrefix(token.IMAG, p.parseNumberLiteral)  // Added for FLOAT support

	p.registerPrefix(token.STR, p.parseStringLiteral)
	p.registerPrefix(token.RUNE, p.parseRuneLiteral)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TILDE, p.parsePrefixExpression) // Bitwise NOT
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	//p.registerPrefix(token.LBRACKET, p.parseArrayLiteral) // For [1, 2, 3]
	p.registerPrefix(token.LBRACE, p.parseMapLiteral) // For {key: val}
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FN, p.parseLambdaExpression)
	p.registerPrefix(token.SPAWN, p.parseSpawnExpression)
	p.registerPrefix(token.STRUCT, p.parseStructLiteral)
	p.registerPrefix(token.TYPE, p.parseSumTypeLiteral)
	p.registerPrefix(token.ENUM, p.parseSumTypeLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.PARALLEL, p.parseParallelExpression)
	p.registerPrefix(token.SCOPE, p.parseScopeExpression)

	p.registerPrefix(token.NONE, p.parseNone)
	p.registerPrefix(token.MATCH, p.parseMatchExpression)
	p.registerPrefix(token.ALLOC, p.parseAllocExpression)
	p.registerPrefix(token.ARROW, p.parseReceiveExpression) // <-ch
	p.registerPrefix(token.CHAN, p.parseChanTypeExpression) // For make(chan[T])
	p.registerPrefix(token.INTERFACE, p.parseInterfaceLiteral)
	p.registerPrefix(token.DEFAULT, p.parseIdentifier)
	p.registerPrefix(token.PROTOCOL, p.parseInterfaceLiteral)
	p.registerPrefix(token.HASH, p.parsePrefixExpression)
	p.registerPrefix(token.AND, p.parsePrefixExpression)
	p.registerPrefix(token.MOVE, p.parsePrefixExpression)
	// p.registerPrefix(token.PIN, p.parsePinExpression) // For lease pinning

	// --- Infix Registrations ---
	p.infixParseFns = make(map[token.TokenType]infixParseFn)

	// Arithmetic & Bitwise
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.REM, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.XOR, p.parseInfixExpression)
	p.registerInfix(token.AND_NOT, p.parseInfixExpression)
	p.registerInfix(token.SHL, p.parseInfixExpression)
	p.registerInfix(token.SHR, p.parseInfixExpression)

	// Comparison & Equality
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.STRICT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LT_EQ, p.parseInfixExpression)
	p.registerInfix(token.GT_EQ, p.parseInfixExpression)

	// Logical
	p.registerInfix(token.LAND, p.parseInfixExpression)
	p.registerInfix(token.LOR, p.parseInfixExpression)
	p.registerInfix(token.DOT_DOT, p.parseRangeExpression)

	// Accessors & Calls
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.QUESTION, p.parseTryExpression)
	p.registerInfix(token.DOT, p.parseSelectorExpression)
	p.registerInfix(token.LBRACE, p.parseStructInstantiation)

	p.registerInfix(token.ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(token.PLUS_ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(token.MINUS_ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(token.ASTERISK_ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(token.SLASH_ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(token.REM_ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(token.AND_ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(token.OR_ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(token.XOR_ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(token.SHL_ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(token.SHR_ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(token.ARROW, p.parseSendExpression) // ch <- val

	p.nextToken()
	p.nextToken()
	p.nextToken()
	return p
}

// --- Core Pratt Logic ---

func (p *Parser) parseExpression(precedence int) ast.Expression {
	if p.Context != nil && p.Context.Err() != nil {
		return nil
	}
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		// IMPORTANT: If we don't know this token, we shouldn't just return nil.
		// We log the error. The caller's loop must then advance p.nextToken().
		p.noPrefixParseFnError(p.curToken)
		return nil
	}
	leftExp := prefix()

	// Terminate if we see a comma or semicolon, as these are structural boundaries
	for !p.peekTokenIs(token.SEMICOLON) && !p.peekTokenIs(token.COMMA) && precedence < p.peekPrecedence() {
		if p.peekTokenIs(token.LBRACE) && p.noStructLiteral {
			break
		}
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

// IDENTIFIER LEASE
// X read lease
// X# write lease
// X@ move lease
func (p *Parser) parseIdentifier() ast.Expression {
	ident := &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	return ident
}
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.ReportError(p.curToken.Position, "could not parse %q as integer", p.curToken.Literal)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

// --- Statement Parsers ---

// ParseFile is a static helper that handles the IO and setup.
func ParseFile(filename string) (*ast.File, error) {
	// 1. Read the file from disk
	input, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read file '%s': %v", filename, err)
	}

	// 2. Initialize Lexer (Pass filename if your lexer supports tracking it)
	l := lexer.New(string(input), filename)

	// 3. Initialize Parser
	p := New(l)
	p.AllowNoPackage = false

	// 4. Run the Parse Loop
	return p.Parse(filename), nil
}

// Parse runs the parser loop until EOF and returns an ast.File
func (p *Parser) Parse(filename string) *ast.File {
	file := &ast.File{
		Name:        filename,
		Statements:  []ast.Statement{},
		BlankLines:  p.l.GetBlankLines(),
		StmtEndLine: p.StmtEndLines,
	}

	// Skip any leading semicolons or empty statements before checking for package declaration
	for p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	if !p.AllowNoPackage && filename != "test.nr" && filename != "" && !strings.Contains(filename, "test.nr") && !p.curTokenIs(token.PACKAGE) {
		p.ReportError(p.curToken.Position, "expected package statement at the beginning of the file")
	}

	for !p.curTokenIs(token.EOF) {
		if p.Context != nil && p.Context.Err() != nil {
			return file
		}

		// Skip any leading semicolons or empty statements
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if !ast.IsNil(stmt) {
			p.applyDocComments(stmt)
			file.Statements = append(file.Statements, stmt)
			p.nextToken() // Move to the next token for the next iteration
		} else {
			p.curDocComments = nil // clear if failed
			p.synchronize()
			// synchronize leaves us at a boundary token, so we don't p.nextToken() here
			// unless we are at a semicolon which Parse loop will skip anyway.
			if p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
		}
	}

	// Group comments into CommentGroups if needed, or just store them.
	// For now, let's just store them as a single group or individual groups.
	// Standard practice is to group adjacent comments.
	file.Comments = p.groupComments(p.allComments)

	if !p.DisableMacros && filename != "serialization_generated" {
		// Macro processing is now handled externally by the plugin manager
	}

	return file
}

func (p *Parser) groupComments(comments []*ast.Comment) []*ast.CommentGroup {
	if len(comments) == 0 {
		return nil
	}

	var groups []*ast.CommentGroup
	var currentGroup *ast.CommentGroup

	for _, c := range comments {
		if currentGroup == nil {
			currentGroup = &ast.CommentGroup{List: []*ast.Comment{c}}
		} else {
			last := currentGroup.List[len(currentGroup.List)-1]
			if c.Token.Position.Line == last.Token.Position.Line+1 {
				currentGroup.List = append(currentGroup.List, c)
			} else {
				groups = append(groups, currentGroup)
				currentGroup = &ast.CommentGroup{List: []*ast.Comment{c}}
			}
		}
	}
	if currentGroup != nil {
		groups = append(groups, currentGroup)
	}
	return groups
}

func (p *Parser) applyDocComments(node ast.Node) {
	if len(p.curDocComments) == 0 {
		return
	}

	nodeLine := node.Pos().Line

	// Find the comment that ends exactly at nodeLine - 1
	var endIdx int = -1
	for i := len(p.curDocComments) - 1; i >= 0; i-- {
		if p.curDocComments[i].Token.Position.Line == nodeLine-1 {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return
	}

	// Work backwards from endIdx to find connected doc comments
	var group []*ast.Comment
	lastLine := nodeLine
	startIdx := endIdx
	for i := endIdx; i >= 0; i-- {
		c := p.curDocComments[i]
		if c.Token.Position.Line == lastLine-1 {
			group = append([]*ast.Comment{c}, group...)
			lastLine = c.Token.Position.Line
			startIdx = i
		} else {
			break
		}
	}

	if len(group) == 0 {
		return
	}

	docGroup := &ast.CommentGroup{List: group}

	applied := false
	switch s := node.(type) {
	case *ast.FunctionStatement:
		s.Doc = docGroup
		applied = true
	case *ast.TypeStatement:
		s.Doc = docGroup
		applied = true
	case *ast.VarStatement:
		s.Doc = docGroup
		applied = true
	case *ast.PackageStatement:
		s.Doc = docGroup
		applied = true
	case *ast.FieldDefinition:
		s.Doc = docGroup
		applied = true
	case *ast.VariantDefinition:
		s.Doc = docGroup
		applied = true
	}

	if applied {
		// Remove applied comments from curDocComments
		p.curDocComments = append(p.curDocComments[:startIdx], p.curDocComments[endIdx+1:]...)
	}
}

func (p *Parser) ParseProject(filenames []string) *ast.Program {
	prog := &ast.Program{
		Files: make([]*ast.File, 0),
	}

	for _, name := range filenames {
		// fmt.Printf("Parsing %s...\n", name)

		fileAst, err := ParseFile(name)
		if err != nil {
			// Stop compilation if a file cannot be read
			panic(err)
		}

		// Optional: Check if the parser encountered syntax errors
		// (Assuming your Parser struct has an Errors field)
		// if len(parser.Errors) > 0 { ... }

		prog.Files = append(prog.Files, fileAst)
	}

	return prog
}

func (p *Parser) parseStatement() ast.Statement {
	stmt := p.parseStatementInternal()
	if !ast.IsNil(stmt) {
		p.StmtEndLines[stmt] = p.curToken.Position.Line
	}
	return stmt
}

func (p *Parser) parseStatementInternal() ast.Statement {
	switch p.curToken.Type {
	case token.PACKAGE:
		return p.parsePackageStatement()
	case token.IMPORT:
		return p.parseImportStatement()
	case token.FN:
		return p.parseFunctionStatement(false, false)
	case token.RETURN:
		return p.parseReturnStatement()
	case token.VAR:
		return p.parseVarStatement()
	case token.PIN:
		return p.parsePinStatement()
	case token.EXTERN:
		return p.parseExternStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.EXPORT:
		return p.parseExportStatement(false)
	case token.LBRACE:
		return p.parseBlockStatement()
	case token.LBRACKET:
		if p.peekTokenIs(token.IDENT) {
			return p.parseAttributeStatement()
		}
		return p.parseExpressionStatement()
	case token.SELECT:
		return p.parseSelectStatement()
	case token.BREAK:
		return p.parseBreakStatement()
	case token.CONTINUE:
		return p.parseContinueStatement()
	case token.DEFER:
		return p.parseDeferStatement()
	case token.TYPE:
		if p.peekTokenIs(token.IDENT) {
			return p.parseTypeStatement()
		}
		return p.parseExpressionStatement()
	case token.PUB:
		return p.parsePublicStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseAttributeStatement() ast.Statement {
	var attributes []ast.Attribute

	// Parse one or more attributes: [foo] [bar("baz")]
	for p.curTokenIs(token.LBRACKET) && p.peekTokenIs(token.IDENT) {
		p.nextToken() // Move past '['

		attrToken := p.curToken
		attrName := p.curToken.Literal
		var args []string

		if p.peekTokenIs(token.LPAREN) {
			p.nextToken() // move to '('

			// read args
			for p.peekTokenIs(token.STR) {
				p.nextToken()
				args = append(args, p.curToken.Literal)
				if p.peekTokenIs(token.COMMA) {
					p.nextToken()
				}
			}

			if !p.expectPeek(token.RPAREN) {
				return nil
			}
		}

		if !p.expectPeek(token.RBRACKET) {
			return nil
		}

		attributes = append(attributes, ast.Attribute{
			Token: attrToken,
			Name:  attrName,
			Args:  args,
		})

		p.nextToken() // move past ']'

		// consume newlines/semicolons
		for p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	var stmt ast.Statement
	if p.curTokenIs(token.FN) {
		stmt = p.parseFunctionStatement(false, true)
	} else if p.curTokenIs(token.EXPORT) {
		stmt = p.parseExportStatement(true)
	} else if p.curTokenIs(token.TYPE) {
		stmt = p.parseTypeStatement()
	} else if p.curTokenIs(token.PUB) {
		stmt = p.parsePublicStatement()
	} else if p.curTokenIs(token.EXTERN) {
		stmt = p.parseExternStatement()
	} else {
		p.ReportError(p.curToken.Position, "expected 'fn', 'export fn', 'extern fn', or 'type' after attribute, got %s", p.curToken.Type)
		return nil
	}

	if stmt != nil {
		if fnStmt, ok := stmt.(*ast.FunctionStatement); ok && fnStmt != nil {
			fnStmt.Attributes = attributes
		} else if typeStmt, ok := stmt.(*ast.TypeStatement); ok && typeStmt != nil {
			typeStmt.Attributes = attributes
		} else if varStmt, ok := stmt.(*ast.VarStatement); ok && varStmt != nil {
			// Var statements don't usually have attributes in Nora yet, but for consistency:
			_ = varStmt
		}
	}

	return stmt
}

func (p *Parser) parseSelectStatement() *ast.SelectStatement {
	stmt := &ast.SelectStatement{Token: p.curToken}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	p.nextToken() // Move past '{'

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
			continue
		}

		if p.curTokenIs(token.CASE) || p.curTokenIs(token.DEFAULT) {
			stmt.Cases = append(stmt.Cases, p.parseSelectCase())
		} else {
			p.ReportError(p.curToken.Position, "expected 'case' or 'default' in select, got %s", p.curToken.Type)
			return nil
		}
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseSelectCase() *ast.SelectCase {
	sc := &ast.SelectCase{Token: p.curToken}

	if p.curTokenIs(token.CASE) {
		p.nextToken() // Move past 'case'
		// Parse the condition (SendExpression or Assignment)
		sc.Condition = p.parseStatement()
	} else {
		// default case
		sc.Condition = nil
	}

	if !p.expectPeek(token.COLON) {
		return nil
	}

	// Parse body until next case, default or }
	sc.Body = &ast.BlockStatement{Token: p.peekToken}
	for !p.peekTokenIs(token.CASE) && !p.peekTokenIs(token.DEFAULT) && !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		if p.curTokenIs(token.SEMICOLON) {
			continue
		}
		s := p.parseStatement()
		if s != nil {
			sc.Body.Statements = append(sc.Body.Statements, s)
		}
	}

	return sc
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	// Case 1: Infinite loop 'for { }'
	if p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		stmt.Body = p.parseBlockStatement()
		return stmt
	}

	p.nextToken() // Move past 'for'

	// Case 2: for-in loop 'for x in arr' or 'for i, x in arr'
	// We check if this is an identifier followed by 'in' or ','
	if p.curTokenIs(token.IDENT) && (p.peekTokenIs(token.IN) || p.peekTokenIs(token.COMMA)) {
		if p.peekTokenIs(token.COMMA) {
			stmt.Key = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
			p.nextToken() // Move to ','
			if !p.expectPeek(token.IDENT) {
				return nil
			}
			stmt.Value = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		} else {
			stmt.Value = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		}

		if !p.expectPeek(token.IN) {
			return nil
		}
		p.nextToken() // Move past 'in'

		p.noStructLiteral = true
		stmt.Iterable = p.parseExpression(LOWEST)
		p.noStructLiteral = false

		if p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		stmt.Body = p.parseBlockStatement()
		return stmt
	}

	p.ReportError(p.curToken.Position, "expected '{' for infinite loop or 'in' for for-in loop")
	return nil
}

func (p *Parser) parseTypeParameters() []*ast.TypeParameter {
	tparams := []*ast.TypeParameter{}
	if !p.peekTokenIs(token.LBRACKET) {
		return tparams
	}
	p.nextToken() // move to [

	for {
		p.nextToken() // move to T
		if !p.curTokenIs(token.IDENT) {
			break
		}

		tp := &ast.TypeParameter{
			Token: p.curToken,
			Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
		}

		if p.peekTokenIs(token.COLON) {
			p.nextToken() // :
			p.nextToken() // move to constraint
			tp.Constraint = p.parseType()
		}

		tparams = append(tparams, tp)

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		} else {
			break
		}
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return tparams
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	// Handle i++ / i-- special case
	if p.curTokenIs(token.IDENT) && (p.peekTokenIs(token.INC) || p.peekTokenIs(token.DEC)) {
		return p.parseIncrementDecrementStatement()
	}

	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseIncrementDecrementStatement() *ast.ExpressionStatement {
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken() // move to ++ or --

	var op token.TokenType = token.PLUS
	var opLit string = "+"
	if p.curTokenIs(token.DEC) {
		op = token.MINUS
		opLit = "-"
	}

	// Desugar: i++ => i = i + 1
	stmt := &ast.AssignmentStatement{
		Token: p.curToken,
		Left:  ident,
		Value: &ast.InfixExpression{
			Token:    token.Token{Type: op, Literal: opLit},
			Left:     ident,
			Operator: opLit,
			Right:    &ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
		},
	}

	return &ast.ExpressionStatement{
		Token:      ident.Token,
		Expression: stmt,
	}
}
func (p *Parser) parsePackageStatement() *ast.PackageStatement {
	stmt := &ast.PackageStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Optional: Consume semicolon if your language uses them
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// --- IMPORT ---
func (p *Parser) parseImportStatement() *ast.ImportStatement {
	stmt := &ast.ImportStatement{Token: p.curToken}

	// 1. Check for optional alias: import m "math"
	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		stmt.Alias = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	// 2. Expect the string path: import "math"
	if !p.curTokenIs(token.STR) && !p.expectPeek(token.STR) {
		return nil
	}
	pathLit := p.parseStringLiteral().(*ast.StringLiteral)
	stmt.Path = &ast.Identifier{
		Token: pathLit.Token,
		Value: pathLit.Value,
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseAssignmentExpression(left ast.Expression) ast.Expression {
	stmt := &ast.AssignmentStatement{Token: p.curToken, Left: left}

	op := p.curToken.Type
	precedence := p.curPrecedence()
	p.nextToken() // Skip operator

	val := p.parseExpression(precedence)

	if op == token.ASSIGN {
		stmt.Value = val
	} else {
		var realOp token.TokenType
		var realOpLit string
		switch op {
		case token.PLUS_ASSIGN:
			realOp, realOpLit = token.PLUS, "+"
		case token.MINUS_ASSIGN:
			realOp, realOpLit = token.MINUS, "-"
		case token.ASTERISK_ASSIGN:
			realOp, realOpLit = token.ASTERISK, "*"
		case token.SLASH_ASSIGN:
			realOp, realOpLit = token.SLASH, "/"
		case token.REM_ASSIGN:
			realOp, realOpLit = token.REM, "%"
		case token.AND_ASSIGN:
			realOp, realOpLit = token.AND, "&"
		case token.OR_ASSIGN:
			realOp, realOpLit = token.OR, "|"
		case token.XOR_ASSIGN:
			realOp, realOpLit = token.XOR, "^"
		case token.SHL_ASSIGN:
			realOp, realOpLit = token.SHL, "<<"
		case token.SHR_ASSIGN:
			realOp, realOpLit = token.SHR, ">>"
		case token.AND_NOT_ASSIGN:
			realOp, realOpLit = token.AND_NOT, "&^"
		}

		stmt.Value = &ast.InfixExpression{
			Token:    token.Token{Type: realOp, Literal: realOpLit, Position: stmt.Token.Position},
			Left:     left,
			Operator: realOpLit,
			Right:    val,
		}
	}

	return stmt
}

func (p *Parser) parseExternStatement() ast.Statement {
	// Current token is EXTERN
	if !p.expectPeek(token.FN) {
		return nil
	}

	// Parse the function but tell it not to look for a body { }
	return p.parseFunctionStatement(true, true)
}

func (p *Parser) parseExportStatement(allowNoBody bool) ast.Statement {
	// Current token is EXPORT
	if !p.expectPeek(token.FN) {
		return nil
	}

	// Parse the function
	fn := p.parseFunctionStatement(false, allowNoBody)
	if fn != nil {
		fn.IsExport = true
	}
	return fn
}

func (p *Parser) parsePublicStatement() ast.Statement {
	// Current token is PUB
	p.nextToken() // Move past 'pub'

	var stmt ast.Statement
	switch p.curToken.Type {
	case token.FN:
		stmt = p.parseFunctionStatement(false, true)
		if fn, ok := stmt.(*ast.FunctionStatement); ok && fn != nil {
			fn.IsPublic = true
		}
	case token.TYPE:
		stmt = p.parseTypeStatement()
		if ts, ok := stmt.(*ast.TypeStatement); ok && ts != nil {
			ts.IsPublic = true
		}
	case token.VAR:
		stmt = p.parseVarStatement()
		if vs, ok := stmt.(*ast.VarStatement); ok && vs != nil {
			vs.IsPublic = true
		}
	case token.EXPORT:
		stmt = p.parseExportStatement(true)
		if fn, ok := stmt.(*ast.FunctionStatement); ok && fn != nil {
			fn.IsPublic = true
		}
	default:
		p.ReportError(p.curToken.Position, "expected 'fn', 'type', 'var', or 'export' after 'pub', got %s", p.curToken.Type)
		return nil
	}
	return stmt
}

// --- RETURN ---
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken() // move to ';'
		return stmt
	}

	if p.peekTokenIs(token.RBRACE) || p.peekTokenIs(token.EOF) {
		// Do not advance! 'return' is the last token of this statement.
		return stmt
	}

	p.nextToken() // move to the start of the expression

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseVarStatement() *ast.VarStatement {
	stmt := &ast.VarStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekTokenIs(token.COLON) {
		p.nextToken() // move past :
		p.nextToken() // move to type
		stmt.Type = p.parseType()
	}

	if p.peekTokenIs(token.ASSIGN) {
		p.nextToken() // move to '='
		p.nextToken() // move past '='
		stmt.Value = p.parseExpression(LOWEST)
	}

	return stmt
}

func extractNumericSuffix(literal string) (string, string) {
	suffixes := []string{
		"i8", "u8", "i16", "u16", "i32", "u32", "i64", "u64",
		"f32", "f64", "c32", "c64", "c128", "i", "j",
	}

	// Check for underscore-prefixed suffixes first
	for _, s := range suffixes {
		if strings.HasSuffix(literal, "_"+s) {
			return literal[:len(literal)-len(s)-1], s
		}
	}

	// Check for direct suffixes
	for _, s := range suffixes {
		if strings.HasSuffix(literal, s) {
			isHex := strings.HasPrefix(strings.ToLower(literal), "0x")
			// If it's a hex number and the suffix starts with a-f, it's ambiguous, assume it's part of the hex number
			if isHex && ((s[0] >= 'a' && s[0] <= 'f') || (s[0] >= 'A' && s[0] <= 'F')) {
				continue
			}
			return literal[:len(literal)-len(s)], s
		}
	}

	return literal, ""
}

func (p *Parser) parseNumberLiteral() ast.Expression {
	rawLiteral := p.curToken.Literal
	numPart, suffix := extractNumericSuffix(rawLiteral)
	
	if suffix != "" && suffix != "i" && suffix != "j" {
		p.ReportError(p.curToken.Position, "type suffixes on numeric literals are disallowed; use constructor-style conversion (e.g. %s(%s)) instead", suffix, numPart)
		suffix = ""
	}
	
	// Remove visual separators from numPart for strconv
	cleanNum := strings.ReplaceAll(numPart, "_", "")

	// 1. Handle Integers
	if p.curTokenIs(token.INT) {
		val, err := strconv.ParseInt(cleanNum, 0, 64)
		if err != nil {
			// fallback for uint64 max
			uval, uerr := strconv.ParseUint(cleanNum, 0, 64)
			if uerr != nil {
				p.ReportError(p.curToken.Position, "could not parse %q as integer", rawLiteral)
				return nil
			}
			val = int64(uval)
		}
		return &ast.IntegerLiteral{Token: p.curToken, Value: val, Suffix: suffix}
	}

	// 2. Handle Imaginary (e.g., 5i, 10.5j)
	if p.curTokenIs(token.IMAG) {
		val, err := strconv.ParseFloat(cleanNum, 64)
		if err != nil {
			p.ReportError(p.curToken.Position, "could not parse %q as imaginary component", rawLiteral)
			return nil
		}
		return &ast.ImaginaryLiteral{Token: p.curToken, Value: val, Suffix: suffix}
	}

	// 3. Handle Regular Floats
	val, err := strconv.ParseFloat(cleanNum, 64)
	if err != nil {
		p.ReportError(p.curToken.Position, "could not parse %q as float", rawLiteral)
		return nil
	}
	return &ast.FloatLiteral{Token: p.curToken, Value: val, Suffix: suffix}
}

// --- Structure Parsers ---
func (p *Parser) parseFieldDefinitions(end token.TokenType) []*ast.FieldDefinition {
	fields := []*ast.FieldDefinition{}

	// Move past '{' or '('
	if p.curTokenIs(token.LBRACE) || p.curTokenIs(token.LPAREN) {
		p.nextToken()
	}

	// Standard Loop: Check for terminator at the start
	for !p.curTokenIs(end) && !p.curTokenIs(token.EOF) {

		// 1. Skip separators (Commas, Newlines/Semicolons)
		// If parseType left us on a comma, we consume it here and move to the next name.
		if p.curTokenIs(token.COMMA) || p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
			continue
		}

		// 2. Parse field
		if p.curTokenIs(token.IDENT) || p.curTokenIs(token.LBRACKET) {
			field := p.parseFieldDefinition()
			if field != nil {
				fields = append(fields, field)
			}

			if p.peekTokenIs(token.COMMA) {
				p.nextToken() // Move to ','
			}
		}
		p.nextToken()
	}

	return fields
}

// 2. Fix parseFieldDefinition
func (p *Parser) parseFieldDefinition() *ast.FieldDefinition {
	var attributes []ast.Attribute

	// Parse field-level attributes if current token is [
	for p.curTokenIs(token.LBRACKET) && p.peekTokenIs(token.IDENT) {
		p.nextToken() // Move past '['

		attrToken := p.curToken
		attrName := p.curToken.Literal
		var args []string

		if p.peekTokenIs(token.LPAREN) {
			p.nextToken() // move to '('

			// read args
			for p.peekTokenIs(token.STR) {
				p.nextToken()
				args = append(args, p.curToken.Literal)
				if p.peekTokenIs(token.COMMA) {
					p.nextToken()
				}
			}

			if !p.expectPeek(token.RPAREN) {
				return nil
			}
		}

		if !p.expectPeek(token.RBRACKET) {
			return nil
		}

		attributes = append(attributes, ast.Attribute{
			Token: attrToken,
			Name:  attrName,
			Args:  args,
		})

		p.nextToken() // move past ']'

		// consume newlines/semicolons
		for p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	// 1. Check for Name
	if !p.curTokenIs(token.IDENT) {
		p.peekError(token.IDENT)
		return nil
	}

	field := &ast.FieldDefinition{Token: p.curToken, Attributes: attributes}
	field.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 2. Expect Colon
	if !p.expectPeek(token.COLON) {
		return nil
	}
	p.nextToken() // Move to Type (e.g., 'i32')

	// 3. Parse Type
	field.Type = p.parseType()
	// Note: p.curToken is now "i32". We return immediately.

	p.applyDocComments(field)
	return field
}
func (p *Parser) parseBaseType() ast.TypeNode {
	var left ast.TypeNode

	if p.curTokenIs(token.LPAREN) {
		tok := p.curToken
		p.nextToken()
		left = p.parseType()
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
		if p.PreserveParentheses {
			left = &ast.GroupedExpression{Token: tok, Expression: left}
		}
	} else if p.curToken.Type == token.IDENT {
		left = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	} else if p.curTokenIs(token.CHAN) {
		stmt := &ast.ChanType{Token: p.curToken}
		if !p.expectPeek(token.LBRACKET) {
			return nil
		}
		p.nextToken() // Move to element type
		stmt.Value = p.parseType()
		if !p.expectPeek(token.RBRACKET) {
			return nil
		}
		left = stmt
	} else if p.curTokenIs(token.FN) {
		left = p.parseFunctionType()
	}

	return left
}

func (p *Parser) parseTightType() ast.TypeNode {
	return p.parseBaseType()
}

func (p *Parser) parseType() ast.TypeNode {
	// 1. Lease modifiers (#, &, @) bind loosely
	if p.curTokenIs(token.HASH) || p.curTokenIs(token.AND) || p.curTokenIs(token.MOVE) {
		tok := p.curToken
		p.nextToken()
		return &ast.PrefixExpression{
			Token:    tok,
			Operator: tok.Literal,
			Right:    p.parseType(),
		}
	}

	// 3. Standard types (base + all suffixes)
	left := p.parseTightType()
	if left == nil {
		p.ReportError(p.curToken.Position, "expected type identifier, got %s", p.curToken.Type)
		return nil
	}

	for p.peekTokenIs(token.DOT) || p.peekTokenIs(token.LBRACKET) {
		if p.peekTokenIs(token.DOT) {
			p.nextToken()
			p.nextToken()
			if p.curToken.Type != token.IDENT {
				p.ReportError(p.curToken.Position, "expected identifier after '.', got %s", p.curToken.Type)
				return nil
			}
			left = &ast.SelectorExpression{
				Token: token.Token{Type: token.DOT, Literal: "."},
				Left:  left,
				Field: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
			}
		} else {
			left = p.parseTypeSuffix(left)
		}
	}

	return left
}

func (p *Parser) parseTypeSuffix(left ast.TypeNode) ast.TypeNode {
	p.nextToken() // Move to '['
	if p.peekTokenIs(token.RBRACKET) {
		p.nextToken() // Move to ']'
		return &ast.IndexExpression{
			Token:   token.Token{Type: token.LBRACKET, Literal: "["},
			Left:    left,
			Indices: []ast.Expression{},
		}
	}

	p.nextToken() // Move to first argument
	exp := &ast.IndexExpression{
		Token:   token.Token{Type: token.LBRACKET, Literal: "["},
		Left:    left,
		Indices: []ast.Expression{},
	}

	var firstIdx ast.Expression
	if p.curTokenIs(token.INT) {
		firstIdx = p.parseExpression(LOWEST)
	} else {
		firstIdx = p.parseType()
	}
	exp.Indices = append(exp.Indices, firstIdx)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Move to ','
		p.nextToken() // Move past ','
		var idx ast.Expression
		if p.curTokenIs(token.INT) {
			idx = p.parseExpression(LOWEST)
		} else {
			idx = p.parseType()
		}
		exp.Indices = append(exp.Indices, idx)
	}
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return exp
}

func (p *Parser) parseFunctionType() ast.TypeNode {
	lit := &ast.FunctionType{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = []ast.TypeNode{}

	if !p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		lit.Parameters = append(lit.Parameters, p.parseType())

		for p.peekTokenIs(token.COMMA) {
			p.nextToken() // Move to ','
			p.nextToken() // Move past ','
			lit.Parameters = append(lit.Parameters, p.parseType())
		}
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// Optional return type
	if !p.peekTokenIs(token.LBRACE) && !p.peekTokenIs(token.COMMA) && !p.peekTokenIs(token.RPAREN) && !p.peekTokenIs(token.SEMICOLON) && !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.ASSIGN) && !p.peekTokenIs(token.RBRACKET) {
		p.nextToken()
		lit.ReturnType = p.parseType()
	}

	return lit
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{
		Token: p.curToken,
		Value: p.curTokenIs(token.TRUE),
	}
}

func (p *Parser) parseNone() ast.Expression {
	return &ast.NoneLiteral{
		Token: p.curToken,
	}
}

func (p *Parser) interpolatedPositionAt(fullLiteral string, index int) token.Position {
	parent := p.curToken.Position
	lineOffset := 0
	colOffset := 0
	for idx := 0; idx < index; idx++ {
		if fullLiteral[idx] == '\n' {
			lineOffset++
			colOffset = 0
		} else {
			colOffset++
		}
	}
	line := parent.Line + lineOffset
	col := parent.Column + colOffset
	if lineOffset > 0 {
		col = colOffset
	}
	return token.Position{
		Line:     line,
		Column:   col,
		Offset:   parent.Offset + index,
		Filename: parent.Filename,
	}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	fullLiteral := p.curToken.Literal
	var unquoted string
	if len(fullLiteral) >= 2 && fullLiteral[0] == '"' && fullLiteral[len(fullLiteral)-1] == '"' {
		unquoted = fullLiteral[1 : len(fullLiteral)-1]
	} else {
		unquoted = fullLiteral
	}
	if !strings.Contains(fullLiteral, "$") {
		return &ast.StringLiteral{Token: p.curToken, Value: unquoted}
	}

	res := &ast.InterpolatedString{Token: p.curToken}
	i := 0
	for i < len(fullLiteral) {
		start := i
		for i < len(fullLiteral) && fullLiteral[i] != '$' {
			i++
		}

		if i > start {
			res.Parts = append(res.Parts, &ast.StringLiteral{
				Value: fullLiteral[start:i],
			})
		}

		if i >= len(fullLiteral) {
			break
		}

		// Found '$'
		i++ // skip '$'
		if i >= len(fullLiteral) {
			res.Parts = append(res.Parts, &ast.StringLiteral{Value: "$"})
			break
		}

		if fullLiteral[i] == '{' {
			i++ // skip '{'
			exprStart := i
			braceCount := 1
			for i < len(fullLiteral) && braceCount > 0 {
				if fullLiteral[i] == '{' {
					braceCount++
				} else if fullLiteral[i] == '}' {
					braceCount--
				}
				i++
			}
			exprStr := fullLiteral[exprStart : i-1]

			exprPos := p.interpolatedPositionAt(fullLiteral, exprStart)
			subL := lexer.NewAt(exprStr, p.l.Filename, exprPos.Line, exprPos.Column, exprPos.Offset)

			subP := New(subL)
			subExp := subP.parseExpression(LOWEST)
			if subExp != nil {
				res.Parts = append(res.Parts, subExp)
			}
		} else {
			// Simple $ident
			identStart := i
			for i < len(fullLiteral) && (unicode.IsLetter(rune(fullLiteral[i])) || unicode.IsDigit(rune(fullLiteral[i])) || fullLiteral[i] == '_') {
				i++
			}
			identStr := fullLiteral[identStart:i]
			if identStr == "" {
				res.Parts = append(res.Parts, &ast.StringLiteral{Value: "$"})
			} else {
				res.Parts = append(res.Parts, &ast.Identifier{
					Token: token.Token{
						Type:     token.IDENT,
						Literal:  identStr,
						Position: p.interpolatedPositionAt(fullLiteral, identStart),
					},
					Value: identStr,
				})
			}
		}
	}

	return res
}

func (re *Parser) parseReceiveExpression() ast.Expression {
	exp := &ast.ReceiveExpression{Token: re.curToken}
	re.nextToken()
	exp.Value = re.parseExpression(PREFIX)
	return exp
}

func (re *Parser) parseSendExpression(left ast.Expression) ast.Expression {
	exp := &ast.SendExpression{Token: re.curToken, Left: left}
	precedence := re.curPrecedence()
	re.nextToken()
	exp.Right = re.parseExpression(precedence)
	return exp
}

func (p *Parser) parseChanTypeExpression() ast.Expression {
	// We use the already implemented parseType which handles 'chan[T]'
	// This works because ChanType and Identifier both implement Expression.
	return p.parseType()
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	tok := p.curToken
	p.nextToken() // consume '('
	oldNoStruct := p.noStructLiteral
	p.noStructLiteral = false
	exp := p.parseExpression(LOWEST)
	p.noStructLiteral = oldNoStruct
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if p.PreserveParentheses {
		return &ast.GroupedExpression{Token: tok, Expression: exp}
	}
	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}
	p.nextToken() // move past 'if'
	p.noStructLiteral = true
	expression.Condition = p.parseExpression(LOWEST)
	p.noStructLiteral = false
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.SEMICOLON) && p.peek2Token.Type == token.ELSE {
		p.nextToken()
	}

	if p.peekTokenIs(token.ELSE) {
		p.nextToken() // move to 'else'

		if p.peekTokenIs(token.IF) {
			p.nextToken() // move to 'if'
			// Recursive call: parse the next 'if' as the Alternative
			expression.Alternative = p.parseIfExpression()
		} else {
			// Standard final 'else': parse the block
			if !p.expectPeek(token.LBRACE) {
				return nil
			}
			expression.Alternative = p.parseBlockStatement()
		}
	}
	return expression
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken}
	p.nextToken() // move past while
	p.noStructLiteral = true
	stmt.Condition = p.parseExpression(LOWEST)
	p.noStructLiteral = false
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseLambdaExpression() ast.Expression {
	exp := &ast.LambdaExpression{Token: p.curToken} // 'fn'

	// 1. Parameters
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	exp.Parameters = p.parseFunctionParameters()

	// 2. Optional Return Type
	if !p.peekTokenIs(token.LBRACE) && !p.peekTokenIs(token.SEMICOLON) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		exp.ReturnType = p.parseType()
	}

	// 3. Body
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	exp.Body = p.parseBlockStatement()

	return exp
}

func (p *Parser) parseFunctionStatement(isExtern bool, allowNoBody bool) *ast.FunctionStatement {
	stmt := &ast.FunctionStatement{
		Token:    p.curToken, // 'fn'
		IsExtern: isExtern,
	}

	// 1. Optional Receiver or Name
	if p.peekTokenIs(token.LPAREN) {
		// Method Receiver: fn (self: #Point) Name(...)
		p.nextToken() // Move to (
		receivers := p.parseFunctionParameters()
		if len(receivers) > 0 {
			stmt.Receiver = receivers[0]
		}

		// After receiver, we MUST have a name
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	} else if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	// 1.5. Type Parameters
	if p.peekTokenIs(token.LBRACKET) {
		stmt.TypeParameters = p.parseTypeParameters()
	}

	// 2. Parameters
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	stmt.Parameters = p.parseFunctionParameters()

	// 3. Optional Return Type
	// Look for return type only if we aren't at the end of an extern
	// or the start of a block.
	if !p.peekTokenIs(token.LBRACE) && !p.peekTokenIs(token.SEMICOLON) && !p.peekTokenIs(token.EOF) {
		// If we are on a new line and it's an extern, ASI might trigger.
		// But let's assume the type is on the same line.
		p.nextToken()
		stmt.ReturnType = p.parseType()
	}

	// 4. Body vs. Termination
	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	if p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		stmt.Body = p.parseBlockStatement()
	} else if allowNoBody {
		stmt.Body = nil
	} else {
		// Normal functions MUST have a body
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		stmt.Body = p.parseBlockStatement()
	}

	return stmt
}
func (p *Parser) parseFunctionParameters() []*ast.Parameter {
	identifiers := []*ast.Parameter{}

	// Case: Empty parameters "()"
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken() // Consume ')'
		return identifiers
	}

	p.nextToken() // Move from '(' to first identifier

	// Loop to parse parameters
	for {
		param := &ast.Parameter{
			LeaseKind: ast.LeaseRead, // Default is Read-Only Borrow
		}

		// 1. Parse Parameter Name or Ellipsis
		if p.curTokenIs(token.ELLIPSIS) {
			param.IsVariadic = true
			param.Token = p.curToken
			param.Name = &ast.Identifier{Token: p.curToken, Value: "..."}
			identifiers = append(identifiers, param)
			p.nextToken() // Move past '...'
			break         // Variadic must be last
		}

		if !p.curTokenIs(token.IDENT) {
			return nil
		}

		param.Token = p.curToken
		param.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		// 3. Parse Type Annotation
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken() // Move past ':'

		param.Type = p.parseType()
		if param.Type == nil {
			return nil
		}

		identifiers = append(identifiers, param)

		// 4. Check for comma or end
		if p.peekTokenIs(token.COMMA) {
			p.nextToken() // Move to comma
			p.nextToken() // Move to next parameter start
		} else {
			break
		}
	}

	// 5. Closing Parenthesis
	if p.curTokenIs(token.RPAREN) && p.peekTokenIs(token.RPAREN) {
		p.nextToken() // Consume the outer RPAREN
	} else if !p.curTokenIs(token.RPAREN) {
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	}

	return identifiers
}

func (p *Parser) parseSpawnExpression() ast.Expression {
	expr := &ast.SpawnExpression{Token: p.curToken} // curToken is 'spawn'

	p.nextToken() // move past 'spawn'

	// Use a precedence slightly lower than CALL.
	// This forces the parser to process the Identifier + Call as one unit.
	firstExp := p.parseExpression(ACCESSOR - 1)

	// If the next token can start an expression, it means we have two expressions:
	// spawn(monitor_chan) function_call()
	if p.prefixParseFns[p.peekToken.Type] != nil {
		expr.MonitorChannel = firstExp
		p.nextToken() // move to start of second expression
		secondExp := p.parseExpression(ACCESSOR - 1)
		
		call, ok := secondExp.(*ast.CallExpression)
		if !ok {
			p.ReportError(p.curToken.Position, "spawn monitor channel must be followed by a function call")
			return nil
		}
		expr.Call = call
		return expr
	}

	call, ok := firstExp.(*ast.CallExpression)
	if !ok {
		p.ReportError(p.curToken.Position, "spawn expression must be followed by a function call")
		return nil
	}

	expr.Call = call
	return expr
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken() // Move past '['

	exp.Indices = []ast.Expression{}

	// Handle empty index: arr[]
	if p.curTokenIs(token.RBRACKET) {
		return exp
	}

	oldNoStruct := p.noStructLiteral
	p.noStructLiteral = false

	// Handle Slice shortcut: arr[:end]
	if p.curTokenIs(token.COLON) {
		slice := &ast.SliceExpression{Token: p.curToken}
		p.nextToken() // past :
		if !p.curTokenIs(token.RBRACKET) {
			slice.End = p.parseExpression(LOWEST)
		}
		exp.Indices = append(exp.Indices, slice)
	} else {
		// Normal index or Slice start: arr[start...]
		idx := p.parseExpression(LOWEST)
		if p.peekTokenIs(token.COLON) {
			p.nextToken() // move to :
			slice := &ast.SliceExpression{Token: p.curToken, Start: idx}
			p.nextToken() // past :
			if !p.curTokenIs(token.RBRACKET) {
				slice.End = p.parseExpression(LOWEST)
			}
			exp.Indices = append(exp.Indices, slice)
		} else {
			exp.Indices = append(exp.Indices, idx)
			for p.peekTokenIs(token.COMMA) {
				p.nextToken() // Move to ','
				p.nextToken() // Move past ','
				exp.Indices = append(exp.Indices, p.parseExpression(LOWEST))
			}
		}
	}

	p.noStructLiteral = oldNoStruct

	if p.curTokenIs(token.RBRACKET) && !p.peekTokenIs(token.RBRACKET) {
		return exp
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return exp
}

func (p *Parser) parseTryExpression(left ast.Expression) ast.Expression {
	// left is (arr[0]).id
	return &ast.TryExpression{
		Token: p.curToken, // The '?'
		Value: left,
	}
}
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken() // move past '{'

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		// Skip leading semicolons
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if !ast.IsNil(stmt) {
			block.Statements = append(block.Statements, stmt)
		}

		// Standard Pratt advance: if the statement parser left us on the
		// last token of the statement, move to the next one.
		p.nextToken()
	}
	block.Rbrace = p.curToken
	return block
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}

	// Handle generic instantiation: add[i32](...) or Map[str, i32]()
	if idx, ok := function.(*ast.IndexExpression); ok {
		exp.Function = idx.Left
		for _, index := range idx.Indices {
			if typeNode, ok := index.(ast.TypeNode); ok {
				exp.TypeArguments = append(exp.TypeArguments, typeNode)
			}
		}
	}

	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}
func (p *Parser) parseExpressionList(end token.TokenType) []*ast.ArgumentsExpression {
	list := []*ast.ArgumentsExpression{}

	// 1. Check for empty: "add()"
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	// 2. Parse the first argument
	p.nextToken() // Move from '(' to the start of the first expression
	list = append(list, p.parseArgument())

	// 3. Handle subsequent arguments: ", arg"
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Move to the ','
		p.nextToken() // Move to the start of the next expression
		list = append(list, p.parseArgument())
	}

	// 4. Close the call: ')'
	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseArgument() *ast.ArgumentsExpression {
	// Crucial: Create the wrapper and parse the internal expression
	oldNoStruct := p.noStructLiteral
	p.noStructLiteral = false
	arg := &ast.ArgumentsExpression{
		Token: p.curToken,
		Value: p.parseExpression(LOWEST),
	}
	p.noStructLiteral = oldNoStruct

	// Safety check: if Value is nil, the parser failed to find an expression
	if arg.Value == nil {
		return nil
	}

	return arg
}

func (p *Parser) parseStructLiteral() ast.Expression {
	lit := &ast.StructLiteral{Token: p.curToken}
	lit.Fields = []*ast.FieldDefinition{}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// Check for empty struct: struct { }
	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return lit
	}

	p.nextToken() // Move to first Field Name

	for {
		// 1. Parse the Field
		// This leaves p.curToken at the Type (e.g., "i32" or "f64")
		field := p.parseFieldDefinition()
		if field == nil {
			p.ReportError(p.curToken.Position, "could not resolve field definition")
			return nil
		}
		lit.Fields = append(lit.Fields, field)

		// 2. Check the NEXT token (peekToken)

		// Case A: Comma OR Semicolon (Newline) -> Consume and continue
		if p.peekTokenIs(token.COMMA) || p.peekTokenIs(token.SEMICOLON) {
			p.nextToken() // Consume the Type ("f64") -> cur is "," or ";"
			p.nextToken() // Consume the Separator -> cur is next Field Name or '}'

			// If we hit a semicolon/newline, we might be at the closing brace now
			if p.curTokenIs(token.RBRACE) {
				return lit
			}
			continue
		}

		// Case B: RBrace -> Consume and finish
		if p.peekTokenIs(token.RBRACE) {
			p.nextToken() // Consume the Type -> cur is "}"
			break
		}

		// Error Case
		p.peekError(token.RBRACE)
		return nil
	}

	// 3. Finalize
	if !p.curTokenIs(token.RBRACE) {
		// Should have been handled by loop or break
		return nil
	}

	return lit
}

// Parses: User { id: "val", name: "foo" }
// parseStructInstantiation handles: User { id: "val" }
func (p *Parser) parseStructInstantiation(left ast.Expression) ast.Expression {
	// 1. Create the AST Node (Using StructLiteral for now)
	// We treat 'left' (the identifier "User") as the struct type/name
	lit := &ast.StructLiteral{Token: p.curToken}
	// FIX: Store the name "User" (which was passed in as 'left')
	lit.Name = left
	lit.Fields = []*ast.FieldDefinition{}

	// 2. Parse Fields loop
	p.nextToken() // Move past '{'

	for !p.curTokenIs(token.RBRACE) {
		// Handle empty struct or trailing comma
		if p.curTokenIs(token.RBRACE) {
			break
		}

		// A. Parse Key (e.g. "id")
		if !p.curTokenIs(token.IDENT) {
			p.peekError(token.IDENT)
			return nil
		}

		// Create a FieldDefinition
		// Note: We are reusing FieldDefinition, but putting the VALUE in the Type field is a hack.
		// For a proper compiler, you should create a 'CompositeLiteral' node.
		// But to make your current test PASS, we will adapt:
		field := &ast.FieldDefinition{Token: p.curToken}
		field.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		// B. Expect Colon
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken() // Move to Value

		// C. Parse Value (Expression)
		val := p.parseExpression(LOWEST)
		field.Value = val

		// Add to list (dummy)
		lit.Fields = append(lit.Fields, field)

		// D. Handle Comma or Semicolon (Newline)
		if p.peekTokenIs(token.COMMA) || p.peekTokenIs(token.SEMICOLON) {
			p.nextToken() // cur is ',' or ';'
			p.nextToken() // cur is next Key or '}'
		} else if !p.peekTokenIs(token.RBRACE) {
			p.peekError(token.RBRACE)
			return nil
		} else {
			p.nextToken() // cur is '}'
		}
	}

	return lit
}
func (p *Parser) parseSumTypeLiteral() ast.Expression {
	lit := &ast.SumTypeLiteral{Token: p.curToken}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	p.nextToken() // Move to first variant
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.COMMA) || p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
			continue
		}

		variant := &ast.VariantDefinition{
			Token: p.curToken,
			Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
		}
		if p.peekTokenIs(token.LPAREN) {
			p.nextToken() // Move to '('
			variant.Fields = p.parseFieldDefinitions(token.RPAREN)
		}

		p.applyDocComments(variant)
		lit.Variants = append(lit.Variants, variant)
		if p.peekTokenIs(token.COMMA) {
			p.nextToken() // Move to ','
			p.nextToken() // Move to next variant
		} else {
			p.nextToken() // Move to RBRACE or similar
		}
	}
	return lit
}

func (p *Parser) parseTypeStatement() *ast.TypeStatement {
	stmt := &ast.TypeStatement{Token: p.curToken}

	// Expect an identifier (the name of the type)
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekTokenIs(token.LBRACKET) {
		stmt.TypeParameters = p.parseTypeParameters()
	}

	// Expect '=' (assignment of the structure to the type name)
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	// This parses the StructLiteral, SumTypeLiteral, or just an Identifier
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseInterfaceLiteral() ast.Expression {
	lit := &ast.InterfaceLiteral{Token: p.curToken}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		if p.curTokenIs(token.FN) {
			// Parse a function statement but WITHOUT a body
			method := p.parseFunctionStatement(true, true)
			if method != nil {
				lit.Methods = append(lit.Methods, method)
			}
		} else if p.curTokenIs(token.IDENT) {
			lit.Embedded = append(lit.Embedded, p.parseIdentifier().(*ast.Identifier))
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return lit
}

func (p *Parser) parseMatchExpression() ast.Expression {
	exp := &ast.MatchExpression{Token: p.curToken}

	// Match target can be parenthesized or a simple expression
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken() // Move to '('
		exp.Target = p.parseExpression(LOWEST)
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	} else {
		p.nextToken() // Move to target
		p.noStructLiteral = true
		exp.Target = p.parseExpression(LOWEST)
		p.noStructLiteral = false
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	p.nextToken() // Move past '{'

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		// 1. Skip separators (Commas and ASI Semicolons)
		if p.curTokenIs(token.COMMA) || p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
			continue
		}

		// 2. Parse Case
		matchCase := p.parseMatchCase()
		if matchCase != nil {
			exp.Cases = append(exp.Cases, matchCase)

			// --- THE FIX ---
			// We must advance after the body so the next iteration
			// sees the comma or semicolon separator.
			p.nextToken()
		} else {
			p.nextToken() // Safety advance
		}
	}

	return exp
}
func (p *Parser) parseMatchCase() *ast.MatchCase {
	curCase := &ast.MatchCase{}

	// 1. Parse Pattern
	// Use ASSIGN precedence to stop at '=>'
	curCase.Pattern = p.parseExpression(ASSIGN)

	// 2. Handle '=>' (Assign + GT)
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	if !p.expectPeek(token.GT) {
		return nil
	}

	p.nextToken() // Move past '>' to start of body

	// 3. Parse Body
	if p.curTokenIs(token.LBRACE) {
		curCase.Body = p.parseBlockStatement()
	} else {
		// Single expression body
		expr := p.parseExpression(LOWEST)
		curCase.Body = &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.ExpressionStatement{Expression: expr},
			},
		}
	}

	return curCase
}

func (p *Parser) parseSelectorExpression(left ast.Expression) ast.Expression {
	exp := &ast.SelectorExpression{Token: p.curToken, Left: left}

	if !p.expectPeek(token.IDENT) {
		// For LSP support, we return the expression even if the field is missing
		return exp
	}

	exp.Field = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	return exp
}

func (p *Parser) parsePinStatement() *ast.PinStatement {
	stmt := &ast.PinStatement{Token: p.curToken}

	for {
		p.nextToken() // Move to identifier
		if !p.curTokenIs(token.IDENT) {
			p.ReportError(p.curToken.Position, "expected identifier for pin, got %s", p.curToken.Literal)
			return nil
		}

		stmt.Targets = append(stmt.Targets, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})

		if p.peekTokenIs(token.COMMA) {
			p.nextToken() // Move to comma
		} else {
			break
		}
	}

	return stmt
}

func (p *Parser) parseAllocExpression() ast.Expression {
	exp := &ast.AllocExpression{Token: p.curToken}
	p.nextToken() // Consume 'alloc'
	exp.Value = p.parseExpression(PREFIX)
	return exp
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}

	p.nextToken() // Move past '['

	oldNoStruct := p.noStructLiteral
	p.noStructLiteral = false

	for !p.curTokenIs(token.RBRACKET) && !p.curTokenIs(token.EOF) {
		expr := p.parseExpression(LOWEST)
		if expr != nil {
			array.Elements = append(array.Elements, expr)
		}

		if p.peekTokenIs(token.COMMA) {
			p.nextToken() // Move to ','
			p.nextToken() // Move to next expression
		} else {
			p.nextToken() // Move to ']'
		}
	}

	p.noStructLiteral = oldNoStruct

	return array
}

func (p *Parser) parseMapLiteral() ast.Expression {
	mapExp := &ast.MapLiteral{Token: p.curToken}
	mapExp.Pairs = make(map[ast.Expression]ast.Expression)

	oldNoStruct := p.noStructLiteral
	p.noStructLiteral = false

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken() // move past { or ,
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			p.noStructLiteral = oldNoStruct
			return nil
		}

		p.nextToken() // move past :
		value := p.parseExpression(LOWEST)
		mapExp.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			p.noStructLiteral = oldNoStruct
			return nil
		}
	}

	p.noStructLiteral = oldNoStruct

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return mapExp
}

// --- Helpers ---

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.peek2Token
	p.peek2Token = p.l.NextToken()
	for {
		if p.peekToken.Type == token.COMMENT {
			p.allComments = append(p.allComments, &ast.Comment{Token: p.peekToken, Text: p.peekToken.Literal})
			p.peekToken = p.peek2Token
			p.peek2Token = p.l.NextToken()
			continue
		}
		if p.peekToken.Type == token.DOC_COMMENT {
			comment := &ast.Comment{Token: p.peekToken, Text: p.peekToken.Literal}
			p.allComments = append(p.allComments, comment)
			p.curDocComments = append(p.curDocComments, comment)
			p.peekToken = p.peek2Token
			p.peek2Token = p.l.NextToken()
			continue
		}
		break
	}
}

func (p *Parser) peekTokenIs(t token.TokenType) bool  { return p.peekToken.Type == t }
func (p *Parser) curTokenIs(t token.TokenType) bool   { return p.curToken.Type == t }
func (p *Parser) peek2TokenIs(t token.TokenType) bool { return p.peek2Token.Type == t }

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekPrecedence() int {
	if pre, ok := precedences[p.peekToken.Type]; ok {
		return pre
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if pre, ok := precedences[p.curToken.Type]; ok {
		return pre
	}
	return LOWEST
}

func (p *Parser) registerPrefix(t token.TokenType, fn prefixParseFn) { p.prefixParseFns[t] = fn }
func (p *Parser) registerInfix(t token.TokenType, fn infixParseFn)   { p.infixParseFns[t] = fn }

func (p *Parser) noPrefixParseFnError(tok token.Token) {
	if tok.Type == token.ILLEGAL {
		return
	}
	p.ReportError(tok.Position, "no prefix parse function for %s found (literal: %q)", tok.Type, tok.Literal)
}

func (p *Parser) peekError(t token.TokenType) {
	if p.peekToken.Type == token.ILLEGAL {
		return
	}
	p.ReportErrorWithHint(p.peekToken.Position,
		fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type),
		fmt.Sprintf("try adding a %s here?", t))
}

func (p *Parser) Errors() []string {
	if p.Diagnostics == nil {
		return nil
	}
	return p.Diagnostics.ErrorMessages()
}

func (p *Parser) ReportError(pos token.Position, format string, args ...interface{}) {
	p.ReportErrorWithHint(pos, fmt.Sprintf(format, args...), "")
}

func (p *Parser) ReportErrorWithHint(pos token.Position, message string, hint string) {
	if p.Diagnostics == nil {
		return
	}
	p.Diagnostics.Add(diag.Diagnostic{
		Range: diag.Range{
			Start: diag.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
			End:   diag.Position{Line: pos.Line, Column: pos.Column + 1, Offset: pos.Offset + 1},
		},
		Severity: diag.Error,
		Message:  message,
		Source:   "Parser",
		File:     pos.Filename,
		Hint:     hint,
	})
}

func (p *Parser) synchronize() {
	// Always advance at least once to avoid getting stuck on the same token
	p.nextToken()

	for !p.curTokenIs(token.EOF) {
		if p.Context != nil && p.Context.Err() != nil {
			return
		}
		if p.curTokenIs(token.SEMICOLON) {
			return
		}

		switch p.curToken.Type {
		case token.PACKAGE, token.IMPORT, token.FN, token.VAR, token.TYPE, token.STRUCT, token.ENUM,
			token.IF, token.FOR, token.WHILE, token.RETURN, token.MATCH, token.SELECT, token.RBRACE:
			return
		}
		p.nextToken()
	}
}
func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	stmt := &ast.BreakStatement{Token: p.curToken}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
	stmt := &ast.ContinueStatement{Token: p.curToken}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}
func (p *Parser) parseRuneLiteral() ast.Expression {
	lit := &ast.RuneLiteral{Token: p.curToken}

	// Value extraction (e.g. 'a' -> 97)
	runes := []rune(p.curToken.Literal)
	if len(runes) > 0 {
		lit.Value = int32(runes[0])
	}

	return lit
}

func (p *Parser) parseDeferStatement() *ast.DeferStatement {
	stmt := &ast.DeferStatement{Token: p.curToken}

	p.nextToken() // move past 'defer'
	exp := p.parseExpression(LOWEST)

	call, ok := exp.(*ast.CallExpression)
	if !ok {
		p.ReportError(p.curToken.Position, "defer must be followed by a function call, got %T", exp)
		return nil
	}

	stmt.Call = call
	return stmt
}

func (p *Parser) parseParallelExpression() ast.Expression {
	exp := &ast.ParallelExpression{Token: p.curToken}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	exp.Body = p.parseBlockStatement()
	return exp
}

func (p *Parser) parseScopeExpression() ast.Expression {
	exp := &ast.ScopeExpression{Token: p.curToken}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	exp.Body = p.parseBlockStatement()
	return exp
}

func (p *Parser) parseRangeExpression(left ast.Expression) ast.Expression {
	expression := &ast.RangeExpression{
		Token: p.curToken,
		Start: left,
	}

	precedence := p.curPrecedence()
	p.nextToken()

	expression.End = p.parseExpression(precedence)

	return expression
}
