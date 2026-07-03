package pkg_b

import "../pkg_a"

pub type Handle = struct { handle: pkg_a.Handle }

pub fn NewHandle() Handle {
    return Handle { handle: pkg_a.NewHandle() }
}
