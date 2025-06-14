// UI Resilience System - EntityDB v2.31.0
// Enhanced UI error recovery and stability mechanisms

class UIResilience {
    constructor() {
        this.componentStates = new Map();
        this.errorBoundaries = new Map();
        this.recoveryAttempts = new Map();
        this.maxRecoveryAttempts = 3;
        this.healthChecks = new Map();
        this.renderingErrors = [];
        this.componentMetrics = new Map();
        
        this.initializeResilience();
    }
    
    // Initialize UI resilience system
    initializeResilience() {
        this.setupVueErrorHandler();
        this.setupDOMObserver();
        this.setupPerformanceMonitoring();
        this.setupTabValidation();
        this.startHealthChecks();
    }
    
    // Setup Vue.js error handling
    setupVueErrorHandler() {
        // Global Vue error handler
        window.addEventListener('vue:error', (event) => {
            this.handleVueError(event.detail);
        });
        
        // Override Vue's error handler if available
        if (window.Vue && window.Vue.config) {
            const originalErrorHandler = window.Vue.config.errorHandler;
            window.Vue.config.errorHandler = (err, vm, info) => {
                this.handleVueError({
                    error: err,
                    vm: vm,
                    info: info
                });
                
                if (originalErrorHandler) {
                    originalErrorHandler(err, vm, info);
                }
            };
        }
    }
    
    // Setup DOM mutation observer
    setupDOMObserver() {
        const observer = new MutationObserver((mutations) => {
            mutations.forEach((mutation) => {
                if (mutation.type === 'childList') {
                    this.validateDOMChanges(mutation);
                }
            });
        });
        
        observer.observe(document.body, {
            childList: true,
            subtree: true,
            attributes: true
        });
        
        this.domObserver = observer;
    }
    
    // Setup performance monitoring
    setupPerformanceMonitoring() {
        // Monitor long tasks
        if ('PerformanceObserver' in window) {
            const observer = new PerformanceObserver((list) => {
                list.getEntries().forEach((entry) => {
                    if (entry.duration > 50) { // Tasks longer than 50ms
                        this.handlePerformanceIssue({
                            type: 'long_task',
                            duration: entry.duration,
                            startTime: entry.startTime,
                            name: entry.name
                        });
                    }
                });
            });
            
            try {
                observer.observe({ entryTypes: ['longtask'] });
            } catch (e) {
                console.warn('Long task monitoring not supported');
            }
        }
        
        // Monitor memory usage
        this.startMemoryMonitoring();
    }
    
    // Setup tab structure validation
    setupTabValidation() {
        this.tabValidator = {
            validate: () => {
                const tabs = document.querySelectorAll('.nav-tab');
                const tabContents = document.querySelectorAll('.tab-content');
                
                const issues = [];
                
                // Check if tabs and content match
                if (tabs.length !== tabContents.length) {
                    issues.push(`Tab count mismatch: ${tabs.length} tabs, ${tabContents.length} content`);
                }
                
                // Check for orphaned content
                tabContents.forEach((content, index) => {
                    if (!content.classList.contains('active') && !tabs[index]) {
                        issues.push(`Orphaned tab content at index ${index}`);
                    }
                });
                
                // Check for missing Vue directives
                const vueElements = document.querySelectorAll('[v-if], [v-for], [v-show]');
                vueElements.forEach((element) => {
                    if (element.style.display === 'none' && !element.hasAttribute('v-show')) {
                        issues.push(`Potentially broken Vue directive on ${element.tagName}`);
                    }
                });
                
                return issues;
            }
        };
    }
    
    // Start health checks
    startHealthChecks() {
        setInterval(() => {
            this.performHealthCheck();
        }, 30000); // Check every 30 seconds
    }
    
    // Handle Vue.js errors
    handleVueError(errorData) {
        const errorId = this.generateErrorId();
        const componentName = this.getComponentName(errorData.vm);
        
        const error = {
            id: errorId,
            type: 'vue_error',
            component: componentName,
            message: errorData.error.message,
            stack: errorData.error.stack,
            info: errorData.info,
            timestamp: new Date().toISOString(),
            severity: 'high'
        };
        
        this.renderingErrors.push(error);
        
        // Attempt component recovery
        this.attemptComponentRecovery(componentName, errorData.vm);
        
        // Notify error handler
        if (window.ErrorHandler) {
            window.ErrorHandler.createError(
                'vue_error',
                `Vue component error in ${componentName}: ${error.message}`,
                'ui',
                'high',
                error
            );
        }
    }
    
