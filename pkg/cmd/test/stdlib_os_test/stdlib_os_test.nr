package main

import "io"
import "os"

fn assert_bool(val: bool, expected: bool, name: str) {
    if (val != expected) {
        io.PrintLn("Assertion failed for: ")
        io.PrintLn(name)
        panic("Error: boolean assertion failed")
    }
}

fn main() {
    io.PrintLn("--- Starting std/os Premium Test Suite ---")

    // 1. Verify Working Directory
    io.PrintLn("Testing Working Directory...")
    var cwd_opt = os.GetWD()
    if (!cwd_opt.IsSome()) {
        panic("Error: GetWD failed to retrieve path")
    }
    var cwd = cwd_opt.Unwrap()
    io.PrintLn("Current Working Directory: " + cwd)

    // 2. Verify Command Line Arguments
    io.PrintLn("Testing Command Line Arguments...")
    var args = os.Args()
    var arg_len = args.Len()
    io.PrintLn("Number of arguments: " + io.FormatInt(arg_len))
    assert_bool(arg_len > 0, true, "Args count > 0")

    var exec_path = args.Get(0)
    io.PrintLn("Executable path: " + exec_path)

    // 3. Verify Environment Variables
    io.PrintLn("Testing Environment Variables...")

    // Ensure test key starts empty/unset
    var _ignore = os.UnsetEnv("NORA_TEST_KEY")
    var initial_val = os.GetEnv("NORA_TEST_KEY")
    if (initial_val.IsSome()) {
        panic("Error: NORA_TEST_KEY should be unset initially")
    }

    // Set variable
    var set_ok = os.SetEnv("NORA_TEST_KEY", "nora_premium_os_works")
    assert_bool(set_ok, true, "SetEnv success")

    // Get variable
    var get_opt = os.GetEnv("NORA_TEST_KEY")
    if (!get_opt.IsSome()) {
        panic("Error: GetEnv failed to retrieve set value")
    }
    var get_val = get_opt.Unwrap()
    io.PrintLn("Environment variable NORA_TEST_KEY value: " + get_val)
    if (get_val != "nora_premium_os_works") {
        panic("Error: retrieved environment variable mismatch")
    }

    // Unset variable
    var unset_ok = os.UnsetEnv("NORA_TEST_KEY")
    assert_bool(unset_ok, true, "UnsetEnv success")

    // Get variable again (should be None)
    var get_opt_after = os.GetEnv("NORA_TEST_KEY")
    if (get_opt_after.IsSome()) {
        panic("Error: NORA_TEST_KEY should be unset")
    }

    io.PrintLn("std/os Premium Test Suite passed successfully!")

    // 4. Verify process exit FFI hook
    io.PrintLn("Testing Exit FFI hook...")
    os.Exit(0)
}
