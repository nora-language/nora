// Gap 4.2c: HashMap.Remove (std/collections/map.nr:204-210) moves @m.entries[entry_idx].value out
// without clearing the slot, AND when m is dropped at scope exit, m.entries's auto-drop drops all
// capacity slots (even where occupied == false), freeing the removed value while caller still holds it!
package main

import "collections"
import "io"
import "os"

var g_val_drops: i32 = 0

type ValObj = struct {
    v: i32
}

fn (o: #ValObj) drop() {
    g_val_drops = g_val_drops + 1
}

fn main() i32 {
    var m = collections.NewHashMap[i32, ValObj](4)
    m.Insert[i32, ValObj](1, ValObj { v: 100 })
    var k = 1
    var rem = m.Remove[i32, ValObj](#k)
    match rem {
        Some(obj) => {
            if obj.v != 100 {
                os.Exit(1)
            }
        }
        None => {
            os.Exit(1)
        }
    }
    // At this point, rem was consumed and dropped once, and m is still alive.
    // Yet when m is dropped (or if we check drops after scope), how many times was ValObj dropped?
    if g_val_drops != 1 {
        io.PrintLn("FAIL: expected ValObj to be dropped once, dropped ${g_val_drops} times")
        os.Exit(1)
    }
    return 0
}
