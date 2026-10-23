// Gap 2.1a: cloneAndSubstituteHelper (pkg/semantic/analyzer.go:6966) omits *ast.TryExpression (?),
// so generic functions using `?` fail to clone and substitute the TryExpression node across monomorphizations!
package main

import "os"

fn unwrap_or_err[T](r: Result[T, str]) Result[T, str] {
    var v = r?
    return Ok[T, str](v)
}

fn main() i32 {
    var r1 = unwrap_or_err[i32](Ok[i32, str](123))
    match r1 {
        Ok(v) => {
            if v != 123 {
                os.Exit(1)
            }
        }
        Err(e) => {
            os.Exit(1)
        }
    }
    return 0
}
