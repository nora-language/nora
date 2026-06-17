// Generated runtime header for Nora
#ifndef NORA_RUNTIME_H
#define NORA_RUNTIME_H

#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>
#include <string.h>
#include <ctype.h>
#include <stdint.h>
#include <stdarg.h>
#include <time.h>
#include <math.h>
#include <setjmp.h>

#ifdef _WIN32
#define WIN32_LEAN_AND_MEAN
#include <winsock2.h>
#include <ws2tcpip.h>
#include <windows.h>
#include <direct.h>
#elif defined(__EMSCRIPTEN__)
#include <emscripten.h>
#else
#include <pthread.h>
#include <stdatomic.h>
#include <unistd.h>
#endif

// Base macros
// --- BASE MACROS ---
#ifdef _WIN32
#define NR_ATOMIC_INT volatile long
#define NR_ATOMIC_LONG volatile long long
#define NR_INT long
#define NR_LONG long long
#define NR_ATOMIC_INC(p) InterlockedIncrement((volatile long*)(p))
#define NR_ATOMIC_DEC(p) InterlockedDecrement((volatile long*)(p))
#define NR_ATOMIC_ADD(p, v) InterlockedExchangeAdd((volatile long*)(p), (long)(v))
#define NR_ATOMIC_SUB(p, v) InterlockedExchangeAdd((volatile long*)(p), -(long)(v))
#define NR_ATOMIC_LOAD(p) InterlockedCompareExchange((volatile long*)(p), 0, 0)
#define NR_ATOMIC_STORE(p, v) InterlockedExchange((volatile long*)(p), (long)(v))
#define NR_ATOMIC_EXCHANGE(p, v) InterlockedExchange((volatile long*)(p), (long)(v))

static inline bool nr_atomic_cas_win32(volatile long* p, long* exp, long des) {
    long old = InterlockedCompareExchange(p, des, *exp);
    bool success = (old == *exp);
    if (!success) *exp = old;
    return success;
}
#define NR_ATOMIC_CAS(p, exp, des) nr_atomic_cas_win32((volatile long*)(p), (long*)(exp), (long)(des))

#define NR_MUTEX_T CRITICAL_SECTION
#define NR_MUTEX_INIT(l) InitializeCriticalSection(l)
#define NR_MUTEX_LOCK(l) EnterCriticalSection(l)
#define NR_MUTEX_UNLOCK(l) LeaveCriticalSection(l)
#define NR_MUTEX_DESTROY(l) DeleteCriticalSection(l)
#include <dbghelp.h>

#elif defined(__EMSCRIPTEN__)
#define NR_ATOMIC_INT int
#define NR_ATOMIC_LONG long long
#define NR_INT int
#define NR_LONG long long
#define NR_ATOMIC_INC(p) (++(*(p)))
#define NR_ATOMIC_DEC(p) (--(*(p)))
#define NR_ATOMIC_LOAD(p) (*(p))
#define NR_ATOMIC_STORE(p, v) (*(p)) = (v)

static inline int nr_atomic_add_emscripten(int* p, int v) {
    int old = *p;
    *p += v;
    return old;
}
#define NR_ATOMIC_ADD(p, v) nr_atomic_add_emscripten(p, v)

static inline int nr_atomic_sub_emscripten(int* p, int v) {
    int old = *p;
    *p -= v;
    return old;
}
#define NR_ATOMIC_SUB(p, v) nr_atomic_sub_emscripten(p, v)

static inline int nr_atomic_exchange_emscripten(int* p, int v) {
    int old = *p;
    *p = v;
    return old;
}
#define NR_ATOMIC_EXCHANGE(p, v) nr_atomic_exchange_emscripten(p, v)

static inline bool nr_atomic_cas_emscripten(int* p, int* exp, int des) {
    bool success = (*p == *exp);
    if (success) *p = des;
    else *exp = *p;
    return success;
}
#define NR_ATOMIC_CAS(p, exp, des) nr_atomic_cas_emscripten(p, exp, des)

#define NR_MUTEX_T int
#define NR_MUTEX_INIT(l) (*(l)) = 0
#define NR_MUTEX_LOCK(l) (void)(l)
#define NR_MUTEX_UNLOCK(l) (void)(l)
#define NR_MUTEX_DESTROY(l) (void)(l)

