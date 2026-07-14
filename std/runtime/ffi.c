#include "runtime.h"

void nr_write_u8(void* p, uint8_t val) { if (p) { *((uint8_t*)p) = val; } }
uint8_t nr_read_u8(void* p) { return p ? *((uint8_t*)p) : 0; }

void nr_write_i8(void* p, int8_t val) { if (p) { *((int8_t*)p) = val; } }
int8_t nr_read_i8(void* p) { return p ? *((int8_t*)p) : 0; }

void nr_write_u16(void* p, uint16_t val) { if (p) { *((uint16_t*)p) = val; } }
uint16_t nr_read_u16(void* p) { return p ? *((uint16_t*)p) : 0; }

void nr_write_i16(void* p, int16_t val) { if (p) { *((int16_t*)p) = val; } }
int16_t nr_read_i16(void* p) { return p ? *((int16_t*)p) : 0; }

void nr_write_u32(void* p, uint32_t val) { if (p) { *((uint32_t*)p) = val; } }
uint32_t nr_read_u32(void* p) { return p ? *((uint32_t*)p) : 0; }

void nr_write_i32(void* p, int32_t val) { if (p) { *((int32_t*)p) = val; } }
int32_t nr_read_i32(void* p) { return p ? *((int32_t*)p) : 0; }

void nr_write_u64(void* p, uint64_t val) { if (p) { *((uint64_t*)p) = val; } }
uint64_t nr_read_u64(void* p) { return p ? *((uint64_t*)p) : 0; }

void nr_write_i64(void* p, int64_t val) { if (p) { *((int64_t*)p) = val; } }
int64_t nr_read_i64(void* p) { return p ? *((int64_t*)p) : 0; }

void nr_write_f32(void* p, float val) { if (p) { *((float*)p) = val; } }
float nr_read_f32(void* p) { return p ? *((float*)p) : 0.0f; }

void nr_write_f64(void* p, double val) { if (p) { *((double*)p) = val; } }
double nr_read_f64(void* p) { return p ? *((double*)p) : 0.0; }

void nr_write_ptr(void* p, void* val) { if (p) { *((void**)p) = val; } }
void* nr_read_ptr(void* p) { return p ? *((void**)p) : NULL; }
