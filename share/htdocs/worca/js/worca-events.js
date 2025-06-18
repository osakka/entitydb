// Worca Event System
// Handles real-time synchronization, notifications, and multi-user collaboration

class WorcaEvents {
    constructor(config = null, client = null) {
        this.config = config || window.worcaConfig;
        this.client = client || window.entityDBClient;
        
        this.eventListeners = new Map();
        this.syncTimer = null;
        this.lastSyncTime = Date.now();
        this.offlineQueue = [];
        this.isOnline = navigator.onLine;
        this.notifications = [];
        
        // Performance tracking
        this.metrics = {
            syncCycles: 0,
            entitiesSynced: 0,
            conflictsResolved: 0,
            errorsOccurred: 0
        };
        
        console.log('🌐 WorcaEvents initialized');
        this.initialize();
    }

    async initialize() {
        // Set up network monitoring
        this.initializeNetworkMonitoring();
        
        // Set up configuration listeners
        this.initializeConfigListeners();
        
        // Start real-time sync if enabled
        if (this.config.get('features.realTimeSync')) {
            this.startRealTimeSync();
        }
        
        // Initialize notifications
        this.initializeNotifications();
    }

    // Network Monitoring
    initializeNetworkMonitoring() {
        window.addEventListener('online', () => {
            this.isOnline = true;
            this.emit('network-online');
            this.showNotification('🌐 Connection restored', 'Back online - syncing data...', 'success');
            this.processOfflineQueue();
        });

        window.addEventListener('offline', () => {
            this.isOnline = false;
            this.emit('network-offline');
            this.showNotification('📴 Connection lost', 'Working offline - changes will sync when reconnected', 'warning');
        });
    }

    initializeConfigListeners() {
        this.config.on('config-changed', (event) => {
            if (event.path === 'features.realTimeSync') {
                if (event.value) {
                    this.startRealTimeSync();
                } else {
                    this.stopRealTimeSync();
                }
            }
        });

        this.config.on('workspace-changed', (event) => {
            this.lastSyncTime = Date.now(); // Reset sync time for new workspace
            this.emit('workspace-sync-reset');
        });

        this.config.on('entitydb-health', (event) => {
            if (event.status === 'healthy') {
                this.emit('entitydb-reconnected');
            } else {
                this.emit('entitydb-disconnected', event.error);
            }
        });
    }

    // Real-time Synchronization
    startRealTimeSync() {
        if (this.syncTimer) {
            clearInterval(this.syncTimer);
        }

        const interval = this.config.get('features.realTimeSync.interval') || 5000;
        
        console.log(`🔄 Starting real-time sync (${interval}ms interval)`);
        
        this.syncTimer = setInterval(() => {
            this.performSync();
        }, interval);

        // Initial sync
        setTimeout(() => this.performSync(), 1000);
    }

    stopRealTimeSync() {
        if (this.syncTimer) {
            clearInterval(this.syncTimer);
            this.syncTimer = null;
            console.log('⏹️ Real-time sync stopped');
        }
    }

    async performSync() {
        if (!this.isOnline || !this.config.status.connected) {
            return;
        }

        try {
            this.metrics.syncCycles++;
            
            // Get entities changed since last sync
            const changes = await this.detectChanges();
            
            if (changes.length > 0) {
                console.log(`🔄 Detected ${changes.length} changes since last sync`);
                
                // Process each change
                for (const change of changes) {
                    await this.processChange(change);
                }
                
                this.metrics.entitiesSynced += changes.length;
                this.emit('sync-complete', {
                    changesProcessed: changes.length,
                    timestamp: new Date().toISOString()
                });
            }
            
            this.lastSyncTime = Date.now();
            
        } catch (error) {
            this.metrics.errorsOccurred++;
            console.error('❌ Sync error:', error);
            this.emit('sync-error', error);
            
            if (error.message.includes('authentication')) {
                this.showNotification('🔐 Authentication required', 'Please log in again', 'error');
            }
        }
    }

