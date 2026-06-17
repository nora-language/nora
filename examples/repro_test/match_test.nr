
import "io"

type Status = enum {
    Ready
    Loading
    Error(msg: str)
}

fn test_match_int(x: i32) {
    match x {
        1 => io.println("one")
        2 => io.println("two")
        _ => io.println("something else")
    }
}

fn test_match_str(s: str) {
    match s {
        "hello" => io.println("world")
        "foo" => io.println("bar")
        _ => io.println("unknown")
    }
}

fn test_match_binding(x: i32) {
    match x {
        1 => io.println("one")
        val => io.println("got $val")
    }
}

fn main() {
    test_match_int(1)
    test_match_int(3)
    test_match_str("hello")
    test_match_str("baz")
    test_match_binding(1)
    test_match_binding(42)

    var res = match 10 {
        10 => 100
        _ => 0
    }
    io.println("res: $res")

    var s = Status.Loading
    match s {
        Ready => io.println("Ready")
        Loading => io.println("Loading")
        Error(msg) => io.println("Error: $msg")
    }
}
