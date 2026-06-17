package gecs

import "collections"

pub fn NewEventQueue() @EventQueue {
    return alloc EventQueue {
        events: collections.NewVector[any](16)
    }
}

pub fn NewEventManager() @EventManager {
    return alloc EventManager {
        queues: collections.NewHashMap[i32, @EventQueue](16)
    }
}

pub fn (em: &EventManager) Emit[T](cid: i32, event: T) {
    var queue: &@EventQueue = none
    var queue_opt = em.queues.GetMut[i32, @EventQueue](#cid)
    match queue_opt {
        Some(p) => {
            queue = p
        }
        None => {
            var q = NewEventQueue()
            var _inserted = em.queues.Insert[i32, @EventQueue](cid + 0, q)
            var q_opt = em.queues.GetMut[i32, @EventQueue](#cid)
            match q_opt {
                Some(p) => {
                    queue = p
                }
                None => {}
            }
        }
    }
    var val_any: any = @event
    queue.events.Push[any](val_any)
}

pub fn (em: &EventManager) ReadEvents(cid: i32) Option[#EventQueue] {
    var queue_opt = em.queues.Get[i32, @EventQueue](#cid)
    match queue_opt {
        Some(p) => {
            return Some[#EventQueue](p)
        }
        None => {
            return None[#EventQueue]
        }
    }
}

pub fn (em: &EventManager) Clear() {
    em.queues = collections.NewHashMap[i32, @EventQueue](16)
}
