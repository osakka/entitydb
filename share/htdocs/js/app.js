// Alpine.js component for entity management
function entityManager() {
    return {
        // State
        user: null,
        entities: [],
        filteredEntities: [],
        loading: false,
        error: null,
        passwordError: null,
        passwordStrength: 0,
        showCreateForm: false,
        showCreateUserForm: false,
        showChangePasswordModal: false,
        showTypeDropdown: false,
        
        // Theme management
        isDarkMode: localStorage.getItem('entitydb-theme') === 'dark',
        
        // Change password form
        changePasswordForm: {
            currentPassword: '',
            newPassword: '',
            confirmPassword: ''
        },
        
        // Active tab and view management
        activeTab: 'all',
        
        // Default entity types (will be replaced by actual types from API)
        availableTypes: ['user', 'session', 'issue', 'test', 'note'],
        
        // Helper functions for entity content extraction
        getEntityTitle(entity) {
            if (!entity || !entity.content) return this.getEntityDisplayName(entity);
            
            // If content is an array, look for title, display_name, or username
            if (Array.isArray(entity.content)) {
                for (const item of entity.content) {
                    if (item.type === 'title' || item.type === 'display_name' || item.type === 'username') {
                        return item.value;
                    }
                }
            }
            
            // If content is a string, try to parse it as JSON
            if (typeof entity.content === 'string' && entity.content.startsWith('{')) {
                try {
                    const parsed = JSON.parse(entity.content);
                    return parsed.title || parsed.display_name || parsed.username || this.getEntityDisplayName(entity);
                } catch (e) {
                    // Not valid JSON, use the display name
                    return this.getEntityDisplayName(entity);
                }
            }
            
            return this.getEntityDisplayName(entity);
        },
        
        // Get display name based on entity type or tags
        getEntityDisplayName(entity) {
            if (!entity) return 'Unknown Entity';
            
            // Check for specific entity types
            const entityType = this.getEntityType(entity);
            
            // For users, try to get username from tags
            if (entityType === 'user') {
                // Try to extract username from id:username tag
                const usernameTag = entity.tags?.find(tag => tag.startsWith('id:username:'));
                if (usernameTag) {
                    return usernameTag.substring(12); // Remove "id:username:" prefix
                }
            }
            
            // Check for name tags
            const nameTag = entity.tags?.find(tag => tag.startsWith('name:'));
            if (nameTag) {
                return nameTag.substring(5); // Remove "name:" prefix
            }
            
            // For sessions, show a more readable display
            if (entityType === 'session') {
                return `Session ${this.truncateId(entity.id)}`;
            }
            
            // Look for other descriptive tags
            const titleTag = entity.tags?.find(tag => tag.startsWith('title:'));
            if (titleTag) {
                return titleTag.substring(6);
            }
            
            // Default to entity type with shortened ID
            return `${this.capitalizeFirstLetter(entityType)} ${this.truncateId(entity.id)}`;
        },
        
        getEntityDescription(entity) {
            if (!entity || !entity.content) return this.getDescriptionFromTags(entity);
            
            // If content is an array, look for description
            if (Array.isArray(entity.content)) {
                for (const item of entity.content) {
                    if (item.type === 'description') {
                        return item.value;
                    }
                }
            }
            
            // If content is a string, try to parse it as JSON
            if (typeof entity.content === 'string' && entity.content.startsWith('{')) {
                try {
                    const parsed = JSON.parse(entity.content);
                    if (parsed.description) return parsed.description;
                    
                    // For user entities, try to use other fields
                    if (this.getEntityType(entity) === 'user') {
                        if (parsed.email) return parsed.email;
                        if (parsed.full_name) return parsed.full_name;
                    }
                    
                    return this.getDescriptionFromTags(entity);
                } catch (e) {
                    // Not valid JSON, try to extract from tags
                    return this.getDescriptionFromTags(entity);
                }
            }
            
            return this.getDescriptionFromTags(entity);
        },
        
        // Extract a description from entity tags
        getDescriptionFromTags(entity) {
            if (!entity || !entity.tags || !Array.isArray(entity.tags)) {
                return '';
            }
            
            // Check for description tag
            const descTag = entity.tags.find(tag => tag.startsWith('description:'));
            if (descTag) {
                return descTag.substring(12); // Remove "description:" prefix
            }
            
            // Get the entity type for context-specific descriptions
            const type = this.getEntityType(entity);
            
            // For different entity types, different tags might be interesting
            if (type === 'user') {
                const roleTag = entity.tags.find(tag => tag.startsWith('rbac:role:'));
                if (roleTag) {
                    return `Role: ${roleTag.substring(10)}`; // Remove "rbac:role:" prefix
                }
            }
            
            // For config entities, show namespace
            if (type === 'config') {
                const confTag = entity.tags.find(tag => tag.startsWith('conf:'));
                if (confTag) {
                    return confTag;
                }
            }
            
            // For sessions, show user
            if (type === 'session') {
                const userTag = entity.tags.find(tag => tag.startsWith('user:'));
                if (userTag) {
                    return `User: ${userTag.substring(5)}`; // Remove "user:" prefix
                }
            }
            
            // For other entities, show status if available
            const statusTag = entity.tags.find(tag => tag.startsWith('status:'));
            if (statusTag) {
                return `Status: ${statusTag.substring(7)}`; // Remove "status:" prefix
            }
            
            // No meaningful description available
            return '';
        },
        
        getEntityUsername(entity) {
            if (!entity || !entity.content) return '';
            
            // If content is an array, look for username
            if (Array.isArray(entity.content)) {
                for (const item of entity.content) {
                    if (item.type === 'username') {
                        return item.value;
                    }
                }
            }
            
            // Check for username in tags (rbac:username:value)
            if (entity.tags && Array.isArray(entity.tags)) {
                const usernameTag = entity.tags.find(tag => tag.startsWith('rbac:username:'));
                if (usernameTag) {
                    return usernameTag.split(':')[2];
                }
            }
            
            return '';
        },
        
        getEntityRoles(entity) {
            // Extract roles from tags (rbac:role:*)
            if (!entity || !entity.tags || !Array.isArray(entity.tags)) return '';
            
            const roleTags = entity.tags.filter(tag => tag.startsWith('rbac:role:'));
            const roles = roleTags.map(tag => tag.split(':')[2]);
            
            return roles.join(', ');
        },
        
        getEntityExpiresAt(entity) {
            if (!entity || !entity.content) return '';
            
            // If content is an array, look for expires_at
            if (Array.isArray(entity.content)) {
                for (const item of entity.content) {
                    if (item.type === 'expires_at') {
                        return item.value;
                    }
                }
            }
            
            // Check in tags for expires
            if (entity.tags && Array.isArray(entity.tags)) {
                const expiresTag = entity.tags.find(tag => tag.startsWith('expires:'));
                if (expiresTag) {
                    return expiresTag.split(':')[1];
                }
            }
            
            return '';
        },
        
        // Login form
        loginForm: {
            username: '',
            password: ''
        },
        
        // Create form
        createForm: {
            title: '',
            description: '',
            type: '',
            customType: '',
            parsedTags: [],
            currentTag: ''
        },
        
        // Create user form
        createUserForm: {
            username: '',
            display_name: '',
            password: '',
            roles: ['user']
        },
        
        // Tag suggestions
        tagSuggestions: [],
        
        // Edit state
        editingEntity: null,
        editForm: {
            title: '',
            description: '',
            parsedTags: [],
            currentTag: ''
        },
        
        // Filters
        filters: {
            search: '',
            type: '',
            role: ''
        },
        
        // Sorting
        sortConfig: {
            field: 'updated_at',
            direction: 'desc'
        },
        
        // Entity relationships
        relationshipsCache: {},

        // Auto-refresh
        autoRefresh: false,
        refreshInterval: null,
        refreshDelay: 30000, // 30 seconds
        availableTypes: [], // Dynamic list of available types
        availableTags: [], // Dynamic list of available tags
        
        // Initialize
        init() {
            console.log("🚀 Initializing EntityDB Dashboard");
            
            // Initialize theme
            this.initTheme();

            // Clear any stuck localStorage for debugging
            if (window.location.search.includes('clear_cache=1')) {
                console.log("🧹 Clearing storage due to clear_cache parameter");
                localStorage.removeItem('entitydb_token');
                localStorage.removeItem('entitydb_username');
            }
            
            // Check for saved token
            const token = localStorage.getItem('entitydb_token');
            const username = localStorage.getItem('entitydb_username');

            console.log("🔐 Auth check:", username ? "Username found in storage" : "No username in storage");

            if (token && username) {
                console.log("👤 Setting user from storage:", username);
                this.user = { username, token };
                this.fetchEntities();
                this.startAutoRefresh();
            } else {
                console.log("🔑 No valid credentials found, showing login form");
                // Ensure user is null
                this.user = null;
            }

            // Watch for filter changes
            this.$watch('filters', () => this.filterEntities());
        },
        
        // Tab management
        changeTab(tab) {
            console.log("📌 Changing tab to:", tab);
            this.activeTab = tab;
            this.filters.type = tab === 'all' ? '' : tab;
            console.log("🔍 Updated filter type to:", this.filters.type);
            this.filterEntities();
        },
        
        getActiveTabTitle() {
            if (this.activeTab === 'all') return 'All Entities';
            if (this.activeTab === 'user') return 'User Management';
            if (this.activeTab === 'session') return 'Active Sessions';
            
            // For dynamic entity types, format them nicely
            const capitalizedType = this.capitalizeFirstLetter(this.activeTab);
            // Add 's' if the type doesn't already end with 's'
            const pluralizedType = this.activeTab.endsWith('s') ? capitalizedType : capitalizedType + 's';
            return pluralizedType;
        },
        
        // Helper functions for entity type indicators
        getEntityType(entity) {
            if (!entity || !entity.tags || !Array.isArray(entity.tags)) return 'unknown';
            
            const typeTag = entity.tags.find(tag => tag.startsWith('type:'));
            return typeTag ? typeTag.substring(5) : 'unknown';
        },
        
        getEntityTypeClass(entity) {
            const type = this.getEntityType(entity);
            return `type-${type}`;
        },
        
        getEntityTypeIcon(entity) {
            const type = this.getEntityType(entity);
            
            // Icons for different entity types
            const icons = {
                'user': '👤',
                'session': '🔑',
                'issue': '🐛',
                'test': '🧪',
                'config': '⚙️',
                'note': '📝',
                'unknown': '❓'
            };
            
            return icons[type] || icons['unknown'];
        },
        
        // Content type helpers
        hasContent(entity) {
            if (!entity) return false;
            
            // Check for content object or string
            if (entity.content) {
                return true;
            }
            
            // Look for content-type tag
            if (entity.tags && Array.isArray(entity.tags)) {
                return entity.tags.some(tag => tag.startsWith('content-type:'));
            }
            
            return false;
        },
        
        getContentType(entity) {
            if (!entity) return 'unknown';
            
            // Check for content-type tag
            if (entity.tags && Array.isArray(entity.tags)) {
                const contentTypeTag = entity.tags.find(tag => tag.startsWith('content-type:'));
                if (contentTypeTag) {
                    return contentTypeTag.substring(13);
                }
            }
            
            // If the content is a string that looks like JSON
            if (typeof entity.content === 'string' && 
                entity.content.trim().startsWith('{') && 
                entity.content.trim().endsWith('}')) {
                return 'json';
            }
            
            // If content is an array
            if (Array.isArray(entity.content)) {
                return 'structured';
            }
            
            // Check if content is a string
            if (typeof entity.content === 'string') {
                // Check if it might be HTML
                if (entity.content.includes('<html') || 
                    entity.content.includes('<div') || 
                    entity.content.includes('<p>')) {
                    return 'html';
                }
                
                // Check if it might be markdown
                if (entity.content.includes('# ') || 
                    entity.content.includes('## ') || 
                    entity.content.includes('```')) {
                    return 'markdown';
                }
                
                return 'text';
            }
            
            // Default
            return entity.content ? 'data' : 'unknown';
        },
        
        getContentTypeIcon(entity) {
            const contentType = this.getContentType(entity);
            
            // Icons for different content types
            const icons = {
                'json': '📊',
                'html': '🌐',
                'markdown': '📝',
                'text': '📄',
                'structured': '🗂️',
                'binary': '📦',
                'data': '💾',
                'unknown': '📄'
            };
            
            return icons[contentType] || icons['unknown'];
        },
        
        // Helper functions
        truncateId(id) {
            if (!id) return '';
            if (id.length <= 12) return id;
            return id.substring(0, 6) + '...' + id.substring(id.length - 6);
        },
        
        userHasAdminRole() {
            return this.user && this.user.roles && this.user.roles.includes('admin');
        },
        
        addTagFilter(tag) {
            // Add a tag to the search filter
            const tagText = tag.toLowerCase();
            if (!this.filters.search.toLowerCase().includes(tagText)) {
                this.filters.search = this.filters.search ? `${this.filters.search} ${tag}` : tag;
                this.filterEntities();
            }
        },
        
        // Start auto-refresh
        startAutoRefresh() {
            if (this.autoRefresh && !this.refreshInterval) {
                this.refreshInterval = setInterval(() => {
                    this.fetchEntities();
                }, this.refreshDelay);
            }
        },

        // Stop auto-refresh
        stopAutoRefresh() {
            if (this.refreshInterval) {
                clearInterval(this.refreshInterval);
                this.refreshInterval = null;
            }
        },
        
        // API methods
        async apiRequest(endpoint, options = {}) {
            const defaultOptions = {
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.user?.token}`
                }
            };
            
            const response = await fetch(endpoint, { ...defaultOptions, ...options });
            
            if (!response.ok) {
                if (response.status === 401) {
                    this.logout();
                    throw new Error('Authentication required');
                }
                const errorText = await response.text();
                throw new Error(`API error: ${response.status} - ${errorText || response.statusText}`);
            }
            
            return response.json();
        },
        
        // Login
        async login() {
            this.error = null;
            try {
                console.log("⏳ Attempting login with username:", this.loginForm.username);
                
                const response = await this.apiRequest('/api/v1/auth/login', {
                    method: 'POST',
                    body: JSON.stringify(this.loginForm)
                });
                
                console.log("✅ Login API response received:", response);
                
                // Direct response format - not wrapped in status/data
                if (response.token) {
                    console.log("✅ Login successful, token received");
                    
                    // Store user info including roles if available
                    this.user = {
                        username: this.loginForm.username,
                        token: response.token,
                        roles: response.user?.roles || []
                    };
                    
                    console.log("👤 User object created:", this.user);
                    
                    // Save to localStorage
                    localStorage.setItem('entitydb_token', response.token);
                    localStorage.setItem('entitydb_username', this.loginForm.username);
                    
                    // Clear form
                    this.loginForm = { username: '', password: '' };
                    
                    // Load entities
                    console.log("🔄 Loading entities after successful login");
                    await this.fetchEntities();

                    // Start auto-refresh
                    this.startAutoRefresh();
                    
                    console.log("✨ Login and initialization complete");
                } else {
                    console.error("❌ Login response missing token:", response);
                    this.error = "Login failed: Invalid response from server";
                }
            } catch (err) {
                console.error("❌ Login error:", err);
                this.error = err.message || "Login failed";
            }
        },

        // Logout
        logout() {
            // Attempt to logout on the server
            try {
                this.apiRequest('/api/v1/auth/logout', {
                    method: 'POST'
                }).catch(() => {});
            } catch (err) {
                // Ignore errors during logout
            }
            
            this.user = null;
            this.activeTab = 'all';

            // Stop auto-refresh
            this.stopAutoRefresh();
            this.entities = [];
            this.filteredEntities = [];
            localStorage.removeItem('entitydb_token');
            localStorage.removeItem('entitydb_username');
        },
        
        // Fetch entities
        async fetchEntities() {
            console.log("⏳ Starting fetchEntities, current tab:", this.activeTab);
            this.loading = true;
            try {
                // Different endpoint based on active tab
                let endpoint = '/api/v1/entities/list';
                let params = [];
                
                if (this.activeTab !== 'all') {
                    params.push(`tag=type:${this.activeTab}`);
                }
                
                if (params.length > 0) {
                    endpoint += '?' + params.join('&');
                }
                
                console.log("🔍 Fetching entities from:", endpoint);
                const response = await this.apiRequest(endpoint);
                console.log("✅ API response received");
                
                // Handle different response formats
                if (Array.isArray(response)) {
                    console.log("📊 Response is an array with", response.length, "entities");
                    
                    // Check for missing created_at dates
                    const entitiesWithoutDates = response.filter(e => !e.created_at).length;
                    if (entitiesWithoutDates > 0) {
                        console.warn(`⚠️ Found ${entitiesWithoutDates} entities without created_at dates`);
                        
                        // Sample the first entity without a date
                        const sampleEntity = response.find(e => !e.created_at);
                        if (sampleEntity) {
                            console.log("Sample entity without date:", sampleEntity);
                        }
                    }
                    
                    this.entities = response;
                } else if (response.entities) {
                    console.log("📊 Response has entities array with", response.entities.length, "entities");
                    
                    // Check for missing created_at dates
                    const entitiesWithoutDates = response.entities.filter(e => !e.created_at).length;
                    if (entitiesWithoutDates > 0) {
                        console.warn(`⚠️ Found ${entitiesWithoutDates} entities without created_at dates`);
                        
                        // Sample the first entity without a date
                        const sampleEntity = response.entities.find(e => !e.created_at);
                        if (sampleEntity) {
                            console.log("Sample entity without date:", sampleEntity);
                        }
                    }
                    
                    this.entities = response.entities;
                } else {
                    console.log("📊 Using data property or empty array", response.data ? response.data.length : 0, "entities");
                    
                    // Check the data property if it exists
                    if (response.data && Array.isArray(response.data)) {
                        const entitiesWithoutDates = response.data.filter(e => !e.created_at).length;
                        if (entitiesWithoutDates > 0) {
                            console.warn(`⚠️ Found ${entitiesWithoutDates} entities without created_at dates`);
                            
                            // Sample the first entity without a date
                            const sampleEntity = response.data.find(e => !e.created_at);
                            if (sampleEntity) {
                                console.log("Sample entity without date:", sampleEntity);
                            }
                        }
                    }
                    
                    this.entities = response.data || [];
                }
                
                // Batch fetch all types of entities and relationships
                try {
                    console.log("🔄 Fetching all entity types and relationships");
                    
                    // Fetch all entities regardless of type for comprehensive type extraction
                    const allEntitiesResponse = await this.apiRequest('/api/v1/entities/list');
                    let allEntities = [];
                    
                    if (Array.isArray(allEntitiesResponse)) {
                        allEntities = allEntitiesResponse;
                    } else if (allEntitiesResponse.entities) {
                        allEntities = allEntitiesResponse.entities;
                    } else {
                        allEntities = allEntitiesResponse.data || [];
                    }
                    
                    console.log("📊 All entities fetched:", allEntities.length);
                    
                    // Explicitly fetch sessions to make sure we have them
                    console.log("🔑 Explicitly fetching sessions");
                    const sessionsResponse = await this.apiRequest('/api/v1/entities/list?tag=type:session');
                    
                    // Merge any found sessions into all entities list if not already there
                    let sessions = [];
                    if (Array.isArray(sessionsResponse)) {
                        sessions = sessionsResponse;
                    } else if (sessionsResponse.entities) {
                        sessions = sessionsResponse.entities;
                    } else {
                        sessions = sessionsResponse.data || [];
                    }
                    
                    if (sessions.length > 0) {
                        console.log("🔑 Found", sessions.length, "sessions");
                        // Add sessions to allEntities if they're not already there
                        sessions.forEach(session => {
                            if (!allEntities.some(e => e.id === session.id)) {
                                allEntities.push(session);
                            }
                        });
                    } else {
                        console.log("⚠️ No sessions found, which is unusual if you're logged in");
                    }
                    
                    // Attempt to fetch relationships
                    try {
                        console.log("🔗 Fetching entity relationships");
                        const relationshipsResponse = await this.apiRequest('/api/v1/entity-relationships');
                        
                        if (relationshipsResponse && Array.isArray(relationshipsResponse.relationships)) {
                            console.log("🔗 Found", relationshipsResponse.relationships.length, "relationships");
                            // Store relationships for later use
                            this.relationships = relationshipsResponse.relationships;
                            
                            // Attach relationships to entities
                            this.attachRelationshipsToEntities();
                        } else {
                            console.log("⚠️ No relationships found or unexpected response format");
                            this.relationships = [];
                        }
                    } catch (relErr) {
                        console.error("❌ Error fetching relationships:", relErr);
                        this.relationships = [];
                    }
                    
                    // Extract available types from all entities
                    console.log("🔍 Extracting types from", allEntities.length, "entities");
                    this.extractAvailableTypes(allEntities);
                    this.updateAvailableTags();
                    
                    // Debug log for first entity
                    console.log("🧩 Sample entity:", allEntities.length > 0 ? JSON.stringify(allEntities[0]) : "No entities found");
                } catch (err) {
                    console.error("❌ Error in batch fetching:", err);
                }
                
                // Try to fix missing created_at dates by looking for timestamp data
                console.log("🔄 Trying to fix missing created_at dates");
                this.fixMissingDates();
                
                console.log("⚙️ Applying filters");
                this.applyFilters();
                console.log("✅ Fetch entities complete");
            } catch (err) {
                console.error("❌ Error in fetchEntities:", err);
                this.error = err.message;
            } finally {
                this.loading = false;
            }
        },
        
        // Fix missing created_at dates
        fixMissingDates() {
            if (!this.entities || !Array.isArray(this.entities)) {
                return;
            }
            
            const entitiesWithoutDates = this.entities.filter(e => !e.created_at);
            if (entitiesWithoutDates.length === 0) {
                console.log("✅ All entities have created_at dates");
                return;
            }
            
            console.log(`🔄 Fixing ${entitiesWithoutDates.length} entities without created_at dates`);
            
            // Check for alternate date sources
            let fixedCount = 0;
            
            this.entities.forEach(entity => {
                if (entity.created_at) return; // Already has date
                
                // Try to find dates in tags
                if (entity.tags && Array.isArray(entity.tags)) {
                    // Look for timestamp tags
                    const timestampTag = entity.tags.find(tag => 
                        tag.startsWith('created:') || 
                        tag.startsWith('timestamp:') || 
                        tag.startsWith('created_at:')
                    );
                    
                    if (timestampTag) {
                        const parts = timestampTag.split(':');
                        if (parts.length >= 2) {
                            const timestamp = parts[1];
                            entity.created_at = timestamp;
                            fixedCount++;
                            return;
                        }
                    }
                }
                
                // If no date found in tags, use current time as fallback
                entity.created_at = new Date().toISOString();
                fixedCount++;
            });
            
            console.log(`✅ Fixed ${fixedCount} entities with missing dates`);
        },
            
        attachRelationshipsToEntities() {
            if (!this.relationships || !Array.isArray(this.relationships) || !this.entities) {
                return;
            }
            
            console.log("🔗 Attaching relationships to entities");
            
            this.entities.forEach(entity => {
                if (!entity.relationships) {
                    entity.relationships = [];
                }
                
                // Find relationships where this entity is the source
                const sourceRels = this.relationships.filter(rel => rel.source_id === entity.id);
                if (sourceRels.length > 0) {
                    entity.relationships.push(...sourceRels);
                }
                
                // Find relationships where this entity is the target
                const targetRels = this.relationships.filter(rel => rel.target_id === entity.id);
                if (targetRels.length > 0) {
                    // Map target relationships and mark them as incoming
                    targetRels.forEach(rel => {
                        entity.relationships.push({
                            ...rel,
                            is_incoming: true
                        });
                    });
                }
            });
        },
        
        // Extract available types from a list of entities
        extractAvailableTypes(entities) {
            console.log("Extracting types from", entities.length, "entities");
            const types = new Set();
            
            entities.forEach(entity => {
                if (entity.tags && Array.isArray(entity.tags)) {
                    entity.tags.forEach(tag => {
                        if (tag && typeof tag === 'string' && tag.startsWith('type:')) {
                            const type = tag.substring(5);
                            // Log when we find a new type
                            if (!types.has(type)) {
                                console.log("Found entity type:", type);
                            }
                            types.add(type);
                        }
                    });
                }
            });
            
            // Sort alphabetically but give priority to common types
            const commonTypesOrder = ['user', 'session', 'issue', 'test', 'config'];
            
            // Convert to array and sort
            const sortedTypes = [...types].sort((a, b) => {
                // If both are common types, sort by their position in commonTypesOrder
                const aIndex = commonTypesOrder.indexOf(a);
                const bIndex = commonTypesOrder.indexOf(b);
                
                if (aIndex !== -1 && bIndex !== -1) {
                    return aIndex - bIndex;
                }
                
                // If only a is a common type, it comes first
                if (aIndex !== -1) return -1;
                
                // If only b is a common type, it comes first
                if (bIndex !== -1) return 1;
                
                // Otherwise, alphabetical sort
                return a.localeCompare(b);
            });
            
            this.availableTypes = sortedTypes;
            console.log("Available entity types:", this.availableTypes);
        },
        
        // Extract all available tags from entities
        updateAvailableTags() {
            const allTags = [];
            
            this.entities.forEach(entity => {
                if (entity.tags && Array.isArray(entity.tags)) {
                    allTags.push(...entity.tags);
                }
            });
            
            // Create unique sorted list
            this.availableTags = [...new Set(allTags)].sort();
        },
        
        // Create entity
        async createEntity() {
            try {
                // Get the entity type
                let entityType = this.createForm.type;
                if (entityType === 'custom' && this.createForm.customType) {
                    entityType = this.createForm.customType;
                }
                
                // Prepare tags
                const tags = [...this.createForm.parsedTags];
                
                // Add type tag if not already present
                if (!tags.some(tag => tag.startsWith('type:'))) {
                    tags.push(`type:${entityType}`);
                }
                
                // Build content array
                const content = [];
                if (this.createForm.title) {
                    content.push({ type: 'title', value: this.createForm.title });
                }
                if (this.createForm.description) {
                    content.push({ type: 'description', value: this.createForm.description });
                }
                
                const payload = {
                    content: content,
                    tags: tags
                };
                
                await this.apiRequest('/api/v1/entities/create', {
                    method: 'POST',
                    body: JSON.stringify(payload)
                });
                
                // Reset form and reload
                this.cancelCreate();
                await this.fetchEntities();
            } catch (err) {
                this.error = err.message;
            }
        },
        
        // Create user
        async createUser() {
            try {
                // Prepare user payload
                const content = [];
                if (this.createUserForm.username) {
                    content.push({ type: 'username', value: this.createUserForm.username });
                }
                if (this.createUserForm.display_name) {
                    content.push({ type: 'display_name', value: this.createUserForm.display_name });
                }
                
                // Prepare tags including roles
                const tags = [
                    'type:user'
                ];
                
                // Add role tags
                this.createUserForm.roles.forEach(role => {
                    tags.push(`rbac:role:${role}`);
                });
                
                // Add username tag for easier querying
                tags.push(`rbac:username:${this.createUserForm.username}`);
                
                const payload = {
                    username: this.createUserForm.username,
                    password: this.createUserForm.password,
                    display_name: this.createUserForm.display_name,
                    roles: this.createUserForm.roles,
                    content,
                    tags
                };
                
                await this.apiRequest('/api/v1/users/create', {
                    method: 'POST',
                    body: JSON.stringify(payload)
                });
                
                // Reset form and reload
                this.cancelCreateUser();
                await this.fetchEntities();
            } catch (err) {
                this.error = err.message;
            }
        },
        
        // Change current user's password
        async changeCurrentUserPassword() {
            this.passwordError = null;
            
            try {
                // Validate passwords match
                if (this.changePasswordForm.newPassword !== this.changePasswordForm.confirmPassword) {
                    this.passwordError = "New passwords don't match";
                    return;
                }
                
                if (this.changePasswordForm.newPassword.length < 4) {
                    this.passwordError = "Password must be at least 4 characters long";
                    return;
                }
                
                console.log("🔑 Changing password for current user:", this.user.username);
                
                // Use the change-password endpoint if it exists, otherwise use reset-password
                let endpoint = '/api/v1/users/change-password';
                let payload = {
                    username: this.user.username,
                    current_password: this.changePasswordForm.currentPassword,
                    new_password: this.changePasswordForm.newPassword
                };
                
                try {
                    await this.apiRequest(endpoint, {
                        method: 'POST',
                        body: JSON.stringify(payload)
                    });
                    
                    // Reset the form
                    this.changePasswordForm = {
                        currentPassword: '',
                        newPassword: '',
                        confirmPassword: ''
                    };
                    
                    this.showChangePasswordModal = false;
                    
                    // Show success message
                    alert('Password changed successfully');
                } catch (err) {
                    // If the endpoint doesn't exist, try the reset-password endpoint as admin
                    if (err.message.includes('404') || err.message.includes('Not Found')) {
                        console.log("⚠️ Change password endpoint not found, trying reset-password");
                        
                        // Check if user has admin role
                        if (!this.userHasAdminRole()) {
                            this.passwordError = "You need admin privileges to change your password";
                            return;
                        }
                        
                        // Find the current user entity
                        const userEntity = this.entities.find(entity => 
                            this.getEntityUsername(entity) === this.user.username
                        );
                        
                        if (!userEntity) {
                            this.passwordError = "Could not find your user account";
                            return;
                        }
                        
                        await this.apiRequest('/api/v1/users/reset-password', {
                            method: 'POST',
                            body: JSON.stringify({
                                user_id: userEntity.id,
                                username: this.user.username,
                                password: this.changePasswordForm.newPassword
                            })
                        });
                        
                        // Reset the form
                        this.changePasswordForm = {
                            currentPassword: '',
                            newPassword: '',
                            confirmPassword: ''
                        };
                        
                        this.showChangePasswordModal = false;
                        
                        // Show success message
                        alert('Password reset successfully');
                    } else {
                        // Other error
                        throw err;
                    }
                }
            } catch (err) {
                console.error("❌ Password change error:", err);
                this.passwordError = err.message || "Failed to change password";
            }
        },
        
        // Reset user password (for admin)
        async resetUserPassword(user) {
            if (!confirm(`Are you sure you want to reset the password for user "${this.getEntityUsername(user)}"?`)) {
                return;
            }
            
            try {
                const username = this.getEntityUsername(user);
                const newPassword = prompt('Enter new password:');
                
                if (!newPassword) return;
                
                await this.apiRequest('/api/v1/users/reset-password', {
                    method: 'POST',
                    body: JSON.stringify({
                        user_id: user.id,
                        username,
                        password: newPassword
                    })
                });
                
                alert('Password reset successfully.');
            } catch (err) {
                this.error = err.message;
                alert('Failed to reset password: ' + err.message);
            }
        },
        
        // Cancel create
        cancelCreate() {
            this.showCreateForm = false;
            this.createForm = {
                title: '',
                description: '',
                type: '',
                customType: '',
                parsedTags: [],
                currentTag: ''
            };
            this.tagSuggestions = [];
        },
        
        // Cancel create user
        cancelCreateUser() {
            this.showCreateUserForm = false;
            this.createUserForm = {
                username: '',
                display_name: '',
                password: '',
                roles: ['user']
            };
            this.passwordStrength = 0;
        },
        
        // Password strength checker
        checkPasswordStrength() {
            const password = this.createUserForm.password;
            
            if (!password) {
                this.passwordStrength = 0;
                return;
            }
            
            let strength = 0;
            
            // Length contribution (up to 30%)
            strength += Math.min(30, (password.length * 3));
            
            // Character variety contribution
            if (password.match(/[A-Z]/)) strength += 10; // Uppercase
            if (password.match(/[a-z]/)) strength += 10; // Lowercase
            if (password.match(/[0-9]/)) strength += 10; // Numbers
            if (password.match(/[^A-Za-z0-9]/)) strength += 15; // Special characters
            
            // Length bonus for longer passwords
            if (password.length > 8) strength += 15;
            if (password.length > 12) strength += 10;
            
            // Cap at 100
            this.passwordStrength = Math.min(100, strength);
        },
        
        // Get password strength color
        getPasswordStrengthColor() {
            if (this.passwordStrength < 30) return '#e74c3c'; // Weak - Red
            if (this.passwordStrength < 60) return '#f39c12'; // Medium - Orange
            if (this.passwordStrength < 80) return '#f1c40f'; // Good - Yellow
            return '#2ecc71'; // Strong - Green
        },
        
        // Get password strength text
        getPasswordStrengthText() {
            if (this.passwordStrength === 0) return 'Enter a password';
            if (this.passwordStrength < 30) return 'Very weak';
            if (this.passwordStrength < 60) return 'Weak';
            if (this.passwordStrength < 80) return 'Good';
            return 'Strong';
        },
        
        // Check change password strength
        checkChangePasswordStrength() {
            const password = this.changePasswordForm.newPassword;
            
            if (!password) {
                this.passwordStrength = 0;
                return;
            }
            
            let strength = 0;
            
            // Length contribution (up to 30%)
            strength += Math.min(30, (password.length * 3));
            
            // Character variety contribution
            if (password.match(/[A-Z]/)) strength += 10; // Uppercase
            if (password.match(/[a-z]/)) strength += 10; // Lowercase
            if (password.match(/[0-9]/)) strength += 10; // Numbers
            if (password.match(/[^A-Za-z0-9]/)) strength += 15; // Special characters
            
            // Length bonus for longer passwords
            if (password.length > 8) strength += 15;
            if (password.length > 12) strength += 10;
            
            // Cap at 100
            this.passwordStrength = Math.min(100, strength);
        },
        
        // Add tag
        addTag() {
            const tag = this.createForm.currentTag.trim();
            if (tag && !this.createForm.parsedTags.includes(tag)) {
                this.createForm.parsedTags.push(tag);
                this.createForm.currentTag = '';
                this.tagSuggestions = [];
            }
        },
        
        // Remove tag
        removeTag(index) {
            this.createForm.parsedTags.splice(index, 1);
        },
        
        // Update tag suggestions
        updateTagSuggestions() {
            const input = this.createForm.currentTag.toLowerCase();
            if (input.length < 1) {
                this.tagSuggestions = [];
                return;
            }
            
            // Filter available tags based on input
            this.tagSuggestions = this.availableTags
                .filter(tag => 
                    tag.toLowerCase().includes(input) && 
                    !this.createForm.parsedTags.includes(tag)
                )
                .slice(0, 5);
        },
        
        // Select suggestion
        selectSuggestion(suggestion) {
            this.createForm.currentTag = suggestion;
            this.addTag();
        },
        
        // Apply filters
        applyFilters() {
            console.log("🔍 Applying filters:", this.filters);
            console.log("📊 Starting with", this.entities.length, "entities");
            
            let filtered = this.entities;
            
            // Handle special tab cases directly
            if (this.activeTab === 'users') {
                console.log("👥 User tab active, filtering to user entities");
                filtered = filtered.filter(entity => 
                    entity.tags && entity.tags.includes('type:user')
                );
            } else if (this.activeTab === 'sessions') {
                console.log("🔑 Session tab active, filtering to session entities");
                filtered = filtered.filter(entity => 
                    entity.tags && entity.tags.includes('type:session')
                );
            } else if (this.activeTab !== 'all' && this.activeTab) {
                console.log(`🏷️ Tab ${this.activeTab} active, filtering by type:${this.activeTab}`);
                const typeTag = `type:${this.activeTab}`;
                filtered = filtered.filter(entity => 
                    entity.tags && entity.tags.includes(typeTag)
                );
            }
            
            // Additional type filter (if different from tab)
            if (this.filters.type && (this.activeTab === 'all' || `type:${this.filters.type}` !== `type:${this.activeTab}`)) {
                console.log(`🏷️ Additional type filter: ${this.filters.type}`);
                const typeTag = `type:${this.filters.type}`;
                filtered = filtered.filter(entity => 
                    entity.tags && entity.tags.includes(typeTag)
                );
            }
            
            // Search filter
            if (this.filters.search) {
                console.log(`🔍 Applying search filter: ${this.filters.search}`);
                const searchTerms = this.filters.search.toLowerCase().split(' ');
                
                filtered = filtered.filter(entity => {
                    const title = this.getEntityTitle(entity)?.toLowerCase() || '';
                    const description = this.getEntityDescription(entity)?.toLowerCase() || '';
                    const username = this.getEntityUsername(entity)?.toLowerCase() || '';
                    const tags = entity.tags?.join(' ').toLowerCase() || '';
                    
                    // Multiple search terms - all must match
                    return searchTerms.every(term => {
                        return title.includes(term) || 
                               description.includes(term) || 
                               username.includes(term) || 
                               tags.includes(term);
                    });
                });
            }
            
            // Role filter (for user tab)
            if (this.activeTab === 'users' && this.filters.role) {
                console.log(`👤 Filtering users by role: ${this.filters.role}`);
                const roleTag = `rbac:role:${this.filters.role}`;
                filtered = filtered.filter(entity => 
                    entity.tags && entity.tags.includes(roleTag)
                );
            }
            
            console.log("📊 Filtered to", filtered.length, "entities");
            this.filteredEntities = filtered;
            
            // Apply sorting
            this.sortEntities();
        },
        
        // Filter entities using applyFilters and ensure UI updates
        filterEntities() {
            console.log("🧹 Running filterEntities()");
            this.applyFilters();
            console.log("✅ Filtering complete, entities filtered:", this.filteredEntities.length);
        },
        
        // Sort entities
        sortEntities() {
            // Clone the array to avoid mutating the original
            const sorted = [...this.filteredEntities];
            
            const field = this.sortConfig.field;
            const direction = this.sortConfig.direction;
            
            sorted.sort((a, b) => {
                let valueA, valueB;
                
                if (field === 'title') {
                    valueA = this.getEntityTitle(a)?.toLowerCase() || '';
                    valueB = this.getEntityTitle(b)?.toLowerCase() || '';
                } else if (field === 'created_at' || field === 'updated_at') {
                    valueA = new Date(a[field] || a.created_at || 0).getTime();
                    valueB = new Date(b[field] || b.created_at || 0).getTime();
                } else {
                    valueA = a[field] || '';
                    valueB = b[field] || '';
                }
                
                if (direction === 'asc') {
                    return valueA > valueB ? 1 : valueA < valueB ? -1 : 0;
                } else {
                    return valueA < valueB ? 1 : valueA > valueB ? -1 : 0;
                }
            });
            
            this.filteredEntities = sorted;
        },
        
        // Toggle sort direction
        toggleSortDirection() {
            this.sortConfig.direction = this.sortConfig.direction === 'asc' ? 'desc' : 'asc';
            this.sortEntities();
        },
        
        // Relationship functions
        hasRelationships(entity) {
            if (!entity) return false;
            
            // Check for relationship tags (legacy format)
            const hasRelTags = entity.tags && Array.isArray(entity.tags) && 
                               entity.tags.some(tag => tag.startsWith('rel:'));
            
            // Check for direct relationship data
            const hasRelData = this.hasRelationshipData(entity);
            
            return hasRelTags || hasRelData;
        },
        
        hasRelationshipData(entity) {
            if (!entity || !entity.relationships) return false;
            
            return Array.isArray(entity.relationships) && entity.relationships.length > 0;
        },
        
        // Helper function to check if entities of a given type exist
        hasEntitiesOfType(type) {
            if (!this.entities || !Array.isArray(this.entities)) return false;
            
            const typeTag = `type:${type}`;
            return this.entities.some(entity => 
                entity.tags && Array.isArray(entity.tags) && entity.tags.includes(typeTag)
            );
        },
        
        // Helper to capitalize first letter of a string
        capitalizeFirstLetter(string) {
            if (!string) return '';
            return string.charAt(0).toUpperCase() + string.slice(1);
        },
        
        // Get entity relationships
        getEntityRelationships(entity) {
            if (!entity) return [];
            
            const relationships = [];
            
            // Safely check both data sources for relationships
            try {
                // Check direct relationship data if available (preferred format)
                if (entity.relationships && Array.isArray(entity.relationships)) {
                    // These relationships were already processed in attachRelationshipsToEntities
                    return entity.relationships;
                }
                
                // Fallback: Check for relationship tags (rel:type:id) - legacy format
                if (entity.tags && Array.isArray(entity.tags)) {
                    entity.tags.forEach(tag => {
                        if (tag && tag.startsWith('rel:')) {
                            const parts = tag.split(':');
                            if (parts.length >= 3) {
                                const type = parts[1];
                                const id = parts.slice(2).join(':'); // Join remaining parts in case ID contains colons
                                
                                if (id) {
                                    relationships.push({
                                        id: `${entity.id || 'unknown'}_${type}_${id}`,
                                        relationship_type: type,
                                        target_id: id,
                                        is_incoming: tag.includes(':from:')
                                    });
                                }
                            }
                        }
                    });
                }
            } catch (err) {
                console.error("Error processing relationships:", err);
            }
            
            return relationships;
        },
        
        // View related entity
        async viewRelatedEntity(relationship) {
            try {
                // Determine the related entity ID based on relationship direction
                const relatedEntityId = relationship.is_incoming 
                    ? relationship.source_id 
                    : (relationship.target_id || relationship.entity_id);
                
                if (!relatedEntityId) {
                    console.error("❌ Cannot view related entity: No valid ID found in relationship", relationship);
                    return;
                }
                
                console.log("🔍 Viewing related entity:", relatedEntityId);
                
                // Find entity in current list
                const relatedEntity = this.entities.find(entity => entity.id === relatedEntityId);
                
                if (relatedEntity) {
                    console.log("✅ Related entity found in cache");
                    // Entity exists, highlight it
                    this.highlightEntity(relatedEntity);
                } else {
                    console.log("🔄 Fetching related entity by ID");
                    // Fetch entity
                    this.fetchEntityById(relatedEntityId);
                }
            } catch (err) {
                console.error("❌ Error viewing related entity:", err);
                this.error = err.message;
            }
        },
        
        // Download entity content
        downloadEntityContent(entity) {
            try {
                if (!entity || !this.hasContent(entity)) {
                    console.error("❌ Cannot download: Entity has no content");
                    return;
                }
                
                console.log("💾 Downloading content for entity:", entity.id);
                
                // Determine file name
                const entityType = this.getEntityType(entity);
                const contentType = this.getContentType(entity);
                const entityTitle = this.getEntityTitle(entity) || 'untitled';
                const safeTitle = entityTitle.replace(/[^a-z0-9]/gi, '_').toLowerCase();
                
                // Create base filename
                let fileName = `${safeTitle}_${entityType}`;
                
                // Determine file extension and content
                let fileContent;
                let fileExtension;
                let mimeType;
                
                if (typeof entity.content === 'string') {
                    // Handle string content
                    fileContent = entity.content;
                    
                    // Determine extension based on content type
                    switch (contentType) {
                        case 'json':
                            fileExtension = 'json';
                            mimeType = 'application/json';
                            break;
                        case 'html':
                            fileExtension = 'html';
                            mimeType = 'text/html';
                            break;
                        case 'markdown':
                            fileExtension = 'md';
                            mimeType = 'text/markdown';
                            break;
                        default:
                            fileExtension = 'txt';
                            mimeType = 'text/plain';
                    }
                } else if (Array.isArray(entity.content)) {
                    // For array content, convert to JSON
                    fileContent = JSON.stringify(entity.content, null, 2);
                    fileExtension = 'json';
                    mimeType = 'application/json';
                } else if (typeof entity.content === 'object') {
                    // For object content, convert to JSON
                    fileContent = JSON.stringify(entity.content, null, 2);
                    fileExtension = 'json';
                    mimeType = 'application/json';
                } else {
                    // Fallback for other content types
                    fileContent = String(entity.content);
                    fileExtension = 'txt';
                    mimeType = 'text/plain';
                }
                
                // Combine filename and extension
                fileName = `${fileName}.${fileExtension}`;
                
                // Create a Blob with the content
                const blob = new Blob([fileContent], { type: mimeType });
                
                // Create a download link
                const url = URL.createObjectURL(blob);
                const link = document.createElement('a');
                link.href = url;
                link.download = fileName;
                
                // Append to document, click, and clean up
                document.body.appendChild(link);
                link.click();
                setTimeout(() => {
                    document.body.removeChild(link);
                    URL.revokeObjectURL(url);
                }, 100);
                
                console.log("✅ Download initiated for", fileName);
            } catch (err) {
                console.error("❌ Error downloading entity content:", err);
                this.error = "Failed to download content: " + err.message;
            }
        },
        
        // Highlight entity
        highlightEntity(entity) {
            // Switch to the entity's type tab
            const typeTag = entity.tags?.find(tag => tag.startsWith('type:'));
            if (typeTag) {
                const type = typeTag.substring(5);
                if (type === 'user') {
                    this.changeTab('users');
                } else if (type === 'session') {
                    this.changeTab('sessions');
                } else if (this.availableTypes.includes(type)) {
                    this.changeTab(type);
                } else {
                    this.changeTab('all');
                }
            } else {
                this.changeTab('all');
            }
            
            // Clear filters to ensure entity is visible
            this.filters.search = '';
            this.applyFilters();
            
            // Find the entity card DOM element and scroll to it
            setTimeout(() => {
                const entityCard = document.querySelector(`[data-entity-id="${entity.id}"]`);
                if (entityCard) {
                    entityCard.scrollIntoView({ behavior: 'smooth', block: 'center' });
                    entityCard.classList.add('highlight');
                    setTimeout(() => {
                        entityCard.classList.remove('highlight');
                    }, 2000);
                }
            }, 100);
        },
        
        // Fetch entity by ID
        async fetchEntityById(id) {
            try {
                const response = await this.apiRequest(`/api/v1/entities/get?id=${id}`);
                
                if (response) {
                    // Add to entities if not already there
                    if (!this.entities.some(e => e.id === response.id)) {
                        this.entities.push(response);
                    }
                    
                    // Update available tags
                    this.updateAvailableTags();
                    
                    // Switch to all tab and clear filters
                    this.changeTab('all');
                    this.filters.search = '';
                    this.applyFilters();
                    
                    // Highlight entity
                    this.highlightEntity(response);
                }
            } catch (err) {
                this.error = err.message;
            }
        },
        
        // Format date
        formatDate(dateString) {
            if (!dateString) return 'Unknown';
            
            try {
                // Handle both ISO strings and timestamp numbers
                let date;
                
                // If it looks like a timestamp (milliseconds or nanoseconds)
                if (typeof dateString === 'number' || (typeof dateString === 'string' && !isNaN(dateString))) {
                    // If it's a nanosecond timestamp (too large for JavaScript)
                    if (dateString > 1000000000000000) {
                        date = new Date(dateString / 1000000); // Convert to milliseconds
                    } else if (dateString > 1000000000000) {
                        date = new Date(Number(dateString)); // Normal millisecond timestamp
                    } else {
                        date = new Date(Number(dateString) * 1000); // Second timestamp
                    }
                } else {
                    // Regular date string
                    date = new Date(dateString);
                }
                
                // Check if date is valid
                if (isNaN(date.getTime())) {
                    console.warn("Invalid date format:", dateString);
                    return 'Unknown';
                }
                
                // Format date nicely
                const now = new Date();
                const isToday = date.toDateString() === now.toDateString();
                
                if (isToday) {
                    // For today, show just the time
                    return 'Today ' + date.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
                } else {
                    // For other dates, show date and time
                    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
                }
            } catch (e) {
                console.error("Date formatting error:", e, "for input:", dateString);
                return 'Unknown';
            }
        },
        
        // Start editing an entity
        startEdit(entity) {
            this.editingEntity = { ...entity };
            this.editForm = {
                title: this.getEntityTitle(entity),
                description: this.getEntityDescription(entity),
                parsedTags: [...(entity.tags || [])],
                currentTag: ''
            };
            
            // Focus the title input after the next render
            setTimeout(() => {
                const input = document.querySelector('.edit-title');
                if (input) input.focus();
            }, 50);
        },
        
        // Cancel editing
        cancelEdit() {
            this.editingEntity = null;
            this.editForm = {
                title: '',
                description: '',
                parsedTags: [],
                currentTag: ''
            };
        },
        
        // Save edited entity
        async saveEdit() {
            try {
                // Extract type from tags
                const tags = [...this.editForm.parsedTags];
                let entityType = null;
                
                // Find the type tag
                tags.forEach(tag => {
                    if (tag.startsWith('type:')) {
                        entityType = tag.substring(5);
                    }
                });
                
                // If no type found, use original type or default to 'note'
                if (!entityType) {
                    const origTypeTag = this.editingEntity.tags?.find(t => t.startsWith('type:'));
                    entityType = origTypeTag ? origTypeTag.substring(5) : 'note';
                    tags.push(`type:${entityType}`);
                }
                
                // Build content array
                const content = [];
                
                // Add title
                if (this.editForm.title) {
                    if (entityType === 'user') {
                        content.push({ type: 'display_name', value: this.editForm.title });
                    } else {
                        content.push({ type: 'title', value: this.editForm.title });
                    }
                }
                
                // Add description
                if (this.editForm.description) {
                    content.push({ type: 'description', value: this.editForm.description });
                }
                
                // For user entities, preserve username
                if (entityType === 'user') {
                    const username = this.getEntityUsername(this.editingEntity);
                    if (username) {
                        content.push({ type: 'username', value: username });
                    }
                }
                
                const payload = {
                    id: this.editingEntity.id,
                    content: content,
                    tags: tags
                };
                
                await this.apiRequest('/api/v1/entities/update', {
                    method: 'PUT',
                    body: JSON.stringify(payload)
                });
                
                this.cancelEdit();
                await this.fetchEntities();
            } catch (err) {
                this.error = err.message;
            }
        },
        
        // Add tag during edit
        addEditTag() {
            const tag = this.editForm.currentTag.trim();
            if (tag && !this.editForm.parsedTags.includes(tag)) {
                this.editForm.parsedTags.push(tag);
                this.editForm.currentTag = '';
            }
        },
        
        // Remove tag during edit
        removeEditTag(index) {
            this.editForm.parsedTags.splice(index, 1);
        },
        
        // Filter entities by tag when clicking on a tag
        filterByTag(tag) {
            if (tag.startsWith('type:')) {
                const type = tag.substring(5);
                if (this.availableTypes.includes(type)) {
                    this.changeTab(type);
                    return;
                }
            }
            
            this.addTagFilter(tag);
        },
        
        // Copy text to clipboard with notification
        copyToClipboard(text) {
            if (!text) return;
            
            // Create temporary input element
            const input = document.createElement('input');
            input.value = text;
            document.body.appendChild(input);
            
            // Select and copy
            input.select();
            document.execCommand('copy');
            
            // Clean up
            document.body.removeChild(input);
            
            // Show notification
            const notification = document.createElement('div');
            notification.textContent = 'ID copied to clipboard!';
            notification.style.position = 'fixed';
            notification.style.bottom = '20px';
            notification.style.left = '50%';
            notification.style.transform = 'translateX(-50%)';
            notification.style.backgroundColor = '#2ecc71';
            notification.style.color = 'white';
            notification.style.padding = '8px 16px';
            notification.style.borderRadius = '4px';
            notification.style.zIndex = '9999';
            notification.style.boxShadow = '0 2px 8px rgba(0,0,0,0.2)';
            notification.style.transition = 'opacity 0.3s ease';
            
            document.body.appendChild(notification);
            
            // Remove after a delay
            setTimeout(() => {
                notification.style.opacity = '0';
                setTimeout(() => {
                    document.body.removeChild(notification);
                }, 300);
            }, 2000);
        },
        
        // Log app state for debugging
        logAppState() {
            console.log("=================== APP STATE ===================");
            console.log("Active Tab:", this.activeTab);
            console.log("Available Types:", this.availableTypes);
            console.log("Show Type Dropdown:", this.showTypeDropdown);
            console.log("Entity Count:", this.entities.length);
            console.log("Filtered Entity Count:", this.filteredEntities.length);
            
            if (this.entities.length > 0) {
                console.log("Sample Entity:", this.entities[0]);
                
                // Check tags on first 5 entities
                for (let i = 0; i < Math.min(5, this.entities.length); i++) {
                    const entity = this.entities[i];
                    console.log(`Entity ${i} Tags:`, entity.tags || 'No tags');
                    
                    if (entity.tags && Array.isArray(entity.tags)) {
                        const typeTags = entity.tags.filter(tag => tag.startsWith('type:'));
                        console.log(`Entity ${i} Type Tags:`, typeTags);
                    }
                }
            }
            
            console.log("===============================================");
        },
        
        // Theme management
        toggleTheme() {
            this.isDarkMode = !this.isDarkMode;
            localStorage.setItem('entitydb-theme', this.isDarkMode ? 'dark' : 'light');
            document.documentElement.setAttribute('data-theme', this.isDarkMode ? 'dark' : 'light');
        },
        
        // Initialize theme on startup
        initTheme() {
            document.documentElement.setAttribute('data-theme', this.isDarkMode ? 'dark' : 'light');
        }
    };
}

// Initialize Alpine
document.addEventListener('alpine:init', () => {
    Alpine.data('entityManager', entityManager);
    Alpine.data('debugConsole', debugConsole);
});