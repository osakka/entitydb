/**
 * Entity Explorer Component - Professional entity browsing and management
 * Implements modern data grid with filtering, sorting, and bulk operations
 */
class EntityExplorer extends BaseComponent {
    get defaultOptions() {
        return {
            pageSize: 50,
            allowEdit: true,
            allowDelete: true,
            allowBulkOperations: true,
            showRelationships: true,
            autoRefresh: false,
            refreshInterval: 30000
        };
    }

    get defaultState() {
        return {
            entities: [],
            filteredEntities: [],
            selectedEntities: [],
            currentPage: 1,
            totalPages: 1,
            sortField: 'created_at',
            sortDirection: 'desc',
            searchQuery: '',
            filters: {},
            loading: false,
            error: null
        };
    }

    init() {
        super.init();
        this.loadEntities();
        
        if (this.options.autoRefresh) {
            this.startAutoRefresh();
        }
    }

    setupEventListeners() {
        // Wait for render to complete before setting up listeners
        setTimeout(() => {
            // Search input
            const searchInput = this.find('.search-input');
            if (searchInput) {
                this.addEventListener(searchInput, 'input', this.debounce(this.handleSearch, 300));
            }
            
            // Filter controls
            const filterControls = this.find('.filter-controls');
            if (filterControls) {
                this.addEventListener(filterControls, 'change', this.handleFilter);
            }
            
            // Sort controls
            const sortControls = this.find('.sort-controls');
            if (sortControls) {
                this.addEventListener(sortControls, 'change', this.handleSort);
            }
            
            // Pagination
            const pagination = this.find('.pagination');
            if (pagination) {
                this.addEventListener(pagination, 'click', this.handlePagination);
            }
            
            // Bulk operations
            const bulkActions = this.find('.bulk-actions');
            if (bulkActions) {
                this.addEventListener(bulkActions, 'click', this.handleBulkAction);
            }
            
            // Row selection
            const entityTable = this.find('.entity-table');
            if (entityTable) {
                this.addEventListener(entityTable, 'change', this.handleRowSelection);
                this.addEventListener(entityTable, 'dblclick', this.handleRowDoubleClick);
            }
        }, 100);
    }

    render() {
        this.container.innerHTML = this.renderTemplate(this.getTemplate(), {
            searchQuery: this.state.searchQuery,
            totalEntities: this.state.entities.length,
            selectedCount: this.state.selectedEntities.length
        });

        this.renderTable();
        this.renderPagination();
        this.updateBulkActions();
    }

    getTemplate() {
        return `
            <div class="entity-explorer">
                <div class="explorer-header">
                    <div class="search-section">
                        <input type="text" class="search-input" placeholder="Search entities..." value="{{searchQuery}}">
                        <button class="search-btn">
                            <i class="icon-search"></i>
                        </button>
                    </div>
                    
                    <div class="filter-controls">
                        <select class="type-filter">
                            <option value="">All Types</option>
                            <option value="user">Users</option>
                            <option value="session">Sessions</option>
                            <option value="config">Configuration</option>
                        </select>
                        
                        <select class="time-filter">
                            <option value="">All Time</option>
                            <option value="today">Today</option>
                            <option value="week">This Week</option>
                            <option value="month">This Month</option>
                        </select>
                    </div>
                    
                    <div class="sort-controls">
                        <select class="sort-field">
                            <option value="created_at">Created</option>
                            <option value="updated_at">Updated</option>
                            <option value="type">Type</option>
                            <option value="name">Name</option>
                        </select>
                        
                        <button class="sort-direction" data-direction="desc">
                            <i class="icon-arrow-down"></i>
                        </button>
                    </div>
                    
                    <div class="action-buttons">
                        <button class="create-entity-btn">
                            <i class="icon-plus"></i> Create Entity
                        </button>
                        <button class="refresh-btn">
                            <i class="icon-refresh"></i>
                        </button>
                    </div>
                </div>
                
                <div class="bulk-actions" style="display: none;">
                    <span class="selection-info">{{selectedCount}} entities selected</span>
                    <button class="bulk-delete" data-action="delete">Delete</button>
                    <button class="bulk-export" data-action="export">Export</button>
                    <button class="bulk-tag" data-action="tag">Add Tag</button>
                    <button class="clear-selection" data-action="clear">Clear Selection</button>
                </div>
                
                <div class="table-container">
                    <table class="entity-table">
                        <thead>
                            <tr>
                                <th class="select-column">
                                    <input type="checkbox" class="select-all">
                                </th>
                                <th class="type-column">Type</th>
                                <th class="name-column">Name/ID</th>
                                <th class="tags-column">Tags</th>
                                <th class="created-column">Created</th>
                                <th class="size-column">Size</th>
                                <th class="actions-column">Actions</th>
                            </tr>
                        </thead>
                        <tbody class="entity-rows">
                            <!-- Dynamically populated -->
                        </tbody>
                    </table>
                </div>
                
                <div class="pagination">
                    <!-- Dynamically populated -->
                </div>
                
                <div class="loading-overlay" style="display: none;">
                    <div class="loading-spinner"></div>
                    <span>Loading entities...</span>
                </div>
            </div>
        `;
    }

