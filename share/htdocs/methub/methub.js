// MetHub - Main Application
function methub() {
    return {
        // State
        loading: false,
        autoRefresh: true,
        refreshInterval: null,
        timeRange: '1h',
        selectedHost: null,
        hosts: [],
        widgets: [],
        showAddWidget: false,
        editingWidget: null,
        gridColumns: 4,
        
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
        
        // Initialization
        async init() {
            console.log('🚀 Initializing MetHub...');
            
            // Initialize API
            this.api = new MetHubAPI();
            this.widgetRenderer = new MetHubWidgets();
            
            // Check authentication
            if (!this.api.token) {
                // Redirect to login or show login modal
                await this.login();
            }
            
            // Load saved widgets
            this.loadWidgets();
            
            // Load hosts
            await this.loadHosts();
            
            // Initial data load
            await this.refreshData();
            
            // Start auto-refresh
            this.startAutoRefresh();
            
            console.log('✅ MetHub initialized');
        },
        
        // Simple login (you might want to add a proper login screen)
        async login() {
            try {
                await this.api.login('admin', 'admin');
                console.log('✅ Logged in successfully');
            } catch (error) {
                console.error('❌ Login failed:', error);
                alert('Login failed. Please refresh and try again.');
            }
        },
        
        // Load available hosts
        async loadHosts() {
            try {
                this.hosts = await this.api.getHosts();
                console.log('📊 Found hosts:', this.hosts);
            } catch (error) {
                console.error('Error loading hosts:', error);
            }
        },
        
        // Load saved widgets from localStorage
        loadWidgets() {
            const saved = localStorage.getItem('methub-widgets');
            if (saved) {
                this.widgets = JSON.parse(saved);
            } else {
                // Default widgets
                this.widgets = [
                    {
                        id: this.generateId(),
                        title: 'CPU Usage',
                        type: 'gauge',
                        metricType: 'cpu',
                        metricName: 'cpu_usage',
                        span: 1,
                        thresholds: { warning: 70, critical: 90 }
                    },
                    {
                        id: this.generateId(),
                        title: 'Memory Usage',
                        type: 'gauge',
                        metricType: 'memory',
                        metricName: 'mem_percent',
                        span: 1,
                        thresholds: { warning: 80, critical: 95 }
                    },
                    {
                        id: this.generateId(),
                        title: 'CPU History',
                        type: 'line',
                        metricType: 'cpu',
                        metricName: 'cpu_usage',
                        span: 2
                    },
                    {
                        id: this.generateId(),
                        title: 'Disk Usage',
                        type: 'bar',
                        metricType: 'disk',
                        metricName: 'disk_percent',
                        span: 2
                    }
                ];
                this.saveWidgets();
            }
        },
        
        // Save widgets to localStorage
        saveWidgets() {
            localStorage.setItem('methub-widgets', JSON.stringify(this.widgets));
        },
        
        // Refresh all widget data
        async refreshData() {
            if (this.loading) return;
            
            this.loading = true;
            console.log('🔄 Refreshing data...');
            
            try {
                // Update each widget
                for (const widget of this.widgets) {
                    await this.updateWidget(widget);
                }
            } catch (error) {
                console.error('Error refreshing data:', error);
            } finally {
                this.loading = false;
            }
        },
        
        // Update single widget data
        async updateWidget(widget) {
            try {
                // Query metrics for this widget
                const metrics = await this.api.queryMetrics(
                    this.timeRange,
                    this.selectedHost,
                    widget.metricType,
                    widget.metricName
                );
                
                // Find widget container
                const container = document.querySelector(`#widget-${widget.id} .widget-content`);
                if (container) {
                    widget.api = this.api; // Pass API reference
                    await this.widgetRenderer.renderWidget(widget, metrics, container);
                }
            } catch (error) {
                console.error(`Error updating widget ${widget.title}:`, error);
            }
        },
        
        // Render widget HTML structure
        renderWidget(widget) {
            return `<div id="widget-${widget.id}" class="widget-content"></div>`;
        },
        
        // Start auto-refresh
        startAutoRefresh() {
            if (this.refreshInterval) {
                clearInterval(this.refreshInterval);
            }
            
            if (this.autoRefresh) {
                this.refreshInterval = setInterval(() => {
                    this.refreshData();
                }, 30000); // Refresh every 30 seconds
            }
        },
        
        // Add new widget
        addWidget() {
            this.editingWidget = null;
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
            this.showAddWidget = true;
        },
        
        // Edit existing widget
        editWidget(widget) {
            this.editingWidget = widget;
            this.widgetForm = { ...widget };
            this.showAddWidget = true;
        },
        
        // Save widget (add or update)
        saveWidget() {
            if (this.editingWidget) {
                // Update existing widget
                const index = this.widgets.findIndex(w => w.id === this.editingWidget.id);
                if (index !== -1) {
                    this.widgets[index] = { ...this.widgetForm, id: this.editingWidget.id };
                }
            } else {
                // Add new widget
                const newWidget = {
                    ...this.widgetForm,
                    id: this.generateId()
                };
                this.widgets.push(newWidget);
            }
            
            this.saveWidgets();
            this.closeWidgetModal();
            this.refreshData();
        },
        
        // Remove widget
        removeWidget(widgetId) {
            if (confirm('Remove this widget?')) {
                this.widgetRenderer.destroyChart(widgetId);
                this.widgets = this.widgets.filter(w => w.id !== widgetId);
                this.saveWidgets();
            }
        },
        
        // Close widget modal
        closeWidgetModal() {
            this.showAddWidget = false;
            this.editingWidget = null;
        },
        
        // Update widget defaults based on type
        updateWidgetDefaults() {
            const defaults = {
                gauge: { span: 1 },
                line: { span: 2 },
                bar: { span: 2 },
                value: { span: 1 },
                table: { span: 4 },
                heatmap: { span: 4 }
            };
            
            if (defaults[this.widgetForm.type]) {
                Object.assign(this.widgetForm, defaults[this.widgetForm.type]);
            }
        },
        
        // Generate unique ID
        generateId() {
            return 'widget_' + Math.random().toString(36).substr(2, 9);
        },
        
        // Watch for changes
        $watch('autoRefresh', function(value) {
            this.startAutoRefresh();
        }),
        
        $watch('timeRange', function(value) {
            this.refreshData();
        }),
        
        $watch('selectedHost', function(value) {
            this.refreshData();
        })
    };
}