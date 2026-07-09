# Scheduler Worker Thread CPU Overhead in Single-Fiber Programs

## Status
Resolved

## Problem
A Nora program that runs a tight render loop (no spawned fibers) exhibits high CPU usage proportional to the number of logical CPU cores on the machine. In testing with the `nora_wgpu` spinning cube example, the process consumed ~20% CPU on a multi-core machine while rendering at a capped ~60 FPS with a `sys.Sleep(16)` call per frame. Setting `NORA_NUM_WORKERS=1` dropped CPU usage to ~0%.

## Reproduction
Run any tight Nora loop that does not spawn fibers but performs per-frame work:

```nora
// examples/cube/main.nr
while win.PollEvents() {
    // ... render work (matrix updates, wgpu submit, present) ...
    sys.Sleep(16) // Direct Win32 Sleep - correctly yields OS thread
}
```

Run with the default scheduler (one worker thread per CPU core):
```powershell
.\build\debug\cube.exe
# --> ~20% CPU on a 5-core machine (Task Manager Details tab)
```

Run with a single worker thread:
```powershell
$env:NORA_NUM_WORKERS = "1"; .\build\debug\cube.exe
# --> ~0% CPU
```

## Root Cause

### 1. Cooperative Yield Checkpoints
The Nora compiler inserts `NR_COOPERATIVE_YIELD_CHECKPOINT()` macros into the prologue of every generated function. In the generated C output (`build/debug/out_pkg_*.c`), there are hundreds of these checkpoints across all packages. This macro calls `nr_cooperative_yield()` periodically, which:
1. Sets `yield_pending = true` on the current fiber
2. Calls `park()` → which calls `SwitchToFiber(main_fiber)` to return to the scheduler

### 2. Worker Thread Wakeup Storm
When `park()` is called, the `resume()` function is eventually called to re-queue the fiber. `resume()` calls `ReleaseSemaphore(g_worker_sem, 1, ...)` to signal a sleeping worker thread to wake up and process the re-queued fiber. 

With `N` worker threads sleeping on `g_worker_sem`, each cooperative yield point in the render loop triggers:
- 1 semaphore release → wakes 1 worker thread
- Worker thread finds the fiber in the queue, checks state, potentially re-parks
- Context switches back and forth between OS threads

This happens dozens of times per frame due to the dense yield checkpoints in every function call. At 60 FPS × N checkpoints × (N-1) worker threads thrashing on the semaphore, the overhead becomes significant:

```
~50 yield checkpoints/frame × 60 FPS × 4 workers = ~12,000 unnecessary thread wakeups/second
```

### 3. Why `sys.Sleep` Doesn't Help
`sys.Sleep(16)` calls Win32 `Sleep()` directly and correctly suspends the OS thread. However, this only adds a 16ms pause to the main thread. The problem occurs during the active frame work **before** the sleep, where cooperative yield checkpoints cause repeated park/unpark cycles that signal idle worker threads.

## Impact
- Any single-fiber Nora graphics or game loop will consume excess CPU proportional to CPU core count
- A 10-core machine would show ~40% wasted CPU for a program that should idle at ~0%
- This makes Nora unsuitable for real-time applications without the `NORA_NUM_WORKERS=1` workaround

## Workaround
Set the `NORA_NUM_WORKERS=1` environment variable before launching any single-fiber Nora program:

```powershell
$env:NORA_NUM_WORKERS = "1"; .\build\debug\cube.exe
```

Or, for building and running via the CLI:
```powershell
$env:NORA_NUM_WORKERS = "1"; nora run --example cube
```

## Proposed Fixes

### Fix A: Runtime Heuristic (Recommended Short-Term)
In `scheduler_init()` inside `pkg/std/runtime/fiber.c`, default `num_workers = 1` unless the program explicitly spawns a fiber. Once the first `spawn` is called, dynamically scale up to `sysinfo.dwNumberOfProcessors`. This requires an atomic flag `g_any_fiber_ever_spawned`.

```c
void scheduler_init() {
    // Start with 1 worker. Scale up lazily when fibers are first spawned.
    num_workers = 1;
    char* env_workers = getenv("NORA_NUM_WORKERS");
    if (env_workers != NULL) {
        int parsed = atoi(env_workers);
        if (parsed >= 1 && parsed <= MAX_WORKERS) num_workers = parsed;
    }
    ...
}
```