    // Attempt component recovery
    attemptComponentRecovery(componentName, vm) {
        const attemptKey = `${componentName}_${Date.now()}`;
        const currentAttempts = this.recoveryAttempts.get(componentName) || 0;
        
        if (currentAttempts >= this.maxRecoveryAttempts) {
            this.handleComponentFailure(componentName);
            return;
        }
        
        this.recoveryAttempts.set(componentName, currentAttempts + 1);
        
        setTimeout(() => {
            try {
                // Force component re-render
                if (vm && vm.$forceUpdate) {
                    vm.$forceUpdate();
                } else if (vm && vm.forceUpdate) {
                    vm.forceUpdate();
                }
                
                // Validate component state
                this.validateComponentState(componentName, vm);
                
                // Reset recovery counter on success
                this.recoveryAttempts.set(componentName, 0);
                
            } catch (recoveryError) {
                console.error(`Component recovery failed for ${componentName}:`, recoveryError);
                this.handleComponentFailure(componentName);
            }
        }, 1000 * (currentAttempts + 1)); // Exponential backoff
    }
    
    // Handle component failure
    handleComponentFailure(componentName) {
        console.error(`Component ${componentName} failed to recover after ${this.maxRecoveryAttempts} attempts`);
        
        // Create fallback UI
        this.createFallbackUI(componentName);
        
        // Emit component failure event
        window.dispatchEvent(new CustomEvent('ui:component-failure', {
            detail: {
                component: componentName,
                timestamp: new Date().toISOString()
            }
        }));
    }
    
    // Create fallback UI
    createFallbackUI(componentName) {
        const fallbackElement = document.createElement('div');
        fallbackElement.className = 'component-fallback';
        fallbackElement.innerHTML = `
            <div class="fallback-content">
                <i class="fas fa-exclamation-triangle" style="color: #ffc107; font-size: 24px; margin-bottom: 12px;"></i>
                <h3>Component Error</h3>
                <p>The ${componentName} component encountered an error and needs to be reloaded.</p>
                <button onclick="location.reload()" class="btn btn-primary">
                    <i class="fas fa-sync-alt"></i> Reload Page
                </button>
            </div>
        `;
        
        // Find and replace the failed component
        const selector = this.getComponentSelector(componentName);
        const targetElement = document.querySelector(selector);
        if (targetElement && targetElement.parentNode) {
            targetElement.parentNode.replaceChild(fallbackElement, targetElement);
        }
    }
    
    // Get component selector
    getComponentSelector(componentName) {
        const selectorMap = {
            'dashboard': '.tab-content[data-tab="dashboard"]',
            'entities': '.tab-content[data-tab="entities"]', 
            'performance': '.tab-content[data-tab="performance"]',
            'errors': '.tab-content[data-tab="errors"]',
            'default': '.tab-content'
        };
        
        return selectorMap[componentName] || selectorMap.default;
    }
    
    // Validate DOM changes
    validateDOMChanges(mutation) {
        // Check for broken tab structure
        if (mutation.target.classList && mutation.target.classList.contains('tab-content')) {
            const issues = this.tabValidator.validate();
            if (issues.length > 0) {
                this.handleTabStructureIssues(issues);
            }
        }
        
        // Check for missing critical elements
        this.validateCriticalElements();
    }
    
    // Validate critical elements
    validateCriticalElements() {
        const criticalSelectors = [
            '#app',
            '.app-header',
            '.nav-tabs',
            '.main-content'
        ];
        
        criticalSelectors.forEach(selector => {
            const element = document.querySelector(selector);
            if (!element) {
                this.handleMissingCriticalElement(selector);
            }
        });
    }
    
