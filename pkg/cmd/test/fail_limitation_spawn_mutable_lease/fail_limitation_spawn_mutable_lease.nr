package main

fn worker(val: &i32) {
    val = val + 1
}

pub fn main() {
    var x = 10

    // REAL UNRESOLVED LIMITATION #3: Parallel Array Element Mutation Across Fibers (spawn &array[i])
    // What you want to write:
    // Spawning fibers over mutable write leases (&x or &bodies[i]) inside loops for multi-threading.
    //
    // The Reality / Compiler Error:
    // Error: cannot pass mutable lease (write) across fiber boundary
    // The static lease solver cannot prove at compile-time that array element indices across fibers are disjoint.

    scope {
        spawn worker(&x) // FAILS
    }
}
