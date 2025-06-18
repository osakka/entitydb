// Worca EntityDB API Integration
// Complete backend functionality using EntityDB v2.32.4

class WorcaAPI {
    constructor() {
        // Use EntityDB client for all API operations
        this.client = window.entityDBClient || new EntityDBClient();
        this.config = window.worcaConfig;
        
        // Legacy compatibility
        this.baseURL = this.client.getBaseURL();
        this.token = this.client.token;
        
        console.log('🚀 WorcaAPI initialized with EntityDB client');
        console.log('📡 Server URL:', this.baseURL);
        console.log('🔐 Authentication:', this.token ? 'Configured' : 'Not configured');
    }

    // Authentication (delegated to EntityDB client)
    async login(username, password) {
        try {
            const result = await this.client.login(username, password);
            
            // Update legacy token reference
            this.token = this.client.token;
            
            return result;
        } catch (error) {
            console.error('🔍 Worca login failed:', error);
            throw error;
        }
    }

    async logout() {
        await this.client.logout();
        this.token = null;
    }

    // Generic EntityDB operations (delegated to EntityDB client)
    async createEntity(entityData) {
        try {
            // Ensure worca namespace is added
            const worcaData = this.addWorcaNamespace(entityData);
            return await this.client.createEntity(worcaData);
        } catch (error) {
            console.error('🔍 Worca create entity failed:', error);
            throw error;
        }
    }

    async updateEntity(id, entityData) {
        try {
            return await this.client.updateEntity(id, entityData);
        } catch (error) {
            console.error('🔍 Worca update entity failed:', error);
            throw error;
        }
    }

    async getEntity(id) {
        try {
            const entity = await this.client.getEntity(id);
            return this.transformEntity(entity);
        } catch (error) {
            console.error('🔍 Worca get entity failed:', error);
            throw error;
        }
    }

    async queryEntities(filters = {}) {
        try {
            // Add worca namespace filter automatically
            const worcaFilters = this.addWorcaFilters(filters);
            const entities = await this.client.queryEntities(worcaFilters);
            
            // Transform all entities for worca compatibility
            return entities.map(entity => this.transformEntity(entity));
        } catch (error) {
            console.error('🔍 Worca query entities failed:', error);
            throw error;
        }
    }

    // Helper methods for worca integration
    addWorcaNamespace(entityData) {
        const worcaData = JSON.parse(JSON.stringify(entityData));
        
        if (!worcaData.tags) {
            worcaData.tags = [];
        }
        
        // Add worca namespace if not present
        const namespace = this.config.get('dataset.namespace');
        const hasNamespace = worcaData.tags.some(tag => tag.startsWith(`namespace:${namespace}`));
        if (!hasNamespace) {
            worcaData.tags.unshift(`namespace:${namespace}`);
        }
        
        return worcaData;
    }

    addWorcaFilters(filters) {
        const worcaFilters = { ...filters };
        
        // If no specific tag filter, add worca namespace
        if (!worcaFilters.tag && !worcaFilters.namespace) {
            const namespace = this.config.get('dataset.namespace');
            worcaFilters.tag = `namespace:${namespace}`;
        }
        
        return worcaFilters;
    }

    // Worca-specific operations
    async createOrganization(name, description) {
        const orgData = {
            tags: [
                'type:organization',
                `name:${name}`,
                'status:active'
            ],
            content: [{
                type: 'description',
                value: description || ''
            }]
        };

        return await this.createEntity(orgData);
    }

    async createProject(name, orgId, description) {
        const projectData = {
            tags: [
                'type:project',
                `name:${name}`,
                `org:${orgId}`,
                'status:active'
            ],
            content: [{
                type: 'description',
                value: description || ''
            }]
        };

        return await this.createEntity(projectData);
    }

    async createEpic(title, projectId, description) {
        const epicData = {
            tags: [
                'type:epic',
                `title:${title}`,
                `project:${projectId}`,
                'status:todo'
            ],
            content: [{
                type: 'description',
                value: description || ''
            }]
        };

        return await this.createEntity(epicData);
    }

    async createStory(title, epicId, description) {
        const storyData = {
            tags: [
                'type:story',
                `title:${title}`,
                `epic:${epicId}`,
                'status:todo'
            ],
            content: [{
                type: 'description',
                value: description || ''
            }]
        };

        return await this.createEntity(storyData);
    }

    async createTask(title, description, assignee = null, priority = 'medium', traits = {}) {
        const tags = [
            'type:task',
            `title:${title}`,
            `status:${traits.status || 'todo'}`,
            `priority:${priority}`,
            `created:${new Date().toISOString()}`
        ];

        if (assignee) tags.push(`assignee:${assignee}`);
        
        // Add traits as tags (except status which is handled above)
        Object.entries(traits).forEach(([key, value]) => {
            if (value && key !== 'status') tags.push(`${key}:${value}`);
        });

        const taskData = {
            tags,
            content: [{
                type: 'description',
                value: description || ''
            }]
        };

        return await this.createEntity(taskData);
    }

