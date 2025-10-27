#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdbool.h>
#include <stdatomic.h>
#include <pthread.h>
#include <sys/mman.h>
#include <unistd.h>

// Lock-free queue node
struct LFQueueNode {
    void *data;
    _Atomic(struct LFQueueNode*) next;
};

// Lock-free queue
struct LFQueue {
    _Atomic(struct LFQueueNode*) head;
    _Atomic(struct LFQueueNode*) tail;
    _Atomic size_t size;
};

// Memory arena for fast allocation
struct Arena {
    uint8_t *buffer;
    size_t capacity;
    _Atomic size_t offset;
    struct Arena *next;
};

// Arena allocator
struct ArenaAllocator {
    struct Arena *current;
    size_t arena_size;
    pthread_mutex_t lock;
};

// Reference counted object
struct RefCounted {
    void *data;
    _Atomic int ref_count;
    void (*destructor)(void *data);
};

// Ring buffer for IPC
struct RingBuffer {
    uint8_t *buffer;
    size_t capacity;
    _Atomic size_t read_pos;
    _Atomic size_t write_pos;
    bool shared_memory;
};

// Lock-free queue operations
struct LFQueue* lfqueue_create(void) {
    struct LFQueue *queue = malloc(sizeof(struct LFQueue));
    if (!queue) return NULL;

    struct LFQueueNode *dummy = malloc(sizeof(struct LFQueueNode));
    if (!dummy) {
        free(queue);
        return NULL;
    }

    dummy->data = NULL;
    atomic_store(&dummy->next, NULL);

    atomic_store(&queue->head, dummy);
    atomic_store(&queue->tail, dummy);
    atomic_store(&queue->size, 0);

    return queue;
}

bool lfqueue_enqueue(struct LFQueue *queue, void *data) {
    if (!queue || !data) return false;

    struct LFQueueNode *node = malloc(sizeof(struct LFQueueNode));
    if (!node) return false;

    node->data = data;
    atomic_store(&node->next, NULL);

    while (true) {
        struct LFQueueNode *tail = atomic_load(&queue->tail);
        struct LFQueueNode *next = atomic_load(&tail->next);

        if (tail == atomic_load(&queue->tail)) {
            if (next == NULL) {
                if (atomic_compare_exchange_weak(&tail->next, &next, node)) {
                    atomic_compare_exchange_weak(&queue->tail, &tail, node);
                    atomic_fetch_add(&queue->size, 1);
                    return true;
                }
            } else {
                atomic_compare_exchange_weak(&queue->tail, &tail, next);
            }
        }
    }
}

void* lfqueue_dequeue(struct LFQueue *queue) {
    if (!queue) return NULL;

    while (true) {
        struct LFQueueNode *head = atomic_load(&queue->head);
        struct LFQueueNode *tail = atomic_load(&queue->tail);
        struct LFQueueNode *next = atomic_load(&head->next);

        if (head == atomic_load(&queue->head)) {
            if (head == tail) {
                if (next == NULL) {
                    return NULL;
                }
                atomic_compare_exchange_weak(&queue->tail, &tail, next);
            } else {
                void *data = next->data;
                if (atomic_compare_exchange_weak(&queue->head, &head, next)) {
                    free(head);
                    atomic_fetch_sub(&queue->size, 1);
                    return data;
                }
            }
        }
    }
}

size_t lfqueue_size(struct LFQueue *queue) {
    return queue ? atomic_load(&queue->size) : 0;
}

void lfqueue_destroy(struct LFQueue *queue) {
    if (!queue) return;

    while (lfqueue_dequeue(queue) != NULL);

    struct LFQueueNode *head = atomic_load(&queue->head);
    free(head);
    free(queue);
}

// Arena allocator operations
struct ArenaAllocator* arena_create(size_t arena_size) {
    struct ArenaAllocator *allocator = malloc(sizeof(struct ArenaAllocator));
    if (!allocator) return NULL;

    struct Arena *arena = malloc(sizeof(struct Arena));
    if (!arena) {
        free(allocator);
        return NULL;
    }

    arena->buffer = malloc(arena_size);
    if (!arena->buffer) {
        free(arena);
        free(allocator);
        return NULL;
    }

