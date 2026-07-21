package main

import "io"
import "./pkg_core_user"
import "./pkg_local_user"

fn main() {
    pkg_core_user.RunCoreTest()
    pkg_local_user.RunLocalTest()
    io.PrintLn("done")
}
