/**
 * EntityDB Component Lazy Loading System
 * Dynamic component loading with caching and preloading
 */

class LazyLoader {
    constructor() {
        this.components = new Map();
        this.loadingPromises = new Map();
        this.preloadQueue = [];
        this.intersectionObserver = null;
        this.loadedScripts = new Set();
        this.loadedStyles = new Set();
        
        // Initialize Intersection Observer for lazy loading
        this.initIntersectionObserver();
    }

    /**
     * Register a component for lazy loading
     */
    register(name, loader, options = {}) {
        this.components.set(name, {
            loader,
            loaded: false,
            instance: null,
            preload: options.preload || false,
            priority: options.priority || 0,
            dependencies: options.dependencies || [],
            styles: options.styles || [],
            minDelay: options.minDelay || 0
        });

        // Add to preload queue if needed
        if (options.preload) {
            this.preloadQueue.push({ name, priority: options.priority || 0 });
            this.preloadQueue.sort((a, b) => b.priority - a.priority);
        }
    }

    /**
     * Load a component
     */
    async load(name) {
        const component = this.components.get(name);
        if (!component) {
            throw new Error(`Component '${name}' not registered`);
        }

        // Return cached instance if already loaded
        if (component.loaded && component.instance) {
            return component.instance;
        }

        // Return existing loading promise if in progress
        if (this.loadingPromises.has(name)) {
            return this.loadingPromises.get(name);
        }

        // Create loading promise
        const loadingPromise = this.loadComponent(name, component);
        this.loadingPromises.set(name, loadingPromise);

        try {
            const instance = await loadingPromise;
            this.loadingPromises.delete(name);
            return instance;
        } catch (error) {
            this.loadingPromises.delete(name);
            throw error;
        }
    }

    /**
     * Internal component loading logic
     */
    async loadComponent(name, component) {
        const startTime = performance.now();

        try {
            // Load dependencies first
            if (component.dependencies.length > 0) {
                await Promise.all(
                    component.dependencies.map(dep => this.load(dep))
                );
            }

            // Load styles
            if (component.styles.length > 0) {
                await Promise.all(
                    component.styles.map(style => this.loadStyle(style))
                );
            }

            // Show loading indicator
            this.showLoadingIndicator(name);

            // Load the component
            const instance = await component.loader();

            // Apply minimum delay if specified
            const loadTime = performance.now() - startTime;
            if (component.minDelay > 0 && loadTime < component.minDelay) {
                await this.delay(component.minDelay - loadTime);
            }

            // Cache the instance
            component.instance = instance;
            component.loaded = true;

            // Hide loading indicator
            this.hideLoadingIndicator(name);

            logger.debug(`Component '${name}' loaded in ${Math.round(loadTime)}ms`);

            return instance;
        } catch (error) {
            this.hideLoadingIndicator(name);
            logger.error(`Failed to load component '${name}':`, error);
            throw error;
        }
    }

    /**
     * Load a script dynamically
     */
    async loadScript(src) {
        if (this.loadedScripts.has(src)) {
            return;
        }

        return new Promise((resolve, reject) => {
            const script = document.createElement('script');
            script.src = src;
            script.async = true;
            
            script.onload = () => {
                this.loadedScripts.add(src);
                resolve();
            };
            
            script.onerror = () => {
                reject(new Error(`Failed to load script: ${src}`));
            };
            
            document.head.appendChild(script);
        });
    }

    /**
     * Load a stylesheet dynamically
     */
    async loadStyle(href) {
        if (this.loadedStyles.has(href)) {
            return;
        }

        return new Promise((resolve, reject) => {
            const link = document.createElement('link');
            link.rel = 'stylesheet';
            link.href = href;
            
            link.onload = () => {
                this.loadedStyles.add(href);
                resolve();
            };
            
            link.onerror = () => {
                reject(new Error(`Failed to load stylesheet: ${href}`));
            };
            
            document.head.appendChild(link);
        });
    }

    /**
     * Preload components based on priority
     */
    async preloadComponents() {
        for (const { name } of this.preloadQueue) {
            try {
                await this.load(name);
                logger.debug(`Preloaded component: ${name}`);
            } catch (error) {
                logger.error(`Failed to preload component '${name}':`, error);
            }
        }
    }