    async loadEntities() {
        try {
            this.setLoading(true);
            this.setState({ error: null });
            
            const entities = await apiClient.listEntities();
            
            this.setState({ 
                entities: entities || [],
                filteredEntities: entities || []
            });
            
            this.applyFiltersAndSort();
            this.renderTable();
            
        } catch (error) {
            this.handleError(error, 'EntityExplorer.loadEntities');
            this.setState({ error: error.message });
        } finally {
            this.setLoading(false);
        }
    }

    renderTable() {
        const tbody = this.find('.entity-rows');
        if (!tbody) return;

        const startIndex = (this.state.currentPage - 1) * this.options.pageSize;
        const endIndex = startIndex + this.options.pageSize;
        const pageEntities = this.state.filteredEntities.slice(startIndex, endIndex);

        tbody.innerHTML = pageEntities.map(entity => this.renderEntityRow(entity)).join('');
    }

    renderEntityRow(entity) {
        const isSelected = this.state.selectedEntities.includes(entity.id);
        const entityName = this.getEntityName(entity);
        const entityType = this.getEntityType(entity);
        const createdDate = new Date(entity.created_at * 1000000).toLocaleDateString();
        const entitySize = this.formatBytes(entity.content?.length || 0);
        
        return `
            <tr class="entity-row ${isSelected ? 'selected' : ''}" data-entity-id="${entity.id}">
                <td class="select-column">
                    <input type="checkbox" class="row-select" ${isSelected ? 'checked' : ''}>
                </td>
                <td class="type-column">
                    <span class="entity-type ${entityType}">${entityType}</span>
                </td>
                <td class="name-column">
                    <div class="entity-name">${entityName}</div>
                    <div class="entity-id">${entity.id.substring(0, 8)}...</div>
                </td>
                <td class="tags-column">
                    <div class="tag-list">
                        ${this.renderEntityTags(entity).slice(0, 3).map(tag => 
                            `<span class="tag">${tag}</span>`
                        ).join('')}
                        ${this.renderEntityTags(entity).length > 3 ? '<span class="tag-more">+' + (this.renderEntityTags(entity).length - 3) + '</span>' : ''}
                    </div>
                </td>
                <td class="created-column">${createdDate}</td>
                <td class="size-column">${entitySize}</td>
                <td class="actions-column">
                    <button class="action-btn view-btn" data-action="view" data-entity-id="${entity.id}">
                        <i class="icon-eye"></i>
                    </button>
                    ${this.options.showRelationships ? `
                        <button class="action-btn relationships-btn" data-action="relationships" data-entity-id="${entity.id}">
                            <i class="icon-network"></i>
                        </button>
                    ` : ''}
                    ${this.options.allowEdit ? `
                        <button class="action-btn edit-btn" data-action="edit" data-entity-id="${entity.id}">
                            <i class="icon-edit"></i>
                        </button>
                    ` : ''}
                    ${this.options.allowDelete ? `
                        <button class="action-btn delete-btn" data-action="delete" data-entity-id="${entity.id}">
                            <i class="icon-trash"></i>
                        </button>
                    ` : ''}
                </td>
            </tr>
        `;
    }

    renderEntityTags(entity) {
        if (!entity.tags) return [];
        
        return entity.tags
            .filter(tag => !tag.includes('|')) // Filter out temporal timestamps
            .map(tag => {
                // Remove namespace for display
                const parts = tag.split(':');
                return parts.length > 1 ? parts.slice(1).join(':') : tag;
            });
    }

    getEntityName(entity) {
        const nameTag = entity.tags?.find(tag => tag.startsWith('name:'));
        if (nameTag) {
            return nameTag.split(':').slice(1).join(':');
        }
        return entity.id.substring(0, 12) + '...';
    }

