package core

type Engine = struct {
    running: i32,
}

pub fn NewEngine() Engine {
    return Engine {
        running: 1
    }
}

pub fn (self: &Engine) Quit() {
    self.running = 0
}

pub fn (self: #Engine) IsRunning() i32 {
    return self.running
}
