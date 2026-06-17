package token

type TokenType string

type Position struct {
	Line     int
	Column   int
	Offset   int
	Filename string
}

type Token struct {
	Type     TokenType
	Literal  string
	Position Position
}

func (t Token) EndPosition() Position {
	// Note: This is a simple approximation. Multiline literals would need more complex logic.
	return Position{
		Line:     t.Position.Line,
		Column:   t.Position.Column + len(t.Literal),
		Offset:   t.Position.Offset + len(t.Literal),
		Filename: t.Position.Filename,
	}
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + Literals
	IDENT       = "IDENT"       // add, foobar, x, y, ...
	INT         = "INT"         // 134345, 0x1A, 0b1101
	FLOAT       = "FLOAT"       // 3.14, 0x1.fp3
	IMAG        = "IMAG"        // 5i, 10j
	STR         = "STR"         // "hello", `raw`, """multi"""
	RUNE        = "RUNE"        // 'a'
	DOC_COMMENT = "DOC_COMMENT" // /// ...
	COMMENT     = "COMMENT"     // // ...

	// lease
	HASH = "#"
	MOVE = "@"

	// Arithmetic Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	REM      = "%"
	INC      = "++"
	DEC      = "--"

	// Compound Assignment
	PLUS_ASSIGN     = "+="
	MINUS_ASSIGN    = "-="
	ASTERISK_ASSIGN = "*="
	SLASH_ASSIGN    = "/="
	REM_ASSIGN      = "%="

	// Bitwise Operators
	AND     = "&"
	OR      = "|"
	XOR     = "^"
	TILDE   = "~"  // Bitwise NOT
	AND_NOT = "&^" // Bit Clear (Go-style)
	SHL     = "<<"
	SHR     = ">>"

	// Bitwise Compound
	AND_ASSIGN     = "&="
	OR_ASSIGN      = "|="
	XOR_ASSIGN     = "^="
	AND_NOT_ASSIGN = "&^="
	SHL_ASSIGN     = "<<="
	SHR_ASSIGN     = ">>="

	// Comparison Operators
	EQ        = "=="
	NOT_EQ    = "!="
	STRICT_EQ = "===" // Identity Equality (Lease-based)
	LT        = "<"
	GT        = ">"
	LT_EQ     = "<="
	GT_EQ     = ">="

	// Logical Operators
	LAND = "&&"
	LOR  = "||"

	// Punctuation & Scoping
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	DOT       = "."
	DOT_DOT   = ".."
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	LBRACKET  = "["
	RBRACKET  = "]"
	ELLIPSIS  = "..."
	QUESTION  = "?"  // Error unwrapping (Try operator)
	ARROW     = "<-" // Channels

	// Keywords
	PACKAGE = "PACKAGE"
	IMPORT  = "IMPORT"
	FN      = "FN"
	VAR     = "VAR"
	ENUM    = "ENUM"

	PIN      = "PIN" // Manual lease override - Keep leases alive during C calls
	IF       = "IF"
	ELSE     = "ELSE"
	FOR      = "FOR"
	WHILE    = "WHILE" // While loop support
	IN       = "IN"    // For-in loop
	MATCH    = "MATCH" // Pattern matching
	TYPE     = "TYPE"  // Struct/Sum type definition
	STRUCT   = "STRUCT"
	PROTOCOL = "PROTOCOL" // Interface/Trait
	RETURN   = "RETURN"
	BREAK    = "BREAK"
	CONTINUE = "CONTINUE"
	DEFER    = "DEFER"
	SPAWN    = "SPAWN"    // Start Fiber
	PARALLEL = "PARALLEL" // Multi-core block
	SCOPE    = "SCOPE"    // Structured concurrency block
	ALLOC    = "ALLOC"    // Heap allocation

	CHAN      = "CHAN"   // Communication
	EXTERN    = "EXTERN" // FFI -  Declare C functions
	EXPORT    = "EXPORT" // Reverse FFI -  Allow C to call Nora
	TRUE      = "TRUE"
	FALSE     = "FALSE"
	NONE      = "NONE" // Null/Option alternative
	SELECT    = "SELECT"
	CASE      = "CASE"
	DEFAULT   = "DEFAULT"
	INTERFACE = "INTERFACE"
	PUB       = "PUB"
)

// Keywords mapping for the Lexer
var keywords = map[string]TokenType{
	"package": PACKAGE,
	"import":  IMPORT,
	"fn":      FN,
	"var":     VAR,
	"enum":    ENUM,

	"pin":       PIN,
	"if":        IF,
	"else":      ELSE,
	"for":       FOR,
	"while":     WHILE,
	"in":        IN,
	"match":     MATCH,
	"type":      TYPE,
	"struct":    STRUCT,
	"protocol":  PROTOCOL,
	"return":    RETURN,
	"break":     BREAK,
	"continue":  CONTINUE,
	"defer":     DEFER,
	"spawn":     SPAWN,
	"parallel":  PARALLEL,
	"scope":     SCOPE,
	"alloc":     ALLOC,
	"chan":      CHAN,
	"extern":    EXTERN,
	"export":    EXPORT,
	"true":      TRUE,
	"false":     FALSE,
	"none":      NONE,
	"select":    SELECT,
	"case":      CASE,
	"default":   DEFAULT,
	"interface": INTERFACE,
	"pub":       PUB,
}

// LookupIdent checks if an identifier is a reserved keyword
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
