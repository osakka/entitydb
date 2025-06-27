/**
 * Dashboard Layout Component - Modern layout manager
 * Orchestrates all UI components with proper separation of concerns
 */
class DashboardLayout extends BaseComponent {
    get defaultOptions() {
        return {
            theme: 'dark',
            sidebarCollapsed: false,
            autoSave: true,
            refreshInterval: 30000
        };
    }

    get defaultState() {
        return {
            currentView: 'entities',
            user: null,
            authenticated: false,
            loading: false,
            sidebarCollapsed: false,
            notifications: []
        };
    }

    init() {
        this.initializeLayout();
        this.setupLayoutListeners();
        this.checkAuthentication();
        this.setupRouting();
        this.startPeriodicRefresh();
    }

    async checkAuthentication() {
        try {
            const token = localStorage.getItem('entitydb_token');
            if (!token) {
                // Wait a bit for DOM to be ready
                setTimeout(() => this.showLoginModal(), 100);
                return;
            }

            // Verify token is still valid
            const userInfo = await apiClient.whoami();
            this.setState({ 
                authenticated: true, 
                user: userInfo 
            });
            
            this.renderDashboard();
            
        } catch (error) {
            console.log('Authentication check failed:', error.message);
            setTimeout(() => this.showLoginModal(), 100);
        }
    }

    initializeLayout() {
        console.log('DashboardLayout: Initializing layout...');
        this.container.innerHTML = this.getLayoutTemplate();
        console.log('DashboardLayout: Layout template rendered');
        
        // Verify essential elements were created
        const loginModal = this.container.querySelector('.login-modal');
        const dashboardContainer = this.container.querySelector('.dashboard-container');
        console.log('DashboardLayout: Login modal found:', !!loginModal);
        console.log('DashboardLayout: Dashboard container found:', !!dashboardContainer);
    }

    getLayoutTemplate() {
        return `
            <!-- Login Modal -->
            <div class="login-modal" style="display: none;">
                <div class="login-modal-content">
                    <div class="login-header">
                        <div class="entitydb-logo">
                            <div class="logo-icon"></div>
                            <h1>EntityDB</h1>
                        </div>
                        <p>Temporal Database Command Center</p>
                    </div>
                    
                    <form class="login-form">
                        <div class="form-group">
                            <label for="username">Username</label>
                            <input type="text" id="username" name="username" required>
                        </div>
                        
                        <div class="form-group">
                            <label for="password">Password</label>
                            <input type="password" id="password" name="password" required>
                        </div>
                        
                        <button type="submit" class="login-btn">
                            <span class="btn-text">Sign In</span>
                            <span class="btn-spinner" style="display: none;"></span>
                        </button>
                        
                        <div class="login-error" style="display: none;"></div>
                    </form>
                </div>
            </div>

            <!-- Main Dashboard -->
            <div class="dashboard-container" style="display: none;">
                <!-- Header -->
                <header class="dashboard-header">
                    <div class="header-left">
                        <button class="sidebar-toggle">
                            <i class="icon-menu"></i>
                        </button>
                        
                        <div class="entitydb-brand">
                            <div class="brand-icon"></div>
                            <span class="brand-text">EntityDB</span>
                        </div>
                    </div>
                    
                    <div class="header-center">
                        <div class="global-search">
                            <input type="text" class="global-search-input" placeholder="Search entities, relationships...">
                            <button class="global-search-btn">
                                <i class="icon-search"></i>
                            </button>
                        </div>
                    </div>
                    
                    <div class="header-right">
                        <div class="header-actions">
                            <button class="notification-btn">
                                <i class="icon-bell"></i>
                                <span class="notification-badge" style="display: none;">0</span>
                            </button>
                            
                            <div class="user-menu">
                                <button class="user-menu-btn">
                                    <div class="user-avatar"></div>
                                    <span class="user-name">Loading...</span>
                                    <i class="icon-chevron-down"></i>
                                </button>
                                
                                <div class="user-menu-dropdown" style="display: none;">
                                    <a href="#" class="menu-item" data-action="profile">
                                        <i class="icon-user"></i> Profile
                                    </a>
                                    <a href="#" class="menu-item" data-action="settings">
                                        <i class="icon-settings"></i> Settings
                                    </a>
                                    <div class="menu-divider"></div>
                                    <a href="#" class="menu-item" data-action="logout">
                                        <i class="icon-logout"></i> Logout
                                    </a>
                                </div>
                            </div>
                        </div>
                    </div>
                </header>

                <!-- Sidebar -->
                <aside class="dashboard-sidebar">
                    <nav class="sidebar-nav">
                        <a href="#entities" class="nav-item active" data-view="entities">
                            <i class="icon-database"></i>
                            <span class="nav-text">Entities</span>
                        </a>
                        
                        <a href="#relationships" class="nav-item" data-view="relationships">
                            <i class="icon-network"></i>
                            <span class="nav-text">Relationships</span>
                        </a>
                        
                        <a href="#temporal" class="nav-item" data-view="temporal">
                            <i class="icon-clock"></i>
                            <span class="nav-text">Temporal</span>
                        </a>
                        
                        <a href="#search" class="nav-item" data-view="search">
                            <i class="icon-search"></i>
                            <span class="nav-text">Advanced Search</span>
                        </a>
                        
                        <a href="#analytics" class="nav-item" data-view="analytics">
                            <i class="icon-chart"></i>
                            <span class="nav-text">Analytics</span>
                        </a>
                        
                        <div class="nav-divider"></div>
                        
                        <a href="#system" class="nav-item" data-view="system">
                            <i class="icon-server"></i>
                            <span class="nav-text">System</span>
                        </a>
                        
                        <a href="#admin" class="nav-item" data-view="admin">
                            <i class="icon-shield"></i>
                            <span class="nav-text">Administration</span>
                        </a>
                    </nav>
                </aside>

                <!-- Main Content -->
                <main class="dashboard-main">
                    <div class="content-header">
                        <h1 class="content-title">Entity Explorer</h1>
                        <div class="content-actions">
                            <!-- Dynamic action buttons -->
                        </div>
                    </div>
                    
                    <div class="content-body">
                        <!-- Dynamic content area -->
                    </div>
                </main>

                <!-- Notifications -->
                <div class="notification-container">
                    <!-- Dynamic notifications -->
                </div>
            </div>
        `;
    }

