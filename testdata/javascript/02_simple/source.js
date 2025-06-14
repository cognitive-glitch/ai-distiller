/**
 * Simple JavaScript user model with classes
 * Demonstrates: ES6 classes, private fields, getters/setters, static methods
 */

/**
 * User model class with validation and formatting
 */
class UserModel {
    // Private fields (ES2022)
    #id;
    #email;
    #createdAt;
    
    /**
     * Creates a new user
     * @param {number} id - User ID
     * @param {string} email - User email
     * @param {string} name - User name
     */
    constructor(id, email, name) {
        this.#id = id;
        this.#email = email;
        this.name = name;
        this.#createdAt = new Date();
        this._validateEmail();
    }
    
    /**
     * Gets user ID (readonly)
     * @returns {number} User ID
     */
    get id() {
        return this.#id;
    }
    
    /**
     * Gets user email
     * @returns {string} User email
     */
    get email() {
        return this.#email;
    }
    
    /**
     * Sets user email with validation
     * @param {string} email - New email
     */
    set email(email) {
        this.#email = email;
        this._validateEmail();
    }
    
    /**
     * Gets creation timestamp
     * @returns {Date} Creation date
     */
    get createdAt() {
        return new Date(this.#createdAt);
    }
    
    /**
     * Private email validation
     * @private
     */
    _validateEmail() {
        if (!this.#email || !this.#email.includes('@')) {
            throw new Error('Invalid email format');
        }
    }
    
    /**
     * Gets user display string
     * @returns {string} Formatted user info
     */
    toString() {
        const age = this._getAccountAge();
        return `User #${this.#id}: ${this.name} (${this.#email}) - ${age} days old`;
    }
    
    /**
     * Private helper for account age calculation
     * @private
     * @returns {number} Account age in days
     */
    _getAccountAge() {
        const now = new Date();
        const diffTime = Math.abs(now - this.#createdAt);
        return Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    }
    
    /**
     * Converts user to JSON object
     * @returns {Object} User data
     */
    toJSON() {
        return {
            id: this.#id,
            email: this.#email,
            name: this.name,
            createdAt: this.#createdAt.toISOString()
        };
    }
    
    /**
     * Static factory method for creating users
     * @static
     * @param {Object} data - User data
     * @returns {UserModel} New user instance
     */
    static fromJSON(data) {
        const user = new UserModel(data.id, data.email, data.name);
        if (data.createdAt) {
            user.#createdAt = new Date(data.createdAt);
        }
        return user;
    }
    
    /**
     * Static method to validate user data
     * @static
     * @param {Object} data - Data to validate
     * @returns {boolean} True if valid
     */
    static isValidUserData(data) {
        return data && 
               typeof data.id === 'number' && 
               typeof data.email === 'string' && 
               typeof data.name === 'string' &&
               data.email.includes('@');
    }
}

module.exports = UserModel;