package ops

pub type Addable[T, R] = interface {
    fn add(other: T) R
}

pub type Subtractable[T, R] = interface {
    fn sub(other: T) R
}

pub type Multipliable[T, R] = interface {
    fn mul(other: T) R
}

pub type Divisible[T, R] = interface {
    fn div(other: T) R
}

pub type Negatable[R] = interface {
    fn neg() R
}

pub type Notable[R] = interface {
    fn not() R
}

pub type BitNotable[R] = interface {
    fn bitnot() R
}

pub type Indexable[K, V] = interface {
    fn index(key: K) V
}

pub type IndexMutable[K, V] = interface {
    fn index_mut(key: K) &V
}

pub type BitAndable[T, R] = interface {
    fn bitand(other: T) R
}

pub type BitOrable[T, R] = interface {
    fn bitor(other: T) R
}

pub type BitXorable[T, R] = interface {
    fn bitxor(other: T) R
}

pub type ShiftLeftable[T, R] = interface {
    fn shl(other: T) R
}

pub type ShiftRightable[T, R] = interface {
    fn shr(other: T) R
}
