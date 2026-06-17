package main

import "io"
import "collections"

pub type MyInterface = interface {
    fn DoSomething(val: i32)
}

pub type MyImpl = struct {
    x: i32
}

pub fn (m: &MyImpl) DoSomething(val: i32) {
    m.x = val
}

pub fn main() {
    var impl = alloc MyImpl{ x: 0 }
    var iface: MyInterface = impl
    var map = collections.NewHashMap[i32, MyInterface](16)
    map.Insert(1, @iface)
    
    var m_opt = map.GetMut(1)
    match m_opt {
        Some(m) => {
            var m_ptr = m
            m_ptr.DoSomething(42)
        }
        None => {}
    }
}
