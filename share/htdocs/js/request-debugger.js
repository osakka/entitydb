// Request Debugger - EntityDB v2.31.0
// Advanced API request monitoring and debugging tools

class RequestDebugger {
    constructor() {
        this.interceptors = [];
        this.debugLevel = 'info'; // 'verbose', 'info', 'warn', 'error'
        this.maxLogSize = 1000;
        this.logs = [];
        this.filters = {
            methods: [],
            statusCodes: [],
            urls: [],
            enabled: true
        };
        this.performance = {
            slowRequestThreshold: 1000, // ms
            enabled: true
        };
        this.retryConfig = {
            maxRetries: 3,
            retryDelay: 1000,
            retryOnStatus: [500, 502, 503, 504, 408, 429]
        };
        
        this.initializeInterceptors();
    }
    
    // Initialize request/response interceptors
    initializeInterceptors() {
        if (window.ErrorHandler) {
            window.ErrorHandler.addListener((error) => {
                if (error.category === 'api' || error.category === 'network') {
                    this.logRequest({
                        type: 'error',
                        error: error,
                        timestamp: new Date().toISOString()
                    });
                }
            });
        }
    }
    
    // Enhanced fetch wrapper with detailed logging
    wrapFetchDetailed() {
        const originalFetch = window.fetch;
        
        window.fetch = async (...args) => {
            const [url, options = {}] = args;
            const requestId = this.generateRequestId();
            const startTime = performance.now();
            
            // Prepare request details
            const requestDetails = {
                id: requestId,
                url: typeof url === 'string' ? url : url.url,
                method: options.method || 'GET',
                headers: this.sanitizeHeaders(options.headers),
                body: this.sanitizeBody(options.body),
                timestamp: new Date().toISOString(),
                startTime: startTime,
                type: 'request'
            };
            
            this.logRequest(requestDetails);
            
            try {
                // Apply retry logic if enabled
                let response;
                let attempt = 0;
                
                while (attempt <= this.retryConfig.maxRetries) {
                    try {
                        response = await originalFetch(...args);
                        break;
                    } catch (error) {
                        attempt++;
                        if (attempt > this.retryConfig.maxRetries) {
                            throw error;
                        }
                        
                        this.logRequest({
                            id: requestId,
                            type: 'retry',
                            attempt: attempt,
                            error: error.message,
                            timestamp: new Date().toISOString()
                        });
                        
                        await this.delay(this.retryConfig.retryDelay * attempt);
                    }
                }
                
                const endTime = performance.now();
                const duration = endTime - startTime;
                
                // Clone response for body reading
                const responseClone = response.clone();
                let responseBody = '';
                
                try {
                    const contentType = response.headers.get('content-type');
                    if (contentType && contentType.includes('application/json')) {
                        responseBody = await responseClone.json();
                    } else {
                        responseBody = await responseClone.text();
                        if (responseBody.length > 1000) {
                            responseBody = responseBody.substring(0, 1000) + '... (truncated)';
                        }
                    }
                } catch (e) {
                    responseBody = 'Unable to read response body';
                }
                
                // Log response details
                const responseDetails = {
                    id: requestId,
                    type: 'response',
                    status: response.status,
                    statusText: response.statusText,
                    headers: this.sanitizeHeaders(response.headers),
                    body: responseBody,
                    duration: duration,
                    timestamp: new Date().toISOString(),
                    ok: response.ok,
                    redirected: response.redirected,
                    url: response.url
                };
                
                this.logRequest(responseDetails);
                
                // Check for slow requests
                if (this.performance.enabled && duration > this.performance.slowRequestThreshold) {
                    this.logRequest({
                        id: requestId,
                        type: 'performance_warning',
                        message: `Slow request detected: ${duration.toFixed(2)}ms`,
                        threshold: this.performance.slowRequestThreshold,
                        timestamp: new Date().toISOString()
                    });
                }
                
                // Retry on specific status codes if configured
                if (!response.ok && this.retryConfig.retryOnStatus.includes(response.status) && attempt < this.retryConfig.maxRetries) {
                    attempt++;
                    this.logRequest({
                        id: requestId,
                        type: 'retry_on_status',
                        status: response.status,
                        attempt: attempt,
                        timestamp: new Date().toISOString()
                    });
                    
                    await this.delay(this.retryConfig.retryDelay * attempt);
                    return window.fetch(...args);
                }
                
                return response;
                
            } catch (error) {
                const endTime = performance.now();
                const duration = endTime - startTime;
                
                this.logRequest({
                    id: requestId,
                    type: 'error',
                    error: {
                        message: error.message,
                        name: error.name,
                        stack: error.stack
                    },
                    duration: duration,
                    timestamp: new Date().toISOString()
                });
                
                throw error;
            }
        };
    }
    
