// EntityDB Widget System - Clean, Simple, Maintainable
class WidgetSystem {
    constructor(container, options = {}) {
        this.container = container;
        this.widgets = new Map();
        this.layout = [];
        this.options = {
            columns: 3,
            gap: 16,
            animationDuration: 300,
            ...options
        };
        
        this.init();
    }
    
    init() {
        // Setup container
        this.container.classList.add('widget-container');
        this.container.style.display = 'flex';
        this.container.style.flexWrap = 'wrap';
        this.container.style.gap = `${this.options.gap}px`;
        
        // Setup drag and drop
        this.setupDragAndDrop();
    }
    
    // Register a widget type
    registerWidget(type, config) {
        this.widgets.set(type, {
            component: config.component,
            defaultSize: config.defaultSize || 'medium',
            minRefreshInterval: config.minRefreshInterval || 30000,
            render: config.render,
            onMount: config.onMount,
            onUnmount: config.onUnmount,
            onResize: config.onResize
        });
    }
    
    // Add widget to dashboard
    addWidget(widgetData) {
        const widgetConfig = this.widgets.get(widgetData.type);
        if (!widgetConfig) {
            console.error(`Unknown widget type: ${widgetData.type}`);
            return null;
        }
        
        // Create widget element
        const widget = document.createElement('div');
        widget.className = 'widget';
        widget.dataset.widgetId = widgetData.id;
        widget.dataset.widgetType = widgetData.type;
        widget.dataset.size = widgetData.size || widgetConfig.defaultSize;
        widget.draggable = true;
        
        // Widget structure
        widget.innerHTML = `
            <div class="widget-header">
                <h3 class="widget-title">${widgetData.config?.title || widgetData.type}</h3>
                <div class="widget-actions">
                    <button class="widget-action" data-action="refresh" title="Refresh">
                        <i class="fas fa-sync-alt"></i>
                    </button>
                    <button class="widget-action" data-action="settings" title="Settings">
                        <i class="fas fa-cog"></i>
                    </button>
                    <button class="widget-action" data-action="remove" title="Remove">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
            </div>
            <div class="widget-content"></div>
        `;
        
        // Add to layout
        this.layout.push({
            id: widgetData.id,
            type: widgetData.type,
            size: widgetData.size || widgetConfig.defaultSize,
            config: widgetData.config || {}
        });
        
        // Render widget content
        const content = widget.querySelector('.widget-content');
        if (widgetConfig.render) {
            widgetConfig.render(content, widgetData.config || {});
        }
        
        // Setup event handlers
        this.setupWidgetEvents(widget, widgetData);
        
        // Add to container
        this.container.appendChild(widget);
        
        // Call mount lifecycle
        if (widgetConfig.onMount) {
            widgetConfig.onMount(widget, widgetData.config || {});
        }
        
        return widget;
    }
    
    // Remove widget
    removeWidget(widgetId) {
        const widget = this.container.querySelector(`[data-widget-id="${widgetId}"]`);
        if (!widget) return;
        
        const widgetType = widget.dataset.widgetType;
        const widgetConfig = this.widgets.get(widgetType);
        
        // Call unmount lifecycle
        if (widgetConfig?.onUnmount) {
            widgetConfig.onUnmount(widget);
        }
        
        // Remove from layout
        this.layout = this.layout.filter(w => w.id !== widgetId);
        
        // Remove from DOM with animation
        widget.style.opacity = '0';
        widget.style.transform = 'scale(0.9)';
        setTimeout(() => {
            widget.remove();
            // Emit layout changed event after removal
            this.saveLayout();
        }, this.options.animationDuration);
    }
    
    // Setup widget events
    setupWidgetEvents(widget, widgetData) {
        const header = widget.querySelector('.widget-header');
        
        // Action buttons
        widget.addEventListener('click', (e) => {
            const action = e.target.closest('[data-action]')?.dataset.action;
            if (!action) return;
            
            switch(action) {
                case 'refresh':
                    this.refreshWidget(widgetData.id);
                    break;
                case 'settings':
                    this.showWidgetSettings(widgetData.id);
                    break;
                case 'remove':
                    this.removeWidget(widgetData.id);
                    break;
            }
        });
        
        // Size toggle on double-click
        header.addEventListener('dblclick', () => {
            this.cycleWidgetSize(widgetData.id);
        });
    }
    
    // Cycle widget size
    cycleWidgetSize(widgetId) {
        const widget = this.container.querySelector(`[data-widget-id="${widgetId}"]`);
        if (!widget) return;
        
        const sizes = ['small', 'medium', 'large'];
        const currentSize = widget.dataset.size;
        const currentIndex = sizes.indexOf(currentSize);
        const newSize = sizes[(currentIndex + 1) % sizes.length];
        
        widget.dataset.size = newSize;
        
        // Update layout
        const layoutItem = this.layout.find(w => w.id === widgetId);
        if (layoutItem) {
            layoutItem.size = newSize;
        }
        
        // Trigger resize event
        const widgetConfig = this.widgets.get(widget.dataset.widgetType);
        if (widgetConfig?.onResize) {
            widgetConfig.onResize(widget, newSize);
        }
        
        // Save layout after size change
        this.saveLayout();
    }
    
