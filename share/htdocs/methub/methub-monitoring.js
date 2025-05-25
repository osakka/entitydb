// MetHub Monitoring - Professional Monitoring Dashboard
window.methub = function() {
    return {
        // Authentication State
        isAuthenticated: false,
        currentUser: null,
        sessionToken: null,
        loginForm: {
            username: '',
            password: ''
        },
        loginError: '',
        
        // Dashboard State
        currentView: 'dashboard',
        sidebarCollapsed: false,
        loading: false,
        autoRefresh: true,
        refreshInterval: null,
        timeRange: '1h',
        selectedHost: null,
        hosts: [],
        widgets: [],
        agents: [],
        showAddWidget: false,
        editingWidget: null,
        gridColumns: 3,
        initialized: false,
        
        // Monitoring Stats
        systemStats: {
            totalAgents: 0,
            activeAgents: 0,
            criticalAlerts: 0,
            warnings: 0,
            systemHealth: 'unknown'
        },
        
        // API instances
        api: null,
        widgetRenderer: null,
        
        // Widget form
        widgetForm: {
            title: '',
            type: 'gauge',
            metricType: 'cpu',
            metricName: 'cpu_usage',
            span: 1,
            thresholds: {
                warning: 70,
                critical: 90
            }
        },
        
        // Theme management
        isDarkMode: false,
        
        toggleTheme() {
            this.isDarkMode = !this.isDarkMode;
            document.documentElement.setAttribute('data-theme', this.isDarkMode ? 'dark' : 'light');
            localStorage.setItem('methub-theme', this.isDarkMode ? 'dark' : 'light');
        },
        
        // Authentication methods
        async login() {
            this.loading = true;
            this.loginError = '';
            
            try {
                const response = await fetch('/api/v1/auth/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        username: this.loginForm.username,
                        password: this.loginForm.password
                    })
                });
                
                const data = await response.json();
                
                if (response.ok && data.token) {
                    this.sessionToken = data.token;
                    this.currentUser = data.user || { username: this.loginForm.username };
                    this.isAuthenticated = true;
                    
                    // Store authentication state
                    localStorage.setItem('methub-token', this.sessionToken);
                    localStorage.setItem('methub-user', JSON.stringify(this.currentUser));
                    
                    // Update API token
                    if (this.api) {
                        this.api.setToken(this.sessionToken);
                    }
                    
                    // Reset form
                    this.loginForm.username = '';
                    this.loginForm.password = '';
                    
                    // Initialize dashboard
                    await this.initializeDashboard();
                } else {
                    this.loginError = data.error || 'Login failed. Please check your credentials.';
                }
            } catch (error) {
                console.error('Login error:', error);
                this.loginError = 'Connection error. Please try again.';
            } finally {
                this.loading = false;
            }
        },
        
        async logout() {
            try {
                // Attempt to logout on server
                if (this.sessionToken) {
                    await fetch('/api/v1/auth/logout', {
                        method: 'POST',
                        headers: {
                            'Authorization': `Bearer ${this.sessionToken}`
                        }
                    });
                }
            } catch (error) {
                console.warn('Logout request failed:', error);
            }
            
            // Clear local state
            this.isAuthenticated = false;
            this.currentUser = null;
            this.sessionToken = null;
            
            // Clear stored authentication
            localStorage.removeItem('methub-token');
            localStorage.removeItem('methub-user');
            
            // Stop auto-refresh
            if (this.refreshInterval) {
                clearInterval(this.refreshInterval);
                this.refreshInterval = null;
            }
        },
        
        // Check stored authentication
        checkStoredAuth() {
            const token = localStorage.getItem('methub-token');
            const user = localStorage.getItem('methub-user');
            
            if (token && user) {
                try {
                    this.sessionToken = token;
                    this.currentUser = JSON.parse(user);
                    this.isAuthenticated = true;
                    
                    // Set API token if API is already initialized
                    if (this.api) {
                        this.api.setToken(token);
                    }
                    
                    return true;
                } catch (error) {
                    console.warn('Invalid stored auth data:', error);
                    this.clearStoredAuth();
                }
            }
            return false;
        },
        
        clearStoredAuth() {
            localStorage.removeItem('methub-token');
            localStorage.removeItem('methub-user');
        },
        
        // Initialization
        async init() {
            if (this.initialized) {
                console.log('⚠️ MetHub already initialized, skipping...');
                return;
            }
            
            console.log('🚀 Initializing MetHub Monitoring Dashboard...');
            
            // Check for stored authentication
            if (this.checkStoredAuth()) {
                await this.initializeDashboard();
            }
            
            // Load theme preference
            const savedTheme = localStorage.getItem('methub-theme');
            if (savedTheme === 'dark') {
                this.isDarkMode = true;
                document.documentElement.setAttribute('data-theme', 'dark');
            }
            
            this.initialized = true;
        },
        
        async initializeDashboard() {
            try {
                // Initialize API
                this.api = new MetHubAPI();
                this.widgetRenderer = new MetHubWidgets();
                
                // Set token if we have one
                if (this.sessionToken) {
                    this.api.setToken(this.sessionToken);
                }
                
                // Load data
                await Promise.all([
                    this.loadAgents(),
                    this.loadDefaultWidgets(),
                    this.updateSystemStats()
                ]);
                
                // Start auto-refresh
                if (this.autoRefresh) {
                    this.startAutoRefresh();
                }
                
                console.log('✅ MetHub dashboard initialized successfully');
            } catch (error) {
                console.error('Failed to initialize dashboard:', error);
            }
        },
        
        // Navigation
        setView(view) {
            this.currentView = view;
            console.log(`📱 Switched to view: ${view}`);
        },
        
        toggleSidebar() {
            this.sidebarCollapsed = !this.sidebarCollapsed;
        },
        
        // Data loading
        async loadAgents() {
            try {
                if (!this.api) {
                    console.error('API not initialized, cannot load agents');
                    this.agents = [];
                    return;
                }

                // Load real agents from EntityDB
                const agents = await this.api.request('GET', '/api/v1/entities/list?tags=type:agent&matchAll=true');
                
                if (agents && agents.length > 0) {
                    this.agents = await Promise.all(agents.map(async (entity) => {
                        const agent = await this.transformEntityToAgent(entity);
                        return agent;
                    }));
                    console.log(`🖥️ Loaded ${this.agents.length} real agents from EntityDB`);
                } else {
                    // Discover hosts from metrics - this is the real data source
                    console.log('No agent entities found, discovering hosts from metrics...');
                    await this.loadHostsFromMetrics();
                }
            } catch (error) {
                console.error('Failed to load agents from EntityDB:', error);
                this.agents = [];
            }
        },

        async loadHostsFromMetrics() {
            try {
                const hosts = await this.api.getHosts();
                console.log('Found hosts/instances:', hosts);
                
                this.agents = await Promise.all(hosts.map(async (hostname) => {
                    // For now, get some sample metrics for this instance
                    const instanceMetrics = await this.getInstanceMetrics(hostname);
                    
                    return {
                        id: hostname,
                        name: hostname.charAt(0).toUpperCase() + hostname.slice(1) + ' Instance',
                        hostname: hostname,
                        status: this.determineInstanceStatus(instanceMetrics),
                        lastSeen: new Date(), // Use current time since metrics are recent
                        metrics: {
                            activeUsers: instanceMetrics.activeUsers || 0,
                            productivity: instanceMetrics.productivity || 0,
                            tasksCompleted: instanceMetrics.tasksCompleted || 0
                        }
                    };
                }));
                console.log(`🖥️ Discovered ${this.agents.length} instances from metrics`);
            } catch (error) {
                console.error('Failed to load instances from metrics:', error);
                this.agents = [];
            }
        },

        async getInstanceMetrics(instance) {
            try {
                // Get real metrics for this instance
                const metrics = await this.api.request('GET', `/api/v1/entities/list?tags=type:metric,metric:instance:${instance}`);
                
                const result = {};
                (metrics.entities || metrics || []).forEach(entity => {
                    if (entity.tags) {
                        for (const tag of entity.tags) {
                            if (tag.startsWith('metric:value:')) {
                                const valueMatch = tag.match(/metric:value:(\d+\.?\d*):(\w+)/);
                                if (valueMatch) {
                                    const value = parseFloat(valueMatch[1]);
                                    const unit = valueMatch[2];
                                    
                                    // Extract metric name
                                    const nameTag = entity.tags.find(t => t.startsWith('metric:name:'));
                                    if (nameTag) {
                                        const metricName = nameTag.substring(12);
                                        if (metricName.includes('active_users')) result.activeUsers = value;
                                        if (metricName.includes('productivity')) result.productivity = value;
                                        if (metricName.includes('tasks_completed')) result.tasksCompleted = value;
                                    }
                                }
                            }
                        }
                    }
                });
                
                return result;
            } catch (error) {
                console.error('Failed to get instance metrics:', error);
                return {};
            }
        },

        determineInstanceStatus(metrics) {
            if (metrics.productivity) {
                if (metrics.productivity >= 90) return 'online';
                if (metrics.productivity >= 70) return 'warning';
                return 'critical';
            }
            return 'online';
        },

        async transformEntityToAgent(entity) {
            // Extract agent info from EntityDB entity
            const hostname = this.extractTagValue(entity.tags, 'hostname') || 
                            this.extractTagValue(entity.tags, 'host') || 
                            entity.id;
            
            const status = this.extractTagValue(entity.tags, 'status') || 'unknown';
            const lastSeen = entity.updated_at ? new Date(entity.updated_at) : new Date();
            
            // Get latest metrics for this agent
            let metrics = { cpu: 0, memory: 0, disk: 0 };
            try {
                const cpuMetric = await this.api.getLatestMetric(hostname, 'system', 'cpu_usage');
                const memoryMetric = await this.api.getLatestMetric(hostname, 'system', 'memory_usage');
                const diskMetric = await this.api.getLatestMetric(hostname, 'system', 'disk_usage');
                
                metrics = {
                    cpu: cpuMetric ? cpuMetric.value : 0,
                    memory: memoryMetric ? memoryMetric.value : 0,
                    disk: diskMetric ? diskMetric.value : 0
                };
            } catch (error) {
                console.warn(`Failed to load metrics for agent ${hostname}:`, error);
            }
            
            return {
                id: entity.id,
                name: this.extractTagValue(entity.tags, 'name') || hostname,
                hostname: hostname,
                status: this.determineAgentStatus(metrics.cpu, metrics.memory, metrics.disk, status),
                lastSeen: lastSeen,
                metrics: metrics
            };
        },


        extractTagValue(tags, key) {
            if (!tags) return null;
            for (const tag of tags) {
                if (tag.startsWith(`${key}:`)) {
                    return tag.substring(key.length + 1);
                }
            }
            return null;
        },

        determineAgentStatus(cpu, memory, disk, existingStatus = null) {
            if (existingStatus && ['offline', 'error'].includes(existingStatus)) {
                return existingStatus;
            }
            
            const cpuValue = typeof cpu === 'object' ? cpu.value : cpu;
            const memoryValue = typeof memory === 'object' ? memory.value : memory;
            const diskValue = typeof disk === 'object' ? disk.value : disk;
            
            if (cpuValue >= 95 || memoryValue >= 95 || diskValue >= 95) {
                return 'critical';
            } else if (cpuValue >= 80 || memoryValue >= 85 || diskValue >= 90) {
                return 'warning';
            } else {
                return 'online';
            }
        },

        getLatestTimestamp(...metrics) {
            let latest = new Date(0);
            for (const metric of metrics) {
                if (metric && metric.timestamp) {
                    const ts = new Date(metric.timestamp);
                    if (ts > latest) latest = ts;
                }
            }
            return latest.getTime() === 0 ? new Date() : latest;
        },
        
        async loadDefaultWidgets() {
            if (this.widgets.length === 0) {
                this.widgets = [
                    {
                        id: 'system-health',
                        title: 'System Health',
                        type: 'gauge',
                        metricType: 'health',
                        span: 1,
                        thresholds: { warning: 70, critical: 90 },
                        data: null,
                        loading: false
                    },
                    {
                        id: 'active-agents',
                        title: 'Active Agents',
                        type: 'single',
                        metricType: 'count',
                        metricName: 'agent_count',
                        span: 1,
                        data: null,
                        loading: false
                    },
                    {
                        id: 'cpu-usage',
                        title: 'CPU Usage',
                        type: 'line',
                        metricType: 'cpu',
                        span: 2,
                        thresholds: { warning: 70, critical: 90 },
                        data: null,
                        loading: false
                    },
                    {
                        id: 'memory-usage',
                        title: 'Memory Usage',
                        type: 'line',
                        metricType: 'memory',
                        span: 2,
                        thresholds: { warning: 80, critical: 95 },
                        data: null,
                        loading: false
                    }
                ];
                console.log(`📊 Loaded ${this.widgets.length} default widgets`);
            }
            
            await this.refreshWidgets();
        },
        
        async updateSystemStats() {
            const activeAgents = this.agents.filter(a => a.status === 'online').length;
            const criticalAgents = this.agents.filter(a => a.status === 'critical').length;
            const warningAgents = this.agents.filter(a => a.status === 'warning').length;
            
            this.systemStats = {
                totalAgents: this.agents.length,
                activeAgents: activeAgents,
                criticalAlerts: criticalAgents,
                warnings: warningAgents,
                systemHealth: criticalAgents > 0 ? 'critical' : 
                             warningAgents > 0 ? 'warning' : 'ok'
            };
        },
        
        // Auto-refresh
        startAutoRefresh() {
            if (this.refreshInterval) {
                clearInterval(this.refreshInterval);
            }
            
            this.refreshInterval = setInterval(async () => {
                if (this.autoRefresh && this.isAuthenticated) {
                    await this.refreshData();
                }
            }, 30000); // Refresh every 30 seconds
            
            console.log('🔄 Auto-refresh started');
        },
        
        stopAutoRefresh() {
            if (this.refreshInterval) {
                clearInterval(this.refreshInterval);
                this.refreshInterval = null;
                console.log('⏹️ Auto-refresh stopped');
            }
        },
        
        async refreshData() {
            await Promise.all([
                this.refreshWidgets(),
                this.updateSystemStats()
            ]);
        },
        
        async refreshWidgets() {
            for (let widget of this.widgets) {
                widget.loading = true;
                try {
                    widget.data = await this.fetchWidgetData(widget);
                } catch (error) {
                    console.error(`Failed to refresh widget ${widget.title}:`, error);
                    widget.data = null;
                } finally {
                    widget.loading = false;
                }
            }
        },
        
        async fetchWidgetData(widget) {
            // Mock data generation - replace with actual API calls
            const mockData = this.generateMockData(widget);
            
            // Simulate API delay
            await new Promise(resolve => setTimeout(resolve, Math.random() * 300 + 100));
            
            return mockData;
        },
        
        generateMockData(widget) {
            const now = Date.now();
            const timeRangeMs = this.parseTimeRange(this.timeRange);
            const points = Math.min(50, Math.max(10, Math.floor(timeRangeMs / 60000)));
            
            switch (widget.type) {
                case 'gauge':
                    const value = Math.random() * 100;
                    return {
                        value: value,
                        unit: '%',
                        status: this.getStatusFromValue(value, widget.thresholds)
                    };
                    
                case 'single':
                    if (widget.metricName === 'agent_count') {
                        return {
                            value: this.agents.length,
                            unit: 'agents',
                            status: 'ok'
                        };
                    }
                    return {
                        value: Math.floor(Math.random() * 1000),
                        unit: 'count',
                        status: 'ok'
                    };
                    
                case 'line':
                    const data = [];
                    for (let i = points; i >= 0; i--) {
                        const timestamp = now - (i * 60000);
                        const value = Math.random() * 100;
                        data.push({
                            timestamp,
                            value,
                            status: this.getStatusFromValue(value, widget.thresholds)
                        });
                    }
                    return { series: data };
                    
                default:
                    return null;
            }
        },
        
        getStatusFromValue(value, thresholds) {
            if (!thresholds) return 'ok';
            
            if (value >= thresholds.critical) return 'critical';
            if (value >= thresholds.warning) return 'warning';
            return 'ok';
        },
        
        parseTimeRange(range) {
            const multipliers = {
                'm': 60 * 1000,
                'h': 60 * 60 * 1000,
                'd': 24 * 60 * 60 * 1000
            };
            
            const match = range.match(/^(\d+)([mhd])$/);
            if (match) {
                return parseInt(match[1]) * multipliers[match[2]];
            }
            return 60 * 60 * 1000; // Default to 1 hour
        },
        
        // Widget management
        addWidget() {
            this.showAddWidget = true;
            this.editingWidget = null;
            this.resetWidgetForm();
        },
        
        editWidget(widget) {
            this.editingWidget = widget;
            this.widgetForm = { ...widget };
            this.showAddWidget = true;
        },
        
        async saveWidget() {
            const widget = {
                id: this.editingWidget ? this.editingWidget.id : `widget-${Date.now()}`,
                ...this.widgetForm,
                data: null,
                loading: false
            };
            
            if (this.editingWidget) {
                const index = this.widgets.findIndex(w => w.id === this.editingWidget.id);
                this.widgets[index] = widget;
            } else {
                this.widgets.push(widget);
            }
            
            this.showAddWidget = false;
            this.editingWidget = null;
            
            // Refresh the new/updated widget
            widget.data = await this.fetchWidgetData(widget);
        },
        
        deleteWidget(widget) {
            if (confirm(`Delete widget "${widget.title}"?`)) {
                this.widgets = this.widgets.filter(w => w.id !== widget.id);
            }
        },
        
        resetWidgetForm() {
            this.widgetForm = {
                title: '',
                type: 'gauge',
                metricType: 'cpu',
                metricName: 'cpu_usage',
                span: 1,
                thresholds: {
                    warning: 70,
                    critical: 90
                }
            };
        },
        
        // Utility methods
        formatBytes(bytes) {
            const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
            if (bytes === 0) return '0 B';
            const i = Math.floor(Math.log(bytes) / Math.log(1024));
            return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
        },
        
        formatDuration(ms) {
            const seconds = Math.floor(ms / 1000);
            const minutes = Math.floor(seconds / 60);
            const hours = Math.floor(minutes / 60);
            const days = Math.floor(hours / 24);
            
            if (days > 0) return `${days}d ${hours % 24}h`;
            if (hours > 0) return `${hours}h ${minutes % 60}m`;
            if (minutes > 0) return `${minutes}m ${seconds % 60}s`;
            return `${seconds}s`;
        },
        
        timeAgo(date) {
            const now = new Date();
            const diff = now - date;
            
            if (diff < 60000) return 'Just now';
            if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
            if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
            return `${Math.floor(diff / 86400000)}d ago`;
        }
    };
};

// Theme store for Alpine.js
document.addEventListener('alpine:init', () => {
    Alpine.store('theme', {
        dark: localStorage.getItem('methub-theme') === 'dark',
        
        toggle() {
            this.dark = !this.dark;
            localStorage.setItem('methub-theme', this.dark ? 'dark' : 'light');
            document.documentElement.setAttribute('data-theme', this.dark ? 'dark' : 'light');
        }
    });
});