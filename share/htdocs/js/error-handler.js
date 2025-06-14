// Error Handler System - EntityDB v2.31.0
// Comprehensive error handling and troubleshooting framework

class ErrorHandler {
    constructor() {
        this.errors = [];
        this.maxErrors = 100;
        this.errorThreshold = 5; // Max errors per minute before alert
        this.listeners = [];
        this.debugMode = false;
        this.requestHistory = [];
        this.maxRequestHistory = 50;
        
        // Error categories for classification
        this.categories = {
            API: 'api',
            UI: 'ui',
            NETWORK: 'network',
            AUTH: 'auth',
            VALIDATION: 'validation',
            PERFORMANCE: 'performance',
            UNKNOWN: 'unknown'
        };
        
        // Severity levels
        this.severity = {
            LOW: 'low',
            MEDIUM: 'medium',
            HIGH: 'high',
            CRITICAL: 'critical'
        };
        
        this.initializeGlobalHandlers();
        this.startCleanupTimer();
    }
    
    // Initialize global error handlers
    initializeGlobalHandlers() {
        // Global JavaScript error handler
        window.addEventListener('error', (event) => {
            this.handleError({
                type: 'javascript_error',
                message: event.message,
                filename: event.filename,
                lineno: event.lineno,
                colno: event.colno,
                error: event.error,
                category: this.categories.UI,
                severity: this.severity.HIGH,
                timestamp: new Date().toISOString(),
                stack: event.error?.stack
            });
        });
        
        // Unhandled promise rejection handler
        window.addEventListener('unhandledrejection', (event) => {
            this.handleError({
                type: 'unhandled_promise_rejection',
                message: event.reason?.message || 'Unhandled promise rejection',
                reason: event.reason,
                category: this.categories.API,
                severity: this.severity.HIGH,
                timestamp: new Date().toISOString(),
                stack: event.reason?.stack
            });
        });
        
        // Fetch API wrapper for request monitoring
        this.wrapFetch();
    }
    
    // Wrap fetch to monitor API requests
    wrapFetch() {
        const originalFetch = window.fetch;
        window.fetch = async (...args) => {
            const requestId = this.generateRequestId();
            const startTime = performance.now();
            const [url, options = {}] = args;
            
            // Log request start
            this.logRequest({
                id: requestId,
                url: url,
                method: options.method || 'GET',
                headers: options.headers,
                body: options.body,
                timestamp: new Date().toISOString(),
                startTime: startTime,
                status: 'pending'
            });
            
            try {
                const response = await originalFetch(...args);
                const endTime = performance.now();
                const duration = endTime - startTime;
                
                // Log successful response
                this.logRequest({
                    id: requestId,
                    url: url,
                    method: options.method || 'GET',
                    status: 'completed',
                    statusCode: response.status,
                    statusText: response.statusText,
                    duration: duration,
                    timestamp: new Date().toISOString(),
                    success: response.ok
                });
                
                // Handle HTTP errors
                if (!response.ok) {
                    const errorBody = await response.text().catch(() => 'Unable to read response body');
                    this.handleError({
                        type: 'http_error',
                        message: `HTTP ${response.status}: ${response.statusText}`,
                        url: url,
                        method: options.method || 'GET',
                        statusCode: response.status,
                        statusText: response.statusText,
                        responseBody: errorBody,
                        category: this.categories.API,
                        severity: this.getHttpErrorSeverity(response.status),
                        timestamp: new Date().toISOString(),
                        requestId: requestId,
                        duration: duration
                    });
                }
                
                return response;
            } catch (error) {
                const endTime = performance.now();
                const duration = endTime - startTime;
                
                // Log failed request
                this.logRequest({
                    id: requestId,
                    url: url,
                    method: options.method || 'GET',
                    status: 'failed',
                    error: error.message,
                    duration: duration,
                    timestamp: new Date().toISOString(),
                    success: false
                });
                
                // Handle network errors
                this.handleError({
                    type: 'network_error',
                    message: error.message,
                    url: url,
                    method: options.method || 'GET',
                    category: this.categories.NETWORK,
                    severity: this.severity.HIGH,
                    timestamp: new Date().toISOString(),
                    requestId: requestId,
                    duration: duration,
                    stack: error.stack
                });
                
                throw error;
            }
        };
    }
    
