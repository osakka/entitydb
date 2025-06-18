// Worca Configuration Management System
// Handles EntityDB server connection, dataset management, and feature flags

class WorcaConfig {
    constructor() {
        this.version = '2.32.4';
        this.configKey = 'worca-config-v2';
        
        // Default configuration
        this.defaults = {
            entitydb: {
                host: window?.location?.hostname || 'localhost',
                port: 8085, // EntityDB's actual port
                ssl: true,  // EntityDB is running HTTPS on 8085
                basePath: '/api/v1',
                timeout: 30000,
                retries: 3,
                autoDetect: true,
                healthCheckInterval: 30000 // 30 seconds
            },
            dataset: {
                name: 'worca-workspace',
                namespace: 'worca',
                autoBootstrap: true,
                sampleData: true,
                validation: true
            },
            features: {
                realTimeSync: true,
                syncInterval: 5000, // 5 seconds
                offlineMode: false,
                autoSave: true,
                debugMode: false,
                notifications: true,
                darkMode: false
            },
            ui: {
                theme: 'ocean-light',
                sidebarCollapsed: false,
                gridLayout: 'default',
                chartColors: ['#0891b2', '#06b6d4', '#0ea5e9', '#3b82f6', '#6366f1'],
                animationsEnabled: true
            },
            auth: {
                autoLogin: false,
                rememberUser: true,
                sessionTimeout: 3600000, // 1 hour
                tokenRefreshBuffer: 300000 // 5 minutes before expiry
            }
        };
        
        this.config = this.loadConfig();
        this.status = {
            entitydb: 'unknown',
            lastCheck: null,
            version: null,
            connected: false,
            authenticated: false,
            workspace: null
        };
        
        this.eventListeners = new Map();
        this.healthCheckTimer = null;
        
        // Initialize health monitoring
        this.startHealthMonitoring();
    }

    // Configuration Management
    loadConfig() {
        try {
            const stored = localStorage.getItem(this.configKey);
            if (stored) {
                const parsed = JSON.parse(stored);
                return this.mergeConfig(this.defaults, parsed);
            }
        } catch (error) {
            console.warn('Failed to load worca config, using defaults:', error);
        }
        return JSON.parse(JSON.stringify(this.defaults));
    }

    saveConfig() {
        try {
            localStorage.setItem(this.configKey, JSON.stringify(this.config));
            this.emit('config-saved', this.config);
            return true;
        } catch (error) {
            console.error('Failed to save worca config:', error);
            return false;
        }
    }

    mergeConfig(defaults, override) {
        const result = JSON.parse(JSON.stringify(defaults));
        for (const [key, value] of Object.entries(override)) {
            if (typeof value === 'object' && !Array.isArray(value) && value !== null) {
                result[key] = this.mergeConfig(defaults[key] || {}, value);
            } else {
                result[key] = value;
            }
        }
        return result;
    }

    // Getters and Setters
    get(path) {
        const keys = path.split('.');
        let current = this.config;
        for (const key of keys) {
            if (current && typeof current === 'object') {
                current = current[key];
            } else {
                return undefined;
            }
        }
        return current;
    }

    set(path, value) {
        const keys = path.split('.');
        let current = this.config;
        
        for (let i = 0; i < keys.length - 1; i++) {
            const key = keys[i];
            if (!current[key] || typeof current[key] !== 'object') {
                current[key] = {};
            }
            current = current[key];
        }
        
        const lastKey = keys[keys.length - 1];
        const oldValue = current[lastKey];
        current[lastKey] = value;
        
        this.saveConfig();
        this.emit('config-changed', { path, value, oldValue });
        
        return true;
    }

    reset(section = null) {
        if (section) {
            this.config[section] = JSON.parse(JSON.stringify(this.defaults[section]));
        } else {
            this.config = JSON.parse(JSON.stringify(this.defaults));
        }
        this.saveConfig();
        this.emit('config-reset', section);
    }

