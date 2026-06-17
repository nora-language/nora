package lexer

import (
	"testing"

	"github.com/nora-language/nora/pkg/token"
)

func TestNextToken(t *testing.T) {
	// Input is formatted to avoid triggering ASI (no newlines between statements)
	input := `var five = 5; var ten = 10; type User = struct { name: String, age: u8 } fn add(x, y) { x + y } var result = add(five, ten); !-/*5; 5 < 10 > 5; if (5 < 10) { return true } else { return false } 10 == 10; 10 != 9; obj1 === obj2; "foobar" "foo bar" ` + "`raw string`" + ` 'a' 0x1_000 0b1101 12.34 5i 10.5j &^= ... ? spawn worker() parallel { }`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.VAR, "var"}, {token.IDENT, "five"}, {token.ASSIGN, "="}, {token.INT, "5"}, {token.SEMICOLON, ";"},
		{token.VAR, "var"}, {token.IDENT, "ten"}, {token.ASSIGN, "="}, {token.INT, "10"}, {token.SEMICOLON, ";"},

		// Type and Struct
		{token.TYPE, "type"}, {token.IDENT, "User"}, {token.ASSIGN, "="}, {token.STRUCT, "struct"}, {token.LBRACE, "{"},
		{token.IDENT, "name"}, {token.COLON, ":"}, {token.IDENT, "String"}, {token.COMMA, ","},
		{token.IDENT, "age"}, {token.COLON, ":"}, {token.IDENT, "u8"}, {token.RBRACE, "}"},

		// Function
		{token.FN, "fn"}, {token.IDENT, "add"}, {token.LPAREN, "("}, {token.IDENT, "x"}, {token.COMMA, ","}, {token.IDENT, "y"}, {token.RPAREN, ")"},
		{token.LBRACE, "{"}, {token.IDENT, "x"}, {token.PLUS, "+"}, {token.IDENT, "y"}, {token.RBRACE, "}"},

		// Var and Arithmetic
		{token.VAR, "var"}, {token.IDENT, "result"}, {token.ASSIGN, "="}, {token.IDENT, "add"}, {token.LPAREN, "("}, {token.IDENT, "five"}, {token.COMMA, ","}, {token.IDENT, "ten"}, {token.RPAREN, ")"}, {token.SEMICOLON, ";"},
		{token.BANG, "!"}, {token.MINUS, "-"}, {token.SLASH, "/"}, {token.ASTERISK, "*"}, {token.INT, "5"}, {token.SEMICOLON, ";"},

		// Comparisons
		{token.INT, "5"}, {token.LT, "<"}, {token.INT, "10"}, {token.GT, ">"}, {token.INT, "5"}, {token.SEMICOLON, ";"},

		// Control Flow
		{token.IF, "if"}, {token.LPAREN, "("}, {token.INT, "5"}, {token.LT, "<"}, {token.INT, "10"}, {token.RPAREN, ")"},
		{token.LBRACE, "{"}, {token.RETURN, "return"}, {token.TRUE, "true"}, {token.RBRACE, "}"},
		{token.ELSE, "else"}, {token.LBRACE, "{"}, {token.RETURN, "return"}, {token.FALSE, "false"}, {token.RBRACE, "}"},

		// Equality
		{token.INT, "10"}, {token.EQ, "=="}, {token.INT, "10"}, {token.SEMICOLON, ";"},
		{token.INT, "10"}, {token.NOT_EQ, "!="}, {token.INT, "9"}, {token.SEMICOLON, ";"},
		{token.IDENT, "obj1"}, {token.STRICT_EQ, "==="}, {token.IDENT, "obj2"}, {token.SEMICOLON, ";"},

		// Literals
		{token.STR, "foobar"}, {token.STR, "foo bar"}, {token.STR, "raw string"},
		{token.RUNE, "a"}, {token.INT, "0x1_000"}, {token.INT, "0b1101"},
		{token.FLOAT, "12.34"}, {token.IMAG, "5i"}, {token.IMAG, "10.5j"},

		// Operators
		{token.AND_NOT_ASSIGN, "&^="}, {token.ELLIPSIS, "..."}, {token.QUESTION, "?"},

		// Concurrency
		{token.SPAWN, "spawn"}, {token.IDENT, "worker"}, {token.LPAREN, "("}, {token.RPAREN, ")"},
		{token.PARALLEL, "parallel"}, {token.LBRACE, "{"}, {token.RBRACE, "}"},

		{token.EOF, ""},
	}

	l := New(input, "test.nr")

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestStructLexing(t *testing.T) {
	input := `
    type User = struct {
        id: i32,
        name: str
    }
	`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.TYPE, "type"},
		{token.IDENT, "User"},
		{token.ASSIGN, "="},
		{token.STRUCT, "struct"},
		{token.LBRACE, "{"},
		{token.IDENT, "id"},
		{token.COLON, ":"},
		{token.IDENT, "i32"},
		{token.COMMA, ","},
		{token.IDENT, "name"},
		{token.COLON, ":"},
		{token.IDENT, "str"},
		{token.SEMICOLON, ";"}, // <--- EXPECTED: Injected after 'str\n'
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"}, // <--- EXPECTED: Injected after '}\n' (ends TypeStatement)
		{token.EOF, ""},
	}
	l := New(input, "test.nr")

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
func TestSemicolonInsertion(t *testing.T) {
	input := `var x = 10
fn add(a, b) {
    return a + b
}
	
spawn work()
var y = 20`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.VAR, "var"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"}, // Injected after 10\n

		{token.FN, "fn"},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "a"},
		{token.COMMA, ","},
		{token.IDENT, "b"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.IDENT, "a"},
		{token.PLUS, "+"},
		{token.IDENT, "b"},
		{token.SEMICOLON, ";"}, // Injected after b\n
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"}, // Injected after }\n

		{token.SPAWN, "spawn"},
		{token.IDENT, "work"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"}, // Injected after ()\n

		{token.VAR, "var"},
		{token.IDENT, "y"},
		{token.ASSIGN, "="},
		{token.INT, "20"},
		{token.EOF, ""}, // EOF also acts as a terminator
	}

	l := New(input, "test.nr")

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestSemicolonInsertionWithComment(t *testing.T) {
	input := `pub fn scanln() // comment
pub fn FormatInt()`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.PUB, "pub"},
		{token.FN, "fn"},
		{token.IDENT, "scanln"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.COMMENT, "comment"},
		{token.SEMICOLON, ";"}, // Injected after ) despite trailing comment
		{token.PUB, "pub"},
		{token.FN, "fn"},
		{token.IDENT, "FormatInt"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.EOF, ""},
	}

	l := New(input, "test.nr")

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