    // Setup drag and drop
    setupDragAndDrop() {
        let draggedWidget = null;
        let dropPlaceholder = null;
        
        this.container.addEventListener('dragstart', (e) => {
            const widget = e.target.closest('.widget');
            if (!widget) return;
            
            draggedWidget = widget;
            widget.classList.add('dragging');
            e.dataTransfer.effectAllowed = 'move';
            
            // Create placeholder
            dropPlaceholder = document.createElement('div');
            dropPlaceholder.className = 'widget-placeholder';
            dropPlaceholder.dataset.size = widget.dataset.size;
        });
        
        this.container.addEventListener('dragend', (e) => {
            const widget = e.target.closest('.widget');
            if (!widget) return;
            
            widget.classList.remove('dragging');
            dropPlaceholder?.remove();
            draggedWidget = null;
            dropPlaceholder = null;
            
            // Save new layout order
            this.saveLayout();
        });
        
        this.container.addEventListener('dragover', (e) => {
            e.preventDefault();
            if (!draggedWidget) return;
            
            const afterElement = this.getDragAfterElement(e.clientY);
            if (afterElement == null) {
                this.container.appendChild(dropPlaceholder);
            } else {
                this.container.insertBefore(dropPlaceholder, afterElement);
            }
        });
        
        this.container.addEventListener('drop', (e) => {
            e.preventDefault();
            if (!draggedWidget || !dropPlaceholder) return;
            
            dropPlaceholder.replaceWith(draggedWidget);
        });
    }
    
    // Get element after drag position
    getDragAfterElement(y) {
        const draggableElements = [...this.container.querySelectorAll('.widget:not(.dragging)')];
        
        return draggableElements.reduce((closest, child) => {
            const box = child.getBoundingClientRect();
            const offset = y - box.top - box.height / 2;
            
            if (offset < 0 && offset > closest.offset) {
                return { offset: offset, element: child };
            } else {
                return closest;
            }
        }, { offset: Number.NEGATIVE_INFINITY }).element;
    }
    
    // Save layout to storage
    saveLayout() {
        // Update layout array based on DOM order
        const widgets = [...this.container.querySelectorAll('.widget')];
        const newLayout = widgets.map(widget => {
            return this.layout.find(w => w.id === widget.dataset.widgetId);
        }).filter(Boolean);
        
        this.layout = newLayout;
        
        // Emit save event
        this.container.dispatchEvent(new CustomEvent('layout-changed', {
            detail: { layout: this.layout }
        }));
    }
    
    // Load layout
    loadLayout(layout) {
        // Clear existing widgets
        this.container.innerHTML = '';
        this.layout = [];
        
        // Add widgets in order
        layout.forEach(widgetData => {
            this.addWidget(widgetData);
        });
    }
    
    // Refresh widget
    refreshWidget(widgetId) {
        const widget = this.container.querySelector(`[data-widget-id="${widgetId}"]`);
        if (!widget) return;
        
        const widgetType = widget.dataset.widgetType;
        const widgetConfig = this.widgets.get(widgetType);
        const layoutItem = this.layout.find(w => w.id === widgetId);
        
        if (widgetConfig?.render && layoutItem) {
            const content = widget.querySelector('.widget-content');
            const refreshBtn = widget.querySelector('[data-action="refresh"] i');
            
            // Add spinning animation
            refreshBtn?.classList.add('fa-spin');
            
            // Re-render content
            widgetConfig.render(content, layoutItem.config || {});
            
            // Remove spinning after delay
            setTimeout(() => {
                refreshBtn?.classList.remove('fa-spin');
            }, 1000);
        }
    }
    
    // Show widget settings
    showWidgetSettings(widgetId) {
        // This would open a modal with widget-specific settings
        console.log('Show settings for widget:', widgetId);
        // Implementation depends on your modal system
    }
}

// Widget Registry - Define available widget types
const widgetRegistry = {
    // System Stats Widget
    systemStats: {
        defaultSize: 'medium',
        render: (container, config) => {
            container.innerHTML = `
                <div class="stats-grid">
                    <div class="stat-item">
                        <div class="stat-value">${config.totalEntities || '0'}</div>
                        <div class="stat-label">Total Entities</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">${config.activeUsers || '0'}</div>
                        <div class="stat-label">Active Users</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">${config.uptime || '0d'}</div>
                        <div class="stat-label">Uptime</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">${config.healthScore || '0%'}</div>
                        <div class="stat-label">Health Score</div>
                    </div>
                </div>
            `;
        }
    },
    
    // Performance Chart Widget
    performanceChart: {
        defaultSize: 'large',
        render: (container, config) => {
            container.innerHTML = `
                <div class="chart-container">
                    <canvas id="perf-chart-${Date.now()}"></canvas>
                </div>
            `;
            // Initialize chart here
        },
        onMount: (widget, config) => {
            // Initialize Chart.js
        },
        onUnmount: (widget) => {
            // Destroy chart instance
        },
        onResize: (widget, newSize) => {
            // Update chart size
        }
    },
    
    // Activity Feed Widget
    activityFeed: {
        defaultSize: 'small',
        render: (container, config) => {
            container.innerHTML = `
                <div class="activity-feed">
                    <div class="activity-item">
                        <span class="activity-time">2m ago</span>
                        <span class="activity-text">User login: admin</span>
                    </div>
                    <!-- More items -->
                </div>
            `;
        }
    }
};

// Export for use
window.WidgetSystem = WidgetSystem;
window.widgetRegistry = widgetRegistry;