    // Handle tab structure issues
    handleTabStructureIssues(issues) {
        console.warn('Tab structure issues detected:', issues);
        
        if (window.ErrorHandler) {
            window.ErrorHandler.createError(
                'tab_structure_error',
                `Tab structure validation failed: ${issues.join(', ')}`,
                'ui',
                'medium',
                { issues: issues }
            );
        }
        
        // Attempt to fix tab structure
        this.attemptTabRepair();
    }
    
    // Attempt tab repair
    attemptTabRepair() {
        try {
            // Force tab reinitialization
            const app = window.entitydbApp;
            if (app && app.switchTab) {
                const currentTab = app.activeTab;
                app.switchTab('dashboard');
                setTimeout(() => {
                    app.switchTab(currentTab);
                }, 100);
            }
        } catch (error) {
            console.error('Tab repair failed:', error);
        }
    }
    
    // Handle missing critical element
    handleMissingCriticalElement(selector) {
        console.error('Critical element missing:', selector);
        
        if (window.ErrorHandler) {
            window.ErrorHandler.createError(
                'missing_critical_element',
                `Critical UI element missing: ${selector}`,
                'ui',
                'critical'
            );
        }
        
        // Suggest page reload for critical elements
        if (selector === '#app') {
            this.suggestPageReload('Critical application container missing');
        }
    }
    
    // Handle performance issues
    handlePerformanceIssue(issue) {
        console.warn('Performance issue detected:', issue);
        
        this.componentMetrics.set(issue.name || 'unknown', {
            lastIssue: issue,
            issueCount: (this.componentMetrics.get(issue.name)?.issueCount || 0) + 1,
            timestamp: new Date().toISOString()
        });
        
        // Throttle performance warnings
        if (issue.duration > 100) {
            if (window.ErrorHandler) {
                window.ErrorHandler.createError(
                    'performance_issue',
                    `Long task detected: ${issue.duration.toFixed(2)}ms`,
                    'performance',
                    'medium',
                    issue
                );
            }
        }
    }
    
    // Start memory monitoring
    startMemoryMonitoring() {
        if ('memory' in performance) {
            setInterval(() => {
                const memory = performance.memory;
                const usageRatio = memory.usedJSHeapSize / memory.jsHeapSizeLimit;
                
                if (usageRatio > 0.8) {
                    this.handleMemoryPressure(memory);
                }
            }, 60000); // Check every minute
        }
    }
    
    // Handle memory pressure
    handleMemoryPressure(memory) {
        console.warn('High memory usage detected:', memory);
        
        if (window.ErrorHandler) {
            window.ErrorHandler.createError(
                'memory_pressure',
                `High memory usage: ${(memory.usedJSHeapSize / 1024 / 1024).toFixed(2)}MB`,
                'performance',
                'high',
                memory
            );
        }
        
        // Suggest memory cleanup
        this.performMemoryCleanup();
    }
    
    // Perform memory cleanup
    performMemoryCleanup() {
        try {
            // Clear old error entries
            if (window.ErrorHandler) {
                const oldErrors = window.ErrorHandler.getErrors({ 
                    since: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString() 
                });
                if (oldErrors.length > 100) {
                    window.ErrorHandler.clearErrors();
                }
            }
            
            // Clear old request logs
            if (window.RequestDebugger) {
                const stats = window.RequestDebugger.getStats();
                if (stats.totalRequests > 200) {
                    window.RequestDebugger.clearLogs();
                }
            }
            
            // Force garbage collection if available
            if (window.gc) {
                window.gc();
            }
            
        } catch (error) {
            console.error('Memory cleanup failed:', error);
        }
    }
    
    // Perform health check
    performHealthCheck() {
        const health = {
            timestamp: new Date().toISOString(),
            status: 'healthy',
            issues: []
        };
        
        // Check tab structure
        const tabIssues = this.tabValidator.validate();
        if (tabIssues.length > 0) {
            health.status = 'degraded';
            health.issues.push(...tabIssues);
        }
        
        // Check component states
        this.componentStates.forEach((state, component) => {
            if (state.status === 'failed') {
                health.status = 'unhealthy';
                health.issues.push(`Component ${component} is in failed state`);
            }
        });
        
        // Check error rate
        const recentErrors = this.renderingErrors.filter(error => 
            new Date(error.timestamp).getTime() > Date.now() - 300000 // 5 minutes
        );
        
        if (recentErrors.length > 5) {
            health.status = 'unhealthy';
            health.issues.push(`High error rate: ${recentErrors.length} errors in 5 minutes`);
        }
        
        // Store health check result
        this.healthChecks.set(Date.now(), health);
        
        // Keep only last 24 health checks
        const healthEntries = Array.from(this.healthChecks.entries()).slice(-24);
        this.healthChecks.clear();
        healthEntries.forEach(([timestamp, healthData]) => {
            this.healthChecks.set(timestamp, healthData);
        });
        
        // Emit health status
        window.dispatchEvent(new CustomEvent('ui:health-check', {
            detail: health
        }));
        
        return health;
    }
    
