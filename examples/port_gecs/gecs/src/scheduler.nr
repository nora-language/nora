package gecs

import "collections"
import "sync"

pub fn NewScheduler() @Scheduler {
    return alloc Scheduler {
        systems: collections.NewVector[System](16),
        batches: collections.NewVector[@collections.Vector[#System]](8)
    }
}

pub fn (s: &Scheduler) AddSystem(sys: System) {
    s.systems.Push[System](sys)
    s.rebuild()
}

pub fn (s: &Scheduler) rebuild() {
    import "io"
    io.PrintLn("      [rebuild] starting...")
    s.batches.size = 0

    io.PrintLn("      [rebuild] remaining vector allocating...")
    var remaining = collections.NewVector[#System](s.systems.size)
    for i in 0..s.systems.size {
        var sys_val: #System = #s.systems.data[i]
        remaining.Push[#System](sys_val)
    }

    io.PrintLn("      [rebuild] entering while loop...")
    while (remaining.size > 0) {
        var batch = collections.NewVector[#System](8)
        var conflicts = NewBitSet()
        var reads = NewBitSet()

        var next_remaining = collections.NewVector[#System](remaining.size)

        for i in 0..remaining.size {
            var sys = remaining.data[i]

            var has_conflict = false
            io.PrintLn("      [rebuild] calling sys.GetAccess()...")
            var sys_access = sys.GetAccess()
            io.PrintLn("      [rebuild] sys.GetAccess() succeeded.")

            var iter_write = sys_access.writes.capacity * 64
            for id in 0..iter_write {
                if (sys_access.writes.Has(id)) {
                    if (conflicts.Has(id) || reads.Has(id)) {
                        has_conflict = true
                    }
                }
            }

            if (!has_conflict) {
                var iter_read = sys_access.reads.capacity * 64
                for id in 0..iter_read {
                    if (sys_access.reads.Has(id)) {
                        if (conflicts.Has(id)) {
                            has_conflict = true
                        }
                    }
                }
            }

            if (!has_conflict) {
                batch.Push[#System](sys)

                var cap_write = sys_access.writes.capacity * 64
                for id in 0..cap_write {
                    if (sys_access.writes.Has(id)) {
                        conflicts.Set(id)
                    }
                }
                var cap_read = sys_access.reads.capacity * 64
                for id in 0..cap_read {
                    if (sys_access.reads.Has(id)) {
                        reads.Set(id)
                    }
                }
            } else {
                next_remaining.Push[#System](sys)
            }
        }

        s.batches.Push[@collections.Vector[#System]](@batch)
        remaining = @next_remaining
    }
}

pub fn (s: &Scheduler) Run(w: &World) {
    import "io"
    import "strconv"
    for b in 0..s.batches.size {
        var batch: &collections.Vector[#System] = s.batches.data[b]

        var wg = sync.NewWaitGroup()
        var cb_chans = alloc chan[((@CommandBuffer))][batch.size]

        for i in 0..batch.size {
            cb_chans[i] = make(chan[((@CommandBuffer))], 1)
        }

        for i in 0..batch.size {
            wg.Add(1)
            var cb = NewCommandBuffer()
            var sys = batch.data[i]
            spawn run_system_worker(sys, #w, @cb, #wg, cb_chans[i])
        }

        wg.Wait()
        io.PrintLn("   [Scheduler] Wait finished.")
        for i in 0..batch.size {
            io.PrintLn("   [Scheduler] Playback loop i = " + io.FormatInt(i))
            var cb = <-cb_chans[i]
            io.PrintLn("   [Scheduler] Playback calling...")
            cb.Playback(w)
            io.PrintLn("   [Scheduler] Playback done.")
        }
    }
    io.PrintLn("   [Scheduler] Run exiting...")
}

fn run_system_worker(sys: #System, w: #World, cb: @CommandBuffer, wg: #sync.WaitGroup, cb_chan: #chan[@CommandBuffer]) {
    io.PrintLn("   [Worker] Starting system update...")
    sys.Update(w, &cb)
    io.PrintLn("   [Worker] System update finished.")
    cb_chan <- @cb
    io.PrintLn("   [Worker] WaitGroup Done calling...")
    wg.Done()
}
