# Standard Library: OS

## Overview

The `std/os` package provides a platform-independent interface to operating system functionality. The design is Unix-like, although error handling is Nora-idiomatic (returning `Option[Error]`). It handles file manipulation, environment variables, system execution, and path abstractions.

## Core Types

### `File`
Represents an open file descriptor. It implements `io.Reader`, `io.Writer`, and `io.Closer`.
```nora
pub type File = struct { ... }
```

## Functions

### File Operations
*   `Open(name: str) (*File, Option[Error])`: Opens the named file for reading.
*   `Create(name: str) (*File, Option[Error])`: Creates or truncates the named file.
*   `Remove(name: str) Option[Error]`: Removes the named file or (empty) directory.
*   `Rename(oldpath: str, newpath: str) Option[Error]`: Renames a file.

### Environment & System Variables
*   `Getenv(key: str) Option[str]`: Retrieves the value of the environment variable named by the key.
*   `Setenv(key: str, value: str) Option[Error]`: Sets the value of the environment variable.
*   `Args() []str`: Returns the command-line arguments, starting with the program name.
*   `Exit(code: i32)`: Causes the current program to exit with the given status code.

## Path Separation
Platform-specific path separators are provided natively:
*   `os.PathSeparator`: `\` on Windows, `/` on Unix.

## RAII Integration
Files managed by `os.File` implicitly close themselves via compiler-injected RAII `Drop()` calls if not manually closed, effectively preventing file-descriptor exhaustion.