    getEntityType(entity) {
        const typeTag = entity.tags?.find(tag => tag.startsWith('type:'));
        if (typeTag) {
            return typeTag.split(':')[1];
        }
        return 'unknown';
    }

    formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    renderPagination() {
        const pagination = this.find('.pagination');
        if (!pagination) return;

        const totalPages = Math.ceil(this.state.filteredEntities.length / this.options.pageSize);
        this.setState({ totalPages });

        if (totalPages <= 1) {
            pagination.style.display = 'none';
            return;
        }

        pagination.style.display = 'block';
        pagination.innerHTML = this.renderPaginationButtons(this.state.currentPage, totalPages);
    }

    renderPaginationButtons(currentPage, totalPages) {
        let buttons = [];
        
        // Previous button
        buttons.push(`
            <button class="page-btn prev" ${currentPage === 1 ? 'disabled' : ''} data-page="${currentPage - 1}">
                <i class="icon-chevron-left"></i> Previous
            </button>
        `);

        // Page numbers
        const startPage = Math.max(1, currentPage - 2);
        const endPage = Math.min(totalPages, currentPage + 2);

        if (startPage > 1) {
            buttons.push(`<button class="page-btn" data-page="1">1</button>`);
            if (startPage > 2) {
                buttons.push(`<span class="page-ellipsis">...</span>`);
            }
        }

        for (let i = startPage; i <= endPage; i++) {
            buttons.push(`
                <button class="page-btn ${i === currentPage ? 'active' : ''}" data-page="${i}">
                    ${i}
                </button>
            `);
        }

        if (endPage < totalPages) {
            if (endPage < totalPages - 1) {
                buttons.push(`<span class="page-ellipsis">...</span>`);
            }
            buttons.push(`<button class="page-btn" data-page="${totalPages}">${totalPages}</button>`);
        }

        // Next button
        buttons.push(`
            <button class="page-btn next" ${currentPage === totalPages ? 'disabled' : ''} data-page="${currentPage + 1}">
                Next <i class="icon-chevron-right"></i>
            </button>
        `);

        return buttons.join('');
    }

    // Event handlers
    handleSearch(event) {
        const query = event.target.value.toLowerCase();
        this.setState({ searchQuery: query, currentPage: 1 });
        this.applyFiltersAndSort();
    }

    handleFilter(event) {
        const filterType = event.target.className;
        const filterValue = event.target.value;
        
        this.setState({
            filters: { ...this.state.filters, [filterType]: filterValue },
            currentPage: 1
        });
        
        this.applyFiltersAndSort();
    }

    handleSort(event) {
        if (event.target.matches('.sort-field')) {
            this.setState({ sortField: event.target.value });
        } else if (event.target.matches('.sort-direction') || event.target.closest('.sort-direction')) {
            const btn = event.target.closest('.sort-direction');
            const currentDirection = btn.dataset.direction;
            const newDirection = currentDirection === 'asc' ? 'desc' : 'asc';
            
            btn.dataset.direction = newDirection;
            btn.innerHTML = `<i class="icon-arrow-${newDirection === 'asc' ? 'up' : 'down'}"></i>`;
            
            this.setState({ sortDirection: newDirection });
        }
        
        this.applyFiltersAndSort();
    }

    handlePagination(event) {
        if (event.target.matches('.page-btn') && !event.target.disabled) {
            const page = parseInt(event.target.dataset.page);
            if (page && page !== this.state.currentPage) {
                this.setState({ currentPage: page });
                this.renderTable();
                this.renderPagination();
            }
        }
    }

    handleRowSelection(event) {
        if (event.target.matches('.row-select')) {
            const entityId = event.target.closest('.entity-row').dataset.entityId;
            const isChecked = event.target.checked;
            
            let selectedEntities = [...this.state.selectedEntities];
            
            if (isChecked) {
                selectedEntities.push(entityId);
            } else {
                selectedEntities = selectedEntities.filter(id => id !== entityId);
            }
            
            this.setState({ selectedEntities });
            this.updateBulkActions();
            
        } else if (event.target.matches('.select-all')) {
            const isChecked = event.target.checked;
            const pageEntities = this.getCurrentPageEntities();
            
            let selectedEntities = [...this.state.selectedEntities];
            
            if (isChecked) {
                pageEntities.forEach(entity => {
                    if (!selectedEntities.includes(entity.id)) {
                        selectedEntities.push(entity.id);
                    }
                });
            } else {
                const pageEntityIds = pageEntities.map(e => e.id);
                selectedEntities = selectedEntities.filter(id => !pageEntityIds.includes(id));
            }
            
            this.setState({ selectedEntities });
            this.updateBulkActions();
            this.renderTable();
        }
    }