    arena->capacity = arena_size;
    atomic_store(&arena->offset, 0);
    arena->next = NULL;

    allocator->current = arena;
    allocator->arena_size = arena_size;
    pthread_mutex_init(&allocator->lock, NULL);

    return allocator;
}

void* arena_alloc(struct ArenaAllocator *allocator, size_t size) {
    if (!allocator || size == 0) return NULL;

    // Align to 8 bytes
    size = (size + 7) & ~7;

    pthread_mutex_lock(&allocator->lock);

    struct Arena *arena = allocator->current;
    size_t offset = atomic_load(&arena->offset);

    if (offset + size > arena->capacity) {
        // Need new arena
        struct Arena *new_arena = malloc(sizeof(struct Arena));
        if (!new_arena) {
            pthread_mutex_unlock(&allocator->lock);
            return NULL;
        }

        size_t new_size = allocator->arena_size;
        if (size > new_size) {
            new_size = size * 2;
        }

        new_arena->buffer = malloc(new_size);
        if (!new_arena->buffer) {
            free(new_arena);
            pthread_mutex_unlock(&allocator->lock);
            return NULL;
        }

        new_arena->capacity = new_size;
        atomic_store(&new_arena->offset, 0);
        new_arena->next = arena;

        allocator->current = new_arena;
        arena = new_arena;
        offset = 0;
    }

    void *ptr = arena->buffer + offset;
    atomic_store(&arena->offset, offset + size);

    pthread_mutex_unlock(&allocator->lock);
    return ptr;
}

void arena_reset(struct ArenaAllocator *allocator) {
    if (!allocator) return;

    pthread_mutex_lock(&allocator->lock);

    struct Arena *arena = allocator->current;
    while (arena) {
        atomic_store(&arena->offset, 0);
        arena = arena->next;
    }

    pthread_mutex_unlock(&allocator->lock);
}

void arena_destroy(struct ArenaAllocator *allocator) {
    if (!allocator) return;

    struct Arena *arena = allocator->current;
    while (arena) {
        struct Arena *next = arena->next;
        free(arena->buffer);
        free(arena);
        arena = next;
    }

    pthread_mutex_destroy(&allocator->lock);
    free(allocator);
}

// Reference counting operations
struct RefCounted* refcount_create(void *data, void (*destructor)(void*)) {
    struct RefCounted *rc = malloc(sizeof(struct RefCounted));
    if (!rc) return NULL;

    rc->data = data;
    atomic_store(&rc->ref_count, 1);
    rc->destructor = destructor;

    return rc;
}

void refcount_retain(struct RefCounted *rc) {
    if (!rc) return;
    atomic_fetch_add(&rc->ref_count, 1);
}

void refcount_release(struct RefCounted *rc) {
    if (!rc) return;

    int old_count = atomic_fetch_sub(&rc->ref_count, 1);
    if (old_count == 1) {
        if (rc->destructor) {
            rc->destructor(rc->data);
        }
        free(rc);
    }
}

int refcount_get(struct RefCounted *rc) {
    return rc ? atomic_load(&rc->ref_count) : 0;
}

// Ring buffer operations
struct RingBuffer* ringbuffer_create(size_t capacity, bool use_shared_memory) {
    struct RingBuffer *rb = malloc(sizeof(struct RingBuffer));
    if (!rb) return NULL;

    if (use_shared_memory) {
        rb->buffer = mmap(NULL, capacity, PROT_READ | PROT_WRITE,
                         MAP_SHARED | MAP_ANONYMOUS, -1, 0);
        if (rb->buffer == MAP_FAILED) {
            free(rb);
            return NULL;
        }
    } else {
        rb->buffer = malloc(capacity);
        if (!rb->buffer) {
            free(rb);
            return NULL;
        }
    }

    rb->capacity = capacity;
    atomic_store(&rb->read_pos, 0);
    atomic_store(&rb->write_pos, 0);
    rb->shared_memory = use_shared_memory;

    return rb;
}

