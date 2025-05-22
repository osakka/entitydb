// Worcha EntityDB API Integration
// Complete backend functionality using EntityDB

class WorchaAPI {
    constructor() {
        this.baseURL = '/api/v1';
        this.token = localStorage.getItem('authToken');
    }

    // Authentication
    async login(username, password) {
        try {
            const response = await fetch(`${this.baseURL}/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username, password })
            });

            if (response.ok) {
                const data = await response.json();
                this.token = data.token;
                localStorage.setItem('authToken', this.token);
                return data;
            } else {
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
            const response = await fetch(`${this.baseURL}/entities/create`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.token}`
                },
                body: JSON.stringify(entityData)
            });

            if (response.ok) {
                return await response.json();
            } else {
                throw new Error(`Failed to create entity: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Create entity error:', error);
            throw error;
        }
    }

    async updateEntity(id, entityData) {
        try {
            const response = await fetch(`${this.baseURL}/entities/update`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.token}`
                },
                body: JSON.stringify({ id, ...entityData })
            });

            if (response.ok) {
                return await response.json();
            } else {
                throw new Error(`Failed to update entity: ${response.statusText}`);
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
            const params = new URLSearchParams();
            
            // Add filters as query parameters
            Object.entries(filters).forEach(([key, value]) => {
                if (value !== undefined && value !== null && value !== '') {
                    params.append(key, value);
                }
            });

            const response = await fetch(`${this.baseURL}/entities/list?${params.toString()}`, {
                headers: {
                    'Authorization': `Bearer ${this.token}`
                }
            });

            if (response.ok) {
                return await response.json();
            } else {
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

    async createTask(title, description, storyId = null, assignee = null, priority = 'medium') {
        const tags = [
            'type:task',
            `title:${title}`,
            'status:todo',
            `priority:${priority}`,
            `created:${new Date().toISOString()}`
        ];

        if (storyId) tags.push(`story:${storyId}`);
        if (assignee) tags.push(`assignee:${assignee}`);

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
        const result = await this.queryEntities();
        return this.filterEntitiesByType(result.entities || [], 'organization');
    }

    async getProjects(orgId = null) {
        const result = await this.queryEntities();
        let projects = this.filterEntitiesByType(result.entities || [], 'project');
        
        if (orgId) {
            projects = projects.filter(p => this.getTagValue(p.tags, 'org') === orgId);
        }
        
        return projects;
    }

    async getEpics(projectId = null) {
        const result = await this.queryEntities();
        let epics = this.filterEntitiesByType(result.entities || [], 'epic');
        
        if (projectId) {
            epics = epics.filter(e => this.getTagValue(e.tags, 'project') === projectId);
        }
        
        return epics;
    }

    async getStories(epicId = null) {
        const result = await this.queryEntities();
        let stories = this.filterEntitiesByType(result.entities || [], 'story');
        
        if (epicId) {
            stories = stories.filter(s => this.getTagValue(s.tags, 'epic') === epicId);
        }
        
        return stories;
    }

    async getTasks(filters = {}) {
        const result = await this.queryEntities();
        let tasks = this.filterEntitiesByType(result.entities || [], 'task');
        
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
        return this.filterEntitiesByType(result.entities || [], 'user');
    }

    async getSprints() {
        const result = await this.queryEntities();
        return this.filterEntitiesByType(result.entities || [], 'sprint');
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
        return entities.filter(entity => 
            entity.tags && entity.tags.some(tag => tag === `type:${type}`)
        ).map(entity => this.transformEntity(entity));
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

        // Extract common properties from tags
        transformed.type = this.getTagValue(entity.tags, 'type');
        transformed.status = this.getTagValue(entity.tags, 'status');
        transformed.name = this.getTagValue(entity.tags, 'name');
        transformed.title = this.getTagValue(entity.tags, 'title');
        transformed.username = this.getTagValue(entity.tags, 'username');
        transformed.assignee = this.getTagValue(entity.tags, 'assignee');
        transformed.priority = this.getTagValue(entity.tags, 'priority') || 'medium';

        // Extract description from content
        const descContent = entity.content?.find(c => c.type === 'description');
        transformed.description = descContent?.value || '';

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
        const tag = tags.find(t => t.startsWith(`${key}:`));
        return tag ? tag.substring(key.length + 1) : null;
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