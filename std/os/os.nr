package os

import "collections"

// FFI declarations matching the C runtime helpers defined in emitOSRuntime()
extern fn nr_get_argc() i32

extern fn nr_get_argv(index: i32) str

extern fn nr_os_getenv(key: str) str

extern fn nr_os_setenv(key: str, val: str) bool

extern fn nr_os_unsetenv(key: str) bool

extern fn nr_os_getcwd() str

extern fn exit(code: i32)

// Args returns the command-line arguments passed to the program as a safe, owned Vector[str].
pub fn Args() @collections.Vector[str] {
    var count = nr_get_argc()
    var args = collections.NewVector[str](count)
    var i = 0
    while i < count {
        args.Push[str](nr_get_argv(i))
        i = i + 1
    }
    return args
}

// GetEnv retrieves the value of the environment variable named by the key.
// If the variable is not present, it returns None.
pub fn GetEnv(key: str) Option[str] {
    var val = nr_os_getenv(key)
    if (val == "") {
        return None[str]
    }
    return Some[str](val)
}

// SetEnv sets the value of the environment variable named by the key.
// Returns true if the operation succeeded, false otherwise.
pub fn SetEnv(key: str, value: str) bool {
    return nr_os_setenv(key, value)
}

// UnsetEnv unsets/deletes the environment variable named by the key.
// Returns true if the operation succeeded, false otherwise.
pub fn UnsetEnv(key: str) bool {
    return nr_os_unsetenv(key)
}

// GetWD returns the current working directory path.
// If it fails to determine the working directory, it returns None.
pub fn GetWD() Option[str] {
    var s = nr_os_getcwd()
    if (s == "") {
        return None[str]
    }
    return Some[str](s)
}

// Exit terminates the program with the specified exit status code.
pub fn Exit(code: i32) {
    exit(code)
}
