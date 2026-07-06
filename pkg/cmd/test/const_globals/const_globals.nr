package main

// Global primitive constants
pub const MAX_RETRIES: i32 = 5
pub const TIMEOUT = 500

// Global string constant
const GREETING = "Hello from global const!"

// Global struct constant (compile-time instantiated)
pub type Config = struct {
    retries: i32
    timeout: i32
}

pub const DEFAULT_CONFIG = Config {
    retries: MAX_RETRIES,
    timeout: TIMEOUT,
}

fn test_local_const() i32 {
    const LOCAL_VAL = 42
    return LOCAL_VAL
}

fn test_global_consts() i32 {
    var sum = MAX_RETRIES + TIMEOUT
    sum = sum + DEFAULT_CONFIG.retries
    sum = sum + DEFAULT_CONFIG.timeout
    var s = GREETING
    return sum
}

pub fn main() i32 {
    test_local_const()
    test_global_consts()
    return 0
}
