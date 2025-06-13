/**
 * @file CommonJS module for testing module interoperability
 * Part of Construct 3: The Module System Bridge
 */

// Private variable not exported
const privateKey = "secret";

/**
 * Legacy processor function
 * @param {any} data - Data to process
 * @returns {{processed: any, timestamp: number}} Processed result
 */
function legacyProcessor(data) {
    return { 
        processed: data, 
        timestamp: Date.now(),
        _internal: privateKey // Uses private variable
    };
}

/**
 * Helper function
 * @private
 */
function _internalHelper() {
    return "This is internal";
}

// CommonJS class pattern
function LegacyClass(name) {
    this.name = name;
    this._id = Math.random();
}

LegacyClass.prototype.getName = function() {
    return this.name;
};

LegacyClass.staticMethod = function() {
    return "I am static";
};

// Export using CommonJS
module.exports = {
    legacyProcessor,
    version: "1.0-cjs",
    LegacyClass,
    // Dynamic export
    getConfig: function() {
        return {
            mode: process.env.NODE_ENV || 'development',
            helper: _internalHelper()
        };
    }
};

// Also export a single function as the module
module.exports.default = legacyProcessor;