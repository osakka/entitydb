// Worca EntityDB API Integration
// Complete backend functionality using EntityDB

class WorcaAPI {
    constructor() {
        this.baseURL = '/api/v1';
        this.token = localStorage.getItem('authToken');
    }

    // Authentication
    async login(username, password) {
        try {
            console.log('🔍 API login starting...');
            console.log('🔍 Login URL:', `${this.baseURL}/auth/login`);
            console.log('🔍 Login credentials:', { username, password: password ? '***' : 'MISSING' });
            
            const response = await fetch(`${this.baseURL}/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username, password })
            });
            
            console.log('🔍 Login response status:', response.status, response.statusText);

            if (response.ok) {
                const data = await response.json();
                console.log('🔍 Login response data:', data);
                this.token = data.token;
                console.log('🔍 Token set:', this.token ? 'SUCCESS' : 'FAILED');
                localStorage.setItem('authToken', this.token);
                return data;
            } else {
                const errorText = await response.text();
                console.error('🔍 Login failed:', response.status, errorText);
                throw new Error('Login failed');
            }
        } catch (error) {
            console.error('Login error:', error);
            throw error;
        }
    }

    logout() {
        this.token = null;
        localStorage.removeItem('authToken');
    }

    // Generic EntityDB operations
    async createEntity(entityData) {
        try {
            // Ensure entityData has proper structure
            if (!entityData.tags || !Array.isArray(entityData.tags)) {
                throw new Error('EntityData must have tags array');
            }

            // Add creation timestamp if not present
            const hasCreatedTag = entityData.tags.some(tag => tag.startsWith('created:'));
            if (!hasCreatedTag) {
                entityData.tags.push(`created:${new Date().toISOString()}`);
            }

            console.log('🔍 Creating entity with data:', {
                tags: entityData.tags,
                content: entityData.content,
                fullPayload: entityData
            });

            const response = await fetch(`${this.baseURL}/entities/create`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.token}`
                },
                body: JSON.stringify(entityData)
            });

            console.log('🔍 EntityDB response status:', response.status, response.statusText);