    // EntityDB Connection Management
    getEntityDBUrl() {
        const { host, port, ssl, basePath } = this.config.entitydb;
        const protocol = ssl ? 'https' : 'http';
        return `${protocol}://${host}:${port}${basePath}`;
    }

    async detectEntityDBServer() {
        // Get current host from browser location
        const currentHost = window.location.hostname;
        const currentProtocol = window.location.protocol === 'https:';
        
        const candidates = [
            // Try current host with HTTPS on 8085 (EntityDB's actual config)
            { host: currentHost, port: 8085, ssl: true },
            // Try current host with HTTP on 8085
            { host: currentHost, port: 8085, ssl: false },
            // Try current host with standard ports
            { host: currentHost, port: currentProtocol ? 8443 : 8085, ssl: currentProtocol },
            { host: currentHost, port: currentProtocol ? 8085 : 8443, ssl: !currentProtocol },
            // Fallback to localhost with EntityDB's actual config
            { host: 'localhost', port: 8085, ssl: true },
            { host: 'localhost', port: 8085, ssl: false },
            { host: 'localhost', port: 8443, ssl: true },
            { host: '127.0.0.1', port: 8085, ssl: true },
            { host: '127.0.0.1', port: 8085, ssl: false },
            { host: '127.0.0.1', port: 8443, ssl: true }
        ];

        for (const candidate of candidates) {
            try {
                const protocol = candidate.ssl ? 'https' : 'http';
                const url = `${protocol}://${candidate.host}:${candidate.port}/health`;
                
                console.log(`🔍 Testing: ${url}`);
                
                const response = await fetch(url, {
                    method: 'GET',
                    timeout: 5000,
                    signal: AbortSignal.timeout(5000)
                });

                if (response.ok) {
                    const data = await response.json();
                    if (data.status === 'healthy' || data.status === 'ok' || data.service === 'EntityDB') {
                        console.log('✅ Detected EntityDB server:', url);
                        
                        // Update configuration
                        this.set('entitydb.host', candidate.host);
                        this.set('entitydb.port', candidate.port);
                        this.set('entitydb.ssl', candidate.ssl);
                        
                        return {
                            ...candidate,
                            url,
                            version: data.version || 'unknown',
                            detected: true
                        };
                    }
                }
            } catch (error) {
                // Try next candidate
                continue;
            }
        }

        console.warn('❌ No EntityDB server detected on standard ports');
        return null;
    }