    // Generate unique request ID
    generateRequestId() {
        return 'req_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }
    
    // Get severity based on HTTP status code
    getHttpErrorSeverity(statusCode) {
        if (statusCode >= 500) return this.severity.CRITICAL;
        if (statusCode === 429 || statusCode === 408) return this.severity.HIGH;
        if (statusCode >= 400) return this.severity.MEDIUM;
        return this.severity.LOW;
    }
    
    // Main error handling method
    handleError(errorData) {
        const error = {
            id: this.generateErrorId(),
            ...errorData,
            timestamp: errorData.timestamp || new Date().toISOString(),
            userAgent: navigator.userAgent,
            url: window.location.href,
            resolved: false,
            count: 1
        };
        
        // Check for duplicate errors
        const existingError = this.findSimilarError(error);
        if (existingError) {
            existingError.count++;
            existingError.lastOccurrence = error.timestamp;
        } else {
            this.errors.unshift(error);
            
            // Limit error history
            if (this.errors.length > this.maxErrors) {
                this.errors = this.errors.slice(0, this.maxErrors);
            }
        }
        
        // Notify listeners
        this.notifyListeners(error);
        
        // Auto-recovery attempts
        this.attemptAutoRecovery(error);
        
        // Console logging for development
        if (this.debugMode) {
            console.error('ErrorHandler:', error);
        }
        
        // Check error rate
        this.checkErrorRate();
    }
    
    // Generate unique error ID
    generateErrorId() {
        return 'err_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }
    
    // Find similar existing errors
    findSimilarError(newError) {
        return this.errors.find(error => 
            error.type === newError.type &&
            error.message === newError.message &&
            error.url === newError.url &&
            !error.resolved
        );
    }
    
    // Log API requests for troubleshooting
    logRequest(requestData) {
        this.requestHistory.unshift(requestData);
        
        // Limit request history
        if (this.requestHistory.length > this.maxRequestHistory) {
            this.requestHistory = this.requestHistory.slice(0, this.maxRequestHistory);
        }
        
        // Update existing request if it's a completion/failure
        if (requestData.status !== 'pending') {
            const existingIndex = this.requestHistory.findIndex(req => req.id === requestData.id);
            if (existingIndex > 0) {
                this.requestHistory[existingIndex] = { ...this.requestHistory[existingIndex], ...requestData };
            }
        }
    }
    
    // Attempt automatic error recovery
    attemptAutoRecovery(error) {
        switch (error.category) {
            case this.categories.AUTH:
                this.handleAuthError(error);
                break;
            case this.categories.NETWORK:
                this.handleNetworkError(error);
                break;
            case this.categories.API:
                this.handleApiError(error);
                break;
            case this.categories.UI:
                this.handleUiError(error);
                break;
        }
    }
    
    // Handle authentication errors
    handleAuthError(error) {
        if (error.statusCode === 401) {
            // Token expired - attempt refresh
            const event = new CustomEvent('auth:token-expired', { detail: error });
            window.dispatchEvent(event);
        } else if (error.statusCode === 403) {
            // Insufficient permissions
            const event = new CustomEvent('auth:insufficient-permissions', { detail: error });
            window.dispatchEvent(event);
        }
    }
    
    // Handle network errors
    handleNetworkError(error) {
        // Retry mechanism for network failures
        if (error.type === 'network_error') {
            const event = new CustomEvent('network:retry-available', { detail: error });
            window.dispatchEvent(event);
        }
    }
    
    // Handle API errors
    handleApiError(error) {
        if (error.statusCode >= 500) {
            // Server error - suggest retry
            const event = new CustomEvent('api:server-error', { detail: error });
            window.dispatchEvent(event);
        }
    }
    
    // Handle UI errors
    handleUiError(error) {
        // Component re-initialization might be needed
        const event = new CustomEvent('ui:component-error', { detail: error });
        window.dispatchEvent(event);
    }
    
    // Check error rate and alert if threshold exceeded
    checkErrorRate() {
        const oneMinuteAgo = new Date(Date.now() - 60000);
        const recentErrors = this.errors.filter(error => 
            new Date(error.timestamp) > oneMinuteAgo
        );
        
        if (recentErrors.length >= this.errorThreshold) {
            const event = new CustomEvent('error:rate-threshold-exceeded', {
                detail: {
                    count: recentErrors.length,
                    threshold: this.errorThreshold,
                    errors: recentErrors
                }
            });
            window.dispatchEvent(event);
        }
    }
    
    // Add error listener
    addListener(callback) {
        this.listeners.push(callback);
    }
    
    // Remove error listener
    removeListener(callback) {
        const index = this.listeners.indexOf(callback);
        if (index > -1) {
            this.listeners.splice(index, 1);
        }
    }
    
    // Notify all listeners
    notifyListeners(error) {
        this.listeners.forEach(callback => {
            try {
                callback(error);
            } catch (e) {
                console.warn('Error in error listener:', e);
            }
        });
    }
    
    // Get error statistics
    getErrorStats() {
        const now = Date.now();
        const oneHourAgo = now - 3600000;
        const oneDayAgo = now - 86400000;
        
        const recentErrors = this.errors.filter(error => 
            new Date(error.timestamp).getTime() > oneHourAgo
        );
        const dailyErrors = this.errors.filter(error => 
            new Date(error.timestamp).getTime() > oneDayAgo
        );
        
        const categoryCounts = {};
        const severityCounts = {};
        
        this.errors.forEach(error => {
            categoryCounts[error.category] = (categoryCounts[error.category] || 0) + error.count;
            severityCounts[error.severity] = (severityCounts[error.severity] || 0) + error.count;
        });
        
        return {
            total: this.errors.length,
            recent: recentErrors.length,
            daily: dailyErrors.length,
            categories: categoryCounts,
            severity: severityCounts,
            resolved: this.errors.filter(e => e.resolved).length,
            unresolved: this.errors.filter(e => !e.resolved).length
        };
    }
    
    // Get request statistics
    getRequestStats() {
        const successfulRequests = this.requestHistory.filter(req => req.success === true);
        const failedRequests = this.requestHistory.filter(req => req.success === false);
        const pendingRequests = this.requestHistory.filter(req => req.status === 'pending');
        
        const avgDuration = successfulRequests.length > 0 
            ? successfulRequests.reduce((sum, req) => sum + (req.duration || 0), 0) / successfulRequests.length 
            : 0;
        
        return {
            total: this.requestHistory.length,
            successful: successfulRequests.length,
            failed: failedRequests.length,
            pending: pendingRequests.length,
            successRate: this.requestHistory.length > 0 
                ? (successfulRequests.length / (successfulRequests.length + failedRequests.length)) * 100 
                : 0,
            averageDuration: avgDuration
        };
    }
    
    // Mark error as resolved
    resolveError(errorId) {
        const error = this.errors.find(e => e.id === errorId);
        if (error) {
            error.resolved = true;
            error.resolvedAt = new Date().toISOString();
        }
    }
    
    // Clear all errors
    clearErrors() {
        this.errors = [];
    }
    
    // Clear request history
    clearRequestHistory() {
        this.requestHistory = [];
    }
    
    // Get filtered errors
    getErrors(filter = {}) {
        let filtered = [...this.errors];
        
        if (filter.category) {
            filtered = filtered.filter(e => e.category === filter.category);
        }
        if (filter.severity) {
            filtered = filtered.filter(e => e.severity === filter.severity);
        }
        if (filter.resolved !== undefined) {
            filtered = filtered.filter(e => e.resolved === filter.resolved);
        }
        if (filter.since) {
            const sinceDate = new Date(filter.since);
            filtered = filtered.filter(e => new Date(e.timestamp) > sinceDate);
        }
        
        return filtered;
    }
    
    // Get filtered requests
    getRequests(filter = {}) {
        let filtered = [...this.requestHistory];
        
        if (filter.method) {
            filtered = filtered.filter(r => r.method === filter.method);
        }
        if (filter.status) {
            filtered = filtered.filter(r => r.status === filter.status);
        }
        if (filter.success !== undefined) {
            filtered = filtered.filter(r => r.success === filter.success);
        }
        if (filter.since) {
            const sinceDate = new Date(filter.since);
            filtered = filtered.filter(r => new Date(r.timestamp) > sinceDate);
        }
        
        return filtered;
    }
    
    // Export error data for analysis
    exportErrorData() {
        return {
            errors: this.errors,
            requests: this.requestHistory,
            stats: this.getErrorStats(),
            requestStats: this.getRequestStats(),
            timestamp: new Date().toISOString(),
            userAgent: navigator.userAgent,
            url: window.location.href
        };
    }
    
    // Import error data
    importErrorData(data) {
        if (data.errors) {
            this.errors = data.errors;
        }
        if (data.requests) {
            this.requestHistory = data.requests;
        }
    }
    
    // Enable debug mode
    enableDebug() {
        this.debugMode = true;
        console.log('ErrorHandler: Debug mode enabled');
    }
    
    // Disable debug mode
    disableDebug() {
        this.debugMode = false;
    }
    
    // Start cleanup timer
    startCleanupTimer() {
        // Clean up old errors every 5 minutes
        setInterval(() => {
            const cutoff = new Date(Date.now() - 24 * 60 * 60 * 1000); // 24 hours ago
            this.errors = this.errors.filter(error => 
                new Date(error.timestamp) > cutoff
            );
            this.requestHistory = this.requestHistory.filter(request => 
                new Date(request.timestamp) > cutoff
            );
        }, 5 * 60 * 1000);
    }
    
    // Create error manually (for testing or custom errors)
    createError(type, message, category = this.categories.UNKNOWN, severity = this.severity.MEDIUM, additional = {}) {
        this.handleError({
            type: type,
            message: message,
            category: category,
            severity: severity,
            ...additional
        });
    }
}

// Global error handler instance
window.ErrorHandler = window.ErrorHandler || new ErrorHandler();

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ErrorHandler;
}