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
        const options = {
            method,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': this.token ? `Bearer ${this.token}` : ''
            }
        };

        if (data) {
            options.body = JSON.stringify(data);
        }

        try {
            const response = await fetch(`${this.baseUrl}${endpoint}`, options);
            
            if (!response.ok) {
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
        const result = await this.request('POST', '/api/v1/auth/login', {
            username,
            password
        });
        
        if (result.token) {
            this.setToken(result.token);
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
        
        // Query entities
        const result = await this.request('POST', `/api/v1/hubs/${this.hub}/entities/query`, query);
        
        return this.transformMetrics(result.entities || []);
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
        const result = await this.request('POST', `/api/v1/hubs/${this.hub}/entities/query`, {
            hub: this.hub,
            tags: ['type:metric'],
            limit: 1000
        });
        
        const hosts = new Set();
        (result.entities || []).forEach(entity => {
            if (entity.traits && entity.traits.host) {
                hosts.add(entity.traits.host);
            }
        });
        
        return Array.from(hosts).sort();
    }

    // Transform entities to metric format
    transformMetrics(entities) {
        return entities.map(entity => {
            const metric = {
                id: entity.id,
                timestamp: parseInt(entity.self.timestamp || entity.created_at),
                host: entity.traits.host || entity.self.host,
                type: entity.traits.metric_type,
                name: entity.self.name,
                value: parseFloat(entity.self.value),
                unit: this.extractTagValue(entity.tags, 'unit')
            };
            
            // Add device/mount info for disk metrics
            const device = this.extractTagValue(entity.tags, 'device');
            if (device) metric.device = device;
            
            const mount = this.extractTagValue(entity.tags, 'mount');
            if (mount) metric.mount = mount;
            
            return metric;
        });
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