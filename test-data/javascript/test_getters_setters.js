// Test getters and setters
const user = {
    _name: 'John',
    _age: 30,
    
    get name() {
        return this._name;
    },
    
    set name(value) {
        this._name = value;
    },
    
    get age() {
        return this._age;
    },
    
    set age(value) {
        if (value > 0) {
            this._age = value;
        }
    },
    
    // Regular method
    greet(greeting = 'Hello') {
        return `${greeting}, ${this._name}!`;
    },
    
    // Method with multiple parameters
    updateInfo(name, age, email) {
        this._name = name;
        this._age = age;
        this._email = email;
    }
};

// Class with getters/setters
class Rectangle {
    constructor(width, height) {
        this.width = width;
        this.height = height;
    }
    
    get area() {
        return this.width * this.height;
    }
    
    get perimeter() {
        return 2 * (this.width + this.height);
    }
    
    set dimensions({width, height}) {
        this.width = width;
        this.height = height;
    }
}