### Fix B: Smarter Yield Checkpoint (Recommended Long-Term)
The `NR_COOPERATIVE_YIELD_CHECKPOINT()` macro should check whether there are any *other* runnable fibers before yielding. If there is only one fiber alive (`g_active_fibers == 0` for spawned fibers and no queue entries), the checkpoint should be a no-op:

```c
#define NR_COOPERATIVE_YIELD_CHECKPOINT() do { \
    if (NR_ATOMIC_LOAD(&g_active_fibers) > 0) { \
        nr_cooperative_yield(); \
    } \
} while(0)
```

This avoids the park/resume cycle entirely when there is no other work to schedule, eliminating the wakeup storm at zero cost.

### Fix C: Batch Yield Checkpoints
Rather than yielding on every function call, the checkpoint should use a counter and only yield every N calls (already partially implemented via `g_yield_ticks`). Verify the tick threshold is high enough that single-fiber tight loops don't yield excessively.

## Applied Fixes

### Fix B — Yield Checkpoint Guard (`std/runtime/runtime.h`)

**Key discovery:** `nr_main` itself is spawned as a fiber via `scheduler_spawn`, so `g_active_fibers == 1` throughout the entire program lifetime. The original check `> 0` was therefore always true. The correct threshold is `> 1` (user-spawned fibers exist beyond the main fiber itself).

```diff
- if (++g_yield_ticks >= 1000) {
+ if (NR_ATOMIC_LOAD(&g_active_fibers) > 1 && ++g_yield_ticks >= 1000) {
```

Effect: the yield checkpoint becomes a no-op for single-fiber programs, eliminating the worker thread wakeup storm.

---

### Smart `nr_sleep_ms` (`std/runtime/time.c`)

`nr_time_init()` creates a timer poller OS thread that loops with `Sleep(1)` 1000×/second. It must NOT be called when we're taking the OS sleep path. The check must occur **before** `nr_time_init()` is called, and also uses threshold `<= 1`:

```c
bool no_other_fibers = NR_ATOMIC_LOAD(&g_active_fibers) <= 1;
if (!self || no_other_fibers) {
    Sleep(ms);  // OS-level thread suspend, 0% idle CPU
    return;
}
nr_time_init();  // Only initialized when actually needed
// ... fiber park path
```

---

## Validation Results

| Configuration | CPU % | Notes |
|---|---|---|
| Default workers, no sleep cap | ~50% | Original state — full busy-wait |
| `NORA_NUM_WORKERS=1`, `sys.Sleep(16)` | **~0%** | Confirmed workaround |
| Fix B only, `sys.Sleep(16)` | **~0%** | Confirmed — yield wakeup storm eliminated |
| Fix B + smart `time.Sleep` | ~10% | ⚠️ Unexplained residual overhead |
| Fix B + `sys.Sleep(16)` | **~0%** | ✅ Final confirmed state |

## Known Remaining Issue: `time.Sleep` Overhead

Despite the smart `nr_sleep_ms` correctly routing to `Sleep(ms)` when `g_active_fibers <= 1`, importing the `time` package and calling `time.Sleep` in a render loop still results in ~10% CPU vs. 0% with `sys.Sleep`.

**Hypotheses (not yet confirmed):**
1. The `time.Sleep` Nora wrapper (`time_Sleep`) goes through multiple function call layers that change clang's inlining decisions in `-O0` mode, increasing render work overhead
2. The `time` package's compiled code (`out_pkg_time.c`) includes channel-based `Ticker` infrastructure that interacts with the scheduler's internal state
3. The `NR_COOPERATIVE_YIELD_CHECKPOINT()` in `time_Sleep`'s prologue, even as a no-op, creates a measurable footprint across 60 calls/second × N checkpoints per call

**Current workaround for render loops:** Use `sys.Sleep(ms)` (direct Win32 `Sleep` FFI binding) which bypasses the Nora time package and fiber scheduler entirely, consistently giving ~0% idle CPU.

**Outstanding work:** Requires profiling (e.g., with Very Sleepy or ETW traces) to identify exactly which instruction accounts for the 10% residual when `time.Sleep` is used.
