/**
 * EntityDB Nexus Widget Framework
 * Revolutionary UI components for temporal data visualization
 * v2.34.5
 */

class NexusWidgetFramework {
    constructor() {
        this.widgets = new Map();
        this.errorBoundaries = new Map();
        this.performanceMonitor = new PerformanceMonitor();
        this.initializeFramework();
    }

    initializeFramework() {
        console.log('🚀 Initializing Nexus Widget Framework...');
        this.setupErrorBoundaries();
        this.setupPerformanceMonitoring();
        this.registerCoreWidgets();
    }

    setupErrorBoundaries() {
        // Global error boundary for all widgets
        window.addEventListener('error', (event) => {
            this.handleWidgetError(event.error, 'global', event);
        });

        window.addEventListener('unhandledrejection', (event) => {
            this.handleWidgetError(event.reason, 'promise', event);
        });
    }

    setupPerformanceMonitoring() {
        // Monitor widget render performance
        if (window.PerformanceObserver) {
            const observer = new PerformanceObserver((list) => {
                for (const entry of list.getEntries()) {
                    if (entry.name.startsWith('nexus-widget-')) {
                        this.performanceMonitor.recordMetric(entry);
                    }
                }
            });
            observer.observe({ entryTypes: ['measure'] });
        }
    }

    registerWidget(name, widgetClass) {
        try {
            this.widgets.set(name, widgetClass);
            console.log(`✅ Widget registered: ${name}`);
        } catch (error) {
            this.handleWidgetError(error, 'registration', { widgetName: name });
        }
    }

    createWidget(name, container, config = {}) {
        performance.mark(`nexus-widget-${name}-start`);
        
        try {
            // Validate container exists
            if (!container) {
                throw new Error(`Container for widget '${name}' is null or undefined`);
            }
            
            // If container is a string, try to find the element
            if (typeof container === 'string') {
                const element = document.getElementById(container) || document.querySelector(container);
                if (!element) {
                    throw new Error(`Container '${container}' not found in DOM`);
                }
                container = element;
            }
            
            // Validate container is a valid DOM element
            if (!container.nodeType || container.nodeType !== Node.ELEMENT_NODE) {
                throw new Error(`Invalid container for widget '${name}' - must be a DOM element`);
            }

            const WidgetClass = this.widgets.get(name);
            if (!WidgetClass) {
                throw new Error(`Widget '${name}' not found`);
            }

            const widget = new WidgetClass(container, config);
            
            // Add widget validation before render
            if (!widget.container) {
                throw new Error(`Widget '${name}' failed to initialize - container is null`);
            }
            
            widget.render();
            
            performance.mark(`nexus-widget-${name}-end`);
            performance.measure(`nexus-widget-${name}`, `nexus-widget-${name}-start`, `nexus-widget-${name}-end`);
            
            console.log(`🎨 Widget created: ${name}`);
            return widget;
        } catch (error) {
            this.handleWidgetError(error, 'creation', { widgetName: name, config });
            return null;
        }
    }

    handleWidgetError(error, context, details = {}) {
        const errorInfo = {
            timestamp: new Date().toISOString(),
            error: error.message || error,
            stack: error.stack,
            context,
            details,
            userAgent: navigator.userAgent,
            url: window.location.href
        };

        console.error('🚨 Widget Framework Error:', errorInfo);
        
        // Emit custom event for debug console
        window.dispatchEvent(new CustomEvent('nexus-widget-error', {
            detail: errorInfo
        }));

        // Attempt graceful degradation
        this.attemptRecovery(context, details);
    }

    attemptRecovery(context, details) {
        switch (context) {
            case 'creation':
                // Show fallback widget
                this.createFallbackWidget(details.widgetName);
                break;
            case 'render':
                // Try re-render with safe defaults
                setTimeout(() => this.safeRerender(details.widget), 1000);
                break;
        }
    }

    createFallbackWidget(widgetName) {
        const fallback = document.createElement('div');
        fallback.className = 'nexus-widget-error';
        fallback.innerHTML = `
            <div style="padding: 20px; text-align: center; color: var(--nexus-error);">
                <div style="font-size: 24px;">⚠️</div>
                <div style="margin-top: 8px;">Widget '${widgetName}' failed to load</div>
                <button onclick="location.reload()" style="margin-top: 12px; padding: 8px 16px; background: var(--nexus-accent); color: var(--nexus-bg); border: none; border-radius: 4px; cursor: pointer;">
                    Reload Page
                </button>
            </div>
        `;
        return fallback;
    }

