package main
import "io"

fn printMessage(ch: #chan[i32]) {
    io.println("Hello from a concurrent Nora Fiber compiled via GCC!")
    ch <- 1
}

fn signalChannel(ch: #chan[i32]) {
    ch <- 1
}

fn main() i32 {
    io.println("=========================================================")
    io.println("   Welcome to Nora Decoupled Toolchain Verification")
    io.println("=========================================================")
    io.println("Targeting GCC for Native Compilation...")
    
    // Use a channel to synchronize both fibers
    var c = make(chan[i32], 2)
    spawn printMessage(#c)
    spawn signalChannel(#c)
    
    <-c
    <-c
    io.println("System successfully built and executed under GCC!")
    io.println("=========================================================")
    return 0
}
