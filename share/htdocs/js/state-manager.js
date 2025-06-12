/**
 * EntityDB State Management System
 * Centralized state management with reactive stores
 */

class StateManager {
    constructor() {
        this.stores = new Map();
        this.subscribers = new Map();
        this.middleware = [];
        this.history = [];
        this.maxHistorySize = 50;
        this.debug = false;
    }

    /**
     * Create a new store
     */
    createStore(name, initialState = {}, options = {}) {
        if (this.stores.has(name)) {
            throw new Error(`Store '${name}' already exists`);
        }

        const store = {
            name,
            state: this.deepClone(initialState),
            mutations: options.mutations || {},
            actions: options.actions || {},
            getters: options.getters || {},
            modules: options.modules || {},
            strict: options.strict !== false,
            persist: options.persist || false,
            persistKey: options.persistKey || `entitydb-store-${name}`
        };

        // Load persisted state if enabled
        if (store.persist) {
            this.loadPersistedState(store);
        }

        // Create reactive proxy
        store.proxy = this.createReactiveProxy(store);

        this.stores.set(name, store);
        this.subscribers.set(name, new Set());

        return store.proxy;
    }

    /**
     * Get a store by name
     */
    getStore(name) {
        const store = this.stores.get(name);
        return store ? store.proxy : null;
    }

    /**
     * Create reactive proxy for store
     */
    createReactiveProxy(store) {
        const self = this;
        
        return new Proxy(store.state, {
            get(target, prop) {
                // Handle getters
                if (store.getters[prop]) {
                    return store.getters[prop](store.state);
                }
                
                // Handle mutations
                if (store.mutations[prop]) {
                    return (payload) => self.commit(store.name, prop, payload);
                }
                
                // Handle actions
                if (store.actions[prop]) {
                    return (payload) => self.dispatch(store.name, prop, payload);
                }
                
                // Return state property
                return target[prop];
            },
            
            set(target, prop, value) {
                if (store.strict) {
                    console.error(`Cannot directly mutate store state. Use mutations instead.`);
                    return false;
                }
                
                const oldValue = target[prop];
                target[prop] = value;
                
                self.notifySubscribers(store.name, {
                    type: 'direct',
                    property: prop,
                    oldValue,
                    newValue: value
                });
                
                return true;
            }
        });
    }

    /**
     * Commit a mutation
     */
    commit(storeName, mutationName, payload) {
        const store = this.stores.get(storeName);
        if (!store) {
            throw new Error(`Store '${storeName}' not found`);
        }

        const mutation = store.mutations[mutationName];
        if (!mutation) {
            throw new Error(`Mutation '${mutationName}' not found in store '${storeName}'`);
        }

        // Create state snapshot for history
        const snapshot = {
            storeName,
            mutationName,
            payload,
            timestamp: Date.now(),
            prevState: this.deepClone(store.state)
        };

        // Apply middleware
        for (const mw of this.middleware) {
            if (mw.beforeMutation) {
                mw.beforeMutation(snapshot);
            }
        }

        // Execute mutation
        mutation(store.state, payload);

        // Add to history
        this.addToHistory(snapshot);

        // Persist if enabled
        if (store.persist) {
            this.persistState(store);
        }

        // Notify subscribers
        this.notifySubscribers(storeName, {
            type: 'mutation',
            mutation: mutationName,
            payload,
            state: store.state
        });

        // Apply middleware
        for (const mw of this.middleware) {
            if (mw.afterMutation) {
                mw.afterMutation(snapshot);
            }
        }

        if (this.debug) {
            console.log(`[StateManager] Mutation: ${storeName}/${mutationName}`, payload);
        }
    }

    /**
     * Dispatch an action
     */
    async dispatch(storeName, actionName, payload) {
        const store = this.stores.get(storeName);
        if (!store) {
            throw new Error(`Store '${storeName}' not found`);
        }

        const action = store.actions[actionName];
        if (!action) {
            throw new Error(`Action '${actionName}' not found in store '${storeName}'`);
        }

        if (this.debug) {
            console.log(`[StateManager] Action: ${storeName}/${actionName}`, payload);
        }

        // Create action context
        const context = {
            state: store.state,
            commit: (mutation, payload) => this.commit(storeName, mutation, payload),
            dispatch: (action, payload) => this.dispatch(storeName, action, payload),
            getters: store.getters
        };

        // Execute action
        return await action(context, payload);
    }

