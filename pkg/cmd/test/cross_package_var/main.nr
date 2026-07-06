package main

import "./pkg_a"

fn main() {
    var c = pkg_a.GlobalConfig
    var m = pkg_a.MAX
    
    // just use them
    if c.val == 5 && m == 10 {
        return
    }
}
