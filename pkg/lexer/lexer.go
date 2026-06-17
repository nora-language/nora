package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"context"
	"github.com/DwiYI/Project-Nora/pkg/diag"
	"github.com/DwiYI/Project-Nora/pkg/token"
)

type Lexer struct {
	Filename     string
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           rune // current char under examination
	prevToken    token.TokenType

	// New Coordinate Tracking
	line   int
	column int

	// Nesting depths for semicolon injection control
	parenDepth   int
	bracketDepth int

	// Byte offset in the parent file where input[0] begins (0 for standalone lexing).
	baseOffset int

	Diagnostics *diag.Collection
	Context     context.Context // Added for cancellation
}

func New(input string, filename string) *Lexer {
	return NewAt(input, filename, 1, 1, 0)
}

// NewAt creates a lexer for input that begins at the given 1-based line/column in the parent file.
// offset is the byte offset in the parent file where input[0] starts (0 when unknown).
func NewAt(input string, filename string, line, col, offset int) *Lexer {
	l := &Lexer{
		input:        input,
		Filename:     filename,
		line:         line,
		column:       col - 1,
		position:     -1,
		readPosition: 0,
		baseOffset:   offset,
		Diagnostics:  &diag.Collection{},
	}
	l.readChar()
	return l
}

func (l *Lexer) fileOffset() int {
	if l.position < 0 {
		return l.baseOffset
	}
	return l.baseOffset + l.position
}

func (l *Lexer) AdjustCoordinates(line int, col int) {
	l.line = line
	l.column = col
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
		l.position = l.readPosition
		l.readPosition++
		return
	}

	r, size := utf8.DecodeRuneInString(l.input[l.readPosition:])
	l.ch = r

	// Logic for coordinate advancement
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}

	l.position = l.readPosition
	l.readPosition += size
}

// peekChar allows us to look ahead without advancing the position.
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

