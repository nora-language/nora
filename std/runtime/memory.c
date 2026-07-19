#include "runtime.h"


#ifdef _WIN32
#include <windows.h>
#elif defined(__EMSCRIPTEN__)
#include <emscripten.h>
#else
#include <stdatomic.h>
#endif

NR_ATOMIC_INT g_net_waiters_count = 0;
NR_ATOMIC_INT g_timer_waiters_count = 0;
NR_ATOMIC_LONG g_total_allocations = 0;
NR_ATOMIC_LONG g_num_allocations = 0;
nr_header_t* g_allocations_head = NULL;
NR_MUTEX_T g_mem_lock;
static NR_MUTEX_T g_untracked_lock;
static bool g_untracked_lock_init = false;

void nr_mem_init() {
#if NR_DEBUG_MEM
    NR_MUTEX_INIT(&g_mem_lock);
#endif
    NR_MUTEX_INIT(&g_untracked_lock);
    g_untracked_lock_init = true;
}

#define nr_malloc(s) nr_malloc_debug(s, __FILE__, __LINE__)

void* nr_malloc_debug(size_t size, const char* file, int line) {
    void* raw = malloc(size + NR_HEADER_SIZE);
    if (!raw) return NULL;
    memset(raw, 0, size + NR_HEADER_SIZE);
    nr_header_t* h = (nr_header_t*)raw;
    h->magic = NR_HEADER_MAGIC;
    h->count = 0;
    h->elem_size = (int)size;
    h->ref_count = 0;
    h->file = file;
    h->line = line;

#if NR_DEBUG_MEM
    NR_MUTEX_LOCK(&g_mem_lock);
    h->next = g_allocations_head;
    if (g_allocations_head) g_allocations_head->prev = h;
    g_allocations_head = h;
    NR_MUTEX_UNLOCK(&g_mem_lock);
#endif

    NR_ATOMIC_ADD(&g_total_allocations, (long long)size);
    NR_ATOMIC_ADD(&g_num_allocations, 1);
    return (char*)raw + NR_HEADER_SIZE;
}

#define NR_MALLOC(s) nr_malloc_debug(s, __FILE__, __LINE__)

// nr_malloc is now a macro

void nr_free(void* ptr) {
    if (!ptr) return;
    nr_header_t* h = (nr_header_t*)((char*)ptr - NR_HEADER_SIZE);
    if (h->magic == NR_MAGIC_STATIC) return; 
    if (h->magic != NR_HEADER_MAGIC) return;

    if (NR_ATOMIC_SUB(&h->ref_count, 1) <= 0) {
        int old_magic = NR_ATOMIC_EXCHANGE(&h->magic, NR_MAGIC_FREE);
        if (old_magic == NR_HEADER_MAGIC) {
            
#if NR_DEBUG_MEM
            NR_MUTEX_LOCK(&g_mem_lock);
            if (h->prev) h->prev->next = h->next;
            if (h->next) h->next->prev = h->prev;
            if (h == g_allocations_head) g_allocations_head = h->next;
            NR_MUTEX_UNLOCK(&g_mem_lock);
#endif

            long long sz = h->elem_size;
            if (h->count > 0) sz *= h->count;
            NR_ATOMIC_SUB(&g_total_allocations, sz);
            NR_ATOMIC_SUB(&g_num_allocations, 1);
            free(h);
        }
    }
}

typedef struct nr_untracked_chunk {
    struct nr_untracked_chunk* next;
} nr_untracked_chunk_t;

static nr_untracked_chunk_t* g_untracked_freelist = NULL;

#define NR_UNTRACKED_MAGIC 0x554E5452
#define NR_UNTRACKED_LARGE_MAGIC 0x4C415247

typedef struct {
    uint32_t magic;
    uint32_t pad;
} nr_untracked_header_t;

