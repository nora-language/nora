package cmp

pub type Equatable[T] = interface {
    fn eq(other: #T) bool
}

pub type Comparable[T] = interface {
    fn cmp(other: #T) i32
}
