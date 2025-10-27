#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdbool.h>
#include <pthread.h>

// Memory pool for efficient allocation
struct MemoryPool {
    void *memory;
    size_t block_size;
    size_t num_blocks;
    size_t free_blocks;
    uint8_t *free_list;
    pthread_mutex_t lock;
};

// Hash table entry
struct HashEntry {
    char *key;
    void *value;
    struct HashEntry *next;
};

// Hash table
struct HashTable {
    struct HashEntry **buckets;
    size_t capacity;
    size_t size;
    uint32_t (*hash_func)(const char *key);
    pthread_rwlock_t lock;
};

// Thread pool task
struct Task {
    void (*function)(void *arg);
    void *argument;
    struct Task *next;
};

// Thread pool
struct ThreadPool {
    pthread_t *threads;
    size_t thread_count;
    struct Task *task_queue;
    pthread_mutex_t queue_lock;
    pthread_cond_t queue_cond;
    bool shutdown;
};

// Memory pool operations
struct MemoryPool* mempool_create(size_t block_size, size_t num_blocks) {
    struct MemoryPool *pool = malloc(sizeof(struct MemoryPool));
    if (!pool) return NULL;

    pool->block_size = block_size;
    pool->num_blocks = num_blocks;
    pool->free_blocks = num_blocks;

    pool->memory = malloc(block_size * num_blocks);
    pool->free_list = malloc(num_blocks);

    if (!pool->memory || !pool->free_list) {
        free(pool->memory);
        free(pool->free_list);
        free(pool);
        return NULL;
    }

    for (size_t i = 0; i < num_blocks; i++) {
        pool->free_list[i] = 1;
    }

    pthread_mutex_init(&pool->lock, NULL);
    return pool;
}

void* mempool_alloc(struct MemoryPool *pool) {
    if (!pool) return NULL;

    pthread_mutex_lock(&pool->lock);

    for (size_t i = 0; i < pool->num_blocks; i++) {
        if (pool->free_list[i]) {
            pool->free_list[i] = 0;
            pool->free_blocks--;
            void *ptr = (uint8_t*)pool->memory + (i * pool->block_size);
            pthread_mutex_unlock(&pool->lock);
            return ptr;
        }
    }

    pthread_mutex_unlock(&pool->lock);
    return NULL;
}

void mempool_free(struct MemoryPool *pool, void *ptr) {
    if (!pool || !ptr) return;

    pthread_mutex_lock(&pool->lock);

    size_t offset = (uint8_t*)ptr - (uint8_t*)pool->memory;
    size_t index = offset / pool->block_size;

    if (index < pool->num_blocks) {
        pool->free_list[index] = 1;
        pool->free_blocks++;
    }

    pthread_mutex_unlock(&pool->lock);
}

void mempool_destroy(struct MemoryPool *pool) {
    if (!pool) return;

    pthread_mutex_destroy(&pool->lock);
    free(pool->memory);
    free(pool->free_list);
    free(pool);
}

// Hash table operations
static uint32_t default_hash(const char *key) {
    uint32_t hash = 5381;
    int c;

    while ((c = *key++)) {
        hash = ((hash << 5) + hash) + c;
    }

    return hash;
}

struct HashTable* hashtable_create(size_t capacity) {
    struct HashTable *table = malloc(sizeof(struct HashTable));
    if (!table) return NULL;

    table->buckets = calloc(capacity, sizeof(struct HashEntry*));
    if (!table->buckets) {
        free(table);
        return NULL;
    }

    table->capacity = capacity;
    table->size = 0;
    table->hash_func = default_hash;
    pthread_rwlock_init(&table->lock, NULL);

    return table;
}

bool hashtable_insert(struct HashTable *table, const char *key, void *value) {
    if (!table || !key) return false;

    pthread_rwlock_wrlock(&table->lock);

    uint32_t hash = table->hash_func(key);
    size_t index = hash % table->capacity;

    struct HashEntry *entry = table->buckets[index];
    while (entry) {
        if (strcmp(entry->key, key) == 0) {
            entry->value = value;
            pthread_rwlock_unlock(&table->lock);
            return true;
        }
        entry = entry->next;
    }

    struct HashEntry *new_entry = malloc(sizeof(struct HashEntry));
    if (!new_entry) {
        pthread_rwlock_unlock(&table->lock);
        return false;
    }

    new_entry->key = strdup(key);
    new_entry->value = value;
    new_entry->next = table->buckets[index];
    table->buckets[index] = new_entry;
    table->size++;

    pthread_rwlock_unlock(&table->lock);
    return true;
}

void* hashtable_get(struct HashTable *table, const char *key) {
    if (!table || !key) return NULL;

    pthread_rwlock_rdlock(&table->lock);

    uint32_t hash = table->hash_func(key);
    size_t index = hash % table->capacity;

    struct HashEntry *entry = table->buckets[index];
    while (entry) {
        if (strcmp(entry->key, key) == 0) {
            void *value = entry->value;
            pthread_rwlock_unlock(&table->lock);
            return value;
        }
        entry = entry->next;
    }

    pthread_rwlock_unlock(&table->lock);
    return NULL;
}