    async createUser(username, displayName, role = 'user') {
        const userData = {
            tags: [
                'type:user',
                `username:${username}`,
                `displayName:${displayName}`,
                `role:${role}`,
                'status:active'
            ],
            content: [{
                type: 'profile',
                value: JSON.stringify({
                    username,
                    displayName,
                    role,
                    createdAt: new Date().toISOString()
                })
            }]
        };

        // Validate that all required fields are present
        if (!username || !displayName) {
            throw new Error('Username and displayName are required for user creation');
        }

        console.log('🔍 Creating user:', { username, displayName, role });

        return await this.createEntity(userData);
    }

    async createSprint(name, startDate, endDate, goal, capacity = 40) {
        const sprintData = {
            tags: [
                'type:sprint',
                `name:${name}`,
                'status:planning',
                `capacity:${capacity}`,
                `start:${startDate.toISOString()}`,
                `end:${endDate.toISOString()}`
            ],
            content: [{
                type: 'goal',
                value: goal || ''
            }]
        };

        return await this.createEntity(sprintData);
    }


    // Query methods for different entity types
    async getOrganizations() {
        console.log('🏢 Getting organizations...');
        const result = await this.queryEntities();
        console.log('🔍 Raw query result for orgs:', result?.length || 0, 'entities');
        const filtered = this.filterEntitiesByType(Array.isArray(result) ? result : [], 'organization');
        console.log('🔍 Filtered organizations:', filtered.length, 'found');
        console.log('🔍 Sample org:', filtered[0]);
        return filtered;
    }

    async getProjects(orgId = null) {
        const result = await this.queryEntities();
        let projects = this.filterEntitiesByType(Array.isArray(result) ? result : [], 'project');
        
        if (orgId) {
            projects = projects.filter(p => this.getTagValue(p.tags, 'org') === orgId);
        }
        
        return projects;
    }

    async getEpics(projectId = null) {
        const result = await this.queryEntities();
        let epics = this.filterEntitiesByType(Array.isArray(result) ? result : [], 'epic');
        
        if (projectId) {
            epics = epics.filter(e => this.getTagValue(e.tags, 'project') === projectId);
        }
        
        return epics;
    }

    async getStories(epicId = null) {
        const result = await this.queryEntities();
        let stories = this.filterEntitiesByType(Array.isArray(result) ? result : [], 'story');
        
        if (epicId) {
            stories = stories.filter(s => this.getTagValue(s.tags, 'epic') === epicId);
        }
        
        return stories;
    }

    async getTasks(filters = {}) {
        console.log('📋 Getting tasks...');
        const result = await this.queryEntities();
        console.log('🔍 Raw query result for tasks:', result?.length || 0, 'entities');
        let tasks = this.filterEntitiesByType(Array.isArray(result) ? result : [], 'task');
        console.log('🔍 Filtered tasks:', tasks.length, 'found');
        
        // Apply filters
        if (filters.status) {
            tasks = tasks.filter(t => this.getTagValue(t.tags, 'status') === filters.status);
        }
        
        if (filters.assignee) {
            tasks = tasks.filter(t => this.getTagValue(t.tags, 'assignee') === filters.assignee);
        }
        
        if (filters.story) {
            tasks = tasks.filter(t => this.getTagValue(t.tags, 'story') === filters.story);
        }
        
        if (filters.sprint) {
            tasks = tasks.filter(t => this.getTagValue(t.tags, 'sprint') === filters.sprint);
        }
        
        return tasks;
    }

    async getUsers() {
        const result = await this.queryEntities();
        const users = this.filterEntitiesByType(Array.isArray(result) ? result : [], 'user');
        
        // Debug logging for users (can be removed in production)
        console.log('🔍 Loaded', users.length, 'users from EntityDB');
        users.forEach((user, index) => {
            console.log(`🔍 User ${index + 1}:`, user.name || 'No Name', '-', user.role || 'No Role');
        });
        
        return users;
    }

    async getSprints() {
        const result = await this.queryEntities();
        return this.filterEntitiesByType(Array.isArray(result) ? result : [], 'sprint');
    }

    // Update operations
    async updateTaskStatus(taskId, newStatus) {
        const task = await this.getEntity(taskId);
        if (!task) throw new Error('Task not found');

        // Update status tag
        const updatedTags = task.tags.map(tag => 
            tag.startsWith('status:') ? `status:${newStatus}` : tag
        );

        return await this.updateEntity(taskId, { tags: updatedTags });
    }

