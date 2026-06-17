#include <stddef.h>
#include <stdint.h>
#include <stdarg.h>

// Minimal WASM Runtime for Nora Plugins

// Simple bump allocator
static uint8_t heap[128 * 1024];
static size_t heap_ptr = 0;

void* malloc(size_t size) {
    // Align to 8 bytes
    size = (size + 7) & ~7;
    if (heap_ptr + size > sizeof(heap)) return NULL;
    void* ptr = &heap[heap_ptr];
    heap_ptr += size;
    return ptr;
}

void* calloc(size_t nmemb, size_t size) {
    size_t total = nmemb * size;
    void* ptr = malloc(total);
    if (!ptr) return NULL;
    uint8_t* p = (uint8_t*)ptr;
    for (size_t i = 0; i < total; i++) p[i] = 0;
    return ptr;
}

void free(void* ptr) {
    // No-op for bump allocator
}

void* memcpy(void* dest, const void* src, size_t n) {
    uint8_t* d = (uint8_t*)dest;
    const uint8_t* s = (const uint8_t*)src;
    while (n--) *d++ = *s++;
    return dest;
}

void* memset(void* s, int c, size_t n) {
    uint8_t* p = (uint8_t*)s;
    while (n--) *p++ = (uint8_t)c;
    return s;
}

size_t strlen(const char* s) {
    size_t len = 0;
    while (s[len]) len++;
    return len;
}

char* strstr(const char* haystack, const char* needle) {
    if (!*needle) return (char*)haystack;
    for (; *haystack; haystack++) {
        if (*haystack != *needle) continue;
        const char* h = haystack;
        const char* n = needle;
        while (*h && *n && *h == *n) {
            h++;
            n++;
        }
        if (!*n) return (char*)haystack;
    }
    return NULL;
}

int strcmp(const char* s1, const char* s2) {
    while (*s1 && (*s1 == *s2)) {
        s1++;
        s2++;
    }
    return *(unsigned char*)s1 - *(unsigned char*)s2;
}

char* strcpy(char* dest, const char* src) {
    char* d = dest;
    while ((*d++ = *src++));
    return dest;
}

// Minimal snprintf supporting only %s and %%
int snprintf(char* str, size_t size, const char* format, ...) {
    va_list args;
    va_start(args, format);
    int len = 0;
    const char* p = format;

    while (*p) {
        if (*p == '%') {
            p++;
            if (*p == 's') {
                const char* s = va_arg(args, const char*);
                if (!s) s = "(null)";
                while (*s) {
                    if (str && (size_t)len < size - 1) str[len] = *s;
                    len++;
                    s++;
                }
            } else if (*p == '%') {
                if (str && (size_t)len < size - 1) str[len] = '%';
                len++;
            }
            // Add more formatters if needed (%d, etc.)
        } else {
            if (str && (size_t)len < size - 1) str[len] = *p;
            len++;
        }
        p++;
    }

    if (str && size > 0) {
        if ((size_t)len < size) str[len] = '\0';
        else str[size - 1] = '\0';
    }

    va_end(args);
    return len;
}

// Dummy symbols required by generated code but not used in plugins
void nr_panic(const char* msg, const char* file, int line) {
    // In WASM we could trap or call an imported print
    __builtin_trap();
}