    /**
     * Subscribe to store changes
     */
    subscribe(storeName, callback) {
        const subscribers = this.subscribers.get(storeName);
        if (!subscribers) {
            throw new Error(`Store '${storeName}' not found`);
        }

        subscribers.add(callback);

        // Return unsubscribe function
        return () => {
            subscribers.delete(callback);
        };
    }

    /**
     * Notify all subscribers of a store
     */
    notifySubscribers(storeName, change) {
        const subscribers = this.subscribers.get(storeName);
        if (subscribers) {
            subscribers.forEach(callback => {
                try {
                    callback(change);
                } catch (error) {
                    console.error('Subscriber error:', error);
                }
            });
        }
    }

    /**
     * Add middleware
     */
    use(middleware) {
        this.middleware.push(middleware);
    }

    /**
     * Time travel to a previous state
     */
    timeTravel(index) {
        if (index < 0 || index >= this.history.length) {
            throw new Error('Invalid history index');
        }

        const snapshot = this.history[index];
        const store = this.stores.get(snapshot.storeName);
        
        if (store) {
            store.state = this.deepClone(snapshot.prevState);
            this.notifySubscribers(snapshot.storeName, {
                type: 'timeTravel',
                index,
                state: store.state
            });
        }
    }

    /**
     * Get state history
     */
    getHistory() {
        return this.history.map((item, index) => ({
            index,
            storeName: item.storeName,
            mutation: item.mutationName,
            timestamp: item.timestamp,
            payload: item.payload
        }));
    }

    /**
     * Clear history
     */
    clearHistory() {
        this.history = [];
    }

    /**
     * Add to history with size limit
     */
    addToHistory(snapshot) {
        this.history.push(snapshot);
        if (this.history.length > this.maxHistorySize) {
            this.history.shift();
        }
    }

    /**
     * Persist state to localStorage
     */
    persistState(store) {
        try {
            localStorage.setItem(store.persistKey, JSON.stringify(store.state));
        } catch (error) {
            console.error('Failed to persist state:', error);
        }
    }

    /**
     * Load persisted state from localStorage
     */
    loadPersistedState(store) {
        try {
            const saved = localStorage.getItem(store.persistKey);
            if (saved) {
                const parsed = JSON.parse(saved);
                Object.assign(store.state, parsed);
            }
        } catch (error) {
            console.error('Failed to load persisted state:', error);
        }
    }

    /**
     * Deep clone utility
     */
    deepClone(obj) {
        if (obj === null || typeof obj !== 'object') return obj;
        if (obj instanceof Date) return new Date(obj.getTime());
        if (obj instanceof Array) return obj.map(item => this.deepClone(item));
        if (obj instanceof Object) {
            const cloned = {};
            for (const key in obj) {
                if (obj.hasOwnProperty(key)) {
                    cloned[key] = this.deepClone(obj[key]);
                }
            }
            return cloned;
        }
    }

    /**
     * Create a computed property
     */
    computed(storeName, computeFn) {
        let cached = null;
        let isDirty = true;

        const unsubscribe = this.subscribe(storeName, () => {
            isDirty = true;
        });

        return {
            get value() {
                if (isDirty) {
                    const store = this.getStore(storeName);
                    cached = computeFn(store);
                    isDirty = false;
                }
                return cached;
            },
            destroy() {
                unsubscribe();
            }
        };
    }
}

// Create default stores
const stateManager = new StateManager();

// Application store
stateManager.createStore('app', {
    user: null,
    sessionToken: null,
    isAuthenticated: false,
    currentDataset: 'default',
    theme: 'light',
    locale: 'en',
    notifications: []
}, {
    mutations: {
        setUser(state, user) {
            state.user = user;
            state.isAuthenticated = !!user;
        },
        setSessionToken(state, token) {
            state.sessionToken = token;
        },
        setDataset(state, dataset) {
            state.currentDataset = dataset;
        },
        setTheme(state, theme) {
            state.theme = theme;
        },
        addNotification(state, notification) {
            state.notifications.push({
                id: Date.now(),
                timestamp: new Date(),
                ...notification
            });
        },
        removeNotification(state, id) {
            state.notifications = state.notifications.filter(n => n.id !== id);
        }
    },
    actions: {
        async login({ commit }, { username, password }) {
            try {
                const response = await apiClient.login(username, password);
                commit('setUser', response.user);
                commit('setSessionToken', response.token);
                return response;
            } catch (error) {
                throw error;
            }
        },
        logout({ commit }) {
            commit('setUser', null);
            commit('setSessionToken', null);
            localStorage.removeItem('entitydb-admin-token');
            localStorage.removeItem('entitydb-admin-user');
        }
    },
    getters: {
        isAdmin(state) {
            return state.user && state.user.tags && 
                   state.user.tags.includes('rbac:role:admin');
        }
    },
    persist: true
});

