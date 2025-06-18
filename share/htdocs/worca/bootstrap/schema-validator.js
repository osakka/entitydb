// Worca Schema Validator
// Validates worca entity structure and relationships in EntityDB

class SchemaValidator {
    constructor(config = null, client = null) {
        this.config = config || window.worcaConfig;
        this.client = client || window.entityDBClient;
        
        // Define worca entity schemas
        this.schemas = {
            workspace: {
                required: ['type', 'name', 'namespace'],
                optional: ['description', 'template', 'status', 'created'],
                type: 'workspace'
            },
            organization: {
                required: ['type', 'name'],
                optional: ['description', 'industry', 'size', 'status', 'created'],
                type: 'organization'
            },
            project: {
                required: ['type', 'name', 'org'],
                optional: ['description', 'priority', 'status', 'created'],
                type: 'project',
                relationships: {
                    org: 'organization'
                }
            },
            epic: {
                required: ['type', 'title', 'project'],
                optional: ['description', 'status', 'priority', 'created'],
                type: 'epic',
                relationships: {
                    project: 'project'
                }
            },
            story: {
                required: ['type', 'title', 'epic'],
                optional: ['description', 'status', 'priority', 'storyPoints', 'created'],
                type: 'story',
                relationships: {
                    epic: 'epic'
                }
            },
            task: {
                required: ['type', 'title'],
                optional: ['description', 'status', 'priority', 'assignee', 'story', 'sprint', 'storyPoints', 'created'],
                type: 'task',
                relationships: {
                    story: 'story',
                    sprint: 'sprint',
                    assignee: 'user'
                }
            },
            user: {
                required: ['type'],
                optional: ['username', 'displayName', 'name', 'email', 'role', 'status', 'timezone', 'skills', 'created'],
                type: 'user'
            },
            sprint: {
                required: ['type', 'name'],
                optional: ['status', 'capacity', 'start', 'end', 'goal', 'created'],
                type: 'sprint'
            }
        };
        
        console.log('🔍 SchemaValidator initialized with', Object.keys(this.schemas).length, 'schemas');
    }

    // Entity Validation
    async validateEntity(entity) {
        try {
            const type = this.getTagValue(entity.tags, 'type');
            if (!type) {
                return {
                    valid: false,
                    errors: ['Entity missing type tag'],
                    warnings: [],
                    type: null
                };
            }

            const schema = this.schemas[type];
            if (!schema) {
                return {
                    valid: false,
                    errors: [`Unknown entity type: ${type}`],
                    warnings: [],
                    type
                };
            }

            const validation = {
                valid: true,
                errors: [],
                warnings: [],
                type
            };

            // Check required tags
            for (const requiredTag of schema.required) {
                const value = this.getTagValue(entity.tags, requiredTag);
                if (!value) {
                    validation.errors.push(`Missing required tag: ${requiredTag}`);
                    validation.valid = false;
                }
            }

            // Validate tag formats
            const tagValidation = this.validateTags(entity.tags, schema);
            validation.errors.push(...tagValidation.errors);
            validation.warnings.push(...tagValidation.warnings);

            // Validate content structure
            const contentValidation = this.validateContent(entity.content, schema);
            validation.warnings.push(...contentValidation.warnings);

            if (validation.errors.length > 0) {
                validation.valid = false;
            }

            return validation;

        } catch (error) {
            console.error('Entity validation error:', error);
            return {
                valid: false,
                errors: [error.message],
                warnings: [],
                type: null
            };
        }
    }

    async validateWorkspace(namespace) {
        try {
            console.log(`🔍 Validating workspace schema: ${namespace}`);

            const validation = {
                valid: true,
                errors: [],
                warnings: [],
                entityValidations: [],
                relationships: {
                    valid: true,
                    errors: [],
                    warnings: []
                },
                stats: {
                    total: 0,
                    valid: 0,
                    invalid: 0,
                    types: {}
                }
            };

            // Get all entities in workspace
            const entities = await this.client.queryEntities({
                tag: `namespace:${namespace}`
            });

            validation.stats.total = entities.length;

            if (entities.length === 0) {
                validation.warnings.push('Workspace is empty');
                return validation;
            }

            // Validate each entity
            for (const entity of entities) {
                const entityValidation = await this.validateEntity(entity);
                validation.entityValidations.push({
                    id: entity.id,
                    ...entityValidation
                });

                // Update stats
                if (entityValidation.valid) {
                    validation.stats.valid++;
                } else {
                    validation.stats.invalid++;
                    validation.valid = false;
                }

                // Count types
                if (entityValidation.type) {
                    validation.stats.types[entityValidation.type] = 
                        (validation.stats.types[entityValidation.type] || 0) + 1;
                }
            }

            // Validate relationships
            const relationshipValidation = await this.validateRelationships(entities);
            validation.relationships = relationshipValidation;

            if (!relationshipValidation.valid) {
                validation.valid = false;
            }

            console.log(`✅ Workspace validation complete: ${validation.valid ? 'VALID' : 'INVALID'}`);
            console.log(`📊 Stats: ${validation.stats.valid}/${validation.stats.total} entities valid`);

            return validation;

        } catch (error) {
            console.error('Workspace validation error:', error);
            return {
                valid: false,
                errors: [error.message],
                warnings: [],
                entityValidations: [],
                relationships: { valid: false, errors: [error.message], warnings: [] },
                stats: { total: 0, valid: 0, invalid: 0, types: {} }
            };
        }
    }

