#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>

// Forward declarations
struct List;
struct Iterator;

// Callback function type
typedef int (*compare_fn)(const void *a, const void *b);
typedef void (*free_fn)(void *data);

// Generic list node
struct ListNode {
    void *data;
    struct ListNode *next;
    struct ListNode *prev;
};

// Doubly linked list
struct List {
    struct ListNode *head;
    struct ListNode *tail;
    size_t size;
    compare_fn compare;
    free_fn free_data;
};

// Iterator for list traversal
struct Iterator {
    struct ListNode *current;
    struct List *list;
};

// Status codes
enum Status {
    STATUS_OK = 0,
    STATUS_ERROR = -1,
    STATUS_NOT_FOUND = -2,
    STATUS_INVALID = -3
};

// Configuration structure
struct Config {
    char *name;
    int max_size;
    bool enabled;
    double threshold;
};

// Create a new list
struct List* list_create(compare_fn cmp, free_fn free_func) {
    struct List *list = malloc(sizeof(struct List));
    if (list) {
        list->head = NULL;
        list->tail = NULL;
        list->size = 0;
        list->compare = cmp;
        list->free_data = free_func;
    }
    return list;
}

// Destroy a list
void list_destroy(struct List *list) {
    if (!list) return;

    struct ListNode *current = list->head;
    while (current) {
        struct ListNode *next = current->next;
        if (list->free_data) {
            list->free_data(current->data);
        }
        free(current);
        current = next;
    }
    free(list);
}

// Add element to list
enum Status list_add(struct List *list, void *data) {
    if (!list || !data) {
        return STATUS_INVALID;
    }

    struct ListNode *node = malloc(sizeof(struct ListNode));
    if (!node) {
        return STATUS_ERROR;
    }

    node->data = data;
    node->next = NULL;
    node->prev = list->tail;

    if (list->tail) {
        list->tail->next = node;
    } else {
        list->head = node;
    }

    list->tail = node;
    list->size++;

    return STATUS_OK;
}

// Remove element from list
enum Status list_remove(struct List *list, void *data) {
    if (!list || !data) {
        return STATUS_INVALID;
    }

    struct ListNode *current = list->head;
    while (current) {
        if (list->compare && list->compare(current->data, data) == 0) {
            if (current->prev) {
                current->prev->next = current->next;
            } else {
                list->head = current->next;
            }

            if (current->next) {
                current->next->prev = current->prev;
            } else {
                list->tail = current->prev;
            }

            if (list->free_data) {
                list->free_data(current->data);
            }
            free(current);
            list->size--;

            return STATUS_OK;
        }
        current = current->next;
    }

    return STATUS_NOT_FOUND;
}

// Get list size
size_t list_size(const struct List *list) {
    return list ? list->size : 0;
}

// Create iterator
struct Iterator* list_iterator(struct List *list) {
    if (!list) return NULL;

    struct Iterator *iter = malloc(sizeof(struct Iterator));
    if (iter) {
        iter->current = list->head;
        iter->list = list;
    }
    return iter;
}

// Check if iterator has next
bool iterator_has_next(struct Iterator *iter) {
    return iter && iter->current != NULL;
}

// Get next element
void* iterator_next(struct Iterator *iter) {
    if (!iter || !iter->current) {
        return NULL;
    }

    void *data = iter->current->data;
    iter->current = iter->current->next;
    return data;
}

// Free iterator
void iterator_free(struct Iterator *iter) {
    free(iter);
}

// Static helper functions
static struct ListNode* find_node(struct List *list, void *data) {
    struct ListNode *current = list->head;
    while (current) {
        if (list->compare && list->compare(current->data, data) == 0) {
            return current;
        }
        current = current->next;
    }
    return NULL;
}

static void swap_nodes(struct ListNode *a, struct ListNode *b) {
    void *temp = a->data;
    a->data = b->data;
    b->data = temp;
}

// Sort list using bubble sort
void list_sort(struct List *list) {
    if (!list || !list->compare || list->size < 2) {
        return;
    }

    bool swapped;
    do {
        swapped = false;
        struct ListNode *current = list->head;

        while (current && current->next) {
            if (list->compare(current->data, current->next->data) > 0) {
                swap_nodes(current, current->next);
                swapped = true;
            }
            current = current->next;
        }
    } while (swapped);
}

// Reverse list
void list_reverse(struct List *list) {
    if (!list || list->size < 2) {
        return;
    }

    struct ListNode *current = list->head;
    struct ListNode *temp = NULL;

    list->tail = list->head;

    while (current) {
        temp = current->prev;
        current->prev = current->next;
        current->next = temp;
        current = current->prev;
    }

    if (temp) {
        list->head = temp->prev;
    }
}

// Configuration management
struct Config* config_create(const char *name) {
    struct Config *config = malloc(sizeof(struct Config));
    if (config) {
        config->name = strdup(name);
        config->max_size = 100;
        config->enabled = true;
        config->threshold = 0.5;
    }
    return config;
}

void config_destroy(struct Config *config) {
    if (config) {
        free(config->name);
        free(config);
    }
}

bool config_is_valid(const struct Config *config) {
    return config && config->name && config->max_size > 0;
}
