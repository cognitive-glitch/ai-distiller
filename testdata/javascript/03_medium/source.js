/**
 * Medium complexity JavaScript notification service
 * Demonstrates: inheritance, mixins, async/await, event patterns, composition
 */

const EventEmitter = require('events');

/**
 * Base notification interface (mixin pattern)
 */
const NotificationMixin = {
    /**
     * Send notification method to be implemented
     * @abstract
     * @param {Object} payload - Notification data
     * @returns {Promise<boolean>} Success status
     */
    async send(payload) {
        throw new Error('send() must be implemented');
    },

    /**
     * Validates notification payload
     * @protected
     * @param {Object} payload - Payload to validate
     * @returns {boolean} Valid or not
     */
    _validatePayload(payload) {
        return payload && payload.message && payload.recipient;
    }
};

/**
 * Abstract base notification service
 */
class BaseNotificationService extends EventEmitter {
    /**
     * @param {string} serviceName - Name of the service
     */
    constructor(serviceName) {
        super();
        this.serviceName = serviceName;
        this._stats = {
            sent: 0,
            failed: 0,
            lastSent: null
        };
        this._isActive = true;
    }

    /**
     * Gets service statistics
     * @returns {Object} Service stats
     */
    get stats() {
        return { ...this._stats };
    }

    /**
     * Checks if service is active
     * @returns {boolean} Active status
     */
    get isActive() {
        return this._isActive;
    }

    /**
     * Abstract send method
     * @abstract
     * @param {Object} payload - Notification payload
     * @returns {Promise<boolean>} Success status
     */
    async send(payload) {
        throw new Error('send() must be implemented by subclass');
    }

    /**
     * Processes notification with error handling and stats
     * @param {Object} payload - Notification data
     * @returns {Promise<boolean>} Success status
     */
    async processNotification(payload) {
        if (!this._isActive) {
            throw new Error('Service is not active');
        }

        try {
            this.emit('beforeSend', payload);

            if (!this._validatePayload(payload)) {
                throw new Error('Invalid payload');
            }

            const result = await this.send(payload);

            if (result) {
                this._stats.sent++;
                this._stats.lastSent = new Date();
                this.emit('sent', payload);
            } else {
                this._stats.failed++;
                this.emit('failed', payload);
            }

            return result;
        } catch (error) {
            this._stats.failed++;
            this.emit('error', error, payload);
            throw error;
        }
    }

    /**
     * Protected validation method
     * @protected
     * @param {Object} payload - Payload to validate
     * @returns {boolean} Valid or not
     */
    _validatePayload(payload) {
        return NotificationMixin._validatePayload(payload);
    }

    /**
     * Deactivates the service
     */
    deactivate() {
        this._isActive = false;
        this.emit('deactivated');
    }

    /**
     * Reactivates the service
     */
    activate() {
        this._isActive = true;
        this.emit('activated');
    }
}

/**
 * Email notification service implementation
 */
class EmailNotificationService extends BaseNotificationService {
    /**
     * @param {Object} config - Email configuration
     */
    constructor(config = {}) {
        super('email');
        this._config = {
            smtp: config.smtp || 'localhost',
            port: config.port || 587,
            timeout: config.timeout || 5000,
            ...config
        };
        this._connectionPool = new Map();
    }

    /**
     * Sends email notification
     * @param {Object} payload - Email data
     * @returns {Promise<boolean>} Success status
     */
    async send(payload) {
        const { recipient, message, subject = 'Notification' } = payload;

        // Simulate async email sending
        const connection = await this._getConnection();

        try {
            await this._sendEmail(connection, {
                to: recipient,
                subject,
                body: message
            });

            this._logEmailSent(recipient);
            return true;
        } catch (error) {
            this._logEmailError(recipient, error);
            return false;
        } finally {
            this._releaseConnection(connection);
        }
    }

    /**
     * Gets email connection from pool
     * @private
     * @returns {Promise<Object>} Connection object
     */
    async _getConnection() {
        // Simulate connection pooling
        const connectionId = Math.random().toString(36);
        const connection = {
            id: connectionId,
            smtp: this._config.smtp,
            connected: true
        };

        this._connectionPool.set(connectionId, connection);
        return connection;
    }

    /**
     * Sends actual email
     * @private
     * @param {Object} connection - SMTP connection
     * @param {Object} emailData - Email to send
     * @returns {Promise<void>}
     */
    async _sendEmail(connection, emailData) {
        // Simulate network delay
        await new Promise(resolve => setTimeout(resolve, 100));

        if (Math.random() < 0.1) { // 10% failure rate
            throw new Error('SMTP connection failed');
        }

        // Log successful send
        console.log(`Email sent to ${emailData.to} via ${connection.smtp}`);
    }

    /**
     * Releases connection back to pool
     * @private
     * @param {Object} connection - Connection to release
     */
    _releaseConnection(connection) {
        this._connectionPool.delete(connection.id);
    }

    /**
     * Logs successful email
     * @private
     * @param {string} recipient - Email recipient
     */
    _logEmailSent(recipient) {
        console.log(`✓ Email delivered to ${recipient}`);
    }

    /**
     * Logs email error
     * @private
     * @param {string} recipient - Email recipient
     * @param {Error} error - Error that occurred
     */
    _logEmailError(recipient, error) {
        console.error(`✗ Email failed to ${recipient}: ${error.message}`);
    }

    /**
     * Gets current connection count
     * @returns {number} Active connections
     */
    getActiveConnections() {
        return this._connectionPool.size;
    }
}

/**
 * SMS notification service implementation
 */
class SMSNotificationService extends BaseNotificationService {
    /**
     * @param {string} provider - SMS provider name
     */
    constructor(provider = 'default') {
        super('sms');
        this._provider = provider;
        this._rateLimitQueue = [];
    }

    /**
     * Sends SMS notification
     * @param {Object} payload - SMS data
     * @returns {Promise<boolean>} Success status
     */
    async send(payload) {
        const { recipient, message } = payload;

        // Check rate limiting
        if (!this._checkRateLimit()) {
            throw new Error('Rate limit exceeded');
        }

        // Simulate SMS API call
        try {
            await this._sendSMS(recipient, message);
            return true;
        } catch (error) {
            console.error(`SMS failed: ${error.message}`);
            return false;
        }
    }

    /**
     * Checks SMS rate limiting
     * @private
     * @returns {boolean} Can send SMS
     */
    _checkRateLimit() {
        const now = Date.now();
        const minute = 60 * 1000;

        // Remove old entries
        this._rateLimitQueue = this._rateLimitQueue.filter(time => now - time < minute);

        // Check limit (10 per minute)
        if (this._rateLimitQueue.length >= 10) {
            return false;
        }

        this._rateLimitQueue.push(now);
        return true;
    }

    /**
     * Sends actual SMS
     * @private
     * @param {string} recipient - Phone number
     * @param {string} message - SMS text
     * @returns {Promise<void>}
     */
    async _sendSMS(recipient, message) {
        // Simulate API delay
        await new Promise(resolve => setTimeout(resolve, 200));

        if (message.length > 160) {
            throw new Error('Message too long for SMS');
        }

        console.log(`SMS sent to ${recipient}: ${message}`);
    }
}

// Apply mixin to classes
Object.assign(BaseNotificationService.prototype, NotificationMixin);

module.exports = {
    BaseNotificationService,
    EmailNotificationService,
    SMSNotificationService,
    NotificationMixin
};