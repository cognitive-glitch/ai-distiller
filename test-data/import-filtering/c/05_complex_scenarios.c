// Test Pattern 5: Complex Scenarios
// Tests mixed includes with various patterns

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdarg.h>
#include <assert.h>
#include <limits.h>
#include <float.h>
#include <setjmp.h>
#include <locale.h>
#include <wchar.h>

// Local headers
#include "app/core.h"
#include "app/utils.h"
#include "lib/parser.h"
#include "lib/validator.h"

// Nested includes
#include "vendor/json/json.h"
#include "vendor/xml/parser.h"

// Platform-specific
#ifdef __GNUC__
    #include <execinfo.h>
#endif

// Not using: stdarg.h, limits.h, float.h, setjmp.h, locale.h, wchar.h
// Not using: app/utils.h, lib/validator.h, vendor/xml/parser.h, execinfo.h

typedef struct Node {
    void* data;
    struct Node* next;
} Node;

Node* create_node(void* data) {
    // Using stdlib.h
    Node* node = (Node*)malloc(sizeof(Node));
    node->data = data;
    node->next = NULL;
    return node;
}

void print_list(Node* head) {
    // Using stdio.h
    printf("List contents:\n");

    Node* current = head;
    while (current != NULL) {
        printf("  Node at %p\n", current->data);
        current = current->next;
    }
}

void free_list(Node* head) {
    Node* current = head;
    while (current != NULL) {
        Node* next = current->next;
        // Using stdlib.h
        free(current);
        current = next;
    }
}

char* duplicate_string(const char* str) {
    if (!str) return NULL;

    // Using string.h
    size_t len = strlen(str);

    // Using stdlib.h
    char* dup = (char*)malloc(len + 1);

    // Using string.h
    strcpy(dup, str);

    return dup;
}

int main() {
    // Using assert.h
    assert(sizeof(int) >= 4);

    // Using stdio.h
    printf("Starting complex test\n");

    // Create linked list
    Node* head = create_node("First");
    head->next = create_node("Second");
    head->next->next = create_node("Third");

    print_list(head);

    // Test string duplication
    char* str = duplicate_string("Test string");
    printf("Duplicated: %s\n", str);

    // Cleanup
    free(str);
    free_list(head);

    return 0;
}