    async checkEntityDBHealth() {
        try {
            const url = `${this.getEntityDBUrl().replace('/api/v1', '')}/health`;
            
            const response = await fetch(url, {
                method: 'GET',
                timeout: this.config.entitydb.timeout,
                signal: AbortSignal.timeout(this.config.entitydb.timeout)
            });

            if (response.ok) {
                const data = await response.json();
                
                this.status.entitydb = 'healthy';
                this.status.connected = true;
                this.status.lastCheck = new Date().toISOString();
                this.status.version = data.version || 'unknown';
                
                this.emit('entitydb-health', {
                    status: 'healthy',
                    data,
                    timestamp: this.status.lastCheck
                });
                
                return { healthy: true, data };
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            this.status.entitydb = 'unhealthy';
            this.status.connected = false;
            this.status.lastCheck = new Date().toISOString();
            
            this.emit('entitydb-health', {
                status: 'unhealthy',
                error: error.message,
                timestamp: this.status.lastCheck
            });
            
            return { healthy: false, error: error.message };
        }
    }

    async validateConnection() {
        // First try health check
        const health = await this.checkEntityDBHealth();
        
        if (!health.healthy) {
            // Try auto-detection if health check failed
            if (this.config.entitydb.autoDetect) {
                console.log('🔍 Auto-detecting EntityDB server...');
                const detected = await this.detectEntityDBServer();
                if (detected) {
                    // Retry health check with new settings
                    return await this.checkEntityDBHealth();
                }
            }
            return health;
        }
        
        return health;
    }

    // Health Monitoring
    startHealthMonitoring() {
        if (this.healthCheckTimer) {
            clearInterval(this.healthCheckTimer);
        }
        
        const interval = this.config.entitydb.healthCheckInterval;
        if (interval > 0) {
            this.healthCheckTimer = setInterval(() => {
                this.checkEntityDBHealth();
            }, interval);
            
            // Initial check
            setTimeout(() => this.validateConnection(), 1000);
        }
    }

    stopHealthMonitoring() {
        if (this.healthCheckTimer) {
            clearInterval(this.healthCheckTimer);
            this.healthCheckTimer = null;
        }
    }

    // Event System
    on(event, callback) {
        if (!this.eventListeners.has(event)) {
            this.eventListeners.set(event, []);
        }
        this.eventListeners.get(event).push(callback);
    }

    off(event, callback) {
        if (this.eventListeners.has(event)) {
            const listeners = this.eventListeners.get(event);
            const index = listeners.indexOf(callback);
            if (index > -1) {
                listeners.splice(index, 1);
            }
        }
    }

    emit(event, data = null) {
        if (this.eventListeners.has(event)) {
            this.eventListeners.get(event).forEach(callback => {
                try {
                    callback(data);
                } catch (error) {
                    console.error(`Error in event listener for ${event}:`, error);
                }
            });
        }
        
        // Also emit to global event system if available
        if (window.worcaEvents) {
            window.worcaEvents.emit(event, data);
        }
    }

    // Workspace Management
    setWorkspace(workspaceName) {
        this.set('dataset.name', workspaceName);
        this.status.workspace = workspaceName;
        this.emit('workspace-changed', workspaceName);
    }

    getWorkspace() {
        return this.config.dataset.name;
    }

    // Theme Management
    setTheme(theme) {
        this.set('ui.theme', theme);
        this.set('features.darkMode', theme.includes('dark'));
        this.emit('theme-changed', theme);
        
        // Apply theme to document
        if (theme.includes('dark')) {
            document.documentElement.setAttribute('data-theme', 'dark');
        } else {
            document.documentElement.removeAttribute('data-theme');
        }
    }

    // Export/Import Configuration
    exportConfig() {
        return {
            version: this.version,
            timestamp: new Date().toISOString(),
            config: this.config
        };
    }

    importConfig(configData) {
        try {
            if (configData.version && configData.config) {
                this.config = this.mergeConfig(this.defaults, configData.config);
                this.saveConfig();
                this.emit('config-imported', configData);
                
                // Restart health monitoring with new settings
                this.startHealthMonitoring();
                
                return true;
            }
            throw new Error('Invalid configuration format');
        } catch (error) {
            console.error('Failed to import configuration:', error);
            return false;
        }
    }

    // Debug and Diagnostics
    getDiagnostics() {
        return {
            version: this.version,
            config: this.config,
            status: this.status,
            entitydbUrl: this.getEntityDBUrl(),
            localStorage: {
                available: typeof Storage !== 'undefined',
                used: JSON.stringify(this.config).length,
                quota: this.getStorageQuota()
            },
            browser: {
                userAgent: navigator.userAgent,
                language: navigator.language,
                online: navigator.onLine
            },
            timestamp: new Date().toISOString()
        };
    }

    getStorageQuota() {
        try {
            // Estimate localStorage usage
            let used = 0;
            for (let key in localStorage) {
                if (localStorage.hasOwnProperty(key)) {
                    used += localStorage[key].length + key.length;
                }
            }
            return { used, estimated: true };
        } catch (error) {
            return { error: error.message };
        }
    }

    // Cleanup
    destroy() {
        this.stopHealthMonitoring();
        this.eventListeners.clear();
        this.emit('config-destroyed');
    }
}

// Global instance
window.worcaConfig = new WorcaConfig();

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = WorcaConfig;
}