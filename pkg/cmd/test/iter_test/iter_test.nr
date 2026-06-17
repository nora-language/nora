import "io"

pub type IntIterator = interface {
    fn Next() Option[i32]
}

pub type ArrayIter = struct {
    arr: #i32[]
    idx: i32
}

[inline]
pub fn (a: &ArrayIter) Next() Option[i32] {
    if a.idx >= len(a.arr) {
        return None[i32]
    }
    var val = a.arr[a.idx]
    a.idx = a.idx + 1
    return Some[i32](val)
}

pub type TakeIter[I: IntIterator] = struct {
    iter: I
    limit: i32
    current: i32
}

[inline]
pub fn (t: &TakeIter[I]) Next[I: IntIterator]() Option[i32] {
    if t.current >= t.limit {
        return None[i32]
    }
    t.current = t.current + 1
    return t.iter.Next()
}

pub fn main() {
    var data = alloc i32[5]
    data[0] = 10
    data[1] = 20
    data[2] = 30
    data[3] = 40
    data[4] = 50

    var arr_iter = alloc ArrayIter { arr: data, idx: 0 }
    
    var take_iter = alloc TakeIter[&ArrayIter] {
        iter: arr_iter,
        limit: 2,
        current: 0,
    }

    var sum = 0

    // Take only the first 2 elements (10 + 20)
    var opt1 = take_iter.Next[&ArrayIter]()
    if opt1.IsSome[i32]() {
        sum = sum + opt1.Unwrap[i32]()
    }

    var opt2 = take_iter.Next[&ArrayIter]()
    if opt2.IsSome[i32]() {
        sum = sum + opt2.Unwrap[i32]()
    }

    var opt3 = take_iter.Next[&ArrayIter]()
    if opt3.IsSome[i32]() {
        // Should not reach here
        sum = sum + 999
    }

    if sum == 30 {
        io.PrintLn("success")
    } else {
        io.PrintLn("failed")
    }
}
