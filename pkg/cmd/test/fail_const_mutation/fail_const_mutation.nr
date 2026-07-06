package main

pub const MAX_RETRIES: i32 = 5

pub fn main() i32 {
    MAX_RETRIES = 10 // Violation: cannot assign to 'MAX_RETRIES' (it is a Constant)
    
    const LOCAL_VAL = 42
    LOCAL_VAL = 50 // Violation: cannot assign to 'LOCAL_VAL' (it is a Constant)
    
    return 0
}