    setupLayoutListeners() {
        // Login form
        this.addEventListener(this.find('.login-form'), 'submit', this.handleLogin);
        
        // Sidebar toggle
        this.addEventListener(this.find('.sidebar-toggle'), 'click', this.toggleSidebar);
        
        // Navigation
        this.addEventListener(this.find('.sidebar-nav'), 'click', this.handleNavigation);
        
        // User menu
        this.addEventListener(this.find('.user-menu-btn'), 'click', this.toggleUserMenu);
        this.addEventListener(this.find('.user-menu-dropdown'), 'click', this.handleUserMenuAction);
        
        // Global search
        this.addEventListener(this.find('.global-search-input'), 'keypress', this.handleGlobalSearch);
        this.addEventListener(this.find('.global-search-btn'), 'click', this.handleGlobalSearch);
        
        // Click outside to close dropdowns
        this.addEventListener(document, 'click', this.handleDocumentClick);
    }

    setupRouting() {
        // Simple hash-based routing
        window.addEventListener('hashchange', () => {
            const hash = window.location.hash.slice(1);
            if (hash) {
                this.navigateToView(hash);
            }
        });

        // Set initial view
        const initialHash = window.location.hash.slice(1) || 'entities';
        this.navigateToView(initialHash);
    }

    async handleLogin(event) {
        event.preventDefault();
        
        const form = event.target;
        const formData = new FormData(form);
        const username = formData.get('username');
        const password = formData.get('password');

        try {
            this.setLoginLoading(true);
            this.hideLoginError();

            await apiClient.login(username, password);
            
            // Get user info after successful login
            const userInfo = await apiClient.whoami();
            
            this.setState({ 
                authenticated: true, 
                user: userInfo 
            });
            
            this.hideLoginModal();
            this.renderDashboard();
            this.showNotification('Welcome back!', 'success');
            
        } catch (error) {
            this.showLoginError(error.message);
        } finally {
            this.setLoginLoading(false);
        }
    }