func (l *Lexer) NextToken() token.Token {
	if l.Context != nil && l.Context.Err() != nil {
		return token.Token{Type: token.EOF}
	}
	var tok token.Token

	// Skip horizontal whitespace
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}

	// Handle Newlines for Semicolon Injection
	if l.ch == '\n' {
		if l.shouldInsertSemicolon() {
			// Create the virtual semicolon
			pos := token.Position{Line: l.line, Column: l.column, Offset: l.fileOffset(), Filename: l.Filename}
			tok = token.Token{Type: token.SEMICOLON, Literal: ";", Position: pos}

			l.prevToken = token.SEMICOLON
			l.readChar() // Move past the \n
			return tok
		}
		// If not inserting, just consume newline and recurse
		l.readChar()
		return l.NextToken()
	}

	// Capture position for the actual token
	pos := token.Position{Line: l.line, Column: l.column, Offset: l.fileOffset(), Filename: l.Filename}

	switch l.ch {

	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			if l.peekChar() == '=' { // Handle ===
				l.readChar()
				tok = token.Token{Type: token.STRICT_EQ, Literal: string(ch) + "=" + "=", Position: pos}
			} else {
				tok = token.Token{Type: token.EQ, Literal: string(ch) + "=", Position: pos}
			}
		} else {
			tok = l.newToken(token.ASSIGN, l.ch)
		}
	case '+':
		if l.peekChar() == '+' {
			l.readChar()
			tok = token.Token{Type: token.INC, Literal: "++", Position: pos}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.PLUS_ASSIGN, Literal: "+=", Position: pos}
		} else {
			tok = l.newToken(token.PLUS, l.ch)
		}
	case '-':
		if l.peekChar() == '-' {
			l.readChar()
			tok = token.Token{Type: token.DEC, Literal: "--", Position: pos}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.MINUS_ASSIGN, Literal: "-=", Position: pos}
		} else {
			tok = l.newToken(token.MINUS, l.ch)
		}
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok = token.Token{Type: token.LAND, Literal: "&&", Position: pos}
		} else if l.peekChar() == '^' {
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = token.Token{Type: token.AND_NOT_ASSIGN, Literal: "&^=", Position: pos}
			} else {
				tok = token.Token{Type: token.AND_NOT, Literal: "&^", Position: pos}
			}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.AND_ASSIGN, Literal: "&=", Position: pos}
		} else {
			tok = l.newToken(token.AND, l.ch)
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = token.Token{Type: token.LOR, Literal: "||", Position: pos}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.OR_ASSIGN, Literal: "|=", Position: pos}
		} else {
			tok = l.newToken(token.OR, l.ch)
		}
	case '<':
		if l.peekChar() == '-' {
			l.readChar()
			tok = token.Token{Type: token.ARROW, Literal: "<-", Position: pos}
		} else if l.peekChar() == '<' {
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = token.Token{Type: token.SHL_ASSIGN, Literal: "<<=", Position: pos}
			} else {
				tok = token.Token{Type: token.SHL, Literal: "<<", Position: pos}
			}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.LT_EQ, Literal: "<=", Position: pos}
		} else {
			tok = l.newToken(token.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '>' {
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = token.Token{Type: token.SHR_ASSIGN, Literal: ">>=", Position: pos}
			} else {
				tok = token.Token{Type: token.SHR, Literal: ">>", Position: pos}
			}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.GT_EQ, Literal: ">=", Position: pos}
		} else {
			tok = l.newToken(token.GT, l.ch)
		}
	case '^':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.XOR_ASSIGN, Literal: "^=", Position: pos}
		} else {
			tok = l.newToken(token.XOR, l.ch)
		}
	case '%':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.REM_ASSIGN, Literal: "%=", Position: pos}
		} else {
			tok = l.newToken(token.REM, l.ch)
		}
	case '*':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.ASTERISK_ASSIGN, Literal: "*=", Position: pos}
		} else {
			tok = l.newToken(token.ASTERISK, l.ch)
		}
	case '/':
		if l.peekChar() == '/' {
			l.readChar() // move to second /
			if l.peekChar() == '/' {
				l.readChar() // move to third /

				tok.Type = token.DOC_COMMENT
				tok.Literal = l.readDocComment()
				tok.Position = pos
				return tok
			}

			tok.Type = token.COMMENT
			tok.Literal = l.readComment()
			tok.Position = pos
			return tok
		}
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.SLASH_ASSIGN, Literal: "/=", Position: pos}
		} else {
			tok = l.newToken(token.SLASH, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: string(ch) + "=", Position: pos}
		} else {
			tok = l.newToken(token.BANG, l.ch)
		}
	case ':':
		tok = l.newToken(token.COLON, l.ch)
	case ';':
		tok = l.newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = l.newToken(token.COMMA, l.ch)
	case '.':
		if l.peekChar() == '.' {
			l.readChar()
			if l.peekChar() == '.' {
				l.readChar()
				tok = token.Token{Type: token.ELLIPSIS, Literal: "...", Position: pos}
			} else {
				tok = token.Token{Type: token.DOT_DOT, Literal: "..", Position: pos}
			}
		} else {
			tok = l.newToken(token.DOT, l.ch)
		}
	case '(':
		l.parenDepth++
		tok = l.newToken(token.LPAREN, l.ch)
	case ')':
		l.parenDepth--
		tok = l.newToken(token.RPAREN, l.ch)
	case '{':
		tok = l.newToken(token.LBRACE, l.ch)
	case '}':
		tok = l.newToken(token.RBRACE, l.ch)
	case '[':
		l.bracketDepth++
		tok = l.newToken(token.LBRACKET, l.ch)
	case ']':
		l.bracketDepth--
		tok = l.newToken(token.RBRACKET, l.ch)
	case '~':
		tok = l.newToken(token.TILDE, l.ch)
	case '?':
		tok = l.newToken(token.QUESTION, l.ch)
	case '#':
		tok = l.newToken(token.HASH, l.ch)
	case '@':
		tok = l.newToken(token.MOVE, l.ch)

	case '"', '`':
		tok.Type = token.STR
		tok.Literal = l.readString(l.ch)
		tok.Position = pos
		l.prevToken = tok.Type
		return tok
	case '\'':
		tok.Type = token.RUNE
		tok.Literal = l.readRune()
		tok.Position = pos
		l.prevToken = tok.Type
		return tok
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
		tok.Position = pos
		return tok
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			tok.Position = pos
			l.prevToken = tok.Type
			return tok
		} else if isDigit(l.ch) {
			lit, kind := l.readNumber()
			tok.Type = kind
			tok.Literal = lit
			tok.Position = pos
			l.prevToken = tok.Type
			return tok
		} else {
			tok = l.newToken(token.ILLEGAL, l.ch)
			l.ReportError(tok.Position, "illegal character: %q (code: %d)", l.ch, l.ch)
		}
	}

	l.readChar()
	l.prevToken = tok.Type
	return tok
}

