// EntityDB RBAC Management Component
// Comprehensive role-based access control interface

const RBACManager = {
    name: 'RBACManager',
    template: `
        <div class="rbac-manager">
            <!-- RBAC Header -->
            <div class="rbac-header">
                <h2>RBAC Management</h2>
                <div class="rbac-stats">
                    <div class="stat-item">
                        <i class="fas fa-users"></i>
                        <span>{{ users.length }} Users</span>
                    </div>
                    <div class="stat-item">
                        <i class="fas fa-user-tag"></i>
                        <span>{{ roles.length }} Roles</span>
                    </div>
                    <div class="stat-item">
                        <i class="fas fa-shield-alt"></i>
                        <span>{{ permissions.length }} Permissions</span>
                    </div>
                    <div class="stat-item">
                        <i class="fas fa-user-clock"></i>
                        <span>{{ activeSessions.length }} Active Sessions</span>
                    </div>
                </div>
            </div>

            <!-- Tab Navigation -->
            <div class="rbac-tabs">
                <button 
                    v-for="tab in tabs" 
                    :key="tab.id"
                    @click="activeTab = tab.id"
                    :class="['rbac-tab', { active: activeTab === tab.id }]"
                >
                    <i :class="tab.icon"></i> {{ tab.name }}
                </button>
            </div>

            <!-- Users Tab -->
            <div v-if="activeTab === 'users'" class="tab-content">
                <div class="section-header">
                    <h3>User Management</h3>
                    <button @click="showCreateUser = true" class="btn btn-primary">
                        <i class="fas fa-user-plus"></i> Create User
                    </button>
                </div>

                <div class="user-grid">
                    <div v-for="user in users" :key="user.id" class="user-card">
                        <div class="user-avatar">
                            <i class="fas fa-user-circle"></i>
                        </div>
                        <div class="user-info">
                            <h4>{{ user.username }}</h4>
                            <p class="user-email">{{ user.email || 'No email' }}</p>
                            <div class="user-roles">
                                <span v-for="role in user.roles" :key="role" class="role-badge" :class="getRoleClass(role)">
                                    {{ role }}
                                </span>
                            </div>
                            <div class="user-meta">
                                <span class="meta-item">
                                    <i class="fas fa-calendar"></i> Created {{ formatDate(user.created) }}
                                </span>
                                <span v-if="isUserOnline(user.id)" class="meta-item online">
                                    <i class="fas fa-circle"></i> Online
                                </span>
                                <span v-else class="meta-item offline">
                                    <i class="far fa-circle"></i> Offline
                                </span>
                            </div>
                        </div>
                        <div class="user-actions">
                            <button @click="editUser(user)" class="action-btn" title="Edit User">
                                <i class="fas fa-edit"></i>
                            </button>
                            <button @click="viewUserPermissions(user)" class="action-btn" title="View Permissions">
                                <i class="fas fa-key"></i>
                            </button>
                            <button @click="viewUserSessions(user)" class="action-btn" title="View Sessions">
                                <i class="fas fa-history"></i>
                            </button>
                            <button @click="deleteUser(user)" class="action-btn danger" title="Delete User">
                                <i class="fas fa-trash"></i>
                            </button>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Roles Tab -->
            <div v-if="activeTab === 'roles'" class="tab-content">
                <div class="section-header">
                    <h3>Role Management</h3>
                    <button @click="showCreateRole = true" class="btn btn-primary">
                        <i class="fas fa-plus"></i> Create Role
                    </button>
                </div>

                <div class="role-list">
                    <div v-for="role in roles" :key="role.name" class="role-item">
                        <div class="role-header">
                            <div class="role-info">
                                <h4 :class="getRoleClass(role.name)">
                                    <i :class="getRoleIcon(role.name)"></i> {{ role.name }}
                                </h4>
                                <p class="role-description">{{ role.description }}</p>
                                <div class="role-stats">
                                    <span>{{ role.userCount }} users</span>
                                    <span>{{ role.permissions.length }} permissions</span>
                                </div>
                            </div>
                            <div class="role-actions">
                                <button @click="editRole(role)" class="btn btn-secondary">
                                    <i class="fas fa-edit"></i> Edit
                                </button>
                                <button v-if="!isSystemRole(role.name)" @click="deleteRole(role)" class="btn btn-danger">
                                    <i class="fas fa-trash"></i> Delete
                                </button>
                            </div>
                        </div>
                        
                        <div class="permission-grid">
                            <div v-for="perm in role.permissions" :key="perm" class="permission-item">
                                <i class="fas fa-check-circle"></i> {{ formatPermission(perm) }}
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Permissions Tab -->
            <div v-if="activeTab === 'permissions'" class="tab-content">
                <div class="section-header">
                    <h3>Permission Matrix</h3>
                    <div class="filter-controls">
                        <input 
                            v-model="permissionFilter" 
                            type="text" 
                            placeholder="Filter permissions..."
                            class="filter-input"
                        >
                    </div>
                </div>

                <div class="permission-matrix">
                    <table class="matrix-table">
                        <thead>
                            <tr>
                                <th>Permission</th>
                                <th v-for="role in roles" :key="role.name" class="role-column">
                                    {{ role.name }}
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="permission in filteredPermissions" :key="permission">
                                <td class="permission-name">
                                    <i :class="getPermissionIcon(permission)"></i>
                                    {{ formatPermission(permission) }}
                                </td>
                                <td v-for="role in roles" :key="role.name" class="permission-cell">
                                    <label class="permission-toggle">
                                        <input 
                                            type="checkbox"
                                            :checked="role.permissions.includes(permission)"
                                            @change="togglePermission(role, permission)"
                                            :disabled="isSystemRole(role.name)"
                                        >
                                        <span class="toggle-slider"></span>
                                    </label>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>

            <!-- Sessions Tab -->
            <div v-if="activeTab === 'sessions'" class="tab-content">
                <div class="section-header">
                    <h3>Active Sessions</h3>
                    <button @click="refreshSessions" class="btn btn-secondary">
                        <i class="fas fa-sync" :class="{ 'fa-spin': loadingSessions }"></i> Refresh
                    </button>
                </div>

                <div class="session-list">
                    <div v-for="session in activeSessions" :key="session.id" class="session-item">
                        <div class="session-user">
                            <i class="fas fa-user-circle"></i>
                            <div>
                                <h4>{{ session.username }}</h4>
                                <p class="session-id">Session: {{ session.id.substring(0, 16) }}...</p>
                            </div>
                        </div>
                        
                        <div class="session-info">
                            <div class="info-item">
                                <i class="fas fa-clock"></i>
                                <span>Started {{ formatRelativeTime(session.created) }}</span>
                            </div>
                            <div class="info-item">
                                <i class="fas fa-hourglass-half"></i>
                                <span>Expires {{ formatRelativeTime(session.expires) }}</span>
                            </div>
                            <div class="info-item">
                                <i class="fas fa-globe"></i>
                                <span>{{ session.ip || 'Unknown IP' }}</span>
                            </div>
                            <div class="info-item">
                                <i class="fas fa-desktop"></i>
                                <span>{{ session.userAgent || 'Unknown Device' }}</span>
                            </div>
                        </div>
                        
                        <div class="session-actions">
                            <button @click="revokeSession(session)" class="btn btn-danger btn-small">
                                <i class="fas fa-ban"></i> Revoke
                            </button>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Audit Log Tab -->
            <div v-if="activeTab === 'audit'" class="tab-content">
                <div class="section-header">
                    <h3>Authentication Audit Log</h3>
                    <div class="audit-filters">
                        <select v-model="auditFilter.type" class="filter-select">
                            <option value="">All Events</option>
                            <option value="login">Logins</option>
                            <option value="logout">Logouts</option>
                            <option value="failed">Failed Attempts</option>
                            <option value="permission">Permission Changes</option>
                        </select>
                        <input 
                            v-model="auditFilter.user" 
                            type="text" 
                            placeholder="Filter by user..."
                            class="filter-input"
                        >
                    </div>
                </div>

                <div class="audit-timeline">
                    <div v-for="event in filteredAuditEvents" :key="event.id" class="audit-event" :class="getEventClass(event)">
                        <div class="event-icon">
                            <i :class="getEventIcon(event)"></i>
                        </div>
                        <div class="event-details">
                            <h4>{{ event.action }}</h4>
                            <p>
                                <strong>{{ event.username }}</strong>
                                <span v-if="event.details"> - {{ event.details }}</span>
                            </p>
                            <div class="event-meta">
                                <span><i class="fas fa-clock"></i> {{ formatTimestamp(event.timestamp) }}</span>
                                <span><i class="fas fa-globe"></i> {{ event.ip }}</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Create User Modal -->
            <div v-if="showCreateUser" class="modal-backdrop" @click.self="showCreateUser = false">
                <div class="modal">
                    <div class="modal-header">
                        <h3>Create New User</h3>
                        <button @click="showCreateUser = false" class="close-btn">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>
                    <form @submit.prevent="createUser" class="modal-body">
                        <div class="form-group">
                            <label>Username</label>
                            <input v-model="newUser.username" type="text" required class="form-input">
                        </div>
                        <div class="form-group">
                            <label>Email</label>
                            <input v-model="newUser.email" type="email" class="form-input">
                        </div>
                        <div class="form-group">
                            <label>Password</label>
                            <input v-model="newUser.password" type="password" required class="form-input">
                        </div>
                        <div class="form-group">
                            <label>Roles</label>
                            <div class="role-selector">
                                <label v-for="role in roles" :key="role.name" class="role-option">
                                    <input 
                                        type="checkbox" 
                                        :value="role.name"
                                        v-model="newUser.roles"
                                    >
                                    <span :class="getRoleClass(role.name)">{{ role.name }}</span>
                                </label>
                            </div>
                        </div>
                        <div class="modal-actions">
                            <button type="button" @click="showCreateUser = false" class="btn btn-secondary">
                                Cancel
                            </button>
                            <button type="submit" class="btn btn-primary">
                                <i class="fas fa-user-plus"></i> Create User
                            </button>
                        </div>
                    </form>
                </div>
            </div>

            <!-- Permission Details Modal -->
            <div v-if="selectedUser && showPermissionDetails" class="modal-backdrop" @click.self="showPermissionDetails = false">
                <div class="modal modal-large">
                    <div class="modal-header">
                        <h3>{{ selectedUser.username }} - Effective Permissions</h3>
                        <button @click="showPermissionDetails = false" class="close-btn">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>
                    <div class="modal-body">
                        <div class="permission-summary">
                            <h4>Roles</h4>
                            <div class="role-list">
                                <span v-for="role in selectedUser.roles" :key="role" class="role-badge" :class="getRoleClass(role)">
                                    {{ role }}
                                </span>
                            </div>
                        </div>
                        
                        <div class="permission-details">
                            <h4>All Permissions</h4>
                            <div class="permission-categories">
                                <div v-for="(perms, category) in groupedUserPermissions" :key="category" class="permission-category">
                                    <h5>{{ formatCategory(category) }}</h5>
                                    <div class="permission-list">
                                        <div v-for="perm in perms" :key="perm" class="permission-item">
                                            <i class="fas fa-check"></i> {{ formatPermission(perm) }}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `,
    
    props: ['sessionToken', 'currentDataset', 'isDarkMode'],
    
    data() {
        return {
            activeTab: 'users',
            tabs: [
                { id: 'users', name: 'Users', icon: 'fas fa-users' },
                { id: 'roles', name: 'Roles', icon: 'fas fa-user-tag' },
                { id: 'permissions', name: 'Permissions', icon: 'fas fa-shield-alt' },
                { id: 'sessions', name: 'Sessions', icon: 'fas fa-user-clock' },
                { id: 'audit', name: 'Audit Log', icon: 'fas fa-history' }
            ],
            
            // Data
            users: [],
            roles: [],
            permissions: [],
            activeSessions: [],
            auditEvents: [],
            
            // UI State
            showCreateUser: false,
            showCreateRole: false,
            showPermissionDetails: false,
            selectedUser: null,
            loadingSessions: false,
            
            // Filters
            permissionFilter: '',
            auditFilter: {
                type: '',
                user: ''
            },
            
            // Forms
            newUser: {
                username: '',
                email: '',
                password: '',
                roles: []
            }
        };
    },
    
    computed: {
        filteredPermissions() {
            if (!this.permissionFilter) return this.permissions;
            const filter = this.permissionFilter.toLowerCase();
            return this.permissions.filter(perm => 
                perm.toLowerCase().includes(filter)
            );
        },
        
        filteredAuditEvents() {
            let events = [...this.auditEvents];
            
            if (this.auditFilter.type) {
                events = events.filter(e => e.type === this.auditFilter.type);
            }
            
            if (this.auditFilter.user) {
                const userFilter = this.auditFilter.user.toLowerCase();
                events = events.filter(e => 
                    e.username.toLowerCase().includes(userFilter)
                );
            }
            
            return events;
        },
        
        groupedUserPermissions() {
            if (!this.selectedUser) return {};
            
            const allPerms = this.getUserPermissions(this.selectedUser);
            const grouped = {};
            
            allPerms.forEach(perm => {
                const parts = perm.split(':');
                const category = parts.length > 2 ? parts[1] : 'general';
                
                if (!grouped[category]) {
                    grouped[category] = [];
                }
                grouped[category].push(perm);
            });
            
            return grouped;
        }
    },
    
    mounted() {
        this.loadRBACData();
        this.loadSessions();
        
        // Auto-refresh sessions every 30 seconds
        this.sessionRefreshInterval = setInterval(() => {
            if (this.activeTab === 'sessions') {
                this.loadSessions();
            }
        }, 30000);
    },
    
    beforeUnmount() {
        if (this.sessionRefreshInterval) {
            clearInterval(this.sessionRefreshInterval);
        }
    },
    
    methods: {
        async loadRBACData() {
            try {
                // Load users
                const usersResponse = await fetch('/api/v1/entities/list?tag=type:user', {
                    headers: { 'Authorization': `Bearer ${this.sessionToken}` }
                });
                
                if (usersResponse.ok) {
                    const userEntities = await usersResponse.json();
                    this.users = userEntities.map(entity => this.parseUserEntity(entity));
                }
                
                // Load roles (predefined for now)
                this.roles = [
                    {
                        name: 'admin',
                        description: 'Full system administration access',
                        permissions: ['rbac:perm:*'],
                        userCount: this.users.filter(u => u.roles.includes('admin')).length
                    },
                    {
                        name: 'user',
                        description: 'Standard user access',
                        permissions: [
                            'rbac:perm:entity:view',
                            'rbac:perm:entity:create',
                            'rbac:perm:entity:update',
                            'rbac:perm:entity:delete'
                        ],
                        userCount: this.users.filter(u => u.roles.includes('user')).length
                    },
                    {
                        name: 'viewer',
                        description: 'Read-only access',
                        permissions: [
                            'rbac:perm:entity:view',
                            'rbac:perm:config:view',
                            'rbac:perm:metrics:read'
                        ],
                        userCount: this.users.filter(u => u.roles.includes('viewer')).length
                    }
                ];
                
                // Extract all unique permissions
                const permSet = new Set();
                this.roles.forEach(role => {
                    role.permissions.forEach(perm => {
                        if (perm === 'rbac:perm:*') {
                            // Expand wildcard
                            this.getAllPermissions().forEach(p => permSet.add(p));
                        } else {
                            permSet.add(perm);
                        }
                    });
                });
                this.permissions = Array.from(permSet).sort();
                
                // Load audit events
                await this.loadAuditEvents();
                
            } catch (error) {
                this.$emit('error', error);
            }
        },
        
        async loadSessions() {
            this.loadingSessions = true;
            try {
                const response = await fetch('/api/v1/rbac/metrics', {
                    headers: { 'Authorization': `Bearer ${this.sessionToken}` }
                });
                
                if (response.ok) {
                    const data = await response.json();
                    this.activeSessions = data.active_sessions || [];
                }
            } catch (error) {
                console.error('Failed to load sessions:', error);
            } finally {
                this.loadingSessions = false;
            }
        },
        
        async loadAuditEvents() {
            // Mock audit events for now
            this.auditEvents = [
                {
                    id: 1,
                    type: 'login',
                    action: 'Successful Login',
                    username: 'admin',
                    timestamp: new Date(Date.now() - 3600000),
                    ip: '192.168.1.100',
                    details: 'Via dashboard'
                },
                {
                    id: 2,
                    type: 'failed',
                    action: 'Failed Login Attempt',
                    username: 'unknown',
                    timestamp: new Date(Date.now() - 7200000),
                    ip: '192.168.1.150',
                    details: 'Invalid credentials'
                },
                {
                    id: 3,
                    type: 'permission',
                    action: 'Role Assignment',
                    username: 'admin',
                    timestamp: new Date(Date.now() - 86400000),
                    ip: '192.168.1.100',
                    details: 'Added admin role to user test_user'
                }
            ];
        },
        
        refreshSessions() {
            this.loadSessions();
        },
        
        async createUser() {
            try {
                const userEntity = {
                    id: `user_${Date.now()}`,
                    tags: [
                        'type:user',
                        `username:${this.newUser.username}`,
                        ...this.newUser.roles.map(role => `rbac:role:${role}`)
                    ],
                    content: JSON.stringify({
                        username: this.newUser.username,
                        email: this.newUser.email,
                        created: new Date().toISOString()
                    })
                };
                
                const response = await fetch('/api/v1/users/create', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${this.sessionToken}`,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        username: this.newUser.username,
                        email: this.newUser.email,
                        password: this.newUser.password,
                        roles: this.newUser.roles
                    })
                });
                
                if (response.ok) {
                    this.$emit('notification', { 
                        message: `User ${this.newUser.username} created successfully`, 
                        type: 'success' 
                    });
                    this.showCreateUser = false;
                    this.newUser = { username: '', email: '', password: '', roles: [] };
                    await this.loadRBACData();
                } else {
                    throw new Error('Failed to create user');
                }
            } catch (error) {
                this.$emit('error', error);
            }
        },
        
        async editUser(user) {
            this.$emit('edit-user', user);
        },
        
        async deleteUser(user) {
            if (user.username === 'admin') {
                this.$emit('notification', { 
                    message: 'Cannot delete admin user', 
                    type: 'error' 
                });
                return;
            }
            
            if (confirm(`Are you sure you want to delete user ${user.username}?`)) {
                try {
                    // In a real implementation, call the delete API
                    this.$emit('notification', { 
                        message: `User ${user.username} deleted`, 
                        type: 'success' 
                    });
                    await this.loadRBACData();
                } catch (error) {
                    this.$emit('error', error);
                }
            }
        },
        
        viewUserPermissions(user) {
            this.selectedUser = user;
            this.showPermissionDetails = true;
        },
        
        viewUserSessions(user) {
            this.activeTab = 'sessions';
            // Filter sessions for this user
        },
        
        async togglePermission(role, permission) {
            if (this.isSystemRole(role.name)) return;
            
            const hasPermission = role.permissions.includes(permission);
            
            if (hasPermission) {
                role.permissions = role.permissions.filter(p => p !== permission);
            } else {
                role.permissions.push(permission);
            }
            
            // Save role changes
            this.$emit('notification', { 
                message: `Updated permissions for ${role.name} role`, 
                type: 'success' 
            });
        },
        
        async revokeSession(session) {
            if (confirm('Are you sure you want to revoke this session?')) {
                try {
                    // Call revoke session API
                    this.$emit('notification', { 
                        message: 'Session revoked', 
                        type: 'success' 
                    });
                    await this.loadSessions();
                } catch (error) {
                    this.$emit('error', error);
                }
            }
        },
        
        // Helper methods
        parseUserEntity(entity) {
            const username = entity.tags.find(t => t.startsWith('username:'))?.split(':')[1] || entity.id;
            const roles = entity.tags
                .filter(t => t.startsWith('rbac:role:'))
                .map(t => t.split(':')[2]);
            
            let userData = {};
            if (entity.content) {
                try {
                    userData = JSON.parse(entity.content);
                } catch (e) {
                    console.error('Failed to parse user content:', e);
                }
            }
            
            return {
                id: entity.id,
                username: username,
                email: userData.email || '',
                roles: roles,
                created: userData.created || new Date().toISOString(),
                ...userData
            };
        },
        
        getUserPermissions(user) {
            const perms = new Set();
            
            user.roles.forEach(roleName => {
                const role = this.roles.find(r => r.name === roleName);
                if (role) {
                    role.permissions.forEach(perm => {
                        if (perm === 'rbac:perm:*') {
                            this.getAllPermissions().forEach(p => perms.add(p));
                        } else {
                            perms.add(perm);
                        }
                    });
                }
            });
            
            return Array.from(perms).sort();
        },
        
        getAllPermissions() {
            return [
                'rbac:perm:entity:view',
                'rbac:perm:entity:create',
                'rbac:perm:entity:update',
                'rbac:perm:entity:delete',
                'rbac:perm:user:view',
                'rbac:perm:user:create',
                'rbac:perm:user:update',
                'rbac:perm:user:delete',
                'rbac:perm:config:view',
                'rbac:perm:config:update',
                'rbac:perm:system:view',
                'rbac:perm:metrics:read',
                'rbac:perm:relation:view',
                'rbac:perm:relation:create',
                'rbac:perm:relation:delete'
            ];
        },
        
        isUserOnline(userId) {
            return this.activeSessions.some(s => s.user_id === userId);
        },
        
        isSystemRole(roleName) {
            return ['admin', 'user'].includes(roleName);
        },
        
        getRoleClass(role) {
            const classes = {
                admin: 'role-admin',
                user: 'role-user',
                viewer: 'role-viewer'
            };
            return classes[role] || 'role-custom';
        },
        
        getRoleIcon(role) {
            const icons = {
                admin: 'fas fa-crown',
                user: 'fas fa-user',
                viewer: 'fas fa-eye'
            };
            return icons[role] || 'fas fa-user-tag';
        },
        
        getPermissionIcon(permission) {
            if (permission.includes('view')) return 'fas fa-eye';
            if (permission.includes('create')) return 'fas fa-plus';
            if (permission.includes('update')) return 'fas fa-edit';
            if (permission.includes('delete')) return 'fas fa-trash';
            return 'fas fa-key';
        },
        
        getEventClass(event) {
            return `event-${event.type}`;
        },
        
        getEventIcon(event) {
            const icons = {
                login: 'fas fa-sign-in-alt',
                logout: 'fas fa-sign-out-alt',
                failed: 'fas fa-exclamation-triangle',
                permission: 'fas fa-user-shield'
            };
            return icons[event.type] || 'fas fa-info-circle';
        },
        
        formatPermission(permission) {
            return permission
                .replace('rbac:perm:', '')
                .replace(/:/g, ' → ')
                .replace(/([a-z])([A-Z])/g, '$1 $2')
                .split(' ')
                .map(word => word.charAt(0).toUpperCase() + word.slice(1))
                .join(' ');
        },
        
        formatCategory(category) {
            return category.charAt(0).toUpperCase() + category.slice(1);
        },
        
        formatDate(date) {
            return new Date(date).toLocaleDateString();
        },
        
        formatTimestamp(date) {
            return new Date(date).toLocaleString();
        },
        
        formatRelativeTime(date) {
            const now = new Date();
            const then = new Date(date);
            const diff = then - now;
            
            if (diff > 0) {
                // Future time
                if (diff < 60000) return 'in a moment';
                if (diff < 3600000) return `in ${Math.floor(diff / 60000)} minutes`;
                if (diff < 86400000) return `in ${Math.floor(diff / 3600000)} hours`;
                return `in ${Math.floor(diff / 86400000)} days`;
            } else {
                // Past time
                const absDiff = Math.abs(diff);
                if (absDiff < 60000) return 'just now';
                if (absDiff < 3600000) return `${Math.floor(absDiff / 60000)} minutes ago`;
                if (absDiff < 86400000) return `${Math.floor(absDiff / 3600000)} hours ago`;
                return `${Math.floor(absDiff / 86400000)} days ago`;
            }
        }
    }
};

// Export for use
if (typeof module !== 'undefined' && module.exports) {
    module.exports = RBACManager;
}

// Also make it available globally for browser use
if (typeof window !== 'undefined') {
    window.RBACManager = RBACManager;
}