    async handleLogout() {
        try {
            await apiClient.logout();
        } catch (error) {
            console.warn('Logout API call failed:', error);
        } finally {
            this.setState({ 
                authenticated: false, 
                user: null 
            });
            this.showLoginModal();
        }
    }

    navigateToView(viewName) {
        // Update navigation state
        this.setState({ currentView: viewName });
        
        // Update active nav item
        this.findAll('.nav-item').forEach(item => {
            item.classList.toggle('active', item.dataset.view === viewName);
        });

        // Update content
        this.renderViewContent(viewName);
        
        // Update URL
        if (window.location.hash.slice(1) !== viewName) {
            window.location.hash = viewName;
        }
    }

    renderViewContent(viewName) {
        const contentBody = this.find('.content-body');
        const contentTitle = this.find('.content-title');
        
        // Check if elements exist before proceeding
        if (!contentBody || !contentTitle) {
            console.warn('Content elements not found, skipping view render');
            return;
        }
        
        // Clear existing content
        contentBody.innerHTML = '';
        
        // Destroy existing components
        if (this.currentComponent) {
            this.currentComponent.destroy();
            this.currentComponent = null;
        }

        switch (viewName) {
            case 'entities':
                contentTitle.textContent = 'Entity Explorer';
                this.renderEntityExplorer(contentBody);
                break;
                
            case 'relationships':
                contentTitle.textContent = 'Relationship Network';
                this.renderRelationshipNetwork(contentBody);
                break;
                
            case 'temporal':
                contentTitle.textContent = 'Temporal Explorer';
                this.renderTemporalExplorer(contentBody);
                break;
                
            case 'search':
                contentTitle.textContent = 'Advanced Search';
                this.renderAdvancedSearch(contentBody);
                break;
                
            case 'analytics':
                contentTitle.textContent = 'Analytics Dashboard';
                this.renderAnalytics(contentBody);
                break;
                
            case 'system':
                contentTitle.textContent = 'System Monitoring';
                this.renderSystemMonitoring(contentBody);
                break;
                
            case 'admin':
                contentTitle.textContent = 'Administration';
                this.renderAdministration(contentBody);
                break;
                
            default:
                contentTitle.textContent = 'Not Found';
                contentBody.innerHTML = '<div class="empty-state">View not found</div>';
        }
    }

    renderEntityExplorer(container) {
        this.currentComponent = new EntityExplorer(container, {
            pageSize: 50,
            allowEdit: true,
            allowDelete: true,
            showRelationships: true
        });
        
        // Listen for component events
        container.addEventListener('entity-details', (event) => {
            const { entityId } = event.detail;
            this.showEntityDetails(entityId);
        });
        
        container.addEventListener('bulk-action', (event) => {
            const { action, entityIds } = event.detail;
            this.handleBulkAction(action, entityIds);
        });
    }

    renderRelationshipNetwork(container) {
        // Create container for network component
        const networkContainer = this.createElement('div', 'relationship-network-container');
        container.appendChild(networkContainer);
        
        this.currentComponent = new RelationshipNetwork(networkContainer, {
            width: 1000,
            height: 600,
            interactive: true,
            zoom: true
        });
        
        // Listen for network events
        networkContainer.addEventListener('node-select', (event) => {
            const { node } = event.detail;
            this.showEntityDetails(node.id);
        });
        
        // Add entity selector
        const selector = this.createElement('div', 'entity-selector', `
            <div class="selector-header">
                <h3>Select Entity to Explore</h3>
                <input type="text" class="entity-search" placeholder="Search entities...">
            </div>
            <div class="entity-list">
                <!-- Populated dynamically -->
            </div>
        `);
        
        container.insertBefore(selector, networkContainer);
        this.setupEntitySelector(selector);
    }

    renderTemporalExplorer(container) {
        container.innerHTML = `
            <div class="temporal-explorer">
                <div class="temporal-controls">
                    <h3>Temporal Query Builder</h3>
                    <!-- Temporal controls will be implemented -->
                </div>
                <div class="temporal-results">
                    <!-- Results will be shown here -->
                </div>
            </div>
        `;
    }

    renderAdvancedSearch(container) {
        container.innerHTML = `
            <div class="advanced-search">
                <div class="search-builder">
                    <h3>Search Builder</h3>
                    <!-- Advanced search interface -->
                </div>
                <div class="search-results">
                    <!-- Search results -->
                </div>
            </div>
        `;
    }

