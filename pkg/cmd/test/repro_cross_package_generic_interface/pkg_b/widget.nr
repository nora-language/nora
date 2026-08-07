package pkg_b

pub type Widget = struct {
    name: str
}

pub fn (w: Widget) to_string() str {
    return w.name
}
