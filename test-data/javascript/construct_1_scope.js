/**
 * @file Construct 1: Scope & Closure Gauntlet
 * Tests identifier resolution, hoisting, and closure analysis
 */

// Test var hoisting
console.log(hoistedVar); // undefined, not ReferenceError
var hoistedVar = 'I am hoisted';

// Test function hoisting
hoistedFunction(); // Works due to hoisting

function hoistedFunction() {
    console.log('Function declarations are hoisted');
}

// Test block scoping
function scopeTest() {
    // Test 1: Classic var hoisting in a loop
    for (var i = 0; i < 3; i++) {
        setTimeout(() => console.log(`var i: ${i}`), 10); // Expects 3, 3, 3
    }

    // Test 2: let block-scoping in a loop
    for (let j = 0; j < 3; j++) {
        setTimeout(() => console.log(`let j: ${j}`), 10); // Expects 0, 1, 2
    }

    // Test 3: Closure over a changing variable
    let a = 1;
    const closure = {
        getA: () => a,
        setA: (val) => { a = val; }
    };
    
    console.log(`Initial a: ${closure.getA()}`); // Expects 1
    a = 5; // Mutate from outside
    console.log(`Mutated a: ${closure.getA()}`); // Expects 5
    
    return closure;
}

// Test IIFE (Immediately Invoked Function Expression)
const privateScope = (function() {
    var privateVar = 'I am private';
    
    return {
        getPrivate: function() {
            return privateVar;
        }
    };
})();

// Test arrow function vs regular function context
const contextTest = {
    name: 'outer',
    
    regularMethod: function() {
        return this.name;
    },
    
    arrowMethod: () => {
        return this.name; // 'this' is lexically bound
    },
    
    nestedTest: function() {
        const inner = () => {
            return this.name; // Still 'outer' due to arrow function
        };
        return inner();
    }
};

// Test default parameters
function defaultParams(a = 1, b = a * 2) {
    return a + b;
}

// Export for module testing
module.exports = {
    scopeTest,
    privateScope,
    contextTest,
    defaultParams
};