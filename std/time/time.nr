package time

extern fn nr_time_now() f64

extern fn nr_sleep_ms(ms: i32)

extern fn nr_time_format(sec: f64, format_str: #str) str



pub type DateParts = struct {
    year: i32,
    month: i32,
    day: i32,
    hour: i32,
    minute: i32,
    second: i32,
    weekday: i32,
    yearday: i32,
    is_dst: bool
}

extern fn nr_time_to_parts(sec: f64, utc: bool, parts: ptr)

pub type Duration = struct {
    nanoseconds: i64
}

// --- Duration Constructors ---
pub fn Nanosecond() Duration {
    return Duration {
        nanoseconds: 1
    }
}

pub fn Microsecond() Duration {
    return Duration {
        nanoseconds: 1000
    }
}

pub fn Millisecond() Duration {
    return Duration {
        nanoseconds: 1000000
    }
}

pub fn Second() Duration {
    return Duration {
        nanoseconds: 1000000000
    }
}

pub fn Minute() Duration {
    return Duration {
        nanoseconds: 60000000000
    }
}

pub fn Hour() Duration {
    return Duration {
        nanoseconds: 3600000000000
    }
}

pub fn (self: #Duration) Nanoseconds() i64 {
    return self.nanoseconds
}

pub fn (self: #Duration) Microseconds() i64 {
    return self.nanoseconds / 1000
}

pub fn (self: #Duration) Milliseconds() i64 {
    return self.nanoseconds / 1000000
}

pub fn (self: #Duration) Seconds() f64 {
    return f64(self.nanoseconds) / 1000000000.0
}

pub type Time = struct {
    unix_sec: i64,
    unix_nsec: i32
}

// Now returns the current high-precision timestamp
pub fn Now() Time {
    var raw = nr_time_now()
    var sec_i64 = i64(raw)
    var sec_f64 = f64(sec_i64)
    var frac = raw - sec_f64
    var nsec = i32(frac * 1000000000.0)
    return Time {
        unix_sec: sec_i64,
        unix_nsec: nsec
    }
}

pub fn Unix(sec: i64, nsec: i32) Time {
    return Time {
        unix_sec: sec,
        unix_nsec: nsec
    }
}

pub fn (self: #Time) Unix() i64 {
    return self.unix_sec
}

pub fn (self: #Time) UnixNano() i64 {
    return (self.unix_sec * 1000000000) + i64(self.unix_nsec)
}

pub fn (self: #Time) Format(format_str: #str) str {
    var sec_double = f64(self.unix_sec) + f64(self.unix_nsec) / 1000000000.0
    return nr_time_format(sec_double, format_str)
}

pub fn (self: #Time) UTC() DateParts {
    var parts = DateParts {
        year: 0,
        month: 0,
        day: 0,
        hour: 0,
        minute: 0,
        second: 0,
        weekday: 0,
        yearday: 0,
        is_dst: false
    }
    var sec_double = f64(self.unix_sec) + f64(self.unix_nsec) / 1000000000.0
    nr_time_to_parts(sec_double, true, &parts)
    return parts
}

pub fn (self: #Time) Local() DateParts {
    var parts = DateParts {
        year: 0,
        month: 0,
        day: 0,
        hour: 0,
        minute: 0,
        second: 0,
        weekday: 0,
        yearday: 0,
        is_dst: false
    }
    var sec_double = f64(self.unix_sec) + f64(self.unix_nsec) / 1000000000.0
    nr_time_to_parts(sec_double, false, &parts)
    return parts
}

pub fn (self: #Time) Add(d: #Duration) Time {
    var total_nsec = i64(self.unix_nsec) + d.nanoseconds
    var sec_add = total_nsec / 1000000000
    var nsec_rem = total_nsec % 1000000000
    if (nsec_rem < 0) {
        sec_add = sec_add - 1
        nsec_rem = nsec_rem + 1000000000
    }
    return Time {
        unix_sec: self.unix_sec + sec_add,
        unix_nsec: i32(nsec_rem)
    }
}

pub fn (self: #Time) Sub(other: #Time) Duration {
    var self_ns = (self.unix_sec * 1000000000) + i64(self.unix_nsec)
    var other_ns = (other.unix_sec * 1000000000) + i64(other.unix_nsec)
    return Duration {
        nanoseconds: self_ns - other_ns
    }
}

pub fn (self: #Time) Before(other: #Time) bool {
    if (self.unix_sec < other.unix_sec) {
        return true
    }
    if (self.unix_sec > other.unix_sec) {
        return false
    }
    return self.unix_nsec < other.unix_nsec
}

pub fn (self: #Time) After(other: #Time) bool {
    if (self.unix_sec > other.unix_sec) {
        return true
    }
    if (self.unix_sec < other.unix_sec) {
        return false
    }
    return self.unix_nsec > other.unix_nsec
}

pub fn (self: #Time) Eq(other: #Time) bool {
    return self.unix_sec == other.unix_sec && self.unix_nsec == other.unix_nsec
}

pub fn Sleep(d: #Duration) {
    var ms = d.nanoseconds / 1000000
    if (ms < 1 && d.nanoseconds > 0) {
        ms = 1
    }
    nr_sleep_ms(i32(ms))
}

pub type Ticker = struct {
    C: chan[Time],
    duration: Duration,
    stop_ch: chan[bool]
}

pub fn NewTicker(d: Duration) @Ticker {
    var c = make(chan[Time], 1)
    var stop_ch = make(chan[bool], 1)
    var duration = Duration {
        nanoseconds: d.nanoseconds
    }
    var ticker = alloc Ticker {
        C: c,
        duration: duration,
        stop_ch: stop_ch
    }

    spawn ticker_worker(ticker.C.clone(), d, ticker.stop_ch.clone())

    return ticker
}

fn ticker_worker(c: chan[Time], d: Duration, stop_ch: chan[bool]) {
    while (true) {
        Sleep(#d)

        select {
            case <-stop_ch:
                break;
            default:
                select {
                    case c <- Now():
                    default:
                }
        }
    }
}

pub fn (self: &Ticker) Stop() {
    select {
        case self.stop_ch <- true:
        default:
    }
}

pub type Timer = struct {
    C: chan[Time],
    stop_ch: chan[bool]
}

pub fn NewTimer(d: Duration) @Timer {
    var c = make(chan[Time], 1)
    var stop_ch = make(chan[bool], 1)
    
    var duration = Duration {
        nanoseconds: d.nanoseconds
    }
    
    var timer = alloc Timer {
        C: c,
        stop_ch: stop_ch
    }

    spawn timer_worker(timer.C.clone(), duration, timer.stop_ch.clone())

    return timer
}

fn timer_worker(c: chan[Time], d: Duration, stop_ch: chan[bool]) {
    Sleep(#d)

    select {
        case <-stop_ch:
            return
        default:
            select {
                case c <- Now():
                default:
            }
    }
}

pub fn (self: &Timer) Stop() {
    select {
        case self.stop_ch <- true:
        default:
    }
}

pub fn After(d: Duration) chan[Time] {
    var timer = NewTimer(d)
    return timer.C.clone()
}
