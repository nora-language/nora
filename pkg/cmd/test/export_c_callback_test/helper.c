typedef int (*my_cmp_t)(void*, void*);
typedef int (*my_enum_t)(void*, void*);

void call_stateless(my_cmp_t cb) {
    cb(0, 0);
}

void call_stateful(my_enum_t cb, void* user_data) {
    cb((void*)0x123, user_data);
}
