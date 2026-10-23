// Gap 4.2a: Vector.Clear (std/collections/vector.nr:130) sets v.size = 0 without dropping
// owned elements, and Vector.Set (line 81) overwrites v.data[index] without dropping the previous element.
package main

import "collections"

fn main() i32 {
    var v = collections.NewVector[str](4)
    var p = "heap_"
    var s1 = p + "one"
    var s2 = p + "two"
    var s3 = p + "three"
    v.Push[str](@s1)
    v.Push[str](@s2)
    v.Set[str](0, @s3)
    v.Clear[str]()
    return 0
}
