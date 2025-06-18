// Worca Widget System - Programmable Dashboard Framework
// Powered by EntityDB

// Widget Registry - All available widget types
const WIDGET_REGISTRY = {
    // System Metrics Widgets
    systemInfo: {
        name: 'System Information',
        icon: 'fa-server',
        category: 'metrics',
        defaultSize: { w: 4, h: 3 },
        component: 'SystemInfoWidget'
    },
    memoryUsage: {
        name: 'Memory Usage',
        icon: 'fa-memory',
        category: 'metrics',
        defaultSize: { w: 4, h: 3 },
        component: 'MemoryUsageWidget'
    },
    databaseStats: {
        name: 'Database Statistics',
        icon: 'fa-database',
        category: 'metrics',
        defaultSize: { w: 4, h: 3 },
        component: 'DatabaseStatsWidget'
    },
    entityBreakdown: {
        name: 'Entity Breakdown',
        icon: 'fa-chart-pie',
        category: 'metrics',
        defaultSize: { w: 6, h: 4 },
        component: 'EntityBreakdownWidget'
    },
    storageUsage: {
        name: 'Storage Usage',
        icon: 'fa-hdd',
        category: 'metrics',
        defaultSize: { w: 4, h: 3 },
        component: 'StorageUsageWidget'
    },
    temporalStats: {
        name: 'Temporal Statistics',
        icon: 'fa-clock',
        category: 'metrics',
        defaultSize: { w: 4, h: 3 },
        component: 'TemporalStatsWidget'
    },
    performanceMetrics: {
        name: 'Performance Metrics',
        icon: 'fa-tachometer-alt',
        category: 'metrics',
        defaultSize: { w: 6, h: 4 },
        component: 'PerformanceMetricsWidget'
    },
    
    // Task Management Widgets
    taskOverview: {
        name: 'Task Overview',
        icon: 'fa-tasks',
        category: 'tasks',
        defaultSize: { w: 4, h: 3 },
        component: 'TaskOverviewWidget'
    },
    kanbanMini: {
        name: 'Mini Kanban',
        icon: 'fa-columns',
        category: 'tasks',
        defaultSize: { w: 8, h: 6 },
        component: 'MiniKanbanWidget'
    },
    recentActivity: {
        name: 'Recent Activity',
        icon: 'fa-stream',
        category: 'activity',
        defaultSize: { w: 4, h: 6 },
        component: 'RecentActivityWidget'
    },
    
    // Team Widgets
    teamWorkload: {
        name: 'Team Workload',
        icon: 'fa-users',
        category: 'team',
        defaultSize: { w: 6, h: 4 },
        component: 'TeamWorkloadWidget'
    },
    teamMembers: {
        name: 'Team Members',
        icon: 'fa-user-friends',
        category: 'team',
        defaultSize: { w: 4, h: 3 },
        component: 'TeamMembersWidget'
    },
    
    // Project Widgets
    projectStatus: {
        name: 'Project Status',
        icon: 'fa-project-diagram',
        category: 'projects',
        defaultSize: { w: 6, h: 4 },
        component: 'ProjectStatusWidget'
    },
    sprintProgress: {
        name: 'Sprint Progress',
        icon: 'fa-running',
        category: 'projects',
        defaultSize: { w: 4, h: 3 },
        component: 'SprintProgressWidget'
    },
    
    // Analytics Widgets
    taskTrends: {
        name: 'Task Trends',
        icon: 'fa-chart-line',
        category: 'analytics',
        defaultSize: { w: 6, h: 4 },
        component: 'TaskTrendsWidget'
    },
    velocityChart: {
        name: 'Velocity Chart',
        icon: 'fa-chart-area',
        category: 'analytics',
        defaultSize: { w: 6, h: 4 },
        component: 'VelocityChartWidget'
    },
    
    // Custom Widgets
    customHtml: {
        name: 'Custom HTML',
        icon: 'fa-code',
        category: 'custom',
        defaultSize: { w: 4, h: 3 },
        component: 'CustomHtmlWidget',
        configurable: true
    },
    customChart: {
        name: 'Custom Chart',
        icon: 'fa-chart-bar',
        category: 'custom',
        defaultSize: { w: 6, h: 4 },
        component: 'CustomChartWidget',
        configurable: true
    }
};