    async detectChanges() {
        try {
            // Query for entities changed since last sync
            const sinceTime = new Date(this.lastSyncTime).toISOString();
            
            // Use EntityDB's changes endpoint if available
            if (this.client.makeRequest) {
                try {
                    const response = await this.client.makeRequest(`/entities/changes?since=${encodeURIComponent(sinceTime)}`);
                    if (response.ok) {
                        const data = await response.json();
                        return data.changes || data.entities || [];
                    }
                } catch (error) {
                    console.warn('Changes endpoint not available, falling back to full query');
                }
            }
            
            // Fallback: query all entities and filter by timestamp
            const allEntities = await this.client.queryEntities();
            const changes = allEntities.filter(entity => {
                const updatedAt = entity.updatedAt || entity.updated_at;
                if (!updatedAt) return false;
                
                const entityTime = typeof updatedAt === 'number' 
                    ? updatedAt / 1000000 // Convert nanoseconds to milliseconds
                    : new Date(updatedAt).getTime();
                    
                return entityTime > this.lastSyncTime;
            });
            
            return changes;
            
        } catch (error) {
            console.error('Failed to detect changes:', error);
            return [];
        }
    }

    async processChange(entity) {
        try {
            // Emit change event for UI updates
            this.emit('entity-changed', {
                entity,
                type: 'update',
                timestamp: new Date().toISOString()
            });
            
            // Check for conflicts with local changes
            const localVersion = await this.getLocalVersion(entity.id);
            if (localVersion && this.hasConflict(entity, localVersion)) {
                await this.resolveConflict(entity, localVersion);
                this.metrics.conflictsResolved++;
            }
            
        } catch (error) {
            console.error('Failed to process change:', error);
            throw error;
        }
    }

    async getLocalVersion(entityId) {
        // Check if entity exists in local cache/storage
        try {
            return await this.client.getEntity(entityId);
        } catch (error) {
            return null;
        }
    }

    hasConflict(remoteEntity, localEntity) {
        // Simple conflict detection based on timestamps
        const remoteTime = remoteEntity.updatedAt || remoteEntity.updated_at || 0;
        const localTime = localEntity.updatedAt || localEntity.updated_at || 0;
        
        // Convert nanoseconds to milliseconds if needed
        const remoteMs = typeof remoteTime === 'number' && remoteTime > 1e15 
            ? remoteTime / 1000000 
            : new Date(remoteTime).getTime();
            
        const localMs = typeof localTime === 'number' && localTime > 1e15 
            ? localTime / 1000000 
            : new Date(localTime).getTime();
        
        // Conflict if local was modified after remote but remote is newer
        return Math.abs(remoteMs - localMs) > 1000; // More than 1 second difference
    }

    async resolveConflict(remoteEntity, localEntity) {
        try {
            console.log(`🤝 Resolving conflict for entity ${remoteEntity.id}`);
            
            // For now, implement simple "remote wins" strategy
            // In a more sophisticated system, you might:
            // 1. Show conflict resolution UI
            // 2. Merge changes intelligently
            // 3. Use vector clocks or CRDTs
            
            this.emit('conflict-resolved', {
                entityId: remoteEntity.id,
                strategy: 'remote-wins',
                remoteEntity,
                localEntity
            });
            
            this.showNotification(
                '🤝 Conflict resolved',
                `Entity ${remoteEntity.id.slice(0, 8)}... updated by another user`,
                'info'
            );
            
        } catch (error) {
            console.error('Failed to resolve conflict:', error);
            throw error;
        }
    }

    // Offline Queue Management
    addToOfflineQueue(operation) {
        this.offlineQueue.push({
            ...operation,
            timestamp: Date.now(),
            id: Math.random().toString(36).substr(2, 9)
        });
        
        this.emit('offline-operation-queued', operation);
        
        // Limit queue size
        if (this.offlineQueue.length > 1000) {
            this.offlineQueue.shift();
        }
    }

    async processOfflineQueue() {
        if (!this.isOnline || this.offlineQueue.length === 0) {
            return;
        }

        console.log(`📤 Processing ${this.offlineQueue.length} offline operations...`);
        
        const queue = [...this.offlineQueue];
        this.offlineQueue = [];
        
        let processed = 0;
        let failed = 0;
        
        for (const operation of queue) {
            try {
                await this.executeOfflineOperation(operation);
                processed++;
            } catch (error) {
                console.warn('Failed to process offline operation:', error);
                failed++;
                
                // Re-queue if it's worth retrying
                if (operation.retries < 3) {
                    this.offlineQueue.push({
                        ...operation,
                        retries: (operation.retries || 0) + 1
                    });
                }
            }
        }
        
        this.showNotification(
            '📤 Offline changes synced',
            `${processed} operations processed${failed > 0 ? `, ${failed} failed` : ''}`,
            failed > 0 ? 'warning' : 'success'
        );
        
        this.emit('offline-queue-processed', { processed, failed });
    }

