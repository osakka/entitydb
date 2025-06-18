// Worca Dataset Manager
// Handles workspace creation, validation, and management in EntityDB

class DatasetManager {
    constructor(config = null, client = null) {
        this.config = config || window.worcaConfig;
        this.client = client || window.entityDBClient;
        this.currentWorkspace = null;
        this.workspaceCache = new Map();
        
        console.log('📊 DatasetManager initialized');
    }

    // Workspace Management
    async listWorkspaces() {
        try {
            console.log('📋 Listing available workspaces...');
            
            // Query for namespace entities to find workspaces
            const entities = await this.client.queryEntities({
                tag: 'type:workspace'
            });
            
            const workspaces = entities.map(entity => {
                const name = this.getTagValue(entity.tags, 'name');
                const namespace = this.getTagValue(entity.tags, 'namespace');
                const description = this.getTagValue(entity.tags, 'description');
                const created = this.getTagValue(entity.tags, 'created');
                
                return {
                    id: entity.id,
                    name: name || namespace,
                    namespace,
                    description: description || '',
                    created: created || entity.createdAt,
                    entityCount: 0 // Will be populated separately
                };
            });
            
            // Get entity counts for each workspace
            for (const workspace of workspaces) {
                try {
                    const entities = await this.client.queryEntities({
                        tag: `namespace:${workspace.namespace}`
                    });
                    workspace.entityCount = entities.length;
                } catch (error) {
                    console.warn(`Failed to count entities for workspace ${workspace.name}:`, error);
                    workspace.entityCount = 0;
                }
            }
            
            console.log(`✅ Found ${workspaces.length} workspaces`);
            return workspaces;
            
        } catch (error) {
            console.error('❌ Failed to list workspaces:', error);
            return [];
        }
    }

    async createWorkspace(name, options = {}) {
        try {
            console.log(`🏗️ Creating workspace: ${name}`);
            
            const namespace = options.namespace || this.sanitizeNamespace(name);
            const description = options.description || `Worca workspace: ${name}`;
            const template = options.template || 'startup';
            
            // Check if workspace already exists
            const existing = await this.getWorkspace(namespace);
            if (existing) {
                throw new Error(`Workspace '${name}' already exists`);
            }
            
            // Create workspace entity
            const workspaceEntity = {
                tags: [
                    'type:workspace',
                    `name:${name}`,
                    `namespace:${namespace}`,
                    `description:${description}`,
                    `template:${template}`,
                    'status:active',
                    `created:${new Date().toISOString()}`
                ],
                content: {
                    name,
                    namespace,
                    description,
                    template,
                    created: new Date().toISOString(),
                    settings: options.settings || {}
                }
            };
            
            const result = await this.client.createEntity(workspaceEntity);
            console.log(`✅ Workspace created: ${result.id}`);
            
            // Initialize workspace with template data if requested
            if (options.initializeWithSample) {
                await this.initializeWorkspaceData(namespace, template);
            }
            
            // Cache the workspace
            this.workspaceCache.set(namespace, result);
            
            return {
                id: result.id,
                name,
                namespace,
                description,
                template,
                created: new Date().toISOString()
            };
            
        } catch (error) {
            console.error(`❌ Failed to create workspace '${name}':`, error);
            throw error;
        }
    }

    async getWorkspace(namespaceOrName) {
        try {
            // Check cache first
            if (this.workspaceCache.has(namespaceOrName)) {
                return this.workspaceCache.get(namespaceOrName);
            }
            
            // Query by namespace or name
            const entities = await this.client.queryEntities({
                tag: 'type:workspace'
            });
            
            const workspace = entities.find(entity => {
                const namespace = this.getTagValue(entity.tags, 'namespace');
                const name = this.getTagValue(entity.tags, 'name');
                return namespace === namespaceOrName || name === namespaceOrName;
            });
            
            if (workspace) {
                // Cache it
                const namespace = this.getTagValue(workspace.tags, 'namespace');
                this.workspaceCache.set(namespace, workspace);
            }
            
            return workspace || null;
            
        } catch (error) {
            console.error(`❌ Failed to get workspace '${namespaceOrName}':`, error);
            return null;
        }
    }

    async switchWorkspace(namespace) {
        try {
            console.log(`🔄 Switching to workspace: ${namespace}`);
            
            // Validate workspace exists
            const workspace = await this.getWorkspace(namespace);
            if (!workspace) {
                throw new Error(`Workspace '${namespace}' not found`);
            }
            
            // Update configuration
            this.config.set('dataset.name', namespace);
            this.config.set('dataset.namespace', namespace);
            
            // Update current workspace
            this.currentWorkspace = workspace;
            
            // Emit workspace changed event
            this.config.emit('workspace-changed', {
                namespace,
                workspace: this.currentWorkspace
            });
            
            console.log(`✅ Switched to workspace: ${namespace}`);
            return workspace;
            
        } catch (error) {
            console.error(`❌ Failed to switch to workspace '${namespace}':`, error);
            throw error;
        }
    }

