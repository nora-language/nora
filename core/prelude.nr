package main

// --- Option[T] Prelude Methods ---
pub fn (self: #Option[T]) IsSome[T]() bool {
    match self {
        Some(val) => {
            return true
        }
        None => {
            return false
        }
    }
}

pub fn (self: #Option[T]) IsNone[T]() bool {
    match self {
        Some(val) => {
            return false
        }
        None => {
            return true
        }
    }
}

pub fn (self: Option[T]) Unwrap[T]() T {
    match self {
        Some(val) => {
            return val
        }
        None => {
            panic("called Option.Unwrap() on a None value")
        }
    }
}

pub fn (self: Option[T]) UnwrapOr[T](def_val: T) T {
    match self {
        Some(val) => {
            return val
        }
        None => {
            return def_val
        }
    }
}

// --- Result[T, E] Prelude Methods ---
pub fn (self: #Result[T, E]) IsOk[T, E]() bool {
    match self {
        Ok(val) => {
            return true
        }
        Err(err) => {
            return false
        }
    }
}

pub fn (self: #Result[T, E]) IsErr[T, E]() bool {
    match self {
        Ok(val) => {
            return false
        }
        Err(err) => {
            return true
        }
    }
}

pub fn (self: Result[T, E]) Unwrap[T, E]() T {
    match self {
        Ok(val) => {
            return val
        }
        Err(err) => {
            panic("called Result.Unwrap() on an Err value")
        }
    }
}

pub fn (self: Result[T, E]) UnwrapOr[T, E](def_val: T) T {
    match self {
        Ok(val) => {
            return val
        }
        Err(err) => {
            return def_val
        }
    }
}

pub fn (self: Result[T, E]) UnwrapErr[T, E]() E {
    match self {
        Ok(val) => {
            panic("called Result.UnwrapErr() on an Ok value")
        }
        Err(err) => {
            return err
        }
    }
}

// --- Range Prelude Definition ---
pub type Range = struct {
    start: i32
    end: i32
    current: i32
}

pub fn (r: &Range) Next() Option[i32] {
    if r.current >= r.end {
        return None[i32]
    }
    var val = r.current
    r.current = r.current + 1
    return Some[i32](val)
}

