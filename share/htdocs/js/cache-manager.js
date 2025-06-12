/**
 * EntityDB Cache Management System
 * Advanced caching with TTL, LRU eviction, and storage adapters
 */

class CacheManager {
    constructor(options = {}) {
        this.maxSize = options.maxSize || 1000;
        this.defaultTTL = options.defaultTTL || 300000; // 5 minutes
        this.checkInterval = options.checkInterval || 60000; // 1 minute
        this.storage = options.storage || 'memory';
        
        // Storage adapters
        this.adapters = {
            memory: new MemoryAdapter(),
            localStorage: new LocalStorageAdapter(),
            indexedDB: new IndexedDBAdapter()
        };
        
        this.adapter = this.adapters[this.storage];
        this.stats = {
            hits: 0,
            misses: 0,
            sets: 0,
            deletes: 0,
            evictions: 0
        };
        
        // Start cleanup interval
        this.cleanupInterval = setInterval(() => {
            this.cleanup();
        }, this.checkInterval);
        
        this.logger = window.logger || console;
    }

    /**
     * Get value from cache
     */
    async get(key, defaultValue = null) {
        try {
            const item = await this.adapter.get(key);
            
            if (!item) {
                this.stats.misses++;
                return defaultValue;
            }
            
            // Check if expired
            if (item.expiresAt && Date.now() > item.expiresAt) {
                await this.delete(key);
                this.stats.misses++;
                return defaultValue;
            }
            
            // Update access time for LRU
            item.accessedAt = Date.now();
            await this.adapter.set(key, item);
            
            this.stats.hits++;
            return item.value;
        } catch (error) {
            this.logger.error('Cache get error:', error);
            this.stats.misses++;
            return defaultValue;
        }
    }

    /**
     * Set value in cache
     */
    async set(key, value, ttl = null) {
        try {
            const now = Date.now();
            const item = {
                value,
                createdAt: now,
                accessedAt: now,
                expiresAt: ttl ? now + ttl : (this.defaultTTL ? now + this.defaultTTL : null),
                size: this.calculateSize(value)
            };
            
            // Check size limits
            await this.ensureCapacity(item.size);
            
            await this.adapter.set(key, item);
            this.stats.sets++;
            
            return true;
        } catch (error) {
            this.logger.error('Cache set error:', error);
            return false;
        }
    }

    /**
     * Delete value from cache
     */
    async delete(key) {
        try {
            const result = await this.adapter.delete(key);
            if (result) {
                this.stats.deletes++;
            }
            return result;
        } catch (error) {
            this.logger.error('Cache delete error:', error);
            return false;
        }
    }

    /**
     * Check if key exists
     */
    async has(key) {
        try {
            const item = await this.adapter.get(key);
            if (!item) return false;
            
            // Check if expired
            if (item.expiresAt && Date.now() > item.expiresAt) {
                await this.delete(key);
                return false;
            }
            
            return true;
        } catch (error) {
            this.logger.error('Cache has error:', error);
            return false;
        }
    }

    /**
     * Get or set pattern
     */
    async getOrSet(key, factory, ttl = null) {
        let value = await this.get(key);
        
        if (value === null) {
            if (typeof factory === 'function') {
                value = await factory();
            } else {
                value = factory;
            }
            
            if (value !== null) {
                await this.set(key, value, ttl);
            }
        }
        
        return value;
    }

    /**
     * Get multiple keys
     */
    async getMany(keys) {
        const results = {};
        
        await Promise.all(keys.map(async (key) => {
            results[key] = await this.get(key);
        }));
        
        return results;
    }

    /**
     * Set multiple key-value pairs
     */
    async setMany(items, ttl = null) {
        const promises = Object.entries(items).map(([key, value]) => 
            this.set(key, value, ttl)
        );
        
        return Promise.all(promises);
    }

    /**
     * Clear cache by pattern
     */
    async clear(pattern = null) {
        try {
            if (pattern) {
                const keys = await this.adapter.keys();
                const regex = new RegExp(pattern);
                const matchingKeys = keys.filter(key => regex.test(key));
                
                await Promise.all(matchingKeys.map(key => this.delete(key)));
                return matchingKeys.length;
            } else {
                const count = await this.adapter.size();
                await this.adapter.clear();
                return count;
            }
        } catch (error) {
            this.logger.error('Cache clear error:', error);
            return 0;
        }
    }

