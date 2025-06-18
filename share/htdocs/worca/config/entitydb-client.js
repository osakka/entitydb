// EntityDB Client Wrapper for Worca
// Handles all EntityDB API communication with proper v2.32.4 compatibility

class EntityDBClient {
    constructor(config = null) {
        this.config = config || window.worcaConfig;
        this.token = localStorage.getItem('authToken');
        this.refreshTimer = null;
        this.requestQueue = [];
        this.isOnline = true;
        this.retryQueue = [];
        
        // Bind methods for proper context
        this.refreshToken = this.refreshToken.bind(this);
        
        // Initialize token refresh monitoring
        this.initializeTokenManagement();
        
        // Monitor online/offline status
        this.initializeNetworkMonitoring();
    }

    // Connection Management
    getBaseURL() {
        return this.config.getEntityDBUrl();
    }

    getAuthHeaders() {
        const headers = {
            'Content-Type': 'application/json'
        };
        
        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }
        
        return headers;
    }

    // Authentication
    async login(username, password) {
        try {
            const baseURL = this.getBaseURL();
            console.log('🔐 Attempting login to EntityDB...');
            console.log('🌐 EntityDB URL:', baseURL);
            console.log('🔍 Current window location:', window.location.href);
            console.log('📡 Login endpoint:', `${baseURL}/auth/login`);
            
            const response = await this.makeRequest('/auth/login', {
                method: 'POST',
                skipAuth: true,
                body: JSON.stringify({ username, password })
            });

            if (response.ok) {
                const data = await response.json();
                console.log('✅ Login successful');
                
                this.token = data.token;
                localStorage.setItem('authToken', this.token);
                
                // Parse token to get expiry
                this.parseTokenExpiry(data.token);
                
                // Update config status
                this.config.status.authenticated = true;
                this.config.emit('auth-login', { user: data.user, token: data.token });
                
                return data;
            } else {
                const error = await this.parseErrorResponse(response);
                throw new Error(error.message || 'Login failed');
            }
        } catch (error) {
            console.error('❌ Login failed:', error);
            this.config.status.authenticated = false;
            this.config.emit('auth-error', { type: 'login', error: error.message });
            throw error;
        }
    }

    async logout() {
        try {
            if (this.token) {
                // Attempt server-side logout
                await this.makeRequest('/auth/logout', {
                    method: 'POST'
                });
            }
        } catch (error) {
            console.warn('Server logout failed:', error);
        } finally {
            // Always clear local state
            this.token = null;
            localStorage.removeItem('authToken');
            this.clearTokenRefresh();
            this.config.status.authenticated = false;
            this.config.emit('auth-logout');
        }
    }

    async refreshToken() {
        try {
            if (!this.token) return false;
            
            const response = await this.makeRequest('/auth/refresh', {
                method: 'POST'
            });

            if (response.ok) {
                const data = await response.json();
                this.token = data.token;
                localStorage.setItem('authToken', this.token);
                this.parseTokenExpiry(data.token);
                this.config.emit('auth-refresh', data);
                return true;
            } else {
                console.warn('Token refresh failed, logging out');
                await this.logout();
                return false;
            }
        } catch (error) {
            console.error('Token refresh error:', error);
            await this.logout();
            return false;
        }
    }

    parseTokenExpiry(token) {
        try {
            // JWT tokens have expiry in payload (basic parsing)
            const payload = JSON.parse(atob(token.split('.')[1]));
            if (payload.exp) {
                const expiryTime = payload.exp * 1000; // Convert to milliseconds
                const refreshTime = expiryTime - this.config.get('auth.tokenRefreshBuffer');
                const now = Date.now();
                
                if (refreshTime > now) {
                    this.scheduleTokenRefresh(refreshTime - now);
                } else {
                    // Token already expired or will expire soon
                    setTimeout(() => this.refreshToken(), 1000);
                }
            }
        } catch (error) {
            console.warn('Failed to parse token expiry:', error);
            // Schedule refresh in 30 minutes as fallback
            this.scheduleTokenRefresh(30 * 60 * 1000);
        }
    }

    scheduleTokenRefresh(delay) {
        this.clearTokenRefresh();
        this.refreshTimer = setTimeout(this.refreshToken, delay);
    }

    clearTokenRefresh() {
        if (this.refreshTimer) {
            clearTimeout(this.refreshTimer);
            this.refreshTimer = null;
        }
    }

    initializeTokenManagement() {
        // Check if we have a stored token
        if (this.token) {
            this.parseTokenExpiry(this.token);
            this.config.status.authenticated = true;
        }
    }

    // Network Monitoring
    initializeNetworkMonitoring() {
        window.addEventListener('online', () => {
            this.isOnline = true;
            this.config.emit('network-online');
            this.processRetryQueue();
        });

        window.addEventListener('offline', () => {
            this.isOnline = false;
            this.config.emit('network-offline');
        });
    }

    // HTTP Request Handler
    async makeRequest(endpoint, options = {}) {
        const url = `${this.getBaseURL()}${endpoint}`;
        
        const defaultOptions = {
            method: 'GET',
            headers: options.skipAuth ? { 'Content-Type': 'application/json' } : this.getAuthHeaders(),
            timeout: this.config.get('entitydb.timeout'),
            skipAuth: false,
            retry: true
        };

        const requestOptions = { ...defaultOptions, ...options };
        
        // Add timeout handling
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), requestOptions.timeout);
        requestOptions.signal = controller.signal;

        try {
            console.log(`🌐 ${requestOptions.method} ${url}`);
            console.log('🔧 Request options:', {
                method: requestOptions.method,
                headers: requestOptions.headers,
                timeout: requestOptions.timeout
            });
            
            const response = await fetch(url, requestOptions);
            clearTimeout(timeoutId);
            
            // Handle authentication errors
            if (response.status === 401) {
                if (!options.skipAuth && options.retry !== false) {
                    console.log('🔄 Attempting token refresh...');
                    const refreshed = await this.refreshToken();
                    if (refreshed) {
                        // Retry with new token
                        return this.makeRequest(endpoint, { ...options, retry: false });
                    }
                }
                
                // Force logout on auth failure
                await this.logout();
                throw new Error('Authentication failed');
            }
            
            return response;
            
        } catch (error) {
            clearTimeout(timeoutId);
            
            if (error.name === 'AbortError') {
                throw new Error(`Request timeout (${requestOptions.timeout}ms)`);
            }
            
            // Handle network errors with retry queue
            if (!this.isOnline && requestOptions.retry !== false) {
                this.addToRetryQueue(endpoint, options);
                throw new Error('Network unavailable - request queued for retry');
            }
            
            throw error;
        }
    }

    async parseErrorResponse(response) {
        try {
            const text = await response.text();
            try {
                return JSON.parse(text);
            } catch {
                return { message: text || response.statusText, status: response.status };
            }
        } catch {
            return { message: response.statusText, status: response.status };
        }
    }

    // Retry Queue Management
    addToRetryQueue(endpoint, options) {
        this.retryQueue.push({
            endpoint,
            options,
            timestamp: Date.now(),
            attempts: 0
        });
        
        // Limit queue size
        if (this.retryQueue.length > 100) {
            this.retryQueue.shift();
        }
    }

    async processRetryQueue() {
        if (!this.isOnline || this.retryQueue.length === 0) return;
        
        console.log(`🔄 Processing ${this.retryQueue.length} queued requests...`);
        
        const queue = [...this.retryQueue];
        this.retryQueue = [];
        
        for (const item of queue) {
            try {
                item.attempts++;
                if (item.attempts <= this.config.get('entitydb.retries')) {
                    await this.makeRequest(item.endpoint, { ...item.options, retry: false });
                    console.log(`✅ Retry successful: ${item.endpoint}`);
                } else {
                    console.warn(`❌ Retry limit exceeded: ${item.endpoint}`);
                }
            } catch (error) {
                console.warn(`🔄 Retry failed: ${item.endpoint}:`, error.message);
                // Re-queue if not at retry limit
                if (item.attempts < this.config.get('entitydb.retries')) {
                    this.retryQueue.push(item);
                }
            }
        }
    }

    // Entity Operations
    async createEntity(entityData) {
        try {
            // Ensure proper format for EntityDB v2.32.4
            const formattedData = this.formatEntityForCreation(entityData);
            
            console.log('📝 Creating entity:', formattedData);
            
            const response = await this.makeRequest('/entities/create', {
                method: 'POST',
                body: JSON.stringify(formattedData)
            });

            if (response.ok) {
                const result = await response.json();
                console.log('✅ Entity created:', result.id);
                this.config.emit('entity-created', result);
                return result;
            } else {
                const error = await this.parseErrorResponse(response);
                throw new Error(error.message || 'Failed to create entity');
            }
        } catch (error) {
            console.error('❌ Create entity failed:', error);
            this.config.emit('entity-error', { operation: 'create', error: error.message });
            throw error;
        }
    }

    async updateEntity(id, entityData) {
        try {
            const formattedData = this.formatEntityForUpdate(id, entityData);
            
            console.log('📝 Updating entity:', id);
            
            const response = await this.makeRequest('/entities/update', {
                method: 'PUT',
                body: JSON.stringify(formattedData)
            });

            if (response.ok) {
                const result = await response.json();
                console.log('✅ Entity updated:', id);
                this.config.emit('entity-updated', result);
                return result;
            } else {
                const error = await this.parseErrorResponse(response);
                throw new Error(error.message || 'Failed to update entity');
            }
        } catch (error) {
            console.error('❌ Update entity failed:', error);
            this.config.emit('entity-error', { operation: 'update', error: error.message });
            throw error;
        }
    }

    async getEntity(id) {
        try {
            const response = await this.makeRequest(`/entities/get?id=${encodeURIComponent(id)}`);

            if (response.ok) {
                const result = await response.json();
                return this.transformEntityFromDB(result);
            } else {
                const error = await this.parseErrorResponse(response);
                throw new Error(error.message || 'Failed to get entity');
            }
        } catch (error) {
            console.error('❌ Get entity failed:', error);
            throw error;
        }
    }

    async queryEntities(filters = {}) {
        try {
            const params = new URLSearchParams();
            
            // Add worca namespace filter
            if (!filters.tag && !filters.namespace) {
                params.append('tag', `namespace:${this.config.get('dataset.namespace')}`);
            }
            
            // Add other filters
            Object.entries(filters).forEach(([key, value]) => {
                if (value !== undefined && value !== null && value !== '') {
                    params.append(key, value);
                }
            });

            const url = `/entities/list?${params.toString()}`;
            console.log('🔍 Querying entities:', url);

            const response = await this.makeRequest(url);

            if (response.ok) {
                const result = await response.json();
                
                // Handle different response formats
                let entities = [];
                if (result && Array.isArray(result.entities)) {
                    entities = result.entities;
                } else if (Array.isArray(result)) {
                    entities = result;
                } else {
                    console.warn('Unexpected query response format:', result);
                    entities = [];
                }
                
                // Transform entities for worca
                const transformed = entities.map(entity => this.transformEntityFromDB(entity));
                
                console.log(`✅ Query returned ${transformed.length} entities`);
                return transformed;
            } else {
                const error = await this.parseErrorResponse(response);
                throw new Error(error.message || 'Failed to query entities');
            }
        } catch (error) {
            console.error('❌ Query entities failed:', error);
            throw error;
        }
    }

    // Entity Format Transformation
    formatEntityForCreation(entityData) {
        const formatted = {
            tags: [],
            content: null
        };

        // Add namespace tag
        const namespace = this.config.get('dataset.namespace');
        formatted.tags.push(`namespace:${namespace}`);

        // Copy existing tags
        if (entityData.tags && Array.isArray(entityData.tags)) {
            formatted.tags.push(...entityData.tags);
        }

        // Add creation timestamp if not present
        const hasCreatedTag = formatted.tags.some(tag => tag.startsWith('created:'));
        if (!hasCreatedTag) {
            formatted.tags.push(`created:${new Date().toISOString()}`);
        }

        // Format content for EntityDB binary storage
        if (entityData.content) {
            if (typeof entityData.content === 'string') {
                // Store as base64 for binary compatibility
                formatted.content = btoa(entityData.content);
            } else if (Array.isArray(entityData.content) || typeof entityData.content === 'object') {
                // Encode complex objects as JSON then base64
                const jsonString = JSON.stringify(entityData.content);
                formatted.content = btoa(jsonString);
            } else {
                formatted.content = entityData.content;
            }
        }

        return formatted;
    }

    formatEntityForUpdate(id, entityData) {
        const formatted = { id };

        if (entityData.tags) {
            formatted.tags = [...entityData.tags];
            
            // Update timestamp
            formatted.tags = formatted.tags.filter(tag => !tag.startsWith('updated:'));
            formatted.tags.push(`updated:${new Date().toISOString()}`);
        }

        if (entityData.content !== undefined) {
            if (typeof entityData.content === 'string') {
                formatted.content = btoa(entityData.content);
            } else if (Array.isArray(entityData.content) || typeof entityData.content === 'object') {
                const jsonString = JSON.stringify(entityData.content);
                formatted.content = btoa(jsonString);
            } else {
                formatted.content = entityData.content;
            }
        }

        return formatted;
    }

    transformEntityFromDB(entity) {
        if (!entity) return null;

        const transformed = {
            id: entity.id,
            tags: entity.tags || [],
            content: null,
            createdAt: entity.created_at,
            updatedAt: entity.updated_at
        };

        // Decode content from EntityDB format
        if (entity.content) {
            try {
                if (typeof entity.content === 'string') {
                    // Decode base64 content
                    const decoded = atob(entity.content);
                    
                    // Try to parse as JSON
                    try {
                        transformed.content = JSON.parse(decoded);
                    } catch {
                        // Use as plain text if not JSON
                        transformed.content = decoded;
                    }
                } else {
                    transformed.content = entity.content;
                }
            } catch (error) {
                console.warn('Failed to decode entity content:', error);
                transformed.content = entity.content;
            }
        }

        return transformed;
    }

    // Health and Status
    async checkHealth() {
        try {
            const response = await this.makeRequest('/health', { skipAuth: true });
            if (response.ok) {
                const data = await response.json();
                return { healthy: true, data };
            } else {
                return { healthy: false, error: `HTTP ${response.status}` };
            }
        } catch (error) {
            return { healthy: false, error: error.message };
        }
    }

    async getSystemMetrics() {
        try {
            const response = await this.makeRequest('/system/metrics', { skipAuth: true });
            if (response.ok) {
                return await response.json();
            } else {
                throw new Error(`HTTP ${response.status}`);
            }
        } catch (error) {
            console.error('Failed to get system metrics:', error);
            return null;
        }
    }

    // Cleanup
    destroy() {
        this.clearTokenRefresh();
        this.retryQueue = [];
        this.config.emit('client-destroyed');
    }
}

// Global instance
window.entityDBClient = new EntityDBClient();

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = EntityDBClient;
}