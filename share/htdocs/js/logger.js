/**
 * EntityDB Logger
 * Structured logging system with remote debugging support
 * Version: v2.29.0
 */

class Logger {
    constructor(component) {
        this.component = component;
        this.levels = {
            debug: 0,
            info: 1,
            warn: 2,
            error: 3
        };
        this.enabled = localStorage.getItem('entitydb-debug') === 'true';
        this.logLevel = localStorage.getItem('entitydb-log-level') || 'info';
        this.maxStoredErrors = 100;
        this.sessionId = this.generateSessionId();
    }

    /**
     * Generate unique session ID for tracking
     */
    generateSessionId() {
        return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    }

    /**
     * Check if log should be output based on level
     */
    shouldLog(level) {
        if (!this.enabled && level !== 'error') return false;
        return this.levels[level] >= this.levels[this.logLevel];
    }

    /**
     * Main logging method
     */
    log(level, message, data = {}, options = {}) {
        if (!this.shouldLog(level)) return;
        
        const timestamp = new Date().toISOString();
        const logEntry = {
            timestamp,
            level,
            component: this.component,
            message,
            data,
            sessionId: this.sessionId,
            url: window.location.href,
            userAgent: navigator.userAgent
        };

        // Add stack trace for errors
        if (level === 'error' && data instanceof Error) {
            logEntry.stack = data.stack;
            logEntry.errorName = data.name;
        }

        // Format console output
        const prefix = `[${timestamp}] [${this.component}] [${level.toUpperCase()}]`;
        const consoleMethod = level === 'error' ? 'error' : level === 'warn' ? 'warn' : 'log';
        
        if (options.group) {
            console.group(prefix + ' ' + message);
            console[consoleMethod](data);
            if (options.groupEnd !== false) {
                console.groupEnd();
            }
        } else {
            console[consoleMethod](prefix, message, data);
        }

        // Store errors for remote debugging
        if (level === 'error') {
            this.storeError(logEntry);
        }

        // Send critical errors to server if configured
        if (level === 'error' && this.shouldSendToServer()) {
            this.sendToServer(logEntry);
        }
    }

    /**
     * Convenience methods
     */
    debug(message, data = {}, options = {}) {
        this.log('debug', message, data, options);
    }

    info(message, data = {}, options = {}) {
        this.log('info', message, data, options);
    }

    warn(message, data = {}, options = {}) {
        this.log('warn', message, data, options);
    }

    error(message, data = {}, options = {}) {
        this.log('error', message, data, options);
    }

    /**
     * Group logging methods
     */
    group(message, data = {}) {
        this.log('info', message, data, { group: true, groupEnd: false });
    }

    groupEnd() {
        console.groupEnd();
    }

    /**
     * Performance logging
     */
    time(label) {
        if (!this.shouldLog('debug')) return;
        console.time(`[${this.component}] ${label}`);
    }

    timeEnd(label) {
        if (!this.shouldLog('debug')) return;
        console.timeEnd(`[${this.component}] ${label}`);
    }

    /**
     * Store error in localStorage for debugging
     */
    storeError(error) {
        try {
            const errors = JSON.parse(localStorage.getItem('entitydb-errors') || '[]');
            errors.push(error);
            
            // Keep only recent errors
            while (errors.length > this.maxStoredErrors) {
                errors.shift();
            }
            
            localStorage.setItem('entitydb-errors', JSON.stringify(errors));
        } catch (e) {
            // Ignore storage errors
            console.error('Failed to store error:', e);
        }
    }

    /**
     * Check if errors should be sent to server
     */
    shouldSendToServer() {
        return localStorage.getItem('entitydb-send-errors') === 'true';
    }

    /**
     * Send error to server for remote debugging
     */
    async sendToServer(logEntry) {
        try {
            const api = new window.EntityDBClient();
            await api.request('/api/v1/logs/error', {
                method: 'POST',
                body: JSON.stringify(logEntry)
            });
        } catch (e) {
            // Don't log errors about logging errors
            console.error('Failed to send error to server:', e);
        }
    }

    /**
     * Get stored errors for debugging
     */
    static getStoredErrors() {
        try {
            return JSON.parse(localStorage.getItem('entitydb-errors') || '[]');
        } catch (e) {
            return [];
        }
    }

    /**
     * Clear stored errors
     */
    static clearStoredErrors() {
        localStorage.removeItem('entitydb-errors');
    }

    /**
     * Export errors as downloadable file
     */
    static exportErrors() {
        const errors = Logger.getStoredErrors();
        const data = JSON.stringify(errors, null, 2);
        const blob = new Blob([data], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `entitydb-errors-${new Date().toISOString()}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }

    /**
     * Set global log level
     */
    static setLogLevel(level) {
        localStorage.setItem('entitydb-log-level', level);
    }

    /**
     * Enable/disable debug mode globally
     */
    static setDebugMode(enabled) {
        localStorage.setItem('entitydb-debug', enabled ? 'true' : 'false');
    }

    /**
     * Create logger instance with context
     */
    static create(component) {
        return new Logger(component);
    }
}

// Export for use in other modules
window.Logger = Logger;

// Create global instance for quick access
window.entitydbLogger = {
    setDebugMode: Logger.setDebugMode,
    setLogLevel: Logger.setLogLevel,
    getStoredErrors: Logger.getStoredErrors,
    clearStoredErrors: Logger.clearStoredErrors,
    exportErrors: Logger.exportErrors
};