    async assignTask(taskId, assigneeId) {
        const task = await this.getEntity(taskId);
        if (!task) throw new Error('Task not found');

        // Remove existing assignee tag and add new one
        let updatedTags = task.tags.filter(tag => !tag.startsWith('assignee:'));
        if (assigneeId) {
            updatedTags.push(`assignee:${assigneeId}`);
        }

        return await this.updateEntity(taskId, { tags: updatedTags });
    }

    async addTaskToSprint(taskId, sprintId) {
        const task = await this.getEntity(taskId);
        if (!task) throw new Error('Task not found');

        // Remove existing sprint tag and add new one
        let updatedTags = task.tags.filter(tag => !tag.startsWith('sprint:'));
        if (sprintId) {
            updatedTags.push(`sprint:${sprintId}`);
        }

        return await this.updateEntity(taskId, { tags: updatedTags });
    }

    // Helper methods
    filterEntitiesByType(entities, type) {
        console.log(`🔍 Filtering ${entities.length} entities for type: ${type}`);
        
        const filtered = entities.filter(entity => {
            if (!entity.tags) {
                console.log('❌ Entity has no tags:', entity.id);
                return false;
            }
            
            // Check for both standard format and dataspace format (worcha legacy + worca current)
            const matches = entity.tags.some(tag => 
                tag === `type:${type}` || 
                tag === `worcha:self:type:${type}` ||
                tag === `worca:self:type:${type}` ||
                (tag.startsWith('dataspace:worcha') && entity.tags.some(t => t === `worcha:self:type:${type}`)) ||
                (tag.startsWith('dataspace:worca') && entity.tags.some(t => t === `worca:self:type:${type}`))
            );
            
            if (matches) {
                console.log(`✅ Found ${type}:`, entity.id, entity.tags.slice(0, 3));
            }
            
            return matches;
        });
        
        console.log(`🔍 Filter result: ${filtered.length} ${type} entities found`);
        return filtered.map(entity => this.transformEntity(entity));
    }

    transformEntity(entity) {
        if (!entity) return null;

        // Entity is already transformed by EntityDB client
        const transformed = {
            id: entity.id,
            tags: entity.tags || [],
            content: entity.content,
            createdAt: entity.createdAt,
            updatedAt: entity.updatedAt
        };

        // Extract worca-specific properties from tags
        transformed.type = this.getTagValue(entity.tags, 'type');
        transformed.status = this.getTagValue(entity.tags, 'status');
        transformed.name = this.getTagValue(entity.tags, 'name');
        transformed.title = this.getTagValue(entity.tags, 'title');
        transformed.username = this.getTagValue(entity.tags, 'username');
        transformed.displayName = this.getTagValue(entity.tags, 'displayName');
        transformed.assignee = this.getTagValue(entity.tags, 'assignee');
        transformed.priority = this.getTagValue(entity.tags, 'priority') || 'medium';
        transformed.role = this.getTagValue(entity.tags, 'role');
        transformed.email = this.getTagValue(entity.tags, 'email');
        
        // Handle description from content
        if (entity.content) {
            if (typeof entity.content === 'string') {
                transformed.description = entity.content;
            } else if (Array.isArray(entity.content)) {
                const descItem = entity.content.find(item => item.type === 'description');
                transformed.description = descItem?.value || '';
            } else if (entity.content.description) {
                transformed.description = entity.content.description;
            } else {
                transformed.description = '';
            }
        } else {
            transformed.description = '';
        }

        // User-specific handling
        if (transformed.type === 'user') {
            // Handle admin user role extraction from rbac tags
            if (!transformed.role && entity.tags) {
                const adminRole = entity.tags.find(t => t === 'rbac:role:admin');
                if (adminRole) {
                    transformed.role = 'Administrator';
                }
            }
            
            // Name fallback logic
            if (!transformed.name) {
                transformed.name = transformed.displayName || transformed.username || 'Unknown User';
            }
            
            // Role fallback
            if (!transformed.role) {
                transformed.role = 'User';
            }
        }

        // Type-specific transformations
        if (transformed.type === 'task') {
            transformed.storyId = this.getTagValue(entity.tags, 'story');
            transformed.sprintId = this.getTagValue(entity.tags, 'sprint');
            transformed.storyPoints = parseInt(this.getTagValue(entity.tags, 'storyPoints')) || 3;
        }

        if (transformed.type === 'epic') {
            transformed.projectId = this.getTagValue(entity.tags, 'project');
        }

        if (transformed.type === 'story') {
            transformed.epicId = this.getTagValue(entity.tags, 'epic');
        }

        if (transformed.type === 'project') {
            transformed.orgId = this.getTagValue(entity.tags, 'org');
        }

        if (transformed.type === 'sprint') {
            transformed.capacity = parseInt(this.getTagValue(entity.tags, 'capacity')) || 40;
            transformed.startDate = this.getTagValue(entity.tags, 'start');
            transformed.endDate = this.getTagValue(entity.tags, 'end');
        }

        return transformed;
    }

