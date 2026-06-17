#include "runtime.h"

// Debugger hook for fiber start
#ifdef _WIN32
__declspec(noinline) void __nora_fiber_started(void* parent) {
    // dummy function for debugger breakpoints
    __asm__(""); // prevent optimization
}
#else
__attribute__((noinline)) void __nora_fiber_started(void* parent) {
    // dummy function for debugger breakpoints
    __asm__(""); // prevent optimization
}
#endif

volatile void* __nora_step_target_fiber = NULL;
volatile int __nora_stepping_into_spawn = 0;
volatile void* __nora_debug_locked_fiber = NULL;

#define NR_DEBUG_CHECK_LOCK(p_info) do { \
    if (__nora_debug_locked_fiber && (p_info) != __nora_debug_locked_fiber) { \
        fiber_info_t* locked = (fiber_info_t*)__nora_debug_locked_fiber; \
        long locked_state = NR_ATOMIC_LOAD(&locked->state); \
        if (locked_state < 3) { \
            if (p_info) { \
                queue_push(&g_queue, (p_info)); \
            } \
            (p_info) = NULL; \
        } \
    } \
} while(0)

// --- FIBER RUNTIME (Global Queue) ---
#ifndef __EMSCRIPTEN__
#ifdef _WIN32
#include <windows.h>
#include <process.h>

#define MAX_WORKERS 64
#define QUEUE_SIZE 4096

#ifdef _MSC_VER
#define THREAD_LOCAL __declspec(thread)
#else
#define THREAD_LOCAL __thread
#endif

typedef LPVOID fiber_t;

typedef struct fiber_info {
    fiber_t handle;
    NR_ATOMIC_INT state; // 0: READY, 1: RUNNING, 2: PARKING, 3: PARKED, 4: TERMINATED
    NR_ATOMIC_INT resume_pending;
    bool is_main;
    char* temp_strs[256];
    int temp_count;
    const char* name;
    const char* file;
    int line;
    int pinned_worker_id;
    volatile bool yield_pending;
    struct fiber_info* next_global;
    struct fiber_info* prev_global;
    void* parent;
    jmp_buf panic_buf;
    const char* panic_msg;
    const char* panic_file;
    int panic_line;
} fiber_info_t;

typedef struct fiber_node {
    fiber_info_t* fiber;
    struct fiber_node* next;
} fiber_node_t;

typedef struct {
    fiber_node_t* head;
    fiber_node_t* tail;
    int count;
    NR_MUTEX_T lock;
} global_queue_t;

// Forward declarations
void park();
void resume(fiber_info_t* info);
_Noreturn void nr_panic(const char* msg, const char* file, int line);
void nr_fiber_report();

typedef struct timer_waiter {
    double wake_time;
    fiber_info_t* fiber;
    struct timer_waiter* next;
} timer_waiter_t;

static timer_waiter_t* g_timer_waiters_head = NULL;
static NR_MUTEX_T g_timer_waiters_lock;
static bool g_timer_poller_running = false;

extern NR_ATOMIC_INT g_timer_waiters_count;
double nr_time_now(void);