    validateTags(tags, schema) {
        const validation = {
            errors: [],
            warnings: []
        };

        if (!tags || !Array.isArray(tags)) {
            validation.errors.push('Entity tags must be an array');
            return validation;
        }

        // Check for namespace tag
        const hasNamespace = tags.some(tag => tag.startsWith('namespace:'));
        if (!hasNamespace) {
            validation.warnings.push('Entity missing namespace tag');
        }

        // Validate tag formats
        for (const tag of tags) {
            if (typeof tag !== 'string') {
                validation.errors.push(`Invalid tag format: ${typeof tag}`);
                continue;
            }

            if (!tag.includes(':')) {
                validation.warnings.push(`Tag missing colon separator: ${tag}`);
                continue;
            }

            const [key, value] = tag.split(':', 2);
            if (!key || !value) {
                validation.warnings.push(`Invalid tag format: ${tag}`);
            }
        }

        // Validate specific tag values
        const type = this.getTagValue(tags, 'type');
        if (type && type !== schema.type) {
            validation.errors.push(`Type mismatch: expected ${schema.type}, got ${type}`);
        }

        const status = this.getTagValue(tags, 'status');
        if (status && !this.isValidStatus(status, schema.type)) {
            validation.warnings.push(`Invalid status for ${schema.type}: ${status}`);
        }

        const priority = this.getTagValue(tags, 'priority');
        if (priority && !this.isValidPriority(priority)) {
            validation.warnings.push(`Invalid priority: ${priority}`);
        }

        return validation;
    }

    validateContent(content, schema) {
        const validation = {
            warnings: []
        };

        if (content === null || content === undefined) {
            // Content is optional for most entities
            return validation;
        }

        // Validate content structure based on type
        if (typeof content === 'string') {
            // String content is valid
            return validation;
        }

        if (Array.isArray(content)) {
            // Array content should have objects with type and value
            for (const item of content) {
                if (!item.type || !item.value) {
                    validation.warnings.push('Content array items should have type and value properties');
                }
            }
            return validation;
        }

        if (typeof content === 'object') {
            // Object content is valid
            return validation;
        }

        validation.warnings.push(`Unexpected content type: ${typeof content}`);
        return validation;
    }

    async validateRelationships(entities) {
        try {
            const validation = {
                valid: true,
                errors: [],
                warnings: []
            };

            // Create entity lookup maps
            const entityMap = new Map();
            const typeMap = new Map();

            for (const entity of entities) {
                entityMap.set(entity.id, entity);
                
                const type = this.getTagValue(entity.tags, 'type');
                if (type) {
                    if (!typeMap.has(type)) {
                        typeMap.set(type, []);
                    }
                    typeMap.get(type).push(entity);
                }
            }

            // Validate relationships for each entity
            for (const entity of entities) {
                const type = this.getTagValue(entity.tags, 'type');
                const schema = this.schemas[type];

                if (!schema || !schema.relationships) {
                    continue;
                }

                // Check each relationship
                for (const [relationTag, expectedType] of Object.entries(schema.relationships)) {
                    const relationValue = this.getTagValue(entity.tags, relationTag);
                    
                    if (!relationValue) {
                        // Optional relationship
                        continue;
                    }

                    // Check if referenced entity exists
                    const referencedEntity = entityMap.get(relationValue);
                    if (!referencedEntity) {
                        // Try to find by other identifiers
                        const foundEntity = entities.find(e => {
                            const username = this.getTagValue(e.tags, 'username');
                            const name = this.getTagValue(e.tags, 'name');
                            return username === relationValue || name === relationValue;
                        });

                        if (!foundEntity) {
                            validation.errors.push(
                                `Entity ${entity.id} references non-existent ${relationTag}: ${relationValue}`
                            );
                            validation.valid = false;
                            continue;
                        }
                    }

                    // Check if referenced entity has correct type
                    const referencedType = this.getTagValue(
                        (referencedEntity || entities.find(e => 
                            this.getTagValue(e.tags, 'username') === relationValue ||
                            this.getTagValue(e.tags, 'name') === relationValue
                        )).tags, 
                        'type'
                    );

                    if (referencedType !== expectedType) {
                        validation.errors.push(
                            `Entity ${entity.id} ${relationTag} references wrong type: expected ${expectedType}, got ${referencedType}`
                        );
                        validation.valid = false;
                    }
                }
            }

            // Check for orphaned entities
            const orphans = this.findOrphanedEntities(entities, typeMap);
            for (const orphan of orphans) {
                validation.warnings.push(
                    `Orphaned ${orphan.type}: ${orphan.id} (${orphan.reason})`
                );
            }

            return validation;

        } catch (error) {
            console.error('Relationship validation error:', error);
            return {
                valid: false,
                errors: [error.message],
                warnings: []
            };
        }
    }