            if (response.ok) {
                const result = await response.json();
                console.log('🔍 EntityDB creation result:', result);
                return result;
            } else {
                const errorText = await response.text();
                console.error('🔍 EntityDB error response:', errorText);
                throw new Error(`Failed to create entity: ${response.status} ${response.statusText} - ${errorText}`);
            }
        } catch (error) {
            console.error('Create entity error:', error);
            throw error;
        }
    }

    async updateEntity(id, entityData) {
        try {
            // Add updated timestamp if we have tags to update
            if (entityData.tags && Array.isArray(entityData.tags)) {
                // Remove any existing updated: tag and add new one
                entityData.tags = entityData.tags.filter(tag => !tag.startsWith('updated:'));
                entityData.tags.push(`updated:${new Date().toISOString()}`);
            }

            console.log('🔍 Updating entity with data:', {
                id,
                tags: entityData.tags,
                content: entityData.content,
                fullPayload: entityData
            });

            const response = await fetch(`${this.baseURL}/entities/update`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.token}`
                },
                body: JSON.stringify({ id, ...entityData })
            });

            console.log('🔍 EntityDB update response status:', response.status, response.statusText);

            if (response.ok) {
                const result = await response.json();
                console.log('🔍 EntityDB update result:', result);
                return result;
            } else {
                const errorText = await response.text();
                console.error('🔍 EntityDB update error response:', errorText);
                throw new Error(`Failed to update entity: ${response.status} ${response.statusText} - ${errorText}`);
            }
        } catch (error) {
            console.error('Update entity error:', error);
            throw error;
        }
    }

    async getEntity(id) {
        try {
            const response = await fetch(`${this.baseURL}/entities/get?id=${encodeURIComponent(id)}`, {
                headers: {
                    'Authorization': `Bearer ${this.token}`
                }
            });

            if (response.ok) {
                return await response.json();
            } else {
                throw new Error(`Failed to get entity: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Get entity error:', error);
            throw error;
        }
    }

    async queryEntities(filters = {}) {
        try {
            console.log('🔍 QueryEntities called with token:', this.token ? 'EXISTS' : 'MISSING');
            const params = new URLSearchParams();
            
            // Add filters as query parameters
            Object.entries(filters).forEach(([key, value]) => {
                if (value !== undefined && value !== null && value !== '') {
                    params.append(key, value);
                }
            });

            const url = `${this.baseURL}/entities/list?${params.toString()}`;
            console.log('🔍 Query URL:', url);

            const response = await fetch(url, {
                headers: {
                    'Authorization': `Bearer ${this.token}`
                }
            });

            console.log('🔍 Query response status:', response.status, response.statusText);

            if (response.ok) {
                const result = await response.json();
                console.log('🔍 Query result type:', typeof result, Array.isArray(result) ? 'ARRAY' : 'NOT_ARRAY');
                console.log('🔍 Query result structure:', result);
                
                // EntityDB returns {entities: [...], total: N} format
                if (result && result.entities && Array.isArray(result.entities)) {
                    console.log('🔍 Extracting entities array:', result.entities.length, 'entities');
                    return result.entities;
                } else if (Array.isArray(result)) {
                    console.log('🔍 Using direct array:', result.length, 'entities');
                    return result;
                } else {
                    console.log('🔍 Unexpected result format, returning empty array');
                    return [];
                }
            } else {
                const errorText = await response.text();
                console.error('🔍 Query failed:', response.status, errorText);
                throw new Error(`Failed to query entities: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Query entities error:', error);
            throw error;
        }
    }

    // Worcha-specific operations
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
            'status:todo',
            `priority:${priority}`,
            `created:${new Date().toISOString()}`
        ];

        if (assignee) tags.push(`assignee:${assignee}`);
        
        // Add traits as tags
        Object.entries(traits).forEach(([key, value]) => {
            if (value) tags.push(`${key}:${value}`);
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

    async createUser(name, role, email, username = null) {
        const userData = {
            tags: [
                'type:user',
                `name:${name}`,
                `role:${role}`,
                `email:${email}`
            ],
            content: [{
                type: 'profile',
                value: { name, role, email, username: username || name.toLowerCase().replace(/\s+/g, '') }
            }]
        };

        if (username) {
            userData.tags.push(`username:${username}`);
        }

        return await this.createEntity(userData);
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
        // Transform EntityDB entity into Worcha format
        const transformed = {
            id: entity.id,
            tags: entity.tags || [],
            content: entity.content || [],
            createdAt: entity.created_at,
            updatedAt: entity.updated_at
        };

        // Extract common properties from tags (handle both standard and dataspace formats)
        transformed.type = this.getTagValue(entity.tags, 'type') || this.getTagValue(entity.tags, 'worca:self:type');
        transformed.status = this.getTagValue(entity.tags, 'status') || this.getTagValue(entity.tags, 'worca:self:status');
        transformed.name = this.getTagValue(entity.tags, 'name') || this.getTagValue(entity.tags, 'worca:self:name');
        transformed.title = this.getTagValue(entity.tags, 'title') || this.getTagValue(entity.tags, 'worca:self:title');
        transformed.username = this.getTagValue(entity.tags, 'username') || this.getTagValue(entity.tags, 'worca:self:username');
        transformed.displayName = this.getTagValue(entity.tags, 'displayName') || this.getTagValue(entity.tags, 'worca:self:displayName');
        transformed.assignee = this.getTagValue(entity.tags, 'assignee') || this.getTagValue(entity.tags, 'worca:self:assignee');
        transformed.priority = this.getTagValue(entity.tags, 'priority') || this.getTagValue(entity.tags, 'worca:self:priority') || 'medium';
        transformed.role = this.getTagValue(entity.tags, 'role') || this.getTagValue(entity.tags, 'worca:self:role');
        transformed.email = this.getTagValue(entity.tags, 'email') || this.getTagValue(entity.tags, 'worca:self:email');
        
        // For users, ensure they always have a name and handle special cases
        if (transformed.type === 'user') {
            // Handle admin user role extraction from rbac tags
            if (!transformed.role && entity.tags) {
                const adminRole = entity.tags.find(t => t === 'rbac:role:admin');
                if (adminRole) {
                    transformed.role = 'Administrator';
                }
            }
            
            // Ensure name fallback logic
            if (!transformed.name && transformed.displayName) {
                transformed.name = transformed.displayName;
            } else if (!transformed.name && transformed.username) {
                transformed.name = transformed.username.charAt(0).toUpperCase() + transformed.username.slice(1);
            } else if (!transformed.name) {
                transformed.name = 'Unknown User';
            }
            
            // Ensure role fallback
            if (!transformed.role) {
                transformed.role = 'User';
            }
        }
        
        // Debug: log user transformation (can be removed in production)
        if (transformed.type === 'user') {
            console.log('🔍 Transformed user:', transformed.name, '-', transformed.role, '(ID:', transformed.id.slice(0, 8) + ')');
        }

        // Extract description from content (handle base64 encoded content)
        if (entity.content) {
            try {
                if (typeof entity.content === 'string') {
                    // Decode base64 content
                    const decodedContent = atob(entity.content);
                    
                    // Try to parse as JSON first
                    try {
                        const jsonContent = JSON.parse(decodedContent);
                        if (jsonContent.description) {
                            transformed.description = jsonContent.description;
                        } else if (Array.isArray(jsonContent)) {
                            // Look for description in array format
                            const descItem = jsonContent.find(item => item.type === 'description');
                            transformed.description = descItem?.value || '';
                        }
                    } catch (jsonError) {
                        // If not JSON, use as plain text description
                        transformed.description = decodedContent;
                    }
                } else if (Array.isArray(entity.content)) {
                    // Handle array format content
                    const descContent = entity.content.find(c => c.type === 'description');
                    transformed.description = descContent?.value || '';
                }
            } catch (e) {
                // If content parsing fails, use empty description
                transformed.description = '';
            }
        } else {
            transformed.description = '';
        }

        // Type-specific transformations
        if (transformed.type === 'task') {
            transformed.storyId = this.getTagValue(entity.tags, 'story');
            transformed.sprintId = this.getTagValue(entity.tags, 'sprint');
            transformed.storyPoints = parseInt(this.getTagValue(entity.tags, 'storyPoints')) || 3;
        }

        if (transformed.type === 'user') {
            transformed.displayName = this.getTagValue(entity.tags, 'displayName');
            transformed.role = this.getTagValue(entity.tags, 'role');
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

// Export for use in Worcha application
window.WorchaAPI = WorchaAPI;