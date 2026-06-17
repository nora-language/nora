#include "runtime.h"

// --- CHANNEL RUNTIME ---
#ifdef _WIN32
#include <windows.h>
#elif defined(__EMSCRIPTEN__)
#include <emscripten.h>
#elif defined(__wasm__)
// Native Wasm headers (if any)
#else
#include <pthread.h>
#include <stdatomic.h>
#include <sched.h>
#endif
#include <string.h>

// Mutexes are defined in Base Macros

typedef struct {
    NR_ATOMIC_INT value;
    NR_ATOMIC_INT ref_count;
} select_state_t;

typedef struct {
    fiber_info_t* info;
    void* data_ptr;
    select_state_t* state;
    int index;
} waiter_t;

void select_state_release(select_state_t* s) {
    if (!s) return;
    NR_ATOMIC_SUB(&s->ref_count, 1);
}

typedef struct waiter_node {
    waiter_t waiter;
    struct waiter_node* next;
} waiter_node_t;

typedef struct {
    waiter_node_t* head;
    waiter_node_t* tail;
    NR_MUTEX_T lock;
} waiter_queue_t;

struct channel_s {
    int capacity;
    int elem_size;
    int size;
    int head;
    int tail;
    void* buffer;
    waiter_queue_t senders;
    waiter_queue_t receivers;
    NR_MUTEX_T lock;
    NR_ATOMIC_INT ref_count;
};

void waiter_queue_init(waiter_queue_t* q) {
    q->head = q->tail = NULL;
    NR_MUTEX_INIT(&q->lock);
}

void waiter_queue_push(waiter_queue_t* q, fiber_info_t* info, void* ptr, select_state_t* state, int index) {
    waiter_node_t* node = (waiter_node_t*)malloc(sizeof(waiter_node_t));
    node->waiter.info = info;
    node->waiter.data_ptr = ptr;
    node->waiter.state = state;
    node->waiter.index = index;
    node->next = NULL;
    if (state) {
        NR_ATOMIC_INC(&state->ref_count);
    }
    
    NR_MUTEX_LOCK(&q->lock);
    if (q->tail) {
        q->tail->next = node;
        q->tail = node;
    } else {
        q->head = q->tail = node;
    }
    NR_MUTEX_UNLOCK(&q->lock);
}

waiter_t waiter_queue_pop(waiter_queue_t* q) {
    NR_MUTEX_LOCK(&q->lock);
    if (!q->head) {
        NR_MUTEX_UNLOCK(&q->lock);
        return (waiter_t){0};
    }
    waiter_node_t* node = q->head;
    waiter_t w = node->waiter;
    q->head = node->next;
    if (!q->head) q->tail = NULL;
    NR_MUTEX_UNLOCK(&q->lock);
    free(node);
    return w;
}

void waiter_queue_remove(waiter_queue_t* q, select_state_t* state) {
    if (!state) return;
    NR_MUTEX_LOCK(&q->lock);
    waiter_node_t* curr = q->head;
    waiter_node_t* prev = NULL;
    while (curr) {
        if (curr->waiter.state == state) {
            waiter_node_t* next = curr->next;
            if (prev) {
                prev->next = next;
            } else {
                q->head = next;
            }
            if (curr == q->tail) {
                q->tail = prev;
            }
            free(curr);
            curr = next;
            NR_ATOMIC_SUB(&state->ref_count, 1);
        } else {
            prev = curr;
            curr = curr->next;
        }
    }
    NR_MUTEX_UNLOCK(&q->lock);
}

channel_t* channel_make(int capacity, int elem_size) {
    channel_t* c = (channel_t*)malloc(sizeof(channel_t));
    waiter_queue_init(&c->senders);
    waiter_queue_init(&c->receivers);
    c->capacity = capacity;
    c->elem_size = elem_size;
    c->size = 0;
    c->head = 0;
    c->tail = 0;
    NR_ATOMIC_STORE(&c->ref_count, 1);
    c->buffer = capacity > 0 ? malloc(capacity * elem_size) : NULL;
    NR_MUTEX_INIT(&c->lock);
    return c;
}

