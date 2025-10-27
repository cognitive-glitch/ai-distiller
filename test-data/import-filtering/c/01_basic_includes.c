// Test Pattern 1: Basic Includes
// Tests standard library includes

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <time.h>
#include <ctype.h>
#include <stdint.h>
#include <stdbool.h>
#include "myheader.h"
#include "utils/helpers.h"

// Not using: time.h, ctype.h, stdint.h, stdbool.h, utils/helpers.h

typedef struct {
    char name[100];
    int age;
} Person;

void print_person(Person* p) {
    // Using stdio.h
    printf("Name: %s, Age: %d\n", p->name, p->age);
}

int main() {
    // Using stdlib.h for malloc
    Person* person = (Person*)malloc(sizeof(Person));

    // Using string.h for strcpy
    strcpy(person->name, "John Doe");
    person->age = 30;

    print_person(person);

    // Using math.h
    double value = 16.0;
    double result = sqrt(value);
    printf("Square root of %.1f is %.2f\n", value, result);

    // Using stdlib.h for free
    free(person);

    return 0;
}
