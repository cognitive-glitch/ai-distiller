/**
 * @file Construct 5: The Metaprogramming Minefield
 * Tests Proxy, Symbols, Reflect API, and advanced meta-programming patterns
 * Note: Decorators excluded due to lack of native runtime support
 */

// Symbol for private data
const secretData = Symbol('secretData');
const metaInfo = Symbol('metaInfo');

/**
 * Proxy handler for dynamic property access
 * @type {ProxyHandler}
 */
const dynamicApiHandler = {
    get(target, prop, receiver) {
        console.log(`[PROXY] Accessing property: ${String(prop)}`);
        
        // Handle existing properties
        if (prop in target) {
            return Reflect.get(...arguments);
        }
        
        // Handle Symbol properties
        if (typeof prop === 'symbol') {
            return target[prop];
        }
        
        // Dynamically create getter methods
        if (typeof prop === 'string' && prop.startsWith('get')) {
            const key = prop.substring(3).toLowerCase();
            return () => {
                const data = target[secretData];
                return data && data[key] 
                    ? `${key}: ${data[key]}` 
                    : `No data for key: ${key}`;
            };
        }
        
        // Handle 'set' methods dynamically
        if (typeof prop === 'string' && prop.startsWith('set')) {
            const key = prop.substring(3).toLowerCase();
            return (value) => {
                if (!target[secretData]) {
                    target[secretData] = {};
                }
                target[secretData][key] = value;
                return value;
            };
        }
        
        return `Property '${String(prop)}' does not exist.`;
    },
    
    set(target, prop, value, receiver) {
        console.log(`[PROXY] Setting '${String(prop)}' to '${value}'`);
        
        // Track metadata about property changes
        if (!target[metaInfo]) {
            target[metaInfo] = { changes: [] };
        }
        target[metaInfo].changes.push({
            property: String(prop),
            value,
            timestamp: Date.now()
        });
        
        return Reflect.set(...arguments);
    },
    
    has(target, prop) {
        console.log(`[PROXY] Checking existence of '${String(prop)}'`);
        return prop in target || 
               (typeof prop === 'string' && prop.startsWith('get'));
    },
    
    deleteProperty(target, prop) {
        console.log(`[PROXY] Deleting property '${String(prop)}'`);
        return Reflect.deleteProperty(target, prop);
    },
    
    ownKeys(target) {
        // Include Symbol keys
        return Reflect.ownKeys(target);
    }
};

/**
 * Factory function for creating dynamic objects
 * @param {Object} initialData - Initial data for the object
 * @returns {Proxy} A proxied object with dynamic behavior
 */
function createDynamicObject(initialData) {
    const obj = {
        [secretData]: initialData,
        [metaInfo]: {
            created: Date.now(),
            changes: []
        }
    };
    return new Proxy(obj, dynamicApiHandler);
}

/**
 * Advanced class using Symbols and WeakMap for true privacy
 */
class SecureStorage {
    static #instances = new WeakMap();
    static #accessCount = new WeakMap();
    
    constructor(password) {
        // Use WeakMap for true private data
        SecureStorage.#instances.set(this, {
            password,
            data: new Map()
        });
        SecureStorage.#accessCount.set(this, 0);
    }
    
    store(key, value, password) {
        const instance = SecureStorage.#instances.get(this);
        if (instance.password !== password) {
            throw new Error('Invalid password');
        }
        instance.data.set(key, value);
        this.#incrementAccess();
    }
    
    retrieve(key, password) {
        const instance = SecureStorage.#instances.get(this);
        if (instance.password !== password) {
            throw new Error('Invalid password');
        }
        this.#incrementAccess();
        return instance.data.get(key);
    }
    
    #incrementAccess() {
        const count = SecureStorage.#accessCount.get(this) || 0;
        SecureStorage.#accessCount.set(this, count + 1);
    }
    
    getAccessCount() {
        return SecureStorage.#accessCount.get(this) || 0;
    }
}

/**
 * Tagged template literal for SQL-like queries
 * @param {string[]} strings - Template strings
 * @param {...any} values - Interpolated values
 * @returns {Object} Query object
 */
function sql(strings, ...values) {
    const query = strings.reduce((result, str, i) => {
        return result + str + (values[i] !== undefined ? `$${i + 1}` : '');
    }, '');
    
    return {
        text: query.trim(),
        values: values,
        execute() {
            console.log(`Executing: ${this.text}`);
            console.log(`With values:`, this.values);
            return Promise.resolve({ rows: [], rowCount: 0 });
        }
    };
}

/**
 * Reflect API usage for safe property manipulation
 */
const safePropertyAccess = {
    getSafe(obj, prop, defaultValue = null) {
        try {
            return Reflect.has(obj, prop) ? Reflect.get(obj, prop) : defaultValue;
        } catch (e) {
            return defaultValue;
        }
    },
    
    setSafe(obj, prop, value) {
        try {
            return Reflect.set(obj, prop, value);
        } catch (e) {
            console.error(`Failed to set property '${prop}':`, e);
            return false;
        }
    },
    
    definePropertySafe(obj, prop, descriptor) {
        try {
            return Reflect.defineProperty(obj, prop, descriptor);
        } catch (e) {
            console.error(`Failed to define property '${prop}':`, e);
            return false;
        }
    }
};

// Usage examples
const user = createDynamicObject({ name: 'Alex', role: 'admin' });

// Dynamic getters
console.log(user.getName()); // "name: Alex"
console.log(user.getRole()); // "role: admin"
console.log(user.getEmail()); // "No data for key: email"

// Dynamic setters
user.setEmail('alex@example.com');
console.log(user.getEmail()); // "email: alex@example.com"

// Regular property access
user.status = 'active';

// Tagged template usage
const query = sql`SELECT * FROM users WHERE name = ${'Alex'} AND role = ${'admin'}`;
query.execute();

// Export all constructs
export {
    createDynamicObject,
    SecureStorage,
    sql,
    safePropertyAccess,
    secretData,
    metaInfo
};