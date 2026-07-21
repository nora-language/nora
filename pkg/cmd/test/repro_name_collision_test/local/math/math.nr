package math

// Generic Sqrt - this is what nora_physics/src/math/math.nr provides
pub fn Sqrt[T](x: T) T {
    var zero = x - x
    if (x == zero) {
        return zero
    }
    var one = x / x
    var two = one + one
    var guess = x * one
    var i = 0
    while (i < 15) {
        if (guess == zero) {
            return zero
        }
        guess = (guess + x / guess) / two
        i = i + 1
    }
    return guess
}
