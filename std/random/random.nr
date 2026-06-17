package random

extern fn nr_i64_to_i32(x: i64) i32

extern fn nr_i32_to_f64(x: i32) f64

extern fn nr_time_now() f64

extern fn nr_f64_to_i64(x: f64) i64

pub type Rand = struct {
    state: i64
}

pub fn (self: &Rand) Seed(seed: i64) {
    self.state = seed
}

pub fn (self: &Rand) Next() i64 {
    // Knuth LCG parameters
    var multiplier = 6364136223846793005
    var increment = 1442695040888963407
    var next = self.state * multiplier + increment
    self.state = next
    return next
}

pub fn (self: &Rand) Int() i32 {
    var raw = self.Next()
    // Shift right by 32 bits to get high 32 bits
    var shifted = raw >> 32
    var val = nr_i64_to_i32(shifted)
    if (val < 0) {
        return -val
    }
    return val
}

pub fn (self: &Rand) IntRange(min: i32, max: i32) i32 {
    if (min >= max) {
        return min
    }
    var r = self.Int()
    return min + (r % (max - min))
}

pub fn (self: &Rand) Float() f64 {
    var r = self.Int()
    // 2147483647.0 is the max positive value of standard 32-bit signed integer
    return nr_i32_to_f64(r) / 2147483647.0
}

pub fn (self: &Rand) FloatRange(min: f64, max: f64) f64 {
    if (min >= max) {
        return min
    }
    var f = self.Float()
    return min + f * (max - min)
}

// Global generator state
var global_rand: Rand = Rand {
    state: 123456789
}

pub fn Seed(seed: i64) {
    global_rand.Seed(seed)
}

pub fn Int() i32 {
    return global_rand.Int()
}

pub fn IntRange(min: i32, max: i32) i32 {
    return global_rand.IntRange(min, max)
}

pub fn Float() f64 {
    return global_rand.Float()
}

pub fn FloatRange(min: f64, max: f64) f64 {
    return global_rand.FloatRange(min, max)
}
