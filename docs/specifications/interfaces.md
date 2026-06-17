# Interfaces & Protocols

## Overview

Nora uses `interface` (or `protocol`) types to enable dynamic dispatch and polymorphism. Interfaces define a set of method signatures. Any concrete type that implements these methods automatically satisfies the interface—no explicit `implements` declaration is needed.

## Motivation

While generic functions (type-erased monomorphization) provide static polymorphism without runtime overhead, there are times when a collection needs to hold heterogeneous types, or a function needs to operate on abstract behavior dynamically. Interfaces solve this by providing existential types and virtual method tables (vtables) under the hood.

## Syntax

### 1. Defining an Interface

An interface is defined using the `interface` keyword and contains a list of function signatures.

```nora
pub type Greeter = interface {
    fn greet() void
}
```

### 2. Implementing an Interface

Implementation is implicit. If a type defines methods that match the interface signatures exactly, it can be assigned to a variable of that interface type.

```nora
pub type User = struct {
    name: str
}

pub fn (self: #User) greet() {
    io.PrintLn("Hello, I am ${self.name}")
}
```

### 3. Usage

You can assign the concrete struct directly to the interface variable. Nora will automatically construct the existential type (pairing the data pointer with the vtable).

```nora
fn main() {
    var u = User { name: "Alice" }
    var g: Greeter = u
    
    // Dynamic dispatch
    g.greet() 
}
```

### 4. The Existential `any` Type

Nora features a built-in `any` type, which serves as an empty interface (meaning every type in the language implicitly satisfies it). The `any` type allows for safe heterogeneous data storage, dynamic type coercion, and safe downcasting, serving as a type-safe replacement for raw C `void*` pointers.

```nora
var heterogeneous_list = alloc make(Vector[any], 10)
heterogeneous_list.Push[any](User { name: "Alice" })
heterogeneous_list.Push[any](100)
```

## Semantics & Type Rules

1.  **Implicit Satisfaction:** Structs do not declare they implement an interface. They just need to have the matching methods (same name, arguments, and return types).
2.  **Receiver Types:** If the interface method signature expects a mutable lease (`&T`) but the struct method expects a read-only lease (`#T`), the assignment may fail depending on variance rules. Usually, the method receiver in the concrete implementation dictates how it can be boxed.
3.  **Dynamic Dispatch (Vtables):** Under the hood, an interface value is a fat pointer: one pointer to the concrete struct's data, and one pointer to a generated Virtual Method Table (vtable) containing function pointers for dispatch.

## Lease Rules

*   When a value is assigned to an interface, ownership semantics remain standard. If you assign an owned struct (`@User`) to an owned interface (`@Greeter`), the interface value takes ownership. 
*   Dropping the owned interface will automatically dynamically dispatch to the concrete type's implicit `drop()` method, ensuring safe resource destruction regardless of the hidden type.

## Examples

### Interface Embedding
Nora supports interface embedding, where one interface inherits the method signatures of another.

```nora
pub type Reader = interface {
    fn Read(buf: &Vector[u8]) i32
}

pub type Writer = interface {
    fn Write(buf: #Vector[u8]) i32
}

pub type ReadWriter = interface {
    Reader
    Writer
}
```

## Errors & Diagnostics

*   **Missing Method:** Attempting to assign a struct to an interface without having all required methods implemented throws a clear compile-time error listing the missing signatures.
*   **Signature Mismatch:** If a struct implements a method but with incorrect argument types or return type compared to the interface, it will not satisfy the protocol.