    handleRowDoubleClick(event) {
        const row = event.target.closest('.entity-row');
        if (row) {
            const entityId = row.dataset.entityId;
            this.emit('entity-details', { entityId });
        }
    }

    handleBulkAction(event) {
        if (event.target.matches('[data-action]')) {
            const action = event.target.dataset.action;
            const selectedIds = this.state.selectedEntities;
            
            this.emit('bulk-action', { action, entityIds: selectedIds });
            
            if (action === 'clear') {
                this.setState({ selectedEntities: [] });
                this.updateBulkActions();
                this.renderTable();
            }
        }
    }

    // Utility methods
    applyFiltersAndSort() {
        let filtered = [...this.state.entities];
        
        // Apply search filter
        if (this.state.searchQuery) {
            filtered = filtered.filter(entity => {
                const searchText = [
                    entity.id,
                    this.getEntityName(entity),
                    this.getEntityType(entity),
                    ...(entity.tags || [])
                ].join(' ').toLowerCase();
                
                return searchText.includes(this.state.searchQuery);
            });
        }
        
        // Apply type filter
        if (this.state.filters['type-filter']) {
            filtered = filtered.filter(entity => 
                this.getEntityType(entity) === this.state.filters['type-filter']
            );
        }
        
        // Apply time filter
        if (this.state.filters['time-filter']) {
            const now = Date.now();
            const filterMap = {
                'today': 24 * 60 * 60 * 1000,
                'week': 7 * 24 * 60 * 60 * 1000,
                'month': 30 * 24 * 60 * 60 * 1000
            };
            
            const timeRange = filterMap[this.state.filters['time-filter']];
            if (timeRange) {
                filtered = filtered.filter(entity => {
                    const entityTime = new Date(entity.created_at * 1000000).getTime();
                    return (now - entityTime) <= timeRange;
                });
            }
        }
        
        // Apply sorting
        filtered.sort((a, b) => {
            let aVal, bVal;
            
            switch (this.state.sortField) {
                case 'type':
                    aVal = this.getEntityType(a);
                    bVal = this.getEntityType(b);
                    break;
                case 'name':
                    aVal = this.getEntityName(a);
                    bVal = this.getEntityName(b);
                    break;
                case 'created_at':
                case 'updated_at':
                    aVal = a[this.state.sortField] || 0;
                    bVal = b[this.state.sortField] || 0;
                    break;
                default:
                    aVal = a[this.state.sortField];
                    bVal = b[this.state.sortField];
            }
            
            const modifier = this.state.sortDirection === 'asc' ? 1 : -1;
            
            if (aVal < bVal) return -1 * modifier;
            if (aVal > bVal) return 1 * modifier;
            return 0;
        });
        
        this.setState({ 
            filteredEntities: filtered,
            currentPage: Math.min(this.state.currentPage, Math.ceil(filtered.length / this.options.pageSize) || 1)
        });
        
        this.renderTable();
        this.renderPagination();
    }

    getCurrentPageEntities() {
        const startIndex = (this.state.currentPage - 1) * this.options.pageSize;
        const endIndex = startIndex + this.options.pageSize;
        return this.state.filteredEntities.slice(startIndex, endIndex);
    }

    updateBulkActions() {
        const bulkActions = this.find('.bulk-actions');
        const hasSelection = this.state.selectedEntities.length > 0;
        
        bulkActions.style.display = hasSelection ? 'block' : 'none';
        
        const selectionInfo = bulkActions.querySelector('.selection-info');
        if (selectionInfo) {
            selectionInfo.textContent = `${this.state.selectedEntities.length} entities selected`;
        }
    }

    startAutoRefresh() {
        this.refreshInterval = setInterval(() => {
            if (!this.state.loading) {
                this.loadEntities();
            }
        }, this.options.refreshInterval);
    }

    onStateChange(oldState, newState) {
        if (oldState.loading !== newState.loading) {
            const overlay = this.find('.loading-overlay');
            if (overlay) {
                overlay.style.display = newState.loading ? 'flex' : 'none';
            }
        }
    }

    destroy() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
        }
        super.destroy();
    }
}

// Export component
window.EntityExplorer = EntityExplorer;