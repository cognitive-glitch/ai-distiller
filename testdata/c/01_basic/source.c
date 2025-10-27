#include <stdio.h>
#include <stdlib.h>
#include "myheader.h"

/**
 * A basic Point structure demonstrating fundamental C concepts
 */
struct Point {
    int x;
    int y;
};

/**
 * Add two integers
 */
int add(int a, int b) {
    return a + b;
}

/**
 * Multiply two integers
 */
int multiply(int x, int y) {
    return x * y;
}

/**
 * Static helper function (internal linkage)
 */
static int helper(void) {
    return 42;
}

/**
 * Process data with pointer parameters
 */
void process(int *data, char *buffer) {
    // Process data
}

/**
 * Initialize a point
 */
void init_point(struct Point *p, int x, int y) {
    p->x = x;
    p->y = y;
}

/**
 * Get point distance
 */
double get_distance(struct Point *p1, struct Point *p2) {
    int dx = p2->x - p1->x;
    int dy = p2->y - p1->y;
    return sqrt(dx * dx + dy * dy);
}
