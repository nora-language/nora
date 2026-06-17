#include "runtime.h"

// --- PREMIUM SYNC RUNTIME ---
typedef struct fiber_waiter {
    fiber_info_t* fiber;
    struct fiber_waiter* next;
} fiber_waiter_t;

typedef struct {
    NR_ATOMIC_INT locked;
    fiber_waiter_t* waiters_head;
    fiber_waiter_t* waiters_tail;
    NR_MUTEX_T queue_lock;
} nr_fiber_mutex_t;

void* nr_sync_mutex_create(void) {
    nr_fiber_mutex_t* m = (nr_fiber_mutex_t*)malloc(sizeof(nr_fiber_mutex_t));
    if (m) {
        NR_ATOMIC_STORE(&m->locked, 0);
        m->waiters_head = m->waiters_tail = NULL;
        NR_MUTEX_INIT(&m->queue_lock);
    }
    return m;
}

void nr_sync_mutex_lock(void* m) {
    nr_fiber_mutex_t* mu = (nr_fiber_mutex_t*)m;
    if (!mu) return;
    
    NR_INT expected = 0;
    if (NR_ATOMIC_CAS(&mu->locked, &expected, 1)) {
        return;
    }
    
    fiber_info_t* self = (fiber_info_t*)GetFiberData();
    if (!self) {
        while (1) {
            expected = 0;
            if (NR_ATOMIC_CAS(&mu->locked, &expected, 1)) return;
#ifdef _WIN32
            Sleep(1);
#else
            usleep(1000);
#endif
        }
    }
    
    fiber_waiter_t* waiter = (fiber_waiter_t*)malloc(sizeof(fiber_waiter_t));
    waiter->fiber = self;
    waiter->next = NULL;
    
    NR_MUTEX_LOCK(&mu->queue_lock);
    expected = 0;
    if (NR_ATOMIC_CAS(&mu->locked, &expected, 1)) {
        NR_MUTEX_UNLOCK(&mu->queue_lock);
        free(waiter);
        return;
    }
    
    if (mu->waiters_tail) {
        mu->waiters_tail->next = waiter;
        mu->waiters_tail = waiter;
    } else {
        mu->waiters_head = mu->waiters_tail = waiter;
    }
    NR_MUTEX_UNLOCK(&mu->queue_lock);
    
    park();
}

void nr_sync_mutex_unlock(void* m) {
    nr_fiber_mutex_t* mu = (nr_fiber_mutex_t*)m;
    if (!mu) return;
    
    NR_MUTEX_LOCK(&mu->queue_lock);
    fiber_waiter_t* waiter = mu->waiters_head;
    if (waiter) {
        mu->waiters_head = waiter->next;
        if (!mu->waiters_head) {
            mu->waiters_tail = NULL;
        }
        NR_MUTEX_UNLOCK(&mu->queue_lock);
        
        resume(waiter->fiber);
        free(waiter);
    } else {
        NR_ATOMIC_STORE(&mu->locked, 0);
        NR_MUTEX_UNLOCK(&mu->queue_lock);
    }
}

void nr_sync_mutex_destroy(void* m) {
    nr_fiber_mutex_t* mu = (nr_fiber_mutex_t*)m;
    if (mu) {
        NR_MUTEX_DESTROY(&mu->queue_lock);
        fiber_waiter_t* curr = mu->waiters_head;
        while (curr) {
            fiber_waiter_t* next = curr->next;
            free(curr);
            curr = next;
        }
        free(mu);
    }
}

typedef struct {
    NR_ATOMIC_INT counter;
    fiber_waiter_t* waiters_head;
    fiber_waiter_t* waiters_tail;
    NR_MUTEX_T queue_lock;
    bool has_panic;
    const char* panic_msg;
    const char* panic_file;
    int panic_line;
} nr_fiber_waitgroup_t;