    /**
     * Initialize Intersection Observer for viewport-based lazy loading
     */
    initIntersectionObserver() {
        if (!('IntersectionObserver' in window)) {
            logger.warn('IntersectionObserver not supported');
            return;
        }

        this.intersectionObserver = new IntersectionObserver(
            (entries) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        const element = entry.target;
                        const componentName = element.getAttribute('data-lazy-component');
                        
                        if (componentName) {
                            this.loadAndMount(componentName, element);
                            this.intersectionObserver.unobserve(element);
                        }
                    }
                });
            },
            {
                rootMargin: '50px' // Start loading 50px before entering viewport
            }
        );
    }

    /**
     * Observe an element for lazy loading
     */
    observe(element) {
        if (this.intersectionObserver) {
            this.intersectionObserver.observe(element);
        }
    }

    /**
     * Load and mount a component to an element
     */
    async loadAndMount(componentName, element) {
        try {
            const Component = await this.load(componentName);
            
            if (typeof Component === 'function') {
                // Vue component
                if (window.Vue) {
                    const app = window.Vue.createApp(Component);
                    app.mount(element);
                }
                // React component
                else if (window.React && window.ReactDOM) {
                    window.ReactDOM.render(
                        window.React.createElement(Component),
                        element
                    );
                }
                // Generic component with mount method
                else if (Component.mount) {
                    Component.mount(element);
                }
            }
        } catch (error) {
            logger.error(`Failed to mount component '${componentName}':`, error);
            element.innerHTML = `<div class="error">Failed to load component</div>`;
        }
    }

    /**
     * Show loading indicator
     */
    showLoadingIndicator(name) {
        const elements = document.querySelectorAll(`[data-lazy-component="${name}"]`);
        elements.forEach(el => {
            if (!el.querySelector('.lazy-loading')) {
                const loader = document.createElement('div');
                loader.className = 'lazy-loading';
                loader.innerHTML = `
                    <div class="spinner-border spinner-border-sm" role="status">
                        <span class="visually-hidden">Loading...</span>
                    </div>
                `;
                el.appendChild(loader);
            }
        });
    }

    /**
     * Hide loading indicator
     */
    hideLoadingIndicator(name) {
        const elements = document.querySelectorAll(`[data-lazy-component="${name}"]`);
        elements.forEach(el => {
            const loader = el.querySelector('.lazy-loading');
            if (loader) {
                loader.remove();
            }
        });
    }

    /**
     * Utility delay function
     */
    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    /**
     * Scan DOM for lazy-loadable components
     */
    scan(root = document) {
        const elements = root.querySelectorAll('[data-lazy-component]');
        elements.forEach(element => this.observe(element));
    }

    /**
     * Cleanup
     */
    destroy() {
        if (this.intersectionObserver) {
            this.intersectionObserver.disconnect();
        }
        this.components.clear();
        this.loadingPromises.clear();
        this.preloadQueue = [];
    }
}

// Create global lazy loader instance
const lazyLoader = new LazyLoader();

// Register core components
lazyLoader.register('entity-browser', async () => {
    await lazyLoader.loadScript('/js/entity-browser.js');
    return window.EntityBrowser;
}, {
    preload: false,
    priority: 5,
    styles: ['/css/entity-browser.css']
});

lazyLoader.register('rbac-manager', async () => {
    await lazyLoader.loadScript('/js/rbac-manager.js');
    return window.RBACManager;
}, {
    preload: false,
    priority: 3
});

lazyLoader.register('dataset-manager', async () => {
    await lazyLoader.loadScript('/js/dataset-manager.js');
    return window.DatasetManager;
}, {
    preload: false,
    priority: 3
});

lazyLoader.register('performance-monitor', async () => {
    await lazyLoader.loadScript('/js/performance-monitor.js');
    return window.PerformanceMonitor;
}, {
    preload: false,
    priority: 2,
    dependencies: ['enhanced-charts']
});

lazyLoader.register('enhanced-charts', async () => {
    await lazyLoader.loadScript('/js/enhanced-charts.js');
    return window.enhancedCharts;
}, {
    preload: true,
    priority: 8,
    minDelay: 100
});

lazyLoader.register('realtime-charts', async () => {
    await lazyLoader.loadScript('/js/realtime-charts.js');
    return window.realtimeCharts;
}, {
    preload: false,
    priority: 4,
    dependencies: ['enhanced-charts']
});

// Auto-scan on DOM ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        lazyLoader.scan();
        lazyLoader.preloadComponents();
    });
} else {
    lazyLoader.scan();
    lazyLoader.preloadComponents();
}

// Export global instance
window.lazyLoader = lazyLoader;