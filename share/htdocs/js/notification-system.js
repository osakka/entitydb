/**
 * EntityDB Notification System
 * Advanced notification management with queuing and persistence
 * Version: v2.29.0
 */

class NotificationSystem {
    constructor() {
        this.notifications = new Map();
        this.container = this.createContainer();
        this.queue = [];
        this.maxVisible = 3;
        this.defaultDuration = 5000;
        this.positions = {
            'top-right': { top: '20px', right: '20px' },
            'top-left': { top: '20px', left: '20px' },
            'bottom-right': { bottom: '20px', right: '20px' },
            'bottom-left': { bottom: '20px', left: '20px' },
            'top-center': { top: '20px', left: '50%', transform: 'translateX(-50%)' },
            'bottom-center': { bottom: '20px', left: '50%', transform: 'translateX(-50%)' }
        };
        this.position = 'top-right';
        this.logger = new Logger('NotificationSystem');
    }
    
    /**
     * Create notification container
     */
    createContainer() {
        const existing = document.getElementById('entitydb-notifications');
        if (existing) {
            return existing;
        }
        
        const container = document.createElement('div');
        container.id = 'entitydb-notifications';
        container.className = 'notification-container';
        this.applyPosition(container);
        document.body.appendChild(container);
        return container;
    }
    
    /**
     * Apply position to container
     */
    applyPosition(container) {
        const styles = this.positions[this.position];
        Object.keys(styles).forEach(key => {
            container.style[key] = styles[key];
        });
    }
    
    /**
     * Change notification position
     */
    setPosition(position) {
        if (this.positions[position]) {
            this.position = position;
            this.applyPosition(this.container);
        }
    }
    
    /**
     * Show notification
     */
    show(message, type = 'info', options = {}) {
        const id = Date.now() + Math.random();
        const notification = {
            id,
            message,
            type,
            timestamp: new Date(),
            duration: options.duration !== undefined ? options.duration : this.defaultDuration,
            persistent: options.persistent || false,
            actions: options.actions || [],
            progress: options.progress || false,
            icon: options.icon || this.getDefaultIcon(type),
            title: options.title || this.getDefaultTitle(type),
            data: options.data || {}
        };
        
        this.logger.debug('Showing notification', notification);
        
        // Check if we need to queue
        if (this.notifications.size >= this.maxVisible) {
            this.queue.push(notification);
            this.logger.debug('Notification queued', { queueLength: this.queue.length });
            return id;
        }
        
        this.notifications.set(id, notification);
        this.render(notification);
        
        // Auto-dismiss if not persistent
        if (notification.duration > 0 && !notification.persistent) {
            setTimeout(() => this.dismiss(id), notification.duration);
        }
        
        return id;
    }
    
    /**
     * Show progress notification
     */
    showProgress(message, progress = 0, options = {}) {
        const id = this.show(message, 'progress', { ...options, progress: true, duration: 0 });
        const notification = this.notifications.get(id);
        if (notification) {
            notification.progress = progress;
            this.updateProgress(id, progress);
        }
        return id;
    }
    
    /**
     * Update progress notification
     */
    updateProgress(id, progress, message) {
        const notification = this.notifications.get(id);
        if (!notification) return;
        
        notification.progress = Math.min(100, Math.max(0, progress));
        if (message) {
            notification.message = message;
        }
        
        const element = document.getElementById(`notification-${id}`);
        if (element) {
            const progressBar = element.querySelector('.notification-progress-bar');
            if (progressBar) {
                progressBar.style.width = `${notification.progress}%`;
            }
            
            if (message) {
                const messageEl = element.querySelector('.notification-message');
                if (messageEl) {
                    messageEl.textContent = message;
                }
            }
        }
        
        // Auto-dismiss when complete
        if (notification.progress >= 100 && !notification.persistent) {
            setTimeout(() => this.dismiss(id), 1000);
        }
    }
    
    /**
     * Render notification
     */
    render(notification) {
        const element = document.createElement('div');
        element.id = `notification-${notification.id}`;
        element.className = `notification notification-${notification.type} notification-enter`;
        
        // Build notification HTML
        let html = `
            <div class="notification-content">
                <div class="notification-header">
                    <i class="${notification.icon}"></i>
                    ${notification.title ? `<span class="notification-title">${notification.title}</span>` : ''}
                    <button class="notification-close" onclick="window.notificationSystem.dismiss(${notification.id})">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                <div class="notification-body">
                    <p class="notification-message">${notification.message}</p>
                    ${notification.progress !== false ? `
                        <div class="notification-progress">
                            <div class="notification-progress-bar" style="width: ${notification.progress || 0}%"></div>
                        </div>
                    ` : ''}
                    ${notification.actions.length > 0 ? `
                        <div class="notification-actions">
                            ${notification.actions.map(action => `
                                <button class="notification-action" onclick="window.notificationSystem.handleAction(${notification.id}, '${action.id}')">
                                    ${action.label}
                                </button>
                            `).join('')}
                        </div>
                    ` : ''}
                </div>
            </div>
        `;
        
        element.innerHTML = html;
        this.container.appendChild(element);
        
        // Trigger enter animation
        requestAnimationFrame(() => {
            element.classList.remove('notification-enter');
            element.classList.add('notification-enter-active');
        });
    }
    