    renderAnalytics(container) {
        container.innerHTML = `
            <div class="analytics-dashboard">
                <div class="analytics-charts">
                    <h3>Entity Analytics</h3>
                    <!-- Analytics charts -->
                </div>
            </div>
        `;
    }

    renderSystemMonitoring(container) {
        container.innerHTML = `
            <div class="system-monitoring">
                <div class="system-metrics">
                    <h3>System Health</h3>
                    <!-- System metrics -->
                </div>
            </div>
        `;
    }

    renderAdministration(container) {
        container.innerHTML = `
            <div class="administration">
                <div class="admin-tools">
                    <h3>Administration Tools</h3>
                    <!-- Admin interface -->
                </div>
            </div>
        `;
    }

    async setupEntitySelector(selector) {
        try {
            const entities = await apiClient.listEntities();
            const entityList = selector.querySelector('.entity-list');
            
            entityList.innerHTML = entities.slice(0, 20).map(entity => `
                <div class="entity-item" data-entity-id="${entity.id}">
                    <div class="entity-name">${this.getEntityName(entity)}</div>
                    <div class="entity-type">${this.getEntityType(entity)}</div>
                </div>
            `).join('');
            
            // Handle entity selection
            this.addEventListener(entityList, 'click', (event) => {
                const item = event.target.closest('.entity-item');
                if (item && this.currentComponent) {
                    const entityId = item.dataset.entityId;
                    this.currentComponent.loadNetwork(entityId);
                }
            });
            
        } catch (error) {
            console.error('Failed to load entities for selector:', error);
        }
    }

    // Event handlers
    handleNavigation(event) {
        if (event.target.matches('.nav-item') || event.target.closest('.nav-item')) {
            event.preventDefault();
            const navItem = event.target.closest('.nav-item');
            const viewName = navItem.dataset.view;
            this.navigateToView(viewName);
        }
    }

    toggleSidebar() {
        const collapsed = !this.state.sidebarCollapsed;
        this.setState({ sidebarCollapsed: collapsed });
        this.container.classList.toggle('sidebar-collapsed', collapsed);
    }

    toggleUserMenu() {
        const dropdown = this.find('.user-menu-dropdown');
        const isVisible = dropdown.style.display !== 'none';
        dropdown.style.display = isVisible ? 'none' : 'block';
    }

    handleUserMenuAction(event) {
        if (event.target.matches('.menu-item')) {
            event.preventDefault();
            const action = event.target.dataset.action;
            
            switch (action) {
                case 'logout':
                    this.handleLogout();
                    break;
                case 'profile':
                    this.showNotification('Profile settings coming soon!', 'info');
                    break;
                case 'settings':
                    this.showNotification('Settings coming soon!', 'info');
                    break;
            }
            
            // Hide dropdown
            this.find('.user-menu-dropdown').style.display = 'none';
        }
    }

    handleGlobalSearch(event) {
        if (event.type === 'keypress' && event.key !== 'Enter') return;
        
        const query = this.find('.global-search-input').value.trim();
        if (query) {
            this.navigateToView('search');
            // TODO: Pass search query to search component
            this.showNotification(`Searching for: ${query}`, 'info');
        }
    }

    handleDocumentClick(event) {
        // Close user menu if clicking outside
        const userMenu = this.find('.user-menu');
        if (userMenu && !userMenu.contains(event.target)) {
            this.find('.user-menu-dropdown').style.display = 'none';
        }
    }

    // UI helpers
    showLoginModal() {
        console.log('DashboardLayout: Showing login modal...');
        const loginModal = this.find('.login-modal');
        const dashboardContainer = this.find('.dashboard-container');
        const usernameInput = this.find('#username');
        
        console.log('DashboardLayout: Login modal element:', !!loginModal);
        console.log('DashboardLayout: Dashboard container element:', !!dashboardContainer);
        console.log('DashboardLayout: Username input element:', !!usernameInput);
        
        if (loginModal) {
            loginModal.style.display = 'flex';
            console.log('DashboardLayout: Login modal displayed');
        } else {
            console.error('DashboardLayout: Login modal not found!');
        }
        if (dashboardContainer) {
            dashboardContainer.style.display = 'none';
        }
        if (usernameInput) {
            usernameInput.focus();
        }
    }

