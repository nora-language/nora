import "io"

pub fn main() {
    io.PrintLn("--- Range Object Test ---")
    
    var r = 1..5
    
    // Manual loop
    var count = 0
    while true {
        var opt = r.Next[i32]()
        if opt.IsNone[i32]() {
            break
        }
        var val = opt.Unwrap[i32]()
        io.PrintLn("Manual loop tick")
        count = count + 1
    }
    
    if count != 4 {
        panic("FAILED! Manual count should be 4")
    }

    // For loop using iterator over new range
    count = 0
    for i in 5..10 {
        io.PrintLn("For loop tick")
        count = count + 1
    }
    
    if count != 5 {
        panic("FAILED! For loop count should be 5")
    }
    
    io.PrintLn("SUCCESS")
}
