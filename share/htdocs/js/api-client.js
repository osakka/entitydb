/**
 * EntityDB API Client - Modular API-first architecture
 * Professional HTTP client with error handling, authentication, and retry logic
 */
class EntityDBAPIClient {
    constructor(baseUrl = '') {
        this.baseUrl = baseUrl;
        this.token = localStorage.getItem('entitydb_token');
        this.retryAttempts = 3;
        this.retryDelay = 1000;
    }

    // Authentication methods
    async login(username, password) {
        const response = await this.post('/auth/login', { username, password });
        if (response.token) {
            this.token = response.token;
            localStorage.setItem('entitydb_token', this.token);
        }
        return response;
    }

    async logout() {
        try {
            await this.post('/auth/logout');
        } finally {
            this.token = null;
            localStorage.removeItem('entitydb_token');
        }
    }

    async whoami() {
        return this.get('/auth/whoami');
    }

    // Entity operations
    async getEntity(id) {
        return this.get(`/entities/get?id=${encodeURIComponent(id)}`);
    }

    async listEntities(params = {}) {
        const query = new URLSearchParams(params).toString();
        return this.get(`/entities/list${query ? '?' + query : ''}`);
    }

    async queryEntities(params = {}) {
        const query = new URLSearchParams(params).toString();
        return this.get(`/entities/query${query ? '?' + query : ''}`);
    }

    async createEntity(entity) {
        return this.post('/entities/create', entity);
    }

    async updateEntity(entity) {
        return this.put('/entities/update', entity);
    }

    // Relationship operations - NEW API-first endpoints
    async discoverRelationships(entityId) {
        return this.get(`/entity-relationships/${entityId}/discover`);
    }

    async getEntityNetwork(entityId, depth = 1) {
        return this.get(`/entity-relationships/${entityId}/network/${depth}`);
    }

    async getRelatedByTags(entityId, params = {}) {
        const query = new URLSearchParams(params).toString();
        return this.get(`/entity-relationships/${entityId}/related${query ? '?' + query : ''}`);
    }

    // Temporal operations
    async getEntityHistory(id, params = {}) {
        const query = new URLSearchParams({ id, ...params }).toString();
        return this.get(`/entities/history?${query}`);
    }

    async getEntityAsOf(id, timestamp) {
        return this.get(`/entities/as-of?id=${encodeURIComponent(id)}&timestamp=${encodeURIComponent(timestamp)}`);
    }

    async getEntityDiff(id, fromTimestamp, toTimestamp) {
        return this.get(`/entities/diff?id=${encodeURIComponent(id)}&from=${encodeURIComponent(fromTimestamp)}&to=${encodeURIComponent(toTimestamp)}`);
    }

    // System operations
    async getSystemMetrics() {
        return this.get('/system/metrics');
    }

    async getHealth() {
        return this.get('/health');
    }

    // HTTP methods with error handling and retry logic
    async get(endpoint) {
        return this.request('GET', endpoint);
    }

    async post(endpoint, data) {
        return this.request('POST', endpoint, data);
    }

    async put(endpoint, data) {
        return this.request('PUT', endpoint, data);
    }

    async delete(endpoint) {
        return this.request('DELETE', endpoint);
    }

    async request(method, endpoint, data = null, attempt = 1) {
        const url = `${this.baseUrl}/api/v1${endpoint}`;
        
        const options = {
            method,
            headers: {
                'Content-Type': 'application/json',
                ...(this.token && { 'Authorization': `Bearer ${this.token}` })
            }
        };

        if (data && (method === 'POST' || method === 'PUT')) {
            options.body = JSON.stringify(data);
        }

        try {
            const response = await fetch(url, options);
            
            // Handle authentication errors
            if (response.status === 401) {
                this.token = null;
                localStorage.removeItem('entitydb_token');
                throw new APIError('Authentication required', 401);
            }

            // Handle other HTTP errors
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new APIError(
                    errorData.error || `HTTP ${response.status} ${response.statusText}`,
                    response.status,
                    errorData
                );
            }

            return await response.json();
        } catch (error) {
            // Retry logic for network errors
            if (attempt < this.retryAttempts && this.shouldRetry(error)) {
                await this.delay(this.retryDelay * attempt);
                return this.request(method, endpoint, data, attempt + 1);
            }
            throw error;
        }
    }

    shouldRetry(error) {
        // Retry on network errors but not on 4xx client errors
        return error instanceof TypeError || 
               (error instanceof APIError && error.status >= 500);
    }

    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
}

/**
 * Custom API Error class for better error handling
 */
class APIError extends Error {
    constructor(message, status, data = null) {
        super(message);
        this.name = 'APIError';
        this.status = status;
        this.data = data;
    }
}

// Export singleton instance
window.apiClient = new EntityDBAPIClient();