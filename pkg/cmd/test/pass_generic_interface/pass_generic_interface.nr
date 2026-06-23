package main

import "io"

// ---------------------------------------------------------
// Scenario 1: Standard generic struct implementing non-generic interface
// ---------------------------------------------------------
pub type ComponentMeta = interface {
    fn Init(cap: i32) i32
}

pub type ComponentMetaImpl[T] = struct {
    cid: i32
}

pub fn (_m: #ComponentMetaImpl[T]) Init(cap: i32) i32 {
    return cap
}

// ---------------------------------------------------------
// Scenario 2: Generic struct method returning generic type `T` 
// matches an interface method returning concrete type.
// ---------------------------------------------------------
pub type IntProvider = interface {
    fn Get() i32
}

pub type ValueProvider[T] = struct {
    val: T
}

pub fn (v: @ValueProvider[T]) Get() T {
    return v.val
}

// ---------------------------------------------------------
// Scenario 3: Multiple type parameters and generic methods
// ---------------------------------------------------------
pub type PairProcessor = interface {
    fn Process() i32
}

pub type Pair[T, U] = struct {
    first: T,
    second: U
}

pub fn (_p: #Pair[T, U]) Process() i32 {
    return 42
}

// Having a generic method on the struct should NOT break its ability
// to implement the interface!
pub fn (_p: #Pair[T, U]) Map[V](val: V) V {
    return val
}

fn main() i32 {
    io.PrintLn("--- Generic Interface Tests ---")
    
    // Test Scenario 1
    var impl1 = ComponentMetaImpl[i32] { cid: 1 }
    var iface1: ComponentMeta = impl1
    io.Print("Scenario 1 Init: ")
    io.PrintLn(iface1.Init(5))

    // Test Scenario 2
    var impl2 = ValueProvider[i32] { val: 100 }
    var iface2: IntProvider = impl2
    io.Print("Scenario 2 Get: ")
    io.PrintLn(iface2.Get())

    // Test Scenario 3
    var impl3 = Pair[i32, str] { first: 1, second: "hello" }
    // Call the generic method on the struct just to ensure it compiles
    io.Print("Scenario 3 Map: ")
    io.PrintLn(impl3.Map[i32](99))

    var iface3: PairProcessor = impl3
    io.Print("Scenario 3 Process: ")
    io.PrintLn(iface3.Process())

    io.PrintLn("All generic interface scenarios passed!")
    return 0
}