    findOrphanedEntities(entities, typeMap) {
        const orphans = [];

        // Check for tasks without stories or assignees
        const tasks = typeMap.get('task') || [];
        for (const task of tasks) {
            const story = this.getTagValue(task.tags, 'story');
            const assignee = this.getTagValue(task.tags, 'assignee');

            if (!story) {
                orphans.push({
                    id: task.id,
                    type: 'task',
                    reason: 'no story assigned'
                });
            }

            if (!assignee) {
                orphans.push({
                    id: task.id,
                    type: 'task',
                    reason: 'no assignee'
                });
            }
        }

        // Check for stories without epics
        const stories = typeMap.get('story') || [];
        for (const story of stories) {
            const epic = this.getTagValue(story.tags, 'epic');
            if (!epic) {
                orphans.push({
                    id: story.id,
                    type: 'story',
                    reason: 'no epic assigned'
                });
            }
        }

        // Check for epics without projects
        const epics = typeMap.get('epic') || [];
        for (const epic of epics) {
            const project = this.getTagValue(epic.tags, 'project');
            if (!project) {
                orphans.push({
                    id: epic.id,
                    type: 'epic',
                    reason: 'no project assigned'
                });
            }
        }

        // Check for projects without organizations
        const projects = typeMap.get('project') || [];
        for (const project of projects) {
            const org = this.getTagValue(project.tags, 'org');
            if (!org) {
                orphans.push({
                    id: project.id,
                    type: 'project',
                    reason: 'no organization assigned'
                });
            }
        }

        return orphans;
    }

    // Validation Helpers
    isValidStatus(status, entityType) {
        const validStatuses = {
            organization: ['active', 'inactive', 'archived'],
            project: ['planning', 'active', 'on-hold', 'completed', 'cancelled'],
            epic: ['todo', 'doing', 'review', 'done', 'cancelled'],
            story: ['todo', 'doing', 'review', 'done', 'cancelled'],
            task: ['todo', 'doing', 'review', 'done', 'cancelled'],
            user: ['active', 'inactive', 'archived'],
            sprint: ['planning', 'active', 'completed', 'cancelled'],
            workspace: ['active', 'archived', 'deleted']
        };

        return validStatuses[entityType]?.includes(status) || false;
    }

    isValidPriority(priority) {
        const validPriorities = ['low', 'medium', 'high', 'urgent'];
        return validPriorities.includes(priority);
    }

    getTagValue(tags, key) {
        if (!tags) return null;
        const tag = tags.find(t => t.startsWith(`${key}:`));
        return tag ? tag.substring(key.length + 1) : null;
    }

    // Schema Information
    getEntitySchema(type) {
        return this.schemas[type] || null;
    }

    getAvailableTypes() {
        return Object.keys(this.schemas);
    }

    getRequiredTags(type) {
        const schema = this.schemas[type];
        return schema ? schema.required : [];
    }

    getOptionalTags(type) {
        const schema = this.schemas[type];
        return schema ? schema.optional : [];
    }

    getRelationships(type) {
        const schema = this.schemas[type];
        return schema ? schema.relationships || {} : {};
    }

    // Repair Helpers
    async generateRepairPlan(validationResult) {
        const repairPlan = {
            actions: [],
            estimates: {
                automatic: 0,
                manual: 0,
                impossible: 0
            }
        };

        // Analyze validation errors and create repair actions
        for (const entityValidation of validationResult.entityValidations) {
            if (!entityValidation.valid) {
                for (const error of entityValidation.errors) {
                    const action = this.createRepairAction(entityValidation.id, error);
                    if (action) {
                        repairPlan.actions.push(action);
                        repairPlan.estimates[action.complexity]++;
                    }
                }
            }
        }

        // Add relationship repair actions
        for (const error of validationResult.relationships.errors) {
            const action = this.createRelationshipRepairAction(error);
            if (action) {
                repairPlan.actions.push(action);
                repairPlan.estimates[action.complexity]++;
            }
        }

        return repairPlan;
    }

    createRepairAction(entityId, error) {
        if (error.includes('Missing required tag')) {
            const tagName = error.split(': ')[1];
            return {
                entityId,
                type: 'add_tag',
                description: `Add missing required tag: ${tagName}`,
                complexity: 'manual',
                priority: 'high'
            };
        }

        if (error.includes('Type mismatch')) {
            return {
                entityId,
                type: 'fix_type',
                description: error,
                complexity: 'manual',
                priority: 'high'
            };
        }

        return null;
    }

    createRelationshipRepairAction(error) {
        if (error.includes('references non-existent')) {
            return {
                type: 'fix_reference',
                description: error,
                complexity: 'manual',
                priority: 'medium'
            };
        }

        return null;
    }
}

// Global instance
window.schemaValidator = new SchemaValidator();

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = SchemaValidator;
}