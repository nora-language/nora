#include "runtime.h"

// --- STRING RUNTIME ---
void nr_free_str(void* ptr); // Prototype
char* nr_str_concat(char* s1, char* s2) {
    if (!s1) s1 = "";
    if (!s2) s2 = "";
    size_t l1 = strlen(s1);
    size_t l2 = strlen(s2);
    void* p = nr_malloc(l1 + l2 + 1);
    char* res = (char*)p;
    nr_header_t* h = (nr_header_t*)((char*)p - NR_HEADER_SIZE);
    h->count = (int)(l1 + l2);
    h->elem_size = 1;
    NR_ATOMIC_STORE(&h->ref_count, 0);
    h->magic = NR_HEADER_MAGIC;
    memcpy(res, s1, l1);
    memcpy(res + l1, s2, l2);
    res[l1 + l2] = 0;
    return res;
}
char* nr_str_concat_free(char* s1, char* s2, bool f1, bool f2) {
    char* res = nr_str_concat(s1, s2);
    if (f1 && s1 && *s1) nr_free(s1);
    if (f2 && s2 && *s2) nr_free(s2);
    return res;
}

char* nr_str_from_cstring(void* s) {
    const char* cs = (const char*)s;
    size_t len = cs ? strlen(cs) : 0;
    void* p = nr_malloc(len + 1);
    char* res = (char*)p;
    nr_header_t* h = (nr_header_t*)((char*)p - NR_HEADER_SIZE);
    h->count = (int)len;
    h->elem_size = 1;
    NR_ATOMIC_STORE(&h->ref_count, 0);
    h->magic = NR_HEADER_MAGIC;
    if (len > 0) {
        memcpy(res, cs, len);
    }
    res[len] = 0;
    return res;
}


bool nr_str_eq(char* s1, char* s2) {
    if (s1 == s2) return true;
    if (!s1 || !s2) return false;
    return strcmp(s1, s2) == 0;
}

// Temporary string management for expressions
char* nr_temp_str(char* s) {
    fiber_info_t* info = (fiber_info_t*)GetFiberData();
    if (info && s && info->temp_count < 256) {
        info->temp_strs[info->temp_count++] = s;
    }
    return s;
}
char* nr_claim_str(char* s) {
    if (s && *s) {
        nr_header_t* h = (nr_header_t*)((char*)s - NR_HEADER_SIZE);
        if (h->magic == NR_HEADER_MAGIC) {
            NR_ATOMIC_INC(&h->ref_count);
        }
    }
    return s;
}
void nr_flush_temps() {
    fiber_info_t* info = (fiber_info_t*)GetFiberData();
    if (info) {
        while (info->temp_count > 0) {
            nr_free(info->temp_strs[--info->temp_count]);
        }
    }
}
char* nr_strdup(const char* s) {
    if (!s) return NULL;
    size_t len = strlen(s);
    void* p = nr_malloc(len + 1);
    char* res = (char*)p;
    nr_header_t* h = (nr_header_t*)((char*)p - NR_HEADER_SIZE);
    h->count = 0;
    h->elem_size = (int)len;
    NR_ATOMIC_STORE(&h->ref_count, 0);
    h->magic = NR_HEADER_MAGIC;
    memcpy(res, s, len + 1);
    return res;
}
char* nr_i32_to_str(int v) {
    char buf[32];
    sprintf(buf, "%d", v);
    return nr_strdup(buf);
}
char* nr_i64_to_str(long long v) {
    char buf[64];
    sprintf(buf, "%lld", v);
    return nr_strdup(buf);
}
char* nr_f64_to_str(double v) {
    char buf[64];
    sprintf(buf, "%g", v);
    return nr_strdup(buf);
}
char* nr_bool_to_str(bool v) {
    return nr_strdup(v ? "true" : "false");
}
char* nr_to_str(void* p) {
    char buf[32];
    sprintf(buf, "%p", p);
    return nr_strdup(buf);
}