bool hashtable_remove(struct HashTable *table, const char *key) {
    if (!table || !key) return false;

    pthread_rwlock_wrlock(&table->lock);

    uint32_t hash = table->hash_func(key);
    size_t index = hash % table->capacity;

    struct HashEntry *entry = table->buckets[index];
    struct HashEntry *prev = NULL;

    while (entry) {
        if (strcmp(entry->key, key) == 0) {
            if (prev) {
                prev->next = entry->next;
            } else {
                table->buckets[index] = entry->next;
            }

            free(entry->key);
            free(entry);
            table->size--;

            pthread_rwlock_unlock(&table->lock);
            return true;
        }
        prev = entry;
        entry = entry->next;
    }

    pthread_rwlock_unlock(&table->lock);
    return false;
}

void hashtable_destroy(struct HashTable *table) {
    if (!table) return;

    for (size_t i = 0; i < table->capacity; i++) {
        struct HashEntry *entry = table->buckets[i];
        while (entry) {
            struct HashEntry *next = entry->next;
            free(entry->key);
            free(entry);
            entry = next;
        }
    }

    pthread_rwlock_destroy(&table->lock);
    free(table->buckets);
    free(table);
}

// Thread pool operations
static void* worker_thread(void *arg) {
    struct ThreadPool *pool = (struct ThreadPool*)arg;

    while (true) {
        pthread_mutex_lock(&pool->queue_lock);

        while (!pool->task_queue && !pool->shutdown) {
            pthread_cond_wait(&pool->queue_cond, &pool->queue_lock);
        }

        if (pool->shutdown) {
            pthread_mutex_unlock(&pool->queue_lock);
            break;
        }

        struct Task *task = pool->task_queue;
        if (task) {
            pool->task_queue = task->next;
        }

        pthread_mutex_unlock(&pool->queue_lock);

        if (task) {
            task->function(task->argument);
            free(task);
        }
    }

    return NULL;
}

struct ThreadPool* threadpool_create(size_t thread_count) {
    struct ThreadPool *pool = malloc(sizeof(struct ThreadPool));
    if (!pool) return NULL;

    pool->threads = malloc(sizeof(pthread_t) * thread_count);
    if (!pool->threads) {
        free(pool);
        return NULL;
    }

    pool->thread_count = thread_count;
    pool->task_queue = NULL;
    pool->shutdown = false;

    pthread_mutex_init(&pool->queue_lock, NULL);
    pthread_cond_init(&pool->queue_cond, NULL);

    for (size_t i = 0; i < thread_count; i++) {
        pthread_create(&pool->threads[i], NULL, worker_thread, pool);
    }

    return pool;
}

bool threadpool_submit(struct ThreadPool *pool, void (*function)(void*), void *arg) {
    if (!pool || !function) return false;

    struct Task *task = malloc(sizeof(struct Task));
    if (!task) return false;

    task->function = function;
    task->argument = arg;
    task->next = NULL;

    pthread_mutex_lock(&pool->queue_lock);

    if (!pool->task_queue) {
        pool->task_queue = task;
    } else {
        struct Task *current = pool->task_queue;
        while (current->next) {
            current = current->next;
        }
        current->next = task;
    }

    pthread_cond_signal(&pool->queue_cond);
    pthread_mutex_unlock(&pool->queue_lock);

    return true;
}

void threadpool_destroy(struct ThreadPool *pool) {
    if (!pool) return;

    pthread_mutex_lock(&pool->queue_lock);
    pool->shutdown = true;
    pthread_cond_broadcast(&pool->queue_cond);
    pthread_mutex_unlock(&pool->queue_lock);

    for (size_t i = 0; i < pool->thread_count; i++) {
        pthread_join(pool->threads[i], NULL);
    }

    struct Task *task = pool->task_queue;
    while (task) {
        struct Task *next = task->next;
        free(task);
        task = next;
    }

    pthread_mutex_destroy(&pool->queue_lock);
    pthread_cond_destroy(&pool->queue_cond);
    free(pool->threads);
    free(pool);
}

// Utility functions
static inline uint32_t next_power_of_two(uint32_t n) {
    n--;
    n |= n >> 1;
    n |= n >> 2;
    n |= n >> 4;
    n |= n >> 8;
    n |= n >> 16;
    n++;
    return n;
}

static int compare_int(const void *a, const void *b) {
    return (*(int*)a - *(int*)b);
}

static void swap(void *a, void *b, size_t size) {
    uint8_t *pa = (uint8_t*)a;
    uint8_t *pb = (uint8_t*)b;

    for (size_t i = 0; i < size; i++) {
        uint8_t temp = pa[i];
        pa[i] = pb[i];
        pb[i] = temp;
    }
}
