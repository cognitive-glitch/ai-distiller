/**
 * Complex JavaScript test file for AI Distiller functional testing
 * Includes classes, inheritance, async/await, generators, modules, JSDoc types
 */

// ES6 imports
import { EventEmitter } from 'events';
import { promisify } from 'util';
import * as fs from 'fs';

// JSDoc types
/**
 * @typedef {Object} User
 * @property {number} id - User ID
 * @property {string} name - User name
 * @property {string} [email] - User email (optional)
 * @property {string[]} roles - User roles
 */

/**
 * @typedef {Object} ProcessOptions
 * @property {boolean} [validate=true] - Whether to validate
 * @property {boolean} [transform=false] - Whether to transform
 */

// Constants
const STATUS = {
    ACTIVE: 'active',
    INACTIVE: 'inactive',
    PENDING: 'pending'
};

const PRIORITY_LEVELS = Object.freeze({
    LOW: 1,
    MEDIUM: 2,
    HIGH: 3
});

// Base class with inheritance
class BaseService extends EventEmitter {
    /**
     * @param {string} name - Service name
     * @param {Object} logger - Logger instance
     */
    constructor(name, logger) {
        super();
        this.name = name;
        this._logger = logger;
        this._cache = new Map();
        this.#privateField = 'private data';
    }
    
    // Private field (ES2022)
    #privateField;
    
    /**
     * Abstract method to be implemented by subclasses
     * @abstract
     * @param {*} entity - Entity to validate
     * @returns {boolean} Validation result
     */
    validate(entity) {
        throw new Error('validate() must be implemented by subclass');
    }
    
    /**
     * Process entity with validation
     * @param {*} entity - Entity to process
     * @returns {Promise<*>} Processed entity
     */
    async process(entity) {
        if (!this.validate(entity)) {
            throw new Error('Validation failed');
        }
        return this._doProcess(entity);
    }
    
    /**
     * @private
     * @param {*} entity - Entity to process
     * @returns {Promise<*>} Processed entity
     */
    async _doProcess(entity) {
        this._logger.info(`Processing entity in ${this.name}`);
        await this._delay(100); // Simulate async work
        return entity;
    }
    
    /**
     * @private
     * @param {number} ms - Delay in milliseconds
     * @returns {Promise<void>}
     */
    _delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
    
    /**
     * @private
     * @returns {string} Private field value
     */
    #getPrivateField() {
        return this.#privateField;
    }
    
    /**
     * Static factory method
     * @static
     * @param {string} name - Service name
     * @returns {BaseService} Service instance
     */
    static create(name) {
        const logger = console; // Simple logger
        return new this(name, logger);
    }
}

// Concrete service implementation
class UserService extends BaseService {
    /**
     * @param {Object} logger - Logger instance
     * @param {EmailService} emailService - Email service
     */
    constructor(logger, emailService) {
        super('UserService', logger);
        this.emailService = emailService;
        this.users = new Map();
    }
    
    /**
     * @override
     * @param {User} user - User to validate
     * @returns {boolean} Validation result
     */
    validate(user) {
        return user && 
               typeof user.name === 'string' && 
               user.name.trim().length > 0 &&
               typeof user.id === 'number' && 
               user.id > 0;
    }
    
    /**
     * Save user to repository
     * @param {User} user - User to save
     * @returns {Promise<User>} Saved user
     */
    async save(user) {
        const processedUser = await this.process(user);
        this.users.set(user.id, processedUser);
        
        if (user.email) {
            await this.emailService.sendWelcome(user.email);
        }
        
        this.emit('userSaved', processedUser);
        return processedUser;
    }
    
    /**
     * Find user by ID
     * @param {number} id - User ID
     * @returns {User|null} Found user or null
     */
    findById(id) {
        return this.users.get(id) || null;
    }
    
    /**
     * Delete user by ID
     * @param {number} id - User ID
     * @returns {boolean} True if deleted, false if not found
     */
    delete(id) {
        const deleted = this.users.delete(id);
        if (deleted) {
            this.emit('userDeleted', id);
        }
        return deleted;
    }
    
    /**
     * Get all users with optional filter
     * @param {Function} [filter] - Optional filter function
     * @returns {User[]} Array of users
     */
    getAllUsers(filter) {
        const users = Array.from(this.users.values());
        return filter ? users.filter(filter) : users;
    }
    
    /**
     * Map users to different format
     * @template T
     * @param {function(User): T} mapper - Mapping function
     * @returns {T[]} Mapped results
     */
    mapUsers(mapper) {
        return this.getAllUsers().map(mapper);
    }
    
    /**
     * Process users in batches
     * @async
     * @generator
     * @param {User[]} users - Users to process
     * @param {number} batchSize - Batch size
     * @yields {Promise<User[]>} Processed batch
     */
    async* processBatches(users, batchSize = 10) {
        for (let i = 0; i < users.length; i += batchSize) {
            const batch = users.slice(i, i + batchSize);
            const processedBatch = await Promise.all(
                batch.map(user => this.process(user))
            );
            yield processedBatch;
        }
    }
}

// Email service class
class EmailService {
    constructor() {
        this.sentEmails = [];
    }
    
    /**
     * Send welcome email
     * @param {string} email - Email address
     * @returns {Promise<void>}
     */
    async sendWelcome(email) {
        await this._sendEmail(email, 'Welcome!', 'Welcome to our service');
    }
    
    /**
     * Send notification email
     * @param {string} email - Email address
     * @param {string} message - Message content
     * @returns {Promise<void>}
     */
    async sendNotification(email, message) {
        await this._sendEmail(email, 'Notification', message);
    }
    
