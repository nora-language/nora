package main

import "collections"

pub type IslandBatch[T] = struct {
    pairs: collections.Vector[T]
}

pub fn main() {
    var _new_island = IslandBatch[f64] {
        pairs: collections.NewVector[f64](4)
    }
}
