// Basic JavaScript structures test

// Imports
import React from 'react';
import { Component, useState } from 'react';
import * as utils from './utils';

// Variable declarations
const PI = 3.14159;
let counter = 0;
var oldStyle = 'legacy';

// Function declarations
function add(a, b) {
    return a + b;
}

// Arrow functions
const multiply = (x, y) => x * y;

const divide = async (numerator, denominator) => {
    if (denominator === 0) {
        throw new Error('Division by zero');
    }
    return numerator / denominator;
};

// Generator function
function* fibonacci() {
    let [a, b] = [0, 1];
    while (true) {
        yield a;
        [a, b] = [b, a + b];
    }
}

// Class declaration
class Animal {
    constructor(name) {
        this.name = name;
        this.alive = true;
    }

    speak() {
        console.log(`${this.name} makes a sound`);
    }

    static createDog(name) {
        return new Dog(name);
    }
}

// Class inheritance
class Dog extends Animal {
    #privateField = 'secret';
    
    constructor(name, breed = 'mixed') {
        super(name);
        this.breed = breed;
    }

    speak() {
        console.log(`${this.name} barks`);
    }

    get description() {
        return `${this.name} is a ${this.breed}`;
    }

    set age(value) {
        if (value < 0) throw new Error('Age cannot be negative');
        this._age = value;
    }
}

// Export statements
export default Animal;
export { Dog, fibonacci };
export { add as sum };