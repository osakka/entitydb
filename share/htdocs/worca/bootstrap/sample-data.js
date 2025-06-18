// Worca Sample Data Generator
// Creates realistic sample data for different workspace templates

class SampleDataGenerator {
    constructor(config = null, client = null) {
        this.config = config || window.worcaConfig;
        this.client = client || window.entityDBClient;
        this.api = new WorcaAPI();
        
        console.log('🎭 SampleDataGenerator initialized');
    }

    async generateFromTemplate(templateConfig) {
        try {
            console.log(`🚀 Generating sample data for template: ${templateConfig.name}`);
            
            const results = {
                organizations: [],
                projects: [],
                epics: [],
                stories: [],
                tasks: [],
                users: [],
                sprints: []
            };
            
            // Load base sample data
            const sampleData = await this.loadSampleData();
            
            // Generate based on template structure
            const structure = templateConfig.structure;
            
            // 1. Create Organizations
            for (let i = 0; i < structure.organizations; i++) {
                const orgData = sampleData.organizations[i % sampleData.organizations.length];
                const org = await this.createOrganization(orgData, i);
                results.organizations.push(org);
            }
            
            // 2. Create Users
            const userPromises = [];
            for (let i = 0; i < structure.users; i++) {
                const userData = sampleData.users[i % sampleData.users.length];
                userPromises.push(this.createUser(userData, i));
            }
            results.users = await Promise.all(userPromises);
            
            // 3. Create Projects
            let projectIndex = 0;
            for (const org of results.organizations) {
                const projectsPerOrg = Math.ceil(structure.projects / structure.organizations);
                for (let i = 0; i < projectsPerOrg && projectIndex < structure.projects; i++) {
                    const projectData = sampleData.projects[projectIndex % sampleData.projects.length];
                    const project = await this.createProject(projectData, org.id, projectIndex);
                    results.projects.push(project);
                    projectIndex++;
                }
            }
            
            // 4. Create Epics
            let epicIndex = 0;
            for (const project of results.projects) {
                const epicsPerProject = Math.ceil(structure.epics / structure.projects);
                for (let i = 0; i < epicsPerProject && epicIndex < structure.epics; i++) {
                    const epic = await this.createEpic(project, epicIndex);
                    results.epics.push(epic);
                    epicIndex++;
                }
            }
            
            // 5. Create Stories
            let storyIndex = 0;
            for (const epic of results.epics) {
                const storiesPerEpic = Math.ceil(structure.stories / structure.epics);
                for (let i = 0; i < storiesPerEpic && storyIndex < structure.stories; i++) {
                    const story = await this.createStory(epic, storyIndex);
                    results.stories.push(story);
                    storyIndex++;
                }
            }
            
            // 6. Create Tasks
            let taskIndex = 0;
            for (const story of results.stories) {
                const tasksPerStory = Math.ceil(structure.tasks / structure.stories);
                for (let i = 0; i < tasksPerStory && taskIndex < structure.tasks; i++) {
                    const assignee = results.users[taskIndex % results.users.length];
                    const task = await this.createTask(story, assignee, taskIndex);
                    results.tasks.push(task);
                    taskIndex++;
                }
            }
            
            // 7. Create Sprints
            for (let i = 0; i < structure.sprints; i++) {
                const sprint = await this.createSprint(results.projects[0], results.tasks, i);
                results.sprints.push(sprint);
            }
            
            // 8. Assign some tasks to sprints
            await this.assignTasksToSprints(results.tasks, results.sprints);
            
            console.log(`✅ Sample data generation complete:`, {
                organizations: results.organizations.length,
                projects: results.projects.length,
                epics: results.epics.length,
                stories: results.stories.length,
                tasks: results.tasks.length,
                users: results.users.length,
                sprints: results.sprints.length
            });
            
            return results;
            
        } catch (error) {
            console.error('❌ Sample data generation failed:', error);
            throw error;
        }
    }

    async loadSampleData() {
        try {
            const response = await fetch('/worca/config/defaults.json');
            if (!response.ok) {
                throw new Error('Failed to load sample data configuration');
            }
            
            const defaults = await response.json();
            return defaults.sampleData;
            
        } catch (error) {
            console.error('Failed to load sample data:', error);
            // Return minimal fallback data
            return {
                organizations: [
                    { name: 'TechCorp Solutions', description: 'Technology solutions provider', industry: 'Technology' }
                ],
                projects: [
                    { name: 'Mobile App', description: 'Mobile application project', priority: 'high' }
                ],
                users: [
                    { username: 'admin', displayName: 'Administrator', role: 'Manager', email: 'admin@company.com' }
                ]
            };
        }
    }

    async createOrganization(orgData, index) {
        try {
            const name = `${orgData.name}${index > 0 ? ` ${index + 1}` : ''}`;
            const description = `${orgData.description} (${orgData.industry || 'Technology'})`;
            
            return await this.api.createOrganization(name, description);
            
        } catch (error) {
            console.error(`Failed to create organization ${index}:`, error);
            throw error;
        }
    }