#elif defined(__wasm__)
#define NR_ATOMIC_INT int
#define NR_ATOMIC_LONG long long
#define NR_INT int
#define NR_LONG long long
#define NR_ATOMIC_INC(p) __atomic_fetch_add(p, 1, __ATOMIC_SEQ_CST)
#define NR_ATOMIC_DEC(p) __atomic_fetch_sub(p, 1, __ATOMIC_SEQ_CST)
#define NR_ATOMIC_ADD(p, v) __atomic_fetch_add(p, v, __ATOMIC_SEQ_CST)
#define NR_ATOMIC_SUB(p, v) __atomic_fetch_sub(p, v, __ATOMIC_SEQ_CST)
#define NR_ATOMIC_LOAD(p) __atomic_load_n(p, __ATOMIC_SEQ_CST)
#define NR_ATOMIC_STORE(p, v) __atomic_store_n(p, v, __ATOMIC_SEQ_CST)
#define NR_ATOMIC_EXCHANGE(p, v) __atomic_exchange_n(p, v, __ATOMIC_SEQ_CST)
#define NR_ATOMIC_CAS(p, exp, des) __atomic_compare_exchange_n(p, exp, des, false, __ATOMIC_SEQ_CST, __ATOMIC_SEQ_CST)

#define NR_MUTEX_T int
#define NR_MUTEX_INIT(l) (*(l)) = 0
static inline void nr_wasm_mutex_lock(int* l) {
    int expected = 0;
    while (!__atomic_compare_exchange_n(l, &expected, 1, false, __ATOMIC_ACQUIRE, __ATOMIC_RELAXED)) {
        expected = 0;
    }
}
static inline void nr_wasm_mutex_unlock(int* l) {
    __atomic_store_n(l, 0, __ATOMIC_RELEASE);
}
#define NR_MUTEX_LOCK(l) nr_wasm_mutex_lock(l)
#define NR_MUTEX_UNLOCK(l) nr_wasm_mutex_unlock(l)
#define NR_MUTEX_DESTROY(l) (void)(l)

#else
#include <stdatomic.h>
#define NR_ATOMIC_INT atomic_int
#define NR_ATOMIC_LONG atomic_long
#define NR_INT int
#define NR_LONG long
#define NR_ATOMIC_INC(p) atomic_fetch_add(p, 1)
#define NR_ATOMIC_DEC(p) atomic_fetch_sub(p, 1)
#define NR_ATOMIC_ADD(p, v) atomic_fetch_add(p, v)
#define NR_ATOMIC_SUB(p, v) atomic_fetch_sub(p, v)
#define NR_ATOMIC_LOAD(p) atomic_load(p)
#define NR_ATOMIC_STORE(p, v) atomic_store(p, v)
#define NR_ATOMIC_EXCHANGE(p, v) atomic_exchange(p, v)
#define NR_ATOMIC_CAS(p, exp, des) atomic_compare_exchange_strong(p, exp, des)

#define NR_MUTEX_T pthread_mutex_t
#define NR_MUTEX_INIT(l) pthread_mutex_init(l, NULL)
#define NR_MUTEX_LOCK(l) pthread_mutex_lock(l)
#define NR_MUTEX_UNLOCK(l) pthread_mutex_unlock(l)
#define NR_MUTEX_DESTROY(l) pthread_mutex_destroy(l)
#include <execinfo.h>
#include <dlfcn.h>
#endif

#ifdef _WIN32
#ifdef _MSC_VER
#define THREAD_LOCAL __declspec(thread)
#else
#define THREAD_LOCAL __thread
#endif
#elif defined(__EMSCRIPTEN__)
#define THREAD_LOCAL
#elif defined(__wasm__)
#define THREAD_LOCAL
#else
#define THREAD_LOCAL __thread
#endif

#ifdef __cplusplus
extern "C" {
#endif
extern THREAD_LOCAL int g_yield_ticks;
#ifdef __cplusplus
}
#endif

void nr_cooperative_yield();

#define NR_COOPERATIVE_YIELD_CHECKPOINT() do { \
    if (++g_yield_ticks >= 1000) { \
        g_yield_ticks = 0; \
        nr_cooperative_yield(); \
    } \
} while(0)

#define NR_HEADER_MAGIC ((int)0xCAFEBABE)
#define NR_MAGIC_STATIC ((int)0xCAFEBA11)
#define NR_MAGIC_FREE   ((int)0xBAADF00D)

typedef struct nr_header {
    int32_t count;
    int32_t elem_size;
    NR_ATOMIC_INT magic;
    NR_ATOMIC_INT ref_count;
    const char* file;
    int32_t line;
    struct nr_header* next;
    struct nr_header* prev;
} __attribute__((aligned(16))) nr_header_t;

#define NR_HEADER_SIZE (sizeof(nr_header_t))

