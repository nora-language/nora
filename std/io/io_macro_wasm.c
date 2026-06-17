#include <stddef.h>
#include <stdint.h>
#include <stdarg.h>
#include <stdbool.h>

static uint8_t heap[512 * 1024];
static size_t heap_ptr = 0;

void* malloc(size_t size) {
    size = (size + 7) & ~7;
    if (heap_ptr + size > sizeof(heap)) return NULL;
    void* ptr = &heap[heap_ptr];
    heap_ptr += size;
    return ptr;
}
void free(void* p) {}
void* memcpy(void* d, const void* s, size_t n) {
    char* dst = (char*)d; const char* src = (const char*)s;
    while (n--) *dst++ = *src++;
    return d;
}
void* memset(void* s, int c, size_t n) {
    char* p = (char*)s; while (n--) *p++ = (char)c;
    return s;
}
size_t strlen(const char* s) {
    size_t l = 0; while (s && s[l]) l++; return l;
}
char* strstr(const char* h, const char* n) {
    if (!h || !n) return NULL;
    if (!*n) return (char*)h;
    for (; *h; h++) {
        const char* hh = h; const char* nn = n;
        while (*hh && *nn && *hh == *nn) { hh++; nn++; }
        if (!*nn) return (char*)h;
    }
    return NULL;
}
int strcmp(const char* a, const char* b) {
    while (*a && (*a == *b)) { a++; b++; }
    return *(unsigned char*)a - *(unsigned char*)b;
}
int strncmp(const char* a, const char* b, size_t n) {
    while (n-- && *a && (*a == *b)) { a++; b++; }
    return n == (size_t)-1 ? 0 : *(unsigned char*)a - *(unsigned char*)b;
}

