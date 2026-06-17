# Code Formatting

## Overview

Consistent code style is critical for readability and maintainability across large codebases. Nora enforces formatting natively via the `pkg/format` module, which is invoked by running the `Nora fmt` CLI command or automatically triggered by the LSP on file save.

## Execution

Running `Nora fmt` in the root of a workspace will recursively find and format all `.nr` files in the project. The formatter parses the code into an Abstract Syntax Tree (AST) and then rebuilds the text output, ensuring that the visual layout strictly maps to the syntactic structure regardless of how the developer typed it.

## Configuration

While Nora's formatter establishes strong defaults, it allows for minor project-specific deviations. The formatter looks for a configuration file (often loaded implicitly or specified in the tooling config mapping to `format.Config`) at the root of the project.

The configuration fields include:

*   **`indent_size` (int):** Determines the number of spaces for a single indentation level (Default: `4`).
*   **`use_tabs` (bool):** If true, uses standard hardware tabs instead of spaces for indentation (Default: `false`).
*   **`max_line_width` (int):** The soft limit for line length. The formatter will attempt to wrap long expressions, function arguments, or chained method calls that exceed this limit (Default: `100`).
*   **`organize_imports` (bool):** If true, the formatter automatically groups and alphabetically sorts all `import` declarations at the top of the file (Default: `true`).

## Formatting Rules

1.  **Block Indentation:** The contents of braces `{}` (functions, structs, logic blocks) are always indented exactly one level deeper than the surrounding scope.
2.  **Vertical Spacing:** The formatter condenses multiple consecutive blank lines into a single blank line.
3.  **Operator Spacing:** Binary operators (`+`, `=`, `==`) are padded with a single space on either side.
4.  **Parameter Lists:** Commas separating parameters or arguments are followed by a single space, but not preceded by one.
