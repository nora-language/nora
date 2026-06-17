package iter

import "prelude"

pub type Iterator[T] = interface {
    fn Next() Option[T]
}

pub type TakeIter[I: Iterator[T], T] = struct {
    iter: I
    limit: i32
    current: i32
}

[inline]
pub fn (t: &TakeIter[I, T]) Next() Option[T] {
    if t.current >= t.limit {
        return None[T]
    }
    t.current = t.current + 1
    return t.iter.Next()
}

pub type FilterIter[I: Iterator[T], T] = struct {
    iter: I
    predicate: fn(#T) bool
}

[inline]
pub fn (f: &FilterIter[I, T]) Next() Option[T] {
    while true {
        var opt = f.iter.Next()
        if opt.IsNone[T]() {
            return None[T]
        }
        var val = opt.Unwrap[T]()
        if f.predicate(#val) {
            return Some[T](val)
        }
    }
    return None[T]
}

pub type MapIter[I: Iterator[T], T, U] = struct {
    iter: I
    mapper: fn(T) U
}

[inline]
pub fn (m: &MapIter[I, T, U]) Next() Option[U] {
    var opt = m.iter.Next()
    if opt.IsNone[T]() {
        return None[U]
    }
    var val = opt.Unwrap[T]()
    var mapped = m.mapper(val)
    return Some[U](mapped)
}

pub type StepByIter[I: Iterator[T], T] = struct {
    iter: I
    step: i32
    first: bool
}

[inline]
pub fn (s: &StepByIter[I, T]) Next() Option[T] {
    if s.first {
        s.first = false
        return s.iter.Next()
    }
    var skips = s.step - 1
    var i = 0
    while i < skips {
        var dummy = s.iter.Next()
        if dummy.IsNone[T]() {
            return None[T]
        }
        i = i + 1
    }
    return s.iter.Next()
}

pub type SkipIter[I: Iterator[T], T] = struct {
    iter: I
    skip: i32
    skipped: i32
}

[inline]
pub fn (s: &SkipIter[I, T]) Next() Option[T] {
    while s.skipped < s.skip {
        var dummy = s.iter.Next()
        if dummy.IsNone[T]() {
            return None[T]
        }
        s.skipped = s.skipped + 1
    }
    return s.iter.Next()
}

// --- Range Iterator Combinators ---

pub fn (r: &prelude.Range) Take(limit: i32) TakeIter[&prelude.Range, i32] {
    return TakeIter[&prelude.Range, i32] {
        iter: r
        limit: limit
        current: 0
    }
}

pub fn (r: &prelude.Range) Filter(predicate: fn(#i32) bool) FilterIter[&prelude.Range, i32] {
    return FilterIter[&prelude.Range, i32] {
        iter: r
        predicate: predicate
    }
}

pub fn (r: &prelude.Range) Map(mapper: fn(i32) i32) MapIter[&prelude.Range, i32, i32] {
    return MapIter[&prelude.Range, i32, i32] {
        iter: r
        mapper: mapper
    }
}

pub fn (r: &prelude.Range) StepBy(step: i32) StepByIter[&prelude.Range, i32] {
    return StepByIter[&prelude.Range, i32] {
        iter: r
        step: step
        first: true
    }
}

pub fn (r: &prelude.Range) Skip(skip: i32) SkipIter[&prelude.Range, i32] {
    return SkipIter[&prelude.Range, i32] {
        iter: r
        skip: skip
        skipped: 0
    }
}
