# Existential `any` Type

## Title & Overview
This specification details the semantics and implementation of the existential **`any`** type in the Nora Programming Language. The `any` type is a universal, type-erased interface that allows heterogeneous data structures, dynamic downcasting, and dynamic polymorphism while maintaining Nora's strict memory safety and RAII rules.

## Motivation
In strict statically-typed languages, creating heterogeneous collections (e.g., `List` containing different component types for an Entity Component System) or returning varying types from a single function requires extensive workarounds, like bulky sum types. The `any` type provides a compiler-supported existential container—an empty protocol (interface)—capable of wrapping any valid Nora struct or value at runtime, allowing safe, dynamic type dispatch.

## Syntax
The `any` keyword is a built-in alias for the empty interface (`interface {}`).

### Wrapping values into `any`
Variables of any concrete struct or interface can be coerced into an `any` type:
```nora
var u = User { name: "Alice" }

// Borrowing a value as 'any'
var a1: any = #u

// Moving a value into 'any'
var a2: any = @u
```

### Retrieving concrete values (Downcasting)
To unwrap the underlying value, use function-style casting:
```nora
var a: any = #user
var u_back: #User = User(a)
```

## Semantics

### Runtime Representation
In Nora, `any` is implemented as an existential interface. At the ABI level (and in the generated C11 code), an `any` is represented as a "fat pointer" containing two fields:
1. `data`: A `void*` pointing to the underlying data payload.
2. `vtable`: A pointer to the Virtual Method Table associated with the original concrete type, enabling safe dynamic dispatch and RAII drops.

### Interface-to-Interface Casting
Nora supports casting existing interface values into `any`, stripping their specialized vtables down to the base universal vtable:
```nora
var g: Greeter = #u
var a: any = g // Safe interface-to-interface coercion
```

### Heterogeneous Lists
The `any` type is commonly used as a generic parameter for standard library collections to enable heterogeneous storage:
```nora
var list = make(List[any])
list = append(list, #user)
list = append(list, #robot)
```

## Type Rules
1. **Concrete to `any`**: Any value type (Structs, Primitives) can be implicitly cast to `any`.
2. **Interface to `any`**: Any interface type can be implicitly cast to `any`.
3. **`any` to Concrete (Downcasting)**: Casting from `any` back to a concrete type requires explicit casting (`ConcreteType(any_val)`). The compiler inserts runtime type checks to ensure the `vtable` matches the requested concrete type.

## Lease Rules
Because `any` is an interface fat-pointer, it seamlessly integrates with Nora's topological lease solver:

* **Borrowed `any` (`#any`)**: A read-only existential type. It cannot be mutated, and the underlying data's lifetime is guaranteed to outlive the borrow.
* **Owned `any` (`@any` or `any`)**: An owned existential type. If the underlying data requires RAII cleanup, the compiler will automatically invoke the destructor registered in the `any`'s `vtable` when the `any` value goes out of scope or is overwritten.

```nora
fn process_any(val: any) {
    // val is an owned 'any'. 
    // If it contains a type that needs dropping, the vtable drop will be invoked here.
}
```

## Examples

### ECS Component Storage
The `any` type is heavily utilized in Entity Component Systems to store heterogeneous components in a single map or array:
```nora
package ecs

pub type ComponentStorage = struct {
    components: collections.HashMap[i32, any] // Entity ID -> any Component
}

pub fn (s: &ComponentStorage) add_component(id: i32, comp: any) {
    s.components.Set(id, @comp)
}

pub fn (s: &ComponentStorage) get_position(id: i32) #Position {
    var comp: any = s.components.Get(id).Unwrap()
    return Position(comp) // Downcast back to Position
}
```

## Edge Cases
* **Downcast Panics**: If an `any` value is explicitly downcast to an incorrect concrete type (e.g. attempting to cast an `any` wrapping a `Robot` into a `User`), the program will safely panic at runtime rather than corrupting memory.
* **Memory Decay on Wrapper Reassignment**: Because `any` abstracts the underlying value size, you must be careful when passing stack-allocated generic arrays explicitly by-move into `any` parameters. The Nora compiler leverages array compound literals to maintain stack lifetimes when passing moved values into `any`.

## Errors & Diagnostics
* **"Downcast failed"**: A runtime panic generated when casting `any` to the wrong concrete type.
* **"Cannot move out of borrowed existential"**: A compile-time semantic error if attempting to cast a borrowed `#any` into an owned concrete `@Type` value.

## Pattern Matching (Dynamic Type Dispatch)
You can safely downcast and branch on an `any` value at runtime using the `match` statement. This functions identically to pattern-matching on a sum type, allowing you to extract the concrete type bound to a variable.

Because `any` is an open existential type, the compiler cannot guarantee that all possible cases are covered. Therefore, **a wildcard `_ =>` fallback branch is strictly required** when matching on `any`.

```nora
fn process_any(val: any) {
    match val {
        User(u) => {
            io.PrintLn("Matched User: " + u.name)
        }
        Robot(r) => {
            io.PrintLn("Matched Robot")
        }
        _ => {
            io.PrintLn("Matched unknown type")
        }
    }
}
```

The matched variables (`u` and `r`) correctly inherit the lease rules applied to the `match` target (e.g., if the target is `#any`, `u` will be a `#User`).