void* nr_malloc_untracked(int size) {
    if (size <= 256) {
        NR_MUTEX_LOCK(&g_untracked_lock);
        nr_untracked_chunk_t* c = g_untracked_freelist;
        if (c) {
            g_untracked_freelist = c->next;
            NR_MUTEX_UNLOCK(&g_untracked_lock);
            nr_untracked_header_t* h = (nr_untracked_header_t*)c;
            h->magic = NR_UNTRACKED_MAGIC;
            return (void*)(h + 1);
        }
        NR_MUTEX_UNLOCK(&g_untracked_lock);
        nr_untracked_header_t* h = (nr_untracked_header_t*)malloc(sizeof(nr_untracked_header_t) + 256);
        if (!h) return NULL;
        h->magic = NR_UNTRACKED_MAGIC;
        return (void*)(h + 1);
    }
    nr_untracked_header_t* h = (nr_untracked_header_t*)malloc(sizeof(nr_untracked_header_t) + size);
    if (!h) return NULL;
    h->magic = NR_UNTRACKED_LARGE_MAGIC;
    return (void*)(h + 1);
}

void nr_free_untracked(void* p) {
    if (!p) return;
    nr_untracked_header_t* h = ((nr_untracked_header_t*)p) - 1;
    if (h->magic == NR_UNTRACKED_MAGIC) {
        NR_MUTEX_LOCK(&g_untracked_lock);
        nr_untracked_chunk_t* c = (nr_untracked_chunk_t*)h;
        c->next = g_untracked_freelist;
        g_untracked_freelist = c;
        NR_MUTEX_UNLOCK(&g_untracked_lock);
        return;
    }
    if (h->magic == NR_UNTRACKED_LARGE_MAGIC) {
        free(h);
        return;
    }
    free(p);
}

void nr_untracked_cleanup() {
    if (!g_untracked_lock_init) return;
    NR_MUTEX_LOCK(&g_untracked_lock);
    nr_untracked_chunk_t* curr = g_untracked_freelist;
    while (curr) {
        nr_untracked_chunk_t* next = curr->next;
        free(curr);
        curr = next;
    }
    g_untracked_freelist = NULL;
    NR_MUTEX_UNLOCK(&g_untracked_lock);
}

void nr_mem_report() {
#if NR_DEBUG_MEM
    long long num = (long long)NR_ATOMIC_LOAD(&g_num_allocations);
    long long total = (long long)NR_ATOMIC_LOAD(&g_total_allocations);
    if (num > 0) {
        printf("\nNora MEMORY LEAK REPORT\n");
        printf("==========================\n");
        NR_MUTEX_LOCK(&g_mem_lock);
        nr_header_t* curr = g_allocations_head;
        int i = 0;
        while (curr && i < 100) { // Limit to 100 to avoid flooding
            long long sz = curr->elem_size;
            if (curr->count > 0) sz *= curr->count;
            char* data = (char*)curr + NR_HEADER_SIZE;
            if (curr->elem_size == 1 && curr->count > 0) {
                printf("Leak: %lld bytes at %s:%d, value: \"", sz, curr->file ? curr->file : "unknown", curr->line);
                for (int j = 0; j < curr->count && j < 100; j++) {
                    char c = data[j];
                    if (c >= 32 && c <= 126) {
                        putchar(c);
                    } else {
                        printf("\\x%02x", (unsigned char)c);
                    }
                }
                printf("\"\n");
            } else {
                printf("Leak: %lld bytes at %s:%d\n", sz, curr->file ? curr->file : "unknown", curr->line);
            }
            curr = curr->next;
            i++;
        }
        if (curr) printf("... and more\n");
        NR_MUTEX_UNLOCK(&g_mem_lock);
        printf("--------------------------\n");
        printf("Active allocations: %lld\n", num);
        printf("Total leaked bytes: %lld\n", total);
        printf("==========================\n");
    }
#endif
}
void nr_free_str(void* p) { nr_free(p); }