int snprintf(char* str, size_t size, const char* format, ...) {
    va_list args; va_start(args, format);
    int len = 0; const char* p = format;
    while (*p) {
        if (*p == '%') {
            p++;
            if (*p == 's') {
                const char* s = va_arg(args, const char*);
                if (!s) s = "(null)";
                while (*s) {
                    if (str && (size_t)len + 1 < size) str[len] = *s;
                    len++; s++;
                }
            } else if (*p == '%') {
                if (str && (size_t)len + 1 < size) str[len] = '%';
                len++;
            }
        } else {
            if (str && (size_t)len + 1 < size) str[len] = *p;
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
void nr_panic(const char* msg, const char* file, int line) { __builtin_trap(); }
void* plugin_alloc(size_t size) { return malloc(size); }
void plugin_reset() { heap_ptr = 0; }

// Append a string to a buffer, returning the new length
int str_append(char* buf, int pos, int max, const char* s) {
    while (*s && pos < max - 1) buf[pos++] = *s++;
    buf[pos] = '\0';
    return pos;
}

// Escape a string for JSON (only " and \)
void json_escape(char* dest, int* pos, int max, const char* src) {
    while (*src && *pos < max - 1) {
        if (*src == '"') {
            if (*pos < max - 2) {
                dest[(*pos)++] = '\\';
                dest[(*pos)++] = '"';
            }
        } else if (*src == '\\') {
            if (*pos < max - 2) {
                dest[(*pos)++] = '\\';
                dest[(*pos)++] = '\\';
            }
        } else if (*src == '\n') {
            if (*pos < max - 2) {
                dest[(*pos)++] = '\\';
                dest[(*pos)++] = 'n';
            }
        } else {
            dest[(*pos)++] = *src;
        }
        src++;
    }
    dest[*pos] = '\0';
}

// Get value of a JSON string field (returns content without surrounding quotes)
// The returned pointer is into a fresh malloc'd buffer with the unescaped content
char* get_json_str_value(char* payload, char* key) {
    char key_buf[256];
    snprintf(key_buf, sizeof(key_buf), "\"%s\":", key);
    char* found = strstr(payload, key_buf);
    if (!found) return NULL;
    char* p = found + strlen(key_buf);
    while (*p == ' ' || *p == '\t' || *p == '\n' || *p == '\r') p++;
    if (*p != '"') return NULL;
    p++; // skip opening quote

    // First pass: calculate unescaped length
    char* p_len = p;
    int len = 0;
    while (*p_len && *p_len != '"') {
        if (*p_len == '\\') {
            p_len++;
            if (!*p_len) break;
        }
        len++;
        p_len++;
    }

    char* buf = malloc(len + 1);
    if (!buf) return NULL;

    int pos = 0;
    while (*p && *p != '"') {
        if (*p == '\\') {
            p++;
            if (*p == '"') { buf[pos++] = '"'; }
            else if (*p == '\\') { buf[pos++] = '\\'; }
            else if (*p == 'n') { buf[pos++] = '\n'; }
            else if (*p == 't') { buf[pos++] = '\t'; }
            else if (*p == 'r') { buf[pos++] = '\r'; }
            else { buf[pos++] = '\\'; buf[pos++] = *p; }
        } else {
            buf[pos++] = *p;
        }
        p++;
    }
    buf[pos] = '\0';
    return buf;
}

// Find the Nth element of a JSON array field, returned as a raw substring (with { })
char* get_json_array_element(char* payload, char* key, int index) {
    char key_buf[256];
    snprintf(key_buf, sizeof(key_buf), "\"%s\":", key);
    char* found = strstr(payload, key_buf);
    if (!found) return NULL;
    char* p = found + strlen(key_buf);
    while (*p && *p != '[') p++;
    if (!*p) return NULL;
    p++; // skip '['

    // Skip to the Nth element
    for (int i = 0; i < index; i++) {
        int depth = 0; bool in_str = false; char prev = 0;
        while (*p) {
            if (in_str) {
                if (*p == '"' && prev != '\\') in_str = false;
            } else {
                if (*p == '"') { in_str = true; }
                else if (*p == '{' || *p == '[') depth++;
                else if (*p == '}' || *p == ']') { depth--; if (depth < 0) return NULL; }
                else if (*p == ',' && depth == 0) { p++; break; }
            }
            prev = *p; p++;
        }
    }

    // Skip whitespace
    while (*p == ' ' || *p == '\t' || *p == '\n' || *p == '\r') p++;
    if (*p == ']' || *p == '\0') return NULL;

    // Now p points at start of the element
    char* start = p;
    int depth = 0; bool in_str = false; char prev = 0;
    while (*p) {
        if (in_str) {
            if (*p == '"' && prev != '\\') in_str = false;
        } else {
            if (*p == '"') { in_str = true; }
            else if (*p == '{' || *p == '[') depth++;
            else if (*p == '}' || *p == ']') {
                if (depth == 0) break;
                depth--;
            }
            else if (*p == ',' && depth == 0) break;
        }
        prev = *p; p++;
    }

    int len = p - start;
    if (len <= 0) return NULL;
    char* res = malloc(len + 1);
    memcpy(res, start, len);
    res[len] = '\0';
    return res;
}

// Get format specifier given type_name and value_raw (the generated C expression)
// Returns:
//   "STR_LITERAL" -> value_raw starts with '"', strip quotes and embed inline
//   "STR_VAR"     -> str variable/interpolated expr, use %s and pass as arg
//   "%d" etc.     -> numeric types
const char* get_fmt_spec(const char* type_name, const char* value_raw) {
    if (!type_name) return "%s";
    
    // Skip Nora pointer/lease prefixes
    const char* t = type_name;
    while (*t == '#' || *t == '&' || *t == '@' || *t == '*') t++;

    if (strcmp(t, "str") == 0) {
        return "%s";
    }
    if (strcmp(t, "i8") == 0 || strcmp(t, "i16") == 0 || strcmp(t, "i32") == 0 || strcmp(t, "int") == 0) return "%d";
    if (strcmp(t, "ptr") == 0) return "%p";
    if (strcmp(t, "u8") == 0 || strcmp(t, "u16") == 0 || strcmp(t, "u32") == 0) return "%u";
    if (strcmp(t, "i64") == 0) return "%lld";
    if (strcmp(t, "u64") == 0) return "%llu";
    if (strcmp(t, "f32") == 0) return "%f";
    if (strcmp(t, "f64") == 0) return "%lf";
    if (strcmp(t, "bool") == 0) return "%s";
    // Default: string
    return "%s";
}

// Build printf/fprintf call.
// newline=true: append \n to the format string
// use_stderr=true: use fprintf(stderr,...) instead of printf(...)
char* build_printf_impl(char* request, bool newline, bool use_stderr) {
    char c_code[16384]; // Much larger buffer for generated code
    int c_pos = 0;
    int i = 0;

    c_code[0] = '\0';

    // Count arguments first
    int arg_count = 0;
    while (1) {
        char* arg_obj = get_json_array_element(request, "arguments", arg_count);
        if (!arg_obj) break;
        arg_count++;
    }

    if (arg_count == 0) {
        if (newline) {
            if (use_stderr) c_pos = snprintf(c_code, sizeof(c_code), "fprintf(stderr, \"\\n\");");
            else c_pos = snprintf(c_code, sizeof(c_code), "printf(\"\\n\");");
        } else {
            c_pos = snprintf(c_code, sizeof(c_code), "(void)0;");
        }
    } else {
        // Wrap in a block to handle multiple printf statements
        c_pos = str_append(c_code, c_pos, sizeof(c_code), "{ ");

        for (i = 0; i < arg_count; i++) {
            char* arg_obj = get_json_array_element(request, "arguments", i);
            if (!arg_obj) break;

            char* val = get_json_str_value(arg_obj, "value_raw");
            char* type_name = get_json_str_value(arg_obj, "type_name");

            if (!val) val = "";
            if (!type_name) type_name = "";

            const char* spec = get_fmt_spec(type_name, val);
            const char* sep = "";
            if (newline && i == arg_count - 1) sep = "\\n";
            else if (i < arg_count - 1) sep = " ";

            char call[2048];
            const char* target = use_stderr ? "fprintf(stderr, " : "printf(";
            
            if (val[0] == '"') {
                // Strip quotes for direct embedding in the printf argument list
                // (but still use %s to be safe against % in literal)
                char* s = val;
                s++; // skip leading quote
                int vlen = strlen(s);
                if (vlen > 0 && s[vlen-1] == '"') s[vlen-1] = '\0';

                snprintf(call, sizeof(call), "%s\"%%s%%s\", \"%s\", \"%s\"); ", target, s, sep);
            } else if (strcmp(type_name, "bool") == 0) {
                snprintf(call, sizeof(call), "%s\"%%s%%s\", (%s) ? \"true\" : \"false\", \"%s\"); ", target, val, sep);
            } else {
                // Use <spec>%s pattern
                snprintf(call, sizeof(call), "%s\"%s%%s\", %s, \"%s\"); ", target, spec, val, sep);
            }
            c_pos = str_append(c_code, c_pos, sizeof(c_code), call);
        }
        c_pos = str_append(c_code, c_pos, sizeof(c_code), "}");
    }

    // Now escape the C code for JSON
    // We need a larger result buffer for many arguments
    char* result = malloc(32768); 
    if (!result) return NULL;
    int res_pos = 0;
    res_pos = str_append(result, res_pos, 32768, "{\"replacement_code\": \"");
    json_escape(result, &res_pos, 32768, c_code);
    res_pos = str_append(result, res_pos, 32768, "\"}");
    
    return result;
}

char* expand_println(char* request) {
    return build_printf_impl(request, true, false);
}

char* expand_print(char* request) {
    return build_printf_impl(request, false, false);
}

char* expand_eprintln(char* request) {
    return build_printf_impl(request, true, true);
}

char* expand_eprint(char* request) {
    return build_printf_impl(request, false, true);
}

// ─── SCAN / SCANLN ────────────────────────────────────────────────────────────
// Get scanf format specifier for a given type_name.
// Supports primitives (i32, f64, etc.) and pointers (*i32, *f64, etc.)
const char* get_scan_fmt_spec(const char* type_name) {
    if (!type_name) return "%s";
    
    // Handle pointer types (e.g. #i32, *i32, &i32, @i32)
    const char* t = type_name;
    while (*t == '#' || *t == '&' || *t == '@' || *t == '*') t++;

    if (strcmp(t, "str") == 0) return "%s";
    if (strcmp(t, "i8") == 0 || strcmp(t, "i16") == 0 || strcmp(t, "i32") == 0 || strcmp(t, "int") == 0) return "%d";
    if (strcmp(t, "u8") == 0 || strcmp(t, "u16") == 0 || strcmp(t, "u32") == 0) return "%u";
    if (strcmp(t, "i64") == 0) return "%lld";
    if (strcmp(t, "u64") == 0) return "%llu";
    if (strcmp(t, "f32") == 0) return "%f";
    if (strcmp(t, "f64") == 0) return "%lf";
    if (strcmp(t, "bool") == 0) return "%d";
    if (strcmp(t, "ptr") == 0) return "%p";
    
    return "%s";
}

bool is_passed_by_value(const char* type_name) {
    if (type_name[0] == '#') {
        const char* t = type_name + 1;
        if (strcmp(t, "i8") == 0 || strcmp(t, "i16") == 0 || strcmp(t, "i32") == 0 || strcmp(t, "int") == 0 ||
            strcmp(t, "u8") == 0 || strcmp(t, "u16") == 0 || strcmp(t, "u32") == 0 ||
            strcmp(t, "i64") == 0 || strcmp(t, "u64") == 0 ||
            strcmp(t, "f32") == 0 || strcmp(t, "f64") == 0 ||
            strcmp(t, "bool") == 0) {
            return true;
        }
    }
    return false;
}


// Build scanf call for read macros.
// scanln = true: consume trailing newline after reading (adds " " to consume whitespace)
char* build_scanf(char* request, bool scanln) {
    char fmt_str[512];
    char args_str[2048];
    int fmt_pos = 0;
    int args_pos = 0;
    int i = 0;

    fmt_str[0] = '\0';
    args_str[0] = '\0';

    while (1) {
        char* arg_obj = get_json_array_element(request, "arguments", i);
        if (!arg_obj) break;

        char* val = get_json_str_value(arg_obj, "value_raw");
        char* type_name = get_json_str_value(arg_obj, "type_name");

        if (!val) val = "";
        if (!type_name) type_name = "";

        const char* spec = get_scan_fmt_spec(type_name);
        fmt_pos = str_append(fmt_str, fmt_pos, sizeof(fmt_str), spec);

        if (args_pos > 0) {
            args_pos = str_append(args_str, args_pos, sizeof(args_str), ", ");
        }

        // str is already a char* pointer — pass directly
        // ptr types are already pointers — pass directly
        // All other types need & (address-of) to receive the value
        bool is_str = (strcmp(type_name, "str") == 0);
        bool is_ptr = (type_name[0] == '*' || type_name[0] == '#' || type_name[0] == '&' || type_name[0] == '@' || strcmp(type_name, "ptr") == 0);
        if (is_passed_by_value(type_name)) {
            is_ptr = false;
        }
        
        if (!is_str && !is_ptr) {
            args_pos = str_append(args_str, args_pos, sizeof(args_str), "&");
        }
        args_pos = str_append(args_str, args_pos, sizeof(args_str), val);
        i++;
    }

    // For scanln: add a space specifier after to consume remaining whitespace/newline
    if (scanln && fmt_pos > 0) {
        fmt_pos = str_append(fmt_str, fmt_pos, sizeof(fmt_str), " ");
    }

    char c_code[4096];
    if (args_pos > 0) {
        snprintf(c_code, 4096, "scanf(\"%s\", %s)", fmt_str, args_str);
    } else {
        // No args: just skip whitespace or newline
        if (scanln) {
            snprintf(c_code, 4096, "scanf(\" \");");
        } else {
            snprintf(c_code, 4096, "(void)0;");
        }
    }

    // Now escape the C code for JSON
    char* result = malloc(8192);
    int res_pos = 0;
    res_pos = str_append(result, res_pos, 8192, "{\"replacement_code\": \"");
    json_escape(result, &res_pos, 8192, c_code);
    res_pos = str_append(result, res_pos, 8192, "\"}");
    
    return result;
}

char* expand_scan(char* request) {
    return build_scanf(request, false);
}

char* expand_scanln(char* request) {
    return build_scanf(request, true);
}