#ifdef _WIN32
static DWORD WINAPI nr_timer_poller_thread(LPVOID arg) {
#else
static void* nr_timer_poller_thread(void* arg) {
#endif
    while (g_timer_poller_running) {
        double now = nr_time_now();
        NR_MUTEX_LOCK(&g_timer_waiters_lock);
        timer_waiter_t* curr = g_timer_waiters_head;
        timer_waiter_t* prev = NULL;
        while (curr) {
            if (now >= curr->wake_time) {
                NR_ATOMIC_SUB(&g_timer_waiters_count, 1);
                resume(curr->fiber);
                timer_waiter_t* next = curr->next;
                if (prev) {
                    prev->next = next;
                } else {
                    g_timer_waiters_head = next;
                }
                free(curr);
                curr = next;
            } else {
                prev = curr;
                curr = curr->next;
            }
        }
        NR_MUTEX_UNLOCK(&g_timer_waiters_lock);

#ifdef _WIN32
        Sleep(1);
#else
        usleep(1000);
#endif
    }
    return 0;
}

void nr_time_init(void) {
    static bool timer_poller_initialized = false;
    if (!timer_poller_initialized) {
        NR_MUTEX_INIT(&g_timer_waiters_lock);
        g_timer_poller_running = true;
#ifdef _WIN32
        CreateThread(NULL, 0, nr_timer_poller_thread, NULL, 0, NULL);
#else
        pthread_t tid;
        pthread_create(&tid, NULL, nr_timer_poller_thread, NULL);
#endif
        timer_poller_initialized = true;
    }
}

#define DEQUE_CAPACITY 65536

extern HANDLE g_worker_sem;

typedef struct {
    NR_ATOMIC_INT bottom;
    NR_ATOMIC_INT top;
    fiber_info_t* volatile buffer[DEQUE_CAPACITY];
} deque_t;

void deque_init(deque_t* q) {
    NR_ATOMIC_STORE(&q->bottom, 0);
    NR_ATOMIC_STORE(&q->top, 0);
    memset((void*)q->buffer, 0, sizeof(q->buffer));
}

void deque_push(deque_t* q, fiber_info_t* f) {
    int b = (int)NR_ATOMIC_LOAD(&q->bottom);
    int t = (int)NR_ATOMIC_LOAD(&q->top);
    if (b - t >= DEQUE_CAPACITY) {
        return;
    }
    q->buffer[b & (DEQUE_CAPACITY - 1)] = f;
    // Sequential consistency release write
    NR_ATOMIC_STORE(&q->bottom, b + 1);
    ReleaseSemaphore(g_worker_sem, 1, NULL);
}

fiber_info_t* deque_pop(deque_t* q) {
    int b = (int)NR_ATOMIC_LOAD(&q->bottom) - 1;
    NR_ATOMIC_STORE(&q->bottom, b);
    
    // CRITICAL: Memory barrier prevents bottom decrement being reordered after top load
    MemoryBarrier();
    
    int t = (int)NR_ATOMIC_LOAD(&q->top);
    if (b > t) {
        // At least one element left safely, no conflict
        return q->buffer[b & (DEQUE_CAPACITY - 1)];
    }
    if (b == t) {
        // Exactly one element left, conflict with concurrent stealers
        fiber_info_t* f = q->buffer[b & (DEQUE_CAPACITY - 1)];
        int expected_t = t;
        if (NR_ATOMIC_CAS(&q->top, &expected_t, t + 1)) {
            // CAS succeeded, we popped it
            NR_ATOMIC_STORE(&q->bottom, t + 1);
            return f;
        }
        // Lost race to a stealer
        NR_ATOMIC_STORE(&q->bottom, t + 1);
        return NULL;
    }
    // Already empty
    NR_ATOMIC_STORE(&q->bottom, t);
    return NULL;
}

fiber_info_t* deque_steal(deque_t* q) {
    int t = (int)NR_ATOMIC_LOAD(&q->top);
    MemoryBarrier();
    int b = (int)NR_ATOMIC_LOAD(&q->bottom);
    if (t >= b) {
        return NULL;
    }
    fiber_info_t* f = q->buffer[t & (DEQUE_CAPACITY - 1)];
    int expected_t = t;
    if (NR_ATOMIC_CAS(&q->top, &expected_t, t + 1)) {
        return f;
    }
    return NULL;
}

global_queue_t g_queue;
global_queue_t g_pinned_queues[MAX_WORKERS];
deque_t g_local_queues[MAX_WORKERS];
int num_workers = 0;
THREAD_LOCAL int worker_id = -1;
THREAD_LOCAL int g_yield_ticks = 0;
fiber_t main_fibers[MAX_WORKERS];
fiber_info_t main_fiber_infos[MAX_WORKERS];
NR_ATOMIC_LONG g_active_fibers = 0;
volatile bool g_running = true;
fiber_info_t* g_fibers_head = NULL;
fiber_info_t* g_terminated_fibers_head = NULL;
NR_MUTEX_T g_fiber_list_lock;

// Signaling
HANDLE g_worker_sem;
NR_ATOMIC_INT g_sleeping_workers = 0;

void queue_init(global_queue_t* q) {
    q->head = q->tail = NULL;
    q->count = 0;
    NR_MUTEX_INIT(&q->lock);
}

void queue_push(global_queue_t* q, fiber_info_t* f) {
    fiber_node_t* node = (fiber_node_t*)malloc(sizeof(fiber_node_t));
    node->fiber = f;
    node->next = NULL;
    NR_MUTEX_LOCK(&q->lock);
    if (q->tail) {
        q->tail->next = node;
        q->tail = node;
    } else {
        q->head = q->tail = node;
    }
    q->count++;
    NR_MUTEX_UNLOCK(&q->lock);
    ReleaseSemaphore(g_worker_sem, 1, NULL);
}

void* queue_pop(global_queue_t* q) {
    fiber_info_t* f = NULL;
    NR_MUTEX_LOCK(&q->lock);
    if (q->head) {
        fiber_node_t* node = q->head;
        f = node->fiber;
        q->head = node->next;
        if (!q->head) q->tail = NULL;
        q->count--;
        free(node);
    }
    NR_MUTEX_UNLOCK(&q->lock);
    return f;
}

void worker_loop(void* arg) {
    worker_id = (int)(intptr_t)arg;
    
    // Processor Affinity [Disabled for dynamic OS thread scheduling]
    // GROUP_AFFINITY affinity = {0};
    // affinity.Mask = (KAFFINITY)1 << (worker_id % 64);
    // affinity.Group = worker_id / 64;
    // SetThreadGroupAffinity(GetCurrentThread(), &affinity, NULL);

    main_fibers[worker_id] = ConvertThreadToFiber(&main_fiber_infos[worker_id]);
    main_fiber_infos[worker_id].handle = main_fibers[worker_id];
    main_fiber_infos[worker_id].state = 0;
    main_fiber_infos[worker_id].is_main = true; 

    while (g_running) {
        fiber_info_t* info = NULL;
        if (worker_id >= 0 && worker_id < num_workers) {
            info = (fiber_info_t*)queue_pop(&g_pinned_queues[worker_id]);
            if (!info) info = deque_pop(&g_local_queues[worker_id]);
        }
        if (!info) {
            info = (fiber_info_t*)queue_pop(&g_queue);
        }
        if (!info && worker_id >= 0) {
            for (int j = 1; j < num_workers; j++) {
                int victim = (worker_id + j) % num_workers;
                info = deque_steal(&g_local_queues[victim]);
                if (info) break;
            }
        }

        NR_DEBUG_CHECK_LOCK(info);
        if (info) {
            NR_ATOMIC_STORE(&info->state, 1); // RUNNING
            SwitchToFiber(info->handle);
            long s = NR_ATOMIC_LOAD(&info->state);
            if (s == 2) { // PARKING
                NR_ATOMIC_STORE(&info->state, 3); // PARKED
                if (info->yield_pending) {
                    resume(info);
                } else {
                    while (1) {
                        long rp = NR_ATOMIC_LOAD(&info->resume_pending);
                        if (rp <= 0) break;
                        long new_rp = rp - 1;
                        if (NR_ATOMIC_CAS(&info->resume_pending, &rp, new_rp)) {
                            resume(info);
                            break;
                        }
                    }
                }
            } else if (s == 4) { // TERMINATED
                // Fiber is done.
            }
        } else {
            if (__nora_debug_locked_fiber) {
                nr_sleep_ms(1);
                continue;
            }
            if (NR_ATOMIC_LOAD(&g_active_fibers) == 0 && worker_id == 0) break;
            
            NR_ATOMIC_INC(&g_sleeping_workers);
            if (NR_ATOMIC_LOAD(&g_sleeping_workers) == num_workers && NR_ATOMIC_LOAD(&g_active_fibers) > 0) {
                NR_MUTEX_LOCK(&g_fiber_list_lock);
                fiber_info_t* curr = g_fibers_head;
                bool any_runnable = false;
                while (curr) {
                    long s = NR_ATOMIC_LOAD(&curr->state);
                    if (s == 0 || s == 1 || s == 2) {
                        any_runnable = true;
                        break;
                    }
                    curr = curr->next_global;
                }
                NR_MUTEX_UNLOCK(&g_fiber_list_lock);
                
                if (!any_runnable && NR_ATOMIC_LOAD(&g_net_waiters_count) == 0 && NR_ATOMIC_LOAD(&g_timer_waiters_count) == 0) {
                    printf("\nFATAL: Deadlock detected! All %ld fibers are blocked.\n", (long)NR_ATOMIC_LOAD(&g_active_fibers));
                    nr_fiber_report();
                    nr_panic("deadlock", "runtime", 0);
                }
            }
            WaitForSingleObject(g_worker_sem, INFINITE);
            NR_ATOMIC_DEC(&g_sleeping_workers);
        }
    }
    if (worker_id > 0) _endthread();
}

void scheduler_init() {
    SYSTEM_INFO sysinfo;
    GetSystemInfo(&sysinfo);
    num_workers = sysinfo.dwNumberOfProcessors;
    char* env_workers = getenv("NORA_NUM_WORKERS");
    if (env_workers != NULL) {
        int parsed = atoi(env_workers);
        if (parsed >= 1 && parsed <= MAX_WORKERS) {
            num_workers = parsed;
        }
    } else {
        if (num_workers > MAX_WORKERS) num_workers = MAX_WORKERS;
    }

    g_worker_sem = CreateSemaphore(NULL, 0, 1000000, NULL);
    NR_MUTEX_INIT(&g_fiber_list_lock);
    queue_init(&g_queue);
    for (int i = 0; i < MAX_WORKERS; i++) {
        queue_init(&g_pinned_queues[i]);
        deque_init(&g_local_queues[i]);
    }

    for (int i = 1; i < num_workers; i++) {
        _beginthread(worker_loop, 0, (void*)(intptr_t)i);
    }
    worker_id = 0;
    main_fiber_infos[0].state = 1;
    main_fiber_infos[0].resume_pending = 0;
    main_fiber_infos[0].is_main = true;
    main_fiber_infos[0].temp_count = 0;
    main_fibers[0] = ConvertThreadToFiber(&main_fiber_infos[0]);
    main_fiber_infos[0].handle = main_fibers[0];
    main_fiber_infos[0].state = 4;
    main_fiber_infos[0].name = "main";
    main_fiber_infos[0].file = "main";
    main_fiber_infos[0].line = 0;

    NR_MUTEX_LOCK(&g_fiber_list_lock);
    main_fiber_infos[0].next_global = g_fibers_head;
    if (g_fibers_head) g_fibers_head->prev_global = &main_fiber_infos[0];
    g_fibers_head = &main_fiber_infos[0];
    NR_MUTEX_UNLOCK(&g_fiber_list_lock);
    
    // GROUP_AFFINITY affinity = {0};
    // affinity.Mask = (KAFFINITY)1;
    // affinity.Group = 0;
    // SetThreadGroupAffinity(GetCurrentThread(), &affinity, NULL);
}
void scheduler_run_loop() {
    while (g_running) {
        fiber_info_t* info = NULL;
        int id = worker_id;
        if (id >= 0 && id < num_workers) {
            info = (fiber_info_t*)queue_pop(&g_pinned_queues[id]);
            if (!info) info = deque_pop(&g_local_queues[id]);
        }
        if (!info) {
            info = (fiber_info_t*)queue_pop(&g_queue);
        }
        if (!info && id >= 0) {
            for (int j = 1; j < num_workers; j++) {
                int victim = (id + j) % num_workers;
                info = deque_steal(&g_local_queues[victim]);
                if (info) break;
            }
        }

        NR_DEBUG_CHECK_LOCK(info);
        if (info) {
#ifdef NR_DEBUG_FIBER
            printf("[C-SCHED] Thread %d: Popped/Stole fiber %s (%p), state = %ld, switching context...\n", id, info->name ? info->name : "unnamed", info, (long)NR_ATOMIC_LOAD(&info->state));
#endif
            NR_ATOMIC_STORE(&info->state, 1); // RUNNING
            SwitchToFiber(info->handle);
#ifdef NR_DEBUG_FIBER
            printf("[C-SCHED] Thread %d: Switched back from fiber %s (%p)\n", id, info->name ? info->name : "unnamed", info);
#endif
            long s = NR_ATOMIC_LOAD(&info->state);
            if (s == 2) { // PARKING
                NR_ATOMIC_STORE(&info->state, 3); // PARKED
                if (info->yield_pending) {
                    resume(info);
                } else {
                    while (1) {
                        long rp = NR_ATOMIC_LOAD(&info->resume_pending);
                        if (rp <= 0) break;
                        long new_rp = rp - 1;
                        if (NR_ATOMIC_CAS(&info->resume_pending, &rp, new_rp)) {
                            resume(info);
                            break;
                        }
                    }
                }
            } else if (s == 4) { // TERMINATED
                // Fiber is done.
            }
        } else {
            if (__nora_debug_locked_fiber) {
                nr_sleep_ms(1);
                continue;
            }
            if (NR_ATOMIC_LOAD(&g_active_fibers) == 0) break;
            
            NR_ATOMIC_INC(&g_sleeping_workers);
            if (NR_ATOMIC_LOAD(&g_sleeping_workers) == num_workers && NR_ATOMIC_LOAD(&g_active_fibers) > 0) {
                NR_MUTEX_LOCK(&g_fiber_list_lock);
                fiber_info_t* curr = g_fibers_head;
                bool any_runnable = false;
                while (curr) {
                    long s = NR_ATOMIC_LOAD(&curr->state);
                    if (s == 0 || s == 1 || s == 2) {
                        any_runnable = true;
                        break;
                    }
                    curr = curr->next_global;
                }
                NR_MUTEX_UNLOCK(&g_fiber_list_lock);
                
                if (!any_runnable && NR_ATOMIC_LOAD(&g_net_waiters_count) == 0 && NR_ATOMIC_LOAD(&g_timer_waiters_count) == 0) {
                    printf("\nFATAL: Deadlock detected! All %ld fibers are blocked.\n", (long)NR_ATOMIC_LOAD(&g_active_fibers));
                    nr_fiber_report();
                    nr_panic("deadlock", "runtime", 0);
                }
            }
            WaitForSingleObject(g_worker_sem, INFINITE);
            NR_ATOMIC_DEC(&g_sleeping_workers);
        }
    }
}

void scheduler_cleanup() {
    g_running = false;
    ReleaseSemaphore(g_worker_sem, num_workers, NULL);
    Sleep(50); // Give workers time to exit
    CloseHandle(g_worker_sem);
    NR_MUTEX_DESTROY(&g_queue.lock);

    // Free all active fibers that didn't terminate cleanly
    NR_MUTEX_LOCK(&g_fiber_list_lock);
    fiber_info_t* curr = g_fibers_head;
    while (curr) {
        fiber_info_t* next = curr->next_global;
        if (!curr->is_main) {
            if (curr->handle) {
                DeleteFiber(curr->handle);
            }
            free(curr);
        }
        curr = next;
    }
    g_fibers_head = NULL;

    // Free all terminated fibers
    curr = g_terminated_fibers_head;
    while (curr) {
        fiber_info_t* next = curr->next_global;
        if (curr->handle) {
            DeleteFiber(curr->handle);
        }
        free(curr);
        curr = next;
    }
    g_terminated_fibers_head = NULL;
    NR_MUTEX_UNLOCK(&g_fiber_list_lock);
    NR_MUTEX_DESTROY(&g_fiber_list_lock);
}

void park() {
    fiber_info_t* info = (fiber_info_t*)GetFiberData();
    if (info) {
        long s = NR_ATOMIC_LOAD(&info->state);
        if (s == 1) NR_ATOMIC_STORE(&info->state, 2); // PARKING
    }
    if (worker_id >= 0) SwitchToFiber(main_fibers[worker_id]);
}

void nr_cooperative_yield() {
    fiber_info_t* self = (fiber_info_t*)GetFiberData();
    if (!self) return;
    self->yield_pending = true;
    park();
}

void resume(fiber_info_t* info) {
    if (!info) return;
    if (NR_ATOMIC_LOAD(&info->state) == 4) return; // Don't resume terminated fibers

    bool is_yield = false;
    if (info->yield_pending) {
        is_yield = true;
        info->yield_pending = false;
    }

    long expected = 3; // PARKED
#ifdef NR_DEBUG_FIBER
    printf("[C-SCHED] resume: Attempting to resume fiber %s (%p), current state = %ld...\n", info->name ? info->name : "unnamed", info, (long)NR_ATOMIC_LOAD(&info->state));
#endif
    if (NR_ATOMIC_CAS(&info->state, &expected, 0)) {
        int id = worker_id;
#ifdef NR_DEBUG_FIBER
        printf("[C-SCHED] resume: Successfully CASed state of fiber %s (%p) to 0. Queuing...\n", info->name ? info->name : "unnamed", info);
#endif
        if (id >= 0 && id < num_workers && num_workers > 1 && !is_yield) {
            if (info->pinned_worker_id >= 0) queue_push(&g_pinned_queues[info->pinned_worker_id], info);
            else deque_push(&g_local_queues[id], info);
        } else {
            if (info->pinned_worker_id >= 0) queue_push(&g_pinned_queues[info->pinned_worker_id], info);
            else queue_push(&g_queue, info);
        }
        ReleaseSemaphore(g_worker_sem, 1, NULL);
    } else {
#ifdef NR_DEBUG_FIBER
        printf("[C-SCHED] resume: CAS failed for fiber %s (%p) (not in PARKED state). Incrementing resume_pending (currently %ld)...\n", info->name ? info->name : "unnamed", info, (long)NR_ATOMIC_LOAD(&info->resume_pending));
#endif
        NR_ATOMIC_ADD(&info->resume_pending, 1);
        // Double check
        expected = 3;
        if (NR_ATOMIC_CAS(&info->state, &expected, 0)) {
#ifdef NR_DEBUG_FIBER
            printf("[C-SCHED] resume: Double-check CAS succeeded for fiber %s (%p)!\n", info->name ? info->name : "unnamed", info);
#endif
            if (NR_ATOMIC_SUB(&info->resume_pending, 1) > 0) {
                int id = worker_id;
                if (id >= 0 && id < num_workers && num_workers > 1 && !is_yield) {
                    if (info->pinned_worker_id >= 0) queue_push(&g_pinned_queues[info->pinned_worker_id], info);
                    else deque_push(&g_local_queues[id], info);
                } else {
                    if (info->pinned_worker_id >= 0) queue_push(&g_pinned_queues[info->pinned_worker_id], info);
                    else queue_push(&g_queue, info);
                }
                ReleaseSemaphore(g_worker_sem, 1, NULL);
            }
        }
    }
}

int get_worker_id() { return worker_id; }

typedef struct {
    void (*fn)(void*);
    void* arg;
} spawn_data_t;

void WINAPI fiber_wrapper(LPVOID p) {
    fiber_info_t* info = (fiber_info_t*)p;
    spawn_data_t* data = (spawn_data_t*)(info + 1);
    data->fn(data->arg);
    nr_flush_temps();
    NR_ATOMIC_DEC(&g_active_fibers);
    
    NR_MUTEX_LOCK(&g_fiber_list_lock);
    if (info->prev_global) info->prev_global->next_global = info->next_global;
    if (info->next_global) info->next_global->prev_global = info->prev_global;
    if (info == g_fibers_head) g_fibers_head = info->next_global;

    // Push to terminated list
    info->next_global = g_terminated_fibers_head;
    info->prev_global = NULL;
    if (g_terminated_fibers_head) g_terminated_fibers_head->prev_global = info;
    g_terminated_fibers_head = info;
    NR_MUTEX_UNLOCK(&g_fiber_list_lock);

    if (info->is_main) {
        g_running = false;
        ReleaseSemaphore(g_worker_sem, num_workers, NULL);
    }
    NR_ATOMIC_STORE(&info->state, 4); // TERMINATED
    park();
}

void* scheduler_spawn(void (*fn)(void*), void* arg, const char* name, const char* file, int line) {
    fiber_info_t* info = (fiber_info_t*)malloc(sizeof(fiber_info_t) + sizeof(spawn_data_t));
    memset(info, 0, sizeof(fiber_info_t) + sizeof(spawn_data_t));
    NR_ATOMIC_STORE(&info->state, 0); // READY
    NR_ATOMIC_STORE(&info->resume_pending, 0);
    info->pinned_worker_id = -1;
    info->temp_count = 0;
    info->name = name;
    info->file = file;
    info->line = line;
    info->parent = nr_fiber_current();

    if (__nora_stepping_into_spawn) {
        __nora_stepping_into_spawn = 0;
        __nora_step_target_fiber = info;
    }

    NR_MUTEX_LOCK(&g_fiber_list_lock);
    info->next_global = g_fibers_head;
    if (g_fibers_head) g_fibers_head->prev_global = info;
    g_fibers_head = info;
    NR_MUTEX_UNLOCK(&g_fiber_list_lock);

    long old_count = NR_ATOMIC_ADD(&g_active_fibers, 1);
    info->is_main = (old_count == 0);
    
    spawn_data_t* data = (spawn_data_t*)(info + 1);
    data->fn = fn;
    data->arg = arg;
    info->handle = CreateFiber(0, (LPFIBER_START_ROUTINE)fiber_wrapper, info);
    int id = worker_id;
    if (id >= 0 && id < num_workers && num_workers > 1) {
        deque_push(&g_local_queues[id], info);
    } else {
        queue_push(&g_queue, info);
    }
    return (fiber_t)info;
}

void* nr_fiber_current() {
    return GetFiberData();
}

void nr_fiber_suspend() {
    park();
}

void nr_fiber_pin_thread() {
    fiber_info_t* self = (fiber_info_t*)nr_fiber_current();
    if (self && worker_id >= 0) {
        self->pinned_worker_id = worker_id;
    }
}

void nr_fiber_unpin_thread() {
    fiber_info_t* self = (fiber_info_t*)nr_fiber_current();
    if (self) {
        self->pinned_worker_id = -1;
    }
}

void nr_fiber_resume(void* f) {
    resume((fiber_info_t*)f);
}

void nr_store_ptr(void* dest, void* val) {
    if (dest) *(void**)dest = val;
}

void nr_store_i32(void* dest, int val) {
    if (dest) *(int*)dest = val;
}

int nr_load_i32(void* p) {
    if (p) return *(volatile int*)p;
    return 0;
}

void* nr_load_ptr(void* p) {
    if (p) return *(void* volatile*)p;
    return NULL;
}

#elif defined(__wasm__)
// --- NATIVE WASM RUNTIME (Typed Continuations) ---
int g_yield_ticks = 0;
void nr_cooperative_yield() {}
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdio.h>

#define MAX_WORKERS 1
#define QUEUE_SIZE 4096

#ifdef NR_USE_FIBERS
// Native Wasm Continuations (Typed Continuations Proposal)
typedef void* fiber_t;

extern fiber_t wasm_cont_new(void (*fn)(void*), void* arg) __attribute__((import_module("wasm_runtime"), import_name("cont_new")));
extern void wasm_cont_resume(fiber_t cont) __attribute__((import_module("wasm_runtime"), import_name("cont_resume")));
extern void wasm_cont_suspend() __attribute__((import_module("wasm_runtime"), import_name("cont_suspend")));
#else
typedef void* fiber_t;
#define wasm_cont_new(fn, arg) (NULL)
#define wasm_cont_resume(cont) (void)(cont)
#define wasm_cont_suspend() (void)0
#endif

typedef struct {
    void (*fn)(void*);
    void* arg;
} spawn_data_t;

typedef struct fiber_info {
    fiber_t handle;
    NR_ATOMIC_INT state; // 0: READY, 1: RUNNING, 2: PARKING, 3: PARKED, 4: TERMINATED
    NR_ATOMIC_INT resume_pending;
    bool is_main;
    char* temp_strs[256];
    int temp_count;
    spawn_data_t data;
    const char* name;
    const char* file;
    int line;
    volatile bool yield_pending;
    struct fiber_info* next_global;
    struct fiber_info* prev_global;
    void* parent;
    jmp_buf panic_buf;
    const char* panic_msg;
    const char* panic_file;
    int panic_line;
} fiber_info_t;

typedef struct fiber_node {
    fiber_info_t* fiber;
    struct fiber_node* next;
} fiber_node_t;

typedef struct {
    fiber_node_t* head;
    fiber_node_t* tail;
    int count;
    NR_MUTEX_T lock;
} global_queue_t;

void park();
void resume(fiber_info_t* info);
_Noreturn void nr_panic(const char* msg, const char* file, int line);
void nr_fiber_report();

global_queue_t g_queue;
static fiber_info_t* g_current_fiber = NULL;
NR_ATOMIC_LONG g_active_fibers = 0;
volatile bool g_running = true;
fiber_info_t* g_fibers_head = NULL;
static fiber_info_t* g_terminated_fibers_head = NULL;
NR_MUTEX_T g_fiber_list_lock;

void queue_init(global_queue_t* q) {
    q->head = q->tail = NULL;
    q->count = 0;
    NR_MUTEX_INIT(&q->lock);
}

void queue_push(global_queue_t* q, fiber_info_t* f) {
    fiber_node_t* node = (fiber_node_t*)malloc(sizeof(fiber_node_t));
    node->fiber = f;
    node->next = NULL;
    NR_MUTEX_LOCK(&q->lock);
    if (q->tail) {
        q->tail->next = node;
        q->tail = node;
    } else {
        q->head = q->tail = node;
    }
    q->count++;
    NR_MUTEX_UNLOCK(&q->lock);
}

void* queue_pop(global_queue_t* q) {
    fiber_info_t* f = NULL;
    NR_MUTEX_LOCK(&q->lock);
    if (q->head) {
        fiber_node_t* node = q->head;
        f = node->fiber;
        q->head = node->next;
        if (!q->head) q->tail = NULL;
        q->count--;
        free(node);
    }
    NR_MUTEX_UNLOCK(&q->lock);
    return f;
}

void scheduler_init() {
    queue_init(&g_queue);
    NR_MUTEX_INIT(&g_fiber_list_lock);
    g_active_fibers = 0;
    g_running = true;
}

void scheduler_run_loop() {
    while (g_running) {
        fiber_info_t* info = (fiber_info_t*)queue_pop(&g_queue);
        if (info) {
            g_current_fiber = info;
            NR_ATOMIC_STORE(&info->state, 1); // RUNNING
            wasm_cont_resume(info->handle);
            
            int s = NR_ATOMIC_LOAD(&info->state);
            if (s == 2) { // PARKING
                NR_ATOMIC_STORE(&info->state, 3); // PARKED
                if (NR_ATOMIC_LOAD(&info->resume_pending) > 0) {
                    NR_ATOMIC_STORE(&info->resume_pending, 0);
                    resume(info);
                }
            }
        } else {
            if (NR_ATOMIC_LOAD(&g_active_fibers) == 0) break;
            
            // Check for deadlock (in Wasm we only have 1 worker)
            NR_MUTEX_LOCK(&g_fiber_list_lock);
            fiber_info_t* curr = g_fibers_head;
            bool any_runnable = false;
            while (curr) {
                long s = NR_ATOMIC_LOAD(&curr->state);
                if (s == 0 || s == 1) { // READY or RUNNING
                    any_runnable = true;
                    break;
                }
                curr = curr->next_global;
            }
            NR_MUTEX_UNLOCK(&g_fiber_list_lock);
            
            if (!any_runnable && NR_ATOMIC_LOAD(&g_net_waiters_count) == 0 && NR_ATOMIC_LOAD(&g_timer_waiters_count) == 0) {
                printf("\nFATAL: Deadlock detected!\n");
                nr_fiber_report();
                nr_panic("deadlock", "runtime", 0);
            }
            break; 
        }
    }
}

void scheduler_cleanup() {
    g_running = false;

    NR_MUTEX_LOCK(&g_fiber_list_lock);
    fiber_info_t* curr = g_fibers_head;
    while (curr) {
        fiber_info_t* next = curr->next_global;
        if (!curr->is_main) {
            free(curr);
        }
        curr = next;
    }
    g_fibers_head = NULL;

    curr = g_terminated_fibers_head;
    while (curr) {
        fiber_info_t* next = curr->next_global;
        free(curr);
        curr = next;
    }
    g_terminated_fibers_head = NULL;
    NR_MUTEX_UNLOCK(&g_fiber_list_lock);
    NR_MUTEX_DESTROY(&g_fiber_list_lock);
}

void* GetFiberData() {
    return g_current_fiber;
}

void park() {
    if (g_current_fiber && NR_ATOMIC_LOAD(&g_current_fiber->state) == 1) {
        NR_ATOMIC_STORE(&g_current_fiber->state, 2); // PARKING
        wasm_cont_suspend();
    }
}

void resume(fiber_info_t* info) {
    if (!info || NR_ATOMIC_LOAD(&info->state) == 4) return;
    if (NR_ATOMIC_LOAD(&info->state) == 3) {
        NR_ATOMIC_STORE(&info->state, 0); // READY
        queue_push(&g_queue, info);
    } else {
        NR_ATOMIC_INC(&info->resume_pending);
    }
}

void fiber_wrapper(void* p) {
    fiber_info_t* info = (fiber_info_t*)p;
    info->data.fn(info->data.arg);
    nr_flush_temps();
    NR_ATOMIC_DEC(&g_active_fibers);
    if (info->is_main) g_running = false;

    NR_MUTEX_LOCK(&g_fiber_list_lock);
    if (info->prev_global) info->prev_global->next_global = info->next_global;
    if (info->next_global) info->next_global->prev_global = info->prev_global;
    if (info == g_fibers_head) g_fibers_head = info->next_global;

    info->next_global = g_terminated_fibers_head;
    info->prev_global = NULL;
    if (g_terminated_fibers_head) g_terminated_fibers_head->prev_global = info;
    g_terminated_fibers_head = info;
    NR_MUTEX_UNLOCK(&g_fiber_list_lock);

    NR_ATOMIC_STORE(&info->state, 4); // TERMINATED
    wasm_cont_suspend();
}

void* scheduler_spawn(void (*fn)(void*), void* arg, const char* name, const char* file, int line) {
    fiber_info_t* info = (fiber_info_t*)malloc(sizeof(fiber_info_t));
    memset(info, 0, sizeof(fiber_info_t));
    info->data.fn = fn;
    info->data.arg = arg;
    info->name = name;
    info->file = file;
    info->line = line;
    info->parent = nr_fiber_current();

    if (__nora_stepping_into_spawn) {
        __nora_stepping_into_spawn = 0;
        __nora_step_target_fiber = info;
    }

    NR_MUTEX_LOCK(&g_fiber_list_lock);
    info->next_global = g_fibers_head;
    if (g_fibers_head) g_fibers_head->prev_global = info;
    g_fibers_head = info;
    NR_MUTEX_UNLOCK(&g_fiber_list_lock);

    long old_count = NR_ATOMIC_ADD(&g_active_fibers, 1);
    info->is_main = (old_count == 0);
#ifdef NR_USE_FIBERS
    info->handle = wasm_cont_new(fiber_wrapper, info);
    queue_push(&g_queue, info);
#else
    if (info->is_main) {
        info->data.fn(info->data.arg);
        g_running = false;
    } else {
        fprintf(stderr, "FATAL: spawn() called but Wasm Stack Switching is disabled in this build.\n");
        exit(1);
    }
#endif
    return (fiber_t)info;
}

int get_worker_id() { return 0; }

#else
// --- LINUX/POSIX RUNTIME (Global Queue) ---
#define _GNU_SOURCE
#include <pthread.h>
#include <ucontext.h>
#include <semaphore.h>
#include <stdatomic.h>
#include <sched.h>
#include <unistd.h>
#include <ctype.h>
#include <sys/sysinfo.h>
#include <errno.h>
#define Sleep(ms) usleep((ms) * 1000)

#define MAX_WORKERS 64
#define QUEUE_SIZE 4096

typedef void* fiber_t;

typedef struct fiber_info {
    ucontext_t context;
    NR_ATOMIC_INT state; // 0: READY, 1: RUNNING, 2: PARKING, 3: PARKED, 4: TERMINATED
    NR_ATOMIC_INT resume_pending;
    bool is_main;
    char* temp_strs[256];
    int temp_count;
    void* stack;
    const char* name;
    const char* file;
    int line;
    volatile bool yield_pending;
    struct fiber_info* next_global;
    struct fiber_info* prev_global;
    void* parent;
    jmp_buf panic_buf;
    const char* panic_msg;
    const char* panic_file;
    int panic_line;
} fiber_info_t;

typedef struct fiber_node {
    fiber_info_t* fiber;
    struct fiber_node* next;
} fiber_node_t;

typedef struct {
    fiber_node_t* head;
    fiber_node_t* tail;
    int count;
    NR_MUTEX_T lock;
} global_queue_t;

void park();
void resume(fiber_info_t* info);
_Noreturn void nr_panic(const char* msg, const char* file, int line);
void nr_fiber_report();

typedef struct timer_waiter {
    double wake_time;
    fiber_info_t* fiber;
    struct timer_waiter* next;
} timer_waiter_t;

static timer_waiter_t* g_timer_waiters_head = NULL;
static NR_MUTEX_T g_timer_waiters_lock;
static bool g_timer_poller_running = false;
pthread_t g_timer_thread;

extern NR_ATOMIC_INT g_timer_waiters_count;
double nr_time_now(void);

static void* nr_timer_poller_thread(void* arg) {
    while (g_timer_poller_running) {
        double now = nr_time_now();
        NR_MUTEX_LOCK(&g_timer_waiters_lock);
        timer_waiter_t* curr = g_timer_waiters_head;
        timer_waiter_t* prev = NULL;
        while (curr) {
            if (now >= curr->wake_time) {
                NR_ATOMIC_SUB(&g_timer_waiters_count, 1);
                resume(curr->fiber);
                timer_waiter_t* next = curr->next;
                if (prev) {
                    prev->next = next;
                } else {
                    g_timer_waiters_head = next;
                }
                free(curr);
                curr = next;
            } else {
                prev = curr;
                curr = curr->next;
            }
        }
        NR_MUTEX_UNLOCK(&g_timer_waiters_lock);
        usleep(1000);
    }
    return NULL;
}

void nr_time_init(void) {
    static bool timer_poller_initialized = false;
    if (!timer_poller_initialized) {
        NR_MUTEX_INIT(&g_timer_waiters_lock);
        g_timer_poller_running = true;
        pthread_create(&g_timer_thread, NULL, nr_timer_poller_thread, NULL);
        timer_poller_initialized = true;
    }
}

#define DEQUE_CAPACITY 65536

#include <semaphore.h>
extern sem_t g_worker_sem;

typedef struct {
    NR_ATOMIC_INT bottom;
    NR_ATOMIC_INT top;
    fiber_info_t* volatile buffer[DEQUE_CAPACITY];
} deque_t;

void deque_init(deque_t* q) {
    NR_ATOMIC_STORE(&q->bottom, 0);
    NR_ATOMIC_STORE(&q->top, 0);
    memset((void*)q->buffer, 0, sizeof(q->buffer));
}

void deque_push(deque_t* q, fiber_info_t* f) {
    int b = (int)NR_ATOMIC_LOAD(&q->bottom);
    int t = (int)NR_ATOMIC_LOAD(&q->top);
    if (b - t >= DEQUE_CAPACITY) {
        return;
    }
    q->buffer[b & (DEQUE_CAPACITY - 1)] = f;
    // Sequential consistency release write
    NR_ATOMIC_STORE(&q->bottom, b + 1);
    sem_post(&g_worker_sem);
}

fiber_info_t* deque_pop(deque_t* q) {
    int b = (int)NR_ATOMIC_LOAD(&q->bottom) - 1;
    NR_ATOMIC_STORE(&q->bottom, b);
    
    // CRITICAL: Memory barrier prevents bottom decrement being reordered after top load
    __sync_synchronize();
    
    int t = (int)NR_ATOMIC_LOAD(&q->top);
    if (b > t) {
        // At least one element left safely, no conflict
        return q->buffer[b & (DEQUE_CAPACITY - 1)];
    }
    if (b == t) {
        // Exactly one element left, conflict with concurrent stealers
        fiber_info_t* f = q->buffer[b & (DEQUE_CAPACITY - 1)];
        int expected_t = t;
        if (NR_ATOMIC_CAS(&q->top, &expected_t, t + 1)) {
            // CAS succeeded, we popped it
            NR_ATOMIC_STORE(&q->bottom, t + 1);
            return f;
        }
        // Lost race to a stealer
        NR_ATOMIC_STORE(&q->bottom, t + 1);
        return NULL;
    }
    // Already empty
    NR_ATOMIC_STORE(&q->bottom, t);
    return NULL;
}

fiber_info_t* deque_steal(deque_t* q) {
    int t = (int)NR_ATOMIC_LOAD(&q->top);
    __sync_synchronize();
    int b = (int)NR_ATOMIC_LOAD(&q->bottom);
    if (t >= b) {
        return NULL;
    }
    fiber_info_t* f = q->buffer[t & (DEQUE_CAPACITY - 1)];
    int expected_t = t;
    if (NR_ATOMIC_CAS(&q->top, &expected_t, t + 1)) {
        return f;
    }
    return NULL;
}

global_queue_t g_queue;
deque_t g_local_queues[MAX_WORKERS];
int num_workers = 0;
__thread int worker_id = -1;
__thread int g_yield_ticks = 0;
ucontext_t main_contexts[MAX_WORKERS];
fiber_info_t main_fiber_infos[MAX_WORKERS];
NR_ATOMIC_LONG g_active_fibers = 0;
volatile bool g_running = true;
fiber_info_t* g_fibers_head = NULL;
fiber_info_t* g_terminated_fibers_head = NULL;
NR_MUTEX_T g_fiber_list_lock;

sem_t g_worker_sem;
NR_ATOMIC_INT g_sleeping_workers = 0;
pthread_t g_worker_threads[MAX_WORKERS];

void queue_init(global_queue_t* q) {
    q->head = q->tail = NULL;
    q->count = 0;
    NR_MUTEX_INIT(&q->lock);
}

void queue_push(global_queue_t* q, fiber_info_t* f) {
    fiber_node_t* node = (fiber_node_t*)malloc(sizeof(fiber_node_t));
    node->fiber = f;
    node->next = NULL;
    NR_MUTEX_LOCK(&q->lock);
    if (q->tail) {
        q->tail->next = node;
        q->tail = node;
    } else {
        q->head = q->tail = node;
    }
    q->count++;
    NR_MUTEX_UNLOCK(&q->lock);
    sem_post(&g_worker_sem);
}

void* queue_pop(global_queue_t* q) {
    fiber_info_t* f = NULL;
    NR_MUTEX_LOCK(&q->lock);
    if (q->head) {
        fiber_node_t* node = q->head;
        f = node->fiber;
        q->head = node->next;
        if (!q->head) q->tail = NULL;
        q->count--;
        free(node);
    }
    NR_MUTEX_UNLOCK(&q->lock);
    return f;
}

static fiber_info_t* g_current_fiber[MAX_WORKERS];

void* worker_loop(void* arg) {
    worker_id = (int)(intptr_t)arg;
    
    // Processor Affinity (Linux) [Disabled for dynamic OS thread scheduling]
    // cpu_set_t cpuset;
    // CPU_ZERO(&cpuset);
    // CPU_SET(worker_id % get_nprocs(), &cpuset);
    // pthread_setaffinity_np(pthread_self(), sizeof(cpu_set_t), &cpuset);

    atomic_store(&main_fiber_infos[worker_id].state, 0);
    main_fiber_infos[worker_id].is_main = true;

    while (g_running) {
        fiber_info_t* info = NULL;
        if (worker_id >= 0 && worker_id < num_workers) {
            info = deque_pop(&g_local_queues[worker_id]);
        }
        if (!info) {
            info = (fiber_info_t*)queue_pop(&g_queue);
        }
        if (!info && worker_id >= 0) {
            for (int j = 1; j < num_workers; j++) {
                int victim = (worker_id + j) % num_workers;
                info = deque_steal(&g_local_queues[victim]);
                if (info) break;
            }
        }

        NR_DEBUG_CHECK_LOCK(info);
        if (info) {
            NR_ATOMIC_STORE(&info->state, 1); // RUNNING
            g_current_fiber[worker_id] = info;
            swapcontext(&main_contexts[worker_id], &info->context);
            
            int s = NR_ATOMIC_LOAD(&info->state);
            if (s == 2) { // PARKING
                NR_ATOMIC_STORE(&info->state, 3); // PARKED
                if (info->yield_pending) {
                    resume(info);
                } else {
                    while (1) {
                        int rp = NR_ATOMIC_LOAD(&info->resume_pending);
                        if (rp <= 0) break;
                        int new_rp = rp - 1;
                        if (NR_ATOMIC_CAS(&info->resume_pending, &rp, new_rp)) {
                            resume(info);
                            break;
                        }
                    }
                }
            } else if (s == 4) { // TERMINATED
                // Fiber is done.
            }
        } else {
            if (__nora_debug_locked_fiber) {
                nr_sleep_ms(1);
                continue;
            }
            if (!g_running) break;
            if (NR_ATOMIC_LOAD(&g_active_fibers) == 0 && worker_id == 0) break;
            NR_ATOMIC_INC(&g_sleeping_workers);
            if (NR_ATOMIC_LOAD(&g_sleeping_workers) == num_workers && NR_ATOMIC_LOAD(&g_active_fibers) > 0) {
                NR_MUTEX_LOCK(&g_fiber_list_lock);
                fiber_info_t* curr = g_fibers_head;
                bool any_runnable = false;
                while (curr) {
                    long s = NR_ATOMIC_LOAD(&curr->state);
                    if (s == 0 || s == 1 || s == 2) {
                        any_runnable = true;
                        break;
                    }
                    curr = curr->next_global;
                }
                NR_MUTEX_UNLOCK(&g_fiber_list_lock);
                
                if (!any_runnable && NR_ATOMIC_LOAD(&g_net_waiters_count) == 0 && NR_ATOMIC_LOAD(&g_timer_waiters_count) == 0) {
                    printf("\nFATAL: Deadlock detected! All %ld fibers are blocked.\n", (long)NR_ATOMIC_LOAD(&g_active_fibers));
                    nr_fiber_report();
                    nr_panic("deadlock", "runtime", 0);
                }
            }
            sem_wait(&g_worker_sem);
            NR_ATOMIC_DEC(&g_sleeping_workers);
        }
    }
    return NULL;
}

void scheduler_init() {
    num_workers = get_nprocs();
    char* env_workers = getenv("NORA_NUM_WORKERS");
    if (env_workers != NULL) {
        int parsed = atoi(env_workers);
        if (parsed >= 1 && parsed <= MAX_WORKERS) {
            num_workers = parsed;
        }
    } else {
        if (num_workers > MAX_WORKERS) num_workers = MAX_WORKERS;
        if (num_workers < 1) num_workers = 1;
        // For stability on WSL, don't use too many workers
        if (num_workers > 8) num_workers = 8;
    }
    setvbuf(stdout, NULL, _IONBF, 0);
    // printf("Using %d workers\n", num_workers);

    sem_init(&g_worker_sem, 0, 0);
    NR_MUTEX_INIT(&g_fiber_list_lock);
    queue_init(&g_queue);
    for (int i = 0; i < MAX_WORKERS; i++) {
        deque_init(&g_local_queues[i]);
    }

    for (int i = 1; i < num_workers; i++) {
        pthread_create(&g_worker_threads[i], NULL, worker_loop, (void*)(intptr_t)i);
    }
    
    worker_id = 0;
    NR_ATOMIC_STORE(&main_fiber_infos[0].state, 4);
    NR_ATOMIC_STORE(&main_fiber_infos[0].resume_pending, 0);
    main_fiber_infos[0].is_main = true;
    main_fiber_infos[0].temp_count = 0;
    
    // cpu_set_t cpuset;
    // CPU_ZERO(&cpuset);
    // CPU_SET(0, &cpuset);
    // pthread_setaffinity_np(pthread_self(), sizeof(cpu_set_t), &cpuset);

    main_fiber_infos[0].name = "main";
    main_fiber_infos[0].file = "main";
    main_fiber_infos[0].line = 0;
    NR_MUTEX_LOCK(&g_fiber_list_lock);
    main_fiber_infos[0].next_global = g_fibers_head;
    if (g_fibers_head) g_fibers_head->prev_global = &main_fiber_infos[0];
    g_fibers_head = &main_fiber_infos[0];
    NR_MUTEX_UNLOCK(&g_fiber_list_lock);
}

void scheduler_run_loop() {
    while (g_running) {
        fiber_info_t* info = NULL;
        int id = worker_id;
        if (id >= 0 && id < num_workers) {
            info = deque_pop(&g_local_queues[id]);
        }
        if (!info) {
            info = (fiber_info_t*)queue_pop(&g_queue);
        }
        if (!info && id >= 0) {
            for (int j = 1; j < num_workers; j++) {
                int victim = (id + j) % num_workers;
                info = deque_steal(&g_local_queues[victim]);
                if (info) break;
            }
        }

        NR_DEBUG_CHECK_LOCK(info);
        if (info) {
            NR_ATOMIC_STORE(&info->state, 1); // RUNNING
            g_current_fiber[worker_id] = info;
            swapcontext(&main_contexts[worker_id], &info->context);
            
            int s = NR_ATOMIC_LOAD(&info->state);
            if (s == 2) { // PARKING
                NR_ATOMIC_STORE(&info->state, 3); // PARKED
                if (info->yield_pending) {
                    resume(info);
                } else {
                    while (1) {
                        int rp = NR_ATOMIC_LOAD(&info->resume_pending);
                        if (rp <= 0) break;
                        int new_rp = rp - 1;
                        if (NR_ATOMIC_CAS(&info->resume_pending, &rp, new_rp)) {
                            resume(info);
                            break;
                        }
                    }
                }
            } else if (s == 4) { // TERMINATED
                // Fiber is done.
            }
        } else {
            if (__nora_debug_locked_fiber) {
                nr_sleep_ms(1);
                continue;
            }
            if (NR_ATOMIC_LOAD(&g_active_fibers) == 0) break;
            
            NR_ATOMIC_INC(&g_sleeping_workers);
            if (NR_ATOMIC_LOAD(&g_sleeping_workers) == num_workers && NR_ATOMIC_LOAD(&g_active_fibers) > 0) {
                NR_MUTEX_LOCK(&g_fiber_list_lock);
                fiber_info_t* curr = g_fibers_head;
                bool any_runnable = false;
                while (curr) {
                    long s = NR_ATOMIC_LOAD(&curr->state);
                    if (s == 0 || s == 1 || s == 2) {
                        any_runnable = true;
                        break;
                    }
                    curr = curr->next_global;
                }
                NR_MUTEX_UNLOCK(&g_fiber_list_lock);
                
                if (!any_runnable && NR_ATOMIC_LOAD(&g_net_waiters_count) == 0 && NR_ATOMIC_LOAD(&g_timer_waiters_count) == 0) {
                    printf("\nFATAL: Deadlock detected! All %ld fibers are blocked. g_net_waiters_count: %d\n", (long)NR_ATOMIC_LOAD(&g_active_fibers), (int)NR_ATOMIC_LOAD(&g_net_waiters_count));
                    nr_fiber_report();
                    nr_panic("deadlock", "runtime", 0);
                }
            }
            sem_wait(&g_worker_sem);
            NR_ATOMIC_DEC(&g_sleeping_workers);
        }
    }
}

void scheduler_cleanup() {
    g_running = false;
    for (int i = 0; i < num_workers; i++) sem_post(&g_worker_sem);
    for (int i = 1; i < num_workers; i++) {
        pthread_join(g_worker_threads[i], NULL);
    }
    if (g_timer_poller_running) {
        g_timer_poller_running = false;
        pthread_join(g_timer_thread, NULL);
    }
    sem_destroy(&g_worker_sem);
    NR_MUTEX_DESTROY(&g_queue.lock);

    // Free all active fibers
    NR_MUTEX_LOCK(&g_fiber_list_lock);
    fiber_info_t* curr = g_fibers_head;
    while (curr) {
        fiber_info_t* next = curr->next_global;
        if (!curr->is_main) {
            free(curr);
        }
        curr = next;
    }
    g_fibers_head = NULL;

    // Free all terminated fibers
    curr = g_terminated_fibers_head;
    while (curr) {
        fiber_info_t* next = curr->next_global;
        free(curr);
        curr = next;
    }
    g_terminated_fibers_head = NULL;
    NR_MUTEX_UNLOCK(&g_fiber_list_lock);
    NR_MUTEX_DESTROY(&g_fiber_list_lock);
}

void* GetFiberData() {
    if (worker_id < 0 || worker_id >= MAX_WORKERS) return NULL;
    return g_current_fiber[worker_id];
}

void park() {
    fiber_info_t* info = (fiber_info_t*)GetFiberData();
    if (info) {
        int s = atomic_load(&info->state);
        if (s == 1) atomic_store(&info->state, 2); // PARKING
    }
    if (worker_id >= 0) swapcontext(&info->context, &main_contexts[worker_id]);
}

void nr_cooperative_yield() {
    fiber_info_t* self = (fiber_info_t*)GetFiberData();
    if (!self) return;
    self->yield_pending = true;
    park();
}

void resume(fiber_info_t* info) {
    if (!info) return;

    bool is_yield = false;
    if (info->yield_pending) {
        is_yield = true;
        info->yield_pending = false;
    }

    int expected = 3; // PARKED
    if (atomic_compare_exchange_strong(&info->state, &expected, 0)) { // PARKED -> READY
        int id = worker_id;
        if (id >= 0 && id < num_workers && num_workers > 1 && !is_yield) {
            deque_push(&g_local_queues[id], info);
        } else {
            queue_push(&g_queue, info);
        }
        sem_post(&g_worker_sem);
    } else {
        atomic_fetch_add(&info->resume_pending, 1);
        expected = 3;
        if (atomic_compare_exchange_strong(&info->state, &expected, 0)) {
            if (atomic_fetch_sub(&info->resume_pending, 1) > 0) {
                int id = worker_id;
                if (id >= 0 && id < num_workers && num_workers > 1 && !is_yield) {
                    deque_push(&g_local_queues[id], info);
                } else {
                    queue_push(&g_queue, info);
                }
                sem_post(&g_worker_sem);
            }
        }
    }
}

int get_worker_id() { return worker_id; }

typedef struct {
    void (*fn)(void*);
    void* arg;
} spawn_data_t;

void fiber_wrapper() {
    fiber_info_t* info = g_current_fiber[worker_id];
    spawn_data_t* data = (spawn_data_t*)(info + 1);
    data->fn(data->arg);
    nr_flush_temps();
    atomic_fetch_sub(&g_active_fibers, 1);
    
    NR_MUTEX_LOCK(&g_fiber_list_lock);
    if (info->prev_global) info->prev_global->next_global = info->next_global;
    if (info->next_global) info->next_global->prev_global = info->prev_global;
    if (info == g_fibers_head) g_fibers_head = info->next_global;

    // Push to terminated list
    info->next_global = g_terminated_fibers_head;
    info->prev_global = NULL;
    if (g_terminated_fibers_head) g_terminated_fibers_head->prev_global = info;
    g_terminated_fibers_head = info;
    NR_MUTEX_UNLOCK(&g_fiber_list_lock);

    if (info->is_main) {
        // printf("Main fiber finished\n");
        g_running = false;
        for (int i = 0; i < num_workers; i++) sem_post(&g_worker_sem);
    }
    atomic_store(&info->state, 4); // TERMINATED
    park();
}

void* scheduler_spawn(void (*fn)(void*), void* arg, const char* name, const char* file, int line) {
    size_t info_size = (sizeof(fiber_info_t) + sizeof(spawn_data_t) + 15) & ~15;
    fiber_info_t* info = (fiber_info_t*)malloc(info_size + 1024 * 1024);
    if (!info) {
        fprintf(stderr, "FATAL: scheduler_spawn failed to allocate fiber\n");
        exit(1);
    }
    memset(info, 0, info_size);
    atomic_store(&info->state, 0); // READY
    atomic_store(&info->resume_pending, 0);
    info->temp_count = 0;
    info->stack = (char*)info + info_size;
    info->name = name;
    info->file = file;
    info->line = line;
    info->parent = nr_fiber_current();

    if (__nora_stepping_into_spawn) {
        __nora_stepping_into_spawn = 0;
        __nora_step_target_fiber = info;
    }

    NR_MUTEX_LOCK(&g_fiber_list_lock);
    info->next_global = g_fibers_head;
    if (g_fibers_head) g_fibers_head->prev_global = info;
    g_fibers_head = info;
    NR_MUTEX_UNLOCK(&g_fiber_list_lock);
    
    long old_count = atomic_fetch_add(&g_active_fibers, 1);
    info->is_main = (old_count == 0);
    
    spawn_data_t* data = (spawn_data_t*)(info + 1);
    data->fn = fn;
    data->arg = arg;

    getcontext(&info->context);
    info->context.uc_stack.ss_sp = info->stack;
    info->context.uc_stack.ss_size = 1024 * 1024;
    info->context.uc_link = NULL;
    makecontext(&info->context, fiber_wrapper, 0);

    int id = worker_id;
    if (id >= 0 && id < num_workers && num_workers > 1) {
        deque_push(&g_local_queues[id], info);
    } else {
        queue_push(&g_queue, info);
    }
    sem_post(&g_worker_sem);
    return (fiber_t)info;
}

void* nr_fiber_current() {
    return GetFiberData();
}

void nr_fiber_suspend() {
    park();
}

void nr_fiber_resume(void* f) {
    resume((fiber_info_t*)f);
}

void nr_store_ptr(void* dest, void* val) {
    if (dest) *(void**)dest = val;
}

void nr_store_i32(void* dest, int val) {
    if (dest) *(int*)dest = val;
}

int nr_load_i32(void* p) {
    if (p) return *(volatile int*)p;
    return 0;
}

void* nr_load_ptr(void* p) {
    if (p) return *(void* volatile*)p;
    return NULL;
}

#endif
#else // __EMSCRIPTEN__
#include <emscripten.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdio.h>

#define MAX_WORKERS 1
#define QUEUE_SIZE 4096

// Asyncify imports
extern void asyncify_start_unwind(void*) __attribute__((import_module("env"), import_name("asyncify_start_unwind")));
extern void asyncify_stop_unwind() __attribute__((import_module("env"), import_name("asyncify_stop_unwind")));
extern void asyncify_start_rewind(void*) __attribute__((import_module("env"), import_name("asyncify_start_rewind")));
extern void asyncify_stop_rewind() __attribute__((import_module("env"), import_name("asyncify_stop_rewind")));
extern int asyncify_get_state() __attribute__((import_module("env"), import_name("asyncify_get_state")));

typedef void* fiber_t;

typedef struct {
    void* stack_ptr;
    void* stack_limit;
} asyncify_buffer_t;

typedef struct fiber_info {
    asyncify_buffer_t asyncify_buf;
    NR_ATOMIC_INT state; // 0: READY, 1: RUNNING, 2: PARKING, 3: PARKED, 4: TERMINATED
    NR_ATOMIC_INT resume_pending;
    bool is_main;
    char* temp_strs[256];
    int temp_count;
    void* stack;
    void (*fn)(void*);
    void* arg;
    const char* name;
    const char* file;
    int line;
    volatile bool yield_pending;
    struct fiber_info* next_global;
    struct fiber_info* prev_global;
    void* parent;
    jmp_buf panic_buf;
    const char* panic_msg;
    const char* panic_file;
    int panic_line;
} fiber_info_t;


typedef struct fiber_node {
    fiber_info_t* fiber;
    struct fiber_node* next;
} fiber_node_t;

typedef struct {
    fiber_node_t* head;
    fiber_node_t* tail;
    int count;
    NR_MUTEX_T lock;
} global_queue_t;

void park();
void resume(fiber_info_t* info);
_Noreturn void nr_panic(const char* msg, const char* file, int line);
void nr_fiber_report();

global_queue_t g_queue;
fiber_info_t* g_current_fiber = NULL;
NR_ATOMIC_LONG g_active_fibers = 0;
volatile bool g_running = true;
fiber_info_t* g_fibers_head = NULL;
fiber_info_t* g_terminated_fibers_head = NULL;
NR_MUTEX_T g_fiber_list_lock;

void queue_init(global_queue_t* q) {
    q->head = q->tail = NULL;
    q->count = 0;
    NR_MUTEX_INIT(&q->lock);
}

void queue_push(global_queue_t* q, fiber_info_t* f) {
    fiber_node_t* node = (fiber_node_t*)malloc(sizeof(fiber_node_t));
    node->fiber = f;
    node->next = NULL;
    NR_MUTEX_LOCK(&q->lock);
    if (q->tail) {
        q->tail->next = node;
        q->tail = node;
    } else {
        q->head = q->tail = node;
    }
    q->count++;
    NR_MUTEX_UNLOCK(&q->lock);
}

void* queue_pop(global_queue_t* q) {
    fiber_info_t* f = NULL;
    NR_MUTEX_LOCK(&q->lock);
    if (q->head) {
        fiber_node_t* node = q->head;
        f = node->fiber;
        q->head = node->next;
        if (!q->head) q->tail = NULL;
        q->count--;
        free(node);
    }
    NR_MUTEX_UNLOCK(&q->lock);
    return f;
}

void fiber_entry(fiber_info_t* info);

void scheduler_init() {
    queue_init(&g_queue);
    NR_MUTEX_INIT(&g_fiber_list_lock);
    g_active_fibers = 0;
    g_running = true;
}

void scheduler_run_loop() {
    while (g_running) {
        if (asyncify_get_state() != 0) {
            // We are in the middle of an unwind/rewind, shouldn't happen here
            return;
        }

        fiber_info_t* info = (fiber_info_t*)queue_pop(&g_queue);
        if (info) {
            g_current_fiber = info;
            NR_ATOMIC_STORE(&info->state, 1); // RUNNING
            fiber_entry(info);
            
            if (asyncify_get_state() == 1) { // UNWINDING
                asyncify_stop_unwind();
                if (NR_ATOMIC_LOAD(&info->state) == 2) { // PARKING
                    NR_ATOMIC_STORE(&info->state, 3); // PARKED
                    if (NR_ATOMIC_LOAD(&info->resume_pending) > 0) {
                        NR_ATOMIC_STORE(&info->resume_pending, 0);
                        resume(info);
                    }
                }
            } else {
                // Fiber finished normally
                NR_ATOMIC_STORE(&info->state, 4); // TERMINATED
            }
        } else {
            if (NR_ATOMIC_LOAD(&g_active_fibers) == 0) break;
            
            // Check for deadlock
            NR_MUTEX_LOCK(&g_fiber_list_lock);
            fiber_info_t* curr = g_fibers_head;
            bool any_runnable = false;
            while (curr) {
                long s = NR_ATOMIC_LOAD(&curr->state);
                if (s == 0 || s == 1 || s == 2) { // READY, RUNNING, or PARKING
                    any_runnable = true;
                    break;
                }
                curr = curr->next_global;
            }
            NR_MUTEX_UNLOCK(&g_fiber_list_lock);
            
            if (!any_runnable && NR_ATOMIC_LOAD(&g_net_waiters_count) == 0 && NR_ATOMIC_LOAD(&g_timer_waiters_count) == 0) {
                printf("\nFATAL: Deadlock detected!\n");
                nr_fiber_report();
                nr_panic("deadlock", "runtime", 0);
            }
            break; 
        }
    }
}

void scheduler_cleanup() {
    g_running = false;

    // Free all active fibers
    NR_MUTEX_LOCK(&g_fiber_list_lock);
    fiber_info_t* curr = g_fibers_head;
    while (curr) {
        fiber_info_t* next = curr->next_global;
        if (!curr->is_main) {
            free(curr->asyncify_buf.stack_ptr);
            free(curr);
        }
        curr = next;
    }
    g_fibers_head = NULL;

    // Free all terminated fibers
    curr = g_terminated_fibers_head;
    while (curr) {
        fiber_info_t* next = curr->next_global;
        free(curr->asyncify_buf.stack_ptr);
        free(curr);
        curr = next;
    }
    g_terminated_fibers_head = NULL;
    NR_MUTEX_UNLOCK(&g_fiber_list_lock);
    NR_MUTEX_DESTROY(&g_fiber_list_lock);
}

void* GetFiberData() {
    return g_current_fiber;
}

void park() {
    if (g_current_fiber && NR_ATOMIC_LOAD(&g_current_fiber->state) == 1) {
        NR_ATOMIC_STORE(&g_current_fiber->state, 2); // PARKING
        asyncify_start_unwind(&g_current_fiber->asyncify_buf);
    }
}

void resume(fiber_info_t* info) {
    if (!info || NR_ATOMIC_LOAD(&info->state) == 4) return;
    if (NR_ATOMIC_LOAD(&info->state) == 3) {
        NR_ATOMIC_STORE(&info->state, 0); // READY
        queue_push(&g_queue, info);
    } else {
        NR_ATOMIC_INC(&info->resume_pending);
    }
}

void fiber_entry(fiber_info_t* info) {
    if (asyncify_get_state() == 2) { // REWINDING
        asyncify_stop_rewind();
        // The call stack will be restored and execution will continue from where it left off
    } else {
        info->fn(info->arg);
        nr_flush_temps();
        NR_ATOMIC_DEC(&g_active_fibers);

        NR_MUTEX_LOCK(&g_fiber_list_lock);
        if (info->prev_global) info->prev_global->next_global = info->next_global;
        if (info->next_global) info->next_global->prev_global = info->prev_global;
        if (info == g_fibers_head) g_fibers_head = info->next_global;

        // Push to terminated list
        info->next_global = g_terminated_fibers_head;
        info->prev_global = NULL;
        if (g_terminated_fibers_head) g_terminated_fibers_head->prev_global = info;
        g_terminated_fibers_head = info;
        NR_MUTEX_UNLOCK(&g_fiber_list_lock);

        if (info->is_main) g_running = false;
    }
}

void* scheduler_spawn(void (*fn)(void*), void* arg, const char* name, const char* file, int line) {
    fiber_info_t* info = (fiber_info_t*)malloc(sizeof(fiber_info_t));
    memset(info, 0, sizeof(fiber_info_t));
    NR_ATOMIC_STORE(&info->state, 0); // READY
    info->fn = fn;
    info->arg = arg;
    info->name = name;
    info->file = file;
    info->line = line;
    info->parent = nr_fiber_current();

    if (__nora_stepping_into_spawn) {
        __nora_stepping_into_spawn = 0;
        __nora_step_target_fiber = info;
    }
    info->asyncify_buf.stack_ptr = malloc(16384); // Asyncify buffer
    info->asyncify_buf.stack_limit = (char*)info->asyncify_buf.stack_ptr + 16384;

    NR_MUTEX_LOCK(&g_fiber_list_lock);
    info->next_global = g_fibers_head;
    if (g_fibers_head) g_fibers_head->prev_global = info;
    g_fibers_head = info;
    NR_MUTEX_UNLOCK(&g_fiber_list_lock);
    
    bool is_main = (NR_ATOMIC_LOAD(&g_active_fibers) == 0);
    NR_ATOMIC_INC(&g_active_fibers);
    info->is_main = is_main;
    
    queue_push(&g_queue, info);
    return (fiber_t)info;
}

void* nr_fiber_current() {
    return GetFiberData();
}

void nr_fiber_suspend() {
    park();
}

void nr_fiber_resume(void* f) {
    resume((fiber_info_t*)f);
}

void nr_store_ptr(void* dest, void* val) {
    if (dest) *(void**)dest = val;
}

void nr_store_i32(void* dest, int val) {
    if (dest) *(int*)dest = val;
}

int nr_load_i32(void* p) {
    if (p) return *(volatile int*)p;
    return 0;
}

void* nr_load_ptr(void* p) {
    if (p) return *(void* volatile*)p;
    return NULL;
}

#endif

void* nr_fiber_parent() {
    fiber_info_t* curr = (fiber_info_t*)nr_fiber_current();
    if (curr) return curr->parent;
    return NULL;
}

jmp_buf* nr_fiber_panic_buf_ptr(void* f) {
    if (!f) return NULL;
    return &((fiber_info_t*)f)->panic_buf;
}

const char* nr_fiber_panic_msg(void* f) {
    if (!f) return NULL;
    return ((fiber_info_t*)f)->panic_msg;
}

const char* nr_fiber_panic_file(void* f) {
    if (!f) return NULL;
    return ((fiber_info_t*)f)->panic_file;
}

int nr_fiber_panic_line(void* f) {
    if (!f) return 0;
    return ((fiber_info_t*)f)->panic_line;
}

