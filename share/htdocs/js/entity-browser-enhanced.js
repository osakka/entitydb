/**
 * EntityDB Enhanced Entity Browser Component
 * Complete CRUD implementation with advanced features
 * Version: v2.30.0+
 */

class EntityBrowserEnhanced {
    constructor() {
        this.currentDataset = localStorage.getItem('entitydb.dataset') || 'default';
        this.entities = [];
        this.filteredEntities = [];
        this.selectedEntities = new Set();
        this.searchQuery = '';
        this.filterTags = [];
        this.loading = false;
        this.container = null;
        this.currentEntity = null;
        this.searchDebounceTimer = null;
        
        // Initialize modal system
        this.initModals();
    }

    async mount(container) {
        this.container = container;
        this.render();
        await this.loadEntities();
        this.setupEventListeners();
    }

    render() {
        if (!this.container) return;
        
        this.container.innerHTML = `
            <div class="entity-browser-enhanced">
                <!-- Toolbar -->
                <div class="toolbar">
                    <div class="toolbar-left">
                        <button class="btn btn-primary" onclick="entityBrowserEnhanced.showCreateDialog()">
                            <i class="fas fa-plus"></i> New Entity
                        </button>
                        <button class="btn btn-secondary" onclick="entityBrowserEnhanced.refresh()">
                            <i class="fas fa-sync"></i> Refresh
                        </button>
                        <div class="bulk-actions" id="bulk-actions" style="display: none;">
                            <button class="btn btn-outline-primary" onclick="entityBrowserEnhanced.bulkTag()">
                                <i class="fas fa-tags"></i> Tag Selected
                            </button>
                            <button class="btn btn-outline-danger" onclick="entityBrowserEnhanced.bulkDelete()">
                                <i class="fas fa-trash"></i> Delete Selected
                            </button>
                            <button class="btn btn-outline-secondary" onclick="entityBrowserEnhanced.exportSelected()">
                                <i class="fas fa-download"></i> Export Selected
                            </button>
                        </div>
                    </div>
                    <div class="toolbar-right">
                        <div class="search-container">
                            <div class="search-input-group">
                                <input 
                                    type="text" 
                                    id="entity-search" 
                                    class="form-input search-input" 
                                    placeholder="Search entities..." 
                                    value="${this.searchQuery}"
                                >
                                <button class="search-clear-btn" onclick="entityBrowserEnhanced.clearSearch()" style="display: ${this.searchQuery ? 'block' : 'none'}">
                                    <i class="fas fa-times"></i>
                                </button>
                            </div>
                            <div class="search-suggestions" id="search-suggestions"></div>
                        </div>
                        <div class="view-controls">
                            <button class="btn btn-outline-secondary view-toggle active" data-view="grid">
                                <i class="fas fa-th"></i>
                            </button>
                            <button class="btn btn-outline-secondary view-toggle" data-view="list">
                                <i class="fas fa-list"></i>
                            </button>
                        </div>
                    </div>
                </div>

                <!-- Filters -->
                <div class="filters-bar" id="filters-bar">
                    <div class="active-filters" id="active-filters"></div>
                    <button class="btn btn-sm btn-outline-secondary" onclick="entityBrowserEnhanced.showFilterDialog()">
                        <i class="fas fa-filter"></i> Add Filter
                    </button>
                </div>

                <!-- Entity List -->
                <div class="entity-list-container">
                    <div class="entity-list" id="entity-list">
                        <div class="loading-placeholder">
                            <div class="loading-content">
                                <i class="fas fa-database loading-icon"></i>
                                <p>Loading entities...</p>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Entity Count and Pagination -->
                <div class="entity-footer">
                    <div class="entity-count">
                        <span id="entity-count-text">Loading...</span>
                    </div>
                    <div class="pagination" id="pagination"></div>
                </div>
            </div>
        `;
    }