void channel_send(channel_t* c, void* val) {
    if (!c) return;
    NR_MUTEX_LOCK(&c->lock);
    waiter_t r;
    while (1) {
        r = waiter_queue_pop(&c->receivers);
        if (!r.info) break;
        if (r.state) {
            NR_INT expected = 0;
            if (!NR_ATOMIC_CAS(&r.state->value, &expected, r.index)) {
                select_state_release(r.state);
                continue;
            }
        }
        break;
    }
    if (r.info) {
        memcpy(r.data_ptr, val, c->elem_size);
        fiber_info_t* f = r.info;
        select_state_t* s = r.state;
        NR_MUTEX_UNLOCK(&c->lock);
        resume(f);
        if (s) select_state_release(s);
        return;
    }
    if (c->size < c->capacity) {
        memcpy((char*)c->buffer + (c->tail * c->elem_size), val, c->elem_size);
        c->tail = (c->tail + 1) % c->capacity;
        c->size++;
        NR_MUTEX_UNLOCK(&c->lock);
        return;
    }

    select_state_t state;
    NR_ATOMIC_STORE(&state.value, 0);
    NR_ATOMIC_STORE(&state.ref_count, 1);
    waiter_queue_push(&c->senders, (fiber_info_t*)GetFiberData(), val, &state, 1);
    NR_MUTEX_UNLOCK(&c->lock);

    while (1) {
        park();
        if (NR_ATOMIC_LOAD(&state.value) > 0) break;
    }
    waiter_queue_remove(&c->senders, &state);
    while (NR_ATOMIC_LOAD(&state.ref_count) > 1) {
        #ifdef _WIN32
        SwitchToThread();
        #else
        sched_yield();
        #endif
    }
}

void channel_recv(channel_t* c, void* res) {
    if (!c) return;
    NR_MUTEX_LOCK(&c->lock);
    if (c->size > 0) {
        memcpy(res, (char*)c->buffer + (c->head * c->elem_size), c->elem_size);
        c->head = (c->head + 1) % c->capacity;
        c->size--;
        waiter_t s;
        while (1) {
            s = waiter_queue_pop(&c->senders);
            if (!s.info) break;
            if (s.state) {
                NR_INT expected = 0;
                if (!NR_ATOMIC_CAS(&s.state->value, &expected, s.index)) {
                    select_state_release(s.state);
                    continue;
                }
            }
            break;
        }
        if (s.info) {
            memcpy((char*)c->buffer + (c->tail * c->elem_size), s.data_ptr, c->elem_size);
            c->tail = (c->tail + 1) % c->capacity;
            c->size++;
            fiber_info_t* f = s.info;
            select_state_t* st = s.state;
            NR_MUTEX_UNLOCK(&c->lock);
            resume(f);
            if (st) select_state_release(st);
            return;
        }
        NR_MUTEX_UNLOCK(&c->lock);
        return;
    }
    waiter_t s;
    while (1) {
        s = waiter_queue_pop(&c->senders);
        if (!s.info) break;
        if (s.state) {
            NR_INT expected = 0;
            if (!NR_ATOMIC_CAS(&s.state->value, &expected, s.index)) {
                select_state_release(s.state);
                continue;
            }
        }
        break;
    }
    if (s.info) {
        memcpy(res, s.data_ptr, c->elem_size);
        fiber_info_t* f = s.info;
        select_state_t* st = s.state;
        NR_MUTEX_UNLOCK(&c->lock);
        resume(f);
        if (st) select_state_release(st);
        return;
    }
    select_state_t state;
    NR_ATOMIC_STORE(&state.value, 0);
    NR_ATOMIC_STORE(&state.ref_count, 1);
    waiter_queue_push(&c->receivers, (fiber_info_t*)GetFiberData(), res, &state, 1);
    NR_MUTEX_UNLOCK(&c->lock);

    while (1) {
        park();
        if (NR_ATOMIC_LOAD(&state.value) > 0) break;
    }
    waiter_queue_remove(&c->receivers, &state);
    while (NR_ATOMIC_LOAD(&state.ref_count) > 1) {
        #ifdef _WIN32
        SwitchToThread();
        #else
        sched_yield();
        #endif
    }
}

void channel_ref(channel_t* c) {
    if (c) NR_ATOMIC_INC(&c->ref_count);
}

void channel_destroy(channel_t* c) {
    NR_MUTEX_DESTROY(&c->lock);
    NR_MUTEX_DESTROY(&c->senders.lock);
    NR_MUTEX_DESTROY(&c->receivers.lock);
    if (c->buffer) free(c->buffer);
    free(c);
}

void channel_free(channel_t* c) {
    if (!c) return;
    if (NR_ATOMIC_SUB(&c->ref_count, 1) == 1) {
        channel_destroy(c);
    }
}