// Entity store
stateManager.createStore('entities', {
    items: [],
    selected: new Set(),
    filters: {
        search: '',
        type: '',
        tags: [],
        dateRange: { start: null, end: null }
    },
    sortBy: 'updated_at',
    sortOrder: 'desc',
    currentPage: 1,
    pageSize: 50,
    totalCount: 0,
    loading: false,
    error: null
}, {
    mutations: {
        setEntities(state, entities) {
            state.items = entities;
        },
        addEntity(state, entity) {
            state.items.unshift(entity);
            state.totalCount++;
        },
        updateEntity(state, { id, updates }) {
            const index = state.items.findIndex(e => e.id === id);
            if (index !== -1) {
                state.items[index] = { ...state.items[index], ...updates };
            }
        },
        removeEntity(state, id) {
            state.items = state.items.filter(e => e.id !== id);
            state.totalCount--;
            state.selected.delete(id);
        },
        toggleSelection(state, id) {
            if (state.selected.has(id)) {
                state.selected.delete(id);
            } else {
                state.selected.add(id);
            }
        },
        clearSelection(state) {
            state.selected.clear();
        },
        setFilters(state, filters) {
            Object.assign(state.filters, filters);
        },
        setSorting(state, { field, order }) {
            state.sortBy = field;
            state.sortOrder = order;
        },
        setPage(state, page) {
            state.currentPage = page;
        },
        setLoading(state, loading) {
            state.loading = loading;
        },
        setError(state, error) {
            state.error = error;
        }
    },
    actions: {
        async fetchEntities({ commit, state }) {
            commit('setLoading', true);
            commit('setError', null);
            
            try {
                const params = {
                    dataset: stateManager.getStore('app').currentDataset,
                    limit: state.pageSize,
                    offset: (state.currentPage - 1) * state.pageSize,
                    sort: state.sortBy,
                    order: state.sortOrder
                };
                
                // Apply filters
                if (state.filters.search) {
                    params.search = state.filters.search;
                }
                if (state.filters.type) {
                    params.tags = `type:${state.filters.type}`;
                }
                
                const response = await apiClient.get('/entities/list', params);
                commit('setEntities', response.data || []);
                
            } catch (error) {
                commit('setError', error.message);
                throw error;
            } finally {
                commit('setLoading', false);
            }
        },
        
        async createEntity({ commit }, entity) {
            const response = await apiClient.post('/entities/create', entity);
            commit('addEntity', response.data);
            return response.data;
        },
        
        async deleteSelected({ state, commit }) {
            const ids = Array.from(state.selected);
            const promises = ids.map(id => apiClient.delete(`/entities/${id}`));
            
            await Promise.all(promises);
            
            ids.forEach(id => commit('removeEntity', id));
            commit('clearSelection');
        }
    },
    getters: {
        filteredEntities(state) {
            return state.items;
        },
        selectedCount(state) {
            return state.selected.size;
        },
        hasSelection(state) {
            return state.selected.size > 0;
        }
    }
});

// Metrics store
stateManager.createStore('metrics', {
    system: {},
    performance: [],
    errors: [],
    sessions: [],
    updateInterval: 5000,
    isUpdating: false
}, {
    mutations: {
        setSystemMetrics(state, metrics) {
            state.system = metrics;
        },
        addPerformanceData(state, data) {
            state.performance.push(data);
            // Keep last 100 data points
            if (state.performance.length > 100) {
                state.performance.shift();
            }
        },
        setErrors(state, errors) {
            state.errors = errors;
        },
        setSessions(state, sessions) {
            state.sessions = sessions;
        },
        setUpdating(state, updating) {
            state.isUpdating = updating;
        }
    },
    actions: {
        async fetchSystemMetrics({ commit }) {
            try {
                const response = await apiClient.get('/system/metrics');
                commit('setSystemMetrics', response.data);
            } catch (error) {
                console.error('Failed to fetch system metrics:', error);
            }
        },
        
        async startMetricsUpdates({ state, dispatch, commit }) {
            if (state.isUpdating) return;
            
            commit('setUpdating', true);
            
            const update = async () => {
                await dispatch('fetchSystemMetrics');
                
                if (state.isUpdating) {
                    setTimeout(update, state.updateInterval);
                }
            };
            
            update();
        },
        
        stopMetricsUpdates({ commit }) {
            commit('setUpdating', false);
        }
    }
});

// Export global state manager
window.stateManager = stateManager;