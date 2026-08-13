package pkg_a

pub type Handle = struct { id: u64 }

pub fn NewHandle() Handle {
    return Handle { id: 0 }
}
