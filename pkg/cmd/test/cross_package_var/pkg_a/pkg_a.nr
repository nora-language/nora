package pkg_a

pub type Config = struct {
    val: i32
}

pub var GlobalConfig: Config = Config { val: 5 }
pub const MAX: i32 = 10
