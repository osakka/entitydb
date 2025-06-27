/**
 * Base Component - Foundation for modular component architecture
 * Provides lifecycle management, event handling, and state management
 */
class BaseComponent {
    constructor(container, options = {}) {
        this.container = typeof container === 'string' 
            ? document.querySelector(container) 
            : container;
        
        if (!this.container) {
            throw new Error(`Container not found: ${container}`);
        }

        this.options = { ...this.defaultOptions, ...options };
        this.state = { ...this.defaultState };
        this.listeners = [];
        this.children = [];
        this.isDestroyed = false;

        this.init();
    }

    // Default options and state (to be overridden by subclasses)
    get defaultOptions() {
        return {};
    }

    get defaultState() {
        return {};
    }

    // Lifecycle methods
    init() {
        this.setupEventListeners();
        this.render();
    }

    setupEventListeners() {
        // Override in subclasses
    }

    render() {
        // Override in subclasses
    }

    destroy() {
        if (this.isDestroyed) return;

        // Clean up event listeners
        this.listeners.forEach(({ element, event, handler }) => {
            element.removeEventListener(event, handler);
        });
        this.listeners = [];

        // Destroy child components
        this.children.forEach(child => {
            if (child && typeof child.destroy === 'function') {
                child.destroy();
            }
        });
        this.children = [];

        // Clear container
        if (this.container) {
            this.container.innerHTML = '';
        }

        this.isDestroyed = true;
    }

    // State management
    setState(newState) {
        const oldState = { ...this.state };
        this.state = { ...this.state, ...newState };
        this.onStateChange(oldState, this.state);
    }

    onStateChange(oldState, newState) {
        // Override in subclasses for reactive updates
    }

    // Event handling helpers
    addEventListener(element, event, handler) {
        if (!element) {
            console.warn('addEventListener: element is null or undefined');
            return;
        }
        
        const wrappedHandler = (e) => {
            if (!this.isDestroyed) {
                handler.call(this, e);
            }
        };
        
        element.addEventListener(event, wrappedHandler);
        this.listeners.push({ element, event, handler: wrappedHandler });
    }

    emit(eventName, data = null) {
        const customEvent = new CustomEvent(eventName, {
            detail: data,
            bubbles: true
        });
        this.container.dispatchEvent(customEvent);
    }

    // DOM helpers
    createElement(tag, className = '', content = '') {
        const element = document.createElement(tag);
        if (className) element.className = className;
        if (content) element.innerHTML = content;
        return element;
    }

    find(selector) {
        return this.container.querySelector(selector);
    }

    findAll(selector) {
        return this.container.querySelectorAll(selector);
    }

    // Error handling
    handleError(error, context = 'Component') {
        console.error(`[${context}] Error:`, error);
        this.emit('error', { error, context });
    }

    // Loading state management
    setLoading(isLoading) {
        this.setState({ loading: isLoading });
        this.container.classList.toggle('loading', isLoading);
    }

    // Template rendering helper
    renderTemplate(template, data = {}) {
        let html = template;
        Object.keys(data).forEach(key => {
            const regex = new RegExp(`{{\\s*${key}\\s*}}`, 'g');
            html = html.replace(regex, data[key] || '');
        });
        return html;
    }

    // Animation helpers
    async fadeIn(element = this.container, duration = 300) {
        return new Promise(resolve => {
            element.style.opacity = '0';
            element.style.transition = `opacity ${duration}ms ease`;
            
            requestAnimationFrame(() => {
                element.style.opacity = '1';
                setTimeout(resolve, duration);
            });
        });
    }

    async fadeOut(element = this.container, duration = 300) {
        return new Promise(resolve => {
            element.style.transition = `opacity ${duration}ms ease`;
            element.style.opacity = '0';
            setTimeout(resolve, duration);
        });
    }

    // Utility methods
    debounce(func, wait) {
        let timeout;
        return (...args) => {
            clearTimeout(timeout);
            timeout = setTimeout(() => func.apply(this, args), wait);
        };
    }

    throttle(func, limit) {
        let inThrottle;
        return (...args) => {
            if (!inThrottle) {
                func.apply(this, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    }
}

// Export for use in other modules
window.BaseComponent = BaseComponent;