    // Generate unique request ID
    generateRequestId() {
        return 'debug_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }
    
    // Sanitize headers for logging
    sanitizeHeaders(headers) {
        if (!headers) return {};
        
        const sanitized = {};
        const sensitiveHeaders = ['authorization', 'cookie', 'x-api-key', 'x-auth-token'];
        
        if (headers instanceof Headers) {
            for (const [key, value] of headers.entries()) {
                const lowerKey = key.toLowerCase();
                sanitized[key] = sensitiveHeaders.includes(lowerKey) ? '[REDACTED]' : value;
            }
        } else if (typeof headers === 'object') {
            Object.keys(headers).forEach(key => {
                const lowerKey = key.toLowerCase();
                sanitized[key] = sensitiveHeaders.includes(lowerKey) ? '[REDACTED]' : headers[key];
            });
        }
        
        return sanitized;
    }
    
    // Sanitize request body for logging
    sanitizeBody(body) {
        if (!body) return null;
        
        if (typeof body === 'string') {
            try {
                const parsed = JSON.parse(body);
                return this.sanitizeObject(parsed);
            } catch (e) {
                return body.length > 500 ? body.substring(0, 500) + '... (truncated)' : body;
            }
        }
        
        if (body instanceof FormData) {
            return '[FormData]';
        }
        
        if (body instanceof URLSearchParams) {
            return '[URLSearchParams]';
        }
        
        return body;
    }
    
    // Sanitize object by removing sensitive fields
    sanitizeObject(obj) {
        if (!obj || typeof obj !== 'object') return obj;
        
        const sensitiveFields = ['password', 'token', 'secret', 'key', 'authorization'];
        const sanitized = { ...obj };
        
        Object.keys(sanitized).forEach(key => {
            const lowerKey = key.toLowerCase();
            if (sensitiveFields.some(field => lowerKey.includes(field))) {
                sanitized[key] = '[REDACTED]';
            } else if (typeof sanitized[key] === 'object') {
                sanitized[key] = this.sanitizeObject(sanitized[key]);
            }
        });
        
        return sanitized;
    }
    
    // Log request/response data
    logRequest(data) {
        if (!this.shouldLog(data)) return;
        
        const logEntry = {
            ...data,
            level: this.getLogLevel(data),
            timestamp: data.timestamp || new Date().toISOString()
        };
        
        this.logs.unshift(logEntry);
        
        // Limit log size
        if (this.logs.length > this.maxLogSize) {
            this.logs = this.logs.slice(0, this.maxLogSize);
        }
        
        // Console output for development
        if (this.debugLevel === 'verbose' || (this.debugLevel === 'info' && logEntry.level !== 'debug')) {
            this.consoleLog(logEntry);
        }
        
        // Emit custom event for UI updates
        window.dispatchEvent(new CustomEvent('request-debugger:log', { detail: logEntry }));
    }
    
    // Determine if request should be logged based on filters
    shouldLog(data) {
        if (!this.filters.enabled) return false;
        
        if (this.filters.methods.length > 0 && data.method && !this.filters.methods.includes(data.method)) {
            return false;
        }
        
        if (this.filters.statusCodes.length > 0 && data.status && !this.filters.statusCodes.includes(data.status)) {
            return false;
        }
        
        if (this.filters.urls.length > 0 && data.url) {
            const matchesUrl = this.filters.urls.some(pattern => 
                new RegExp(pattern).test(data.url)
            );
            if (!matchesUrl) return false;
        }
        
        return true;
    }
    
    // Get log level for entry
    getLogLevel(data) {
        if (data.type === 'error') return 'error';
        if (data.type === 'performance_warning') return 'warn';
        if (data.type === 'retry' || data.type === 'retry_on_status') return 'warn';
        if (data.type === 'request' || data.type === 'response') return 'info';
        return 'debug';
    }
    
    // Console logging with formatting
    consoleLog(entry) {
        const prefix = `[RequestDebugger] ${entry.timestamp} [${entry.level.toUpperCase()}]`;
        
        switch (entry.level) {
            case 'error':
                console.error(prefix, entry);
                break;
            case 'warn':
                console.warn(prefix, entry);
                break;
            case 'info':
                console.info(prefix, entry);
                break;
            default:
                console.log(prefix, entry);
        }
    }
    
    // Delay utility for retries
    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
    
    // Get logs with filtering
    getLogs(filter = {}) {
        let filtered = [...this.logs];
        
        if (filter.level) {
            filtered = filtered.filter(log => log.level === filter.level);
        }
        
        if (filter.type) {
            filtered = filtered.filter(log => log.type === filter.type);
        }
        
        if (filter.requestId) {
            filtered = filtered.filter(log => log.id === filter.requestId);
        }
        
        if (filter.since) {
            const sinceDate = new Date(filter.since);
            filtered = filtered.filter(log => new Date(log.timestamp) > sinceDate);
        }
        
        if (filter.limit) {
            filtered = filtered.slice(0, filter.limit);
        }
        
        return filtered;
    }
    
    // Get request timeline by ID
    getRequestTimeline(requestId) {
        return this.logs.filter(log => log.id === requestId).sort((a, b) => 
            new Date(a.timestamp) - new Date(b.timestamp)
        );
    }
    
    // Get statistics
    getStats() {
        const requests = this.logs.filter(log => log.type === 'request');
        const responses = this.logs.filter(log => log.type === 'response');
        const errors = this.logs.filter(log => log.type === 'error');
        const retries = this.logs.filter(log => log.type === 'retry' || log.type === 'retry_on_status');
        
        const successfulResponses = responses.filter(log => log.ok);
        const failedResponses = responses.filter(log => !log.ok);
        
        const durations = responses.filter(log => log.duration).map(log => log.duration);
        const avgDuration = durations.length > 0 
            ? durations.reduce((sum, d) => sum + d, 0) / durations.length 
            : 0;
        
        return {
            totalRequests: requests.length,
            totalResponses: responses.length,
            successfulResponses: successfulResponses.length,
            failedResponses: failedResponses.length,
            errors: errors.length,
            retries: retries.length,
            successRate: responses.length > 0 ? (successfulResponses.length / responses.length) * 100 : 0,
            averageDuration: avgDuration,
            slowRequests: responses.filter(log => log.duration > this.performance.slowRequestThreshold).length
        };
    }
    
    // Clear logs
    clearLogs() {
        this.logs = [];
        window.dispatchEvent(new CustomEvent('request-debugger:logs-cleared'));
    }
    
    // Export logs for analysis
    exportLogs() {
        return {
            logs: this.logs,
            stats: this.getStats(),
            config: {
                debugLevel: this.debugLevel,
                filters: this.filters,
                performance: this.performance,
                retryConfig: this.retryConfig
            },
            timestamp: new Date().toISOString()
        };
    }
    
    // Configure filters
    setFilters(filters) {
        this.filters = { ...this.filters, ...filters };
        window.dispatchEvent(new CustomEvent('request-debugger:filters-updated', { detail: this.filters }));
    }
    
    // Configure performance monitoring
    setPerformanceConfig(config) {
        this.performance = { ...this.performance, ...config };
    }
    
    // Configure retry behavior
    setRetryConfig(config) {
        this.retryConfig = { ...this.retryConfig, ...config };
    }
    
    // Set debug level
    setDebugLevel(level) {
        if (['verbose', 'info', 'warn', 'error'].includes(level)) {
            this.debugLevel = level;
        }
    }
    
    // Enable detailed fetch wrapping
    enableDetailedLogging() {
        this.wrapFetchDetailed();
        console.log('RequestDebugger: Detailed logging enabled');
    }
    
    // Create manual log entry
    createLog(type, data) {
        this.logRequest({
            type: type,
            timestamp: new Date().toISOString(),
            ...data
        });
    }
}

// Global request debugger instance
window.RequestDebugger = window.RequestDebugger || new RequestDebugger();

// Auto-enable detailed logging in development
if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    window.RequestDebugger.enableDetailedLogging();
}

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = RequestDebugger;
}