    registerCoreWidgets() {
        this.registerWidget('temporal-explorer', TemporalExplorerWidget);
        this.registerWidget('entity-network', EntityNetworkWidget);
        this.registerWidget('storage-metrics', StorageMetricsWidget);
        this.registerWidget('live-monitor', LiveMonitorWidget);
        this.registerWidget('command-palette', CommandPaletteWidget);
    }
}

/**
 * Base Widget Class
 * All Nexus widgets inherit from this
 */
class NexusWidget {
    constructor(container, config = {}) {
        // Validate container before proceeding
        if (!container) {
            throw new Error('Widget container is required');
        }
        
        this.container = container;
        this.config = { ...this.getDefaultConfig(), ...config };
        this.state = {};
        this.isDestroyed = false;
        this.errorBoundary = this.setupErrorBoundary();
        
        // Add container validation marker
        this.container.setAttribute('data-nexus-widget', this.constructor.name);
    }

    getDefaultConfig() {
        return {
            theme: 'nexus-dark',
            animations: true,
            debugMode: false
        };
    }

    setupErrorBoundary() {
        return (fn) => {
            try {
                return fn();
            } catch (error) {
                this.handleError(error);
                return null;
            }
        };
    }

    handleError(error) {
        console.error(`Widget Error (${this.constructor.name}):`, error);
        this.renderErrorState(error);
    }

    renderErrorState(error) {
        if (this.container) {
            this.container.innerHTML = `
                <div class="nexus-widget-error">
                    <div style="color: var(--nexus-error); text-align: center; padding: 20px;">
                        <div>⚠️ Widget Error</div>
                        <div style="font-size: 12px; margin-top: 8px; opacity: 0.7;">
                            ${error.message}
                        </div>
                    </div>
                </div>
            `;
        }
    }
    
    // Safe DOM query helper
    querySelector(selector) {
        try {
            if (!this.container) {
                console.warn(`Widget ${this.constructor.name}: container is null when querying for ${selector}`);
                return null;
            }
            return this.container.querySelector(selector);
        } catch (error) {
            console.error(`Widget ${this.constructor.name}: failed to query ${selector}:`, error);
            return null;
        }
    }
    
    // Safe DOM query all helper
    querySelectorAll(selector) {
        try {
            if (!this.container) {
                console.warn(`Widget ${this.constructor.name}: container is null when querying for ${selector}`);
                return [];
            }
            return Array.from(this.container.querySelectorAll(selector));
        } catch (error) {
            console.error(`Widget ${this.constructor.name}: failed to query all ${selector}:`, error);
            return [];
        }
    }

    setState(newState) {
        this.state = { ...this.state, ...newState };
        this.scheduleRender();
    }

    scheduleRender() {
        if (this.renderTimeout) return;
        this.renderTimeout = requestAnimationFrame(() => {
            this.renderTimeout = null;
            if (!this.isDestroyed) {
                this.errorBoundary(() => this.render());
            }
        });
    }

    render() {
        // Override in subclasses
    }

    destroy() {
        this.isDestroyed = true;
        if (this.renderTimeout) {
            cancelAnimationFrame(this.renderTimeout);
        }
        this.container.innerHTML = '';
    }
}

/**
 * Temporal Explorer Widget
 * Visualizes entity evolution through time
 */
class TemporalExplorerWidget extends NexusWidget {
    constructor(container, config) {
        super(container, config);
        this.timeline = null;
        this.selectedEntity = null;
    }

    getDefaultConfig() {
        return {
            ...super.getDefaultConfig(),
            timeRange: '24h',
            entityTypes: ['all'],
            animationSpeed: 1000
        };
    }

    async render() {
        this.container.innerHTML = `
            <div class="temporal-explorer">
                <div class="temporal-header">
                    <h3>⟲ Temporal Explorer</h3>
                    <div class="temporal-controls">
                        <select class="nexus-select" id="time-range">
                            <option value="1h">Last Hour</option>
                            <option value="24h" selected>Last 24 Hours</option>
                            <option value="7d">Last Week</option>
                            <option value="30d">Last Month</option>
                        </select>
                    </div>
                </div>
                <div class="temporal-timeline" id="timeline-container">
                    <div class="loading-spinner">
                        <div class="nexus-spinner"></div>
                        <div>Loading temporal data...</div>
                    </div>
                </div>
                <div class="temporal-details" id="details-container">
                    <div class="no-selection">
                        Select an entity on the timeline to view its evolution
                    </div>
                </div>
            </div>
        `;

        await this.loadTemporalData();
        this.setupEventHandlers();
    }

