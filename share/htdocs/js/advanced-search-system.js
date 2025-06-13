/**
 * EntityDB Advanced Search System
 * Sophisticated search with real-time filtering, suggestions, and saved queries
 */

class AdvancedSearchSystem {
    constructor() {
        this.searchHistory = JSON.parse(localStorage.getItem('entitydb-search-history') || '[]');
        this.savedQueries = JSON.parse(localStorage.getItem('entitydb-saved-queries') || '[]');
        this.activeFilters = [];
        this.searchCache = new Map();
        this.cacheTimeout = 30000; // 30 seconds
        this.debounceTimer = null;
        this.currentSuggestions = [];
        
        this.init();
    }

    init() {
        this.setupSearchFilters();
        this.setupKeyboardShortcuts();
    }

    setupSearchFilters() {
        this.filterTypes = {
            'type': {
                label: 'Type',
                icon: 'fas fa-tag',
                values: ['user', 'document', 'task', 'note', 'file', 'custom'],
                operator: 'equals'
            },
            'status': {
                label: 'Status',
                icon: 'fas fa-circle',
                values: ['active', 'inactive', 'draft', 'archived'],
                operator: 'equals'
            },
            'created': {
                label: 'Created Date',
                icon: 'fas fa-calendar',
                type: 'date-range',
                operator: 'between'
            },
            'updated': {
                label: 'Updated Date',
                icon: 'fas fa-clock',
                type: 'date-range',
                operator: 'between'
            },
            'tags': {
                label: 'Has Tags',
                icon: 'fas fa-tags',
                type: 'tag-search',
                operator: 'contains'
            },
            'content': {
                label: 'Content Contains',
                icon: 'fas fa-search',
                type: 'text',
                operator: 'contains'
            },
            'size': {
                label: 'Content Size',
                icon: 'fas fa-weight',
                type: 'number-range',
                operator: 'between'
            }
        };
    }

    setupKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            // Ctrl/Cmd + K to focus search
            if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
                e.preventDefault();
                this.focusSearch();
            }
            
            // Ctrl/Cmd + Shift + F to open advanced search
            if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === 'F') {
                e.preventDefault();
                this.openAdvancedSearch();
            }
        });
    }

    // Main search interface
    performSearch(query, entities, options = {}) {
        const cacheKey = this.generateCacheKey(query, this.activeFilters);
        
        // Check cache first
        if (this.searchCache.has(cacheKey) && !options.skipCache) {
            const cached = this.searchCache.get(cacheKey);
            if (Date.now() - cached.timestamp < this.cacheTimeout) {
                return cached.results;
            }
        }

        const results = this.executeSearch(query, entities, options);
        
        // Cache results
        this.searchCache.set(cacheKey, {
            results: results,
            timestamp: Date.now()
        });

        // Add to search history
        if (query.trim()) {
            this.addToSearchHistory(query);
        }

        return results;
    }

    executeSearch(query, entities, options = {}) {
        let results = [...entities];

        // Apply text search
        if (query.trim()) {
            results = this.applyTextSearch(query, results, options);
        }

        // Apply active filters
        results = this.applyFilters(results, options);

        // Apply sorting
        if (options.sortBy) {
            results = this.applySorting(results, options.sortBy, options.sortOrder);
        }

        return results;
    }

    applyTextSearch(query, entities, options = {}) {
        const searchTerms = this.parseSearchQuery(query);
        
        return entities.filter(entity => {
            const searchScore = this.calculateSearchScore(entity, searchTerms);
            entity._searchScore = searchScore;
            return searchScore > 0;
        }).sort((a, b) => (b._searchScore || 0) - (a._searchScore || 0));
    }

    parseSearchQuery(query) {
        const terms = [];
        const regex = /(?:"([^"]*)")|(?:(\S+):(\S+))|(\S+)/g;
        let match;

        while ((match = regex.exec(query)) !== null) {
            if (match[1]) {
                // Quoted phrase
                terms.push({ type: 'phrase', value: match[1], weight: 2 });
            } else if (match[2] && match[3]) {
                // Field search (key:value)
                terms.push({ type: 'field', key: match[2], value: match[3], weight: 3 });
            } else if (match[4]) {
                // Regular term
                terms.push({ type: 'term', value: match[4], weight: 1 });
            }
        }

        return terms;
    }

    calculateSearchScore(entity, searchTerms) {
        let score = 0;
        const searchableContent = this.getSearchableContent(entity);

        searchTerms.forEach(term => {
            let termScore = 0;

            switch (term.type) {
                case 'phrase':
                    if (searchableContent.full.includes(term.value.toLowerCase())) {
                        termScore = 10 * term.weight;
                    }
                    break;
                    
                case 'field':
                    termScore = this.scoreFieldSearch(entity, term.key, term.value);
                    break;
                    
                case 'term':
                    termScore = this.scoreTermSearch(searchableContent, term.value);
                    break;
            }

            score += termScore * term.weight;
        });

        return score;
    }

    getSearchableContent(entity) {
        let content = {
            id: entity.id || '',
            tags: [],
            text: '',
            full: ''
        };

        // Extract tags
        if (entity.tags) {
            content.tags = entity.tags.map(tag => 
                this.stripTimestamp(tag).toLowerCase()
            );
        }

        // Extract text content
        if (entity.content) {
            try {
                content.text = atob(entity.content).toLowerCase();
            } catch (e) {
                content.text = '';
            }
        }

        // Combine all searchable content
        content.full = [
            content.id,
            ...content.tags,
            content.text
        ].join(' ').toLowerCase();

        return content;
    }

    scoreFieldSearch(entity, key, value) {
        const searchableContent = this.getSearchableContent(entity);
        
        switch (key.toLowerCase()) {
            case 'id':
                return entity.id.toLowerCase().includes(value.toLowerCase()) ? 20 : 0;
                
            case 'type':
                const typeTag = searchableContent.tags.find(tag => 
                    tag.startsWith('type:') && tag.includes(value.toLowerCase())
                );
                return typeTag ? 15 : 0;
                
            case 'tag':
                const hasTag = searchableContent.tags.some(tag => 
                    tag.includes(value.toLowerCase())
                );
                return hasTag ? 10 : 0;
                
            case 'content':
                return searchableContent.text.includes(value.toLowerCase()) ? 5 : 0;
                
            default:
                // Generic field search
                return searchableContent.full.includes(`${key}:${value}`.toLowerCase()) ? 8 : 0;
        }
    }

    scoreTermSearch(searchableContent, term) {
        const termLower = term.toLowerCase();
        let score = 0;

        // ID match (highest priority)
        if (searchableContent.id.includes(termLower)) {
            score += 15;
        }

        // Tag matches
        searchableContent.tags.forEach(tag => {
            if (tag.includes(termLower)) {
                score += tag.startsWith(termLower) ? 8 : 4;
            }
        });

        // Content match
        if (searchableContent.text.includes(termLower)) {
            score += 2;
        }

        return score;
    }

    applyFilters(entities, options = {}) {
        let filtered = [...entities];

        this.activeFilters.forEach(filter => {
            filtered = this.applyFilter(filtered, filter);
        });

        return filtered;
    }

    applyFilter(entities, filter) {
        return entities.filter(entity => {
            switch (filter.type) {
                case 'type':
                    return this.filterByType(entity, filter.value);
                case 'status':
                    return this.filterByStatus(entity, filter.value);
                case 'created':
                    return this.filterByDateRange(entity, 'created_at', filter.value);
                case 'updated':
                    return this.filterByDateRange(entity, 'updated_at', filter.value);
                case 'tags':
                    return this.filterByTags(entity, filter.value);
                case 'content':
                    return this.filterByContent(entity, filter.value);
                case 'size':
                    return this.filterBySize(entity, filter.value);
                default:
                    return true;
            }
        });
    }

    filterByType(entity, typeValue) {
        if (!entity.tags) return false;
        return entity.tags.some(tag => {
            const cleanTag = this.stripTimestamp(tag);
            return cleanTag === `type:${typeValue}`;
        });
    }

    filterByStatus(entity, statusValue) {
        if (!entity.tags) return false;
        return entity.tags.some(tag => {
            const cleanTag = this.stripTimestamp(tag);
            return cleanTag === `status:${statusValue}`;
        });
    }

    filterByDateRange(entity, field, range) {
        if (!entity[field]) return false;
        
        const entityDate = new Date(entity[field] / 1000000); // Convert from nanoseconds
        const startDate = new Date(range.start);
        const endDate = new Date(range.end);
        
        return entityDate >= startDate && entityDate <= endDate;
    }

    filterByTags(entity, searchValue) {
        if (!entity.tags) return false;
        const searchLower = searchValue.toLowerCase();
        
        return entity.tags.some(tag => {
            const cleanTag = this.stripTimestamp(tag).toLowerCase();
            return cleanTag.includes(searchLower);
        });
    }

    filterByContent(entity, searchValue) {
        if (!entity.content) return false;
        
        try {
            const content = atob(entity.content).toLowerCase();
            return content.includes(searchValue.toLowerCase());
        } catch (e) {
            return false;
        }
    }

    filterBySize(entity, range) {
        if (!entity.content) return false;
        
        const size = entity.content.length;
        return size >= range.min && size <= range.max;
    }

    applySorting(entities, sortBy, sortOrder = 'desc') {
        return [...entities].sort((a, b) => {
            let aValue, bValue;

            switch (sortBy) {
                case 'relevance':
                    aValue = a._searchScore || 0;
                    bValue = b._searchScore || 0;
                    break;
                case 'created':
                    aValue = a.created_at || 0;
                    bValue = b.created_at || 0;
                    break;
                case 'updated':
                    aValue = a.updated_at || 0;
                    bValue = b.updated_at || 0;
                    break;
                case 'id':
                    aValue = a.id || '';
                    bValue = b.id || '';
                    break;
                case 'size':
                    aValue = a.content ? a.content.length : 0;
                    bValue = b.content ? b.content.length : 0;
                    break;
                default:
                    return 0;
            }

            if (sortOrder === 'asc') {
                return aValue > bValue ? 1 : aValue < bValue ? -1 : 0;
            } else {
                return aValue < bValue ? 1 : aValue > bValue ? -1 : 0;
            }
        });
    }

    // Search suggestions
    generateSuggestions(query, entities) {
        if (!query.trim()) {
            return this.getRecentSuggestions();
        }

        const suggestions = [];
        const queryLower = query.toLowerCase();

        // Add completion suggestions
        const completions = this.generateCompletions(queryLower, entities);
        suggestions.push(...completions);

        // Add field suggestions
        const fieldSuggestions = this.generateFieldSuggestions(queryLower);
        suggestions.push(...fieldSuggestions);

        // Add saved query suggestions
        const savedSuggestions = this.getSavedQuerySuggestions(queryLower);
        suggestions.push(...savedSuggestions);

        return suggestions.slice(0, 8); // Limit to 8 suggestions
    }

    generateCompletions(query, entities) {
        const completions = new Set();

        entities.forEach(entity => {
            if (entity.tags) {
                entity.tags.forEach(tag => {
                    const cleanTag = this.stripTimestamp(tag);
                    if (cleanTag.toLowerCase().includes(query)) {
                        completions.add(cleanTag);
                    }
                });
            }

            // ID completions
            if (entity.id.toLowerCase().includes(query)) {
                completions.add(entity.id);
            }
        });

        return Array.from(completions).map(completion => ({
            type: 'completion',
            text: completion,
            icon: 'fas fa-search'
        }));
    }

    generateFieldSuggestions(query) {
        const suggestions = [];

        Object.entries(this.filterTypes).forEach(([key, config]) => {
            if (key.includes(query) || config.label.toLowerCase().includes(query)) {
                suggestions.push({
                    type: 'field',
                    text: `${key}:`,
                    label: config.label,
                    icon: config.icon
                });
            }
        });

        return suggestions;
    }

    getSavedQuerySuggestions(query) {
        return this.savedQueries
            .filter(saved => saved.name.toLowerCase().includes(query))
            .map(saved => ({
                type: 'saved',
                text: saved.query,
                label: saved.name,
                icon: 'fas fa-bookmark'
            }));
    }

    getRecentSuggestions() {
        return this.searchHistory.slice(-5).reverse().map(query => ({
            type: 'recent',
            text: query,
            icon: 'fas fa-history'
        }));
    }

    // Filter management
    addFilter(filterType, value, operator = 'equals') {
        const filter = {
            id: this.generateFilterId(),
            type: filterType,
            value: value,
            operator: operator,
            label: this.generateFilterLabel(filterType, value, operator)
        };

        this.activeFilters.push(filter);
        this.saveFiltersToStorage();
        return filter;
    }

    removeFilter(filterId) {
        this.activeFilters = this.activeFilters.filter(f => f.id !== filterId);
        this.saveFiltersToStorage();
    }

    clearAllFilters() {
        this.activeFilters = [];
        this.saveFiltersToStorage();
    }

    generateFilterId() {
        return `filter_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    }

    generateFilterLabel(type, value, operator) {
        const config = this.filterTypes[type];
        if (!config) return `${type}: ${value}`;

        switch (operator) {
            case 'equals':
                return `${config.label} is ${value}`;
            case 'contains':
                return `${config.label} contains ${value}`;
            case 'between':
                return `${config.label} between ${value.start} and ${value.end}`;
            case 'greater':
                return `${config.label} > ${value}`;
            case 'less':
                return `${config.label} < ${value}`;
            default:
                return `${config.label}: ${value}`;
        }
    }

    // Search history management
    addToSearchHistory(query) {
        // Remove if already exists
        this.searchHistory = this.searchHistory.filter(q => q !== query);
        
        // Add to beginning
        this.searchHistory.unshift(query);
        
        // Keep only last 50 searches
        this.searchHistory = this.searchHistory.slice(0, 50);
        
        this.saveSearchHistoryToStorage();
    }

    clearSearchHistory() {
        this.searchHistory = [];
        this.saveSearchHistoryToStorage();
    }

    // Saved queries management
    saveQuery(name, query, filters = []) {
        const savedQuery = {
            id: this.generateFilterId(),
            name: name,
            query: query,
            filters: filters,
            created: Date.now()
        };

        this.savedQueries.push(savedQuery);
        this.saveSavedQueriesToStorage();
        return savedQuery;
    }

    loadSavedQuery(queryId) {
        const saved = this.savedQueries.find(q => q.id === queryId);
        if (saved) {
            this.activeFilters = [...saved.filters];
            return saved.query;
        }
        return '';
    }

    deleteSavedQuery(queryId) {
        this.savedQueries = this.savedQueries.filter(q => q.id !== queryId);
        this.saveSavedQueriesToStorage();
    }

    // Storage management
    saveFiltersToStorage() {
        localStorage.setItem('entitydb-active-filters', JSON.stringify(this.activeFilters));
    }

    saveSearchHistoryToStorage() {
        localStorage.setItem('entitydb-search-history', JSON.stringify(this.searchHistory));
    }

    saveSavedQueriesToStorage() {
        localStorage.setItem('entitydb-saved-queries', JSON.stringify(this.savedQueries));
    }

    // Utility methods
    generateCacheKey(query, filters) {
        return `${query}|${JSON.stringify(filters)}`;
    }

    stripTimestamp(tag) {
        if (typeof tag !== 'string') return tag;
        const pipeIndex = tag.indexOf('|');
        return pipeIndex !== -1 ? tag.substring(pipeIndex + 1) : tag;
    }

    focusSearch() {
        const searchInput = document.getElementById('entity-search');
        if (searchInput) {
            searchInput.focus();
            searchInput.select();
        }
    }

    openAdvancedSearch() {
        // This will be implemented when we create the advanced search modal
        if (window.notificationSystem) {
            window.notificationSystem.info('Advanced search dialog - coming in next phase');
        }
    }

    // Public API for integration
    getActiveFilters() {
        return [...this.activeFilters];
    }

    getSearchHistory() {
        return [...this.searchHistory];
    }

    getSavedQueries() {
        return [...this.savedQueries];
    }

    clearCache() {
        this.searchCache.clear();
    }
}

// Initialize the advanced search system
if (typeof window !== 'undefined') {
    window.AdvancedSearchSystem = AdvancedSearchSystem;
    
    // Wait for DOM to be ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            window.advancedSearchSystem = new AdvancedSearchSystem();
        });
    } else {
        window.advancedSearchSystem = new AdvancedSearchSystem();
    }
}