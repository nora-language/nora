package test_integration

import "io"

type Data = struct {
    id: i32
}

fn (self: #Data) drop() {
    io.PrintLn("Cleaning data")
}

fn process_data(d: #Data) {
    io.PrintLn("Processing data")
}

fn test_complex_topology() {
    var d1 = Data {
        id: 1
    }
    var d2 = Data {
        id: 2
    }

    var lease = #d1

    var i = 0
    while (i < 2) {
        if (i == 1) {
            lease = #d2
        }
        process_data(lease)
        i = i + 1
    }
}

fn test_parallel_capture() {
    var d = Data {
        id: 100
    }

    parallel {
        io.PrintLn("Async use of data")
    }
}

fn main() {
    test_complex_topology()
    test_parallel_capture()
    io.PrintLn("Integration Done")
}
