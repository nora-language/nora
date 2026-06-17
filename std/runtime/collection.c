#include "runtime.h"

// --- COLLECTION RUNTIME ---
typedef struct {
    void* key;
    void* val;
    bool occupied;
} MapEntry;

typedef struct {
    int capacity;
    int size;
    MapEntry* entries;
    int key_size;
    int val_size;
    bool is_str_key;
} HashMap;

unsigned int map_hash(HashMap* m, void* key) {
    unsigned int hash = 2166136261u;
    unsigned char* p;
    int len;
    if (m->is_str_key) {
        p = (unsigned char*)(*(char**)key);
        len = strlen((char*)p);
    } else {
        p = (unsigned char*)key;
        len = m->key_size;
    }
    for (int i = 0; i < len; i++) {
        hash = (hash ^ p[i]) * 16777619;
    }
    return hash;
}

bool map_key_eq(HashMap* m, void* k1, void* k2) {
    if (m->is_str_key) {
        return strcmp(*(char**)k1, *(char**)k2) == 0;
    }
    return memcmp(k1, k2, m->key_size) == 0;
}

void* map_make(int key_size, int val_size, bool is_str_key, const char* file, int line) {
    HashMap* m = (HashMap*)nr_malloc_debug(sizeof(HashMap), file, line);
    m->capacity = 16;
    m->size = 0;
    m->key_size = key_size;
    m->val_size = val_size;
    m->is_str_key = is_str_key;
    m->entries = (MapEntry*)nr_malloc_debug(m->capacity * sizeof(MapEntry), file, line);
    memset(m->entries, 0, m->capacity * sizeof(MapEntry));
    return m;
}

void map_set(void* _m, void* key, void* val) {
    HashMap* m = (HashMap*)_m;
    if (m->size >= m->capacity * 0.7) {
        int old_cap = m->capacity;
        MapEntry* old_entries = m->entries;
        m->capacity *= 2;
        m->entries = (MapEntry*)nr_malloc(m->capacity * sizeof(MapEntry));
        memset(m->entries, 0, m->capacity * sizeof(MapEntry));
        m->size = 0;
        for (int i = 0; i < old_cap; i++) {
            if (old_entries[i].occupied) {
                map_set(m, old_entries[i].key, old_entries[i].val);
                if (m->is_str_key) nr_free(*(char**)old_entries[i].key);
                nr_free(old_entries[i].key);
                nr_free(old_entries[i].val);
            }
        }
        nr_free(old_entries);
    }

    int h = map_hash(m, key) % m->capacity;
    while (m->entries[h].occupied) {
        if (map_key_eq(m, m->entries[h].key, key)) {
            memcpy(m->entries[h].val, val, m->val_size);
            return;
        }
        h = (h + 1) % m->capacity;
    }

    m->entries[h].key = nr_malloc(m->key_size);
    memcpy(m->entries[h].key, key, m->key_size);
    if (m->is_str_key) {
        char* original = *(char**)key;
        char* copy = (char*)nr_malloc(strlen(original) + 1);
        strcpy(copy, original);
        *(char**)m->entries[h].key = copy;
    }
    m->entries[h].val = nr_malloc(m->val_size);
    memcpy(m->entries[h].val, val, m->val_size);
    m->entries[h].occupied = true;
    m->size++;
}

void* map_get(void* _m, void* key) {
    HashMap* m = (HashMap*)_m;
    if (!m) return NULL;
    int h = map_hash(m, key) % m->capacity;
    int start = h;
    while (m->entries[h].occupied) {
        if (map_key_eq(m, m->entries[h].key, key)) {
            return m->entries[h].val;
        }
        h = (h + 1) % m->capacity;
        if (h == start) break;
    }
    return NULL;
}

bool map_contains(void* _m, void* key) {
    return map_get(_m, key) != NULL;
}

void map_free(void* _m) {
    HashMap* m = (HashMap*)_m;
    if (!m) return;
    for (int i = 0; i < m->capacity; i++) {
        if (m->entries[i].occupied) {
            if (m->is_str_key) nr_free(*(char**)m->entries[i].key);
            nr_free(m->entries[i].key);
            nr_free(m->entries[i].val);
        }
    }
    nr_free(m->entries);
    nr_free(m);
}

