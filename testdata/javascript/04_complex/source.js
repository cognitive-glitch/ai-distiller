/**
 * @file 04_complex.js
 * @description Demonstrates complex JavaScript features including Proxies for metaprogramming,
 * the Observer design pattern, WeakMaps for memory-safe metadata, Symbols for unique properties,
 * and Generators for iterable history.
 * This module provides a factory for creating "observable" objects.
 */

/**
 * A private symbol used to access the original, raw object from its proxy.
 * Using a Symbol prevents accidental property name collisions.
 * @type {symbol}
 */
const RAW_OBJECT_SYMBOL = Symbol('rawObject');

/**
 * A WeakMap to store event listeners for each observable object.
 * Using a WeakMap is crucial for memory management. If an observed object is garbage collected,
 * its corresponding entry in the WeakMap (and thus its listeners) will also be removed automatically,
* preventing memory leaks that would occur with a standard Map.
 * @type {WeakMap<object, Set<Function>>}
 */
const objectListeners = new WeakMap();

/**
 * A log of all mutations for demonstration of generators.
 * @type {Array<{type: string, key: string, value?: any, timestamp: number}>}
 */
const mutationLog = [];

/**
 * The handler for the Proxy object. It contains the "traps" that intercept operations
 * on the target object. This is the core of the metaprogramming aspect.
 * @type {ProxyHandler<object>}
 */
const observableProxyHandler = {
    /**
     * Trap for getting a property value.
     * @param {object} target - The original object.
     * @param {string|symbol} key - The name of the property to get.
     * @param {object} receiver - The proxy or an object that inherits from it.
     * @returns {any}
     */
    get(target, key, receiver) {
        if (key === RAW_OBJECT_SYMBOL) {
            return target;
        }
        console.log(`[GET] key: ${String(key)}`);
        return Reflect.get(target, key, receiver);
    },

    /**
     * Trap for setting a property value. This is where the observation magic happens.
     * @param {object} target - The original object.
     * @param {string|symbol} key - The name of the property to set.
     * @param {any} value - The new value.
     * @param {object} receiver - The proxy or an object that inherits from it.
     * @returns {boolean} - True if the set was successful.
     */
    set(target, key, value, receiver) {
        const oldValue = target[key];
        if (oldValue === value) {
            return true; // No change, do nothing.
        }

        const result = Reflect.set(target, key, value, receiver);
        if (result) {
            console.log(`[SET] key: ${String(key)}, value:`, value);
            // Log the mutation
            mutationLog.push({ type: 'SET', key: String(key), value, timestamp: Date.now() });
            // Notify listeners associated with this object
            const listeners = objectListeners.get(target);
            if (listeners) {
                listeners.forEach(listener => listener({ key, value, oldValue }));
            }
        }
        return result;
    }
};

/**
 * @class ObservableFactory
 * @description A factory for creating observable configuration objects.
 * Encapsulates the logic of creating proxies and managing subscriptions.
 */
class ObservableFactory {
    /**
     * Creates a new observable object backed by a Proxy.
     * @param {object} initialData - The initial plain JavaScript object.
     * @returns {object} A new Proxy-wrapped observable object.
     */
    static create(initialData) {
        if (typeof initialData !== 'object' || initialData === null) {
            throw new Error('Initial data must be an object.');
        }
        // Associate a new Set of listeners for this object in our WeakMap.
        objectListeners.set(initialData, new Set());
        return new Proxy(initialData, observableProxyHandler);
    }

    /**
     * Subscribes a callback function to changes on an observable object.
     * @param {object} observable - The proxy object created by this factory.
     * @param {Function} callback - The function to call on change.
     */
    static subscribe(observable, callback) {
        const target = observable[RAW_OBJECT_SYMBOL]; // Get the original object via Symbol
        if (!target || !objectListeners.has(target)) {
            throw new Error('Can only subscribe to observables created by this factory.');
        }
        objectListeners.get(target).add(callback);
    }

    /**
     * Unsubscribes a callback function.
     * @param {object} observable - The proxy object.
     * @param {Function} callback - The function to remove.
     */
    static unsubscribe(observable, callback) {
        const target = observable[RAW_OBJECT_SYMBOL];
        if (target && objectListeners.has(target)) {
            objectListeners.get(target).delete(callback);
        }
    }

    /**
     * A generator function that yields the history of all mutations across all observables.
     * This demonstrates a practical use for generators: creating an iterable sequence
     * without loading all data into memory at once.
     * @returns {Generator<string, void, void>} An iterator for the mutation history.
     */
    static *getMutationHistory() {
        for (const log of mutationLog) {
            yield `[${new Date(log.timestamp).toISOString()}] ${log.type}: ${log.key} => ${JSON.stringify(log.value)}`;
        }
    }

    /**
     * Private helper to clear mutation logs
     * @private
     */
    static _clearHistory() {
        mutationLog.length = 0;
    }

    /**
     * Gets current listener count for debugging
     * @private
     * @param {object} observable - The observable object
     * @returns {number} Number of listeners
     */
    static _getListenerCount(observable) {
        const target = observable[RAW_OBJECT_SYMBOL];
        return target && objectListeners.has(target) ? objectListeners.get(target).size : 0;
    }
}

/**
 * Advanced configuration manager with deep observation capabilities
 * @class ConfigManager
 */
class ConfigManager {
    /**
     * @param {object} config - Initial configuration
     */
    constructor(config = {}) {
        this._observable = ObservableFactory.create(config);
        this._changeHistory = [];
        
        // Private change listener
        this._internalListener = (change) => {
            this._changeHistory.push({
                ...change,
                timestamp: Date.now()
            });
        };
        
        ObservableFactory.subscribe(this._observable, this._internalListener);
    }

    /**
     * Gets configuration value
     * @param {string} key - Configuration key
     * @returns {any} Configuration value
     */
    get(key) {
        return this._observable[key];
    }

    /**
     * Sets configuration value
     * @param {string} key - Configuration key
     * @param {any} value - New value
     */
    set(key, value) {
        this._observable[key] = value;
    }

    /**
     * Gets change history for this config
     * @returns {Array} Change history
     */
    getChangeHistory() {
        return [...this._changeHistory];
    }

    /**
     * Private method to validate configuration
     * @private
     * @param {string} key - Key to validate
     * @param {any} value - Value to validate
     * @returns {boolean} Is valid
     */
    _validateChange(key, value) {
        // Simple validation logic
        return key && value !== undefined;
    }
}

// Usage Example
console.log('--- Creating Observable Config ---');
const appConfig = ObservableFactory.create({
    apiEndpoint: 'https://api.example.com/v1',
    timeout: 5000,
    features: {
        newUI: false,
        betaAccess: true
    }
});

const loggerCallback = (change) => {
    console.log(`Logger Callback Notified: Key '${String(change.key)}' changed to`, change.value);
};

console.log('\n--- Subscribing to Changes ---');
ObservableFactory.subscribe(appConfig, loggerCallback);

console.log('\n--- Modifying Config ---');
appConfig.timeout = 7500; // Triggers proxy 'set' trap and notifies subscriber
appConfig.features.newUI = true; // Note: This is a shallow observation. Deep observation is more complex.
appConfig.user = 'admin'; // Adding a new property

console.log('\n--- Unsubscribing ---');
ObservableFactory.unsubscribe(appConfig, loggerCallback);
appConfig.timeout = 10000; // This change will NOT notify the loggerCallback.

console.log('\n--- Iterating Mutation History with Generator ---');
for (const historyEntry of ObservableFactory.getMutationHistory()) {
    console.log(historyEntry);
}

module.exports = { ObservableFactory, ConfigManager, appConfig };