void* nr_sync_waitgroup_create(void) {
    nr_fiber_waitgroup_t* wg = (nr_fiber_waitgroup_t*)malloc(sizeof(nr_fiber_waitgroup_t));
    if (wg) {
        NR_ATOMIC_STORE(&wg->counter, 0);
        wg->waiters_head = wg->waiters_tail = NULL;
        NR_MUTEX_INIT(&wg->queue_lock);
        wg->has_panic = false;
        wg->panic_msg = NULL;
        wg->panic_file = NULL;
        wg->panic_line = 0;
    }
    return wg;
}

void nr_sync_waitgroup_add(void* w, int32_t delta) {
    nr_fiber_waitgroup_t* wg = (nr_fiber_waitgroup_t*)w;
    if (!wg) return;
    
    long prev = NR_ATOMIC_ADD(&wg->counter, delta);
    long current = prev + delta;
    
    if (current < 0) {
        nr_panic("negative WaitGroup counter", "sync", 0);
    }
    
    if (current == 0) {
        NR_MUTEX_LOCK(&wg->queue_lock);
        fiber_waiter_t* curr = wg->waiters_head;
        wg->waiters_head = wg->waiters_tail = NULL;
        NR_MUTEX_UNLOCK(&wg->queue_lock);
        
        while (curr) {
            fiber_waiter_t* next = curr->next;
            resume(curr->fiber);
            free(curr);
            curr = next;
        }
    }
}

void nr_sync_waitgroup_done(void* w) {
    nr_sync_waitgroup_add(w, -1);
}

void nr_sync_waitgroup_panic(void* w, const char* msg, const char* file, int line) {
    nr_fiber_waitgroup_t* wg = (nr_fiber_waitgroup_t*)w;
    if (!wg) return;

    NR_MUTEX_LOCK(&wg->queue_lock);
    if (!wg->has_panic) {
        wg->has_panic = true;
        wg->panic_msg = msg;
        wg->panic_file = file;
        wg->panic_line = line;
    }
    NR_MUTEX_UNLOCK(&wg->queue_lock);
}

void nr_sync_waitgroup_wait(void* w) {
    nr_fiber_waitgroup_t* wg = (nr_fiber_waitgroup_t*)w;
    if (!wg) return;
    
    if (NR_ATOMIC_LOAD(&wg->counter) == 0) {
        if (wg->has_panic) {
            nr_panic(wg->panic_msg, wg->panic_file, wg->panic_line);
        }
        return;
    }
    
    fiber_info_t* self = (fiber_info_t*)GetFiberData();
    if (!self) {
        while (NR_ATOMIC_LOAD(&wg->counter) > 0) {
#ifdef _WIN32
            Sleep(1);
#else
            usleep(1000);
#endif
        }
        if (wg->has_panic) {
            nr_panic(wg->panic_msg, wg->panic_file, wg->panic_line);
        }
        return;
    }
    
    fiber_waiter_t* waiter = (fiber_waiter_t*)malloc(sizeof(fiber_waiter_t));
    waiter->fiber = self;
    waiter->next = NULL;
    
    NR_MUTEX_LOCK(&wg->queue_lock);
    if (NR_ATOMIC_LOAD(&wg->counter) == 0) {
        NR_MUTEX_UNLOCK(&wg->queue_lock);
        free(waiter);
        if (wg->has_panic) {
            nr_panic(wg->panic_msg, wg->panic_file, wg->panic_line);
        }
        return;
    }
    
    if (wg->waiters_tail) {
        wg->waiters_tail->next = waiter;
        wg->waiters_tail = waiter;
    } else {
        wg->waiters_head = wg->waiters_tail = waiter;
    }
    NR_MUTEX_UNLOCK(&wg->queue_lock);
    
    park();

    if (wg->has_panic) {
        nr_panic(wg->panic_msg, wg->panic_file, wg->panic_line);
    }
}

void nr_sync_waitgroup_destroy(void* w) {
    nr_fiber_waitgroup_t* wg = (nr_fiber_waitgroup_t*)w;
    if (wg) {
        NR_MUTEX_DESTROY(&wg->queue_lock);
        fiber_waiter_t* curr = wg->waiters_head;
        while (curr) {
            fiber_waiter_t* next = curr->next;
            free(curr);
            curr = next;
        }
        free(wg);
    }
}

