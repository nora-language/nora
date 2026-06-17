package format

import (
	"strings"
	"testing"

	"github.com/nora-language/nora/pkg/lexer"
	"github.com/nora-language/nora/pkg/parser"
)

func formatSource(t *testing.T, src string) string {
	t.Helper()
	l := lexer.New(src, "test.nr")
	p := parser.New(l)
	file := p.Parse("test.nr")
	if p.Diagnostics.HasErrors() {
		t.Fatalf("parse errors: %v", p.Diagnostics)
	}
	return New(DefaultConfig()).Format(file)
}

func TestFormatMatchMultiline(t *testing.T) {
	src := `package main

fn unwrap_err() {
    match self {
        Ok(val) => {
            panic("called Result.UnwrapErr() on an Ok value")
        }
        Err(err) => {
            return err
        }
    }
}
`
	out := formatSource(t, src)

	if strings.Contains(out, "match self { Ok(val)") {
		t.Fatalf("match should not collapse to one line:\n%s", out)
	}
	if !strings.Contains(out, "Ok(val) => {\n") {
		t.Fatalf("expected multiline match arm, got:\n%s", out)
	}
	if !strings.Contains(out, "panic(\"called Result.UnwrapErr() on an Ok value\")") {
		t.Fatalf("expected panic message preserved, got:\n%s", out)
	}
}

func TestFormatNormalizeBlankLines(t *testing.T) {
	in := "line1\n\n\n\nline2\n\n\nline3\n"
	out := normalizeBlankLines(in)
	if strings.Contains(out, "\n\n\n") {
		t.Fatalf("expected at most one consecutive blank line, got:\n%q", out)
	}
	if !strings.Contains(out, "line1\n\nline2") {
		t.Fatalf("expected single blank line between line1 and line2, got:\n%q", out)
	}
}

func TestFormatIncrementDecrement(t *testing.T) {
	src := `package main

fn test() {
    max++
    min--
    val = val + 1
}
`
	out := formatSource(t, src)
	expected := `package main

fn test() {
    max++
    min--
    val = val + 1
}
`
	if out != expected {
		t.Fatalf("expected:\n%s\n\ngot:\n%s", expected, out)
	}
}

func TestFormatAdditionalASTNodes(t *testing.T) {
	src := `package main

fn test() {
    parallel {
        task_a()
        task_b()
    }
    spawn task_c()
    var arr = [1, 2, 3]
    var m = {"a": 1, "b": 2, "c": 3}
    var slice = arr[1:2]
}
`
	out := formatSource(t, src)
	expected := `package main

fn test() {
    parallel {
        task_a()
        task_b()
    }
    spawn task_c()
    var arr = [1, 2, 3]
    var m = {"a": 1, "b": 2, "c": 3}
    var slice = arr[1:2]
}
`
	if out != expected {
		t.Fatalf("expected:\n%s\n\ngot:\n%s", expected, out)
	}
}

func TestFormatInterpolationNoParentheses(t *testing.T) {
	src := `package main

fn test() {
    var val = "the result is ${a + b}"
}
`
	out := formatSource(t, src)
	expected := `package main

fn test() {
    var val = "the result is ${a + b}"
}
`
	if out != expected {
		t.Fatalf("expected:\n%s\n\ngot:\n%s", expected, out)
	}
}

func TestFormatNewFormatterRefinements(t *testing.T) {
	src := `package main

import "io" // inline import comment

[macro("io_macro", "expand_println")]
pub fn println()

fn test() {
    var x = 5 // inline var comment
    if x > 0 { // '+'
        return
        // comment under it
    } // inline closing brace comment
    select {
        case msg = <-ch1:
            io.println(msg)
        default:
            io.println("default")
    }
    var longExpr = a && b && c && d && e && f && g && h && i && j && k
}
`
	out := formatSource(t, src)
	expected := `package main

import "io" // inline import comment

[macro("io_macro", "expand_println")]
pub fn println()

fn test() {
    var x = 5 // inline var comment
    if x > 0 { // '+'
        return
        // comment under it
    } // inline closing brace comment
    select {
        case msg = <-ch1:
            io.println(msg)
        default:
            io.println("default")
    }
    var longExpr = a && b && c && d && e && f && g && h &&
        i && j && k
}
`
	if out != expected {
		t.Fatalf("expected:\n%s\n\ngot:\n%s", expected, out)
	}
}