    getTagValue(tags, key) {
        if (!tags) return null;
        
        // Debug: log tag search for user-related keys (verbose, can be removed in production)
        const isUserKey = ['name', 'displayName', 'username', 'role', 'email'].includes(key);
        if (isUserKey && false) { // Set to true for debugging
            console.log(`🔍 getTagValue searching for '${key}' in:`, tags.slice(0, 8));
        }
        
        // Try standard format first (key:value)
        let tag = tags.find(t => t.startsWith(`${key}:`));
        let result = tag ? tag.substring(key.length + 1) : null;
        
        // If not found and looking for username, try id:username:value format (for admin user)
        if (!result && key === 'username') {
            tag = tags.find(t => t.startsWith('id:username:'));
            result = tag ? tag.substring('id:username:'.length) : null;
        }
        
        // If still not found and looking for name, try using username as fallback
        if (!result && key === 'name') {
            const username = this.getTagValue(tags, 'username');
            if (username) {
                result = username;
            }
        }
        
        // Debug: log result for user-related keys (verbose, can be removed in production)
        if (isUserKey && false) { // Set to true for debugging
            console.log(`🔍 getTagValue('${key}') = '${result}' (from tag: '${tag}')`);
        }
        
        return result;
    }

    // Initialize sample data (run once)
    async initializeSampleData() {
        try {
            console.log('🚀 Initializing Worcha sample data...');

            // Create organization
            const org = await this.createOrganization(
                'TechCorp Solutions',
                'Leading technology solutions provider'
            );
            console.log('✅ Created organization:', org.id);

            // Create projects
            const mobileProject = await this.createProject(
                'Mobile Banking App',
                org.id,
                'Next-generation mobile banking application'
            );

            const portalProject = await this.createProject(
                'Customer Portal',
                org.id,
                'Web-based customer service portal'
            );
            console.log('✅ Created projects:', mobileProject.id, portalProject.id);

            // Create epics
            const authEpic = await this.createEpic(
                'User Authentication',
                mobileProject.id,
                'Complete user authentication system'
            );

            const accountEpic = await this.createEpic(
                'Account Management',
                mobileProject.id,
                'Bank account management features'
            );
            console.log('✅ Created epics:', authEpic.id, accountEpic.id);

            // Create stories
            const loginStory = await this.createStory(
                'Login Form',
                authEpic.id,
                'Create responsive login form'
            );

            const registerStory = await this.createStory(
                'Registration Flow',
                authEpic.id,
                'User registration workflow'
            );
            console.log('✅ Created stories:', loginStory.id, registerStory.id);

            // Create users
            const alex = await this.createUser('alex', 'Alex Johnson', 'developer');
            const sarah = await this.createUser('sarah', 'Sarah Chen', 'designer');
            const mike = await this.createUser('mike', 'Mike Rodriguez', 'developer');
            const emma = await this.createUser('emma', 'Emma Williams', 'manager');
            console.log('✅ Created users:', alex.id, sarah.id, mike.id, emma.id);

            // Create tasks
            const task1 = await this.createTask(
                'Create login form HTML',
                'Build responsive HTML structure for login form',
                loginStory.id,
                alex.id,
                'high'
            );

            const task2 = await this.createTask(
                'Style login form',
                'Apply CSS styling to login form',
                loginStory.id,
                sarah.id,
                'medium'
            );

            const task3 = await this.createTask(
                'Design registration UI',
                'Create mockups for registration interface',
                registerStory.id,
                sarah.id,
                'high'
            );

            console.log('✅ Created tasks:', task1.id, task2.id, task3.id);

            // Update some task statuses
            await this.updateTaskStatus(task1.id, 'done');
            await this.updateTaskStatus(task2.id, 'done');
            await this.updateTaskStatus(task3.id, 'doing');

            // Create sprint
            const sprint = await this.createSprint(
                'Sprint 23 - Authentication Features',
                new Date(Date.now() - 86400000 * 5), // Started 5 days ago
                new Date(Date.now() + 86400000 * 9), // Ends in 9 days
                'Complete user authentication system'
            );

            // Add tasks to sprint
            await this.addTaskToSprint(task3.id, sprint.id);

            console.log('✅ Created sprint:', sprint.id);
            console.log('🎉 Sample data initialization complete!');

            return {
                organization: org,
                projects: [mobileProject, portalProject],
                epics: [authEpic, accountEpic],
                stories: [loginStory, registerStory],
                users: [alex, sarah, mike, emma],
                tasks: [task1, task2, task3],
                sprint: sprint
            };

        } catch (error) {
            console.error('❌ Error initializing sample data:', error);
            throw error;
        }
    }
}

// Export for use in Worca application
window.WorcaAPI = WorcaAPI;