size_t ringbuffer_write(struct RingBuffer *rb, const void *data, size_t size) {
    if (!rb || !data || size == 0) return 0;

    size_t write_pos = atomic_load(&rb->write_pos);
    size_t read_pos = atomic_load(&rb->read_pos);

    size_t available = rb->capacity - (write_pos - read_pos);
    if (size > available) {
        size = available;
    }

    size_t write_idx = write_pos % rb->capacity;
    size_t first_chunk = rb->capacity - write_idx;

    if (size <= first_chunk) {
        memcpy(rb->buffer + write_idx, data, size);
    } else {
        memcpy(rb->buffer + write_idx, data, first_chunk);
        memcpy(rb->buffer, (uint8_t*)data + first_chunk, size - first_chunk);
    }

    atomic_store(&rb->write_pos, write_pos + size);
    return size;
}

size_t ringbuffer_read(struct RingBuffer *rb, void *data, size_t size) {
    if (!rb || !data || size == 0) return 0;

    size_t write_pos = atomic_load(&rb->write_pos);
    size_t read_pos = atomic_load(&rb->read_pos);

    size_t available = write_pos - read_pos;
    if (size > available) {
        size = available;
    }

    size_t read_idx = read_pos % rb->capacity;
    size_t first_chunk = rb->capacity - read_idx;

    if (size <= first_chunk) {
        memcpy(data, rb->buffer + read_idx, size);
    } else {
        memcpy(data, rb->buffer + read_idx, first_chunk);
        memcpy((uint8_t*)data + first_chunk, rb->buffer, size - first_chunk);
    }

    atomic_store(&rb->read_pos, read_pos + size);
    return size;
}

size_t ringbuffer_available(struct RingBuffer *rb) {
    if (!rb) return 0;

    size_t write_pos = atomic_load(&rb->write_pos);
    size_t read_pos = atomic_load(&rb->read_pos);

    return write_pos - read_pos;
}

void ringbuffer_destroy(struct RingBuffer *rb) {
    if (!rb) return;

    if (rb->shared_memory) {
        munmap(rb->buffer, rb->capacity);
    } else {
        free(rb->buffer);
    }

    free(rb);
}

// Advanced utility functions
static inline bool is_power_of_two(size_t n) {
    return n && !(n & (n - 1));
}

static inline size_t align_up(size_t n, size_t alignment) {
    return (n + alignment - 1) & ~(alignment - 1);
}

static inline size_t align_down(size_t n, size_t alignment) {
    return n & ~(alignment - 1);
}

static uint64_t hash_fnv1a(const void *data, size_t len) {
    const uint8_t *bytes = (const uint8_t*)data;
    uint64_t hash = 14695981039346656037ULL;

    for (size_t i = 0; i < len; i++) {
        hash ^= bytes[i];
        hash *= 1099511628211ULL;
    }

    return hash;
}

static void memory_barrier(void) {
    atomic_thread_fence(memory_order_seq_cst);
}

static bool atomic_cas_ptr(void **ptr, void *expected, void *desired) {
    return atomic_compare_exchange_strong((_Atomic(void*)*) ptr, &expected, desired);
}

// Performance monitoring
struct PerfCounter {
    _Atomic uint64_t count;
    _Atomic uint64_t total_time;
    _Atomic uint64_t min_time;
    _Atomic uint64_t max_time;
};

void perf_counter_init(struct PerfCounter *counter) {
    if (!counter) return;

    atomic_store(&counter->count, 0);
    atomic_store(&counter->total_time, 0);
    atomic_store(&counter->min_time, UINT64_MAX);
    atomic_store(&counter->max_time, 0);
}

void perf_counter_record(struct PerfCounter *counter, uint64_t time) {
    if (!counter) return;

    atomic_fetch_add(&counter->count, 1);
    atomic_fetch_add(&counter->total_time, time);

    uint64_t current_min = atomic_load(&counter->min_time);
    while (time < current_min) {
        if (atomic_compare_exchange_weak(&counter->min_time, &current_min, time)) {
            break;
        }
    }

    uint64_t current_max = atomic_load(&counter->max_time);
    while (time > current_max) {
        if (atomic_compare_exchange_weak(&counter->max_time, &current_max, time)) {
            break;
        }
    }
}

uint64_t perf_counter_avg(struct PerfCounter *counter) {
    if (!counter) return 0;

    uint64_t count = atomic_load(&counter->count);
    if (count == 0) return 0;

    uint64_t total = atomic_load(&counter->total_time);
    return total / count;
}