// Widget Categories
const WIDGET_CATEGORIES = {
    metrics: { name: 'System Metrics', icon: 'fa-server' },
    tasks: { name: 'Task Management', icon: 'fa-tasks' },
    team: { name: 'Team', icon: 'fa-users' },
    projects: { name: 'Projects', icon: 'fa-folder' },
    activity: { name: 'Activity', icon: 'fa-stream' },
    analytics: { name: 'Analytics', icon: 'fa-chart-line' },
    custom: { name: 'Custom', icon: 'fa-palette' }
};

// Widget Manager Component
function widgetManager() {
    return {
        // State
        dashboards: [],
        currentDashboard: null,
        widgets: [],
        gridster: null,
        editMode: false,
        showAddWidget: false,
        showDashboardSettings: false,
        metricsData: null,
        metricsRefreshInterval: null,
        selectedCategory: null,
        
        // Dashboard Form
        dashboardForm: {
            name: '',
            description: '',
            icon: 'fa-tachometer-alt',
            isPublic: false
        },
        
        // Widget Configuration
        selectedWidgetType: null,
        widgetConfig: {},
        
        // Initialize Widget Manager
        async initializeWidgets() {
            console.log('🧩 Initializing Widget Manager...');
            
            // Load dashboards from EntityDB
            await this.loadDashboards();
            
            // Initialize metrics refresh
            await this.refreshMetrics();
            this.startMetricsRefresh();
            
            // Initialize grid system when in dashboard view
            if (this.currentView === 'dashboard') {
                this.$nextTick(() => {
                    this.initializeGrid();
                });
            }
        },
        
        // Load user dashboards from EntityDB
        async loadDashboards() {
            try {
                const entities = await this.api.queryEntities({
                    tags: ['type:dashboard', `owner:${this.loginForm.username}`]
                });
                
                this.dashboards = entities.map(entity => {
                    const content = JSON.parse(atob(entity.content || 'e30='));
                    return {
                        id: entity.id,
                        name: content.name || 'Unnamed Dashboard',
                        description: content.description || '',
                        icon: content.icon || 'fa-tachometer-alt',
                        widgets: content.widgets || [],
                        isPublic: entity.tags.includes('visibility:public'),
                        owner: this.loginForm.username,
                        createdAt: entity.created_at,
                        updatedAt: entity.updated_at
                    };
                });
                
                // If no dashboards exist, create a default one
                if (this.dashboards.length === 0) {
                    await this.createDefaultDashboard();
                }
                
                // Set current dashboard to first one
                if (!this.currentDashboard && this.dashboards.length > 0) {
                    this.currentDashboard = this.dashboards[0];
                    this.widgets = [...this.currentDashboard.widgets];
                }
                
                console.log(`📊 Loaded ${this.dashboards.length} dashboards`);
            } catch (error) {
                console.error('Failed to load dashboards:', error);
            }
        },
        
        // Create default dashboard
        async createDefaultDashboard() {
            const defaultDashboard = {
                name: 'Main Dashboard',
                description: 'Your primary workspace dashboard',
                icon: 'fa-tachometer-alt',
                widgets: [
                    {
                        id: this.generateId(),
                        type: 'systemInfo',
                        x: 0, y: 0, w: 4, h: 3
                    },
                    {
                        id: this.generateId(),
                        type: 'taskOverview',
                        x: 4, y: 0, w: 4, h: 3
                    },
                    {
                        id: this.generateId(),
                        type: 'teamMembers',
                        x: 8, y: 0, w: 4, h: 3
                    },
                    {
                        id: this.generateId(),
                        type: 'recentActivity',
                        x: 0, y: 3, w: 4, h: 6
                    },
                    {
                        id: this.generateId(),
                        type: 'kanbanMini',
                        x: 4, y: 3, w: 8, h: 6
                    }
                ]
            };
            
            await this.saveDashboard(defaultDashboard);
        },
        
        // Save dashboard to EntityDB
        async saveDashboard(dashboard) {
            try {
                const content = {
                    name: dashboard.name,
                    description: dashboard.description,
                    icon: dashboard.icon,
                    widgets: dashboard.widgets || this.widgets
                };
                
                const tags = [
                    'type:dashboard',
                    `owner:${this.loginForm.username}`,
                    dashboard.isPublic ? 'visibility:public' : 'visibility:private'
                ];
                
                if (dashboard.id) {
                    // Update existing dashboard
                    await this.api.updateEntity(dashboard.id, {
                        content: btoa(JSON.stringify(content)),
                        tags
                    });
                } else {
                    // Create new dashboard
                    const entity = await this.api.createEntity({
                        content: btoa(JSON.stringify(content)),
                        tags
                    });
                    dashboard.id = entity.id;
                }
                
                // Reload dashboards
                await this.loadDashboards();
                
                console.log('💾 Dashboard saved:', dashboard.name);
            } catch (error) {
                console.error('Failed to save dashboard:', error);
                alert('Failed to save dashboard: ' + error.message);
            }
        },
        
        // Delete dashboard
        async deleteDashboard(dashboardId) {
            if (!confirm('Are you sure you want to delete this dashboard?')) {
                return;
            }
            
            try {
                await this.api.deleteEntity(dashboardId);
                await this.loadDashboards();
                console.log('🗑️ Dashboard deleted');
            } catch (error) {
                console.error('Failed to delete dashboard:', error);
                alert('Failed to delete dashboard: ' + error.message);
            }
        },
        
        // Switch dashboard
        switchDashboard(dashboard) {
            this.currentDashboard = dashboard;
            this.widgets = [...dashboard.widgets];
            
            // Reinitialize grid
            if (this.gridster) {
                this.gridster.destroy();
            }
            this.$nextTick(() => {
                this.initializeGrid();
            });
        },
        
        // Create new dashboard
        createNewDashboard() {
            this.dashboardForm = {
                name: '',
                description: '',
                icon: 'fa-tachometer-alt',
                isPublic: false
            };
            this.showDashboardSettings = true;
        },
        
        // Save new dashboard
        async saveNewDashboard() {
            if (!this.dashboardForm.name) {
                alert('Please enter a dashboard name');
                return;
            }
            
            const newDashboard = {
                ...this.dashboardForm,
                widgets: []
            };
            
            await this.saveDashboard(newDashboard);
            this.showDashboardSettings = false;
        },
        
        // Initialize grid system
        initializeGrid() {
            if (!this.currentDashboard || this.currentView !== 'dashboard') {
                return;
            }
            
            console.log('📐 Initializing grid system...');
            
            // Wait for DOM
            this.$nextTick(() => {
                const gridContainer = document.querySelector('.widget-grid');
                if (!gridContainer) {
                    console.warn('Grid container not found');
                    return;
                }
                
                // Initialize Gridster
                this.gridster = $('.widget-grid').gridster({
                    widget_selector: '.widget',
                    widget_margins: [10, 10],
                    widget_base_dimensions: [80, 80],
                    max_cols: 12,
                    min_cols: 12,
                    autogenerate_stylesheet: true,
                    resize: {
                        enabled: this.editMode,
                        max_size: [12, 12],
                        min_size: [2, 2]
                    },
                    draggable: {
                        enabled: this.editMode,
                        handle: '.widget-header'
                    },
                    serialize_params: function($w, wgd) {
                        return {
                            id: $w.attr('data-widget-id'),
                            col: wgd.col,
                            row: wgd.row,
                            size_x: wgd.size_x,
                            size_y: wgd.size_y
                        };
                    }
                }).data('gridster');
                
                console.log('✅ Grid system initialized');
            });
        },
        
        // Toggle edit mode
        toggleEditMode() {
            this.editMode = !this.editMode;
            
            if (this.gridster) {
                if (this.editMode) {
                    this.gridster.enable();
                    this.gridster.enable_resize();
                } else {
                    this.gridster.disable();
                    this.gridster.disable_resize();
                    // Save widget positions
                    this.saveWidgetPositions();
                }
            }
        },
        
        // Save widget positions
        async saveWidgetPositions() {
            if (!this.gridster || !this.currentDashboard) return;
            
            const serialized = this.gridster.serialize();
            
            // Update widget positions
            serialized.forEach(item => {
                const widget = this.widgets.find(w => w.id === item.id);
                if (widget) {
                    widget.x = item.col - 1;
                    widget.y = item.row - 1;
                    widget.w = item.size_x;
                    widget.h = item.size_y;
                }
            });
            
            // Save to EntityDB
            await this.saveDashboard(this.currentDashboard);
        },
        
        // Add widget to dashboard
        addWidget(type) {
            const widgetDef = WIDGET_REGISTRY[type];
            if (!widgetDef) return;
            
            const newWidget = {
                id: this.generateId(),
                type: type,
                x: 0,
                y: 0,
                w: widgetDef.defaultSize.w,
                h: widgetDef.defaultSize.h,
                config: {}
            };
            
            this.widgets.push(newWidget);
            
            // Add to grid
            if (this.gridster) {
                const html = this.renderWidget(newWidget);
                this.gridster.add_widget(html, newWidget.w, newWidget.h);
            }
            
            this.showAddWidget = false;
        },
        
        // Remove widget
        removeWidget(widgetId) {
            const index = this.widgets.findIndex(w => w.id === widgetId);
            if (index > -1) {
                this.widgets.splice(index, 1);
                
                // Remove from grid
                if (this.gridster) {
                    const $widget = $(`[data-widget-id="${widgetId}"]`);
                    this.gridster.remove_widget($widget);
                }
            }
        },
        
        // Configure widget
        configureWidget(widgetId) {
            const widget = this.widgets.find(w => w.id === widgetId);
            if (!widget) return;
            
            // Open configuration modal based on widget type
            console.log('Configure widget:', widget);
            // TODO: Implement widget configuration modal
        },
        
        // Refresh metrics data
        async refreshMetrics() {
            try {
                const response = await fetch('/api/v1/system/metrics');
                if (response.ok) {
                    this.metricsData = await response.json();
                    console.log('📊 Metrics refreshed');
                }
            } catch (error) {
                console.error('Failed to fetch metrics:', error);
            }
        },
        
        // Start metrics auto-refresh
        startMetricsRefresh() {
            // Refresh every 30 seconds
            this.metricsRefreshInterval = setInterval(() => {
                this.refreshMetrics();
            }, 30000);
        },
        
        // Stop metrics refresh
        stopMetricsRefresh() {
            if (this.metricsRefreshInterval) {
                clearInterval(this.metricsRefreshInterval);
                this.metricsRefreshInterval = null;
            }
        },
        
        // Get widget data based on type
        getWidgetData(widget) {
            switch (widget.type) {
                case 'systemInfo':
                    return this.metricsData?.system || {};
                case 'memoryUsage':
                    return this.metricsData?.memory || {};
                case 'databaseStats':
                    return this.metricsData?.database || {};
                case 'storageUsage':
                    return this.metricsData?.storage || {};
                case 'temporalStats':
                    return this.metricsData?.temporal || {};
                case 'performanceMetrics':
                    return this.metricsData?.performance || {};
                case 'taskOverview':
                    return this.stats || {};
                case 'recentActivity':
                    return this.recentActivity || [];
                case 'teamMembers':
                    return { members: this.teamMembers || [], count: this.teamMembers?.length || 0 };
                default:
                    return {};
            }
        },
        
        // Format bytes to human readable
        formatBytes(bytes) {
            if (!bytes) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        },
        
        // Format duration
        formatDuration(seconds) {
            if (!seconds) return '0s';
            const days = Math.floor(seconds / 86400);
            const hours = Math.floor((seconds % 86400) / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            
            if (days > 0) return `${days}d ${hours}h`;
            if (hours > 0) return `${hours}h ${minutes}m`;
            return `${minutes}m`;
        },
        
        // Render widget (for dynamic HTML generation)
        renderWidget(widget) {
            const widgetDef = WIDGET_REGISTRY[widget.type];
            return `
                <div class="widget" data-widget-id="${widget.id}">
                    <div class="widget-header">
                        <i class="fas ${widgetDef.icon}"></i>
                        <span>${widgetDef.name}</span>
                        <div class="widget-actions">
                            ${widgetDef.configurable ? `<button onclick="$root.configureWidget('${widget.id}')" title="Configure"><i class="fas fa-cog"></i></button>` : ''}
                            <button onclick="$root.removeWidget('${widget.id}')" title="Remove"><i class="fas fa-times"></i></button>
                        </div>
                    </div>
                    <div class="widget-body" id="widget-body-${widget.id}">
                        <!-- Widget content will be rendered here -->
                    </div>
                </div>
            `;
        }
    };
}

// Export for use in main app
window.widgetManager = widgetManager;
window.WIDGET_REGISTRY = WIDGET_REGISTRY;
window.WIDGET_CATEGORIES = WIDGET_CATEGORIES;