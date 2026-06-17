package hash


pub type Hashable = interface {
    fn Hash() i32
}

pub fn (self: #i32) Hash() i32 {
    return self
}

pub fn (self: #i64) Hash() i32 {
    return i32(self ^ (self >> 32))
}

pub fn (self: #str) Hash() i32 {
    var hash = 5381
    var length = 0
    while self[length] != 0 {
        length = length + 1
    }
    var i = 0
    while i < length {
        hash = ((hash << 5) + hash) + i32(self[i])
        i = i + 1
    }
    return hash
}
