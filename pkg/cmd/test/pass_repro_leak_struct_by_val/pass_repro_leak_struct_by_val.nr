package main

pub type Node = struct {
    data: @i64[]
}

pub fn main() {
    var arr = make(Node[], 10)
    arr[0] = Node { data: make(i64[], 10) }
    
    var val = @arr[0]
}
