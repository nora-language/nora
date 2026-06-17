package format

import (
	"fmt"

	"github.com/nora-language/nora/pkg/lexer"
	"github.com/nora-language/nora/pkg/parser"
)

func testFormat(name string, src string) {
	l := lexer.New(src, "test.nr")
	p := parser.New(l)
	file := p.Parse("test.nr")
	if p.Diagnostics.HasErrors() {
		fmt.Printf("%s: Parse error\n", name)
		return
	}
	formatted := New(DefaultConfig()).Format(file)
	fmt.Printf("=== %s ===\n%s\n", name, formatted)
}

func main() {
	// Test 1: Match expression with proper brace formatting
	testFormat("Match Expression", `pub fn test() {
    match value {
        A => {
            return 1
        }
        B => {
            return 2
        }
        C => {
            return 3
        }
    }
}`)

	// Test 2: If-else blocks
	testFormat("If-Else Blocks", `pub fn test() {
    if x > 0 {
        println("positive")
    } else if x < 0 {
        println("negative")
    } else {
        println("zero")
    }
}`)

	// Test 3: Nested structures
	testFormat("Nested Blocks", `pub fn test() {
    for i in 0..10 {
        if i % 2 == 0 {
            println(i)
        }
    }
}`)

	// Test 4: While loop
	testFormat("While Loop", `pub fn test() {
    while x < 100 {
        x = x + 1
    }
}`)

	// Test 5: Type definition with variants
	testFormat("Type Definition", `pub type Result[T, E] = enum {
    Ok(T)
    Err(E)
}`)

	// Test 6: Function with receiver
	testFormat("Method Definition", `pub fn (self: MyType) Method() {
    return self.value
}`)
}
