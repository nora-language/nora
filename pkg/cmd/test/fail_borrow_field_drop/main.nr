package main

import "io"

pub type SubStruct = struct {
    data: i32[]
}

pub type Container = struct {
    sub: @SubStruct
}

pub fn NewSubStruct() @SubStruct {
    var data = alloc i32[10]
    return @SubStruct {
        data: @data
    }
}

pub fn GetContainer() @Container {
    var sub = NewSubStruct()
    return alloc Container {
        sub: @sub
    }
}

pub fn CheckBug(c: #Container) {
    // This is the bug trigger: accessing a borrowed field and assigning to var
    var s = c.sub
    
    // The compiler lowers this to a pointer assignment in C
    // and sets nr_df_s = true, which drops the underlying data of the borrowed struct!
    io.PrintLn("Accessing borrowed field done")
}

pub fn main() {
    var c = GetContainer()
    CheckBug(#c)
    
    // This will now crash or print garbage because CheckBug dropped c.sub!
    var s = c.sub
    io.PrintLn("Finished without crash")
}
