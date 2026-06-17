#include "runtime.h"

// --- PREMIUM TIME RUNTIME ---
double nr_time_now(void) {
#ifdef _WIN32
    FILETIME ft;
    GetSystemTimeAsFileTime(&ft);
    ULARGE_INTEGER u;
    u.LowPart = ft.dwLowDateTime;
    u.HighPart = ft.dwHighDateTime;
    return (double)(u.QuadPart - 116444736000000000ULL) / 10000000.0;
#elif defined(__EMSCRIPTEN__)
    #include <sys/time.h>
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return (double)tv.tv_sec + (double)tv.tv_usec / 1000000.0;
#elif defined(__wasm__)
    return 0.0;
#else
    #include <sys/time.h>
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return (double)tv.tv_sec + (double)tv.tv_usec / 1000000.0;
#endif
}

void nr_sleep_ms(int32_t ms) {
    nr_time_init();
    fiber_info_t* self = (fiber_info_t*)GetFiberData();
    if (!self) {
#ifdef _WIN32
        Sleep(ms);
#elif defined(__EMSCRIPTEN__)
        emscripten_sleep(ms);
#elif defined(__wasm__)
        // No-op
#else
        #include <unistd.h>
        usleep(ms * 1000);
#endif
        return;
    }

    double now = nr_time_now();
    double delay = (double)ms / 1000.0;

    timer_waiter_t* waiter = (timer_waiter_t*)malloc(sizeof(timer_waiter_t));
    waiter->wake_time = now + delay;
    waiter->fiber = self;
    waiter->next = NULL;

    NR_MUTEX_LOCK(&g_timer_waiters_lock);
    waiter->next = g_timer_waiters_head;
    g_timer_waiters_head = waiter;
    NR_MUTEX_UNLOCK(&g_timer_waiters_lock);

    NR_ATOMIC_ADD(&g_timer_waiters_count, 1);
    park();
}

char* nr_time_format(double sec, char* format) {
    time_t rawtime = (time_t)sec;
    struct tm* timeinfo = localtime(&rawtime);
    char buffer[256];
    if (timeinfo) {
        strftime(buffer, sizeof(buffer), format, timeinfo);
    } else {
        buffer[0] = '\0';
    }
    return nr_strdup(buffer); // Managed securely under Nora's memory tracker
}

typedef struct {
    int32_t day;
    int32_t hour;
    bool is_dst;
    int32_t minute;
    int32_t month;
    int32_t second;
    int32_t weekday;
    int32_t year;
    int32_t yearday;
} nr_date_parts_t;

void nr_time_to_parts(double sec, bool utc, void* parts_ptr) {
    time_t rawtime = (time_t)sec;
    struct tm* t = utc ? gmtime(&rawtime) : localtime(&rawtime);
    nr_date_parts_t* parts = (nr_date_parts_t*)parts_ptr;
    if (t && parts) {
        parts->year = t->tm_year + 1900;
        parts->month = t->tm_mon + 1;
        parts->day = t->tm_mday;
        parts->hour = t->tm_hour;
        parts->minute = t->tm_min;
        parts->second = t->tm_sec;
        parts->weekday = t->tm_wday;
        parts->yearday = t->tm_yday;
        parts->is_dst = t->tm_isdst > 0;
    }
}

double nr_i64_to_f64(long long x) { return (double)x; }
long long nr_f64_to_i64(double x) { return (long long)x; }
int nr_f64_to_i32(double x) { return (int)x; }
double nr_i32_to_f64(int x) { return (double)x; }
long long nr_i32_to_i64(int x) { return (long long)x; }
int nr_i64_to_i32(long long x) { return (int)x; }
