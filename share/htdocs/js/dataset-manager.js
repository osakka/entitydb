// EntityDB Dataset Management Component
// Comprehensive dataset creation, configuration, and isolation management

const DatasetManager = {
    name: 'DatasetManager',
    template: `
        <div class="dataset-manager">
            <!-- Header -->
            <div class="dataset-header">
                <h2>Dataset Management</h2>
                <div class="dataset-stats">
                    <div class="stat-card">
                        <i class="fas fa-layer-group"></i>
                        <div class="stat-value">{{ datasets.length }}</div>
                        <div class="stat-label">Total Datasets</div>
                    </div>
                    <div class="stat-card">
                        <i class="fas fa-database"></i>
                        <div class="stat-value">{{ totalEntities }}</div>
                        <div class="stat-label">Total Entities</div>
                    </div>
                    <div class="stat-card">
                        <i class="fas fa-users"></i>
                        <div class="stat-value">{{ totalUsers }}</div>
                        <div class="stat-label">Total Users</div>
                    </div>
                    <div class="stat-card">
                        <i class="fas fa-hdd"></i>
                        <div class="stat-value">{{ formatSize(totalSize) }}</div>
                        <div class="stat-label">Total Storage</div>
                    </div>
                </div>
            </div>

            <!-- Actions Bar -->
            <div class="actions-bar">
                <button @click="showCreateDataset = true" class="btn btn-primary">
                    <i class="fas fa-plus"></i> Create Dataset
                </button>
                <button @click="refreshDatasets" class="btn btn-secondary">
                    <i class="fas fa-sync" :class="{ 'fa-spin': loading }"></i> Refresh
                </button>
                <div class="view-toggle">
                    <button 
                        @click="viewMode = 'grid'" 
                        :class="['view-btn', { active: viewMode === 'grid' }]"
                    >
                        <i class="fas fa-th-large"></i>
                    </button>
                    <button 
                        @click="viewMode = 'list'" 
                        :class="['view-btn', { active: viewMode === 'list' }]"
                    >
                        <i class="fas fa-list"></i>
                    </button>
                </div>
            </div>

            <!-- Dataset Grid View -->
            <div v-if="viewMode === 'grid'" class="dataset-grid">
                <div 
                    v-for="ds in datasets" 
                    :key="ds.name"
                    :class="['dataset-card', { 
                        active: ds.name === currentDataset,
                        system: ds.isSystem 
                    }]"
                    @click="selectDataset(ds)"
                >
                    <div class="dataset-card-header">
                        <div class="dataset-icon">
                            <i :class="getDatasetIcon(ds)"></i>
                        </div>
                        <div class="dataset-info">
                            <h3>{{ ds.name }}</h3>
                            <p class="dataset-description">{{ ds.description }}</p>
                        </div>
                        <div v-if="ds.name === currentDataset" class="active-badge">
                            <i class="fas fa-check-circle"></i> Active
                        </div>
                    </div>

                    <div class="dataset-stats-grid">
                        <div class="stat-item">
                            <i class="fas fa-cube"></i>
                            <span>{{ ds.entityCount }} entities</span>
                        </div>
                        <div class="stat-item">
                            <i class="fas fa-users"></i>
                            <span>{{ ds.userCount }} users</span>
                        </div>
                        <div class="stat-item">
                            <i class="fas fa-hdd"></i>
                            <span>{{ formatSize(ds.size) }}</span>
                        </div>
                        <div class="stat-item">
                            <i class="fas fa-clock"></i>
                            <span>{{ formatDate(ds.created) }}</span>
                        </div>
                    </div>

                    <div class="dataset-tags">
                        <span v-for="tag in ds.tags" :key="tag" class="dataset-tag">
                            {{ tag }}
                        </span>
                    </div>

                    <div class="dataset-actions">
                        <button 
                            @click.stop="switchToDataset(ds)" 
                            v-if="ds.name !== currentDataset"
                            class="btn btn-small btn-primary"
                        >
                            <i class="fas fa-sign-in-alt"></i> Switch
                        </button>
                        <button 
                            @click.stop="configureDataset(ds)" 
                            class="btn btn-small btn-secondary"
                        >
                            <i class="fas fa-cog"></i> Configure
                        </button>
                        <button 
                            @click.stop="viewDatasetDetails(ds)" 
                            class="btn btn-small btn-secondary"
                        >
                            <i class="fas fa-info-circle"></i> Details
                        </button>
                        <button 
                            @click.stop="deleteDataset(ds)" 
                            v-if="!ds.isSystem"
                            class="btn btn-small btn-danger"
                        >
                            <i class="fas fa-trash"></i> Delete
                        </button>
                    </div>
                </div>
            </div>

            <!-- Dataset List View -->
            <div v-if="viewMode === 'list'" class="dataset-list">
                <table class="dataset-table">
                    <thead>
                        <tr>
                            <th>Name</th>
                            <th>Description</th>
                            <th>Entities</th>
                            <th>Users</th>
                            <th>Size</th>
                            <th>Created</th>
                            <th>Status</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr 
                            v-for="ds in datasets" 
                            :key="ds.name"
                            :class="{ active: ds.name === currentDataset }"
                            @click="selectDataset(ds)"
                        >
                            <td>
                                <div class="name-cell">
                                    <i :class="getDatasetIcon(ds)"></i>
                                    {{ ds.name }}
                                </div>
                            </td>
                            <td>{{ ds.description }}</td>
                            <td>{{ ds.entityCount }}</td>
                            <td>{{ ds.userCount }}</td>
                            <td>{{ formatSize(ds.size) }}</td>
                            <td>{{ formatDate(ds.created) }}</td>
                            <td>
                                <span v-if="ds.name === currentDataset" class="status-badge active">
                                    Active
                                </span>
                                <span v-else-if="ds.isSystem" class="status-badge system">
                                    System
                                </span>
                                <span v-else class="status-badge">
                                    Available
                                </span>
                            </td>
                            <td class="action-cell">
                                <button 
                                    @click.stop="switchToDataset(ds)" 
                                    v-if="ds.name !== currentDataset"
                                    class="action-btn"
                                    title="Switch to this dataset"
                                >
                                    <i class="fas fa-sign-in-alt"></i>
                                </button>
                                <button 
                                    @click.stop="configureDataset(ds)" 
                                    class="action-btn"
                                    title="Configure"
                                >
                                    <i class="fas fa-cog"></i>
                                </button>
                                <button 
                                    @click.stop="viewDatasetDetails(ds)" 
                                    class="action-btn"
                                    title="View details"
                                >
                                    <i class="fas fa-info-circle"></i>
                                </button>
                                <button 
                                    @click.stop="deleteDataset(ds)" 
                                    v-if="!ds.isSystem"
                                    class="action-btn danger"
                                    title="Delete"
                                >
                                    <i class="fas fa-trash"></i>
                                </button>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>

            <!-- Selected Dataset Details -->
            <transition name="slide">
                <div v-if="selectedDataset" class="dataset-details-panel">
                    <div class="details-header">
                        <h3>{{ selectedDataset.name }} Details</h3>
                        <button @click="selectedDataset = null" class="close-btn">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>

                    <div class="details-content">
                        <!-- Overview Tab -->
                        <div class="detail-tabs">
                            <button 
                                v-for="tab in detailTabs" 
                                :key="tab.id"
                                @click="activeDetailTab = tab.id"
                                :class="['tab-btn', { active: activeDetailTab === tab.id }]"
                            >
                                <i :class="tab.icon"></i> {{ tab.name }}
                            </button>
                        </div>

                        <div v-if="activeDetailTab === 'overview'" class="tab-content">
                            <div class="detail-section">
                                <h4>Basic Information</h4>
                                <div class="detail-field">
                                    <label>Name:</label>
                                    <span>{{ selectedDataset.name }}</span>
                                </div>
                                <div class="detail-field">
                                    <label>Description:</label>
                                    <span>{{ selectedDataset.description || 'No description' }}</span>
                                </div>
                                <div class="detail-field">
                                    <label>Created:</label>
                                    <span>{{ formatFullDate(selectedDataset.created) }}</span>
                                </div>
                                <div class="detail-field">
                                    <label>Type:</label>
                                    <span>{{ selectedDataset.isSystem ? 'System' : 'User' }}</span>
                                </div>
                            </div>

                            <div class="detail-section">
                                <h4>Statistics</h4>
                                <div class="stats-grid">
                                    <div class="stat-box">
                                        <div class="stat-number">{{ selectedDataset.entityCount }}</div>
                                        <div class="stat-label">Total Entities</div>
                                    </div>
                                    <div class="stat-box">
                                        <div class="stat-number">{{ selectedDataset.userCount }}</div>
                                        <div class="stat-label">Total Users</div>
                                    </div>
                                    <div class="stat-box">
                                        <div class="stat-number">{{ formatSize(selectedDataset.size) }}</div>
                                        <div class="stat-label">Storage Used</div>
                                    </div>
                                    <div class="stat-box">
                                        <div class="stat-number">{{ selectedDataset.relationshipCount || 0 }}</div>
                                        <div class="stat-label">Relationships</div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div v-if="activeDetailTab === 'entities'" class="tab-content">
                            <div class="entity-type-breakdown">
                                <h4>Entity Type Distribution</h4>
                                <div class="type-chart">
                                    <div 
                                        v-for="type in selectedDataset.entityTypes" 
                                        :key="type.name"
                                        class="type-bar"
                                    >
                                        <div class="type-info">
                                            <span class="type-name">{{ type.name }}</span>
                                            <span class="type-count">{{ type.count }}</span>
                                        </div>
                                        <div class="type-progress">
                                            <div 
                                                class="type-fill"
                                                :style="{ width: (type.count / selectedDataset.entityCount * 100) + '%' }"
                                            ></div>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div class="recent-entities">
                                <h4>Recent Entities</h4>
                                <div class="entity-list">
                                    <div 
                                        v-for="entity in selectedDataset.recentEntities" 
                                        :key="entity.id"
                                        class="entity-item"
                                    >
                                        <i :class="getEntityTypeIcon(entity.type)"></i>
                                        <span class="entity-id">{{ entity.id }}</span>
                                        <span class="entity-time">{{ formatRelativeTime(entity.created) }}</span>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div v-if="activeDetailTab === 'access'" class="tab-content">
                            <div class="access-control">
                                <h4>Access Control</h4>
                                <div class="access-list">
                                    <div 
                                        v-for="access in selectedDataset.accessList" 
                                        :key="access.user"
                                        class="access-item"
                                    >
                                        <div class="access-user">
                                            <i class="fas fa-user"></i>
                                            <span>{{ access.user }}</span>
                                        </div>
                                        <div class="access-permissions">
                                            <span 
                                                v-for="perm in access.permissions" 
                                                :key="perm"
                                                class="permission-badge"
                                            >
                                                {{ perm }}
                                            </span>
                                        </div>
                                        <button @click="editAccess(access)" class="edit-access-btn">
                                            <i class="fas fa-edit"></i>
                                        </button>
                                    </div>
                                </div>
                                <button @click="addAccess" class="btn btn-secondary">
                                    <i class="fas fa-user-plus"></i> Add User Access
                                </button>
                            </div>
                        </div>

                        <div v-if="activeDetailTab === 'settings'" class="tab-content">
                            <div class="dataset-settings">
                                <h4>Dataset Settings</h4>
                                <form @submit.prevent="saveDatasetSettings">
                                    <div class="form-group">
                                        <label>Description</label>
                                        <textarea 
                                            v-model="datasetSettings.description" 
                                            class="form-input"
                                            rows="3"
                                        ></textarea>
                                    </div>
                                    
                                    <div class="form-group">
                                        <label>Isolation Level</label>
                                        <select v-model="datasetSettings.isolationLevel" class="form-select">
                                            <option value="strict">Strict - Complete isolation</option>
                                            <option value="shared">Shared - Allow cross-dataset queries</option>
                                            <option value="public">Public - No isolation</option>
                                        </select>
                                    </div>

                                    <div class="form-group">
                                        <label>
                                            <input 
                                                type="checkbox" 
                                                v-model="datasetSettings.enableVersioning"
                                            >
                                            Enable Entity Versioning
                                        </label>
                                    </div>

                                    <div class="form-group">
                                        <label>
                                            <input 
                                                type="checkbox" 
                                                v-model="datasetSettings.enableAudit"
                                            >
                                            Enable Audit Logging
                                        </label>
                                    </div>

                                    <div class="form-group">
                                        <label>Storage Quota (MB)</label>
                                        <input 
                                            type="number" 
                                            v-model="datasetSettings.storageQuota"
                                            class="form-input"
                                            min="0"
                                        >
                                    </div>

                                    <button type="submit" class="btn btn-primary">
                                        <i class="fas fa-save"></i> Save Settings
                                    </button>
                                </form>
                            </div>
                        </div>
                    </div>
                </div>
            </transition>

            <!-- Create Dataset Modal -->
            <div v-if="showCreateDataset" class="modal-backdrop" @click.self="showCreateDataset = false">
                <div class="modal">
                    <div class="modal-header">
                        <h3>Create New Dataset</h3>
                        <button @click="showCreateDataset = false" class="close-btn">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>
                    <form @submit.prevent="createDataset" class="modal-body">
                        <div class="form-group">
                            <label>Dataset Name</label>
                            <input 
                                v-model="newDataset.name" 
                                type="text" 
                                required 
                                class="form-input"
                                pattern="[a-z0-9_-]+"
                                placeholder="my_dataset"
                            >
                            <small>Lowercase letters, numbers, underscores, and hyphens only</small>
                        </div>

                        <div class="form-group">
                            <label>Description</label>
                            <textarea 
                                v-model="newDataset.description" 
                                class="form-input"
                                rows="3"
                                placeholder="Describe the purpose of this dataset..."
                            ></textarea>
                        </div>

                        <div class="form-group">
                            <label>Initial Configuration</label>
                            <div class="config-options">
                                <label class="config-option">
                                    <input 
                                        type="radio" 
                                        v-model="newDataset.template" 
                                        value="blank"
                                    >
                                    <span>Blank - Empty dataset</span>
                                </label>
                                <label class="config-option">
                                    <input 
                                        type="radio" 
                                        v-model="newDataset.template" 
                                        value="basic"
                                    >
                                    <span>Basic - Include default users and config</span>
                                </label>
                                <label class="config-option">
                                    <input 
                                        type="radio" 
                                        v-model="newDataset.template" 
                                        value="clone"
                                    >
                                    <span>Clone - Copy from existing dataset</span>
                                </label>
                            </div>
                        </div>

                        <div v-if="newDataset.template === 'clone'" class="form-group">
                            <label>Source Dataset</label>
                            <select v-model="newDataset.cloneFrom" class="form-select">
                                <option value="">Select dataset to clone...</option>
                                <option 
                                    v-for="ds in datasets" 
                                    :key="ds.name"
                                    :value="ds.name"
                                >
                                    {{ ds.name }}
                                </option>
                            </select>
                        </div>

                        <div class="modal-actions">
                            <button type="button" @click="showCreateDataset = false" class="btn btn-secondary">
                                Cancel
                            </button>
                            <button type="submit" class="btn btn-primary">
                                <i class="fas fa-plus"></i> Create Dataset
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    `,
    
    props: ['sessionToken', 'currentDataset', 'isDarkMode'],
    
    data() {
        return {
            datasets: [],
            selectedDataset: null,
            loading: false,
            viewMode: 'grid',
            
            // Detail tabs
            activeDetailTab: 'overview',
            detailTabs: [
                { id: 'overview', name: 'Overview', icon: 'fas fa-info-circle' },
                { id: 'entities', name: 'Entities', icon: 'fas fa-cube' },
                { id: 'access', name: 'Access Control', icon: 'fas fa-lock' },
                { id: 'settings', name: 'Settings', icon: 'fas fa-cog' }
            ],
            
            // Create modal
            showCreateDataset: false,
            newDataset: {
                name: '',
                description: '',
                template: 'blank',
                cloneFrom: ''
            },
            
            // Settings
            datasetSettings: {
                description: '',
                isolationLevel: 'strict',
                enableVersioning: false,
                enableAudit: false,
                storageQuota: 1000
            }
        };
    },
    
    computed: {
        totalEntities() {
            return this.datasets.reduce((sum, ds) => sum + ds.entityCount, 0);
        },
        
        totalUsers() {
            return this.datasets.reduce((sum, ds) => sum + ds.userCount, 0);
        },
        
        totalSize() {
            return this.datasets.reduce((sum, ds) => sum + ds.size, 0);
        }
    },
    
    mounted() {
        this.loadDatasets();
    },
    
    methods: {
        async loadDatasets() {
            this.loading = true;
            try {
                // Mock data for now - in real implementation, call API
                this.datasets = [
                    {
                        name: 'default',
                        description: 'Default dataset for general use',
                        isSystem: false,
                        entityCount: 194,
                        userCount: 3,
                        size: 524288,
                        created: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
                        tags: ['production', 'main'],
                        entityTypes: [
                            { name: 'user', count: 3 },
                            { name: 'config', count: 15 },
                            { name: 'dashboard_layout', count: 2 },
                            { name: 'metric', count: 174 }
                        ],
                        recentEntities: [
                            { id: 'metric_123', type: 'metric', created: new Date(Date.now() - 60000) },
                            { id: 'config_456', type: 'config', created: new Date(Date.now() - 300000) }
                        ],
                        accessList: [
                            { user: 'admin', permissions: ['read', 'write', 'admin'] },
                            { user: 'test_user', permissions: ['read', 'write'] }
                        ],
                        relationshipCount: 45
                    },
                    {
                        name: '_system',
                        description: 'System dataset for internal operations',
                        isSystem: true,
                        entityCount: 89,
                        userCount: 1,
                        size: 262144,
                        created: new Date(Date.now() - 60 * 24 * 60 * 60 * 1000),
                        tags: ['system', 'protected'],
                        entityTypes: [
                            { name: 'system_config', count: 45 },
                            { name: 'audit_log', count: 44 }
                        ],
                        recentEntities: [],
                        accessList: [
                            { user: 'admin', permissions: ['read', 'admin'] }
                        ],
                        relationshipCount: 12
                    },
                    {
                        name: 'test',
                        description: 'Test environment for development',
                        isSystem: false,
                        entityCount: 45,
                        userCount: 2,
                        size: 131072,
                        created: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000),
                        tags: ['development', 'test'],
                        entityTypes: [
                            { name: 'test_entity', count: 40 },
                            { name: 'user', count: 2 },
                            { name: 'config', count: 3 }
                        ],
                        recentEntities: [],
                        accessList: [
                            { user: 'admin', permissions: ['read', 'write', 'admin'] },
                            { user: 'developer', permissions: ['read', 'write'] }
                        ],
                        relationshipCount: 8
                    }
                ];
                
                // Sort by name
                this.datasets.sort((a, b) => {
                    if (a.isSystem && !b.isSystem) return -1;
                    if (!a.isSystem && b.isSystem) return 1;
                    return a.name.localeCompare(b.name);
                });
                
            } catch (error) {
                this.$emit('error', error);
            } finally {
                this.loading = false;
            }
        },
        
        refreshDatasets() {
            this.loadDatasets();
        },
        
        selectDataset(dataset) {
            this.selectedDataset = dataset;
            this.datasetSettings = {
                description: dataset.description,
                isolationLevel: 'strict',
                enableVersioning: false,
                enableAudit: false,
                storageQuota: 1000
            };
        },
        
        async switchToDataset(dataset) {
            if (confirm(`Switch to ${dataset.name} dataset?`)) {
                this.$emit('switch-dataset', dataset.name);
                this.$emit('notification', {
                    message: `Switched to ${dataset.name} dataset`,
                    type: 'success'
                });
            }
        },
        
        configureDataset(dataset) {
            this.selectDataset(dataset);
            this.activeDetailTab = 'settings';
        },
        
        viewDatasetDetails(dataset) {
            this.selectDataset(dataset);
            this.activeDetailTab = 'overview';
        },
        
        async createDataset() {
            try {
                // Validate name
                if (!/^[a-z0-9_-]+$/.test(this.newDataset.name)) {
                    throw new Error('Invalid dataset name format');
                }
                
                // In real implementation, call API
                const response = await fetch('/api/v1/dataset/create', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${this.sessionToken}`,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        name: this.newDataset.name,
                        description: this.newDataset.description,
                        template: this.newDataset.template,
                        cloneFrom: this.newDataset.cloneFrom
                    })
                });
                
                if (response.ok) {
                    this.$emit('notification', {
                        message: `Dataset ${this.newDataset.name} created successfully`,
                        type: 'success'
                    });
                    this.showCreateDataset = false;
                    this.newDataset = {
                        name: '',
                        description: '',
                        template: 'blank',
                        cloneFrom: ''
                    };
                    await this.loadDatasets();
                } else {
                    throw new Error('Failed to create dataset');
                }
            } catch (error) {
                this.$emit('error', error);
            }
        },
        
        async deleteDataset(dataset) {
            if (dataset.isSystem) {
                this.$emit('notification', {
                    message: 'Cannot delete system dataset',
                    type: 'error'
                });
                return;
            }
            
            if (dataset.name === this.currentDataset) {
                this.$emit('notification', {
                    message: 'Cannot delete current dataset',
                    type: 'error'
                });
                return;
            }
            
            if (confirm(`Are you sure you want to delete dataset ${dataset.name}? This action cannot be undone.`)) {
                try {
                    // In real implementation, call API
                    this.$emit('notification', {
                        message: `Dataset ${dataset.name} deleted`,
                        type: 'success'
                    });
                    await this.loadDatasets();
                } catch (error) {
                    this.$emit('error', error);
                }
            }
        },
        
        async saveDatasetSettings() {
            try {
                // In real implementation, call API to update dataset settings
                this.$emit('notification', {
                    message: 'Dataset settings saved',
                    type: 'success'
                });
            } catch (error) {
                this.$emit('error', error);
            }
        },
        
        editAccess(access) {
            // Implement access editing
            this.$emit('edit-access', access);
        },
        
        addAccess() {
            // Implement adding access
            this.$emit('add-access', this.selectedDataset);
        },
        
        // Helper methods
        getDatasetIcon(dataset) {
            if (dataset.isSystem) return 'fas fa-shield-alt';
            if (dataset.name === 'default') return 'fas fa-home';
            if (dataset.tags.includes('test')) return 'fas fa-flask';
            if (dataset.tags.includes('production')) return 'fas fa-industry';
            return 'fas fa-layer-group';
        },
        
        getEntityTypeIcon(type) {
            const icons = {
                'user': 'fas fa-user',
                'config': 'fas fa-cog',
                'metric': 'fas fa-chart-line',
                'system_config': 'fas fa-tools',
                'audit_log': 'fas fa-history',
                'test_entity': 'fas fa-vial'
            };
            return icons[type] || 'fas fa-cube';
        },
        
        formatSize(bytes) {
            const units = ['B', 'KB', 'MB', 'GB'];
            let i = 0;
            while (bytes > 1024 && i < units.length - 1) {
                bytes /= 1024;
                i++;
            }
            return `${Math.round(bytes * 10) / 10} ${units[i]}`;
        },
        
        formatDate(date) {
            return new Date(date).toLocaleDateString();
        },
        
        formatFullDate(date) {
            return new Date(date).toLocaleString();
        },
        
        formatRelativeTime(date) {
            const now = new Date();
            const then = new Date(date);
            const diff = now - then;
            
            if (diff < 60000) return 'just now';
            if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
            if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
            return `${Math.floor(diff / 86400000)}d ago`;
        }
    }
};

// Export for use
if (typeof module !== 'undefined' && module.exports) {
    module.exports = DatasetManager;
}

// Also make it available globally for browser use
if (typeof window !== 'undefined') {
    window.DatasetManager = DatasetManager;
}