    async deleteWorkspace(namespace, options = {}) {
        try {
            console.log(`🗑️ Deleting workspace: ${namespace}`);
            
            if (!options.force) {
                // Check if workspace has data
                const entities = await this.client.queryEntities({
                    tag: `namespace:${namespace}`
                });
                
                if (entities.length > 1) { // More than just the workspace entity
                    throw new Error(`Workspace '${namespace}' contains data. Use force option to delete.`);
                }
            }
            
            // Get workspace entity
            const workspace = await this.getWorkspace(namespace);
            if (!workspace) {
                throw new Error(`Workspace '${namespace}' not found`);
            }
            
            // If force delete, remove all entities in the namespace
            if (options.force) {
                const entities = await this.client.queryEntities({
                    tag: `namespace:${namespace}`
                });
                
                console.log(`🗑️ Deleting ${entities.length} entities from workspace...`);
                
                for (const entity of entities) {
                    try {
                        // Update entity to mark as deleted
                        await this.client.updateEntity(entity.id, {
                            tags: [...entity.tags.filter(t => !t.startsWith('status:')), 'status:deleted']
                        });
                    } catch (error) {
                        console.warn(`Failed to delete entity ${entity.id}:`, error);
                    }
                }
            }
            
            // Mark workspace as deleted
            await this.client.updateEntity(workspace.id, {
                tags: [...workspace.tags.filter(t => !t.startsWith('status:')), 'status:deleted']
            });
            
            // Remove from cache
            this.workspaceCache.delete(namespace);
            
            // If this was the current workspace, switch to default
            if (this.currentWorkspace && this.getTagValue(this.currentWorkspace.tags, 'namespace') === namespace) {
                const defaultNamespace = this.config.get('dataset.namespace');
                if (defaultNamespace !== namespace) {
                    await this.switchWorkspace(defaultNamespace);
                } else {
                    this.currentWorkspace = null;
                }
            }
            
            console.log(`✅ Workspace '${namespace}' deleted`);
            return true;
            
        } catch (error) {
            console.error(`❌ Failed to delete workspace '${namespace}':`, error);
            throw error;
        }
    }

    // Workspace Data Initialization
    async initializeWorkspaceData(namespace, template = 'startup') {
        try {
            console.log(`🚀 Initializing workspace '${namespace}' with template '${template}'`);
            
            // Load template configuration
            const templateConfig = await this.getTemplateConfig(template);
            if (!templateConfig) {
                throw new Error(`Template '${template}' not found`);
            }
            
            // Switch to the workspace namespace for data creation
            const originalNamespace = this.config.get('dataset.namespace');
            this.config.set('dataset.namespace', namespace);
            
            try {
                // Create sample data based on template
                const sampleData = new SampleDataGenerator(this.config, this.client);
                const results = await sampleData.generateFromTemplate(templateConfig);
                
                console.log(`✅ Initialized workspace with ${Object.keys(results).length} entity types`);
                return results;
                
            } finally {
                // Restore original namespace
                this.config.set('dataset.namespace', originalNamespace);
            }
            
        } catch (error) {
            console.error(`❌ Failed to initialize workspace data:`, error);
            throw error;
        }
    }

    async validateWorkspace(namespace) {
        try {
            console.log(`🔍 Validating workspace: ${namespace}`);
            
            const validation = {
                valid: true,
                errors: [],
                warnings: [],
                stats: {
                    entities: 0,
                    organizations: 0,
                    projects: 0,
                    tasks: 0,
                    users: 0
                }
            };
            
            // Get all entities in the workspace
            const entities = await this.client.queryEntities({
                tag: `namespace:${namespace}`
            });
            
            validation.stats.entities = entities.length;
            
            // Count entity types
            const typeCounts = {};
            for (const entity of entities) {
                const type = this.getTagValue(entity.tags, 'type');
                if (type) {
                    typeCounts[type] = (typeCounts[type] || 0) + 1;
                }
            }
            
            validation.stats.organizations = typeCounts.organization || 0;
            validation.stats.projects = typeCounts.project || 0;
            validation.stats.tasks = typeCounts.task || 0;
            validation.stats.users = typeCounts.user || 0;
            
            // Validation rules
            if (validation.stats.entities === 0) {
                validation.errors.push('Workspace is empty');
                validation.valid = false;
            }
            
            if (validation.stats.organizations === 0) {
                validation.warnings.push('No organizations found');
            }
            
            if (validation.stats.users === 0) {
                validation.warnings.push('No users found');
            }
            
            // Check for orphaned entities
            const orphanedTasks = entities.filter(entity => {
                const type = this.getTagValue(entity.tags, 'type');
                if (type !== 'task') return false;
                
                const assignee = this.getTagValue(entity.tags, 'assignee');
                if (assignee) {
                    const userExists = entities.some(user => 
                        this.getTagValue(user.tags, 'type') === 'user' && 
                        (this.getTagValue(user.tags, 'username') === assignee || user.id === assignee)
                    );
                    return !userExists;
                }
                return false;
            });
            
            if (orphanedTasks.length > 0) {
                validation.warnings.push(`${orphanedTasks.length} tasks assigned to non-existent users`);
            }
            
            console.log(`✅ Workspace validation complete: ${validation.valid ? 'VALID' : 'INVALID'}`);
            return validation;
            
        } catch (error) {
            console.error(`❌ Failed to validate workspace '${namespace}':`, error);
            return {
                valid: false,
                errors: [error.message],
                warnings: [],
                stats: {}
            };
        }
    }

