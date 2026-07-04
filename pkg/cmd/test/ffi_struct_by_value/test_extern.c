#include <stdio.h>
#include <stdlib.h>

struct MyStruct {
    int a;
    int b;
    void* ptr;
};

void TestByValue(struct MyStruct s) {
    if (s.a == 42 && s.b == 100 && s.ptr == NULL) {
        printf("TestByValue passed!\n");
        exit(0);
    }
    printf("TestByValue failed: %d, %d, %p\n", s.a, s.b, s.ptr);
    exit(1);
}