    /**
     * @private
     * @param {string} to - Recipient email
     * @param {string} subject - Email subject
     * @param {string} body - Email body
     * @returns {Promise<void>}
     */
    async _sendEmail(to, subject, body) {
        // Simulate email sending
        await new Promise(resolve => setTimeout(resolve, 50));
        this.sentEmails.push({ to, subject, body, timestamp: Date.now() });
        console.log(`Email sent to ${to}: ${subject}`);
    }
}

// Factory function
/**
 * Create user service with dependencies
 * @param {Object} [options={}] - Configuration options
 * @returns {UserService} Configured user service
 */
function createUserService(options = {}) {
    const logger = options.logger || console;
    const emailService = options.emailService || new EmailService();
    return new UserService(logger, emailService);
}

// Higher-order function
/**
 * Create a retry wrapper for async functions
 * @param {number} maxAttempts - Maximum retry attempts
 * @returns {Function} Retry decorator
 */
function withRetry(maxAttempts = 3) {
    return function(asyncFn) {
        return async function(...args) {
            let lastError;
            for (let attempt = 1; attempt <= maxAttempts; attempt++) {
                try {
                    return await asyncFn.apply(this, args);
                } catch (error) {
                    lastError = error;
                    if (attempt === maxAttempts) {
                        throw error;
                    }
                    await new Promise(resolve => 
                        setTimeout(resolve, Math.pow(2, attempt) * 100)
                    );
                }
            }
            throw lastError;
        };
    };
}

// Async/await utilities
/**
 * Process data with various options
 * @async
 * @param {User[]} users - Users to process
 * @param {ProcessOptions} [options={}] - Processing options
 * @returns {Promise<User[]>} Processed users
 */
async function processUserData(users, options = {}) {
    const { validate = true, transform = false } = options;
    
    return users
        .filter(user => !validate || isValidUser(user))
        .map(user => transform ? transformUser(user) : user);
}

// Arrow functions
/**
 * @param {User} user - User to validate
 * @returns {boolean} Validation result
 */
const isValidUser = (user) => {
    return user && user.name && user.name.trim().length > 0 && user.id > 0;
};

/**
 * @param {User} user - User to transform
 * @returns {User} Transformed user
 */
const transformUser = (user) => ({
    ...user,
    displayName: `${user.name} (${user.email || 'no email'})`,
    timestamp: Date.now()
});

// Generator function
/**
 * Generate fibonacci sequence
 * @generator
 * @param {number} n - Number of fibonacci numbers to generate
 * @yields {number} Next fibonacci number
 */
function* fibonacci(n) {
    let a = 0, b = 1;
    for (let i = 0; i < n; i++) {
        yield a;
        [a, b] = [b, a + b];
    }
}

// Async generator
/**
 * Generate users asynchronously
 * @async
 * @generator
 * @param {number} count - Number of users to generate
 * @yields {Promise<User>} Next user
 */
async function* generateUsers(count) {
    for (let i = 1; i <= count; i++) {
        await new Promise(resolve => setTimeout(resolve, 10));
        yield {
            id: i,
            name: `User ${i}`,
            email: `user${i}@example.com`,
            roles: ['user']
        };
    }
}

// Module pattern
const UserModule = (function() {
    // Private variables
    let instances = new WeakMap();
    
    // Private functions
    function validateConfig(config) {
        return config && typeof config === 'object';
    }
    
    // Public API
    return {
        /**
         * Create user service instance
         * @param {Object} config - Configuration
         * @returns {UserService} Service instance
         */
        create(config) {
            if (!validateConfig(config)) {
                throw new Error('Invalid configuration');
            }
            const service = createUserService(config);
            instances.set(service, config);
            return service;
        },
        
        /**
         * Get configuration for service instance
         * @param {UserService} service - Service instance
         * @returns {Object|undefined} Configuration
         */
        getConfig(service) {
            return instances.get(service);
        }
    };
})();

// Class with static members
class Utils {
    /**
     * Format user name
     * @static
     * @param {string} first - First name
     * @param {string} last - Last name
     * @returns {string} Formatted name
     */
    static formatName(first, last) {
        return `${first} ${last}`;
    }
    
    /**
     * Capitalize string
     * @static
     * @param {string} str - String to capitalize
     * @returns {string} Capitalized string
     */
    static capitalize(str) {
        return str.charAt(0).toUpperCase() + str.slice(1);
    }
    
    /**
     * Deep clone object
     * @static
     * @param {*} obj - Object to clone
     * @returns {*} Cloned object
     */
    static deepClone(obj) {
        return JSON.parse(JSON.stringify(obj));
    }
}

// Prototype extension
/**
 * Add method to Array prototype
 */
Array.prototype.chunk = function(size) {
    const chunks = [];
    for (let i = 0; i < this.length; i += size) {
        chunks.push(this.slice(i, i + size));
    }
    return chunks;
};

// Export for module systems
export {
    BaseService,
    UserService,
    EmailService,
    createUserService,
    withRetry,
    processUserData,
    isValidUser,
    transformUser,
    fibonacci,
    generateUsers,
    UserModule,
    Utils,
    STATUS,
    PRIORITY_LEVELS
};

// CommonJS fallback
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        BaseService,
        UserService,
        EmailService,
        createUserService,
        withRetry,
        processUserData,
        isValidUser,
        transformUser,
        fibonacci,
        generateUsers,
        UserModule,
        Utils,
        STATUS,
        PRIORITY_LEVELS
    };
}