// Missing Prototypes
#ifdef __cplusplus
extern "C" {
#endif

extern int nr_argc;
extern char** nr_argv;

void* nr_get_stdout(void);
void* nr_get_stderr(void);
void* nr_get_stdin(void);
int nr_atoi(char* s);
double nr_atof(char* s);

char* nr_str_concat(char* s1, char* s2);
char* nr_str_concat_free(char* s1, char* s2, bool f1, bool f2);
char* nr_str_from_cstring(void* s);
bool nr_str_eq(char* s1, char* s2);
char* nr_temp_str(char* s);
char* nr_claim_str(char* s);
void nr_flush_temps();
void nr_free_str(void* ptr);

void nr_mem_init();
void nr_mem_report();
void scheduler_init();
void* scheduler_spawn(void (*func)(void*), void* data, const char* name, const char* file, int line);
void* nr_fiber_current(void);
void* nr_fiber_parent(void);
void nr_fiber_suspend(void);
void nr_fiber_resume(void* f);

// Debugger causality variables for premium stepping
extern volatile void* __nora_step_target_fiber;
extern volatile int __nora_stepping_into_spawn;
extern volatile void* __nora_debug_locked_fiber;

#ifdef _WIN32
#define NR_DEBUGBREAK() DebugBreak()
#else
#define NR_DEBUGBREAK() __builtin_trap()
#endif

// Debugger hook for fiber start
#ifdef _WIN32
__declspec(noinline) void __nora_fiber_started(void* parent);
#else
__attribute__((noinline)) void __nora_fiber_started(void* parent);
#endif

void scheduler_run_loop();
void scheduler_cleanup();
extern volatile bool g_running;

void* array_make(int count, int elem_size, const char* file, int line, ...);
void* array_make_empty(int count, int elem_size, const char* file, int line);
void* array_append(void* arr, void* data);
void* array_data(void* data);
int array_count(void* data);
void* array_slice(void* arr, int start, int end, int elem_size);
char* string_slice(char* s, int start, int end);
void* array_bounds_check(void* arr, int index, const char* file, int line);

void* map_make(int key_size, int val_size, bool is_str_key, const char* file, int line);
void map_set(void* _m, void* key, void* val);
void* map_get(void* _m, void* key);
bool map_contains(void* _m, void* key);
void map_free(void* _m);

void* nr_malloc_debug(size_t size, const char* file, int line);
void* nr_malloc_untracked(int size);
void nr_free(void* ptr);
void nr_free_untracked(void* p);
#define nr_malloc(s) nr_malloc_debug(s, __FILE__, __LINE__)

void nr_panic(const char* msg, const char* file, int line) __attribute__((noreturn));
void nr_print_backtrace();
void nr_fiber_report();

typedef void* fiber_t;
struct fiber_info;
void park();
void resume(struct fiber_info* info);

// First-class function closure fat pointer
typedef struct {
    void* env;
    void* fn_ptr;
    void (*drop_fn)(void*);
} nr_closure_t;

void* nr_sync_mutex_create(void);
void nr_sync_mutex_lock(void* m);
void nr_sync_mutex_unlock(void* m);
void nr_sync_mutex_destroy(void* m);

void* nr_sync_rwmutex_create(void);
void nr_sync_rwmutex_rlock(void* m);
void nr_sync_rwmutex_runlock(void* m);
void nr_sync_rwmutex_lock(void* m);
void nr_sync_rwmutex_unlock(void* m);
void nr_sync_rwmutex_destroy(void* m);
void* nr_sync_waitgroup_create(void);
void nr_sync_waitgroup_add(void* w, int32_t delta);
void nr_sync_waitgroup_done(void* w);
// Panic Unwinding Helpers
jmp_buf* nr_fiber_panic_buf_ptr(void* f);
const char* nr_fiber_panic_msg(void* f);
const char* nr_fiber_panic_file(void* f);
int nr_fiber_panic_line(void* f);

void nr_sync_waitgroup_panic(void* wg_ptr, const char* msg, const char* file, int line);

void nr_sync_waitgroup_wait(void* w);
void nr_sync_waitgroup_destroy(void* w);
void nr_sync_atomic_store(void* p, int32_t v);
bool nr_sync_atomic_cas(void* p, void* exp, int32_t des);

void nr_sleep_ms(int32_t ms);
void nr_time_to_parts(double sec, bool utc, void* parts_ptr);

int nr_get_argc(void);
char* nr_get_argv(int index);
char* nr_os_getenv(char* key);
bool nr_os_setenv(char* key, char* val);
bool nr_os_unsetenv(char* key);
char* nr_os_getcwd(void);

bool nr_crypto_rand_bytes(void* buf, int32_t size);
void nr_net_close(int32_t socket_fd);

char* nr_strdup(const char* s);
char* nr_bool_to_str(bool v);

char* nr_i32_to_str(int v);
char* nr_i64_to_str(long long v);
char* nr_f64_to_str(double v);

char* nr_to_str(void* p);

typedef struct channel_s channel_t;

typedef struct {
    channel_t* chan;
    bool is_send;
    void* data;
} select_op_t;

channel_t* channel_make(int capacity, int elem_size);
void channel_send(channel_t* c, void* val);
void channel_recv(channel_t* c, void* res);
void channel_ref(channel_t* c);
void channel_destroy(channel_t* c);
void channel_free(channel_t* c);
int channel_select(select_op_t* ops, int count, bool has_default);

#ifdef __cplusplus
}
#endif

#endif // NORA_RUNTIME_H