    hideLoginModal() {
        const loginModal = this.find('.login-modal');
        if (loginModal) {
            loginModal.style.display = 'none';
        }
    }

    renderDashboard() {
        console.log('DashboardLayout: Rendering dashboard...');
        const dashboardContainer = this.find('.dashboard-container');
        const userName = this.find('.user-name');
        const loginModal = this.find('.login-modal');
        
        console.log('DashboardLayout: Dashboard container found:', !!dashboardContainer);
        console.log('DashboardLayout: User name element found:', !!userName);
        console.log('DashboardLayout: Login modal found:', !!loginModal);
        
        if (dashboardContainer) {
            dashboardContainer.style.display = 'grid';
            console.log('DashboardLayout: Dashboard container set to grid');
        } else {
            console.error('DashboardLayout: Dashboard container not found!');
        }
        
        if (loginModal) {
            loginModal.style.display = 'none';
            console.log('DashboardLayout: Login modal hidden');
        }
        
        // Update user info
        if (this.state.user && userName) {
            userName.textContent = this.state.user.username || 'User';
            console.log('DashboardLayout: User name updated to:', this.state.user.username);
        }
        
        // Trigger initial view rendering
        this.navigateToView('entities');
    }

    setLoginLoading(loading) {
        const btn = this.find('.login-btn');
        if (!btn) return;
        
        const text = btn.querySelector('.btn-text');
        const spinner = btn.querySelector('.btn-spinner');
        
        btn.disabled = loading;
        if (text) text.style.display = loading ? 'none' : 'inline';
        if (spinner) spinner.style.display = loading ? 'inline-block' : 'none';
    }

    showLoginError(message) {
        const errorDiv = this.find('.login-error');
        if (errorDiv) {
            errorDiv.textContent = message;
            errorDiv.style.display = 'block';
        }
    }

    hideLoginError() {
        const errorDiv = this.find('.login-error');
        if (errorDiv) {
            errorDiv.style.display = 'none';
        }
    }

    showNotification(message, type = 'info') {
        const notification = this.createElement('div', `notification notification-${type}`, `
            <div class="notification-content">
                <span class="notification-message">${message}</span>
                <button class="notification-close">&times;</button>
            </div>
        `);
        
        const container = this.find('.notification-container');
        container.appendChild(notification);
        
        // Auto remove after 5 seconds
        setTimeout(() => {
            if (notification.parentNode) {
                notification.remove();
            }
        }, 5000);
        
        // Manual close
        notification.querySelector('.notification-close').onclick = () => {
            notification.remove();
        };
    }

    showEntityDetails(entityId) {
        // TODO: Implement entity details modal/panel
        this.showNotification(`Entity details for ${entityId} - coming soon!`, 'info');
    }

    handleBulkAction(action, entityIds) {
        this.showNotification(`Bulk ${action} for ${entityIds.length} entities - coming soon!`, 'info');
    }

    getEntityName(entity) {
        const nameTag = entity.tags?.find(tag => tag.startsWith('name:'));
        return nameTag ? nameTag.split(':').slice(1).join(':') : entity.id.substring(0, 12);
    }

    getEntityType(entity) {
        const typeTag = entity.tags?.find(tag => tag.startsWith('type:'));
        return typeTag ? typeTag.split(':')[1] : 'unknown';
    }

    startPeriodicRefresh() {
        // Refresh current view periodically
        this.refreshInterval = setInterval(() => {
            if (this.currentComponent && typeof this.currentComponent.refresh === 'function') {
                this.currentComponent.refresh();
            }
        }, this.options.refreshInterval);
    }

    onStateChange(oldState, newState) {
        // Handle state changes
        if (oldState.sidebarCollapsed !== newState.sidebarCollapsed) {
            localStorage.setItem('entitydb_sidebar_collapsed', newState.sidebarCollapsed);
        }
    }

    destroy() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
        }
        
        if (this.currentComponent) {
            this.currentComponent.destroy();
        }
        
        super.destroy();
    }
}

// Export component
window.DashboardLayout = DashboardLayout;