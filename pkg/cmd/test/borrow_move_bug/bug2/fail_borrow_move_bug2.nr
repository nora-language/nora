package main

pub type TextureDesc = struct {
    name: str
}

pub type Pass = struct {
    desc: TextureDesc
}

pub fn (self: #Pass) GetDesc() @TextureDesc {
    // This should fail because self is a read-only borrow, 
    // and we are trying to MOVE self.desc.name into a new struct literal!
    return alloc TextureDesc {
        name: self.desc.name
    }
}

pub fn main() {
    
}
