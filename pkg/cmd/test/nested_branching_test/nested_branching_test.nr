package main

import "io"
import "random"

type Resource = struct {
    id: i32,
    name: str
}

fn (self: #Resource) drop() {
    io.PrintLn("  [DROP HOOK] Dropping Resource ${self.id}: ${self.name}")
}

type Mode = enum {
    Alpha,
    Beta(res: @Resource),
    Gamma(val: i32)
}

// Up to 4 levels of nested branching if-else
fn test_deep_nested_branches(choice1: i32, choice2: i32, choice3: i32, choice4: i32) {
    io.PrintLn("--- test_deep_nested_branches(c1=${choice1}, c2=${choice2}, c3=${choice3}, c4=${choice4}) ---")
    var r1 = alloc Resource {
        id: 101,
        name: "OuterResource"
    }

    if (choice1 > 50) {
        var r2 = alloc Resource {
            id: 102,
            name: "Branch1_True"
        }
        if (choice2 > 50) {
            var r3 = alloc Resource {
                id: 103,
                name: "Branch2_True"
            }
            if (choice3 > 50) {
                var r4 = alloc Resource {
                    id: 104,
                    name: "Branch3_True"
                }
                if (choice4 > 50) {
                    var r5 = alloc Resource {
                        id: 105,
                        name: "Branch4_True"
                    }
                    io.PrintLn("    Path 1.1.1.1: ${r1.name}, ${r2.name}, ${r3.name}, ${r4.name}, ${r5.name} active")
                } else {
                    var r6 = alloc Resource {
                        id: 106,
                        name: "Branch4_False"
                    }
                    io.PrintLn("    Path 1.1.1.2: ${r1.name}, ${r2.name}, ${r3.name}, ${r4.name}, ${r6.name} active")
                }
            } else {
                io.PrintLn("    Path 1.1.2: ${r1.name}, ${r2.name}, ${r3.name} active")
            }
        } else {
            io.PrintLn("    Path 1.2: ${r1.name}, ${r2.name} active")
        }
    } else {
        var r7 = alloc Resource {
            id: 107,
            name: "Branch1_False"
        }
        if (choice2 <= 50) {
            var r8 = alloc Resource {
                id: 108,
                name: "Branch2_False"
            }
            if (choice3 <= 50) {
                var r9 = alloc Resource {
                    id: 109,
                    name: "Branch3_False"
                }
                if (choice4 <= 50) {
                    var r10 = alloc Resource {
                        id: 110,
                        name: "Branch4_False_End"
                    }
                    io.PrintLn("    Path 2.2.2.2: ${r1.name}, ${r7.name}, ${r8.name}, ${r9.name}, ${r10.name} active")
                } else {
                    io.PrintLn("    Path 2.2.2.1: ${r1.name}, ${r7.name}, ${r8.name}, ${r9.name} active")
                }
            }
        }
    }
}

fn test_match_sum_types(choice: i32) {
    io.PrintLn("--- test_match_sum_types(choice=${choice}) ---")
    var m = Alpha
    if (choice == 0) {
        m = Alpha
    } else if (choice == 1) {
        var r = alloc Resource {
            id: 201,
            name: "BetaResource"
        }
        m = Beta(@r)
    } else {
        m = Gamma(42)
    }

    match @m {
        Alpha => {
            io.PrintLn("    Matched Alpha")
        }
        Beta(res) => {
            var id = res.id
            io.PrintLn("    Matched Beta with resource ID ${id}")
        }
        Gamma(val) => {
            io.PrintLn("    Matched Gamma with val ${val}")
        }
    }
}

fn test_loops_early_exits(iterations: i32) i32 {
    io.PrintLn("--- test_loops_early_exits(iterations=${iterations}) ---")
    var r_loop = alloc Resource {
        id: 301,
        name: "LoopStart"
    }

    var i = 0
    while (i < iterations) {
        var r_iter = alloc Resource {
            id: 400 + i,
            name: "IterResource"
        }

        // Generate pseudo-random value inside the loop using std/random
        var rv = random.IntRange(0, 100)
        io.PrintLn("    Loop iter ${i}: random value = ${rv}")

        if (rv < 20) {
            io.PrintLn("    Early return at rv < 20 (Resource: ${r_loop.name}, Iter: ${r_iter.name})")
            return 99
        }

        if (rv >= 80) {
            io.PrintLn("    Break at rv >= 80 (Resource: ${r_loop.name}, Iter: ${r_iter.name})")
            break;
        }

        i = i + 1
    }

    io.PrintLn("    Exited loop normally (Resource: ${r_loop.name})")
    return 0
}

fn select_producer(ch: #chan[i32]) {
    ch <- 42
}

fn test_select_branching() {
    io.PrintLn("--- test_select_branching() ---")
    var ch = make(chan[i32])
    scope {
        spawn select_producer(#ch)

        var v = 0
        select {
            case v = <-ch:
                {
                    io.PrintLn("    Received ${v} in select case")
                    var r_sel = alloc Resource {
                        id: 501,
                        name: "SelectResource"
                    }
                    if (v == 42) {
                        var r_inner = alloc Resource {
                            id: 502,
                            name: "SelectInner_True"
                        }
                        io.PrintLn("      Nested branching under select case: ${r_sel.name}, ${r_inner.name} active")
                    } else {
                        var r_inner2 = alloc Resource {
                            id: 503,
                            name: "SelectInner_False"
                        }
                        io.PrintLn("      Nested branching under select case: ${r_sel.name}, ${r_inner2.name} active")
                    }
                }
        }

    }
}

fn main() i32 {
    io.PrintLn("=== STARTING DEEP NESTED BRANCHING & RANDOM STRESS TEST ===")

    // Seed the global random generator
    random.Seed(987654321)

    // Demonstrate standard random functionality
    var r_int = random.Int()
    var r_range = random.IntRange(10, 100)
    var r_float = random.Float()
    var r_frange = random.FloatRange(1.5, 5.5)

    io.PrintLn("Standard Random values:")
    io.PrintLn("  Int: ${r_int}")
    io.PrintLn("  IntRange(10, 100): ${r_range}")
    io.PrintLn("  Float: ${r_float}")
    io.PrintLn("  FloatRange(1.5, 5.5): ${r_frange}")

    io.PrintLn("\n=== Starting Dynamic Random Stress Loop ===")
    var i = 0
    while i < 10 {
        io.PrintLn("\n--- Loop Iteration ${i} ---")

        // Generate pseudo-random choices
        var c1 = random.IntRange(0, 100)
        var c2 = random.IntRange(0, 100)
        var c3 = random.IntRange(0, 100)
        var c4 = random.IntRange(0, 100)

        // Run deep nested branching if-else tests with dynamic values
        test_deep_nested_branches(c1, c2, c3, c4)

        // Run match sum type tests with dynamic values (0, 1, or 2)
        var match_choice = random.IntRange(0, 3)
        test_match_sum_types(match_choice)

        // Run loops and early exits tests with dynamic iterations (1 to 5)
        var loop_iterations = random.IntRange(1, 6)
        var ret = test_loops_early_exits(loop_iterations)
        io.PrintLn("  Return value: ${ret}")
        i++
    }

    io.PrintLn("\n=== Starting Concurrent Select Test ===")
    // Run select branching tests
    test_select_branching()

    io.PrintLn("\n=== STRESS TEST COMPLETED SUCCESSFULLY ===")
    return 0
}
