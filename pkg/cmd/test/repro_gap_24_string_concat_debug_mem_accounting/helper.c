#include "runtime.h"

extern NR_ATOMIC_LONG g_total_allocations;

long long get_total_allocations() {
    return (long long)NR_ATOMIC_LOAD(&g_total_allocations);
}