    /**
     * Dismiss notification
     */
    dismiss(id) {
        const notification = this.notifications.get(id);
        if (!notification) return;
        
        const element = document.getElementById(`notification-${id}`);
        if (element) {
            element.classList.add('notification-leave-active');
            element.addEventListener('transitionend', () => {
                element.remove();
                this.notifications.delete(id);
                
                // Process queue
                if (this.queue.length > 0) {
                    const next = this.queue.shift();
                    this.show(next.message, next.type, next);
                }
            }, { once: true });
        }
    }
    
    /**
     * Handle notification action
     */
    handleAction(notificationId, actionId) {
        const notification = this.notifications.get(notificationId);
        if (!notification) return;
        
        const action = notification.actions.find(a => a.id === actionId);
        if (action && action.handler) {
            action.handler(notification);
        }
        
        // Dismiss unless action prevents it
        if (!action || !action.preventDismiss) {
            this.dismiss(notificationId);
        }
    }
    
    /**
     * Clear all notifications
     */
    clearAll() {
        this.notifications.forEach((_, id) => this.dismiss(id));
        this.queue = [];
    }
    
    /**
     * Get default icon for type
     */
    getDefaultIcon(type) {
        const icons = {
            success: 'fas fa-check-circle',
            error: 'fas fa-exclamation-circle',
            warning: 'fas fa-exclamation-triangle',
            info: 'fas fa-info-circle',
            progress: 'fas fa-spinner fa-spin'
        };
        return icons[type] || icons.info;
    }
    
    /**
     * Get default title for type
     */
    getDefaultTitle(type) {
        const titles = {
            success: 'Success',
            error: 'Error',
            warning: 'Warning',
            info: 'Info',
            progress: 'Processing'
        };
        return titles[type] || '';
    }
    
    /**
     * Show specific notification types
     */
    success(message, options = {}) {
        return this.show(message, 'success', options);
    }
    
    error(message, options = {}) {
        return this.show(message, 'error', options);
    }
    
    warning(message, options = {}) {
        return this.show(message, 'warning', options);
    }
    
    info(message, options = {}) {
        return this.show(message, 'info', options);
    }
}

// Add CSS styles
const notificationStyles = document.createElement('style');
notificationStyles.textContent = `
/* Notification Container */
.notification-container {
    position: fixed;
    z-index: 10000;
    pointer-events: none;
}

/* Notification Base */
.notification {
    background: white;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    margin-bottom: 12px;
    min-width: 320px;
    max-width: 480px;
    pointer-events: all;
    transition: all 0.3s ease;
}

body.dark-mode .notification {
    background: #2c3e50;
    color: #e1e8ed;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

/* Notification Types */
.notification-success {
    border-left: 4px solid #10b981;
}

.notification-error {
    border-left: 4px solid #ef4444;
}

.notification-warning {
    border-left: 4px solid #f59e0b;
}

.notification-info {
    border-left: 4px solid #3b82f6;
}

.notification-progress {
    border-left: 4px solid #3b82f6;
}

/* Notification Content */
.notification-content {
    padding: 16px;
}

.notification-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 8px;
}

.notification-header i {
    font-size: 20px;
}

.notification-success .notification-header i {
    color: #10b981;
}

.notification-error .notification-header i {
    color: #ef4444;
}

.notification-warning .notification-header i {
    color: #f59e0b;
}

.notification-info .notification-header i,
.notification-progress .notification-header i {
    color: #3b82f6;
}

.notification-title {
    font-weight: 600;
    font-size: 16px;
    flex: 1;
}

.notification-close {
    background: none;
    border: none;
    color: #6c757d;
    cursor: pointer;
    padding: 4px;
    margin: -4px;
    border-radius: 4px;
    transition: all 0.2s;
}

.notification-close:hover {
    background: rgba(0, 0, 0, 0.05);
    color: #374151;
}

body.dark-mode .notification-close:hover {
    background: rgba(255, 255, 255, 0.1);
    color: #e1e8ed;
}

.notification-message {
    margin: 0;
    color: #374151;
    font-size: 14px;
    line-height: 1.5;
}

body.dark-mode .notification-message {
    color: #e1e8ed;
}

/* Progress Bar */
.notification-progress {
    margin-top: 12px;
    height: 4px;
    background: rgba(0, 0, 0, 0.1);
    border-radius: 2px;
    overflow: hidden;
}

body.dark-mode .notification-progress {
    background: rgba(255, 255, 255, 0.1);
}

.notification-progress-bar {
    height: 100%;
    background: #3b82f6;
    transition: width 0.3s ease;
}

/* Actions */
.notification-actions {
    display: flex;
    gap: 8px;
    margin-top: 12px;
}

.notification-action {
    padding: 6px 12px;
    background: none;
    border: 1px solid #e5e7eb;
    border-radius: 4px;
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
}

.notification-action:hover {
    background: #f3f4f6;
    border-color: #d1d5db;
}

body.dark-mode .notification-action {
    border-color: #34495e;
    color: #e1e8ed;
}

body.dark-mode .notification-action:hover {
    background: #34495e;
    border-color: #415a77;
}

/* Animations */
.notification-enter {
    opacity: 0;
    transform: translateX(100%);
}

.notification-enter-active {
    opacity: 1;
    transform: translateX(0);
}

.notification-leave-active {
    opacity: 0;
    transform: translateX(100%);
}

/* Responsive */
@media (max-width: 480px) {
    .notification {
        min-width: calc(100vw - 40px);
        max-width: calc(100vw - 40px);
    }
}
`;
document.head.appendChild(notificationStyles);

// Create global instance
window.notificationSystem = new NotificationSystem();

// Export for use in other modules
window.NotificationSystem = NotificationSystem;