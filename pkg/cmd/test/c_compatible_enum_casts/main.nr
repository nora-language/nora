package main

import "io"

// 1. Standard signed 32-bit enum
[repr("i32")]
pub type Status = enum {
    Ok = 0,
    Error = -1,
    Pending = 2,
}

// 2. Unsigned 8-bit enum
[repr("u8")]
pub type Flags = enum {
    NoFlags = 0,
    Read = 1,
    Write = 2,
    Execute = 4,
}

pub fn main() i32 {
    io.PrintLn("--- Testing c_compatible_enum Casts ---")

    // --- Scenario 1: Enum to Integer ---
    
    // Status to i32
    var status_err = Status.Error;
    var status_i32 = i32(status_err);
    if status_i32 != -1 {
        io.PrintLn("FAIL: Expected Status.Error to cast to -1, got", status_i32)
        panic("Enum to i32 cast failed")
    }

    // Flags to u8
    var flag_exec = Flags.Execute;
    var flag_u8 = u8(flag_exec);
    if flag_u8 != 4 {
        io.PrintLn("FAIL: Expected Flags.Execute to cast to 4, got", flag_u8)
        panic("Enum to u8 cast failed")
    }

    // Status to i64 (Widening cast via i32)
    var status_i64 = i64(status_err);
    if status_i64 != -1 {
        io.PrintLn("FAIL: Expected Status.Error to cast to i64 -1, got", status_i64)
        panic("Enum to i64 cast failed")
    }

    // --- Scenario 2: Integer to Enum ---

    // i32 to Status
    var int_ok: i32 = 0;
    var restored_status = Status(int_ok);
    if restored_status != Status.Ok {
        io.PrintLn("FAIL: Expected 0 to cast to Status.Ok")
        panic("i32 to Enum cast failed")
    }

    // u8 to Flags
    var int_read: u8 = 1;
    var restored_flag = Flags(int_read);
    if restored_flag != Flags.Read {
        io.PrintLn("FAIL: Expected 1 to cast to Flags.Read")
        panic("u8 to Enum cast failed")
    }

    // --- Scenario 3: Cross Integer Casting ---

    // i32 to Flags (Nora's enum constructor handles primitive coercion natively)
    var int_write: i32 = 2;
    var cross_flag = Flags(int_write);
    if cross_flag != Flags.Write {
        io.PrintLn("FAIL: Expected i32 2 to cast to Flags.Write")
        panic("Cross-primitive to Enum cast failed")
    }

    io.PrintLn("PASS: All enum cast scenarios succeeded!")
    return 0;
}