    /**
     * Get cache statistics
     */
    async getStats() {
        const size = await this.adapter.size();
        const hitRate = this.stats.hits + this.stats.misses > 0 
            ? (this.stats.hits / (this.stats.hits + this.stats.misses) * 100).toFixed(2)
            : 0;
        
        return {
            ...this.stats,
            size,
            hitRate: `${hitRate}%`,
            maxSize: this.maxSize
        };
    }

    /**
     * Cleanup expired items
     */
    async cleanup() {
        try {
            const keys = await this.adapter.keys();
            const now = Date.now();
            let cleaned = 0;
            
            for (const key of keys) {
                const item = await this.adapter.get(key);
                if (item && item.expiresAt && now > item.expiresAt) {
                    await this.delete(key);
                    cleaned++;
                }
            }
            
            if (cleaned > 0) {
                this.logger.debug(`Cache cleanup: removed ${cleaned} expired items`);
            }
        } catch (error) {
            this.logger.error('Cache cleanup error:', error);
        }
    }

    /**
     * Ensure cache capacity
     */
    async ensureCapacity(itemSize) {
        const currentSize = await this.adapter.size();
        
        if (currentSize >= this.maxSize) {
            await this.evictLRU(Math.ceil(this.maxSize * 0.1)); // Evict 10%
        }
    }

    /**
     * Evict least recently used items
     */
    async evictLRU(count) {
        try {
            const keys = await this.adapter.keys();
            const items = [];
            
            // Get all items with access times
            for (const key of keys) {
                const item = await this.adapter.get(key);
                if (item) {
                    items.push({ key, accessedAt: item.accessedAt || 0 });
                }
            }
            
            // Sort by access time (oldest first)
            items.sort((a, b) => a.accessedAt - b.accessedAt);
            
            // Evict oldest items
            const toEvict = items.slice(0, count);
            await Promise.all(toEvict.map(item => this.delete(item.key)));
            
            this.stats.evictions += toEvict.length;
            
            this.logger.debug(`Cache LRU eviction: removed ${toEvict.length} items`);
        } catch (error) {
            this.logger.error('Cache eviction error:', error);
        }
    }

    /**
     * Calculate approximate size of value
     */
    calculateSize(value) {
        if (typeof value === 'string') {
            return value.length * 2; // UTF-16
        } else if (typeof value === 'object') {
            return JSON.stringify(value).length * 2;
        } else {
            return 8; // Approximate for primitives
        }
    }

    /**
     * Create a cache namespace
     */
    namespace(prefix) {
        return new CacheNamespace(this, prefix);
    }

    /**
     * Destroy cache manager
     */
    destroy() {
        if (this.cleanupInterval) {
            clearInterval(this.cleanupInterval);
        }
        
        if (this.adapter && this.adapter.destroy) {
            this.adapter.destroy();
        }
    }
}

/**
 * Cache namespace for logical separation
 */
class CacheNamespace {
    constructor(cache, prefix) {
        this.cache = cache;
        this.prefix = prefix + ':';
    }

    async get(key, defaultValue) {
        return this.cache.get(this.prefix + key, defaultValue);
    }

    async set(key, value, ttl) {
        return this.cache.set(this.prefix + key, value, ttl);
    }

    async delete(key) {
        return this.cache.delete(this.prefix + key);
    }

    async has(key) {
        return this.cache.has(this.prefix + key);
    }

    async getOrSet(key, factory, ttl) {
        return this.cache.getOrSet(this.prefix + key, factory, ttl);
    }

    async clear() {
        return this.cache.clear(this.prefix);
    }
}

/**
 * Memory storage adapter
 */
class MemoryAdapter {
    constructor() {
        this.store = new Map();
    }

    async get(key) {
        return this.store.get(key) || null;
    }

    async set(key, value) {
        this.store.set(key, value);
        return true;
    }

    async delete(key) {
        return this.store.delete(key);
    }

    async keys() {
        return Array.from(this.store.keys());
    }

    async size() {
        return this.store.size;
    }

    async clear() {
        this.store.clear();
        return true;
    }
}

/**
 * LocalStorage adapter
 */
class LocalStorageAdapter {
    constructor() {
        this.prefix = 'entitydb-cache:';
    }

