package main

import "io"

// 1. A universal Iterator interface over type T
pub type Iterator[T] = interface {
    fn Next() Option[T]
}

// 2. A concrete array iterator that implements Iterator[T]
pub type ArrayIter[T] = struct {
    arr: #T[]
    idx: i32
}

pub fn (a: &ArrayIter[T]) Next() Option[T] {
    if a.idx >= len(a.arr) {
        return None[T]
    }
    var val = a.arr[a.idx]
    a.idx = a.idx + 1
    return Some[T](val)
}

// 3. A TakeIter combinator that accepts ANY iterator 'I' that yields 'T'
// This is the problematic syntax in Nora:
pub type TakeIter[I: Iterator[T], T] = struct {
    iter: I
    limit: i32
    current: i32
}

pub fn (t: &TakeIter[I, T]) Next() Option[T] {
    if t.current >= t.limit {
        return None[T]
    }
    t.current = t.current + 1
    return t.iter.Next()
}

pub fn main() {
    var data = alloc i32[5]
    data[0] = 10
    var arr_iter = alloc ArrayIter[i32] { arr: data, idx: 0 }
    
    var take_iter = alloc TakeIter[&ArrayIter[i32], i32] {
        iter: arr_iter,
        limit: 1,
        current: 0,
    }

    var opt1 = take_iter.Next()
    if opt1.IsSome[i32]() {
        io.PrintLn("success")
    }
}
