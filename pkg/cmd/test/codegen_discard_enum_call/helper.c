#include <stdint.h>

int32_t g_was_called = 0;

int32_t test_set_called(void* arg) {
    g_was_called = 1;
    return 0;
}

int32_t test_get_called() {
    return g_was_called;
}
