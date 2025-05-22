// Worcha - Workforce Orchestrator JavaScript Application
// Powered by EntityDB

function worcha() {
    return {
        // State Management
        currentView: 'dashboard',
        sidebarOpen: false,
        showCreateModal: false,
        loading: false,
        isAuthenticated: false,
        
        // EntityDB API
        api: null,
        
        // Data
        organizations: [],
        projects: [],
        epics: [],
        stories: [],
        tasks: [],
        teamMembers: [],
        recentActivity: [],
        
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

        // Initialization
        async init() {
            console.log('🚀 Initializing Worcha...');
            
            // Initialize EntityDB API
            this.api = new WorchaAPI();
            
            // Check authentication
            await this.checkAuth();
            
            if (this.isAuthenticated) {
                await this.loadData();
                this.initializeCharts();
                this.calculateStats();
                this.initializeKanbanDragDrop();
            }
            
            console.log('✅ Worcha initialized successfully');
        },

        // Authentication
        async checkAuth() {
            try {
                // Try to load data to verify token is valid
                const result = await this.api.queryEntities();
                this.isAuthenticated = true;
                console.log('✅ Authentication verified');
            } catch (error) {
                console.log('⚠️ Authentication required');
                this.isAuthenticated = false;
                // Try default admin login
                await this.tryDefaultLogin();
            }
        },

        async tryDefaultLogin() {
            try {
                await this.api.login('admin', 'admin');
                this.isAuthenticated = true;
                console.log('✅ Logged in with default credentials');
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

        logout() {
            this.api.logout();
            this.isAuthenticated = false;
            this.clearData();
            console.log('✅ Logged out');
        },

        // Data Loading
        async loadData() {
            try {
                this.loading = true;
                
                // Load real data from EntityDB
                await this.loadRealData();
                
                this.loading = false;
            } catch (error) {
                console.error('Error loading data:', error);
                this.loading = false;
            }
        },

        async loadRealData() {
            try {
                console.log('📊 Loading data from EntityDB...');

                // Load all entity types in parallel
                const [
                    organizations,
                    projects,
                    epics,
                    stories,
                    tasks,
                    users,
                    sprints
                ] = await Promise.all([
                    this.api.getOrganizations(),
                    this.api.getProjects(),
                    this.api.getEpics(),
                    this.api.getStories(),
                    this.api.getTasks(),
                    this.api.getUsers(),
                    this.api.getSprints()
                ]);

                this.organizations = organizations;
                this.projects = projects;
                this.epics = epics;
                this.stories = stories;
                this.tasks = tasks;
                this.teamMembers = users;

                // Process sprints
                this.currentSprint = sprints.find(s => s.status === 'active' || s.status === 'planning') || null;
                this.pastSprints = sprints.filter(s => s.status === 'completed');

                // Create product backlog from unassigned stories
                this.productBacklog = stories.filter(story => 
                    !this.tasks.some(task => task.storyId === story.id && task.sprintId)
                ).map(story => ({
                    id: story.id,
                    title: story.title,
                    description: story.description,
                    type: 'story',
                    storyPoints: 5, // Default value
                    priority: 'medium',
                    status: 'ready'
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

            } catch (error) {
                console.error('Failed to load real data:', error);
                // Fallback to sample data initialization
                await this.initializeSampleDataIfEmpty();
            }
        },

        async initializeSampleDataIfEmpty() {
            try {
                // Check if we have any data
                const entities = await this.api.queryEntities();
                
                if (!entities.entities || entities.entities.length === 0) {
                    console.log('🔧 No data found, initializing sample data...');
                    await this.api.initializeSampleData();
                    await this.loadRealData();
                } else {
                    console.log('📊 Using existing EntityDB data');
                    await this.loadRealData();
                }
            } catch (error) {
                console.error('Failed to initialize sample data:', error);
            }
        },

        generateRecentActivity() {
            const activities = [];
            
            // Generate activity from recent tasks
            this.tasks.slice(0, 10).forEach(task => {
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
            if (view === 'reports') {
                // Delay chart initialization to ensure DOM is ready
                setTimeout(() => this.updateCharts(), 100);
            }
        },

        getViewTitle() {
            const titles = {
                dashboard: 'Dashboard',
                kanban: 'Kanban Board',
                projects: 'Project Hierarchy',
                team: 'Team Overview',
                reports: 'Analytics & Reports',
                sprints: 'Sprint Planning',
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
                settings: 'Configure your workspace'
            };
            return descriptions[this.currentView] || '';
        },

        setView(view) {
            this.currentView = view;
            
            // Re-initialize specific functionality when switching views
            this.$nextTick(() => {
                if (view === 'kanban') {
                    this.initializeKanbanDragDrop();
                } else if (view === 'reports') {
                    this.updateCharts();
                }
            });
        },

        // Data Queries
        getProjectsByOrg(orgId) {
            return this.projects.filter(p => p.orgId === orgId);
        },

        getEpicsByProject(projectId) {
            return this.epics.filter(e => e.projectId === projectId);
        },

        getStoriesByEpic(epicId) {
            return this.stories.filter(s => s.epicId === epicId);
        },

        getTasksByStory(storyId) {
            return this.tasks.filter(t => t.storyId === storyId);
        },

        getTasksByStatus(status) {
            return this.tasks.filter(t => t.status === status);
        },

        getTasksByAssignee(assigneeId) {
            return this.tasks.filter(t => t.assignee === assigneeId);
        },

        getSprintTasksByStatus(status) {
            if (!this.currentSprint) return [];
            return this.tasks.filter(t => 
                t.sprintId === this.currentSprint.id && t.status === status
            );
        },

        // Statistics
        calculateStats() {
            this.stats.totalTasks = this.tasks.length;
            this.stats.activeTasks = this.tasks.filter(t => ['todo', 'doing', 'review'].includes(t.status)).length;
            this.stats.completedTasks = this.tasks.filter(t => t.status === 'done').length;
            this.stats.teamMembers = this.teamMembers.length;
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

        updateStatusChart() {
            const ctx = document.getElementById('statusChart');
            if (!ctx) return;

            const statusCounts = this.kanbanStatuses.map(status => 
                this.getTasksByStatus(status.id).length
            );

            if (this.statusChart) {
                this.statusChart.destroy();
            }

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
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: {
                            position: 'bottom'
                        }
                    }
                }
            });
        },

        updateWorkloadChart() {
            const ctx = document.getElementById('workloadChart');
            if (!ctx) return;

            const workloadData = this.teamMembers.map(member => 
                this.getTasksByAssignee(member.id).length
            );

            if (this.workloadChart) {
                this.workloadChart.destroy();
            }

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
                    responsive: true,
                    maintainAspectRatio: false,
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
                            this.createForm.description
                        );
                        this.organizations.push(this.api.transformEntity(newItem));
                        break;

                    case 'project':
                        const orgId = this.organizations[0]?.id;
                        if (!orgId) throw new Error('No organization found');
                        
                        newItem = await this.api.createProject(
                            this.createForm.title,
                            orgId,
                            this.createForm.description
                        );
                        this.projects.push(this.api.transformEntity(newItem));
                        break;

                    case 'epic':
                        const projectId = this.projects[0]?.id;
                        if (!projectId) throw new Error('No project found');
                        
                        newItem = await this.api.createEpic(
                            this.createForm.title,
                            projectId,
                            this.createForm.description
                        );
                        this.epics.push(this.api.transformEntity(newItem));
                        break;

                    case 'story':
                        const epicId = this.epics[0]?.id;
                        if (!epicId) throw new Error('No epic found');
                        
                        newItem = await this.api.createStory(
                            this.createForm.title,
                            epicId,
                            this.createForm.description
                        );
                        this.stories.push(this.api.transformEntity(newItem));
                        break;

                    case 'task':
                        const storyId = this.stories[0]?.id; // Optional
                        newItem = await this.api.createTask(
                            this.createForm.title,
                            this.createForm.description,
                            storyId,
                            this.createForm.assignee,
                            'medium'
                        );
                        this.tasks.push(this.api.transformEntity(newItem));
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

                this.currentSprint = this.api.transformEntity(newSprint);
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

                // Create new task from backlog item
                const newTask = await this.api.createTask(
                    backlogItem.title,
                    backlogItem.description,
                    backlogItem.id, // Use backlog item as story
                    null, // No assignee initially
                    backlogItem.priority
                );

                // Add to current sprint
                await this.api.addTaskToSprint(newTask.id, this.currentSprint.id);

                // Add to local tasks
                const transformedTask = this.api.transformEntity(newTask);
                transformedTask.sprintId = this.currentSprint.id;
                transformedTask.storyPoints = backlogItem.storyPoints;
                this.tasks.push(transformedTask);
                
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
                const task = this.tasks.find(t => t.id === taskId);
                if (!task) {
                    console.error('Task not found:', taskId);
                    return;
                }

                // Update in EntityDB
                await this.api.updateTaskStatus(taskId, newStatus);
                
                // Update local task
                task.status = newStatus;
                task.updatedAt = new Date();
                
                // Add to activity log
                this.recentActivity.unshift({
                    id: this.generateId(),
                    description: `Moved "${task.title}" to ${this.kanbanStatuses.find(s => s.id === newStatus)?.name || newStatus}`,
                    timestamp: new Date(),
                    type: 'task_moved'
                });

                // Update stats
                this.calculateStats();
                
                console.log('✅ Task status updated:', taskId, '->', newStatus);
                
            } catch (error) {
                console.error('Error updating task status:', error);
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
            // Wait for DOM to be ready
            this.$nextTick(() => {
                if (typeof Sortable === 'undefined') {
                    console.warn('SortableJS not loaded, drag & drop disabled');
                    return;
                }

                // Initialize drag & drop for each kanban column
                this.kanbanStatuses.forEach(status => {
                    const column = document.querySelector(`[data-status="${status.id}"] .kanban-tasks`);
                    if (column && !column.sortableInstance) {
                        column.sortableInstance = Sortable.create(column, {
                            group: 'kanban',
                            animation: 150,
                            ghostClass: 'task-ghost',
                            chosenClass: 'task-chosen',
                            dragClass: 'task-drag',
                            onEnd: async (evt) => {
                                const taskId = evt.item.getAttribute('data-task-id');
                                const newStatus = evt.to.closest('.kanban-column').getAttribute('data-status');
                                
                                if (taskId && newStatus) {
                                    await this.updateTaskStatus(taskId, newStatus);
                                }
                            }
                        });
                    }
                });

                console.log('✅ Kanban drag & drop initialized');
            });
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
        }
    };
}