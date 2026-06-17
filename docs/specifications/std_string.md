# Standard String Utilities

## Overview

Nora's `std/string` module provides essential routines for parsing, mutating, and manipulating strings. In Nora, the `str` type represents a UTF-8 (or ASCII) string, backed by a C-style null-terminated byte array. String literals are read-only, but dynamically allocated strings (e.g., returned by `Substring`) own their underlying memory. 

## Metadata & Comparison

*   `Length(s: #str) i32`: Returns the number of bytes in the string.
*   `Equals(s1: #str, s2: #str) bool`: Returns `true` if the strings are identical in content.
*   `Compare(s1: #str, s2: #str) i32`: Performs a lexicographical comparison. Returns `0` if equal, `<0` if `s1` is less than `s2`, and `>0` if `s1` is greater.

## Substrings & Searching

*   `Contains(s: #str, substr: #str) bool`: Returns `true` if `substr` exists anywhere within `s`.
*   `HasPrefix(s: #str, prefix: #str) bool`: Returns `true` if `s` begins with `prefix`.
*   `HasSuffix(s: #str, suffix: #str) bool`: Returns `true` if `s` ends with `suffix`.
*   `Substring(s: #str, start: i32, end: i32) str`: Allocates and returns a new string containing the characters from `start` (inclusive) to `end` (exclusive). Supports negative indexing (e.g., `-1` for the last character).
*   `IndexOf(s: #str, substr: #str) i32`: Returns the first index of `substr` in `s`, or `-1` if not found.
*   `LastIndexOf(s: #str, substr: #str) i32`: Returns the last index of `substr` in `s`, or `-1` if not found.

## Transformations

*   `ToUpper(s: #str) str`: Returns a new string with all lowercase ASCII letters converted to uppercase.
*   `ToLower(s: #str) str`: Returns a new string with all uppercase ASCII letters converted to lowercase.
*   `Trim(s: #str) str`: Returns a new string with all leading and trailing whitespace characters (spaces, tabs, newlines) removed.
*   `TrimLeft(s: #str) str`: Removes only leading whitespace.
*   `TrimRight(s: #str) str`: Removes only trailing whitespace.
*   `Repeat(s: #str, count: i32) str`: Returns a new string consisting of `s` repeated `count` times.
*   `Reverse(s: #str) str`: Returns a new string with the characters of `s` in reverse order.
*   `Replace(s: #str, old_str: #str, new_str: #str) str`: Returns a new string where all non-overlapping occurrences of `old_str` are replaced by `new_str`.

## Splitting & Joining

*   `Split(s: #str, sep: #str) @collections.Vector[str]`: Splits the string `s` around occurrences of the separator `sep` and returns an owned Vector of the newly allocated substring fragments.
*   `Join(parts: #collections.Vector[str], sep: #str) str`: Concatenates a vector of strings into a single string, inserting the `sep` string between each element.

## Classifications

*   `IsNumeric(s: #str) bool`: Returns `true` if the string contains only ASCII digits (0-9).
*   `IsAlpha(s: #str) bool`: Returns `true` if the string contains only ASCII letters (A-Z, a-z).
*   `IsAlphaNumeric(s: #str) bool`: Returns `true` if the string contains only letters and digits.