func (l *Lexer) readDocComment() string {
	l.readChar() // skip the third '/'
	if l.ch == ' ' {
		l.readChar() // optionally skip one leading space
	}
	position := l.position
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readComment() string {
	l.readChar() // skip the second '/'
	if l.ch == ' ' {
		l.readChar() // optionally skip one leading space
	}
	position := l.position
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readString(quote rune) string {
	l.readChar() // Skip opening quote
	position := l.position
	braceDepth := 0
	for (l.ch != quote || braceDepth > 0) && l.ch != 0 {
		if l.ch == '\n' {
			pos := token.Position{Line: l.line, Column: l.column, Offset: l.fileOffset(), Filename: l.Filename}
			l.ReportError(pos, "unclosed string literal")
			break
		}
		if l.ch == '$' && l.peekChar() == '{' {
			l.readChar() // skip $
			l.readChar() // skip {
			braceDepth++
		} else if l.ch == '{' && braceDepth > 0 {
			braceDepth++
			l.readChar()
		} else if l.ch == '}' && braceDepth > 0 {
			braceDepth--
			l.readChar()
		} else if l.ch == '\\' {
			l.readChar() // skip backslash
			l.readChar() // skip escaped char
		} else {
			l.readChar()
		}
	}
	result := l.input[position:l.position]
	if l.ch == quote {
		l.readChar() // skip closing quote
	}
	return result
}

func (l *Lexer) readRune() string {
	l.readChar() // skip '
	position := l.position
	if l.ch == '\\' {
		l.readChar() // skip \
	}
	l.readChar() // skip char
	result := l.input[position:l.position]
	if l.ch == '\'' {
		l.readChar() // skip '
	}
	return result
}

func (l *Lexer) readNumber() (string, token.TokenType) {
	position := l.position
	var kind token.TokenType = token.INT

	// Check for hex, binary, octal
	if l.ch == '0' {
		peek := l.peekChar()
		if peek == 'x' || peek == 'X' {
			l.readChar() // 0
			l.readChar() // x
			for isHexDigit(l.ch) || l.ch == '_' || l.ch == '.' || l.ch == 'p' || l.ch == 'P' || ((l.ch == '+' || l.ch == '-') && (l.input[l.position-1] == 'p' || l.input[l.position-1] == 'P')) {
				if l.ch == '.' || l.ch == 'p' || l.ch == 'P' {
					kind = token.FLOAT
				}
				l.readChar()
			}
			l.consumeSuffix()
			return l.input[position:l.position], kind
		}
		if peek == 'b' || peek == 'B' {
			l.readChar() // 0
			l.readChar() // b
			for l.ch == '0' || l.ch == '1' || l.ch == '_' {
				l.readChar()
			}
			l.consumeSuffix()
			return l.input[position:l.position], token.INT
		}
		if peek == 'o' || peek == 'O' {
			l.readChar() // 0
			l.readChar() // o
			for (l.ch >= '0' && l.ch <= '7') || l.ch == '_' {
				l.readChar()
			}
			l.consumeSuffix()
			return l.input[position:l.position], token.INT
		}
	}

	for isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.peekChar()) {
		kind = token.FLOAT
		l.readChar()
		for isDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}
	}

	// Scientific notation
	if l.ch == 'e' || l.ch == 'E' {
		kind = token.FLOAT
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}
	}

	// Imaginary suffix (i, j) or type suffix (i32, f64, etc.)
	suffixStart := l.position
	l.consumeSuffix()

	suffix := l.input[suffixStart:l.position]
	if suffix == "i" || suffix == "j" {
		kind = token.IMAG
	}

	return l.input[position:l.position], kind
}

func (l *Lexer) consumeSuffix() {
	if isLetter(l.ch) {
		for isLetter(l.ch) || isDigit(l.ch) {
			l.readChar()
		}
	}
}

func isHexDigit(ch rune) bool {
	return isDigit(ch) || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}

func (l *Lexer) shouldInsertSemicolon() bool {
	// RULE: Do not insert semicolons inside nested expressions like () or []
	if l.parenDepth > 0 || l.bracketDepth > 0 {
		return false
	}

	switch l.prevToken {
	case token.IDENT, token.INT, token.FLOAT, token.STR, token.RUNE,
		token.RETURN, token.BREAK, token.CONTINUE, token.INC, token.DEC,
		token.RPAREN, token.RBRACKET, token.RBRACE, token.TRUE, token.FALSE, token.NONE:
		return true
	default:
		// If last was a comma, never insert
		return false
	}
}

/* Helper Functions */

func (l *Lexer) newToken(tokenType token.TokenType, ch rune) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(ch),
		Position: token.Position{
			Line:     l.line,
			Column:   l.column,
			Offset:   l.position,
			Filename: l.Filename,
		},
	}
}

func (l *Lexer) ReportError(pos token.Position, format string, args ...interface{}) {
	if l.Diagnostics == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.Diagnostics.Add(diag.Diagnostic{
		Range: diag.Range{
			Start: diag.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
			End:   diag.Position{Line: pos.Line, Column: pos.Column + 1, Offset: pos.Offset + 1},
		},
		Severity: diag.Error,
		Message:  msg,
		Source:   "Lexer",
		File:     l.Filename,
	})
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || unicode.IsDigit(ch)
}

func (l *Lexer) GetBlankLines() []int {
	lines := strings.Split(l.input, "\n")
	var blankLines []int
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			blankLines = append(blankLines, i+1)
		}
	}
	return blankLines
}