    async createUser(userData, index) {
        try {
            const username = `${userData.username}${index > 0 ? index + 1 : ''}`;
            const displayName = `${userData.displayName}${index > 0 ? ` ${index + 1}` : ''}`;
            const email = `${username}@${this.generateCompanyDomain()}`;
            
            // Create user with all properties at once to avoid update conflicts
            const userEntity = {
                tags: [
                    'type:user',
                    `username:${username}`,
                    `displayName:${displayName}`,
                    `role:${userData.role}`,
                    `email:${email}`,
                    `timezone:${userData.timezone || 'America/New_York'}`,
                    `skills:${(userData.skills || ['General']).join(',')}`,
                    'status:active'
                ],
                content: [{
                    type: 'profile',
                    value: JSON.stringify({
                        username,
                        displayName,
                        role: userData.role,
                        email,
                        timezone: userData.timezone || 'America/New_York',
                        skills: userData.skills || ['General'],
                        createdAt: new Date().toISOString()
                    })
                }]
            };
            
            const user = await this.api.createEntity(userEntity);
            return { ...user, email, skills: userData.skills || [] };
            
        } catch (error) {
            console.error(`Failed to create user ${index}:`, error);
            throw error;
        }
    }

    async createProject(projectData, orgId, index) {
        try {
            const name = `${projectData.name}${index > 0 ? ` ${index + 1}` : ''}`;
            const description = `${projectData.description} (Priority: ${projectData.priority || 'medium'})`;
            
            return await this.api.createProject(name, orgId, description);
            
        } catch (error) {
            console.error(`Failed to create project ${index}:`, error);
            throw error;
        }
    }

    async createEpic(project, index) {
        try {
            const epicTitles = [
                'User Authentication System',
                'Data Management Platform',
                'User Interface Redesign',
                'Performance Optimization',
                'Security Enhancement',
                'Mobile Integration',
                'Analytics Dashboard',
                'API Development',
                'Testing Framework',
                'Documentation Portal'
            ];
            
            const title = epicTitles[index % epicTitles.length];
            const description = `Epic ${index + 1}: ${title} for ${project.name || 'project'}`;
            
            return await this.api.createEpic(title, project.id, description);
            
        } catch (error) {
            console.error(`Failed to create epic ${index}:`, error);
            throw error;
        }
    }

    async createStory(epic, index) {
        try {
            const storyTitles = [
                'Login Form Implementation',
                'User Registration Flow',
                'Password Reset Feature',
                'Profile Management',
                'Data Validation',
                'Error Handling',
                'UI Components',
                'Responsive Design',
                'Search Functionality',
                'Export Features'
            ];
            
            const title = storyTitles[index % storyTitles.length];
            const description = `Story ${index + 1}: Implement ${title.toLowerCase()} as part of the epic`;
            
            return await this.api.createStory(title, epic.id, description);
            
        } catch (error) {
            console.error(`Failed to create story ${index}:`, error);
            throw error;
        }
    }

    async createTask(story, assignee, index) {
        try {
            const taskTypes = [
                'Design mockups for',
                'Implement frontend for',
                'Create backend API for',
                'Write tests for',
                'Document functionality for',
                'Review and refactor',
                'Deploy and configure',
                'Optimize performance of',
                'Fix bugs in',
                'Update styles for'
            ];
            
            const taskType = taskTypes[index % taskTypes.length];
            const title = `${taskType} ${story.title || 'component'}`;
            const description = `Task ${index + 1}: ${title} - assigned to ${assignee.displayName || assignee.username}`;
            
            const priorities = ['low', 'medium', 'high'];
            const statuses = ['todo', 'doing', 'review', 'done'];
            
            const priority = priorities[index % priorities.length];
            const status = statuses[index % statuses.length];
            
            const task = await this.api.createTask(
                title,
                description,
                assignee.username || assignee.id,
                priority,
                {
                    story: story.id,
                    status: status,
                    storyPoints: Math.floor(Math.random() * 8) + 1
                }
            );
            
            // Status is already set in createTask through traits parameter
            
            return task;
            
        } catch (error) {
            console.error(`Failed to create task ${index}:`, error);
            throw error;
        }
    }

    async createSprint(project, tasks, index) {
        try {
            const sprintNames = [
                'Foundation Sprint',
                'Feature Development Sprint',
                'Integration Sprint',
                'Polish & Testing Sprint',
                'Release Preparation Sprint',
                'Bug Fix Sprint',
                'Performance Sprint',
                'Security Sprint'
            ];
            
            const name = sprintNames[index % sprintNames.length];
            const startDate = new Date(Date.now() - (index * 14 * 24 * 60 * 60 * 1000)); // 2 weeks ago per sprint
            const endDate = new Date(startDate.getTime() + (14 * 24 * 60 * 60 * 1000)); // 2 weeks duration
            const goal = `Complete ${name.toLowerCase()} objectives for ${project.name || 'project'}`;
            
            return await this.api.createSprint(name, startDate, endDate, goal);
            
        } catch (error) {
            console.error(`Failed to create sprint ${index}:`, error);
            throw error;
        }
    }