    setupEventListeners() {
        // Search input with debounce
        const searchInput = document.getElementById('entity-search');
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                clearTimeout(this.searchDebounceTimer);
                this.searchDebounceTimer = setTimeout(() => {
                    this.searchQuery = e.target.value;
                    this.filterEntities();
                    this.updateSearchSuggestions();
                }, 300);
            });

            searchInput.addEventListener('keydown', (e) => {
                if (e.key === 'Enter') {
                    e.preventDefault();
                    this.performSearch();
                }
            });
        }

        // View toggle
        document.querySelectorAll('.view-toggle').forEach(btn => {
            btn.addEventListener('click', (e) => {
                document.querySelectorAll('.view-toggle').forEach(b => b.classList.remove('active'));
                e.target.closest('button').classList.add('active');
                this.changeView(e.target.closest('button').dataset.view);
            });
        });

        // Global keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.ctrlKey || e.metaKey) {
                switch(e.key) {
                    case 'n':
                        e.preventDefault();
                        this.showCreateDialog();
                        break;
                    case 'f':
                        e.preventDefault();
                        searchInput?.focus();
                        break;
                    case 'a':
                        if (e.shiftKey) {
                            e.preventDefault();
                            this.selectAll();
                        }
                        break;
                }
            }
            if (e.key === 'Escape') {
                this.closeAllModals();
            }
        });
    }

    async loadEntities() {
        const apiClient = window.apiClient || (window.EntityDBClient ? new window.EntityDBClient() : null);
        
        if (!apiClient) {
            this.showError('API client not available');
            return;
        }

        this.loading = true;
        this.updateLoadingState();
        
        try {
            const token = localStorage.getItem('entitydb-admin-token');
            if (token && apiClient.setToken) {
                apiClient.setToken(token);
            }
            
            const response = await apiClient.listEntities({
                dataset: this.currentDataset,
                limit: 1000 // Load more for better search/filter experience
            });
            
            this.entities = response.entities || response.data || response || [];
            this.filteredEntities = [...this.entities];
            this.filterEntities();
            this.renderEntities();
            this.updateEntityCount();
        } catch (error) {
            console.error('Failed to load entities:', error);
            this.showError('Failed to load entities: ' + (error.message || 'Unknown error'));
        } finally {
            this.loading = false;
            this.updateLoadingState();
        }
    }

    filterEntities() {
        if (window.advancedSearchSystem) {
            // Use advanced search system
            this.filteredEntities = window.advancedSearchSystem.performSearch(
                this.searchQuery, 
                this.entities,
                {
                    sortBy: 'relevance',
                    sortOrder: 'desc'
                }
            );
        } else {
            // Fallback to basic search
            this.filteredEntities = this.basicFilterEntities();
        }
        
        this.renderEntities();
        this.updateEntityCount();
    }

    basicFilterEntities() {
        let filtered = [...this.entities];

        // Text search
        if (this.searchQuery.trim()) {
            const query = this.searchQuery.toLowerCase();
            filtered = filtered.filter(entity => {
                // Search in ID
                if (entity.id.toLowerCase().includes(query)) return true;
                
                // Search in tags
                if (entity.tags && entity.tags.some(tag => 
                    this.stripTimestamp(tag).toLowerCase().includes(query)
                )) return true;
                
                // Search in content (if it's text)
                if (entity.content && typeof entity.content === 'string') {
                    try {
                        const decoded = atob(entity.content);
                        if (decoded.toLowerCase().includes(query)) return true;
                    } catch (e) {
                        // Not base64 or not text content
                    }
                }
                
                return false;
            });
        }

        // Tag filters
        if (this.filterTags.length > 0) {
            filtered = filtered.filter(entity => {
                if (!entity.tags) return false;
                return this.filterTags.every(filterTag => 
                    entity.tags.some(tag => 
                        this.stripTimestamp(tag).toLowerCase().includes(filterTag.toLowerCase())
                    )
                );
            });
        }

        return filtered;
    }

    renderEntities() {
        const listContainer = document.getElementById('entity-list');
        if (!listContainer) return;

        if (this.loading) {
            return; // Loading state handled by updateLoadingState()
        }

        if (this.filteredEntities.length === 0) {
            listContainer.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-search empty-icon"></i>
                    <h3>No entities found</h3>
                    <p>Try adjusting your search or filters, or create a new entity.</p>
                    <button class="btn btn-primary" onclick="entityBrowserEnhanced.showCreateDialog()">
                        <i class="fas fa-plus"></i> Create First Entity
                    </button>
                </div>
            `;
            return;
        }

        const viewMode = document.querySelector('.view-toggle.active')?.dataset.view || 'grid';
        
        if (viewMode === 'grid') {
            this.renderGridView(listContainer);
        } else {
            this.renderListView(listContainer);
        }
    }

    renderGridView(container) {
        container.innerHTML = `
            <div class="entity-grid">
                ${this.filteredEntities.map(entity => `
                    <div class="entity-card" data-id="${entity.id}">
                        <div class="entity-card-header">
                            <input type="checkbox" class="entity-checkbox" data-id="${entity.id}" 
                                ${this.selectedEntities.has(entity.id) ? 'checked' : ''}>
                            <div class="entity-actions">
                                <button class="btn btn-sm btn-ghost" onclick="entityBrowserEnhanced.viewEntity('${entity.id}')" title="View">
                                    <i class="fas fa-eye"></i>
                                </button>
                                <button class="btn btn-sm btn-ghost" onclick="entityBrowserEnhanced.editEntity('${entity.id}')" title="Edit">
                                    <i class="fas fa-edit"></i>
                                </button>
                                <button class="btn btn-sm btn-ghost btn-danger" onclick="entityBrowserEnhanced.deleteEntity('${entity.id}')" title="Delete">
                                    <i class="fas fa-trash"></i>
                                </button>
                            </div>
                        </div>
                        <div class="entity-content">
                            <h4 class="entity-title">${this.getEntityTitle(entity)}</h4>
                            <div class="entity-meta">
                                <small class="text-muted">ID: ${entity.id}</small><br>
                                <small class="text-muted">Updated: ${this.formatDate(entity.updated_at)}</small>
                            </div>
                            <div class="entity-tags">
                                ${this.renderEntityTags(entity)}
                            </div>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
        
        this.setupEntityEventListeners();
    }

    renderListView(container) {
        container.innerHTML = `
            <div class="entity-table">
                <div class="entity-table-header">
                    <div class="entity-table-cell checkbox-cell">
                        <input type="checkbox" id="select-all" onchange="entityBrowserEnhanced.toggleSelectAll(this.checked)">
                    </div>
                    <div class="entity-table-cell">Title</div>
                    <div class="entity-table-cell">ID</div>
                    <div class="entity-table-cell">Tags</div>
                    <div class="entity-table-cell">Updated</div>
                    <div class="entity-table-cell">Actions</div>
                </div>
                ${this.filteredEntities.map(entity => `
                    <div class="entity-table-row" data-id="${entity.id}">
                        <div class="entity-table-cell checkbox-cell">
                            <input type="checkbox" class="entity-checkbox" data-id="${entity.id}" 
                                ${this.selectedEntities.has(entity.id) ? 'checked' : ''}>
                        </div>
                        <div class="entity-table-cell">
                            <strong>${this.getEntityTitle(entity)}</strong>
                        </div>
                        <div class="entity-table-cell">
                            <code class="entity-id">${entity.id}</code>
                        </div>
                        <div class="entity-table-cell">
                            ${this.renderEntityTags(entity, 3)}
                        </div>
                        <div class="entity-table-cell">
                            <span class="text-muted">${this.formatDate(entity.updated_at)}</span>
                        </div>
                        <div class="entity-table-cell">
                            <div class="entity-actions">
                                <button class="btn btn-sm btn-ghost" onclick="entityBrowserEnhanced.viewEntity('${entity.id}')" title="View">
                                    <i class="fas fa-eye"></i>
                                </button>
                                <button class="btn btn-sm btn-ghost" onclick="entityBrowserEnhanced.editEntity('${entity.id}')" title="Edit">
                                    <i class="fas fa-edit"></i>
                                </button>
                                <button class="btn btn-sm btn-ghost btn-danger" onclick="entityBrowserEnhanced.deleteEntity('${entity.id}')" title="Delete">
                                    <i class="fas fa-trash"></i>
                                </button>
                            </div>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
        
        this.setupEntityEventListeners();
    }

    setupEntityEventListeners() {
        // Entity selection
        document.querySelectorAll('.entity-checkbox').forEach(checkbox => {
            checkbox.addEventListener('change', (e) => {
                const entityId = e.target.dataset.id;
                if (e.target.checked) {
                    this.selectedEntities.add(entityId);
                } else {
                    this.selectedEntities.delete(entityId);
                }
                this.updateSelectionState();
            });
        });

        // Double-click to edit
        document.querySelectorAll('.entity-card, .entity-table-row').forEach(card => {
            card.addEventListener('dblclick', (e) => {
                if (e.target.type === 'checkbox') return;
                const entityId = card.dataset.id;
                this.editEntity(entityId);
            });
        });
    }

    // Entity CRUD Operations

    showCreateDialog() {
        this.currentEntity = null;
        this.showEntityModal('create');
    }

    async viewEntity(entityId) {
        const entity = this.entities.find(e => e.id === entityId);
        if (!entity) return;
        
        this.currentEntity = entity;
        this.showEntityModal('view');
    }

    async editEntity(entityId) {
        const entity = this.entities.find(e => e.id === entityId);
        if (!entity) return;
        
        this.currentEntity = entity;
        this.showEntityModal('edit');
    }

    async deleteEntity(entityId) {
        const entity = this.entities.find(e => e.id === entityId);
        if (!entity) return;

        const confirmed = await this.showConfirmDialog(
            'Delete Entity',
            `Are you sure you want to delete "${this.getEntityTitle(entity)}"?`,
            'This action cannot be undone.',
            'Delete',
            'btn-danger'
        );

        if (!confirmed) return;

        const apiClient = window.apiClient || (window.EntityDBClient ? new window.EntityDBClient() : null);
        if (!apiClient) {
            this.showNotification('API client not available', 'error');
            return;
        }

        try {
            const token = localStorage.getItem('entitydb-admin-token');
            if (token && apiClient.setToken) {
                apiClient.setToken(token);
            }
            
            await apiClient.deleteEntity(entityId);
            
            this.entities = this.entities.filter(e => e.id !== entityId);
            this.selectedEntities.delete(entityId);
            this.filterEntities();
            this.updateSelectionState();
            
            this.showNotification('Entity deleted successfully', 'success');
        } catch (error) {
            console.error('Failed to delete entity:', error);
            this.showNotification('Failed to delete entity: ' + (error.message || 'Unknown error'), 'error');
        }
    }

    async saveEntity(entityData, isUpdate = false) {
        const apiClient = window.apiClient || (window.EntityDBClient ? new window.EntityDBClient() : null);
        if (!apiClient) {
            this.showNotification('API client not available', 'error');
            return false;
        }

        try {
            const token = localStorage.getItem('entitydb-admin-token');
            if (token && apiClient.setToken) {
                apiClient.setToken(token);
            }

            let response;
            if (isUpdate) {
                response = await apiClient.updateEntity(entityData.id, entityData);
            } else {
                response = await apiClient.createEntity(entityData);
            }

            // Update local entities array
            if (isUpdate) {
                const index = this.entities.findIndex(e => e.id === entityData.id);
                if (index !== -1) {
                    this.entities[index] = response;
                }
            } else {
                this.entities.unshift(response);
            }

            this.filterEntities();
            this.closeAllModals();
            this.showNotification(
                isUpdate ? 'Entity updated successfully' : 'Entity created successfully', 
                'success'
            );
            return true;
        } catch (error) {
            console.error('Failed to save entity:', error);
            this.showNotification(
                'Failed to save entity: ' + (error.message || 'Unknown error'), 
                'error'
            );
            return false;
        }
    }

    // Utility Methods

    stripTimestamp(tag) {
        if (typeof tag !== 'string') return tag;
        // Remove timestamp prefix if present (e.g., "1234567890123456789|type:user" -> "type:user")
        const pipeIndex = tag.indexOf('|');
        return pipeIndex !== -1 ? tag.substring(pipeIndex + 1) : tag;
    }

    getEntityTitle(entity) {
        if (!entity.tags) return `Entity ${entity.id.substring(0, 8)}`;
        
        const titleTag = entity.tags.find(t => this.stripTimestamp(t).startsWith('title:'));
        if (titleTag) {
            return this.escapeHtml(this.stripTimestamp(titleTag).split(':').slice(1).join(':'));
        }
        
        const nameTag = entity.tags.find(t => this.stripTimestamp(t).startsWith('name:'));
        if (nameTag) {
            return this.escapeHtml(this.stripTimestamp(nameTag).split(':').slice(1).join(':'));
        }
        
        const typeTag = entity.tags.find(t => this.stripTimestamp(t).startsWith('type:'));
        if (typeTag) {
            return `${this.escapeHtml(this.stripTimestamp(typeTag).split(':').slice(1).join(':'))} ${entity.id.substring(0, 8)}`;
        }
        
        return `Entity ${entity.id.substring(0, 8)}`;
    }

    renderEntityTags(entity, maxTags = 5) {
        if (!entity.tags || entity.tags.length === 0) {
            return '<span class="text-muted">No tags</span>';
        }

        const displayTags = entity.tags.slice(0, maxTags);
        const remainingCount = entity.tags.length - maxTags;

        return displayTags.map(tag => {
            const cleanTag = this.stripTimestamp(tag);
            return `<span class="badge badge-secondary">${this.escapeHtml(cleanTag)}</span>`;
        }).join('') + 
        (remainingCount > 0 ? `<span class="text-muted">+${remainingCount} more</span>` : '');
    }

    formatDate(timestamp) {
        if (!timestamp) return 'Unknown';
        try {
            return new Date(timestamp / 1000000).toLocaleString();
        } catch (e) {
            return 'Invalid date';
        }
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    updateLoadingState() {
        const listContainer = document.getElementById('entity-list');
        if (!listContainer) return;

        if (this.loading) {
            listContainer.innerHTML = `
                <div class="loading-placeholder">
                    <div class="loading-content">
                        <i class="fas fa-spinner fa-spin loading-icon"></i>
                        <p>Loading entities...</p>
                    </div>
                </div>
            `;
        }
    }

    updateEntityCount() {
        const countElement = document.getElementById('entity-count-text');
        if (countElement) {
            const total = this.entities.length;
            const filtered = this.filteredEntities.length;
            const selected = this.selectedEntities.size;
            
            let text = `${filtered} of ${total} entities`;
            if (selected > 0) {
                text += ` (${selected} selected)`;
            }
            
            countElement.textContent = text;
        }
    }

    updateSelectionState() {
        this.updateEntityCount();
        
        const bulkActions = document.getElementById('bulk-actions');
        if (bulkActions) {
            bulkActions.style.display = this.selectedEntities.size > 0 ? 'flex' : 'none';
        }

        // Update select-all checkbox
        const selectAllCheckbox = document.getElementById('select-all');
        if (selectAllCheckbox) {
            const totalVisible = this.filteredEntities.length;
            const selectedVisible = this.filteredEntities.filter(e => this.selectedEntities.has(e.id)).length;
            
            selectAllCheckbox.checked = totalVisible > 0 && selectedVisible === totalVisible;
            selectAllCheckbox.indeterminate = selectedVisible > 0 && selectedVisible < totalVisible;
        }
    }

    // Selection methods
    selectAll() {
        this.filteredEntities.forEach(entity => {
            this.selectedEntities.add(entity.id);
        });
        this.renderEntities();
    }

    toggleSelectAll(checked) {
        if (checked) {
            this.selectAll();
        } else {
            this.clearSelection();
        }
    }

    clearSelection() {
        this.selectedEntities.clear();
        this.renderEntities();
    }

    // Search and filter methods
    clearSearch() {
        this.searchQuery = '';
        const searchInput = document.getElementById('entity-search');
        if (searchInput) {
            searchInput.value = '';
        }
        this.filterEntities();
        document.querySelector('.search-clear-btn').style.display = 'none';
    }

    performSearch() {
        this.filterEntities();
    }

    updateSearchSuggestions() {
        const suggestionsContainer = document.getElementById('search-suggestions');
        if (!suggestionsContainer) return;

        if (window.advancedSearchSystem) {
            // Use advanced search system for suggestions
            const suggestions = window.advancedSearchSystem.generateSuggestions(
                this.searchQuery, 
                this.entities
            );

            if (suggestions.length > 0) {
                suggestionsContainer.innerHTML = suggestions.map(suggestion => 
                    `<div class="search-suggestion" onclick="entityBrowserEnhanced.applySuggestion('${this.escapeHtml(suggestion.text)}')">
                        <i class="${suggestion.icon}"></i>
                        <span class="suggestion-text">${this.escapeHtml(suggestion.text)}</span>
                        ${suggestion.label ? `<small class="suggestion-label">${this.escapeHtml(suggestion.label)}</small>` : ''}
                    </div>`
                ).join('');
                suggestionsContainer.style.display = 'block';
            } else {
                suggestionsContainer.style.display = 'none';
            }
        } else {
            // Fallback to basic suggestions
            this.basicUpdateSearchSuggestions(suggestionsContainer);
        }
    }

    basicUpdateSearchSuggestions(suggestionsContainer) {
        if (!this.searchQuery.trim()) {
            suggestionsContainer.innerHTML = '';
            return;
        }

        const query = this.searchQuery.toLowerCase();
        const suggestions = new Set();

        this.entities.forEach(entity => {
            if (entity.tags) {
                entity.tags.forEach(tag => {
                    const cleanTag = this.stripTimestamp(tag);
                    if (cleanTag.toLowerCase().includes(query) && suggestions.size < 5) {
                        suggestions.add(cleanTag);
                    }
                });
            }
        });

        if (suggestions.size > 0) {
            suggestionsContainer.innerHTML = Array.from(suggestions).map(suggestion => 
                `<div class="search-suggestion" onclick="entityBrowserEnhanced.applySuggestion('${this.escapeHtml(suggestion)}')">
                    <i class="fas fa-search"></i>
                    <span class="suggestion-text">${this.escapeHtml(suggestion)}</span>
                </div>`
            ).join('');
            suggestionsContainer.style.display = 'block';
        } else {
            suggestionsContainer.style.display = 'none';
        }
    }

    applySuggestion(suggestion) {
        this.searchQuery = suggestion;
        const searchInput = document.getElementById('entity-search');
        if (searchInput) {
            searchInput.value = suggestion;
        }
        this.filterEntities();
        document.getElementById('search-suggestions').style.display = 'none';
    }

    // View methods
    changeView(viewMode) {
        this.renderEntities();
        localStorage.setItem('entitydb.viewMode', viewMode);
    }

    // Notification helper
    showNotification(message, type = 'info') {
        if (window.notificationSystem) {
            window.notificationSystem.show(message, type);
        } else {
            console.log(`${type}: ${message}`);
        }
    }

    showError(message) {
        const listContainer = document.getElementById('entity-list');
        if (!listContainer) return;

        listContainer.innerHTML = `
            <div class="error-state">
                <i class="fas fa-exclamation-triangle error-icon"></i>
                <h3>Error Loading Entities</h3>
                <p>${this.escapeHtml(message)}</p>
                <button class="btn btn-primary" onclick="entityBrowserEnhanced.refresh()">
                    <i class="fas fa-sync"></i> Try Again
                </button>
            </div>
        `;
    }

    refresh() {
        this.selectedEntities.clear();
        this.loadEntities();
    }

    // Modal system methods (to be implemented in next part)
    initModals() {
        // Initialize modal containers
        if (!document.getElementById('entity-modal-container')) {
            const modalContainer = document.createElement('div');
            modalContainer.id = 'entity-modal-container';
            document.body.appendChild(modalContainer);
        }
    }

    showEntityModal(mode) {
        if (window.entityModalSystem) {
            window.entityModalSystem.showEntityModal(mode, this.currentEntity);
        } else {
            this.showNotification(`${mode} entity modal - modal system not loaded`, 'error');
        }
    }

    async showConfirmDialog(title, message, detail, confirmText, confirmClass) {
        // Simple confirm for now, will be enhanced with custom modal
        return confirm(`${title}\n\n${message}\n${detail}`);
    }

    closeAllModals() {
        // Close any open modals
        const modals = document.querySelectorAll('.modal.show');
        modals.forEach(modal => {
            modal.classList.remove('show');
        });
    }

    createModal(id, content, options = {}) {
        const modal = document.createElement('div');
        modal.id = id;
        modal.className = `modal ${options.size || 'medium'}`;
        modal.innerHTML = `
            <div class="modal-backdrop" onclick="entityBrowserEnhanced.closeModal('${id}')"></div>
            <div class="modal-dialog">
                <div class="modal-header">
                    <h2 class="modal-title">${options.title || 'Modal'}</h2>
                    <button class="modal-close" onclick="entityBrowserEnhanced.closeModal('${id}')">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                <div class="modal-body">
                    ${content}
                </div>
            </div>
        `;

        // Add to modal container
        const container = document.getElementById('entity-modal-container') || document.body;
        container.appendChild(modal);
        
        return modal;
    }

    showModal(modal) {
        modal.classList.add('show');
        document.body.classList.add('modal-open');
    }

    closeModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.remove('show');
            setTimeout(() => {
                modal.remove();
                document.body.classList.remove('modal-open');
            }, 300);
        }
    }

    // Bulk operations
    async bulkTag() {
        if (this.selectedEntities.size === 0) {
            this.showNotification('No entities selected for bulk tagging', 'warning');
            return;
        }

        this.showBulkTagDialog();
    }

    async bulkDelete() {
        if (this.selectedEntities.size === 0) {
            this.showNotification('No entities selected for bulk delete', 'warning');
            return;
        }

        const confirmed = await this.showConfirmDialog(
            'Bulk Delete Entities',
            `Are you sure you want to delete ${this.selectedEntities.size} selected entities?`,
            'This action cannot be undone and will permanently remove all selected entities.',
            'Delete All',
            'btn-danger'
        );

        if (!confirmed) return;

        await this.performBulkDelete();
    }

    async performBulkDelete() {
        const apiClient = window.apiClient || (window.EntityDBClient ? new window.EntityDBClient() : null);
        if (!apiClient) {
            this.showNotification('API client not available', 'error');
            return;
        }

        const token = localStorage.getItem('entitydb-admin-token');
        if (token && apiClient.setToken) {
            apiClient.setToken(token);
        }

        let successCount = 0;
        let errorCount = 0;
        const totalCount = this.selectedEntities.size;

        this.showNotification(`Deleting ${totalCount} entities...`, 'info');

        // Process deletions in batches to avoid overwhelming the server
        const selectedIds = Array.from(this.selectedEntities);
        const batchSize = 5;

        for (let i = 0; i < selectedIds.length; i += batchSize) {
            const batch = selectedIds.slice(i, i + batchSize);
            
            await Promise.allSettled(
                batch.map(async (entityId) => {
                    try {
                        await apiClient.request('/api/v1/entities/delete', {
                            method: 'DELETE',
                            body: JSON.stringify({ id: entityId })
                        });
                        
                        // Remove from local entities array
                        this.entities = this.entities.filter(e => e.id !== entityId);
                        this.selectedEntities.delete(entityId);
                        successCount++;
                    } catch (error) {
                        console.error(`Failed to delete entity ${entityId}:`, error);
                        errorCount++;
                    }
                })
            );

            // Update progress
            const processed = Math.min(i + batchSize, selectedIds.length);
            this.showNotification(`Deleted ${processed}/${totalCount} entities...`, 'info');
        }

        // Final update
        this.filterEntities();
        this.updateSelectionState();

        if (errorCount === 0) {
            this.showNotification(`Successfully deleted ${successCount} entities`, 'success');
        } else if (successCount === 0) {
            this.showNotification(`Failed to delete ${errorCount} entities`, 'error');
        } else {
            this.showNotification(`Deleted ${successCount} entities, ${errorCount} failed`, 'warning');
        }
    }

    showBulkTagDialog() {
        const modalId = 'bulk-tag-dialog';
        const selectedCount = this.selectedEntities.size;
        
        const modalContent = `
            <div class="bulk-tag-dialog">
                <div class="bulk-tag-summary">
                    <h3>Add Tags to ${selectedCount} Entities</h3>
                    <p class="text-muted">Add or remove tags from all selected entities</p>
                </div>

                <form id="bulk-tag-form" class="bulk-tag-form">
                    <div class="form-section">
                        <h4 class="section-title">Add Tags</h4>
                        <div class="tag-input-container">
                            <input 
                                type="text" 
                                id="bulk-add-tag-input" 
                                class="form-input tag-input" 
                                placeholder="Enter tag to add (e.g., status:active, type:document)"
                            >
                            <div class="tag-suggestions" id="bulk-add-suggestions"></div>
                        </div>
                        <div class="tags-to-add" id="tags-to-add">
                            <div class="tags-display" style="min-height: 40px;">
                                <span class="text-muted">Tags to add will appear here</span>
                            </div>
                        </div>
                    </div>

                    <div class="form-section">
                        <h4 class="section-title">Remove Tags</h4>
                        <div class="tag-input-container">
                            <input 
                                type="text" 
                                id="bulk-remove-tag-input" 
                                class="form-input tag-input" 
                                placeholder="Enter tag to remove from all selected entities"
                            >
                            <div class="tag-suggestions" id="bulk-remove-suggestions"></div>
                        </div>
                        <div class="tags-to-remove" id="tags-to-remove">
                            <div class="tags-display" style="min-height: 40px;">
                                <span class="text-muted">Tags to remove will appear here</span>
                            </div>
                        </div>
                    </div>

                    <div class="form-section">
                        <h4 class="section-title">Common Tags in Selection</h4>
                        <div class="common-tags" id="common-tags">
                            ${this.getCommonTags()}
                        </div>
                    </div>
                </form>

                <div class="modal-actions">
                    <button type="button" class="btn btn-secondary" onclick="entityBrowserEnhanced.closeBulkTagDialog('${modalId}')">
                        Cancel
                    </button>
                    <button type="button" class="btn btn-primary" onclick="entityBrowserEnhanced.performBulkTag('${modalId}')">
                        <i class="fas fa-tags"></i> Apply Changes
                    </button>
                </div>
            </div>
        `;

        const modal = this.createModal(modalId, modalContent, {
            title: 'Bulk Tag Operations',
            size: 'large'
        });

        this.showModal(modal);
        this.setupBulkTagListeners();
    }

    getCommonTags() {
        if (this.selectedEntities.size === 0) return '<span class="text-muted">No entities selected</span>';

        const selectedEntityObjects = this.filteredEntities.filter(entity => 
            this.selectedEntities.has(entity.id)
        );

        // Find tags that appear in ALL selected entities
        const tagCounts = new Map();
        
        selectedEntityObjects.forEach(entity => {
            if (entity.tags) {
                entity.tags.forEach(tag => {
                    const cleanTag = this.stripTimestamp(tag);
                    tagCounts.set(cleanTag, (tagCounts.get(cleanTag) || 0) + 1);
                });
            }
        });

        const commonTags = Array.from(tagCounts.entries())
            .filter(([tag, count]) => count === selectedEntityObjects.length)
            .map(([tag]) => tag);

        if (commonTags.length === 0) {
            return '<span class="text-muted">No common tags found</span>';
        }

        return commonTags.map(tag => 
            `<span class="badge badge-secondary" onclick="entityBrowserEnhanced.addTagToRemove('${this.escapeHtml(tag)}')" style="cursor: pointer;" title="Click to remove from all">
                ${this.escapeHtml(tag)}
            </span>`
        ).join('');
    }

    setupBulkTagListeners() {
        const addInput = document.getElementById('bulk-add-tag-input');
        const removeInput = document.getElementById('bulk-remove-tag-input');

        if (addInput) {
            addInput.addEventListener('keydown', (e) => {
                if (e.key === 'Enter') {
                    e.preventDefault();
                    this.addTagToAdd(addInput.value.trim());
                    addInput.value = '';
                }
            });
        }

        if (removeInput) {
            removeInput.addEventListener('keydown', (e) => {
                if (e.key === 'Enter') {
                    e.preventDefault();
                    this.addTagToRemove(removeInput.value.trim());
                    removeInput.value = '';
                }
            });
        }

        // Initialize tag collections
        this.tagsToAdd = new Set();
        this.tagsToRemove = new Set();
    }

    addTagToAdd(tag) {
        if (!tag) return;
        
        this.tagsToAdd = this.tagsToAdd || new Set();
        this.tagsToAdd.add(tag);
        this.updateTagsToAddDisplay();
    }

    addTagToRemove(tag) {
        if (!tag) return;
        
        this.tagsToRemove = this.tagsToRemove || new Set();
        this.tagsToRemove.add(tag);
        this.updateTagsToRemoveDisplay();
    }

    updateTagsToAddDisplay() {
        const container = document.getElementById('tags-to-add');
        if (!container) return;

        const tagsDisplay = container.querySelector('.tags-display');
        if (this.tagsToAdd.size === 0) {
            tagsDisplay.innerHTML = '<span class="text-muted">Tags to add will appear here</span>';
        } else {
            tagsDisplay.innerHTML = Array.from(this.tagsToAdd).map(tag => 
                `<div class="tag-item">
                    ${this.escapeHtml(tag)}
                    <button type="button" class="tag-remove" onclick="entityBrowserEnhanced.removeTagFromAdd('${this.escapeHtml(tag)}')">
                        <i class="fas fa-times"></i>
                    </button>
                </div>`
            ).join('');
        }
    }

    updateTagsToRemoveDisplay() {
        const container = document.getElementById('tags-to-remove');
        if (!container) return;

        const tagsDisplay = container.querySelector('.tags-display');
        if (this.tagsToRemove.size === 0) {
            tagsDisplay.innerHTML = '<span class="text-muted">Tags to remove will appear here</span>';
        } else {
            tagsDisplay.innerHTML = Array.from(this.tagsToRemove).map(tag => 
                `<div class="tag-item">
                    ${this.escapeHtml(tag)}
                    <button type="button" class="tag-remove" onclick="entityBrowserEnhanced.removeTagFromRemove('${this.escapeHtml(tag)}')">
                        <i class="fas fa-times"></i>
                    </button>
                </div>`
            ).join('');
        }
    }

    removeTagFromAdd(tag) {
        if (this.tagsToAdd) {
            this.tagsToAdd.delete(tag);
            this.updateTagsToAddDisplay();
        }
    }

    removeTagFromRemove(tag) {
        if (this.tagsToRemove) {
            this.tagsToRemove.delete(tag);
            this.updateTagsToRemoveDisplay();
        }
    }

    async performBulkTag(modalId) {
        const tagsToAdd = Array.from(this.tagsToAdd || []);
        const tagsToRemove = Array.from(this.tagsToRemove || []);

        if (tagsToAdd.length === 0 && tagsToRemove.length === 0) {
            this.showNotification('No tag changes specified', 'warning');
            return;
        }

        const apiClient = window.apiClient || (window.EntityDBClient ? new window.EntityDBClient() : null);
        if (!apiClient) {
            this.showNotification('API client not available', 'error');
            return;
        }

        const token = localStorage.getItem('entitydb-admin-token');
        if (token && apiClient.setToken) {
            apiClient.setToken(token);
        }

        let successCount = 0;
        let errorCount = 0;
        const totalCount = this.selectedEntities.size;

        this.showNotification(`Applying tag changes to ${totalCount} entities...`, 'info');

        const selectedIds = Array.from(this.selectedEntities);
        
        for (const entityId of selectedIds) {
            try {
                const entity = this.entities.find(e => e.id === entityId);
                if (!entity) continue;

                let currentTags = entity.tags ? [...entity.tags] : [];
                
                // Remove specified tags
                tagsToRemove.forEach(tagToRemove => {
                    currentTags = currentTags.filter(tag => 
                        this.stripTimestamp(tag) !== tagToRemove
                    );
                });

                // Add new tags
                tagsToAdd.forEach(tagToAdd => {
                    // Check if tag already exists
                    const exists = currentTags.some(tag => 
                        this.stripTimestamp(tag) === tagToAdd
                    );
                    if (!exists) {
                        currentTags.push(tagToAdd);
                    }
                });

                // Update entity
                const updateData = {
                    id: entityId,
                    tags: currentTags,
                    content: entity.content || ''
                };

                const response = await apiClient.request('/api/v1/entities/update', {
                    method: 'PUT',
                    body: JSON.stringify(updateData)
                });

                // Update local entity
                const index = this.entities.findIndex(e => e.id === entityId);
                if (index !== -1) {
                    this.entities[index] = response;
                }

                successCount++;
            } catch (error) {
                console.error(`Failed to update tags for entity ${entityId}:`, error);
                errorCount++;
            }
        }

        // Update display
        this.filterEntities();
        this.closeModal(modalId);

        if (errorCount === 0) {
            this.showNotification(`Successfully updated tags for ${successCount} entities`, 'success');
        } else if (successCount === 0) {
            this.showNotification(`Failed to update tags for ${errorCount} entities`, 'error');
        } else {
            this.showNotification(`Updated ${successCount} entities, ${errorCount} failed`, 'warning');
        }
    }

    closeBulkTagDialog(modalId) {
        this.tagsToAdd = new Set();
        this.tagsToRemove = new Set();
        this.closeModal(modalId);
    }

    exportSelected() {
        if (this.selectedEntities.size === 0) {
            this.showNotification('No entities selected for export', 'warning');
            return;
        }

        // Get selected entities
        const selectedEntityObjects = this.filteredEntities.filter(entity => 
            this.selectedEntities.has(entity.id)
        );

        // Use the data export system
        if (window.dataExportSystem) {
            window.dataExportSystem.exportSelectedEntities(selectedEntityObjects);
        } else {
            this.showNotification('Export system not available', 'error');
        }
    }

    showFilterDialog() {
        this.showNotification('Advanced filters - coming soon', 'info');
    }
}

// Initialize the enhanced entity browser
if (typeof window !== 'undefined') {
    window.EntityBrowserEnhanced = EntityBrowserEnhanced;
    
    // Wait for DOM to be ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            window.entityBrowserEnhanced = new EntityBrowserEnhanced();
        });
    } else {
        window.entityBrowserEnhanced = new EntityBrowserEnhanced();
    }
}