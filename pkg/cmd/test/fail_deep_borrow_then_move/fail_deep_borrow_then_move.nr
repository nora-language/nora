// fail_deep_borrow_then_move
//
// BUG: solver.go getAliasOrigin() only resolves ONE level of selector for
// borrow expressions. For `#obj.inner.value` (two levels deep), the code is:
//
//   if sel, ok := pref.Right.(*ast.SelectorExpression); ok {
//       if id, ok := sel.Left.(*ast.Identifier); ok {  // ← only works for #obj.field
//           return s.SemanticInfo.Uses[id]
//       }
//   }
//
// For `#obj.inner.value`, `sel.Left` is `obj.inner` (SelectorExpression),
// not an Identifier. So the alias origin is returned as nil, meaning the
// variable `b` is NOT tracked as an alias of `obj`.
//
// Consequence: `@obj` after `#obj.inner.value` does NOT trigger the
// "cannot move out of 'obj' while borrowed by 'b'" check.
//
// Expected: compilation error (moving `obj` while a deep borrow `b` is live).
// Bug present: compiles silently, producing a dangling borrow at runtime.

package main

type Inner = struct {
    value: str
}

type Outer = struct {
    inner: Inner
}

fn use_str(s: #str) {}

fn main() i32 {
    var obj = Outer { inner: Inner { value: "hello" } }

    var b = #obj.inner.value

    // This should FAIL: moving `obj` while `b` (a borrow of obj.inner.value)
    // is still in scope AND used later.
    var moved = @obj

    use_str(b)

    return 0
}
