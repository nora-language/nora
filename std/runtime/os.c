#include "runtime.h"

// --- PREMIUM OS RUNTIME ---
int nr_get_argc(void) {
    return nr_argc;
}

char* nr_get_argv(int index) {
    if (index >= 0 && index < nr_argc) {
        return nr_strdup(nr_argv[index]);
    }
    return "";
}

char* nr_os_getenv(char* key) {
    char* val = getenv(key);
    if (val != NULL) {
        return nr_strdup(val);
    }
    return "";
}

bool nr_os_setenv(char* key, char* val) {
#ifdef _WIN32
    return _putenv_s(key, val) == 0;
#else
    return setenv(key, val, 1) == 0;
#endif
}

bool nr_os_unsetenv(char* key) {
#ifdef _WIN32
    return _putenv_s(key, "") == 0;
#else
    return unsetenv(key) == 0;
#endif
}

char* nr_os_getcwd(void) {
    char buffer[1024];
#ifdef _WIN32
    if (_getcwd(buffer, sizeof(buffer)) != NULL) {
        return nr_strdup(buffer);
    }
#else
    if (getcwd(buffer, sizeof(buffer)) != NULL) {
        return nr_strdup(buffer);
    }
#endif
    return "";
}
