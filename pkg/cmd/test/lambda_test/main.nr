package main

import "std/io"

type User = struct {
    name: str,
    age: i32
}

type Query = struct {
    value: i32
}

fn (self: @Query) Map(op: fn(i32) i32) @Query {
    return Query {
        value: op(self.value)
    }
}

fn (self: @Query) Filter(op: fn(i32) bool) @Query {
    if (op(self.value)) {
        return self
    } else {
        return Query {
            value: 0
        }
    }
}

// Scenario 1: Function returning a closure
fn make_adder(x: i32) fn(i32) i32 {
    return fn(y: i32) i32 {
        return x + y;
    };
}

// Scenario 2: Function taking a closure as argument
fn run_op(a: i32, b: i32, op: fn(i32, i32) i32) i32 {
    return op(a, b);
}

// Scenario 3: Spawning a fiber to execute a closure concurrently
fn concurrent_worker(op: fn() i32, done: #chan[i32]) {
    var res = op();
    done <- res
}

fn main() i32 {
    io.PrintLn("--- Starting Lambda and Closure Integration Tests ---")

    // Test 1: Basic Lambda with no captures
    var sq = fn(x: i32) i32 {
        return x * x;
    };
    var sq_res = sq(6);
    io.PrintLn("1. Basic Lambda sq(6) = ${sq_res}"); // Expected: 36
    if (sq_res != 36) {
        io.PrintLn("Test 1 Failed!")
        return 1;
    }

    // Test 2: Passing lambda to higher order function
    var sum = fn(x: i32, y: i32) i32 {
        return x + y;
    };
    var sum_res = run_op(10, 20, sum);
    io.PrintLn("2. Higher order run_op(10, 20, sum) = ${sum_res}"); // Expected: 30
    if (sum_res != 30) {
        io.PrintLn("Test 2 Failed!")
        return 1;
    }

    // Test 3: Value Capture (capturing primitive variable x by copy/read)
    var x = 10;
    var addX = fn(y: i32) i32 {
        return x + y;
    };
    var addX_res = addX(5);
    io.PrintLn("3. Value Capture addX(5) = ${addX_res}"); // Expected: 15
    if (addX_res != 15) {
        io.PrintLn("Test 3 Failed!")
        return 1;
    }

    // Test 4: Returning a closure from another function (Higher order maker)
    var adder5 = make_adder(5);
    var adder_res = adder5(12);
    io.PrintLn("4. Returning Closure adder5(12) = ${adder_res}"); // Expected: 17
    if (adder_res != 17) {
        io.PrintLn("Test 4 Failed!")
        return 1;
    }

    // Test 5: Closure Capturing Owned Types (str)
    var greeting = "Hello, World!";
    var printGreeting = fn() {
        io.PrintLn("5. Owned Type Capture: " + greeting);
    };
    printGreeting();

    // Test 6: Nested closures/lambdas (closure returning another closure)
    var a = 2;
    var outer_lambda = fn(b: i32) fn(i32) i32 {
        return fn(c: i32) i32 {
            return a + b + c;
        };
    };
    var inner_lambda = outer_lambda(3);
    var nested_res = inner_lambda(4);
    io.PrintLn("6. Nested Closures nested_res = ${nested_res}"); // Expected: 9 (2+3+4)
    if (nested_res != 9) {
        io.PrintLn("Test 6 Failed!")
        return 1;
    }

    // Test 7: Branching Logic inside closure
    var is_premium = true;
    var get_price = fn(base: i32) i32 {
        if (is_premium) {
            return base - 5;
        } else {
            return base;
        }
    };
    var price_res = get_price(20);
    io.PrintLn("7. Branching inside closure get_price(20) = ${price_res}"); // Expected: 15
    if (price_res != 15) {
        io.PrintLn("Test 7 Failed!")
        return 1;
    }

    // Test 8: Loop inside closure
    var limit = 5;
    var sum_to_limit = fn() i32 {
        var sum_val = 0;
        var i = 1;
        while (i <= limit) {
            sum_val = sum_val + i;
            i = i + 1;
        }
        return sum_val;
    };
    var loop_res = sum_to_limit();
    io.PrintLn("8. Loop inside closure sum_to_limit() = ${loop_res}"); // Expected: 15
    if (loop_res != 15) {
        io.PrintLn("Test 8 Failed!")
        return 1;
    }

    // Test 9: User-defined struct captured by closure
    var u = User {
        name: "Alice",
        age: 30
    };
    var printUser = fn() {
        io.PrintLn("9. User-defined Type Capture: name=${u.name}, age=${u.age}");
    };
    printUser();

    // Test 10: Chaining methods with lambda parameters
    var q = Query {
        value: 10
    };
    var chain_res = q.Map(fn(val: i32) i32 {
        return val * 2;
    }).Filter(fn(val: i32) bool {
        return val > 15;
    });
    io.PrintLn("10. Chaining methods with lambdas: value=${chain_res.value}"); // Expected: 20
    if (chain_res.value != 20) {
        io.PrintLn("Test 10 Failed!")
        return 1;
    }

    var q2 = Query {
        value: 5
    };
    var chain_res2 = q2.Map(fn(val: i32) i32 {
        return val * 2;
    }).Filter(fn(val: i32) bool {
        return val > 15;
    });
    io.PrintLn("10b. Chaining methods filter false: value=${chain_res2.value}"); // Expected: 0
    if (chain_res2.value != 0) {
        io.PrintLn("Test 10b Failed!")
        return 1;
    }

    // Test 11: Concurrent Closure Execution (spawn fiber with captured context)
    var factor = 3;
    var task = fn() i32 {
        var local_val = 14;
        return local_val * factor; // Expected: 42 (14 * 3)
    };

    var done_ch = make(chan[i32], 1)
    scope {
        spawn concurrent_worker(task, #done_ch)

        // Wait for the result from the fiber
        var result_val = <-done_ch
        io.PrintLn("11. Concurrent closure result = ${result_val}"); // Expected: 42
        if (result_val != 42) {
            io.PrintLn("Test 11 Failed!")
            return 1;
        }

        io.PrintLn("--- All Lambda and Closure Integration Tests Passed successfully! ---")
        return 0;

    }
}