    async loadTemporalData() {
        try {
            // Get JWT token for authentication
            const token = localStorage.getItem('entitydb_token');
            if (!token) {
                throw new Error('Authentication required');
            }

            // Load entities with temporal data
            const response = await fetch('/api/v1/entities/list?include_timestamps=true', {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const entities = await response.json();
            
            // Transform entities for temporal visualization
            const temporalEvents = [];
            
            entities.forEach(entity => {
                // Parse creation time
                const createdAt = new Date(entity.created_at / 1000000); // Convert nanoseconds to milliseconds
                temporalEvents.push({
                    timestamp: createdAt,
                    type: 'create',
                    entityId: entity.id,
                    entityType: this.getEntityType(entity),
                    entityName: this.getEntityName(entity)
                });
                
                // Parse tag timestamps for modifications
                if (entity.tags) {
                    entity.tags.forEach(tag => {
                        if (tag.includes('|')) {
                            const [timestampStr, tagValue] = tag.split('|', 2);
                            const timestamp = new Date(parseInt(timestampStr) / 1000000);
                            
                            if (timestamp > createdAt) {
                                temporalEvents.push({
                                    timestamp: timestamp,
                                    type: 'modify',
                                    entityId: entity.id,
                                    entityType: this.getEntityType(entity),
                                    entityName: this.getEntityName(entity),
                                    change: tagValue
                                });
                            }
                        }
                    });
                }
            });
            
            // Sort events by timestamp
            temporalEvents.sort((a, b) => a.timestamp - b.timestamp);
            
            this.renderTimeline(temporalEvents);
        } catch (error) {
            this.handleError(error);
        }
    }
    
    getEntityType(entity) {
        const typeTag = entity.tags?.find(tag => {
            const cleanTag = tag.includes('|') ? tag.split('|')[1] : tag;
            return cleanTag.startsWith('type:');
        });
        return typeTag ? typeTag.split(':')[1] : 'unknown';
    }
    
    getEntityName(entity) {
        const nameTag = entity.tags?.find(tag => {
            const cleanTag = tag.includes('|') ? tag.split('|')[1] : tag;
            return cleanTag.startsWith('name:');
        });
        return nameTag ? nameTag.split(':')[1] : entity.id.substring(0, 8);
    }

    renderTimeline(temporalEvents) {
        const container = this.querySelector('#timeline-container');
        
        if (!container) {
            console.warn('Timeline container not found');
            return;
        }
        
        if (!temporalEvents || temporalEvents.length === 0) {
            container.innerHTML = `
                <div style="text-align: center; padding: 40px; color: var(--nexus-text-dim);">
                    <div style="font-size: 48px; margin-bottom: 16px;">⟲</div>
                    <div>No temporal events found</div>
                    <div style="font-size: 12px; margin-top: 8px;">EntityDB temporal data will appear here</div>
                </div>
            `;
            return;
        }
        
        // Group events by time periods
        const now = new Date();
        const last24h = new Date(now - 24 * 60 * 60 * 1000);
        const last7d = new Date(now - 7 * 24 * 60 * 60 * 1000);
        
        const recentEvents = temporalEvents.filter(e => e.timestamp > last24h);
        const weekEvents = temporalEvents.filter(e => e.timestamp > last7d && e.timestamp <= last24h);
        const olderEvents = temporalEvents.filter(e => e.timestamp <= last7d);
        
        container.innerHTML = `
            <div class="timeline-visualization">
                <div class="timeline-stats" style="display: flex; gap: 20px; margin-bottom: 20px; padding: 16px; background: rgba(0, 217, 255, 0.1); border-radius: 8px;">
                    <div style="text-align: center;">
                        <div style="font-size: 24px; font-weight: bold; color: var(--nexus-accent);">${recentEvents.length}</div>
                        <div style="font-size: 12px; color: var(--nexus-text-dim);">Last 24h</div>
                    </div>
                    <div style="text-align: center;">
                        <div style="font-size: 24px; font-weight: bold; color: var(--nexus-temporal);">${weekEvents.length}</div>
                        <div style="font-size: 12px; color: var(--nexus-text-dim);">Past Week</div>
                    </div>
                    <div style="text-align: center;">
                        <div style="font-size: 24px; font-weight: bold; color: var(--nexus-success);">${temporalEvents.length}</div>
                        <div style="font-size: 12px; color: var(--nexus-text-dim);">Total Events</div>
                    </div>
                </div>
                
                <div class="timeline-events" style="max-height: 300px; overflow-y: auto;">
                    ${recentEvents.slice(-10).reverse().map(event => `
                        <div class="timeline-event" style="padding: 12px; margin-bottom: 8px; background: var(--nexus-surface-hover); border-radius: 6px; border-left: 3px solid ${event.type === 'create' ? 'var(--nexus-accent)' : 'var(--nexus-temporal)'};">
                            <div style="display: flex; justify-content: space-between; align-items: center;">
                                <div>
                                    <span style="color: ${event.type === 'create' ? 'var(--nexus-accent)' : 'var(--nexus-temporal)'}; font-weight: 600;">
                                        ${event.type === 'create' ? '+ Created' : '~ Modified'}
                                    </span>
                                    <span style="color: var(--nexus-text); margin-left: 8px;">
                                        ${event.entityName} (${event.entityType})
                                    </span>
                                </div>
                                <div style="font-size: 11px; color: var(--nexus-text-dim);">
                                    ${event.timestamp.toLocaleTimeString()}
                                </div>
                            </div>
                            ${event.change ? `<div style="font-size: 11px; color: var(--nexus-text-dim); margin-top: 4px;">Change: ${event.change}</div>` : ''}
                        </div>
                    `).join('')}
                </div>
                
                <div class="timeline-legend" style="display: flex; gap: 20px; margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--nexus-border);">
                    <div class="legend-item" style="display: flex; align-items: center; gap: 8px;">
                        <span class="legend-color" style="width: 12px; height: 12px; background: var(--nexus-accent); border-radius: 2px;"></span>
                        <span style="font-size: 12px; color: var(--nexus-text-dim);">Entity Created</span>
                    </div>
                    <div class="legend-item" style="display: flex; align-items: center; gap: 8px;">
                        <span class="legend-color" style="width: 12px; height: 12px; background: var(--nexus-temporal); border-radius: 2px;"></span>
                        <span style="font-size: 12px; color: var(--nexus-text-dim);">Entity Modified</span>
                    </div>
                </div>
            </div>
        `;
    }

    setupEventHandlers() {
        const timeRangeSelect = this.querySelector('#time-range');
        if (timeRangeSelect) {
            timeRangeSelect.addEventListener('change', (e) => {
                this.config.timeRange = e.target.value;
                this.loadTemporalData();
            });
        }
    }
}

/**
 * Entity Network Widget
 * Visualizes relationships between entities
 */
class EntityNetworkWidget extends NexusWidget {
    constructor(container, config) {
        super(container, config);
        this.network = null;
        this.nodes = [];
        this.edges = [];
    }

    async render() {
        this.container.innerHTML = `
            <div class="entity-network">
                <div class="network-header">
                    <h3>◊ Entity Relationships</h3>
                    <div class="network-controls">
                        <button class="nexus-btn-ghost" id="reset-view">Reset View</button>
                        <button class="nexus-btn-ghost" id="export-graph">Export</button>
                    </div>
                </div>
                <div class="network-container" id="network-container">
                    <svg id="network-svg" width="100%" height="400"></svg>
                </div>
                <div class="network-stats">
                    <div class="stat-item">
                        <span class="stat-label">Nodes:</span>
                        <span class="stat-value" id="node-count">0</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Edges:</span>
                        <span class="stat-value" id="edge-count">0</span>
                    </div>
                </div>
            </div>
        `;

        await this.loadNetworkData();
        this.renderNetwork();
    }

    async loadNetworkData() {
        try {
            // Get JWT token for authentication
            const token = localStorage.getItem('entitydb_token');
            if (!token) {
                throw new Error('Authentication required');
            }

            // Load entities
            const entitiesResponse = await fetch('/api/v1/entities/list', {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
            
            if (!entitiesResponse.ok) {
                throw new Error(`HTTP ${entitiesResponse.status}: ${entitiesResponse.statusText}`);
            }
            
            const data = await entitiesResponse.json();
            const entities = Array.isArray(data) ? data : (data.entities || []);
            
            // Transform to network format with proper tag parsing
            this.nodes = entities.map(entity => {
                // Parse temporal tags (format: timestamp|tag)
                const cleanTags = entity.tags.map(tag => {
                    if (tag.includes('|')) {
                        return tag.split('|')[1];
                    }
                    return tag;
                });
                
                const nameTag = cleanTags.find(tag => tag.startsWith('name:'));
                const typeTag = cleanTags.find(tag => tag.startsWith('type:'));
                
                return {
                    id: entity.id,
                    label: nameTag ? nameTag.split(':')[1] : entity.id.substring(0, 8),
                    type: typeTag ? typeTag.split(':')[1] : 'unknown',
                    tags: cleanTags
                };
            });

            // Load actual relationships if available
            try {
                const relationshipsResponse = await fetch('/api/v1/entity-relationships', {
                    headers: {
                        'Authorization': `Bearer ${token}`
                    }
                });
                
                if (relationshipsResponse.ok) {
                    const relationships = await relationshipsResponse.json();
                    this.edges = relationships.map(rel => ({
                        source: rel.source_entity_id,
                        target: rel.target_entity_id,
                        type: rel.relationship_type || 'related'
                    }));
                } else {
                    this.edges = [];
                }
            } catch (relError) {
                // Relationships endpoint might not be available
                this.edges = [];
            }
            
        } catch (error) {
            this.handleError(error);
            // Show empty state with error info
            this.nodes = [];
            this.edges = [];
        }
    }

    renderNetwork() {
        const svg = d3.select(this.container.querySelector('#network-svg'));
        const width = 800;
        const height = 400;

        // Store simulation as instance variable for drag functions
        this.simulation = d3.forceSimulation(this.nodes)
            .force('link', d3.forceLink(this.edges).id(d => d.id))
            .force('charge', d3.forceManyBody().strength(-300))
            .force('center', d3.forceCenter(width / 2, height / 2));

        // Clear previous SVG content
        svg.selectAll('*').remove();

        // Render nodes and edges
        const link = svg.append('g')
            .selectAll('line')
            .data(this.edges)
            .enter().append('line')
            .attr('stroke', '#2a3441')
            .attr('stroke-width', 2);

        const node = svg.append('g')
            .selectAll('circle')
            .data(this.nodes)
            .enter().append('circle')
            .attr('r', 8)
            .attr('fill', '#00d9ff')
            .call(d3.drag()
                .on('start', this.dragstarted.bind(this))
                .on('drag', this.dragged.bind(this))
                .on('end', this.dragended.bind(this)));

        this.simulation.on('tick', () => {
            link
                .attr('x1', d => d.source.x)
                .attr('y1', d => d.source.y)
                .attr('x2', d => d.target.x)
                .attr('y2', d => d.target.y);

            node
                .attr('cx', d => d.x)
                .attr('cy', d => d.y);
        });

        // Update stats using safe query helpers
        const nodeCountEl = this.querySelector('#node-count');
        const edgeCountEl = this.querySelector('#edge-count');
        
        if (nodeCountEl) nodeCountEl.textContent = this.nodes.length;
        if (edgeCountEl) edgeCountEl.textContent = this.edges.length;
    }

    dragstarted(event, d) {
        if (!event.active && this.simulation) this.simulation.alphaTarget(0.3).restart();
        d.fx = d.x;
        d.fy = d.y;
    }

    dragged(event, d) {
        d.fx = event.x;
        d.fy = event.y;
    }

    dragended(event, d) {
        if (!event.active && this.simulation) this.simulation.alphaTarget(0);
        d.fx = null;
        d.fy = null;
    }
}

/**
 * Storage Metrics Widget
 * Real-time storage performance monitoring
 */
class StorageMetricsWidget extends NexusWidget {
    constructor(container, config) {
        super(container, config);
        this.chart = null;
        this.metrics = [];
    }

    async render() {
        this.container.innerHTML = `
            <div class="storage-metrics">
                <div class="metrics-header">
                    <h3>▤ Storage Engine</h3>
                    <div class="metrics-controls">
                        <div class="metric-toggle active" data-metric="read">Read</div>
                        <div class="metric-toggle active" data-metric="write">Write</div>
                        <div class="metric-toggle active" data-metric="cache">Cache</div>
                    </div>
                </div>
                <div class="metrics-chart">
                    <canvas id="storage-chart" width="400" height="200"></canvas>
                </div>
                <div class="metrics-summary">
                    <div class="metric-card">
                        <div class="metric-value" id="read-ops">0</div>
                        <div class="metric-label">Read Ops/sec</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value" id="write-ops">0</div>
                        <div class="metric-label">Write Ops/sec</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value" id="cache-hit">0%</div>
                        <div class="metric-label">Cache Hit Rate</div>
                    </div>
                </div>
            </div>
        `;

        this.initChart();
        this.startMetricsCollection();
    }

    initChart() {
        const canvas = this.querySelector('#storage-chart');
        if (!canvas) {
            console.error('Storage chart canvas not found');
            return;
        }
        
        const ctx = canvas.getContext('2d');
        
        // Get computed CSS custom properties for proper dark theme colors
        const computedStyle = getComputedStyle(document.documentElement);
        const accentColor = '#00d9ff'; // Direct hex for chart compatibility
        const temporalColor = '#ff6b35';
        const textColor = '#e2e8f0';
        const textDimColor = '#94a3b8';
        const borderColor = '#2a3441';
        
        this.chart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [
                    {
                        label: 'Read Operations',
                        data: [],
                        borderColor: accentColor,
                        backgroundColor: 'rgba(0, 217, 255, 0.1)',
                        tension: 0.4,
                        fill: true
                    },
                    {
                        label: 'Write Operations', 
                        data: [],
                        borderColor: temporalColor,
                        backgroundColor: 'rgba(255, 107, 53, 0.1)',
                        tension: 0.4,
                        fill: true
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        labels: {
                            color: textColor,
                            usePointStyle: true
                        }
                    }
                },
                scales: {
                    x: {
                        ticks: { 
                            color: textDimColor,
                            maxTicksLimit: 10
                        },
                        grid: { 
                            color: borderColor,
                            drawBorder: false
                        }
                    },
                    y: {
                        ticks: { 
                            color: textDimColor,
                            beginAtZero: true
                        },
                        grid: { 
                            color: borderColor,
                            drawBorder: false
                        }
                    }
                }
            }
        });
    }

    startMetricsCollection() {
        setInterval(async () => {
            await this.collectMetrics();
        }, 1000);
    }

    async collectMetrics() {
        try {
            // Get JWT token from localStorage for authentication
            const token = localStorage.getItem('entitydb_token');
            if (!token) {
                throw new Error('Authentication required');
            }

            const response = await fetch('/api/v1/system/metrics', {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            
            const timestamp = new Date().toLocaleTimeString();
            
            // Extract real metrics from EntityDB system metrics
            const readOps = data.storage?.read_operations_total || 0;
            const writeOps = data.storage?.write_operations_total || 0;
            const cacheHitRate = data.storage?.cache_hit_rate || 0;
            
            // Calculate operations per second (rate)
            const readRate = data.storage?.read_ops_per_second || 0;
            const writeRate = data.storage?.write_ops_per_second || 0;

            // Update chart with rates
            this.chart.data.labels.push(timestamp);
            this.chart.data.datasets[0].data.push(readRate);
            this.chart.data.datasets[1].data.push(writeRate);

            // Keep only last 20 data points
            if (this.chart.data.labels.length > 20) {
                this.chart.data.labels.shift();
                this.chart.data.datasets[0].data.shift();
                this.chart.data.datasets[1].data.shift();
            }

            this.chart.update('none');

            // Update summary cards with real values - using safe query helpers
            const readOpsEl = this.querySelector('#read-ops');
            const writeOpsEl = this.querySelector('#write-ops');
            const cacheHitEl = this.querySelector('#cache-hit');
            
            if (readOpsEl) readOpsEl.textContent = readRate.toFixed(1);
            if (writeOpsEl) writeOpsEl.textContent = writeRate.toFixed(1);
            if (cacheHitEl) cacheHitEl.textContent = Math.round(cacheHitRate * 100) + '%';

        } catch (error) {
            this.handleError(error);
            
            // Show connection status in widgets instead of hiding the error - using safe query helpers
            const readOpsEl = this.querySelector('#read-ops');
            const writeOpsEl = this.querySelector('#write-ops');
            const cacheHitEl = this.querySelector('#cache-hit');
            
            if (readOpsEl) readOpsEl.textContent = 'N/A';
            if (writeOpsEl) writeOpsEl.textContent = 'N/A';
            if (cacheHitEl) cacheHitEl.textContent = 'N/A';
        }
    }
}

/**
 * Live Monitor Widget
 * Real-time system monitoring with alerts
 */
class LiveMonitorWidget extends NexusWidget {
    constructor(container, config) {
        super(container, config);
        this.alerts = [];
        this.metrics = {};
    }

    async render() {
        this.container.innerHTML = `
            <div class="live-monitor">
                <div class="monitor-header">
                    <h3>◉ Live System Monitor</h3>
                    <div class="monitor-status">
                        <span class="status-dot" style="background: var(--nexus-success);"></span>
                        All Systems Operational
                    </div>
                </div>
                <div class="monitor-alerts" id="alerts-container">
                    <div class="no-alerts">No active alerts</div>
                </div>
                <div class="monitor-metrics">
                    <div class="metric-row">
                        <span>Database Status:</span>
                        <span class="metric-value" style="color: var(--nexus-success);">HEALTHY</span>
                    </div>
                    <div class="metric-row">
                        <span>WAL Size:</span>
                        <span class="metric-value" id="wal-size">Loading...</span>
                    </div>
                    <div class="metric-row">
                        <span>Memory Usage:</span>
                        <span class="metric-value" id="memory-usage">Loading...</span>
                    </div>
                    <div class="metric-row">
                        <span>Active Sessions:</span>
                        <span class="metric-value" id="active-sessions">Loading...</span>
                    </div>
                </div>
            </div>
        `;

        this.startMonitoring();
    }

    startMonitoring() {
        setInterval(async () => {
            await this.checkSystemHealth();
        }, 5000);
    }

    async checkSystemHealth() {
        try {
            // Get JWT token for authentication
            const token = localStorage.getItem('entitydb_token');
            if (!token) {
                throw new Error('Authentication required');
            }

            const response = await fetch('/api/v1/system/metrics', {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            
            // Update metrics with real data
            const walSize = data.storage?.wal_size_mb || 0;
            const memoryUsage = Math.round((data.memory?.alloc || 0) / 1024 / 1024);
            const activeSessions = data.rbac?.active_sessions || 0;
            
            // Update metrics using safe query helpers
            const walSizeEl = this.querySelector('#wal-size');
            const memoryUsageEl = this.querySelector('#memory-usage');
            const activeSessionsEl = this.querySelector('#active-sessions');
            
            if (walSizeEl) walSizeEl.textContent = `${walSize.toFixed(1)} MB`;
            if (memoryUsageEl) memoryUsageEl.textContent = `${memoryUsage} MB`;
            if (activeSessionsEl) activeSessionsEl.textContent = activeSessions;
            
            // Real health checks based on EntityDB thresholds
            const alertsFound = [];
            
            if (walSize > 100) {
                alertsFound.push({ level: 'CRITICAL', message: `WAL file size is ${walSize.toFixed(1)} MB (>100MB)` });
            } else if (walSize > 50) {
                alertsFound.push({ level: 'WARNING', message: `WAL file size is ${walSize.toFixed(1)} MB (>50MB)` });
            }
            
            if (memoryUsage > 800) {
                alertsFound.push({ level: 'CRITICAL', message: `High memory usage: ${memoryUsage} MB (>800MB)` });
            } else if (memoryUsage > 500) {
                alertsFound.push({ level: 'WARNING', message: `Elevated memory usage: ${memoryUsage} MB (>500MB)` });
            }
            
            if (data.storage?.corruption_detected) {
                alertsFound.push({ level: 'CRITICAL', message: 'Storage corruption detected' });
            }
            
            if (data.performance?.slow_queries > 10) {
                alertsFound.push({ level: 'WARNING', message: `${data.performance.slow_queries} slow queries detected` });
            }
            
            // Add new alerts
            alertsFound.forEach(alert => {
                this.addAlert(alert.level, alert.message);
            });
            
            // Update system status
            const statusElement = this.container.querySelector('.monitor-status span:last-child');
            if (alertsFound.some(a => a.level === 'CRITICAL')) {
                statusElement.textContent = 'Critical Issues Detected';
                statusElement.style.color = 'var(--nexus-error)';
            } else if (alertsFound.some(a => a.level === 'WARNING')) {
                statusElement.textContent = 'Warnings Present';
                statusElement.style.color = 'var(--nexus-warning)';
            } else {
                statusElement.textContent = 'All Systems Operational';
                statusElement.style.color = 'var(--nexus-success)';
            }
            
        } catch (error) {
            this.handleError(error);
            
            // Show connection error state using safe query helpers
            const walSizeEl = this.querySelector('#wal-size');
            const memoryUsageEl = this.querySelector('#memory-usage');
            const activeSessionsEl = this.querySelector('#active-sessions');
            
            if (walSizeEl) walSizeEl.textContent = 'N/A';
            if (memoryUsageEl) memoryUsageEl.textContent = 'N/A';
            if (activeSessionsEl) activeSessionsEl.textContent = 'N/A';
            
            this.addAlert('ERROR', `Failed to load system metrics: ${error.message}`);
        }
    }

    addAlert(level, message) {
        const alert = {
            id: Date.now(),
            level,
            message,
            timestamp: new Date().toLocaleTimeString()
        };
        
        this.alerts.unshift(alert);
        if (this.alerts.length > 5) {
            this.alerts = this.alerts.slice(0, 5);
        }
        
        this.renderAlerts();
    }

    renderAlerts() {
        const container = this.querySelector('#alerts-container');
        if (container) {
            if (this.alerts.length === 0) {
                container.innerHTML = '<div class="no-alerts">No active alerts</div>';
            } else {
                container.innerHTML = this.alerts.map(alert => `
                    <div class="alert-item ${alert.level.toLowerCase()}">
                        <span class="alert-time">${alert.timestamp}</span>
                        <span class="alert-message">${alert.message}</span>
                    </div>
                `).join('');
            }
        }
    }
}

/**
 * Command Palette Widget
 * Quick command execution interface
 */
class CommandPaletteWidget extends NexusWidget {
    constructor(container, config) {
        super(container, config);
        this.commands = [];
        this.initializeCommands();
    }

    initializeCommands() {
        this.commands = [
            { name: 'Create Entity', command: 'entity:create', icon: '+' },
            { name: 'Search Entities', command: 'entity:search', icon: '🔍' },
            { name: 'System Health', command: 'system:health', icon: '⚡' },
            { name: 'View Metrics', command: 'metrics:view', icon: '📊' },
            { name: 'Export Data', command: 'data:export', icon: '📄' },
        ];
    }

    async render() {
        this.container.innerHTML = `
            <div class="command-palette">
                <div class="command-search">
                    <input type="text" placeholder="Type a command..." id="command-input">
                </div>
                <div class="command-list" id="command-list">
                    ${this.commands.map((cmd, index) => `
                        <div class="command-item" data-command="${cmd.command}">
                            <span class="command-icon">${cmd.icon}</span>
                            <span class="command-name">${cmd.name}</span>
                            <span class="command-shortcut">${cmd.command}</span>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;

        this.setupEventHandlers();
    }

    setupEventHandlers() {
        const input = this.querySelector('#command-input');
        const list = this.querySelector('#command-list');

        if (input) {
            input.addEventListener('input', (e) => {
                this.filterCommands(e.target.value);
            });
        }

        if (list) {
            list.addEventListener('click', (e) => {
                const item = e.target.closest('.command-item');
                if (item) {
                    this.executeCommand(item.dataset.command);
                }
            });
        }
    }

    filterCommands(query) {
        const filtered = this.commands.filter(cmd =>
            cmd.name.toLowerCase().includes(query.toLowerCase()) ||
            cmd.command.toLowerCase().includes(query.toLowerCase())
        );

        const list = this.querySelector('#command-list');
        if (list) {
            list.innerHTML = filtered.map(cmd => `
                <div class="command-item" data-command="${cmd.command}">
                    <span class="command-icon">${cmd.icon}</span>
                    <span class="command-name">${cmd.name}</span>
                    <span class="command-shortcut">${cmd.command}</span>
                </div>
            `).join('');
        }
    }

    executeCommand(command) {
        console.log(`Executing command: ${command}`);
        // Command execution logic would go here
    }
}

/**
 * Performance Monitor
 * Tracks widget performance metrics
 */
class PerformanceMonitor {
    constructor() {
        this.metrics = [];
    }

    recordMetric(entry) {
        this.metrics.push({
            name: entry.name,
            duration: entry.duration,
            timestamp: entry.startTime
        });

        // Keep only last 100 metrics
        if (this.metrics.length > 100) {
            this.metrics = this.metrics.slice(-100);
        }

        // Log slow widgets
        if (entry.duration > 100) {
            console.warn(`🐌 Slow widget render: ${entry.name} took ${entry.duration}ms`);
        }
    }

    getAverageRenderTime(widgetName) {
        const widgetMetrics = this.metrics.filter(m => m.name.includes(widgetName));
        if (widgetMetrics.length === 0) return 0;
        
        const total = widgetMetrics.reduce((sum, m) => sum + m.duration, 0);
        return total / widgetMetrics.length;
    }

    getPerformanceReport() {
        return {
            totalMetrics: this.metrics.length,
            averageRenderTime: this.metrics.reduce((sum, m) => sum + m.duration, 0) / this.metrics.length,
            slowestWidget: this.metrics.reduce((slowest, current) => 
                current.duration > slowest.duration ? current : slowest, 
                { duration: 0 }
            ),
            recentMetrics: this.metrics.slice(-10)
        };
    }
}

// Initialize the framework
window.NexusFramework = new NexusWidgetFramework();

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        NexusWidgetFramework,
        NexusWidget,
        TemporalExplorerWidget,
        EntityNetworkWidget,
        StorageMetricsWidget,
        LiveMonitorWidget,
        CommandPaletteWidget,
        PerformanceMonitor
    };
}