    async get(key) {
        try {
            const item = localStorage.getItem(this.prefix + key);
            return item ? JSON.parse(item) : null;
        } catch (error) {
            return null;
        }
    }

    async set(key, value) {
        try {
            localStorage.setItem(this.prefix + key, JSON.stringify(value));
            return true;
        } catch (error) {
            return false;
        }
    }

    async delete(key) {
        try {
            localStorage.removeItem(this.prefix + key);
            return true;
        } catch (error) {
            return false;
        }
    }

    async keys() {
        const keys = [];
        for (let i = 0; i < localStorage.length; i++) {
            const key = localStorage.key(i);
            if (key && key.startsWith(this.prefix)) {
                keys.push(key.substring(this.prefix.length));
            }
        }
        return keys;
    }

    async size() {
        return (await this.keys()).length;
    }

    async clear() {
        const keys = await this.keys();
        keys.forEach(key => localStorage.removeItem(this.prefix + key));
        return true;
    }
}

/**
 * IndexedDB adapter (for large data)
 */
class IndexedDBAdapter {
    constructor() {
        this.dbName = 'EntityDBCache';
        this.storeName = 'cache';
        this.version = 1;
        this.db = null;
    }

    async init() {
        if (this.db) return this.db;
        
        return new Promise((resolve, reject) => {
            const request = indexedDB.open(this.dbName, this.version);
            
            request.onerror = () => reject(request.error);
            request.onsuccess = () => {
                this.db = request.result;
                resolve(this.db);
            };
            
            request.onupgradeneeded = (event) => {
                const db = event.target.result;
                if (!db.objectStoreNames.contains(this.storeName)) {
                    db.createObjectStore(this.storeName);
                }
            };
        });
    }

    async get(key) {
        const db = await this.init();
        return new Promise((resolve, reject) => {
            const transaction = db.transaction([this.storeName], 'readonly');
            const store = transaction.objectStore(this.storeName);
            const request = store.get(key);
            
            request.onsuccess = () => resolve(request.result || null);
            request.onerror = () => reject(request.error);
        });
    }

    async set(key, value) {
        const db = await this.init();
        return new Promise((resolve, reject) => {
            const transaction = db.transaction([this.storeName], 'readwrite');
            const store = transaction.objectStore(this.storeName);
            const request = store.put(value, key);
            
            request.onsuccess = () => resolve(true);
            request.onerror = () => reject(request.error);
        });
    }

    async delete(key) {
        const db = await this.init();
        return new Promise((resolve, reject) => {
            const transaction = db.transaction([this.storeName], 'readwrite');
            const store = transaction.objectStore(this.storeName);
            const request = store.delete(key);
            
            request.onsuccess = () => resolve(true);
            request.onerror = () => reject(request.error);
        });
    }

    async keys() {
        const db = await this.init();
        return new Promise((resolve, reject) => {
            const transaction = db.transaction([this.storeName], 'readonly');
            const store = transaction.objectStore(this.storeName);
            const request = store.getAllKeys();
            
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async size() {
        const db = await this.init();
        return new Promise((resolve, reject) => {
            const transaction = db.transaction([this.storeName], 'readonly');
            const store = transaction.objectStore(this.storeName);
            const request = store.count();
            
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async clear() {
        const db = await this.init();
        return new Promise((resolve, reject) => {
            const transaction = db.transaction([this.storeName], 'readwrite');
            const store = transaction.objectStore(this.storeName);
            const request = store.clear();
            
            request.onsuccess = () => resolve(true);
            request.onerror = () => reject(request.error);
        });
    }

    destroy() {
        if (this.db) {
            this.db.close();
            this.db = null;
        }
    }
}

// Create global cache instances
const cacheManager = new CacheManager({
    maxSize: 1000,
    defaultTTL: 300000, // 5 minutes
    storage: 'memory'
});

// Create specialized caches
const apiCache = cacheManager.namespace('api');
const entityCache = cacheManager.namespace('entities');
const metricsCache = cacheManager.namespace('metrics');
const uiCache = cacheManager.namespace('ui');

// Export global instances
window.cacheManager = cacheManager;
window.apiCache = apiCache;
window.entityCache = entityCache;
window.metricsCache = metricsCache;
window.uiCache = uiCache;