int32_t nr_sync_atomic_load(void* p) {
    if (!p) return 0;
    return (int32_t)NR_ATOMIC_LOAD((NR_ATOMIC_INT*)p);
}

void nr_sync_atomic_store(void* p, int32_t v) {
    if (p) {
        NR_ATOMIC_STORE((NR_ATOMIC_INT*)p, v);
    }
}

int32_t nr_sync_atomic_add(void* p, int32_t v) {
    if (!p) return 0;
    return (int32_t)NR_ATOMIC_ADD((NR_ATOMIC_INT*)p, v);
}

bool nr_sync_atomic_cas(void* p, void* exp, int32_t des) {
    if (!p || !exp) return false;
    return NR_ATOMIC_CAS((NR_ATOMIC_INT*)p, (int*)exp, des);
}

void nr_sync_add_i32(void* p, int32_t v) {
    if (p) {
        *(int32_t*)p += v;
    }
}

int32_t nr_sync_load_i32(void* p) {
    if (!p) return 0;
    return *(int32_t*)p;
}

typedef struct {
    NR_ATOMIC_INT reader_count;
    NR_ATOMIC_INT writer_active;
    fiber_waiter_t* readers_head;
    fiber_waiter_t* readers_tail;
    fiber_waiter_t* writers_head;
    fiber_waiter_t* writers_tail;
    NR_MUTEX_T queue_lock;
} nr_fiber_rwmutex_t;

void* nr_sync_rwmutex_create(void) {
    nr_fiber_rwmutex_t* rw = (nr_fiber_rwmutex_t*)malloc(sizeof(nr_fiber_rwmutex_t));
    if (rw) {
        NR_ATOMIC_STORE(&rw->reader_count, 0);
        NR_ATOMIC_STORE(&rw->writer_active, 0);
        rw->readers_head = rw->readers_tail = NULL;
        rw->writers_head = rw->writers_tail = NULL;
        NR_MUTEX_INIT(&rw->queue_lock);
    }
    return rw;
}

void nr_sync_rwmutex_rlock(void* m) {
    nr_fiber_rwmutex_t* rw = (nr_fiber_rwmutex_t*)m;
    if (!rw) return;

    fiber_info_t* self = (fiber_info_t*)GetFiberData();
    
    NR_MUTEX_LOCK(&rw->queue_lock);
    if (NR_ATOMIC_LOAD(&rw->writer_active) == 0 && rw->writers_head == NULL) {
        NR_ATOMIC_ADD(&rw->reader_count, 1);
        NR_MUTEX_UNLOCK(&rw->queue_lock);
        return;
    }
    
    if (!self) {
        NR_MUTEX_UNLOCK(&rw->queue_lock);
        while (1) {
            NR_MUTEX_LOCK(&rw->queue_lock);
            if (NR_ATOMIC_LOAD(&rw->writer_active) == 0 && rw->writers_head == NULL) {
                NR_ATOMIC_ADD(&rw->reader_count, 1);
                NR_MUTEX_UNLOCK(&rw->queue_lock);
                return;
            }
            NR_MUTEX_UNLOCK(&rw->queue_lock);
#ifdef _WIN32
            Sleep(1);
#else
            usleep(1000);
#endif
        }
    }
    
    fiber_waiter_t* waiter = (fiber_waiter_t*)malloc(sizeof(fiber_waiter_t));
    waiter->fiber = self;
    waiter->next = NULL;
    if (rw->readers_tail) {
        rw->readers_tail->next = waiter;
        rw->readers_tail = waiter;
    } else {
        rw->readers_head = rw->readers_tail = waiter;
    }
    NR_MUTEX_UNLOCK(&rw->queue_lock);
    
    park();
}

void nr_sync_rwmutex_runlock(void* m) {
    nr_fiber_rwmutex_t* rw = (nr_fiber_rwmutex_t*)m;
    if (!rw) return;
    
    NR_MUTEX_LOCK(&rw->queue_lock);
    NR_INT count = NR_ATOMIC_ADD(&rw->reader_count, -1);
    
    if ((count - 1) == 0 && rw->writers_head != NULL) {
        fiber_waiter_t* w = rw->writers_head;
        rw->writers_head = w->next;
        if (!rw->writers_head) rw->writers_tail = NULL;
        
        NR_ATOMIC_STORE(&rw->writer_active, 1);
        NR_MUTEX_UNLOCK(&rw->queue_lock);
        
        resume(w->fiber);
        free(w);
        return;
    }
    NR_MUTEX_UNLOCK(&rw->queue_lock);
}

