package main

import "io"

type User = struct {
    name: str
}

// 1. u : #User (Read Borrow)
// By default, borrowing must be explicitly marked with # to prevent accidental moves.
fn read_implicit(u: #User) {
    // u.name = "new name" // This would be a COMPILE ERROR: u is immutable
    io.PrintLn("Read (Implicit): ${u.name}")
}

// 2. u : #User (Explicit Read Borrow)
// This is the explicit version of a read-only borrow.
// It behaves identically to the implicit version above.
fn read_explicit(u: #User) {
    io.PrintLn("Read (Explicit): ${u.name}")
}

// 3. u : &User (Write Lease / Mutable Borrow)
// This allows the function to modify the caller's data.
// The Topology Solver ensures that no other leases exist while this is active.
fn update(u: &User) {
    u.name = "Alice Updated"
    io.PrintLn("Updated: ${u.name}")
}

// 4. u : @User (Move Lease / Ownership Transfer)
// This transfers ownership of the resource from the caller to this function.
// Once called, the caller's variable is 'dead' and cannot be used again.
fn consume(u: @User) {
    io.PrintLn("Consuming: ${u.name}")
    // The Topology Solver will automatically drop 'u' here at the end of the scope.
}

fn main() i32 {
    // --- Root Ownership ---
    // u1 is the owner of the User struct.
    // The solver tracks its lifecycle and will drop it at the end of main
    // IF it hasn't been moved.
    var u1 = User {
        name: "Alice"
    }

    io.PrintLn("--- Starting Lease Demo ---")

    // CASE A: Implicit Borrow
    // u1 is borrowed explicitly.
    read_implicit(u1)

    // CASE B: Explicit Borrow (#)
    // u1 is borrowed, not moved.
    read_explicit(u1)

    // CASE C: Mutable Borrow (&)
    // u1 is modified in-place.
    update(u1)
    io.PrintLn("Main: name is now '${u1.name}'")

    // CASE D: Ownership Move (@)
    // u1 is moved into the function.
    consume(u1)

    // --- TERMINAL CASE: Use After Move ---
    // The line below would cause a COMPILER ERROR because u1 was moved above.
    // io.PrintLn("${u1.name}")
    io.PrintLn("--- Demo Completed Successfully ---")
    return 0
}