void* array_make(int count, int elem_size, const char* file, int line, ...) {
    void* p = nr_malloc_debug(count * elem_size, file, line);
    nr_header_t* h = (nr_header_t*)((char*)p - NR_HEADER_SIZE);
    h->count = count;
    h->elem_size = elem_size;
    h->magic = NR_HEADER_MAGIC;
    
    va_list args;
    va_start(args, line);
    for (int i = 0; i < count; i++) {
        // Since we don't know the exact type, we use a heuristic based on elem_size
        // for standard Nora types.
        if (elem_size == sizeof(int)) {
            int val = va_arg(args, int);
            memcpy((char*)p + (i * elem_size), &val, elem_size);
        } else if (elem_size == sizeof(double)) {
            double val = va_arg(args, double);
            memcpy((char*)p + (i * elem_size), &val, elem_size);
        } else if (elem_size == sizeof(void*)) {
            void* val = va_arg(args, void*);
            memcpy((char*)p + (i * elem_size), &val, elem_size);
        } else {
            // Fallback for other sizes (like small structs) — might be tricky with va_arg
            // but Nora mostly uses ptr/int/float for literals.
            void* val = va_arg(args, void*);
            memcpy((char*)p + (i * elem_size), &val, elem_size);
        }
    }
    va_end(args);
    return p;
}

void* array_make_empty(int count, int elem_size, const char* file, int line) {
    void* p = nr_malloc_debug(count * elem_size, file, line);
    nr_header_t* h = (nr_header_t*)((char*)p - NR_HEADER_SIZE);
    h->count = count;
    h->elem_size = elem_size;
    h->magic = NR_HEADER_MAGIC;
    memset(p, 0, count * elem_size);
    return p;
}

void* array_append(void* arr, void* data) {
    if (!arr) return NULL;
    nr_header_t* h = (nr_header_t*)((char*)arr - NR_HEADER_SIZE);
    int count = h->count;
    int elem_size = h->elem_size;
    
    void* new_arr = nr_malloc((count + 1) * elem_size);
    nr_header_t* new_h = (nr_header_t*)((char*)new_arr - NR_HEADER_SIZE);
    new_h->count = count + 1;
    new_h->elem_size = elem_size;
    new_h->magic = NR_HEADER_MAGIC;
    
    if (count > 0) {
        memcpy(new_arr, arr, count * elem_size);
    }
    memcpy((char*)new_arr + (count * elem_size), data, elem_size);
    
    // We free the old array here because Nora's 'append' builtin moves the list.
    // The RAII solver marks the input list as consumed, so it won't be freed elsewhere.
    nr_free(arr);
    
    return new_arr;
}

void* array_data(void* data) { return data; }

int array_count(void* data) {
    if (!data) return 0;
    nr_header_t* h = (nr_header_t*)((char*)data - NR_HEADER_SIZE);
    if (h->magic != NR_HEADER_MAGIC && h->magic != NR_MAGIC_STATIC) return 0;
    return h->count;
}

void* array_slice(void* arr, int start, int end, int elem_size) {
    int count = array_count(arr);
    if (start < 0) start = 0;
    if (end < 0 || end > count) end = count;
    if (start >= end) return nr_malloc(0);

    int new_count = end - start;
    void* new_arr = nr_malloc(new_count * elem_size);
    nr_header_t* h = (nr_header_t*)((char*)new_arr - NR_HEADER_SIZE);
    h->count = new_count;
    h->elem_size = elem_size;
    h->magic = NR_HEADER_MAGIC;
    memcpy(new_arr, (char*)arr + (start * elem_size), new_count * elem_size);
    return new_arr;
}

char* string_slice(char* s, int start, int end) {
    if (!s) return nr_strdup("");
    int len = (int)strlen(s);
    if (start < 0) start = 0;
    if (end < 0 || end > len) end = len;
    if (start >= end) return nr_strdup("");

    int new_len = end - start;
    void* p = nr_malloc(new_len + 1);
    char* res = (char*)p;
    nr_header_t* h = (nr_header_t*)((char*)p - NR_HEADER_SIZE);
    h->count = 0;
    h->elem_size = new_len;
    h->magic = NR_HEADER_MAGIC;
    memcpy(res, s + start, new_len);
    res[new_len] = 0;
    return res;
}

