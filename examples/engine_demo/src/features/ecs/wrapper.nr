package ecs

import "gecs"

type EcsWrapper = struct {
    world: gecs.World,
    scheduler: gecs.Scheduler,
}

pub fn NewEcsWrapper() EcsWrapper {
    return EcsWrapper {
        world: gecs.NewWorld(),
        scheduler: gecs.NewScheduler(),
    }
}

pub fn (self: &EcsWrapper) Step() {
    self.scheduler.Run(self.world)
}

pub fn (self: &EcsWrapper) RegisterSystem(sys: gecs.System) {
    self.scheduler.AddSystem(sys)
}

pub fn CloneEntity(e: #gecs.Entity) gecs.Entity {
    var id_val = e.id
    return gecs.Entity {
        id: id_val
    }
}
