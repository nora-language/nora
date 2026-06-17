package collections

pub type Iterator[T] = interface {
    fn (self: Iterator[T]) Next() Option[#T]
}