    // Suggest page reload
    suggestPageReload(reason) {
        const notification = document.createElement('div');
        notification.className = 'reload-suggestion';
        notification.innerHTML = `
            <div class="notification notification-warning">
                <div class="notification-content">
                    <i class="fas fa-exclamation-triangle"></i>
                    <div>
                        <strong>UI Issue Detected</strong>
                        <p>${reason}. A page reload is recommended.</p>
                    </div>
                </div>
                <div class="notification-actions">
                    <button onclick="location.reload()" class="btn btn-sm btn-primary">
                        <i class="fas fa-sync-alt"></i> Reload
                    </button>
                    <button onclick="this.parentElement.parentElement.parentElement.remove()" class="btn btn-sm btn-secondary">
                        Dismiss
                    </button>
                </div>
            </div>
        `;
        
        document.body.appendChild(notification);
        
        // Auto-remove after 30 seconds
        setTimeout(() => {
            if (notification.parentElement) {
                notification.remove();
            }
        }, 30000);
    }
    
    // Validate component state
    validateComponentState(componentName, vm) {
        const state = {
            name: componentName,
            status: 'healthy',
            timestamp: new Date().toISOString(),
            vm: vm
        };
        
        try {
            // Check if component is responsive
            if (vm && vm.$el) {
                const element = vm.$el;
                if (!element.isConnected) {
                    state.status = 'disconnected';
                }
            }
            
            // Check for Vue reactivity
            if (vm && vm._isVue) {
                if (!vm._watcher || vm._watcher.dirty) {
                    state.status = 'stale';
                }
            }
            
        } catch (error) {
            state.status = 'failed';
            state.error = error.message;
        }
        
        this.componentStates.set(componentName, state);
        return state;
    }
    
    // Get component name from Vue instance
    getComponentName(vm) {
        if (!vm) return 'unknown';
        
        if (vm.$options && vm.$options.name) {
            return vm.$options.name;
        }
        
        if (vm.$el && vm.$el.className) {
            const classes = vm.$el.className.split(' ');
            const tabClass = classes.find(cls => cls.includes('tab-'));
            if (tabClass) {
                return tabClass.replace('tab-', '');
            }
        }
        
        return 'component';
    }
    
    // Generate error ID
    generateErrorId() {
        return 'ui_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }
    
    // Get health status
    getHealthStatus() {
        const latest = Array.from(this.healthChecks.values()).pop();
        return latest || { status: 'unknown', timestamp: new Date().toISOString() };
    }
    
    // Get component metrics
    getComponentMetrics() {
        return Object.fromEntries(this.componentMetrics);
    }
    
    // Get rendering errors
    getRenderingErrors() {
        return [...this.renderingErrors];
    }
    
    // Reset component state
    resetComponentState(componentName) {
        this.componentStates.delete(componentName);
        this.recoveryAttempts.delete(componentName);
    }
    
    // Cleanup
    cleanup() {
        if (this.domObserver) {
            this.domObserver.disconnect();
        }
        
        this.componentStates.clear();
        this.errorBoundaries.clear();
        this.recoveryAttempts.clear();
        this.healthChecks.clear();
        this.renderingErrors.length = 0;
        this.componentMetrics.clear();
    }
}

// Global UI resilience instance
window.UIResilience = window.UIResilience || new UIResilience();

// Auto-initialize on DOM ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        window.UIResilience.performHealthCheck();
    });
} else {
    window.UIResilience.performHealthCheck();
}

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = UIResilience;
}