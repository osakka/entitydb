// Worca - Workforce Orchestrator JavaScript Application
// Powered by EntityDB

function worca() {
    return {
        // State Management
        currentView: 'dashboard',
        sidebarOpen: false,
        sidebarCollapsed: false,
        darkMode: false,
        showCreateModal: false,
        showCreateUserModal: false,
        showEditModal: false,
        loading: false,
        isAuthenticated: false,
        dataLoading: false,
        initialized: false,
        
        // EntityDB API
        api: null,
        
        // Data (initialize as arrays to prevent undefined errors)
        organizations: [],
        projects: [],
        epics: [],
        stories: [],
        tasks: [],
        teamMembers: [],
        recentActivity: [],
        
        // Filters
        selectedProject: null,
        
        // Sprint Data
        currentSprint: null,
        pastSprints: [],
        productBacklog: [],
        
        // Form Data
        createForm: {
            type: 'task',
            title: '',
            description: '',
            assignee: '',
            parent: ''
        },
        
        // User Form Data
        createUserForm: {
            username: '',
            displayName: '',
            role: 'user'
        },

        // Edit Form Data
        editForm: {
            id: '',
            type: '',
            name: '',
            description: '',
            status: '',
            priority: '',
            assignee: '',
            dueDate: ''
        },
        
        // Login Form Data
        loginForm: {
            username: 'admin',
            password: 'admin'
        },
        loginError: '',
        
        // Configuration
        kanbanStatuses: [
            { id: 'todo', name: 'To Do', color: '#fef3c7' },
            { id: 'doing', name: 'In Progress', color: '#dbeafe' },
            { id: 'review', name: 'Review', color: '#fde68a' },
            { id: 'done', name: 'Done', color: '#d1fae5' }
        ],
        
        // Statistics
        stats: {
            totalTasks: 0,
            activeTasks: 0,
            completedTasks: 0,
            teamMembers: 0
        },

        // Widget System
        dashboards: [],
        currentDashboard: null,
        widgets: [],
        editMode: false,
        showAddWidget: false,
        showDashboardSettings: false,
        selectedCategory: null,
        metricsData: null,
        metricsRefreshInterval: null,
        
        // Widget Registry Access
        get WIDGET_REGISTRY() {
            return window.WIDGET_REGISTRY || {};
        },
        
        get WIDGET_CATEGORIES() {
            return window.WIDGET_CATEGORIES || {};
        },
        
        // Dashboard Form
        dashboardForm: {
            name: '',
            description: '',
            icon: 'fa-tachometer-alt',
            isPublic: false
        },

        // Worca v2.32.4 Integration State
        connectionStatus: {
            connected: false,
            entitydb: 'Unknown',
            message: 'Checking connection...',
            lastSync: null
        },
        currentWorkspace: '',
        notifications: [],

        // Configuration and Event Systems
        config: null,
        client: null,
        events: null,
        datasetManager: null,

        // Initialization
        async init() {
            if (this.initialized) {
                console.log('⚠️ Worca already initialized, skipping...');
                return;
            }
            
            this.initCounter = (this.initCounter || 0) + 1;
            console.log(`🚀 Initializing Worca v2.32.4 (attempt ${this.initCounter})...`);
            
            try {
                // Initialize core systems
                await this.initializeCoreIntegration();
                
                // Initialize EntityDB API with new client
                this.api = new WorcaAPI();
                
                // Load user preferences
                this.loadUserPreferences();
                
                // Initialize widget system
                this.initializeWidgets();
                
                // Set up event listeners
                this.initializeEventListeners();
                
                // Check authentication (will show login screen if not authenticated)
                await this.checkAuth();
                
                // Start monitoring systems
                this.startMonitoring();
                
                this.initialized = true;
                console.log('✅ Worca v2.32.4 initialized - ready for login');
                
                // Show initialization success notification
                this.showNotification('🚀 Worca Ready', 'Workforce orchestrator initialized successfully', 'success');
                
            } catch (error) {
                console.error('❌ Worca initialization failed:', error);
                this.initialized = false;
                this.showNotification('❌ Initialization Failed', error.message, 'error');
            }
        },

        async initializeCoreIntegration() {
            console.log('🔧 Initializing core EntityDB integration...');
            console.log('🌐 Current page URL:', window.location.href);
            console.log('🏠 Detected hostname:', window.location.hostname);
            console.log('🔒 Detected protocol:', window.location.protocol);
            
            // Get global instances
            this.config = window.worcaConfig;
            this.client = window.entityDBClient;
            this.events = window.worcaEvents;
            this.datasetManager = window.datasetManager;
            
            if (!this.config || !this.client || !this.events) {
                throw new Error('Required systems not available. Please check configuration.');
            }
            
            // Log configuration details
            console.log('⚙️ EntityDB Configuration:', {
                host: this.config.get('entitydb.host'),
                port: this.config.get('entitydb.port'),
                ssl: this.config.get('entitydb.ssl'),
                url: this.client.getBaseURL()
            });
            
            // Set current workspace
            this.currentWorkspace = this.config.get('dataset.name') || 'worca-workspace';
            
            // Validate EntityDB connection
            console.log('🔍 Testing EntityDB connection...');
            const health = await this.config.validateConnection();
            this.updateConnectionStatus(health);
            
            console.log('✅ Core integration initialized');
        },

        initializeEventListeners() {
            console.log('👂 Setting up event listeners...');
            
            // Configuration events
            this.config.on('entitydb-health', (event) => {
                this.updateConnectionStatus(event);
            });
            
            this.config.on('workspace-changed', (event) => {
                this.currentWorkspace = event.namespace;
                this.refreshData();
            });
            
            // Real-time sync events
            this.events.on('sync-complete', (event) => {
                this.connectionStatus.lastSync = event.timestamp;
                if (event.changesProcessed > 0) {
                    this.refreshData();
                }
            });
            
            this.events.on('entity-changed', (event) => {
                this.handleEntityChange(event);
            });
            
            // Network events
            this.events.on('network-offline', () => {
                this.connectionStatus.connected = false;
                this.connectionStatus.message = 'Working offline';
            });
            
            this.events.on('network-online', () => {
                this.connectionStatus.connected = true;
                this.connectionStatus.message = 'Connection restored';
            });
            
            console.log('✅ Event listeners configured');
        },

        startMonitoring() {
            // Update connection status every 30 seconds
            setInterval(() => {
                this.updateConnectionStatusFromConfig();
            }, 30000);
            
            // Initial status update
            this.updateConnectionStatusFromConfig();
        },

        // Authentication
        async checkAuth() {
            console.log('🔐 Checking authentication...');
            // Don't auto-login anymore - wait for user to login manually
            this.isAuthenticated = false;
        },

        // Refresh data manually
        async refreshData() {
            console.log('🔄 Manually refreshing data...');
            await this.loadData();
        },

        // Debug helper - call from console
        debugTeamMembers() {
            console.log('🔍 Current team members:', this.teamMembers.length);
            this.teamMembers.forEach((member, index) => {
                console.log(`Member ${index + 1}:`, {
                    id: member.id,
                    name: member.name,
                    displayName: member.displayName,
                    username: member.username,
                    role: member.role,
                    type: member.type,
                    tags: member.tags
                });
            });
            return this.teamMembers;
        },

        // Test helper - create a test user with detailed logging
        async testCreateUser() {
            console.log('🧪 Testing user creation...');
            try {
                const result = await this.api.createUser(
                    'test.user',
                    'Test User',
                    'developer'
                );
                console.log('✅ Test user creation successful:', result);
                await this.refreshData();
                return result;
            } catch (error) {
                console.error('❌ Test user creation failed:', error);
                throw error;
            }
        },

        async manualLogin() {
            console.log('🚀 Starting manual login...');
            try {
                this.loading = true;
                this.loginError = '';
                console.log(`🔐 Manual login attempt for: ${this.loginForm.username}`);
                console.log('🔍 API instance:', this.api);
                console.log('🔍 About to call api.login...');
                
                const loginResult = await this.api.login(this.loginForm.username, this.loginForm.password);
                console.log('🔍 Login call completed');
                console.log('🔍 Login API call result:', loginResult);
                console.log('🔍 API token after login:', this.api.token ? 'EXISTS (' + this.api.token.substring(0, 20) + '...)' : 'MISSING');
                
                if (loginResult && this.api.token) {
                    this.isAuthenticated = true;
                    console.log('✅ Authentication status set to true');
                    
                    // Verify the token works by querying entities
                    console.log('🔍 Testing token with queryEntities...');
                    const result = await this.api.queryEntities();
                    console.log('🔍 Raw query result in login:', result);
                    console.log('✅ Login verification: found', result?.length || 0, 'entities');
                    console.log('🔍 Query result type:', typeof result, Array.isArray(result) ? 'ARRAY' : 'NOT_ARRAY');
                    console.log('🔍 Sample entities:', Array.isArray(result) ? result.slice(0, 2) : result);
                    
                    // Reset any stuck loading flags first
                    console.log('🔄 Resetting loading flags...');
                    this.loading = false;
                    this.dataLoading = false;
                    
                    // Load dashboard data
                    console.log('📊 Starting data load...');
                    await this.loadData();
                    console.log('📊 Data load completed');
                    
                    this.initializeCharts();
                    this.calculateStats();
                    this.initializeKanbanDragDrop();
                    console.log('✅ Dashboard fully initialized');
                } else {
                    console.log('❌ Login API returned false');
                    this.loginError = 'Invalid username or password';
                    this.isAuthenticated = false;
                }
            } catch (error) {
                console.error('❌ Login failed with error:', error);
                console.error('Error stack:', error.stack);
                this.loginError = error.message || 'Login failed. Please try again.';
                this.isAuthenticated = false;
            } finally {
                this.loading = false;
                console.log('🔍 Final state - authenticated:', this.isAuthenticated, 'loading:', this.loading);
            }
        },

        async tryDefaultLogin() {
            try {
                console.log('🔐 Attempting default login...');
                const loginResult = await this.api.login('admin', 'admin');
                if (loginResult) {
                    this.isAuthenticated = true;
                    console.log('✅ Logged in with default credentials');
                } else {
                    this.isAuthenticated = false;
                    console.log('❌ Default login returned false');
                }
            } catch (error) {
                console.error('❌ Default login failed:', error);
                this.isAuthenticated = false;
            }
        },

        async login(username, password) {
            try {
                this.loading = true;
                await this.api.login(username, password);
                this.isAuthenticated = true;
                await this.loadData();
                this.calculateStats();
                console.log('✅ Login successful');
            } catch (error) {
                console.error('❌ Login failed:', error);
                throw error;
            } finally {
                this.loading = false;
            }
        },

        // UI Controls
        toggleSidebar() {
            this.sidebarCollapsed = !this.sidebarCollapsed;
            console.log('🔄 Sidebar toggled:', this.sidebarCollapsed ? 'collapsed' : 'expanded');
            
            // Save preference to localStorage
            localStorage.setItem('worca-sidebar-collapsed', this.sidebarCollapsed);
        },

        toggleTheme() {
            this.darkMode = !this.darkMode;
            console.log('🎨 Theme toggled:', this.darkMode ? 'dark' : 'light');
            
            // Apply theme to document
            document.documentElement.setAttribute('data-theme', this.darkMode ? 'dark' : 'light');
            
            // Save preference to localStorage
            localStorage.setItem('worca-dark-mode', this.darkMode);
        },

        loadUserPreferences() {
            // Load sidebar preference
            const sidebarCollapsed = localStorage.getItem('worca-sidebar-collapsed');
            if (sidebarCollapsed !== null) {
                this.sidebarCollapsed = sidebarCollapsed === 'true';
            }
            
            // Load theme preference
            const darkMode = localStorage.getItem('worca-dark-mode');
            if (darkMode !== null) {
                this.darkMode = darkMode === 'true';
                document.documentElement.setAttribute('data-theme', this.darkMode ? 'dark' : 'light');
            }
        },

        logout() {
            this.api.logout();
            this.isAuthenticated = false;
            this.clearData();
            this.loginForm.username = 'admin';
            this.loginForm.password = 'admin';
            this.loginError = '';
            console.log('✅ Logged out - redirected to login screen');
        },

        // Data Loading
        async loadData() {
            console.log('🔍 LoadData called - loading flag:', this.loading, 'dataLoading flag:', this.dataLoading);
            if (this.loading) {
                console.log('⚠️ Data loading already in progress (loading=true), skipping...');
                return;
            }
            
            if (this.dataLoading) {
                console.log('⚠️ Data loading already in progress (dataLoading=true), skipping...');
                return;
            }
            
            try {
                this.loading = true;
                console.log('📡 Loading data from EntityDB...');
                
                // Load real data from EntityDB
                await this.loadRealData();
                
                this.loading = false;
            } catch (error) {
                console.error('Error loading data:', error);
                this.loading = false;
            }
        },

        // Transform hub entities to Worcha format
        transformDataspaceEntities(hubEntities) {
            return hubEntities.map(entity => {
                // Start with self properties
                const transformed = {
                    id: entity.id,
                    dataspace: entity.hub,
                    ...entity.self,
                    traits: entity.traits,
                    created_at: entity.created_at,
                    updated_at: entity.updated_at
                };

                // Add common aliases for backward compatibility
                if (entity.self) {
                    // Map common properties
                    if (entity.self.name) transformed.name = entity.self.name;
                    if (entity.self.title) transformed.title = entity.self.title;
                    if (entity.self.display_name) transformed.name = entity.self.display_name;
                    if (entity.self.username) transformed.username = entity.self.username;
                }

                // Add trait-based properties for easy access
                if (entity.traits) {
                    // Organization hierarchy
                    if (entity.traits.org) transformed.orgId = entity.traits.org;
                    if (entity.traits.project) transformed.projectId = entity.traits.project;
                    if (entity.traits.epic) transformed.epicId = entity.traits.epic;
                    if (entity.traits.story) transformed.storyId = entity.traits.story;
                    if (entity.traits.sprint) transformed.sprintId = entity.traits.sprint;
                    
                    // Team and component info
                    if (entity.traits.team) transformed.team = entity.traits.team;
                    if (entity.traits.component) transformed.component = entity.traits.component;
                }

                // Parse content if available
                if (entity.content) {
                    try {
                        const content = typeof entity.content === 'string' ? 
                            JSON.parse(atob(entity.content)) : entity.content;
                        transformed.description = content.description || transformed.description;
                        transformed.contentData = content;
                    } catch (e) {
                        // If content is not JSON, treat as plain text
                        transformed.description = entity.content;
                    }
                }

                return transformed;
            });
        },

        async loadRealData() {
            console.log('🔍 LoadRealData called - auth:', this.isAuthenticated, 'dataLoading:', this.dataLoading);
            
            if (!this.isAuthenticated) {
                console.log('❌ Cannot load data - not authenticated');
                return;
            }

            if (!this.api || !this.api.token) {
                console.log('❌ Cannot load data - no API token');
                return;
            }

            if (this.dataLoading) {
                console.log('⚠️ Data already loading, resetting flag and continuing...');
                this.dataLoading = false; // Reset stuck flag
            }

            try {
                this.dataLoading = true;
                console.log('📊 Loading data from EntityDB using Hub Architecture...');

                // Load all entity types in parallel using hub-aware API
                const [
                    organizationsResult,
                    projectsResult,
                    epicsResult,
                    storiesResult,
                    tasksResult,
                    usersResult,
                    sprintsResult
                ] = await Promise.all([
                    this.api.getOrganizations(),
                    this.api.getProjects(),
                    this.api.getEpics(),
                    this.api.getStories(),
                    this.api.getTasks(),
                    this.api.getUsers(),
                    this.api.getSprints()
                ]);
                
                // Debug: log raw API results
                console.log('🔍 Raw API results:', {
                    orgs: organizationsResult?.length || 0,
                    projects: projectsResult?.length || 0,
                    tasks: tasksResult?.length || 0,
                    users: usersResult?.length || 0
                });
                
                // Debug: log actual API results
                console.log('🔍 Raw organizations data:', organizationsResult);
                console.log('🔍 Raw tasks data:', tasksResult);

                // Extract entities from hub API responses (ensure arrays)
                this.organizations = Array.isArray(organizationsResult) ? organizationsResult : [];
                this.projects = Array.isArray(projectsResult) ? projectsResult : [];
                this.epics = Array.isArray(epicsResult) ? epicsResult : [];
                this.stories = Array.isArray(storiesResult) ? storiesResult : [];
                this.tasks = Array.isArray(tasksResult) ? tasksResult : [];
                this.teamMembers = Array.isArray(usersResult) ? usersResult : [];
                
                // Verify assignment worked
                console.log('🔍 Data assignment verification:', {
                    orgsAssigned: this.organizations.length,
                    projectsAssigned: this.projects.length,
                    tasksAssigned: this.tasks.length,
                    usersAssigned: this.teamMembers.length
                });
                
                // Force UI update by triggering Alpine.js reactivity
                console.log('🔄 Triggering UI reactivity update...');
                this.$nextTick(() => {
                    console.log('✅ UI update triggered');
                });

                // Process sprints  
                const sprints = Array.isArray(sprintsResult) ? sprintsResult : [];
                this.currentSprint = sprints.find(s => s.status === 'active' || s.status === 'planning') || null;
                this.pastSprints = sprints.filter(s => s.status === 'completed');

                // Create product backlog from unassigned stories
                const validStories = Array.isArray(this.stories) ? this.stories : [];
                const validTasks = Array.isArray(this.tasks) ? this.tasks : [];
                
                this.productBacklog = validStories.filter(story => 
                    !validTasks.some(task => task.storyId === story.id && task.sprintId)
                ).map(story => ({
                    id: story.id,
                    title: story.title || story.name,
                    description: story.description,
                    type: 'story',
                    storyPoints: story.story_points || 5,
                    priority: story.priority || 'medium',
                    status: story.status || 'ready'
                }));

                // Generate recent activity from entity timestamps
                this.recentActivity = this.generateRecentActivity();

                console.log('📊 Data loaded:', {
                    organizations: this.organizations.length,
                    projects: this.projects.length,
                    epics: this.epics.length,
                    stories: this.stories.length,
                    tasks: this.tasks.length,
                    users: this.teamMembers.length,
                    sprints: sprints.length
                });
                
                // Debug: log sample data to see what we actually got
                if (this.organizations.length > 0) {
                    console.log('Sample organization:', this.organizations[0]);
                }
                if (this.tasks.length > 0) {
                    console.log('Sample task:', this.tasks[0]);
                }

            } catch (error) {
                console.error('❌ Failed to load real data:', error);
                console.error('Error details:', error.message, error.stack);
                // Fallback to sample data initialization
                await this.initializeSampleDataIfEmpty();
            } finally {
                this.dataLoading = false;
            }
        },

        async initializeSampleDataIfEmpty() {
            try {
                // Check if we have any data
                const entities = await this.api.queryEntities();
                
                if (!Array.isArray(entities) || entities.length === 0) {
                    console.log('🔧 No data found, creating sample team members...');
                    await this.createSampleTeamMembers();
                } else {
                    console.log('📊 EntityDB has existing data:', entities.length, 'entities');
                    
                    // Check if we have team members specifically
                    const users = await this.api.getUsers();
                    if (!users || users.length === 0) {
                        console.log('👥 No team members found, creating sample users...');
                        await this.createSampleTeamMembers();
                    }
                }
                
            } catch (error) {
                console.error('Failed to initialize sample data:', error);
                // Try to create at least some team members
                try {
                    await this.createSampleTeamMembers();
                } catch (createError) {
                    console.error('Failed to create sample team members:', createError);
                }
            }
        },

        async createSampleTeamMembers() {
            console.log('👥 Creating sample team members...');
            
            const sampleMembers = [
                { name: 'Alex Johnson', role: 'Full Stack Developer', email: 'alex@techcorp.com' },
                { name: 'Sarah Chen', role: 'UI/UX Designer', email: 'sarah@techcorp.com' },
                { name: 'Mike Rodriguez', role: 'Backend Developer', email: 'mike@techcorp.com' },
                { name: 'Emma Williams', role: 'Product Manager', email: 'emma@techcorp.com' }
            ];

            try {
                for (const member of sampleMembers) {
                    console.log(`👤 Creating user: ${member.name}`);
                    await this.api.createUser(member.name, member.role, member.email);
                }
                console.log('✅ Sample team members created successfully');
            } catch (error) {
                console.error('❌ Failed to create team members:', error);
            }
        },

        generateRecentActivity() {
            const activities = [];
            
            // Generate activity from recent tasks
            const tasks = Array.isArray(this.tasks) ? this.tasks : [];
            tasks.slice(0, 10).forEach(task => {
                if (task.status === 'done') {
                    activities.push({
                        id: `activity_${task.id}`,
                        description: `Completed "${task.title}"`,
                        timestamp: new Date(task.updatedAt || task.createdAt),
                        type: 'task_completed',
                        userId: task.assignee
                    });
                } else if (task.status === 'doing') {
                    activities.push({
                        id: `activity_${task.id}_start`,
                        description: `Started working on "${task.title}"`,
                        timestamp: new Date(task.updatedAt || task.createdAt),
                        type: 'task_started',
                        userId: task.assignee
                    });
                }
            });

            return activities.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
        },

        clearData() {
            this.organizations = [];
            this.projects = [];
            this.epics = [];
            this.stories = [];
            this.tasks = [];
            this.teamMembers = [];
            this.recentActivity = [];
            this.currentSprint = null;
            this.pastSprints = [];
            this.productBacklog = [];
        },

        // EntityDB v2.32.4 Integration Methods
        updateConnectionStatus(healthData) {
            if (healthData.healthy || healthData.status === 'healthy') {
                this.connectionStatus.connected = true;
                this.connectionStatus.entitydb = healthData.data?.version || 'Connected';
                this.connectionStatus.message = 'Connected to EntityDB';
            } else {
                this.connectionStatus.connected = false;
                this.connectionStatus.entitydb = 'Disconnected';
                this.connectionStatus.message = healthData.error || 'Connection failed';
            }
        },

        updateConnectionStatusFromConfig() {
            if (this.config) {
                this.connectionStatus.connected = this.config.status.connected;
                this.connectionStatus.entitydb = this.config.status.version || 'Unknown';
                this.connectionStatus.message = this.config.status.connected ? 'Connected' : 'Disconnected';
            }
        },

        async testConnection() {
            try {
                const health = await this.config.validateConnection();
                this.updateConnectionStatus(health);
                
                if (health.healthy) {
                    this.showNotification('✅ Connection Test', 'EntityDB connection successful', 'success');
                } else {
                    this.showNotification('❌ Connection Test', health.error || 'Connection failed', 'error');
                }
            } catch (error) {
                this.showNotification('❌ Connection Test', error.message, 'error');
            }
        },

        openConfiguration() {
            // Open configuration modal or redirect to settings
            this.setView('settings');
            this.showNotification('⚙️ Configuration', 'Opening configuration panel...', 'info');
        },

        handleEntityChange(event) {
            // Handle real-time entity changes
            const { entity, type } = event;
            
            if (!entity || !entity.tags) return;
            
            // Update local data based on entity type
            const entityType = entity.tags.find(tag => tag.startsWith('type:'))?.split(':')[1];
            
            switch (entityType) {
                case 'task':
                    this.handleTaskChange(entity, type);
                    break;
                case 'project':
                    this.handleProjectChange(entity, type);
                    break;
                case 'user':
                    this.handleUserChange(entity, type);
                    break;
                default:
                    console.log('Unknown entity type for real-time update:', entityType);
            }
        },

        handleTaskChange(task, changeType) {
            const index = this.tasks.findIndex(t => t.id === task.id);
            
            if (changeType === 'update' && index > -1) {
                // Update existing task
                this.tasks[index] = { ...this.tasks[index], ...task };
                this.showNotification('📝 Task Updated', `Task "${task.title || task.id}" was updated`, 'info');
            } else if (changeType === 'create' && index === -1) {
                // Add new task
                this.tasks.push(task);
                this.showNotification('➕ New Task', `Task "${task.title || task.id}" was created`, 'success');
            }
            
            // Refresh statistics
            this.updateStats();
        },

        handleProjectChange(project, changeType) {
            const index = this.projects.findIndex(p => p.id === project.id);
            
            if (changeType === 'update' && index > -1) {
                this.projects[index] = { ...this.projects[index], ...project };
                this.showNotification('📁 Project Updated', `Project "${project.name || project.id}" was updated`, 'info');
            } else if (changeType === 'create' && index === -1) {
                this.projects.push(project);
                this.showNotification('🆕 New Project', `Project "${project.name || project.id}" was created`, 'success');
            }
        },

        handleUserChange(user, changeType) {
            const index = this.teamMembers.findIndex(u => u.id === user.id);
            
            if (changeType === 'update' && index > -1) {
                this.teamMembers[index] = { ...this.teamMembers[index], ...user };
            } else if (changeType === 'create' && index === -1) {
                this.teamMembers.push(user);
                this.showNotification('👥 New Team Member', `${user.name || user.username} joined the team`, 'success');
            }
        },

        showNotification(title, message, type = 'info', duration = 5000) {
            const notification = {
                id: Math.random().toString(36).substr(2, 9),
                title,
                message,
                type,
                duration,
                timestamp: new Date().toISOString()
            };
            
            // Emit to notification system
            window.dispatchEvent(new CustomEvent('worca:notification', { detail: notification }));
            
            // Also use events system if available
            if (this.events) {
                this.events.showNotification(title, message, type, duration);
            }
        },

        formatTime(timestamp) {
            if (!timestamp) return 'Never';
            
            try {
                const date = new Date(timestamp);
                const now = new Date();
                const diff = now - date;
                
                if (diff < 60000) return 'Just now';
                if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
                if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
                return `${Math.floor(diff / 86400000)}d ago`;
            } catch (error) {
                return 'Unknown';
            }
        },

        async initializeWorkspace() {
            try {
                if (!this.datasetManager) return;
                
                // Check if current workspace exists
                const workspace = await this.datasetManager.getWorkspace(this.currentWorkspace);
                
                if (!workspace) {
                    console.log(`Creating workspace: ${this.currentWorkspace}`);
                    await this.datasetManager.createWorkspace(this.currentWorkspace, {
                        description: 'Worca workspace for workforce management',
                        template: 'startup',
                        initializeWithSample: true
                    });
                }
                
                // Validate workspace
                const validation = await this.datasetManager.validateWorkspace(this.currentWorkspace);
                if (!validation.valid) {
                    console.warn('Workspace validation issues:', validation.errors);
                }
                
            } catch (error) {
                console.error('Failed to initialize workspace:', error);
                this.showNotification('⚠️ Workspace Error', error.message, 'warning');
            }
        },

        async loadSampleData() {
            // Organizations
            this.organizations = [
                {
                    id: 'org_1',
                    name: 'TechCorp Solutions',
                    description: 'Leading technology solutions provider',
                    status: 'active'
                }
            ];

            // Projects
            this.projects = [
                {
                    id: 'proj_1',
                    orgId: 'org_1',
                    name: 'Mobile Banking App',
                    description: 'Next-generation mobile banking application',
                    status: 'active'
                },
                {
                    id: 'proj_2',
                    orgId: 'org_1',
                    name: 'Customer Portal',
                    description: 'Web-based customer service portal',
                    status: 'active'
                }
            ];

            // Epics
            this.epics = [
                {
                    id: 'epic_1',
                    projectId: 'proj_1',
                    title: 'User Authentication',
                    description: 'Complete user authentication system',
                    status: 'in-progress'
                },
                {
                    id: 'epic_2',
                    projectId: 'proj_1',
                    title: 'Account Management',
                    description: 'Bank account management features',
                    status: 'todo'
                },
                {
                    id: 'epic_3',
                    projectId: 'proj_2',
                    title: 'Support Dashboard',
                    description: 'Customer support dashboard',
                    status: 'in-progress'
                }
            ];

            // Stories
            this.stories = [
                {
                    id: 'story_1',
                    epicId: 'epic_1',
                    title: 'Login Form',
                    description: 'Create responsive login form',
                    status: 'done'
                },
                {
                    id: 'story_2',
                    epicId: 'epic_1',
                    title: 'Registration Flow',
                    description: 'User registration workflow',
                    status: 'in-progress'
                },
                {
                    id: 'story_3',
                    epicId: 'epic_1',
                    title: 'Password Reset',
                    description: 'Password reset functionality',
                    status: 'todo'
                },
                {
                    id: 'story_4',
                    epicId: 'epic_2',
                    title: 'Account Overview',
                    description: 'Account balance and overview',
                    status: 'todo'
                }
            ];

            // Tasks
            this.tasks = [
                {
                    id: 'task_1',
                    storyId: 'story_1',
                    title: 'Create login form HTML',
                    description: 'Build responsive HTML structure for login form',
                    type: 'Development',
                    status: 'done',
                    assignee: 'user_1',
                    priority: 'high',
                    estimatedHours: 4,
                    actualHours: 3.5,
                    createdAt: new Date(Date.now() - 86400000 * 2),
                    completedAt: new Date(Date.now() - 86400000 * 1)
                },
                {
                    id: 'task_2',
                    storyId: 'story_1',
                    title: 'Style login form',
                    description: 'Apply CSS styling to login form',
                    type: 'Design',
                    status: 'done',
                    assignee: 'user_2',
                    priority: 'medium',
                    estimatedHours: 3,
                    actualHours: 2.5,
                    createdAt: new Date(Date.now() - 86400000 * 2),
                    completedAt: new Date(Date.now() - 86400000 * 1)
                },
                {
                    id: 'task_3',
                    storyId: 'story_2',
                    title: 'Design registration UI',
                    description: 'Create mockups for registration interface',
                    type: 'Design',
                    status: 'doing',
                    assignee: 'user_2',
                    priority: 'high',
                    estimatedHours: 6,
                    actualHours: 4,
                    createdAt: new Date(Date.now() - 86400000 * 1)
                },
                {
                    id: 'task_4',
                    storyId: 'story_2',
                    title: 'Implement validation',
                    description: 'Add form validation for registration',
                    type: 'Development',
                    status: 'todo',
                    assignee: 'user_1',
                    priority: 'high',
                    estimatedHours: 5,
                    createdAt: new Date(Date.now() - 86400000 * 1)
                },
                {
                    id: 'task_5',
                    storyId: 'story_3',
                    title: 'Password reset email',
                    description: 'Design and implement password reset email template',
                    type: 'Development',
                    status: 'todo',
                    assignee: 'user_3',
                    priority: 'medium',
                    estimatedHours: 4,
                    createdAt: new Date()
                },
                {
                    id: 'task_6',
                    storyId: 'story_4',
                    title: 'Account balance API',
                    description: 'Create API endpoint for account balance',
                    type: 'Backend',
                    status: 'review',
                    assignee: 'user_1',
                    priority: 'high',
                    estimatedHours: 8,
                    actualHours: 7,
                    createdAt: new Date(Date.now() - 86400000 * 3)
                }
            ];

            // Team Members
            this.teamMembers = [
                {
                    id: 'user_1',
                    name: 'Alex Johnson',
                    role: 'Full Stack Developer',
                    email: 'alex@techcorp.com',
                    avatar: null
                },
                {
                    id: 'user_2',
                    name: 'Sarah Chen',
                    role: 'UI/UX Designer',
                    email: 'sarah@techcorp.com',
                    avatar: null
                },
                {
                    id: 'user_3',
                    name: 'Mike Rodriguez',
                    role: 'Backend Developer',
                    email: 'mike@techcorp.com',
                    avatar: null
                },
                {
                    id: 'user_4',
                    name: 'Emma Williams',
                    role: 'Product Manager',
                    email: 'emma@techcorp.com',
                    avatar: null
                }
            ];

            // Recent Activity
            this.recentActivity = [
                {
                    id: 'activity_1',
                    description: 'Alex completed "Create login form HTML"',
                    timestamp: new Date(Date.now() - 3600000),
                    type: 'task_completed',
                    userId: 'user_1'
                },
                {
                    id: 'activity_2',
                    description: 'Sarah started working on "Design registration UI"',
                    timestamp: new Date(Date.now() - 7200000),
                    type: 'task_started',
                    userId: 'user_2'
                },
                {
                    id: 'activity_3',
                    description: 'Mike moved "Account balance API" to review',
                    timestamp: new Date(Date.now() - 10800000),
                    type: 'task_moved',
                    userId: 'user_3'
                },
                {
                    id: 'activity_4',
                    description: 'Emma created new story "Password Reset"',
                    timestamp: new Date(Date.now() - 14400000),
                    type: 'story_created',
                    userId: 'user_4'
                }
            ];

            // Sprint Data
            this.currentSprint = {
                id: 'sprint_1',
                name: 'Sprint 23 - Authentication Features',
                startDate: new Date(Date.now() - 86400000 * 5), // Started 5 days ago
                endDate: new Date(Date.now() + 86400000 * 9), // Ends in 9 days (2 week sprint)
                status: 'active',
                goal: 'Complete user authentication system',
                capacity: 40, // story points
                commitment: 35 // story points planned
            };

            // Update some tasks to be in current sprint
            this.tasks[2].sprintId = 'sprint_1';
            this.tasks[2].storyPoints = 8;
            this.tasks[3].sprintId = 'sprint_1';
            this.tasks[3].storyPoints = 5;
            this.tasks[4].sprintId = 'sprint_1';
            this.tasks[4].storyPoints = 3;

            // Add story points to other tasks
            this.tasks[0].storyPoints = 2;
            this.tasks[1].storyPoints = 1;
            this.tasks[5].storyPoints = 8;

            this.pastSprints = [
                {
                    id: 'sprint_0',
                    name: 'Sprint 22 - Project Setup',
                    startDate: new Date(Date.now() - 86400000 * 19),
                    endDate: new Date(Date.now() - 86400000 * 5),
                    status: 'completed',
                    capacity: 35,
                    completed: 32
                }
            ];

            this.productBacklog = [
                {
                    id: 'backlog_1',
                    title: 'Password reset functionality',
                    description: 'Implement password reset via email',
                    type: 'story',
                    storyPoints: 5,
                    priority: 'high',
                    status: 'ready'
                },
                {
                    id: 'backlog_2',
                    title: 'Two-factor authentication',
                    description: 'Add 2FA support for enhanced security',
                    type: 'story',
                    storyPoints: 13,
                    priority: 'medium',
                    status: 'ready'
                },
                {
                    id: 'backlog_3',
                    title: 'Social login integration',
                    description: 'Support Google and GitHub login',
                    type: 'story',
                    storyPoints: 8,
                    priority: 'low',
                    status: 'ready'
                }
            ];
        },

        // View Management
        setView(view) {
            this.currentView = view;
            
            // Re-initialize specific functionality when switching views
            this.$nextTick(() => {
                if (view === 'kanban') {
                    this.initializeKanbanDragDrop();
                } else if (view === 'reports') {
                    // Multiple attempts to ensure charts load
                    this.initializeChartsWithRetry();
                }
            });
        },

        getViewTitle() {
            const titles = {
                dashboard: 'Dashboard',
                kanban: 'Kanban Board',
                projects: 'Project Hierarchy',
                team: 'Team Overview',
                reports: 'Analytics & Reports',
                sprints: 'Sprint Planning',
                backlog: 'Backlog',
                settings: 'Settings'
            };
            return titles[this.currentView] || 'Worcha';
        },

        getViewDescription() {
            const descriptions = {
                dashboard: 'Overview of your workforce activities',
                kanban: 'Visual task management board',
                projects: 'Organizational structure and project breakdown',
                team: 'Team members and workload distribution',
                reports: 'Performance metrics and analytics',
                sprints: 'Agile sprint planning and management',
                backlog: 'Manage and track epics, stories, and tasks',
                settings: 'Configure your workspace'
            };
            return descriptions[this.currentView] || '';
        },


        // Settings functions
        refreshSessionData() {
            console.log('🔄 Refreshing session data...');
            // In a real implementation, this would fetch from an API
            console.log('✅ Session data refreshed');
        },

        // Data Queries
        getProjectsByOrg(orgId) {
            return Array.isArray(this.projects) ? this.projects.filter(p => p.orgId === orgId) : [];
        },

        getEpicsByProject(projectId) {
            return Array.isArray(this.epics) ? this.epics.filter(e => e.projectId === projectId) : [];
        },

        getStoriesByEpic(epicId) {
            return Array.isArray(this.stories) ? this.stories.filter(s => s.epicId === epicId) : [];
        },

        getTasksByStory(storyId) {
            return Array.isArray(this.tasks) ? this.tasks.filter(t => t.storyId === storyId) : [];
        },

        getTasksByStatus(status) {
            const filteredTasks = Array.isArray(this.tasks) ? this.tasks.filter(t => {
                // Filter by status
                if (t.status !== status) return false;
                
                // Filter by selected project if one is selected
                if (this.selectedProject) {
                    // Check if task belongs to selected project (via story -> epic -> project)
                    if (t.storyId) {
                        const story = this.stories.find(s => s.id === t.storyId);
                        if (story && story.epicId) {
                            const epic = this.epics.find(e => e.id === story.epicId);
                            if (epic && epic.projectId) {
                                return epic.projectId === this.selectedProject;
                            }
                        }
                    }
                    // If no story/epic/project chain, don't show in filtered view
                    return false;
                }
                
                return true;
            }) : [];
            
            return filteredTasks;
        },

        getTasksByAssignee(assigneeId) {
            return Array.isArray(this.tasks) ? this.tasks.filter(t => t.assignee === assigneeId) : [];
        },

        getSprintTasksByStatus(status) {
            if (!this.currentSprint) return [];
            return Array.isArray(this.tasks) ? this.tasks.filter(t => 
                t.sprintId === this.currentSprint.id && t.status === status
            ) : [];
        },

        // Backlog view helper functions
        getFilteredData(type, searchTerm, sortField, sortDirection) {
            let data = [];
            
            // Get base data based on type
            switch(type) {
                case 'epics':
                    data = this.selectedProject 
                        ? this.epics.filter(epic => epic.projectId === this.selectedProject)
                        : [...this.epics];
                    break;
                case 'stories':
                    if (this.selectedProject) {
                        const projectEpics = this.epics.filter(epic => epic.projectId === this.selectedProject);
                        const epicIds = projectEpics.map(epic => epic.id);
                        data = this.stories.filter(story => epicIds.includes(story.epicId));
                    } else {
                        data = [...this.stories];
                    }
                    break;
                case 'tasks':
                    if (this.selectedProject) {
                        const projectEpics = this.epics.filter(epic => epic.projectId === this.selectedProject);
                        const epicIds = projectEpics.map(epic => epic.id);
                        const projectStories = this.stories.filter(story => epicIds.includes(story.epicId));
                        const storyIds = projectStories.map(story => story.id);
                        data = this.tasks.filter(task => storyIds.includes(task.storyId));
                    } else {
                        data = [...this.tasks];
                    }
                    break;
            }
            
            // Apply search filter
            if (searchTerm) {
                const search = searchTerm.toLowerCase();
                data = data.filter(item => 
                    (item.name || item.title || '').toLowerCase().includes(search) ||
                    (item.description || '').toLowerCase().includes(search)
                );
            }
            
            // Apply sorting
            data.sort((a, b) => {
                let aVal = a[sortField];
                let bVal = b[sortField];
                
                // Handle missing values
                if (sortField === 'name' || sortField === 'title') {
                    aVal = a.name || a.title || '';
                    bVal = b.name || b.title || '';
                }
                
                // Handle special cases
                if (sortField === 'priority') {
                    const priorityOrder = { high: 3, medium: 2, low: 1 };
                    aVal = priorityOrder[aVal] || 0;
                    bVal = priorityOrder[bVal] || 0;
                }
                
                if (sortField === 'createdAt' || sortField === 'dueDate') {
                    aVal = new Date(aVal || 0);
                    bVal = new Date(bVal || 0);
                }
                
                if (sortDirection === 'asc') {
                    return aVal > bVal ? 1 : -1;
                } else {
                    return aVal < bVal ? 1 : -1;
                }
            });
            
            return data;
        },
        
        getProjectName(projectId) {
            const project = this.projects.find(p => p.id === projectId);
            return project ? project.name : 'Unknown';
        },
        
        getEpicName(epicId) {
            const epic = this.epics.find(e => e.id === epicId);
            return epic ? (epic.title || epic.name) : 'Unknown';
        },
        
        getStoryName(storyId) {
            const story = this.stories.find(s => s.id === storyId);
            return story ? (story.title || story.name) : 'Unknown';
        },
        
        getEpicProgress(epicId) {
            const stories = this.stories.filter(s => s.epicId === epicId);
            if (stories.length === 0) return 0;
            
            let completedWeight = 0;
            let totalWeight = 0;
            
            stories.forEach(story => {
                const tasks = this.tasks.filter(t => t.storyId === story.id);
                totalWeight += tasks.length || 1;
                completedWeight += tasks.filter(t => t.status === 'done').length;
            });
            
            return totalWeight > 0 ? Math.round((completedWeight / totalWeight) * 100) : 0;
        },
        
        getStoryProgress(storyId) {
            const tasks = this.tasks.filter(t => t.storyId === storyId);
            if (tasks.length === 0) return 0;
            
            const completed = tasks.filter(t => t.status === 'done').length;
            return Math.round((completed / tasks.length) * 100);
        },
        
        editItem(item) {
            console.log('🖊️ Opening edit modal for:', item);
            
            // Determine item type
            let itemType = 'task';
            if (this.epics.some(e => e.id === item.id)) {
                itemType = 'epic';
            } else if (this.stories.some(s => s.id === item.id)) {
                itemType = 'story';
            }
            
            // Populate edit form
            this.editForm = {
                id: item.id,
                type: itemType,
                name: item.name || item.title || '',
                description: item.description || '',
                status: item.status || 'todo',
                priority: item.priority || 'medium',
                assignee: item.assignee || '',
                dueDate: item.dueDate ? new Date(item.dueDate).toISOString().split('T')[0] : ''
            };
            
            this.showEditModal = true;
        },

        async saveEditItem() {
            try {
                this.loading = true;
                console.log('💾 Saving item:', this.editForm);

                // Find the item in the appropriate array
                let itemArray, itemIndex;
                switch(this.editForm.type) {
                    case 'epic':
                        itemArray = this.epics;
                        itemIndex = this.epics.findIndex(e => e.id === this.editForm.id);
                        break;
                    case 'story':
                        itemArray = this.stories;
                        itemIndex = this.stories.findIndex(s => s.id === this.editForm.id);
                        break;
                    case 'task':
                        itemArray = this.tasks;
                        itemIndex = this.tasks.findIndex(t => t.id === this.editForm.id);
                        break;
                }

                if (itemIndex === -1) {
                    throw new Error(`${this.editForm.type} not found`);
                }

                // Update the item
                const item = itemArray[itemIndex];
                const originalItem = { ...item };

                // Update local item first (optimistic update)
                if (this.editForm.name) {
                    if (this.editForm.type === 'task') {
                        item.title = this.editForm.name;
                    } else {
                        item.name = this.editForm.name;
                    }
                }
                item.description = this.editForm.description;
                item.status = this.editForm.status;
                item.priority = this.editForm.priority;
                item.dueDate = this.editForm.dueDate;
                if (this.editForm.type === 'task') {
                    item.assignee = this.editForm.assignee;
                }
                item.updatedAt = new Date();

                // Update in EntityDB
                try {
                    await this.api.updateEntity(this.editForm.id, {
                        name: this.editForm.name,
                        description: this.editForm.description,
                        status: this.editForm.status,
                        priority: this.editForm.priority,
                        assignee: this.editForm.assignee,
                        dueDate: this.editForm.dueDate
                    });
                } catch (apiError) {
                    // Revert local changes on API failure
                    console.error('❌ API update failed, reverting local changes:', apiError);
                    Object.assign(item, originalItem);
                    throw apiError;
                }

                // Add to activity log
                this.recentActivity.unshift({
                    id: this.generateId(),
                    description: `Updated ${this.editForm.type} "${this.editForm.name}"`,
                    timestamp: new Date(),
                    type: `${this.editForm.type}_updated`
                });

                // Force reactivity update
                switch(this.editForm.type) {
                    case 'epic':
                        this.epics = [...this.epics];
                        break;
                    case 'story':
                        this.stories = [...this.stories];
                        break;
                    case 'task':
                        this.tasks = [...this.tasks];
                        break;
                }

                this.showEditModal = false;
                this.resetEditForm();
                this.calculateStats();

                console.log('✅ Item updated successfully');

            } catch (error) {
                console.error('❌ Error updating item:', error);
                alert(`Error updating ${this.editForm.type}: ${error.message}`);
            } finally {
                this.loading = false;
            }
        },

        resetEditForm() {
            this.editForm = {
                id: '',
                type: '',
                name: '',
                description: '',
                status: '',
                priority: '',
                assignee: '',
                dueDate: ''
            };
        },

        canEdit() {
            // Check if user has edit permissions
            // For now, we'll check if they're authenticated and have admin role or entity:update permission
            if (!this.isAuthenticated || !this.api || !this.api.token) {
                return false;
            }
            
            // TODO: Implement proper RBAC permission checking
            // For now, assume authenticated users can edit (you can enhance this later with proper permission checking)
            return true;
        },

        getAssigneeName(assigneeId) {
            if (!assigneeId) return 'Unassigned';
            const member = this.teamMembers.find(m => m.id === assigneeId);
            return member ? member.name : 'Unknown';
        },

        // Statistics
        calculateStats() {
            const tasks = Array.isArray(this.tasks) ? this.tasks : [];
            const members = Array.isArray(this.teamMembers) ? this.teamMembers : [];
            
            this.stats.totalTasks = tasks.length;
            this.stats.activeTasks = tasks.filter(t => ['todo', 'doing', 'review'].includes(t.status)).length;
            this.stats.completedTasks = tasks.filter(t => t.status === 'done').length;
            this.stats.teamMembers = members.length;
        },

        // Chart Management
        initializeCharts() {
            this.statusChart = null;
            this.workloadChart = null;
        },

        updateCharts() {
            this.updateStatusChart();
            this.updateWorkloadChart();
        },

        async initializeChartsWithRetry(maxAttempts = 3) {
            console.log('📊 Initializing charts for reports view...');
            
            for (let attempt = 1; attempt <= maxAttempts; attempt++) {
                console.log(`📊 Chart initialization attempt ${attempt}/${maxAttempts}`);
                
                // Wait for DOM to be fully ready
                await new Promise(resolve => setTimeout(resolve, 150 * attempt));
                
                const statusCanvas = document.getElementById('statusChart');
                const workloadCanvas = document.getElementById('workloadChart');
                
                if (statusCanvas && workloadCanvas) {
                    console.log('✅ Chart canvases found, initializing charts...');
                    this.updateCharts();
                    return;
                } else {
                    console.warn(`❌ Chart canvases not found on attempt ${attempt}`);
                    if (attempt === maxAttempts) {
                        console.error('❌ Failed to find chart canvases after all attempts');
                    }
                }
            }
        },

        updateStatusChart() {
            const ctx = document.getElementById('statusChart');
            if (!ctx) {
                console.log('❌ Status chart canvas not found');
                return;
            }

            const statusCounts = this.kanbanStatuses.map(status => 
                this.getTasksByStatus(status.id).length
            );

            // Destroy existing chart to prevent memory leaks
            if (this.statusChart) {
                this.statusChart.destroy();
                this.statusChart = null;
            }

            // Set canvas size explicitly to prevent growing
            ctx.style.width = '100%';
            ctx.style.height = '300px';
            ctx.width = ctx.offsetWidth;
            ctx.height = 300;

            this.statusChart = new Chart(ctx, {
                type: 'doughnut',
                data: {
                    labels: this.kanbanStatuses.map(s => s.name),
                    datasets: [{
                        data: statusCounts,
                        backgroundColor: ['#fef3c7', '#dbeafe', '#fde68a', '#d1fae5'],
                        borderWidth: 2,
                        borderColor: '#fff'
                    }]
                },
                options: {
                    responsive: false,
                    maintainAspectRatio: true,
                    animation: {
                        duration: 500,
                        easing: 'easeOutQuart'
                    },
                    plugins: {
                        legend: {
                            position: 'bottom',
                            labels: {
                                padding: 20,
                                usePointStyle: true
                            }
                        }
                    }
                }
            });
        },

        updateWorkloadChart() {
            const ctx = document.getElementById('workloadChart');
            if (!ctx) {
                console.log('❌ Workload chart canvas not found');
                return;
            }

            const workloadData = this.teamMembers.map(member => 
                this.getTasksByAssignee(member.id).length
            );

            // Destroy existing chart to prevent memory leaks
            if (this.workloadChart) {
                this.workloadChart.destroy();
                this.workloadChart = null;
            }

            // Set canvas size explicitly to prevent growing
            ctx.style.width = '100%';
            ctx.style.height = '300px';
            ctx.width = ctx.offsetWidth;
            ctx.height = 300;

            this.workloadChart = new Chart(ctx, {
                type: 'bar',
                data: {
                    labels: this.teamMembers.map(m => m.name),
                    datasets: [{
                        label: 'Active Tasks',
                        data: workloadData,
                        backgroundColor: '#667eea',
                        borderColor: '#5a67d8',
                        borderWidth: 1
                    }]
                },
                options: {
                    responsive: false,
                    maintainAspectRatio: true,
                    animation: {
                        duration: 500,
                        easing: 'easeOutQuart'
                    },
                    scales: {
                        y: {
                            beginAtZero: true,
                            ticks: {
                                stepSize: 1
                            }
                        }
                    },
                    plugins: {
                        legend: {
                            display: false
                        }
                    }
                }
            });
        },

        // CRUD Operations
        async createItem() {
            try {
                this.loading = true;
                let newItem;

                switch (this.createForm.type) {
                    case 'organization':
                        newItem = await this.api.createOrganization(
                            this.createForm.title,
                            this.createForm.description,
                            { industry: 'technology', region: 'global' }
                        );
                        this.organizations.push(newItem);
                        break;

                    case 'project':
                        const orgName = this.organizations[0]?.name || 'DefaultOrg';
                        
                        newItem = await this.api.createProject(
                            this.createForm.title,
                            orgName,
                            this.createForm.description
                        );
                        this.projects.push(newItem);
                        break;

                    case 'epic':
                        const projectName = this.projects[0]?.name || 'DefaultProject';
                        
                        newItem = await this.api.createEpic(
                            this.createForm.title,
                            projectName,
                            this.createForm.description,
                            { org: orgName }
                        );
                        this.epics.push(newItem);
                        break;

                    case 'story':
                        const epicName = this.epics[0]?.name || 'DefaultEpic';
                        const projectForStory = this.epics[0]?.traits?.project || 'DefaultProject';
                        
                        newItem = await this.api.createStory(
                            this.createForm.title,
                            epicName,
                            this.createForm.description,
                            { project: projectForStory, org: orgName }
                        );
                        this.stories.push(newItem);
                        break;

                    case 'task':
                        const storyName = this.stories[0]?.name; // Optional
                        const projectForTask = this.stories[0]?.traits?.project || 'DefaultProject';
                        const epicForTask = this.stories[0]?.traits?.epic || 'DefaultEpic';
                        
                        newItem = await this.api.createTask(
                            this.createForm.title,
                            this.createForm.description,
                            this.createForm.assignee,
                            'medium',
                            { 
                                project: projectForTask, 
                                org: orgName,
                                epic: epicForTask,
                                story: storyName
                            }
                        );
                        this.tasks.push(newItem);
                        break;

                    default:
                        throw new Error(`Unknown type: ${this.createForm.type}`);
                }

                // Add to recent activity
                this.recentActivity.unshift({
                    id: this.generateId(),
                    description: `Created new ${this.createForm.type}: "${this.createForm.title}"`,
                    timestamp: new Date(),
                    type: this.createForm.type + '_created'
                });

                this.showCreateModal = false;
                this.resetCreateForm();
                this.calculateStats();

                console.log(`✅ Created new ${this.createForm.type}:`, newItem.id);

            } catch (error) {
                console.error('Error creating item:', error);
                alert(`Error creating ${this.createForm.type}: ${error.message}`);
            } finally {
                this.loading = false;
            }
        },

        async createUser() {
            try {
                this.loading = true;
                
                const newUser = await this.api.createUser(
                    this.createUserForm.username,
                    this.createUserForm.displayName,
                    this.createUserForm.role
                );
                
                // Refresh team members list to get properly transformed data
                const usersResult = await this.api.getUsers();
                this.teamMembers = Array.isArray(usersResult) ? usersResult : [];
                
                // Debug: log the team members to see what data we have
                console.log('🔍 Updated team members after user creation:', this.teamMembers);
                this.teamMembers.forEach((member, index) => {
                    console.log(`User ${index + 1}:`, {
                        id: member.id,
                        name: member.name,
                        displayName: member.displayName,
                        username: member.username,
                        role: member.role,
                        tags: member.tags
                    });
                });
                
                // Add to recent activity
                this.recentActivity.unshift({
                    id: this.generateId(),
                    description: `Added new team member: "${this.createUserForm.displayName}"`,
                    timestamp: new Date(),
                    type: 'user_created'
                });

                this.showCreateUserModal = false;
                this.resetCreateUserForm();
                this.calculateStats();

                console.log('✅ Created new user:', newUser.id);

            } catch (error) {
                console.error('Error creating user:', error);
                alert(`Error creating user: ${error.message}`);
            } finally {
                this.loading = false;
            }
        },

        resetCreateUserForm() {
            this.createUserForm = {
                username: '',
                displayName: '',
                role: 'user'
            };
        },

        viewTask(task) {
            console.log('Viewing task:', task);
            // In a full implementation, this would open a detailed task modal
        },

        // Utility Functions
        generateId() {
            return 'id_' + Math.random().toString(36).substr(2, 9);
        },

        resetCreateForm() {
            this.createForm = {
                type: 'task',
                title: '',
                description: '',
                assignee: '',
                parent: ''
            };
        },

        formatTime(timestamp) {
            const now = new Date();
            const diff = now - timestamp;
            const hours = Math.floor(diff / 3600000);
            const days = Math.floor(hours / 24);

            if (days > 0) return `${days}d ago`;
            if (hours > 0) return `${hours}h ago`;
            return 'Just now';
        },

        async refreshData() {
            if (!this.isAuthenticated) return;
            
            try {
                this.loading = true;
                console.log('🔄 Refreshing data from EntityDB...');
                await this.loadRealData();
                this.calculateStats();
                if (this.currentView === 'reports') {
                    this.updateCharts();
                }
                console.log('✅ Data refreshed successfully');
            } catch (error) {
                console.error('❌ Error refreshing data:', error);
            } finally {
                this.loading = false;
            }
        },

        // Sprint Management
        getSprintProgress() {
            if (!this.currentSprint) return 0;
            
            const sprintTasks = this.tasks.filter(t => t.sprintId === this.currentSprint.id);
            const completedTasks = sprintTasks.filter(t => t.status === 'done');
            
            return sprintTasks.length > 0 ? Math.round((completedTasks.length / sprintTasks.length) * 100) : 0;
        },

        async createSprint() {
            try {
                this.loading = true;
                
                const sprintNumber = this.pastSprints.length + (this.currentSprint ? 2 : 1);
                const sprintName = `Sprint ${sprintNumber + 22} - New Sprint`;
                const startDate = new Date();
                const endDate = new Date(Date.now() + 86400000 * 14); // 2 weeks

                // Complete current sprint if exists
                if (this.currentSprint) {
                    await this.api.updateEntity(this.currentSprint.id, {
                        tags: this.currentSprint.tags.map(tag => 
                            tag.startsWith('status:') ? 'status:completed' : tag
                        )
                    });
                    this.currentSprint.status = 'completed';
                    this.pastSprints.unshift(this.currentSprint);
                }

                // Create new sprint
                const newSprint = await this.api.createSprint(
                    sprintName,
                    startDate,
                    endDate,
                    'Sprint goal to be defined',
                    40
                );

                this.currentSprint = newSprint;
                console.log('✅ Created new sprint:', this.currentSprint.id);

            } catch (error) {
                console.error('Error creating sprint:', error);
                alert('Error creating sprint: ' + error.message);
            } finally {
                this.loading = false;
            }
        },

        async addToSprint(backlogItem) {
            if (!this.currentSprint) return;

            try {
                this.loading = true;

                // Create new task from backlog item with sprint trait
                const newTask = await this.api.createTask(
                    backlogItem.title,
                    backlogItem.description,
                    null, // No assignee initially
                    backlogItem.priority,
                    {
                        story: backlogItem.title,
                        sprint: this.currentSprint.name,
                        project: 'DefaultProject',
                        org: 'DefaultOrg'
                    }
                );

                // Add to local tasks
                newTask.sprintId = this.currentSprint.id;
                newTask.storyPoints = backlogItem.storyPoints;
                this.tasks.push(newTask);
                
                // Remove from backlog
                const index = this.productBacklog.findIndex(item => item.id === backlogItem.id);
                if (index > -1) {
                    this.productBacklog.splice(index, 1);
                }

                console.log('✅ Added to sprint:', newTask.id);

            } catch (error) {
                console.error('Error adding to sprint:', error);
                alert('Error adding to sprint: ' + error.message);
            } finally {
                this.loading = false;
            }
        },

        formatDate(date) {
            if (!date) return 'N/A';
            return new Date(date).toLocaleDateString();
        },

        // Task Status Updates for Kanban Drag & Drop
        async updateTaskStatus(taskId, newStatus) {
            try {
                const taskIndex = this.tasks.findIndex(t => t.id === taskId);
                if (taskIndex === -1) {
                    console.error('❌ Task not found:', taskId);
                    return;
                }

                const task = this.tasks[taskIndex];
                const oldStatus = task.status;
                
                console.log(`🔄 Moving "${task.title}": ${oldStatus} -> ${newStatus}`);

                // Update local task FIRST (optimistic update)
                task.status = newStatus;
                task.updatedAt = new Date();
                
                // Force Alpine.js reactivity by replacing the array reference
                this.tasks = [...this.tasks];

                // Update in EntityDB
                await this.api.updateTaskStatus(taskId, newStatus);
                
                // Add to activity log
                this.recentActivity.unshift({
                    id: this.generateId(),
                    description: `Moved "${task.title}" to ${this.kanbanStatuses.find(s => s.id === newStatus)?.name || newStatus}`,
                    timestamp: new Date(),
                    type: 'task_moved'
                });

                // Update stats
                this.calculateStats();
                
                console.log(`✅ Task moved successfully: "${task.title}" -> ${newStatus}`);
                
            } catch (error) {
                console.error('❌ Error updating task status:', error);
                alert('Error updating task status: ' + error.message);
            }
        },

        async assignTaskToUser(taskId, userId) {
            try {
                const task = this.tasks.find(t => t.id === taskId);
                if (!task) {
                    console.error('Task not found:', taskId);
                    return;
                }

                // Update in EntityDB
                await this.api.assignTask(taskId, userId);
                
                // Update local task
                task.assignee = userId;
                task.updatedAt = new Date();
                
                const user = this.teamMembers.find(u => u.id === userId);
                const userName = user ? user.name : 'Unassigned';
                
                // Add to activity log
                this.recentActivity.unshift({
                    id: this.generateId(),
                    description: `Assigned "${task.title}" to ${userName}`,
                    timestamp: new Date(),
                    type: 'task_assigned'
                });

                console.log('✅ Task assigned:', taskId, '->', userId);
                
            } catch (error) {
                console.error('Error assigning task:', error);
                alert('Error assigning task: ' + error.message);
            }
        },

        // Initialize Kanban Drag & Drop
        initializeKanbanDragDrop() {
            console.log('🎯 Initializing kanban drag & drop...');
            
            // Wait for DOM to be ready with multiple attempts
            this.$nextTick(() => {
                // Additional delay to ensure Alpine.js has finished rendering
                setTimeout(() => {
                    this.doInitializeKanban();
                }, 100);
            });
        },

        doInitializeKanban() {
            if (typeof Sortable === 'undefined') {
                console.warn('❌ SortableJS not loaded, drag & drop disabled');
                return;
            }

            console.log('✅ SortableJS is loaded');
            
            // Clear existing instances first
            this.kanbanStatuses.forEach(status => {
                const column = document.querySelector(`[data-status="${status.id}"] .kanban-tasks`);
                if (column && column.sortableInstance) {
                    column.sortableInstance.destroy();
                    column.sortableInstance = null;
                    console.log(`🗑️ Destroyed existing sortable for ${status.id}`);
                }
            });

            // Initialize drag & drop for each kanban column
            let successfulInitializations = 0;
            this.kanbanStatuses.forEach(status => {
                const columnSelector = `[data-status="${status.id}"] .kanban-tasks`;
                const column = document.querySelector(columnSelector);
                
                console.log(`🔍 Looking for column: ${columnSelector}`, column);
                
                if (column) {
                    console.log(`📋 Initializing drag & drop for ${status.id} column`);
                    
                    column.sortableInstance = Sortable.create(column, {
                        group: 'kanban',
                        animation: 150,
                        ghostClass: 'task-ghost',
                        chosenClass: 'task-chosen',
                        dragClass: 'task-drag',
                        onStart: (evt) => {
                            console.log('🎯 Drag started:', evt.item.getAttribute('data-task-id'));
                        },
                        onEnd: async (evt) => {
                            const taskId = evt.item.getAttribute('data-task-id');
                            const newStatusColumn = evt.to.closest('.kanban-column');
                            const newStatus = newStatusColumn ? newStatusColumn.getAttribute('data-status') : null;
                            const fromStatus = evt.from.closest('.kanban-column')?.getAttribute('data-status');
                            
                            console.log(`📍 Drag: ${taskId} from ${fromStatus} to ${newStatus}`);
                            
                            if (taskId && newStatus) {
                                // Only update if status actually changed
                                if (fromStatus !== newStatus) {
                                    console.log(`🔄 Status change detected: ${fromStatus} -> ${newStatus} for task ${taskId}`);
                                    await this.updateTaskStatus(taskId, newStatus);
                                } else {
                                    console.log(`ℹ️ No status change needed: ${taskId} already in ${newStatus}`);
                                }
                            } else {
                                console.warn('❌ Missing taskId or newStatus:', { taskId, newStatus });
                                console.warn('❌ Full event object:', evt);
                            }
                        }
                    });
                    
                    successfulInitializations++;
                    console.log(`✅ Sortable created for ${status.id}`);
                } else {
                    console.warn(`❌ Column not found for status: ${status.id} (selector: ${columnSelector})`);
                }
            });

            if (successfulInitializations === 0) {
                console.warn('❌ No kanban columns found! Retrying in 500ms...');
                setTimeout(() => {
                    console.log('🔄 Retrying kanban initialization...');
                    this.doInitializeKanban();
                }, 500);
            } else {
                console.log(`✅ Kanban drag & drop initialization completed (${successfulInitializations}/${this.kanbanStatuses.length} columns)`);
            }
        },

        // EntityDB Integration (Legacy)
        async callEntityDB(endpoint, method = 'GET', data = null) {
            try {
                const options = {
                    method,
                    headers: {
                        'Content-Type': 'application/json'
                    }
                };

                if (data) {
                    options.body = JSON.stringify(data);
                }

                const response = await fetch(`/api/v1/${endpoint}`, options);
                return await response.json();
            } catch (error) {
                console.error('EntityDB API Error:', error);
                throw error;
            }
        },

        // Widget System Methods
        initializeWidgets() {
            console.log('🧩 ========== INITIALIZING WIDGET SYSTEM ==========');
            
            // Load dashboards from storage first
            console.log('📂 Loading dashboards from storage...');
            this.loadDashboardsFromStorage();
            console.log(`📂 Loaded ${this.dashboards.length} dashboards from storage`);
            
            // Create default dashboard if none exist
            if (this.dashboards.length === 0) {
                console.log('🆕 Creating default dashboard - no dashboards found');
                
                const defaultDashboard = {
                    id: this.generateId(),
                    name: 'Main Dashboard',
                    icon: 'tachometer-alt',
                    widgets: [
                        {
                            id: this.generateId(),
                            type: 'systemInfo',
                            x: 0, y: 0, w: 3, h: 2
                        },
                        {
                            id: this.generateId(),
                            type: 'taskOverview',
                            x: 3, y: 0, w: 3, h: 2
                        },
                        {
                            id: this.generateId(),
                            type: 'teamMembers',
                            x: 6, y: 0, w: 3, h: 2
                        }
                    ]
                };
                
                console.log('🆕 Default dashboard created:', defaultDashboard);
                
                this.dashboards = [defaultDashboard];
                this.currentDashboard = this.dashboards[0];
                this.widgets = [...this.currentDashboard.widgets];
                
                console.log(`🆕 Set current dashboard: ${this.currentDashboard.name} with ${this.widgets.length} widgets`);
                
                // Save the default dashboard
                this.saveDashboard();
            } else {
                console.log(`✅ Using existing dashboards - setting current to first one`);
                if (!this.currentDashboard && this.dashboards.length > 0) {
                    this.currentDashboard = this.dashboards[0];
                    this.widgets = [...this.currentDashboard.widgets];
                    console.log(`✅ Set current dashboard: ${this.currentDashboard.name} with ${this.widgets.length} widgets`);
                }
            }
            
            console.log('📊 Starting metrics refresh...');
            this.refreshMetrics();
            this.startMetricsRefresh();
            
            this.logWidgetState('INITIALIZATION_COMPLETE');
            console.log('🧩 ========== WIDGET SYSTEM INITIALIZATION DONE ==========');
        },

        async refreshMetrics() {
            try {
                const response = await fetch('/api/v1/system/metrics');
                if (response.ok) {
                    this.metricsData = await response.json();
                }
            } catch (error) {
                console.error('Failed to fetch metrics:', error);
            }
        },

        startMetricsRefresh() {
            this.metricsRefreshInterval = setInterval(() => {
                this.refreshMetrics();
            }, 30000);
        },

        getWidgetIcon(type) {
            const registry = window.WIDGET_REGISTRY || {};
            return registry[type]?.icon?.replace('fa-', '') || 'cube';
        },

        getWidgetName(type) {
            const registry = window.WIDGET_REGISTRY || {};
            return registry[type]?.name || 'Unknown Widget';
        },

        renderWidgetContent(widget) {
            // Simple widget content rendering
            switch (widget.type) {
                case 'systemInfo':
                    return this.metricsData ? `
                        <div>
                            <div><strong>CPU:</strong> ${(this.metricsData.cpu_percent || 0).toFixed(1)}%</div>
                            <div><strong>Memory:</strong> ${this.formatBytes(this.metricsData.memory_used || 0)}</div>
                            <div><strong>Uptime:</strong> ${this.formatDuration(this.metricsData.uptime_seconds || 0)}</div>
                        </div>
                    ` : '<div>Loading...</div>';
                case 'taskOverview':
                    return `
                        <div>
                            <div><strong>Total:</strong> ${this.stats.totalTasks}</div>
                            <div><strong>Active:</strong> ${this.stats.activeTasks}</div>
                            <div><strong>Done:</strong> ${this.stats.completedTasks}</div>
                        </div>
                    `;
                case 'teamMembers':
                    return `
                        <div>
                            <div><strong>Team Size:</strong> ${this.teamMembers.length}</div>
                            <div>Active members working on ${this.stats.activeTasks} tasks</div>
                        </div>
                    `;
                default:
                    return '<div>Widget content loading...</div>';
            }
        },

        formatBytes(bytes) {
            if (!bytes) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        },

        formatDuration(seconds) {
            if (!seconds) return '0s';
            const days = Math.floor(seconds / 86400);
            const hours = Math.floor((seconds % 86400) / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            
            if (days > 0) return `${days}d ${hours}h`;
            if (hours > 0) return `${hours}h ${minutes}m`;
            return `${minutes}m`;
        },

        // Dashboard Management Methods
        toggleEditMode() {
            console.log(`🔧 TOGGLE EDIT MODE - Before: ${this.editMode}`);
            this.editMode = !this.editMode;
            console.log(`🔧 TOGGLE EDIT MODE - After: ${this.editMode}`);
            
            this.logWidgetState('EDIT_MODE_TOGGLE');
            
            if (!this.editMode) {
                console.log(`💾 Exiting edit mode - saving dashboard`);
                this.saveDashboard();
            }
        },

        async saveDashboard() {
            if (!this.currentDashboard) return;
            
            try {
                // In a real implementation, this would save to EntityDB
                console.log('💾 Saving dashboard:', this.currentDashboard.name);
                
                // Update the dashboard in the dashboards array
                const index = this.dashboards.findIndex(d => d.id === this.currentDashboard.id);
                if (index >= 0) {
                    this.dashboards[index] = { ...this.currentDashboard };
                }
                
                // Save to localStorage as fallback
                localStorage.setItem('worca_dashboards', JSON.stringify(this.dashboards));
                
                console.log('✅ Dashboard saved successfully');
            } catch (error) {
                console.error('❌ Failed to save dashboard:', error);
            }
        },

        addWidget(type) {
            console.log(`➕ ADD WIDGET CALLED - Type: ${type}`);
            
            if (!this.currentDashboard) {
                console.error(`❌ ADD WIDGET FAILED - No current dashboard`);
                return;
            }
            
            console.log(`📋 Current dashboard: ${this.currentDashboard.name}`);
            console.log(`📋 Dashboard has ${this.currentDashboard.widgets.length} widgets before adding`);
            
            // Get widget definition
            const widgetDef = this.WIDGET_REGISTRY[type] || {};
            console.log(`🔍 Widget definition found:`, widgetDef);
            
            // Simple positioning - just add to end of grid
            const existingWidgets = this.currentDashboard.widgets.length;
            const newWidget = {
                id: this.generateId(),
                type: type,
                x: (existingWidgets % 4) * 3, // Simple 4-column layout
                y: Math.floor(existingWidgets / 4) * 2,
                w: 3, // Simple fixed width
                h: 2  // Simple fixed height
            };
            
            console.log(`🆕 Created new widget:`, newWidget);
            
            // Add to dashboard
            this.currentDashboard.widgets.push(newWidget);
            console.log(`📋 Dashboard now has ${this.currentDashboard.widgets.length} widgets`);
            
            // Update local widgets array
            this.widgets = [...this.currentDashboard.widgets];
            console.log(`🔄 Updated local widgets array to ${this.widgets.length} items`);
            
            // Close add widget panel
            this.showAddWidget = false;
            console.log(`✅ Set showAddWidget to false`);
            
            this.logWidgetState('ADD_WIDGET_COMPLETE');
        },

        removeWidget(widgetId) {
            console.log(`🗑️ REMOVE WIDGET CALLED - ID: ${widgetId}`);
            
            if (!this.currentDashboard) {
                console.error(`❌ REMOVE WIDGET FAILED - No current dashboard`);
                return;
            }
            
            console.log(`📋 Dashboard has ${this.currentDashboard.widgets.length} widgets before removal`);
            
            const beforeCount = this.currentDashboard.widgets.length;
            this.currentDashboard.widgets = this.currentDashboard.widgets.filter(w => w.id !== widgetId);
            const afterCount = this.currentDashboard.widgets.length;
            
            console.log(`📋 Removed ${beforeCount - afterCount} widgets`);
            console.log(`📋 Dashboard now has ${afterCount} widgets`);
            
            this.widgets = [...this.currentDashboard.widgets];
            console.log(`🔄 Updated local widgets array to ${this.widgets.length} items`);
            
            this.logWidgetState('REMOVE_WIDGET_COMPLETE');
        },

        createDashboard() {
            const newDashboard = {
                id: this.generateId(),
                name: this.dashboardForm.name || 'New Dashboard',
                icon: this.dashboardForm.icon || 'tachometer-alt',
                widgets: []
            };
            
            this.dashboards.push(newDashboard);
            this.currentDashboard = newDashboard;
            this.widgets = [];
            
            // Reset form
            this.dashboardForm = {
                name: '',
                description: '',
                icon: 'fa-tachometer-alt',
                isPublic: false
            };
            
            this.showDashboardSettings = false;
            console.log(`📊 Created new dashboard: ${newDashboard.name}`);
        },

        // Alias for HTML form
        saveNewDashboard() {
            this.createDashboard();
        },

        switchDashboard(dashboard) {
            this.currentDashboard = dashboard;
            this.widgets = [...dashboard.widgets];
            console.log(`🔄 Switched to dashboard: ${dashboard.name}`);
        },

        loadDashboardsFromStorage() {
            try {
                const stored = localStorage.getItem('worca_dashboards');
                if (stored) {
                    this.dashboards = JSON.parse(stored);
                    if (this.dashboards.length > 0 && !this.currentDashboard) {
                        this.currentDashboard = this.dashboards[0];
                        this.widgets = [...this.currentDashboard.widgets];
                    }
                }
            } catch (error) {
                console.error('Failed to load dashboards from storage:', error);
            }
        },

        // Simple Widget System - No Complex Grid
        logWidgetState(action) {
            console.log(`🧩 WIDGET ${action.toUpperCase()}:`);
            console.log(`   - Current Dashboard:`, this.currentDashboard?.name || 'NONE');
            console.log(`   - Dashboard Widgets:`, this.currentDashboard?.widgets?.length || 0);
            console.log(`   - Local Widgets Array:`, this.widgets.length);
            console.log(`   - Edit Mode:`, this.editMode);
            console.log(`   - Show Add Widget:`, this.showAddWidget);
            console.log(`   - Available Widget Types:`, Object.keys(this.WIDGET_REGISTRY).length);
            
            if (this.currentDashboard?.widgets) {
                this.currentDashboard.widgets.forEach((widget, index) => {
                    console.log(`   - Widget ${index + 1}: ${widget.type} (${widget.id}) at ${widget.x},${widget.y} size ${widget.w}x${widget.h}`);
                });
            }
        }
    };
}