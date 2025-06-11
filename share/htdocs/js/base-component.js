/**
 * EntityDB Base Component
 * Base Vue component with common functionality for all EntityDB components
 * Version: v2.29.0
 */

const BaseComponent = {
    props: {
        sessionToken: {
            type: String,
            required: true
        },
        currentDataset: {
            type: String,
            default: 'default'
        },
        isDarkMode: {
            type: Boolean,
            default: false
        }
    },
    
    emits: ['notification', 'error', 'loading', 'update'],
    
    data() {
        return {
            loading: false,
            error: null,
            logger: null,
            api: null,
            lastError: null,
            loadingMessage: ''
        };
    },
    
    created() {
        // Initialize logger with component name
        this.logger = new Logger(this.$options.name || 'Component');
        this.logger.info('Component created', {
            dataset: this.currentDataset,
            darkMode: this.isDarkMode
        });
        
        // Initialize API client
        this.api = new EntityDBClient();
        this.api.token = this.sessionToken;
        
        // Watch for token changes
        this.$watch('sessionToken', (newToken) => {
            this.api.token = newToken;
            this.logger.debug('Session token updated');
        });
    },
    
    mounted() {
        this.logger.debug('Component mounted');
    },
    
    beforeUnmount() {
        this.logger.debug('Component unmounting');
    },
    
    methods: {
        /**
         * Execute async operation with loading state and error handling
         */
        async executeAsync(operation, loadingMessage = 'Loading...') {
            this.loading = true;
            this.error = null;
            this.loadingMessage = loadingMessage;
            this.$emit('loading', { loading: true, message: loadingMessage });
            
            this.logger.time(loadingMessage);
            
            try {
                this.logger.debug(`Starting operation: ${loadingMessage}`);
                const result = await operation();
                this.logger.info(`Operation completed: ${loadingMessage}`, { result });
                this.logger.timeEnd(loadingMessage);
                return result;
            } catch (error) {
                this.error = error.message || 'An error occurred';
                this.lastError = error;
                this.logger.error(`Operation failed: ${loadingMessage}`, error);
                this.logger.timeEnd(loadingMessage);
                this.$emit('error', {
                    message: this.error,
                    error: error,
                    operation: loadingMessage
                });
                throw error;
            } finally {
                this.loading = false;
                this.loadingMessage = '';
                this.$emit('loading', { loading: false });
            }
        },
        
        /**
         * Show notification
         */
        notify(message, type = 'info', duration = 5000) {
            this.logger.debug(`Notification: ${message}`, { type });
            this.$emit('notification', { message, type, duration });
        },
        
        /**
         * Handle API errors with user-friendly messages
         */
        handleApiError(error, defaultMessage = 'An error occurred') {
            let message = defaultMessage;
            
            if (error instanceof APIError) {
                switch (error.status) {
                    case 401:
                        message = 'Authentication required. Please log in again.';
                        break;
                    case 403:
                        message = 'You do not have permission to perform this action.';
                        break;
                    case 404:
                        message = 'The requested resource was not found.';
                        break;
                    case 409:
                        message = 'This action conflicts with existing data.';
                        break;
                    case 500:
                        message = 'Server error. Please try again later.';
                        break;
                    default:
                        message = error.message || defaultMessage;
                }
            } else {
                message = error.message || defaultMessage;
            }
            
            this.notify(message, 'error');
            return message;
        },
        
        /**
         * Format bytes for display
         */
        formatBytes(bytes) {
            const units = ['B', 'KB', 'MB', 'GB', 'TB'];
            let i = 0;
            while (bytes > 1024 && i < units.length - 1) {
                bytes /= 1024;
                i++;
            }
            return `${bytes.toFixed(i > 0 ? 2 : 0)} ${units[i]}`;
        },
        
        /**
         * Format timestamp for display
         */
        formatTimestamp(timestamp) {
            if (!timestamp) return '';
            const date = new Date(timestamp);
            return date.toLocaleString();
        },
        
        /**
         * Format relative time
         */
        formatRelativeTime(timestamp) {
            if (!timestamp) return '';
            const now = Date.now();
            const then = new Date(timestamp).getTime();
            const diff = now - then;
            
            const seconds = Math.floor(diff / 1000);
            const minutes = Math.floor(seconds / 60);
            const hours = Math.floor(minutes / 60);
            const days = Math.floor(hours / 24);
            
            if (days > 0) return `${days}d ago`;
            if (hours > 0) return `${hours}h ago`;
            if (minutes > 0) return `${minutes}m ago`;
            return `${seconds}s ago`;
        },
        
        /**
         * Debounce function for search/filter inputs
         */
        debounce(func, wait) {
            let timeout;
            return function executedFunction(...args) {
                const later = () => {
                    clearTimeout(timeout);
                    func(...args);
                };
                clearTimeout(timeout);
                timeout = setTimeout(later, wait);
            };
        },
        
        /**
         * Check if user has permission
         */
        hasPermission(permission) {
            // This would check against user's actual permissions
            // For now, return true as placeholder
            return true;
        },
        
        /**
         * Reload component data
         */
        async reload() {
            this.logger.info('Reloading component data');
            // Override in child components
        },
        
        /**
         * Export data as CSV
         */
        exportAsCSV(data, filename) {
            this.logger.info('Exporting data as CSV', { filename, rows: data.length });
            
            if (!data || data.length === 0) {
                this.notify('No data to export', 'warning');
                return;
            }
            
            // Get headers from first row
            const headers = Object.keys(data[0]);
            const csvContent = [
                headers.join(','),
                ...data.map(row => 
                    headers.map(header => {
                        const value = row[header];
                        // Escape quotes and wrap in quotes if contains comma
                        const escaped = String(value).replace(/"/g, '""');
                        return escaped.includes(',') ? `"${escaped}"` : escaped;
                    }).join(',')
                )
            ].join('\n');
            
            const blob = new Blob([csvContent], { type: 'text/csv' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${filename}-${new Date().toISOString().split('T')[0]}.csv`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
            
            this.notify(`Exported ${data.length} rows to ${a.download}`, 'success');
        },
        
        /**
         * Export data as JSON
         */
        exportAsJSON(data, filename) {
            this.logger.info('Exporting data as JSON', { filename });
            
            const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${filename}-${new Date().toISOString().split('T')[0]}.json`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
            
            this.notify(`Exported data to ${a.download}`, 'success');
        }
    },
    
    computed: {
        /**
         * Check if component is in loading state
         */
        isLoading() {
            return this.loading;
        },
        
        /**
         * Check if component has error
         */
        hasError() {
            return !!this.error;
        },
        
        /**
         * Get display-friendly dataset name
         */
        datasetDisplayName() {
            if (this.currentDataset === '_system') return 'System';
            return this.currentDataset.charAt(0).toUpperCase() + this.currentDataset.slice(1);
        }
    }
};

// Export for use in other components
window.BaseComponent = BaseComponent;