    async assignTasksToSprints(tasks, sprints) {
        try {
            if (!sprints.length) return;
            
            // Assign about 60% of tasks to sprints
            const tasksToAssign = Math.floor(tasks.length * 0.6);
            
            for (let i = 0; i < tasksToAssign; i++) {
                const task = tasks[i];
                const sprint = sprints[i % sprints.length];
                
                try {
                    await this.api.addTaskToSprint(task.id, sprint.id);
                } catch (error) {
                    console.warn(`Failed to assign task ${task.id} to sprint ${sprint.id}:`, error);
                }
            }
            
            console.log(`✅ Assigned ${tasksToAssign} tasks to ${sprints.length} sprints`);
            
        } catch (error) {
            console.error('Failed to assign tasks to sprints:', error);
        }
    }

    generateCompanyDomain() {
        const domains = [
            'techcorp.com',
            'startupflow.io',
            'innovation.dev',
            'company.com',
            'workspace.org',
            'team.co'
        ];
        
        return domains[Math.floor(Math.random() * domains.length)];
    }

    // Quick Sample Data Methods
    async generateMinimalSample() {
        try {
            console.log('🚀 Generating minimal sample data...');
            
            // Create single organization
            const org = await this.api.createOrganization(
                'Sample Organization',
                'A sample organization for testing Worca functionality'
            );
            
            // Create sample users
            const users = await Promise.all([
                this.api.createUser('alice', 'Alice Johnson', 'manager'),
                this.api.createUser('bob', 'Bob Smith', 'developer'),
                this.api.createUser('carol', 'Carol Davis', 'designer')
            ]);
            
            // Create project
            const project = await this.api.createProject(
                'Sample Project',
                org.id,
                'A sample project for testing Worca functionality'
            );
            
            // Create epic
            const epic = await this.api.createEpic(
                'Sample Epic',
                project.id,
                'A sample epic containing sample stories and tasks'
            );
            
            // Create story
            const story = await this.api.createStory(
                'Sample Story',
                epic.id,
                'A sample story containing sample tasks'
            );
            
            // Create tasks
            const tasks = await Promise.all([
                this.api.createTask('Design sample interface', 'Create mockups and wireframes', users[2].username, 'high', { story: story.id }),
                this.api.createTask('Implement sample feature', 'Code the main functionality', users[1].username, 'medium', { story: story.id }),
                this.api.createTask('Test sample feature', 'Write and run tests', users[1].username, 'low', { story: story.id })
            ]);
            
            // Task statuses are set during creation
            
            const results = {
                organizations: [org],
                projects: [project],
                epics: [epic],
                stories: [story],
                tasks,
                users,
                sprints: []
            };
            
            console.log('✅ Minimal sample data generated successfully');
            return results;
            
        } catch (error) {
            console.error('❌ Failed to generate minimal sample data:', error);
            throw error;
        }
    }

    async generateDemoData() {
        try {
            console.log('🚀 Generating comprehensive demo data...');
            
            // Use startup template as base
            const template = {
                name: 'demo',
                structure: {
                    organizations: 2,
                    projects: 4,
                    epics: 8,
                    stories: 16,
                    tasks: 40,
                    users: 12,
                    sprints: 4
                }
            };
            
            return await this.generateFromTemplate(template);
            
        } catch (error) {
            console.error('❌ Failed to generate demo data:', error);
            throw error;
        }
    }

    // Utility method to clear all sample data
    async clearSampleData() {
        try {
            console.log('🗑️ Clearing existing sample data...');
            
            const namespace = this.config.get('dataset.namespace');
            const entities = await this.client.queryEntities({
                tag: `namespace:${namespace}`
            });
            
            let deleted = 0;
            for (const entity of entities) {
                try {
                    // Mark as deleted instead of actual deletion for data integrity
                    await this.client.updateEntity(entity.id, {
                        tags: [...entity.tags.filter(t => !t.startsWith('status:')), 'status:deleted']
                    });
                    deleted++;
                } catch (error) {
                    console.warn(`Failed to delete entity ${entity.id}:`, error);
                }
            }
            
            console.log(`✅ Cleared ${deleted} entities from workspace`);
            return deleted;
            
        } catch (error) {
            console.error('❌ Failed to clear sample data:', error);
            throw error;
        }
    }
}

// Global instance
window.sampleDataGenerator = new SampleDataGenerator();

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = SampleDataGenerator;
}