void nr_sync_rwmutex_lock(void* m) {
    nr_fiber_rwmutex_t* rw = (nr_fiber_rwmutex_t*)m;
    if (!rw) return;
    
    fiber_info_t* self = (fiber_info_t*)GetFiberData();
    
    NR_MUTEX_LOCK(&rw->queue_lock);
    if (NR_ATOMIC_LOAD(&rw->writer_active) == 0 && NR_ATOMIC_LOAD(&rw->reader_count) == 0) {
        NR_ATOMIC_STORE(&rw->writer_active, 1);
        NR_MUTEX_UNLOCK(&rw->queue_lock);
        return;
    }
    
    if (!self) {
        NR_MUTEX_UNLOCK(&rw->queue_lock);
        while (1) {
            NR_MUTEX_LOCK(&rw->queue_lock);
            if (NR_ATOMIC_LOAD(&rw->writer_active) == 0 && NR_ATOMIC_LOAD(&rw->reader_count) == 0) {
                NR_ATOMIC_STORE(&rw->writer_active, 1);
                NR_MUTEX_UNLOCK(&rw->queue_lock);
                return;
            }
            NR_MUTEX_UNLOCK(&rw->queue_lock);
#ifdef _WIN32
            Sleep(1);
#else
            usleep(1000);
#endif
        }
    }
    
    fiber_waiter_t* waiter = (fiber_waiter_t*)malloc(sizeof(fiber_waiter_t));
    waiter->fiber = self;
    waiter->next = NULL;
    if (rw->writers_tail) {
        rw->writers_tail->next = waiter;
        rw->writers_tail = waiter;
    } else {
        rw->writers_head = rw->writers_tail = waiter;
    }
    NR_MUTEX_UNLOCK(&rw->queue_lock);
    
    park();
}

void nr_sync_rwmutex_unlock(void* m) {
    nr_fiber_rwmutex_t* rw = (nr_fiber_rwmutex_t*)m;
    if (!rw) return;
    
    NR_MUTEX_LOCK(&rw->queue_lock);
    NR_ATOMIC_STORE(&rw->writer_active, 0);
    
    if (rw->readers_head != NULL) {
        fiber_waiter_t* curr = rw->readers_head;
        rw->readers_head = rw->readers_tail = NULL;
        
        int count = 0;
        fiber_waiter_t* c = curr;
        while(c) { count++; c = c->next; }
        NR_ATOMIC_ADD(&rw->reader_count, count);
        
        NR_MUTEX_UNLOCK(&rw->queue_lock);
        
        while (curr) {
            fiber_waiter_t* next = curr->next;
            resume(curr->fiber);
            free(curr);
            curr = next;
        }
        return;
    }
    
    if (rw->writers_head != NULL) {
        fiber_waiter_t* w = rw->writers_head;
        rw->writers_head = w->next;
        if (!rw->writers_head) rw->writers_tail = NULL;
        
        NR_ATOMIC_STORE(&rw->writer_active, 1);
        NR_MUTEX_UNLOCK(&rw->queue_lock);
        
        resume(w->fiber);
        free(w);
        return;
    }
    
    NR_MUTEX_UNLOCK(&rw->queue_lock);
}

void nr_sync_rwmutex_destroy(void* m) {
    nr_fiber_rwmutex_t* rw = (nr_fiber_rwmutex_t*)m;
    if (rw) {
        NR_MUTEX_DESTROY(&rw->queue_lock);
        fiber_waiter_t* curr = rw->readers_head;
        while (curr) { fiber_waiter_t* n = curr->next; free(curr); curr = n; }
        curr = rw->writers_head;
        while (curr) { fiber_waiter_t* n = curr->next; free(curr); curr = n; }
        free(rw);
    }
}