void* array_bounds_check(void* arr, int index, const char* file, int line) {
    if (!arr) {
        nr_panic("null array access", file, line);
    }
    nr_header_t* h = (nr_header_t*)((char*)arr - NR_HEADER_SIZE);
    if (h->magic != NR_HEADER_MAGIC && h->magic != NR_MAGIC_STATIC) {
        nr_panic("invalid array header", file, line);
    }
    int count = h->count;
    int elem_size = h->elem_size;
    if (index < 0 || index >= count) {
        // Need to format the string for panic message, but since nr_panic takes a const char*,
        // we can use a thread-local buffer or static buffer. 
        // For now, let's just use a static thread-local buffer for the message.
#ifdef _WIN32
        static __declspec(thread) char msg_buf[128];
#else
        static __thread char msg_buf[128];
#endif
        snprintf(msg_buf, sizeof(msg_buf), "array index out of bounds: %d (length %d)", index, count);
        nr_panic(msg_buf, file, line);
    }
    return (char*)arr + (index * elem_size);
}
void nr_print_backtrace() {
#ifdef _WIN32
    void* stack[100];
    unsigned short frames;
    SYMBOL_INFO* symbol;
    HANDLE process;

    process = GetCurrentProcess();
    SymSetOptions(SYMOPT_LOAD_LINES | SYMOPT_UNDNAME);
    SymInitialize(process, NULL, TRUE);

    frames = CaptureStackBackTrace(0, 100, stack, NULL);
    symbol = (SYMBOL_INFO*)calloc(sizeof(SYMBOL_INFO) + 256 * sizeof(char), 1);
    symbol->MaxNameLen = 255;
    symbol->SizeOfStruct = sizeof(SYMBOL_INFO);

    fprintf(stderr, "\nStack Backtrace:\n");
    for (int i = 0; i < frames; i++) {
        DWORD64 address = (DWORD64)(stack[i]);
        if (SymFromAddr(process, address, 0, symbol)) {
            IMAGEHLP_LINE64 line_info;
            line_info.SizeOfStruct = sizeof(IMAGEHLP_LINE64);
            DWORD displacement;
            if (SymGetLineFromAddr64(process, address, &displacement, &line_info)) {
                fprintf(stderr, "  %d: %s (%s:%d)\n", i, symbol->Name, line_info.FileName, (int)line_info.LineNumber);
            } else {
                fprintf(stderr, "  %d: %s - 0x%llX\n", i, symbol->Name, (unsigned long long)symbol->Address);
            }
        } else {
            fprintf(stderr, "  %d: ??? - 0x%p\n", i, stack[i]);
        }
    }

    free(symbol);
#elif defined(__EMSCRIPTEN__) || defined(__wasm__)
    fprintf(stderr, "\nBacktrace not supported on this platform.\n");
#else
    void* array[100];
    size_t size;
    char** strings;

    size = backtrace(array, 100);
    strings = backtrace_symbols(array, size);

    fprintf(stderr, "\nStack Backtrace:\n");
    for (size_t i = 0; i < size; i++) {
        Dl_info info;
        if (dladdr(array[i], &info) && info.dli_sname) {
            fprintf(stderr, "  %zu: %s + %p (%s)\n", i, info.dli_sname, (void*)((char*)array[i] - (char*)info.dli_saddr), info.dli_fname);
        } else {
            fprintf(stderr, "  %zu: %s\n", i, strings[i]);
        }
    }

    free(strings);
#endif
}

_Noreturn void nr_panic(const char* msg, const char* file, int line) {
    void* p = nr_fiber_current();
    if (p) {
        fiber_info_t* self = (fiber_info_t*)p;
        if (!self->is_main) {
            self->panic_msg = msg;
            self->panic_file = file;
            self->panic_line = line;
            longjmp(self->panic_buf, 1);
        }
    }
    fprintf(stderr, "Panic: %s at %s:%d\n", msg, file, line);
    nr_print_backtrace();
    exit(1);
}
void nr_fiber_report() {
    printf("\nNora FIBER INSPECTOR\n");
    printf("================================================================================\n");
    printf("%-5s | %-10s | %-30s | %-20s\n", "ID", "State", "Location", "Function");
    printf("--------------------------------------------------------------------------------\n");

    NR_MUTEX_LOCK(&g_fiber_list_lock);
    fiber_info_t* curr = g_fibers_head;
    int count = 0;
    while (curr) {
        const char* state_str = "UNKNOWN";
        long s = NR_ATOMIC_LOAD(&curr->state);
        switch (s) {
            case 0: state_str = "READY"; break;
            case 1: state_str = "RUNNING"; break;
            case 2: state_str = "PARKING"; break;
            case 3: state_str = "PARKED"; break;
            case 4: state_str = "TERMINATED"; break;
        }

        char loc[64];
        snprintf(loc, sizeof(loc), "%s:%d", curr->file ? curr->file : "?", curr->line);
        
        printf("%-5d | %-10s | %-30s | %-20s\n", count++, state_str, loc, curr->name ? curr->name : "?");
        curr = curr->next_global;
    }
    NR_MUTEX_UNLOCK(&g_fiber_list_lock);
    printf("================================================================================\n");
}

void lose_it(void* p) {}
void nr_free_vector(void* p) { nr_free(p); }
