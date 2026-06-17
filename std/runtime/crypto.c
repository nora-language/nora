#include "runtime.h"

// --- PREMIUM CRYPTO RUNTIME ---
#ifdef _WIN32
#include <windows.h>
#endif

bool nr_crypto_rand_bytes(void* buf, int32_t size) {
    if (!buf || size <= 0) return false;
#ifdef _WIN32
    HMODULE hLib = LoadLibraryA("advapi32.dll");
    if (!hLib) return false;
    typedef BOOLEAN (WINAPI *RtlGenRandomFunc)(PVOID, ULONG);
    RtlGenRandomFunc fnRtlGenRandom = (RtlGenRandomFunc)GetProcAddress(hLib, "SystemFunction036");
    if (!fnRtlGenRandom) {
        FreeLibrary(hLib);
        return false;
    }
    BOOLEAN ok = fnRtlGenRandom(buf, (ULONG)size);
    FreeLibrary(hLib);
    return ok != 0;
#else
    FILE* f = fopen("/dev/urandom", "rb");
    if (!f) return false;
    size_t read_bytes = fread(buf, 1, (size_t)size, f);
    fclose(f);
    return read_bytes == (size_t)size;
#endif
}