    // Template Management
    async getTemplateConfig(templateName) {
        try {
            // Load from defaults.json
            const response = await fetch('/worca/config/defaults.json');
            if (!response.ok) {
                throw new Error('Failed to load defaults configuration');
            }
            
            const defaults = await response.json();
            const template = defaults.workspaces.templates.find(t => t.name === templateName);
            
            if (!template) {
                throw new Error(`Template '${templateName}' not found`);
            }
            
            return template;
            
        } catch (error) {
            console.error(`Failed to get template config for '${templateName}':`, error);
            return null;
        }
    }

    async getAvailableTemplates() {
        try {
            const response = await fetch('/worca/config/defaults.json');
            if (!response.ok) {
                throw new Error('Failed to load defaults configuration');
            }
            
            const defaults = await response.json();
            return defaults.workspaces.templates || [];
            
        } catch (error) {
            console.error('Failed to get available templates:', error);
            return [];
        }
    }

    // Utility Methods
    sanitizeNamespace(name) {
        return name
            .toLowerCase()
            .replace(/[^a-z0-9]+/g, '-')
            .replace(/^-+|-+$/g, '')
            .substring(0, 50);
    }

    getTagValue(tags, key) {
        if (!tags) return null;
        const tag = tags.find(t => t.startsWith(`${key}:`));
        return tag ? tag.substring(key.length + 1) : null;
    }

    // Export/Import
    async exportWorkspace(namespace) {
        try {
            console.log(`📤 Exporting workspace: ${namespace}`);
            
            const entities = await this.client.queryEntities({
                tag: `namespace:${namespace}`
            });
            
            const exportData = {
                version: this.config.version,
                namespace,
                timestamp: new Date().toISOString(),
                entityCount: entities.length,
                entities: entities.map(entity => ({
                    id: entity.id,
                    tags: entity.tags,
                    content: entity.content,
                    createdAt: entity.createdAt,
                    updatedAt: entity.updatedAt
                }))
            };
            
            console.log(`✅ Exported ${entities.length} entities from workspace '${namespace}'`);
            return exportData;
            
        } catch (error) {
            console.error(`❌ Failed to export workspace '${namespace}':`, error);
            throw error;
        }
    }

    async importWorkspace(exportData, options = {}) {
        try {
            console.log(`📥 Importing workspace: ${exportData.namespace}`);
            
            const namespace = options.namespace || exportData.namespace;
            const entities = exportData.entities || [];
            
            // Validate import data
            if (!entities.length) {
                throw new Error('No entities to import');
            }
            
            // Check if workspace exists
            if (!options.overwrite) {
                const existing = await this.getWorkspace(namespace);
                if (existing) {
                    throw new Error(`Workspace '${namespace}' already exists. Use overwrite option.`);
                }
            }
            
            // Update namespace in entity tags if different
            const updatedEntities = entities.map(entity => {
                if (namespace !== exportData.namespace) {
                    const updatedTags = entity.tags.map(tag => 
                        tag.startsWith(`namespace:${exportData.namespace}`) 
                            ? `namespace:${namespace}` 
                            : tag
                    );
                    return { ...entity, tags: updatedTags };
                }
                return entity;
            });
            
            // Switch to target namespace
            const originalNamespace = this.config.get('dataset.namespace');
            this.config.set('dataset.namespace', namespace);
            
            try {
                // Create entities
                let created = 0;
                for (const entity of updatedEntities) {
                    try {
                        await this.client.createEntity({
                            tags: entity.tags,
                            content: entity.content
                        });
                        created++;
                    } catch (error) {
                        console.warn(`Failed to import entity ${entity.id}:`, error);
                    }
                }
                
                console.log(`✅ Imported ${created}/${entities.length} entities to workspace '${namespace}'`);
                return { created, total: entities.length, namespace };
                
            } finally {
                // Restore original namespace
                this.config.set('dataset.namespace', originalNamespace);
            }
            
        } catch (error) {
            console.error(`❌ Failed to import workspace:`, error);
            throw error;
        }
    }

    // Cleanup
    destroy() {
        this.workspaceCache.clear();
        this.currentWorkspace = null;
    }
}

// Global instance
window.datasetManager = new DatasetManager();

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = DatasetManager;
}