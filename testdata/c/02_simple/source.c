#include <stdio.h>
#include <string.h>
#include <stddef.h>

// Color enumeration
enum Color {
    RED,
    GREEN,
    BLUE
};

// Data union
union Data {
    int i;
    float f;
    char str[20];
};

// Node structure with self-reference
struct Node {
    int data;
    struct Node *next;
};

// Rectangle structure
struct Rectangle {
    int width;
    int height;
    enum Color color;
};

// Function prototypes
void initialize(void);
int calculate(int x, int y);

// Create a new node
struct Node* create_node(int value) {
    struct Node *node = malloc(sizeof(struct Node));
    if (node) {
        node->data = value;
        node->next = NULL;
    }
    return node;
}

// Free a linked list
void free_list(struct Node *head) {
    struct Node *current = head;
    while (current) {
        struct Node *next = current->next;
        free(current);
        current = next;
    }
}

// Calculate rectangle area
int rect_area(struct Rectangle *rect) {
    return rect->width * rect->height;
}

// Set union integer value
void set_int(union Data *data, int value) {
    data->i = value;
}

// Get color name
const char* get_color_name(enum Color color) {
    switch (color) {
        case RED: return "Red";
        case GREEN: return "Green";
        case BLUE: return "Blue";
        default: return "Unknown";
    }
}

// Static utility function
static void internal_helper(void) {
    // Internal implementation
}

// Variadic function
int sum_all(int count, ...) {
    // Sum variable arguments
    return 0;
}