    async executeOfflineOperation(operation) {
        switch (operation.type) {
            case 'create':
                return await this.client.createEntity(operation.data);
            case 'update':
                return await this.client.updateEntity(operation.entityId, operation.data);
            case 'delete':
                return await this.client.updateEntity(operation.entityId, {
                    tags: [...operation.data.tags, 'status:deleted']
                });
            default:
                throw new Error(`Unknown operation type: ${operation.type}`);
        }
    }

    // Notification System
    initializeNotifications() {
        if (!this.config.get('features.notifications.enabled')) {
            return;
        }

        // Request notification permission
        if ('Notification' in window && Notification.permission === 'default') {
            Notification.requestPermission();
        }
    }

    showNotification(title, message, type = 'info', duration = null) {
        const notification = {
            id: Math.random().toString(36).substr(2, 9),
            title,
            message,
            type,
            timestamp: new Date().toISOString(),
            duration: duration || this.config.get('features.notifications.duration') || 5000
        };
        
        this.notifications.push(notification);
        this.emit('notification', notification);
        
        // Browser notification for important messages
        if (type === 'error' || type === 'warning') {
            this.showBrowserNotification(title, message);
        }
        
        // Auto-remove after duration
        if (notification.duration > 0) {
            setTimeout(() => {
                this.removeNotification(notification.id);
            }, notification.duration);
        }
        
        return notification.id;
    }

    showBrowserNotification(title, message) {
        if ('Notification' in window && Notification.permission === 'granted') {
            new Notification(title, {
                body: message,
                icon: '/worca/worca-logo-light.svg'
            });
        }
    }

    removeNotification(id) {
        const index = this.notifications.findIndex(n => n.id === id);
        if (index > -1) {
            this.notifications.splice(index, 1);
            this.emit('notification-removed', id);
        }
    }

    clearNotifications() {
        this.notifications = [];
        this.emit('notifications-cleared');
    }

    // Event System
    on(event, callback) {
        if (!this.eventListeners.has(event)) {
            this.eventListeners.set(event, []);
        }
        this.eventListeners.get(event).push(callback);
    }

    off(event, callback) {
        if (this.eventListeners.has(event)) {
            const listeners = this.eventListeners.get(event);
            const index = listeners.indexOf(callback);
            if (index > -1) {
                listeners.splice(index, 1);
            }
        }
    }

    emit(event, data = null) {
        if (this.eventListeners.has(event)) {
            this.eventListeners.get(event).forEach(callback => {
                try {
                    callback(data);
                } catch (error) {
                    console.error(`Error in event listener for ${event}:`, error);
                }
            });
        }
        
        // Also emit to global scope for integration
        if (window.dispatchEvent) {
            window.dispatchEvent(new CustomEvent(`worca:${event}`, { detail: data }));
        }
    }

    // Performance and Analytics
    getMetrics() {
        return {
            ...this.metrics,
            isOnline: this.isOnline,
            syncActive: !!this.syncTimer,
            offlineQueueSize: this.offlineQueue.length,
            notificationCount: this.notifications.length,
            lastSyncTime: this.lastSyncTime
        };
    }

    resetMetrics() {
        this.metrics = {
            syncCycles: 0,
            entitiesSynced: 0,
            conflictsResolved: 0,
            errorsOccurred: 0
        };
    }

    // Collaboration Features
    broadcastUserActivity(activity) {
        if (!this.isOnline) {
            return;
        }

        // For future implementation: WebSocket or Server-Sent Events
        this.emit('user-activity', {
            ...activity,
            userId: this.getCurrentUserId(),
            timestamp: new Date().toISOString()
        });
    }

    getCurrentUserId() {
        // Extract user ID from current session
        try {
            const token = this.client.token;
            if (token) {
                const payload = JSON.parse(atob(token.split('.')[1]));
                return payload.sub || payload.user_id || 'unknown';
            }
        } catch (error) {
            console.warn('Failed to get user ID from token:', error);
        }
        return 'anonymous';
    }

    // Cleanup
    destroy() {
        this.stopRealTimeSync();
        this.eventListeners.clear();
        this.offlineQueue = [];
        this.notifications = [];
        
        window.removeEventListener('online', this.handleOnline);
        window.removeEventListener('offline', this.handleOffline);
    }
}

// Global instance
window.worcaEvents = new WorcaEvents();

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = WorcaEvents;
}