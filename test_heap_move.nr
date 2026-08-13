package main
pub type World = struct { val: i32 }
pub fn NewWorld() @World { return alloc World{val: 0} }
pub fn TakesWorld(w: @World) {}
pub fn main() {
    var w = NewWorld()
    TakesWorld(@w)
}
