/**
 * Basic JavaScript validation utilities
 * Demonstrates: module exports, functions, constants, JSDoc
 */

// Public constants
const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const MIN_PASSWORD_LENGTH = 8;

// Private helper constant
const _VALIDATION_CACHE = new Map();

/**
 * Validates an email address
 * @param {string} email - Email to validate
 * @returns {boolean} True if valid
 */
function isValidEmail(email) {
    if (!email || typeof email !== 'string') {
        return false;
    }

    // Check cache first
    if (_VALIDATION_CACHE.has(email)) {
        return _VALIDATION_CACHE.get(email);
    }

    const isValid = EMAIL_REGEX.test(email.toLowerCase().trim());
    _VALIDATION_CACHE.set(email, isValid);
    return isValid;
}

/**
 * Validates password strength
 * @param {string} password - Password to validate
 * @returns {boolean} True if strong enough
 */
function isValidPassword(password) {
    if (!password || typeof password !== 'string') {
        return false;
    }

    return password.length >= MIN_PASSWORD_LENGTH &&
           /[A-Z]/.test(password) &&
           /[a-z]/.test(password) &&
           /\d/.test(password);
}

/**
 * Private helper to format user ID
 * @private
 * @param {number} id - User ID
 * @returns {string} Formatted ID
 */
function _formatUserID(id) {
    return `USER_${String(id).padStart(6, '0')}`;
}

/**
 * Formats user display name
 * @param {number} id - User ID
 * @param {string} name - User name
 * @returns {string} Formatted display name
 */
function formatUserDisplay(id, name) {
    const formattedId = _formatUserID(id);
    return `${formattedId}: ${name}`;
}

// Export public API
module.exports = {
    EMAIL_REGEX,
    MIN_PASSWORD_LENGTH,
    isValidEmail,
    isValidPassword,
    formatUserDisplay
};