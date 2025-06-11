/**
 * EntityDB API Client
 * Centralized API client with comprehensive error handling and debugging support
 * Version: v2.29.0
 */

class EntityDBClient {
    constructor() {
        this.baseURL = window.location.origin;
        this.token = localStorage.getItem('entitydb-admin-token');
        this.debug = localStorage.getItem('entitydb-debug') === 'true';
        this.defaultTimeout = 30000; // 30 seconds
        this.retryAttempts = 3;
        this.retryDelay = 1000; // 1 second
    }

    /**
     * Set authentication token
     */
    setToken(token) {
        this.token = token;
        if (token) {
            localStorage.setItem('entitydb-admin-token', token);
        } else {
            localStorage.removeItem('entitydb-admin-token');
        }
    }

    /**
     * Enable/disable debug mode
     */
    setDebug(enabled) {
        this.debug = enabled;
        localStorage.setItem('entitydb-debug', enabled ? 'true' : 'false');
    }

    /**
     * Main request method with retry logic
     */
    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            ...options,
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            credentials: 'same-origin',
            mode: 'same-origin'
        };

        // Add auth header if we have a token
        if (this.token && !endpoint.includes('/auth/login')) {
            config.headers['Authorization'] = `Bearer ${this.token}`;
        }

        // Add timeout
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), options.timeout || this.defaultTimeout);
        config.signal = controller.signal;

        if (this.debug) {
            console.group(`API ${options.method || 'GET'} ${endpoint}`);
            console.log('Request Config:', config);
            if (options.body) {
                console.log('Request Body:', options.body);
            }
            console.time('Request Duration');
        }

        let lastError;
        for (let attempt = 1; attempt <= this.retryAttempts; attempt++) {
            try {
                const response = await fetch(url, config);
                clearTimeout(timeoutId);

                let data;
                const contentType = response.headers.get('content-type');
                if (contentType && contentType.includes('application/json')) {
                    data = await response.json();
                } else {
                    data = await response.text();
                }

                if (this.debug) {
                    console.log('Response Status:', response.status);
                    console.log('Response Headers:', Object.fromEntries(response.headers.entries()));
                    console.log('Response Data:', data);
                    console.timeEnd('Request Duration');
                    console.groupEnd();
                }

                if (!response.ok) {
                    throw new APIError(response.status, data.error || data || 'Request failed', endpoint);
                }

                return data;
            } catch (error) {
                lastError = error;

                if (this.debug) {
                    console.error(`Attempt ${attempt} failed:`, error);
                }

                // Don't retry on client errors (4xx) or abort
                if (error.name === 'AbortError' || (error.status >= 400 && error.status < 500)) {
                    break;
                }

                // Don't retry if this was the last attempt
                if (attempt < this.retryAttempts) {
                    if (this.debug) {
                        console.log(`Retrying in ${this.retryDelay}ms...`);
                    }
                    await this.sleep(this.retryDelay * attempt);
                }
            }
        }

        if (this.debug) {
            console.error('All attempts failed');
            console.timeEnd('Request Duration');
            console.groupEnd();
        }

        throw lastError;
    }

    /**
     * Sleep helper for retry delay
     */
    sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    // Authentication endpoints
    async login(username, password) {
        return this.request('/api/v1/auth/login', {
            method: 'POST',
            body: JSON.stringify({ username, password })
        });
    }

    async logout() {
        this.setToken(null);
        return { success: true };
    }

    // Entity endpoints
    async listEntities(params = {}) {
        const query = new URLSearchParams(params).toString();
        return this.request(`/api/v1/entities/list${query ? '?' + query : ''}`);
    }

    async getEntity(id, params = {}) {
        const query = new URLSearchParams(params).toString();
        return this.request(`/api/v1/entities/get?id=${id}${query ? '&' + query : ''}`);
    }

    async createEntity(entity) {
        return this.request('/api/v1/entities/create', {
            method: 'POST',
            body: JSON.stringify(entity)
        });
    }

    async updateEntity(id, updates) {
        return this.request(`/api/v1/entities/update?id=${id}`, {
            method: 'PUT',
            body: JSON.stringify(updates)
        });
    }

    async deleteEntity(id) {
        return this.request(`/api/v1/entities/delete?id=${id}`, {
            method: 'DELETE'
        });
    }

    // Query endpoints
    async queryEntities(params) {
        return this.request('/api/v1/entities/query', {
            method: 'POST',
            body: JSON.stringify(params)
        });
    }

    // Temporal endpoints
    async getEntityAsOf(id, timestamp) {
        return this.request(`/api/v1/entities/as-of?id=${id}&timestamp=${timestamp}`);
    }

    async getEntityHistory(id, params = {}) {
        const query = new URLSearchParams(params).toString();
        return this.request(`/api/v1/entities/history?id=${id}${query ? '&' + query : ''}`);
    }

    async getEntityChanges(params) {
        const query = new URLSearchParams(params).toString();
        return this.request(`/api/v1/entities/changes${query ? '?' + query : ''}`);
    }

    async getEntityDiff(id, from, to) {
        return this.request(`/api/v1/entities/diff?id=${id}&from=${from}&to=${to}`);
    }

    // Relationship endpoints
    async createRelationship(relationship) {
        return this.request('/api/v1/entity-relationships', {
            method: 'POST',
            body: JSON.stringify(relationship)
        });
    }

    async getRelationships(entityId, params = {}) {
        const query = new URLSearchParams({ entity_id: entityId, ...params }).toString();
        return this.request(`/api/v1/entity-relationships?${query}`);
    }

    // User management endpoints
    async createUser(user) {
        return this.request('/api/v1/users/create', {
            method: 'POST',
            body: JSON.stringify(user)
        });
    }

    async listUsers() {
        return this.request('/api/v1/users/list');
    }

    // Dataset endpoints
    async listDatasets() {
        return this.request('/api/v1/datasets');
    }

    async getDataset(id) {
        return this.request(`/api/v1/datasets/${id}`);
    }

    async createDataset(dataset) {
        return this.request('/api/v1/datasets', {
            method: 'POST',
            body: JSON.stringify(dataset)
        });
    }

    async updateDataset(id, updates) {
        return this.request(`/api/v1/datasets/${id}`, {
            method: 'PUT',
            body: JSON.stringify(updates)
        });
    }

    async deleteDataset(id) {
        return this.request(`/api/v1/datasets/${id}`, {
            method: 'DELETE'
        });
    }

    // Metrics endpoints
    async getHealth() {
        return this.request('/health');
    }

    async getMetrics() {
        return this.request('/metrics');
    }

    async getSystemMetrics() {
        return this.request('/api/v1/system/metrics');
    }

    async getRBACMetrics() {
        return this.request('/api/v1/rbac/metrics');
    }

    async getIntegrityMetrics() {
        return this.request('/api/v1/integrity/metrics');
    }

    async getApplicationMetrics(namespace) {
        const params = namespace ? { namespace } : {};
        const query = new URLSearchParams(params).toString();
        return this.request(`/api/v1/application/metrics${query ? '?' + query : ''}`);
    }

    async getMetricsHistory(name, params = {}) {
        const query = new URLSearchParams({ name, ...params }).toString();
        return this.request(`/api/v1/metrics/history?${query}`);
    }

    // Dashboard endpoints
    async getDashboardStats() {
        return this.request('/api/v1/dashboard/stats');
    }

    // Configuration endpoints
    async getConfig() {
        return this.request('/api/v1/config');
    }

    async setFeatureFlag(flag, value) {
        return this.request('/api/v1/feature-flags/set', {
            method: 'POST',
            body: JSON.stringify({ flag, value })
        });
    }
}

/**
 * Custom API Error class
 */
class APIError extends Error {
    constructor(status, message, endpoint) {
        super(message);
        this.name = 'APIError';
        this.status = status;
        this.endpoint = endpoint;
    }

    toString() {
        return `APIError ${this.status}: ${this.message} (${this.endpoint})`;
    }
}

// Export for use in other modules
window.EntityDBClient = EntityDBClient;
window.APIError = APIError;