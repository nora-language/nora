package gecs

import "collections"
import "io"

pub fn SetParent(w: &World, child: Entity, parent: Entity) {
    io.PrintLn("DEBUG SetParent step 1")
    // 1. If child already has a parent, remove from old parent's children list
    var has_old = Has(w,         // CID 3 for Parent
        Entity {
            id: child.id
        }, 3)
    if (has_old) {
        io.PrintLn("DEBUG SetParent step 1.1")
        var old_parent_val = Get[Parent](w, Entity {
                id: child.id
            }, 3)

        var has_children = Has(w,             // CID 4 for Children
            Entity {
                id: old_parent_val.id.id
            }, 4)
        if (has_children) {
            var children_mut: &Children = GetMut[Children](w, Entity {
                    id: old_parent_val.id.id
                }, 4)

            // Remove child from children_val vector
            var new_ents = collections.NewVector[Entity](children_mut.entities.size)
            for i in 0..children_mut.entities.size {
                var c_ent = children_mut.entities.data[i]
                if (c_ent.id != child.id) {
                    new_ents.Push[Entity](Entity {
                            id: c_ent.id
                        })
                }
            }
            children_mut.entities = @new_ents
        }
    }

    io.PrintLn("DEBUG SetParent step 2")
    // 2. Set new parent
    Add[Parent](w, Entity {
            id: child.id
        }, 3, Parent {
            id: Entity {
                id: parent.id
            }
        })

    io.PrintLn("DEBUG SetParent step 3")
    // 3. Update parent's children list
    var has_new_children = Has(w, Entity {
            id: parent.id
        }, 4)
    if (!has_new_children) {
        io.PrintLn("DEBUG SetParent step 3.1")
        var new_vector = collections.NewVector[Entity](4)
        new_vector.Push[Entity](Entity {
                id: child.id
            })
        Add[Children](w, Entity {
                id: parent.id
            }, 4, Children {
                entities: new_vector
            })
    } else {
        io.PrintLn("DEBUG SetParent step 3.2")
        var children_mut: &Children = GetMut[Children](w, Entity {
                id: parent.id
            }, 4)
        children_mut.entities.Push[Entity](Entity {
                id: child.id
            })
    }
    io.PrintLn("DEBUG SetParent done")
}

pub fn RemoveParent(w: &World, child: Entity) {
    var has_parent = Has(w, Entity {
            id: child.id
        }, 3)
    if (has_parent) {
        var old_parent_val = Get[Parent](w, Entity {
                id: child.id
            }, 3)

        var has_children = Has(w, Entity {
                id: old_parent_val.id.id
            }, 4)
        if (has_children) {
            var children_mut: &Children = GetMut[Children](w, Entity {
                    id: old_parent_val.id.id
                }, 4)

            var new_ents = collections.NewVector[Entity](children_mut.entities.size)
            for i in 0..children_mut.entities.size {
                var c_ent = children_mut.entities.data[i]
                if (c_ent.id != child.id) {
                    new_ents.Push[Entity](Entity {
                            id: c_ent.id
                        })
                }
            }
            children_mut.entities = @new_ents
        }
        Remove[Parent](w, Entity {
                id: child.id
            }, 3)
    }
}
