package main

import "collections"
import "io"
import "string"

fn assert(cond: bool, msg: str) {
    if (!cond) {
        panic("Assertion failed: " + msg)
    }
}

fn test_metadata() {
    io.PrintLn("Testing Metadata & Comparison...")
    assert(string.Length("hello") == 5, "Length of 'hello' should be 5")
    assert(string.Length("") == 0, "Length of empty string should be 0")

    assert(string.Equals("abc", "abc"), "'abc' should equal 'abc'")
    assert(!string.Equals("abc", "def"), "'abc' should not equal 'def'")
    assert(!string.Equals("abc", "ABC"), "'abc' should not equal 'ABC'")

    assert(string.Compare("abc", "abc") == 0, "Compare('abc', 'abc') should be 0")
    assert(string.Compare("abc", "def") < 0, "Compare('abc', 'def') should be negative")
    assert(string.Compare("def", "abc") > 0, "Compare('def', 'abc') should be positive")
}

fn test_casing_and_trim() {
    io.PrintLn("Testing Casing and Trim operations...")

    var u1 = string.ToUpper("hello 123")
    assert(string.Equals(u1, "HELLO 123"), "ToUpper failed")

    var l1 = string.ToLower("HeLLo! 123")
    assert(string.Equals(l1, "hello! 123"), "ToLower failed")

    var t1 = string.Trim("  abc  ")
    assert(string.Equals(t1, "abc"), "Trim failed")

    var t2 = string.Trim("\n\t  abc  \r\n")
    assert(string.Equals(t2, "abc"), "Trim with tab/newlines failed")

    var t3 = string.Trim("   ")
    assert(string.Equals(t3, ""), "Trim on whitespace-only failed")

    var tl1 = string.TrimLeft("  abc  ")
    assert(string.Equals(tl1, "abc  "), "TrimLeft failed")

    var tr1 = string.TrimRight("  abc  ")
    assert(string.Equals(tr1, "  abc"), "TrimRight failed")
}

fn test_substring_and_search() {
    io.PrintLn("Testing Substring and Search...")
    var s = "hello nora nora"
    assert(string.Contains(s, "nora"), "Contains 'nora' failed")
    assert(!string.Contains(s, "world"), "Contains 'world' should be false")

    assert(string.HasPrefix(s, "hello"), "HasPrefix 'hello' failed")
    assert(!string.HasPrefix(s, "nora"), "HasPrefix 'nora' should be false")

    assert(string.HasSuffix(s, "nora"), "HasSuffix 'nora' failed")
    assert(!string.HasSuffix(s, "hello"), "HasSuffix 'hello' should be false")

    assert(string.IndexOf(s, "nora") == 6, "IndexOf 'nora' failed")
    assert(string.IndexOf(s, "world") == -1, "IndexOf 'world' should be -1")
    assert(string.IndexOf(s, "") == 0, "IndexOf empty should be 0")

    assert(string.LastIndexOf(s, "nora") == 11, "LastIndexOf 'nora' failed")
    assert(string.LastIndexOf(s, "world") == -1, "LastIndexOf 'world' should be -1")

    // Substring tests capturing owned returns in local variables
    var sub1 = string.Substring("hello", 1, 4)
    assert(string.Equals(sub1, "ell"), "Substring(1, 4) failed")

    var sub2 = string.Substring("hello", -3, -1)
    assert(string.Equals(sub2, "ll"), "Substring(-3, -1) failed")

    var sub3 = string.Substring("hello", 0, 100)
    assert(string.Equals(sub3, "hello"), "Substring(0, 100) failed")

    var sub4 = string.Substring("hello", 3, 2)
    assert(string.Equals(sub4, ""), "Substring(3, 2) failed")
}

fn test_classifications() {
    io.PrintLn("Testing Classifications...")
    assert(string.IsNumeric("12345"), "'12345' is numeric")
    assert(!string.IsNumeric("123a45"), "'123a45' is not numeric")
    assert(!string.IsNumeric(""), "empty is not numeric")

    assert(string.IsAlpha("NoraLanguage"), "'NoraLanguage' is alpha")
    assert(!string.IsAlpha("Nora 123"), "'Nora 123' is not alpha")

    assert(string.IsAlphaNumeric("Nora123"), "'Nora123' is alphanumeric")
    assert(!string.IsAlphaNumeric("Nora 123!"), "'Nora 123!' is not alphanumeric")
}

fn test_transformations() {
    io.PrintLn("Testing Transformations...")

    var r1 = string.Repeat("abc", 3)
    assert(string.Equals(r1, "abcabcabc"), "Repeat failed")

    var r2 = string.Repeat("abc", 0)
    assert(string.Equals(r2, ""), "Repeat 0 should be empty")

    var rev1 = string.Reverse("hello")
    assert(string.Equals(rev1, "olleh"), "Reverse failed")

    var rev2 = string.Reverse("")
    assert(string.Equals(rev2, ""), "Reverse empty failed")

    var rep1 = string.Replace("hello nora nora", "nora", "world")
    assert(string.Equals(rep1, "hello world world"), "Replace failed")

    var rep2 = string.Replace("hello nora", "world", "nora")
    assert(string.Equals(rep2, "hello nora"), "Replace no match failed")
}

fn test_split_and_join() {
    io.PrintLn("Testing Split and Join with collections.Vector...")
    var parts = string.Split("hello,world,nora", ",")
    assert(parts.Len() == 3, "Split parts len should be 3")

    var p0: #str = parts.Get(0)
    io.PrintLn("p0 value: " + p0)
    assert(string.Equals(p0, "hello"), "Split part 0 failed")

    var p1: #str = parts.Get(1)
    assert(string.Equals(p1, "world"), "Split part 1 failed")

    var p2: #str = parts.Get(2)
    assert(string.Equals(p2, "nora"), "Split part 2 failed")

    var joined = string.Join(parts, " - ")
    assert(string.Equals(joined, "hello - world - nora"), "Join failed")

    // Split on empty separator
    var chars = string.Split("nora", "")
    assert(chars.Len() == 4, "Split on empty sep len should be 4")

    var c0: #str = chars.Get(0)
    assert(string.Equals(c0, "n"), "Split char 0 failed")

    var c3: #str = chars.Get(3)
    assert(string.Equals(c3, "a"), "Split char 3 failed")
}

fn main() i32 {
    io.PrintLn("--- Starting std/string Premium Test Suite ---")

    test_metadata()
    test_casing_and_trim()
    test_substring_and_search()
    test_classifications()
    test_transformations()
    test_split_and_join()

    io.PrintLn("std/string Premium Test Suite passed successfully!")
    return 0
}
