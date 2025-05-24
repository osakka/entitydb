// MetHub API - EntityDB wrapper for metrics
class MetHubAPI {
    constructor() {
        this.baseUrl = window.location.origin;
        this.token = localStorage.getItem('entitydb-token');
        this.hub = 'metrics';
    }

    // Set auth token
    setToken(token) {
        this.token = token;
        localStorage.setItem('entitydb-token', token);
    }

    // Make API request
    async request(method, endpoint, data = null) {
        const headers = {
            'Content-Type': 'application/json'
        };
        
        // Only add Authorization header if we have a token
        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }
        
        const options = {
            method,
            headers
        };

        if (data) {
            options.body = JSON.stringify(data);
        }

        console.log(`🌐 API Request: ${method} ${this.baseUrl}${endpoint}`, { 
            hasToken: !!this.token, 
            tokenPreview: this.token ? `${this.token.substring(0, 10)}...` : 'none'
        });
        
        try {
            const response = await fetch(`${this.baseUrl}${endpoint}`, options);
            
            if (!response.ok) {
                console.error(`❌ API Error: ${response.status} ${response.statusText}`, {
                    endpoint: `${this.baseUrl}${endpoint}`,
                    method,
                    hasToken: !!this.token
                });
                throw new Error(`API Error: ${response.status} ${response.statusText}`);
            }

            const text = await response.text();
            try {
                return JSON.parse(text);
            } catch (e) {
                console.warn('Response is not JSON:', text);
                return { error: text };
            }
        } catch (error) {
            console.error('Request failed:', error);
            throw error;
        }
    }

    // Login
    async login(username, password) {
        console.log(`🔐 Attempting login for user: ${username}`);
        
        const result = await this.request('POST', '/api/v1/auth/login', {
            username,
            password
        });
        
        if (result.token) {
            console.log(`✅ Login successful, token received: ${result.token.substring(0, 10)}...`);
            this.setToken(result.token);
        } else {
            console.error('❌ Login failed: No token in response', result);
            throw new Error('Login failed: No token received');
        }
        
        return result;
    }

    // Query metrics with time range
    async queryMetrics(timeRange = '1h', host = null, metricType = null, metricName = null) {
        // Convert time range to timestamp
        const now = Date.now() * 1000000; // Convert to nanoseconds
        let since = now;
        
        switch(timeRange) {
            case '5m': since = now - (5 * 60 * 1e9); break;
            case '15m': since = now - (15 * 60 * 1e9); break;
            case '1h': since = now - (60 * 60 * 1e9); break;
            case '6h': since = now - (6 * 60 * 60 * 1e9); break;
            case '24h': since = now - (24 * 60 * 60 * 1e9); break;
            case '7d': since = now - (7 * 24 * 60 * 60 * 1e9); break;
        }
        
        // Build query
        let query = {
            hub: this.hub,
            tags: ['type:metric'],
            since: since.toString(),
            limit: 10000 // Adjust based on needs
        };
        
        if (host) {
            query.tags.push(`host:${host}`);
        }
        
        if (metricType) {
            query.tags.push(`metric:${metricType}`);
        }
        
        if (metricName) {
            query.tags.push(`name:${metricName}`);
        }
        
        // Query entities - build query string for GET request
        let selfFilter = 'type:metric';
        if (host) selfFilter += `,host:${host}`;
        if (metricName) selfFilter += `,name:${metricName}`;
        
        // For metric type, use traits filter instead
        let traitsFilter = '';
        if (metricType) traitsFilter = `metric_type:${metricType}`;
        
        const params = new URLSearchParams({
            hub: this.hub,
            self: selfFilter,
            include_content: 'true',
            include_raw_tags: 'true'
        });
        
        if (traitsFilter) {
            params.append('traits', traitsFilter);
        }
        
        console.log('Query params:', params.toString());
        
        try {
            const result = await this.request('GET', `/api/v1/hubs/entities/query?${params}`);
            console.log('Query result:', result);
            return this.transformMetrics(result.entities || result || []);
        } catch (error) {
            console.error('Failed to query metrics:', error);
            // Try fallback query without hub
            try {
                const fallbackResult = await this.request('GET', `/api/v1/entities/list?tags=type:metric`);
                return this.transformMetrics(fallbackResult || []);
            } catch (fallbackError) {
                console.error('Fallback query also failed:', fallbackError);
                return [];
            }
        }
    }

    // Get latest metric value
    async getLatestMetric(host, metricType, metricName) {
        const metrics = await this.queryMetrics('5m', host, metricType, metricName);
        
        if (metrics.length === 0) {
            return null;
        }
        
        // Sort by timestamp and get latest
        metrics.sort((a, b) => b.timestamp - a.timestamp);
        return metrics[0];
    }

    // Get available hosts
    async getHosts() {
        const params = new URLSearchParams({
            hub: this.hub,
            self: 'type:metric',
            include_content: 'true',
            include_raw_tags: 'true'
        });
        
        const result = await this.request('GET', `/api/v1/hubs/entities/query?${params}`);
        
        const hosts = new Set();
        (result.entities || []).forEach(entity => {
            const host = (entity.traits && entity.traits.host) || (entity.self && entity.self.host);
            if (host) {
                hosts.add(host);
            }
        });
        
        return Array.from(hosts).sort();
    }

    // Transform entities to metric format
    transformMetrics(entities) {
        console.log('Transforming entities:', entities);
        return entities.map(entity => {
            console.log('Processing entity:', entity);
            
            // Handle timestamp - convert from nanoseconds to milliseconds
            let timestamp = 0;
            if (entity.self && entity.self.timestamp) {
                timestamp = Math.floor(parseInt(entity.self.timestamp) / 1000000); // Convert nanoseconds to milliseconds
            } else if (entity.created_at) {
                timestamp = new Date(entity.created_at).getTime();
            }
            
            const metric = {
                id: entity.id,
                timestamp: timestamp,
                host: (entity.traits && entity.traits.host) || (entity.self && entity.self.host),
                type: (entity.traits && entity.traits.metric_type) || (entity.self && entity.self.type),
                name: (entity.self && entity.self.name),
                value: parseFloat((entity.self && entity.self.value) || 0),
                unit: this.extractTagValue(entity.tags || [], 'unit')
            };
            
            // Add device/mount info for disk metrics
            const device = this.extractTagValue(entity.tags || [], 'device');
            if (device) metric.device = device;
            
            const mount = this.extractTagValue(entity.tags || [], 'mount');
            if (mount) metric.mount = mount;
            
            console.log('Transformed metric:', metric);
            return metric;
        }).filter(metric => metric.value !== undefined && !isNaN(metric.value));
    }

    // Extract value from tag
    extractTagValue(tags, prefix) {
        const tag = tags.find(t => t.startsWith(`${prefix}:`));
        return tag ? tag.split(':')[1] : null;
    }

    // Aggregate metrics for charts
    aggregateMetrics(metrics, interval = '1m') {
        // Group by timestamp intervals
        const intervalMs = this.parseInterval(interval);
        const groups = {};
        
        metrics.forEach(metric => {
            const bucket = Math.floor(metric.timestamp / intervalMs) * intervalMs;
            if (!groups[bucket]) {
                groups[bucket] = [];
            }
            groups[bucket].push(metric.value);
        });
        
        // Calculate averages
        const result = [];
        Object.keys(groups).sort().forEach(timestamp => {
            const values = groups[timestamp];
            const avg = values.reduce((a, b) => a + b, 0) / values.length;
            result.push({
                timestamp: parseInt(timestamp),
                value: avg,
                min: Math.min(...values),
                max: Math.max(...values),
                count: values.length
            });
        });
        
        return result;
    }

    // Parse interval string to milliseconds
    parseInterval(interval) {
        const units = {
            's': 1000,
            'm': 60 * 1000,
            'h': 60 * 60 * 1000,
            'd': 24 * 60 * 60 * 1000
        };
        
        const match = interval.match(/^(\d+)([smhd])$/);
        if (!match) return 60000; // Default 1 minute
        
        return parseInt(match[1]) * units[match[2]];
    }
}