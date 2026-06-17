package core

type Time = struct {
    delta_time: f32,
    time_scale: f32,
    elapsed: f32,
}

pub fn NewTime() Time {
    return Time {
        delta_time: f32(0.016), // default to 60fps
        time_scale: f32(1.0),
        elapsed: f32(0.0)
    }
}

pub fn (self: &Time) Tick(dt: f32) {
    self.delta_time = dt * self.time_scale
    self.elapsed = self.elapsed + self.delta_time
}