int channel_select(select_op_t* ops, int count, bool has_default) {
    while (1) {
        for (int i = 0; i < count; i++) {
            channel_t* c = ops[i].chan;
            if (!c) continue;
            NR_MUTEX_LOCK(&c->lock);
            if (ops[i].is_send) {
                waiter_t r;
                while (1) {
                    r = waiter_queue_pop(&c->receivers);
                    if (!r.info) break;
                    if (r.state) {
                        NR_INT expected = 0;
                        if (!NR_ATOMIC_CAS(&r.state->value, &expected, r.index)) {
                            select_state_release(r.state);
                            continue;
                        }
                    }
                    break;
                }
                if (r.info) {
                    memcpy(r.data_ptr, ops[i].data, c->elem_size);
                    fiber_info_t* f = r.info;
                    select_state_t* s = r.state;
                    NR_MUTEX_UNLOCK(&c->lock);
                    resume(f);
                    if (s) select_state_release(s);
                    return i;
                }
                if (c->size < c->capacity) {
                    memcpy((char*)c->buffer + (c->tail * c->elem_size), ops[i].data, c->elem_size);
                    c->tail = (c->tail + 1) % c->capacity;
                    c->size++;
                    NR_MUTEX_UNLOCK(&c->lock);
                    return i;
                }
            } else {
                if (c->size > 0) {
                    memcpy(ops[i].data, (char*)c->buffer + (c->head * c->elem_size), c->elem_size);
                    c->head = (c->head + 1) % c->capacity;
                    c->size--;
                    waiter_t s;
                    while (1) {
                        s = waiter_queue_pop(&c->senders);
                        if (!s.info) break;
                        if (s.state) {
                            NR_INT expected = 0;
                            if (!NR_ATOMIC_CAS(&s.state->value, &expected, s.index)) {
                                select_state_release(s.state);
                                continue;
                            }
                        }
                        break;
                    }
                    if (s.info) {
                        memcpy((char*)c->buffer + (c->tail * c->elem_size), s.data_ptr, c->elem_size);
                        c->tail = (c->tail + 1) % c->capacity;
                        c->size++;
                        fiber_info_t* f = s.info;
                        select_state_t* st = s.state;
                        NR_MUTEX_UNLOCK(&c->lock);
                        resume(f);
                        if (st) select_state_release(st);
                    } else {
                        NR_MUTEX_UNLOCK(&c->lock);
                    }
                    return i;
                }
                waiter_t s;
                while (1) {
                    s = waiter_queue_pop(&c->senders);
                    if (!s.info) break;
                    if (s.state) {
                        NR_INT expected = 0;
                        if (!NR_ATOMIC_CAS(&s.state->value, &expected, s.index)) {
                            select_state_release(s.state);
                            continue;
                        }
                    }
                    break;
                }
                if (s.info) {
                    memcpy(ops[i].data, s.data_ptr, c->elem_size);
                    fiber_info_t* f = s.info;
                    select_state_t* st = s.state;
                    NR_MUTEX_UNLOCK(&c->lock);
                    resume(f);
                    if (st) select_state_release(st);
                    return i;
                }
            }
            NR_MUTEX_UNLOCK(&c->lock);
        }
        
        if (has_default) return -1;

        select_state_t state;
        NR_ATOMIC_STORE(&state.value, 0);
        NR_ATOMIC_STORE(&state.ref_count, 1);
        fiber_info_t* self = (fiber_info_t*)GetFiberData();
        
        for (int i = 0; i < count; i++) {
            channel_t* c = ops[i].chan;
            if (!c) continue;
            NR_MUTEX_LOCK(&c->lock);
            if (ops[i].is_send) {
                waiter_queue_push(&c->senders, self, ops[i].data, &state, i + 1);
            } else {
                waiter_queue_push(&c->receivers, self, ops[i].data, &state, i + 1);
            }
            NR_MUTEX_UNLOCK(&c->lock);
        }
        
        park();
        for (int i = 0; i < count; i++) {
            channel_t* c = ops[i].chan;
            if (!c) continue;
            if (ops[i].is_send) {
                waiter_queue_remove(&c->senders, &state);
            } else {
                waiter_queue_remove(&c->receivers, &state);
            }
        }
        while (NR_ATOMIC_LOAD(&state.ref_count) > 1) {
            #ifdef _WIN32
            SwitchToThread();
            #else
            sched_yield();
            #endif
        }
        int result = NR_ATOMIC_LOAD(&state.value);
        if (result > 0) return result - 1;
    }
}
