package gecs

import "collections"
import "io"

pub fn NewQuery1[A](w: &World, cidA: i32) @Query1[A] {
    var mask = NewBitSet()
    mask.Set(cidA)
    return alloc Query1[A] {
        world: w,
        mask: mask,
        excl: NewBitSet(),
        cidA: cidA
    }
}

pub fn (q: &Query1[A]) Without[A](cid: i32) {
    q.excl.Set(cid)
}

pub fn (q: &Query1[A]) With[A](cid: i32) {
    q.mask.Set(cid)
}

pub fn (q: &Query1[A]) Each[A](f: fn(Entity, &A)) {
    var iter = q.world.archetypes.Iter[i64, @Archetype]()
    var entry_opt = iter.Next[i64, @Archetype]()

    var loop = true
    while (loop) {
        match entry_opt {
            Some(entry) => {
                var arch: &Archetype = entry.value
                if (arch.mask.Contains(q.mask) && !arch.mask.Intersects(q.excl)) {
                    var ents = arch.GetEntities()

                    var colA_opt = arch.GetColumn(q.cidA)

                    match colA_opt {
                        Some(colA_any) => {
                            var vecA: &collections.Vector[A] = collections.Vector[A](colA_any)
                            var len = ents.Len[Entity]()
                            for i in 0..len {
                                var e_ref = ents.Get[Entity](i)
                                var e = Entity { id: e_ref.id }
                                var refA = vecA.GetMut[A](i)
                                f(e, refA)
                            }
                        }
                        None => {}
                    }
                }
                entry_opt = iter.Next[i64, @Archetype]()
            }
            None => {
                loop = false
            }
        }
    }
}

pub fn NewQuery2[A, B](w: &World, cidA: i32, cidB: i32) @Query2[A, B] {
    var mask = NewBitSet()
    mask.Set(cidA)
    mask.Set(cidB)
    return alloc Query2[A, B] {
        world: w,
        mask: mask,
        excl: NewBitSet(),
        cidA: cidA,
        cidB: cidB
    }
}

pub fn (q: &Query2[A, B]) Without[A, B](cid: i32) {
    q.excl.Set(cid)
}

pub fn (q: &Query2[A, B]) With[A, B](cid: i32) {
    q.mask.Set(cid)
}

pub fn (q: &Query2[A, B]) Each[A, B](f: fn(Entity, &A, &B)) {
    var iter = q.world.archetypes.Iter[i64, @Archetype]()
    var entry_opt = iter.Next[i64, @Archetype]()

    var loop = true
    while (loop) {
        match entry_opt {
            Some(entry) => {
                var arch: &Archetype = entry.value
                if (arch.mask.Contains(q.mask) && !arch.mask.Intersects(q.excl)) {
                    var ents = arch.GetEntities()

                    var colA_opt = arch.GetColumn(q.cidA)
                    var colB_opt = arch.GetColumn(q.cidB)

                    match colA_opt {
                        Some(colA_any) => {
                            match colB_opt {
                                Some(colB_any) => {
                                    var vecA: &collections.Vector[A] = collections.Vector[A](colA_any)
                                    var vecB: &collections.Vector[B] = collections.Vector[B](colB_any)
                                    
                                    var len = ents.Len[Entity]()
                                    for i in 0..len {
                                        var e_ref = ents.Get[Entity](i)
                                        var e = Entity { id: e_ref.id }
                                        var refA = vecA.GetMut[A](i)
                                        var refB = vecB.GetMut[B](i)
                                        f(e, refA, refB)
                                    }
                                }
                                None => {}
                            }
                        }
                        None => {}
                    }
                }
                entry_opt = iter.Next[i64, @Archetype]()
            }
            None => {
                loop = false
            }
        }
    }
}
