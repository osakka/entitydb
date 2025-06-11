// EntityDB Entity Browser Component
// A comprehensive entity browsing and management interface

const EntityBrowser = {
    name: 'EntityBrowser',
    template: `
        <div class="entity-browser">
            <!-- Browser Header -->
            <div class="browser-header">
                <div class="browser-title">
                    <h2>Entity Browser</h2>
                    <span class="browser-subtitle">
                        {{ filteredEntities.length }} of {{ totalEntities }} entities 
                        <span v-if="currentDataset">in {{ currentDataset }}</span>
                    </span>
                </div>
                
                <div class="browser-actions">
                    <button @click="createEntity" class="btn btn-primary">
                        <i class="fas fa-plus"></i> Create Entity
                    </button>
                    <button @click="refreshEntities" class="btn btn-secondary">
                        <i class="fas fa-sync" :class="{ 'fa-spin': loading }"></i> Refresh
                    </button>
                </div>
            </div>

            <!-- Search and Filters -->
            <div class="browser-controls">
                <div class="search-bar">
                    <i class="fas fa-search search-icon"></i>
                    <input 
                        v-model="searchQuery" 
                        @input="debounceSearch"
                        type="text" 
                        class="search-input" 
                        placeholder="Search entities by ID, tags, or content..."
                    >
                    <button v-if="searchQuery" @click="clearSearch" class="clear-search">
                        <i class="fas fa-times"></i>
                    </button>
                </div>

                <div class="filter-controls">
                    <div class="filter-group">
                        <label>Type:</label>
                        <select v-model="filters.type" @change="applyFilters" class="filter-select">
                            <option value="">All Types</option>
                            <option value="user">Users</option>
                            <option value="entity">Entities</option>
                            <option value="config">Config</option>
                            <option value="dashboard_layout">Dashboards</option>
                            <option value="metric">Metrics</option>
                        </select>
                    </div>

                    <div class="filter-group">
                        <label>Time Range:</label>
                        <select v-model="filters.timeRange" @change="applyFilters" class="filter-select">
                            <option value="">All Time</option>
                            <option value="1h">Last Hour</option>
                            <option value="24h">Last 24 Hours</option>
                            <option value="7d">Last 7 Days</option>
                            <option value="30d">Last 30 Days</option>
                        </select>
                    </div>

                    <div class="filter-group">
                        <label>Tags:</label>
                        <div class="tag-filter-input">
                            <input 
                                v-model="tagInput"
                                @keyup.enter="addTagFilter"
                                type="text" 
                                placeholder="Add tag filter..."
                                class="filter-input"
                            >
                            <button @click="addTagFilter" class="btn-small">
                                <i class="fas fa-plus"></i>
                            </button>
                        </div>
                    </div>

                    <div v-if="filters.tags.length > 0" class="active-filters">
                        <span v-for="tag in filters.tags" :key="tag" class="tag-chip">
                            {{ tag }}
                            <button @click="removeTagFilter(tag)" class="tag-remove">
                                <i class="fas fa-times"></i>
                            </button>
                        </span>
                    </div>
                </div>
            </div>

            <!-- View Mode Toggle -->
            <div class="view-controls">
                <div class="view-mode-toggle">
                    <button 
                        @click="viewMode = 'grid'" 
                        :class="['view-btn', { active: viewMode === 'grid' }]"
                        title="Grid View"
                    >
                        <i class="fas fa-th"></i>
                    </button>
                    <button 
                        @click="viewMode = 'list'" 
                        :class="['view-btn', { active: viewMode === 'list' }]"
                        title="List View"
                    >
                        <i class="fas fa-list"></i>
                    </button>
                    <button 
                        @click="viewMode = 'timeline'" 
                        :class="['view-btn', { active: viewMode === 'timeline' }]"
                        title="Timeline View"
                    >
                        <i class="fas fa-stream"></i>
                    </button>
                </div>

                <div class="sort-controls">
                    <label>Sort by:</label>
                    <select v-model="sortBy" @change="sortEntities" class="sort-select">
                        <option value="created">Created Date</option>
                        <option value="modified">Modified Date</option>
                        <option value="id">Entity ID</option>
                        <option value="type">Type</option>
                        <option value="size">Size</option>
                    </select>
                    <button @click="toggleSortOrder" class="sort-order-btn">
                        <i :class="sortOrder === 'asc' ? 'fas fa-arrow-up' : 'fas fa-arrow-down'"></i>
                    </button>
                </div>
            </div>

            <!-- Entity Grid View -->
            <div v-if="viewMode === 'grid'" class="entity-grid">
                <div 
                    v-for="entity in paginatedEntities" 
                    :key="entity.id"
                    @click="selectEntity(entity)"
                    :class="['entity-card', { selected: selectedEntity?.id === entity.id }]"
                >
                    <div class="entity-card-header">
                        <i :class="getEntityIcon(entity)" class="entity-icon"></i>
                        <div class="entity-type">{{ getEntityType(entity) }}</div>
                        <div class="entity-actions">
                            <button @click.stop="viewEntityHistory(entity)" class="action-btn" title="View History">
                                <i class="fas fa-history"></i>
                            </button>
                            <button @click.stop="editEntity(entity)" class="action-btn" title="Edit">
                                <i class="fas fa-edit"></i>
                            </button>
                            <button @click.stop="deleteEntity(entity)" class="action-btn danger" title="Delete">
                                <i class="fas fa-trash"></i>
                            </button>
                        </div>
                    </div>
                    
                    <div class="entity-card-body">
                        <div class="entity-id">{{ truncateId(entity.id) }}</div>
                        <div class="entity-timestamp">
                            <i class="fas fa-clock"></i> {{ formatTimestamp(entity.created) }}
                        </div>
                        
                        <div class="entity-tags">
                            <span v-for="tag in getDisplayTags(entity)" :key="tag" class="entity-tag">
                                {{ tag }}
                            </span>
                            <span v-if="entity.tags.length > 3" class="more-tags">
                                +{{ entity.tags.length - 3 }} more
                            </span>
                        </div>
                        
                        <div v-if="entity.content" class="entity-preview">
                            {{ getContentPreview(entity) }}
                        </div>
                    </div>
                </div>
            </div>

            <!-- Entity List View -->
            <div v-if="viewMode === 'list'" class="entity-list">
                <table class="entity-table">
                    <thead>
                        <tr>
                            <th @click="setSortBy('id')" class="sortable">
                                ID <i v-if="sortBy === 'id'" :class="sortOrder === 'asc' ? 'fas fa-arrow-up' : 'fas fa-arrow-down'"></i>
                            </th>
                            <th @click="setSortBy('type')" class="sortable">
                                Type <i v-if="sortBy === 'type'" :class="sortOrder === 'asc' ? 'fas fa-arrow-up' : 'fas fa-arrow-down'"></i>
                            </th>
                            <th>Tags</th>
                            <th @click="setSortBy('size')" class="sortable">
                                Size <i v-if="sortBy === 'size'" :class="sortOrder === 'asc' ? 'fas fa-arrow-up' : 'fas fa-arrow-down'"></i>
                            </th>
                            <th @click="setSortBy('created')" class="sortable">
                                Created <i v-if="sortBy === 'created'" :class="sortOrder === 'asc' ? 'fas fa-arrow-up' : 'fas fa-arrow-down'"></i>
                            </th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr 
                            v-for="entity in paginatedEntities" 
                            :key="entity.id"
                            @click="selectEntity(entity)"
                            :class="{ selected: selectedEntity?.id === entity.id }"
                        >
                            <td class="entity-id-cell">
                                <i :class="getEntityIcon(entity)"></i>
                                {{ truncateId(entity.id) }}
                            </td>
                            <td>{{ getEntityType(entity) }}</td>
                            <td>
                                <div class="tag-list">
                                    <span v-for="tag in entity.tags.slice(0, 3)" :key="tag" class="mini-tag">
                                        {{ tag }}
                                    </span>
                                    <span v-if="entity.tags.length > 3" class="more-count">
                                        +{{ entity.tags.length - 3 }}
                                    </span>
                                </div>
                            </td>
                            <td>{{ formatSize(entity.size || 0) }}</td>
                            <td>{{ formatTimestamp(entity.created) }}</td>
                            <td class="action-cell">
                                <button @click.stop="viewEntityHistory(entity)" class="action-btn" title="History">
                                    <i class="fas fa-history"></i>
                                </button>
                                <button @click.stop="editEntity(entity)" class="action-btn" title="Edit">
                                    <i class="fas fa-edit"></i>
                                </button>
                                <button @click.stop="deleteEntity(entity)" class="action-btn danger" title="Delete">
                                    <i class="fas fa-trash"></i>
                                </button>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>

            <!-- Entity Timeline View -->
            <div v-if="viewMode === 'timeline'" class="entity-timeline">
                <div v-for="(group, date) in timelineGroups" :key="date" class="timeline-group">
                    <div class="timeline-date">{{ formatDate(date) }}</div>
                    <div class="timeline-entities">
                        <div 
                            v-for="entity in group" 
                            :key="entity.id"
                            @click="selectEntity(entity)"
                            :class="['timeline-entity', { selected: selectedEntity?.id === entity.id }]"
                        >
                            <div class="timeline-time">{{ formatTime(entity.created) }}</div>
                            <div class="timeline-content">
                                <i :class="getEntityIcon(entity)"></i>
                                <span class="timeline-type">{{ getEntityType(entity) }}</span>
                                <span class="timeline-id">{{ truncateId(entity.id) }}</span>
                                <div class="timeline-tags">
                                    <span v-for="tag in entity.tags.slice(0, 2)" :key="tag" class="mini-tag">
                                        {{ tag }}
                                    </span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Pagination -->
            <div v-if="totalPages > 1" class="pagination">
                <button 
                    @click="currentPage = 1" 
                    :disabled="currentPage === 1"
                    class="page-btn"
                >
                    <i class="fas fa-angle-double-left"></i>
                </button>
                <button 
                    @click="currentPage--" 
                    :disabled="currentPage === 1"
                    class="page-btn"
                >
                    <i class="fas fa-angle-left"></i>
                </button>
                
                <span class="page-info">
                    Page {{ currentPage }} of {{ totalPages }}
                </span>
                
                <button 
                    @click="currentPage++" 
                    :disabled="currentPage === totalPages"
                    class="page-btn"
                >
                    <i class="fas fa-angle-right"></i>
                </button>
                <button 
                    @click="currentPage = totalPages" 
                    :disabled="currentPage === totalPages"
                    class="page-btn"
                >
                    <i class="fas fa-angle-double-right"></i>
                </button>

                <select v-model.number="pageSize" @change="currentPage = 1" class="page-size-select">
                    <option value="10">10 per page</option>
                    <option value="25">25 per page</option>
                    <option value="50">50 per page</option>
                    <option value="100">100 per page</option>
                </select>
            </div>

            <!-- Entity Detail Panel -->
            <transition name="slide">
                <div v-if="selectedEntity" class="entity-detail-panel">
                    <div class="detail-header">
                        <h3>Entity Details</h3>
                        <button @click="selectedEntity = null" class="close-btn">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>

                    <div class="detail-content">
                        <div class="detail-section">
                            <h4>Basic Information</h4>
                            <div class="detail-field">
                                <label>Entity ID:</label>
                                <code>{{ selectedEntity.id }}</code>
                                <button @click="copyToClipboard(selectedEntity.id)" class="copy-btn">
                                    <i class="fas fa-copy"></i>
                                </button>
                            </div>
                            <div class="detail-field">
                                <label>Type:</label>
                                <span>{{ getEntityType(selectedEntity) }}</span>
                            </div>
                            <div class="detail-field">
                                <label>Created:</label>
                                <span>{{ formatFullTimestamp(selectedEntity.created) }}</span>
                            </div>
                            <div class="detail-field">
                                <label>Size:</label>
                                <span>{{ formatSize(selectedEntity.size || 0) }}</span>
                            </div>
                        </div>

                        <div class="detail-section">
                            <h4>Tags ({{ selectedEntity.tags.length }})</h4>
                            <div class="tag-cloud">
                                <span v-for="tag in selectedEntity.tags" :key="tag" class="detail-tag">
                                    {{ tag }}
                                </span>
                            </div>
                        </div>

                        <div v-if="selectedEntity.content" class="detail-section">
                            <h4>Content</h4>
                            <div class="content-viewer">
                                <pre v-if="isJSON(selectedEntity.content)">{{ formatJSON(selectedEntity.content) }}</pre>
                                <div v-else class="content-text">{{ selectedEntity.content }}</div>
                            </div>
                        </div>

                        <div class="detail-actions">
                            <button @click="viewEntityHistory(selectedEntity)" class="btn btn-secondary">
                                <i class="fas fa-history"></i> View History
                            </button>
                            <button @click="viewEntityRelationships(selectedEntity)" class="btn btn-secondary">
                                <i class="fas fa-project-diagram"></i> Relationships
                            </button>
                            <button @click="editEntity(selectedEntity)" class="btn btn-primary">
                                <i class="fas fa-edit"></i> Edit Entity
                            </button>
                        </div>
                    </div>
                </div>
            </transition>
        </div>
    `,
    
    props: ['sessionToken', 'currentDataset', 'isDarkMode'],
    
    data() {
        return {
            entities: [],
            filteredEntities: [],
            totalEntities: 0,
            selectedEntity: null,
            loading: false,
            error: null,
            
            // Search and filters
            searchQuery: '',
            searchTimeout: null,
            filters: {
                type: '',
                timeRange: '',
                tags: []
            },
            tagInput: '',
            
            // View controls
            viewMode: 'grid', // grid, list, timeline
            sortBy: 'created',
            sortOrder: 'desc',
            
            // Pagination
            currentPage: 1,
            pageSize: 25,
            
            // Cache
            entityCache: new Map(),
            lastRefresh: null
        };
    },
    
    computed: {
        paginatedEntities() {
            const start = (this.currentPage - 1) * this.pageSize;
            const end = start + this.pageSize;
            return this.filteredEntities.slice(start, end);
        },
        
        totalPages() {
            return Math.ceil(this.filteredEntities.length / this.pageSize);
        },
        
        timelineGroups() {
            const groups = {};
            this.paginatedEntities.forEach(entity => {
                const date = new Date(entity.created).toDateString();
                if (!groups[date]) {
                    groups[date] = [];
                }
                groups[date].push(entity);
            });
            return groups;
        }
    },
    
    mounted() {
        this.loadEntities();
    },
    
    methods: {
        async loadEntities() {
            this.loading = true;
            this.error = null;
            
            try {
                const response = await fetch('/api/v1/entities/list', {
                    headers: {
                        'Authorization': `Bearer ${this.sessionToken}`
                    }
                });
                
                if (!response.ok) {
                    throw new Error('Failed to load entities');
                }
                
                const data = await response.json();
                this.entities = data.map(entity => ({
                    ...entity,
                    created: entity.created || new Date().toISOString(),
                    size: entity.content ? entity.content.length : 0
                }));
                
                this.totalEntities = this.entities.length;
                this.applyFilters();
                this.lastRefresh = new Date();
                
                // Cache entities
                this.entities.forEach(entity => {
                    this.entityCache.set(entity.id, entity);
                });
                
            } catch (error) {
                this.error = error.message;
                this.$emit('error', error);
            } finally {
                this.loading = false;
            }
        },
        
        refreshEntities() {
            this.loadEntities();
        },
        
        debounceSearch() {
            clearTimeout(this.searchTimeout);
            this.searchTimeout = setTimeout(() => {
                this.applyFilters();
            }, 300);
        },
        
        clearSearch() {
            this.searchQuery = '';
            this.applyFilters();
        },
        
        applyFilters() {
            let filtered = [...this.entities];
            
            // Search filter
            if (this.searchQuery) {
                const query = this.searchQuery.toLowerCase();
                filtered = filtered.filter(entity => {
                    return entity.id.toLowerCase().includes(query) ||
                           entity.tags.some(tag => tag.toLowerCase().includes(query)) ||
                           (entity.content && entity.content.toLowerCase().includes(query));
                });
            }
            
            // Type filter
            if (this.filters.type) {
                filtered = filtered.filter(entity => {
                    return entity.tags.some(tag => tag.includes(`type:${this.filters.type}`));
                });
            }
            
            // Time range filter
            if (this.filters.timeRange) {
                const now = new Date();
                const ranges = {
                    '1h': 60 * 60 * 1000,
                    '24h': 24 * 60 * 60 * 1000,
                    '7d': 7 * 24 * 60 * 60 * 1000,
                    '30d': 30 * 24 * 60 * 60 * 1000
                };
                const cutoff = new Date(now - ranges[this.filters.timeRange]);
                
                filtered = filtered.filter(entity => {
                    return new Date(entity.created) >= cutoff;
                });
            }
            
            // Tag filters
            if (this.filters.tags.length > 0) {
                filtered = filtered.filter(entity => {
                    return this.filters.tags.every(tag => 
                        entity.tags.some(entityTag => entityTag.includes(tag))
                    );
                });
            }
            
            this.filteredEntities = filtered;
            this.sortEntities();
            this.currentPage = 1;
        },
        
        addTagFilter() {
            if (this.tagInput && !this.filters.tags.includes(this.tagInput)) {
                this.filters.tags.push(this.tagInput);
                this.tagInput = '';
                this.applyFilters();
            }
        },
        
        removeTagFilter(tag) {
            const index = this.filters.tags.indexOf(tag);
            if (index > -1) {
                this.filters.tags.splice(index, 1);
                this.applyFilters();
            }
        },
        
        setSortBy(field) {
            if (this.sortBy === field) {
                this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
            } else {
                this.sortBy = field;
                this.sortOrder = 'asc';
            }
            this.sortEntities();
        },
        
        toggleSortOrder() {
            this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
            this.sortEntities();
        },
        
        sortEntities() {
            const multiplier = this.sortOrder === 'asc' ? 1 : -1;
            
            this.filteredEntities.sort((a, b) => {
                let aVal, bVal;
                
                switch (this.sortBy) {
                    case 'id':
                        return a.id.localeCompare(b.id) * multiplier;
                    case 'type':
                        aVal = this.getEntityType(a);
                        bVal = this.getEntityType(b);
                        return aVal.localeCompare(bVal) * multiplier;
                    case 'size':
                        return ((a.size || 0) - (b.size || 0)) * multiplier;
                    case 'created':
                    case 'modified':
                        aVal = new Date(a.created).getTime();
                        bVal = new Date(b.created).getTime();
                        return (aVal - bVal) * multiplier;
                    default:
                        return 0;
                }
            });
        },
        
        selectEntity(entity) {
            this.selectedEntity = entity;
            this.$emit('entity-selected', entity);
        },
        
        async createEntity() {
            this.$emit('create-entity');
        },
        
        async editEntity(entity) {
            this.$emit('edit-entity', entity);
        },
        
        async deleteEntity(entity) {
            if (confirm(`Are you sure you want to delete entity ${entity.id}?`)) {
                try {
                    const response = await fetch(`/api/v1/entities/delete?id=${entity.id}`, {
                        method: 'DELETE',
                        headers: {
                            'Authorization': `Bearer ${this.sessionToken}`
                        }
                    });
                    
                    if (response.ok) {
                        this.$emit('entity-deleted', entity);
                        this.loadEntities();
                    } else {
                        throw new Error('Failed to delete entity');
                    }
                } catch (error) {
                    this.$emit('error', error);
                }
            }
        },
        
        async viewEntityHistory(entity) {
            this.$emit('view-history', entity);
        },
        
        async viewEntityRelationships(entity) {
            this.$emit('view-relationships', entity);
        },
        
        // Helper methods
        getEntityType(entity) {
            const typeTag = entity.tags.find(tag => tag.startsWith('type:'));
            return typeTag ? typeTag.split(':')[1] : 'entity';
        },
        
        getEntityIcon(entity) {
            const type = this.getEntityType(entity);
            const icons = {
                user: 'fas fa-user',
                config: 'fas fa-cog',
                dashboard_layout: 'fas fa-th-large',
                metric: 'fas fa-chart-line',
                entity: 'fas fa-cube'
            };
            return icons[type] || 'fas fa-cube';
        },
        
        getDisplayTags(entity) {
            return entity.tags.slice(0, 3);
        },
        
        getContentPreview(entity) {
            if (!entity.content) return '';
            const content = typeof entity.content === 'string' 
                ? entity.content 
                : JSON.stringify(entity.content);
            return content.length > 100 
                ? content.substring(0, 100) + '...' 
                : content;
        },
        
        truncateId(id) {
            return id.length > 20 ? id.substring(0, 8) + '...' + id.substring(id.length - 8) : id;
        },
        
        formatTimestamp(timestamp) {
            const date = new Date(timestamp);
            const now = new Date();
            const diff = now - date;
            
            if (diff < 60000) return 'Just now';
            if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
            if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
            if (diff < 604800000) return `${Math.floor(diff / 86400000)}d ago`;
            
            return date.toLocaleDateString();
        },
        
        formatFullTimestamp(timestamp) {
            return new Date(timestamp).toLocaleString();
        },
        
        formatDate(dateStr) {
            return new Date(dateStr).toLocaleDateString('en-US', {
                weekday: 'long',
                year: 'numeric',
                month: 'long',
                day: 'numeric'
            });
        },
        
        formatTime(timestamp) {
            return new Date(timestamp).toLocaleTimeString('en-US', {
                hour: '2-digit',
                minute: '2-digit'
            });
        },
        
        formatSize(bytes) {
            const units = ['B', 'KB', 'MB', 'GB'];
            let i = 0;
            while (bytes > 1024 && i < units.length - 1) {
                bytes /= 1024;
                i++;
            }
            return `${Math.round(bytes * 10) / 10} ${units[i]}`;
        },
        
        isJSON(content) {
            try {
                JSON.parse(content);
                return true;
            } catch {
                return false;
            }
        },
        
        formatJSON(content) {
            try {
                return JSON.stringify(JSON.parse(content), null, 2);
            } catch {
                return content;
            }
        },
        
        async copyToClipboard(text) {
            try {
                await navigator.clipboard.writeText(text);
                this.$emit('notification', { message: 'Copied to clipboard!', type: 'success' });
            } catch (error) {
                this.$emit('notification', { message: 'Failed to copy', type: 'error' });
            }
        }
    }
};

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = EntityBrowser;
}

// Also make it available globally for browser use
if (typeof window !== 'undefined') {
    window.EntityBrowser = EntityBrowser;
}