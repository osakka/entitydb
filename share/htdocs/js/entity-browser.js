/**
 * EntityDB Enhanced Entity Browser Component
 * Simplified version for compatibility
 */

class EntityBrowser {
    constructor() {
        this.currentDataset = localStorage.getItem('entitydb.dataset') || 'default';
        this.entities = [];
        this.selectedEntities = new Set();
        this.loading = false;
        this.container = null;
    }

    async mount(container) {
        this.container = container;
        this.render();
        await this.loadEntities();
    }

    render() {
        if (!this.container) return;
        
        this.container.innerHTML = `
            <div class="entity-browser">
                <div class="toolbar">
                    <button class="btn btn-primary" onclick="entityBrowser.showCreateDialog()">
                        <i class="fas fa-plus"></i> New Entity
                    </button>
                    <button class="btn btn-secondary" onclick="entityBrowser.refresh()">
                        <i class="fas fa-sync"></i> Refresh
                    </button>
                </div>

                <div class="entity-list" id="entity-list">
                    <div class="loading-placeholder" style="display: flex; align-items: center; justify-content: center; height: 200px; color: #6c757d;">
                        <div>
                            <i class="fas fa-database" style="font-size: 48px; margin-bottom: 16px; opacity: 0.3;"></i>
                            <p>Loading entities...</p>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    async loadEntities() {
        // Try to get API client - it might be set by the main app
        const apiClient = window.apiClient || (window.EntityDBClient ? new window.EntityDBClient() : null);
        
        if (!apiClient) {
            this.showError('API client not available');
            return;
        }

        this.loading = true;
        
        try {
            // Get the token from localStorage
            const token = localStorage.getItem('entitydb-admin-token');
            if (token && apiClient.setToken) {
                apiClient.setToken(token);
            }
            
            const response = await apiClient.get('/entities/list', {
                dataset: this.currentDataset,
                limit: 100
            });
            
            this.entities = response.data || response || [];
            this.renderEntities();
        } catch (error) {
            console.error('Failed to load entities:', error);
            this.showError('Failed to load entities');
        } finally {
            this.loading = false;
        }
    }

    renderEntities() {
        const listContainer = document.getElementById('entity-list');
        if (!listContainer) return;

        if (this.entities.length === 0) {
            listContainer.innerHTML = `
                <div style="text-align: center; padding: 60px; color: #6c757d;">
                    <i class="fas fa-inbox" style="font-size: 48px; margin-bottom: 16px; opacity: 0.3;"></i>
                    <p style="font-size: 18px; margin-bottom: 24px;">No entities found</p>
                    <button class="btn btn-primary" onclick="entityBrowser.showCreateDialog()">
                        <i class="fas fa-plus"></i> Create First Entity
                    </button>
                </div>
            `;
            return;
        }

        listContainer.innerHTML = `
            <div class="entity-grid">
                ${this.entities.map(entity => `
                    <div class="entity-card" data-id="${entity.id}">
                        <div class="entity-header">
                            <h4>${this.getEntityTitle(entity)}</h4>
                            <div class="entity-actions">
                                <button class="btn btn-sm btn-secondary" onclick="entityBrowser.editEntity('${entity.id}')">
                                    <i class="fas fa-edit"></i>
                                </button>
                                <button class="btn btn-sm btn-danger" onclick="entityBrowser.deleteEntity('${entity.id}')">
                                    <i class="fas fa-trash"></i>
                                </button>
                            </div>
                        </div>
                        <div class="entity-meta">
                            <small class="text-muted">ID: ${entity.id}</small><br>
                            <small class="text-muted">Updated: ${this.formatDate(entity.updated_at)}</small>
                        </div>
                        <div class="entity-tags">
                            ${entity.tags ? entity.tags.map(tag => `
                                <span class="badge badge-secondary">${this.escapeHtml(tag)}</span>
                            `).join('') : ''}
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }

    showError(message) {
        const listContainer = document.getElementById('entity-list');
        if (!listContainer) return;

        listContainer.innerHTML = `
            <div style="text-align: center; padding: 60px; color: #dc3545;">
                <i class="fas fa-exclamation-triangle" style="font-size: 48px; margin-bottom: 16px;"></i>
                <p style="font-size: 18px;">${message}</p>
                <button class="btn btn-secondary" onclick="entityBrowser.refresh()">
                    <i class="fas fa-sync"></i> Try Again
                </button>
            </div>
        `;
    }

    showCreateDialog() {
        if (window.notificationSystem) {
            window.notificationSystem.info('Entity creation dialog coming soon');
        } else {
            alert('Entity creation dialog coming soon');
        }
    }

    editEntity(entityId) {
        if (window.notificationSystem) {
            window.notificationSystem.info(`Edit entity ${entityId} - coming soon`);
        } else {
            alert(`Edit entity ${entityId} - coming soon`);
        }
    }

    async deleteEntity(entityId) {
        if (!confirm('Delete this entity?')) return;

        const apiClient = window.apiClient || (window.EntityDBClient ? new window.EntityDBClient() : null);
        
        if (!apiClient) {
            this.showError('API client not available');
            return;
        }

        try {
            // Get the token from localStorage
            const token = localStorage.getItem('entitydb-admin-token');
            if (token && apiClient.setToken) {
                apiClient.setToken(token);
            }
            
            await apiClient.delete(`/entities/${entityId}`);
            this.entities = this.entities.filter(e => e.id !== entityId);
            this.renderEntities();
            
            if (window.notificationSystem) {
                window.notificationSystem.success('Entity deleted');
            }
        } catch (error) {
            console.error('Failed to delete entity:', error);
            if (window.notificationSystem) {
                window.notificationSystem.error('Failed to delete entity');
            }
        }
    }

    refresh() {
        this.loadEntities();
    }

    // Utility methods
    getEntityTitle(entity) {
        if (!entity.tags) return `Entity ${entity.id}`;
        
        const titleTag = entity.tags.find(t => t.startsWith('title:'));
        if (titleTag) {
            return this.escapeHtml(titleTag.split(':').slice(1).join(':'));
        }
        
        const typeTag = entity.tags.find(t => t.startsWith('type:'));
        if (typeTag) {
            const type = typeTag.split(':')[1];
            return `${type.charAt(0).toUpperCase() + type.slice(1)} ${entity.id}`;
        }
        
        return `Entity ${entity.id}`;
    }

    formatDate(dateStr) {
        if (!dateStr) return 'Unknown';
        
        const date = new Date(dateStr);
        const now = new Date();
        const diff = now - date;
        
        if (diff < 60000) return 'just now';
        if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
        if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
        
        return date.toLocaleDateString();
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Add CSS for entity browser
const entityBrowserStyles = document.createElement('style');
entityBrowserStyles.textContent = `
.entity-browser {
    padding: 20px;
}

.toolbar {
    display: flex;
    gap: 12px;
    margin-bottom: 20px;
    padding-bottom: 16px;
    border-bottom: 1px solid #e9ecef;
}

.entity-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 20px;
}

.entity-card {
    background: white;
    border: 1px solid #e9ecef;
    border-radius: 8px;
    padding: 16px;
    transition: all 0.2s;
}

.entity-card:hover {
    border-color: #3498db;
    box-shadow: 0 2px 8px rgba(52, 152, 219, 0.1);
}

body.dark-mode .entity-card {
    background: #2c3e50;
    border-color: #34495e;
    color: #e1e8ed;
}

body.dark-mode .entity-card:hover {
    border-color: #3498db;
    box-shadow: 0 2px 8px rgba(52, 152, 219, 0.2);
}

.entity-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 12px;
}

.entity-header h4 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: #2c3e50;
    flex: 1;
    margin-right: 12px;
}

body.dark-mode .entity-header h4 {
    color: #e1e8ed;
}

.entity-actions {
    display: flex;
    gap: 4px;
}

.entity-meta {
    margin-bottom: 12px;
    line-height: 1.4;
}

.entity-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
}

.badge {
    display: inline-block;
    padding: 2px 8px;
    font-size: 11px;
    font-weight: 500;
    border-radius: 12px;
}

.badge-secondary {
    background-color: #6c757d;
    color: white;
}

.btn {
    padding: 6px 12px;
    border: none;
    border-radius: 4px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    display: inline-flex;
    align-items: center;
    gap: 6px;
}

.btn-primary {
    background: #3498db;
    color: white;
}

.btn-primary:hover {
    background: #2980b9;
}

.btn-secondary {
    background: #6c757d;
    color: white;
}

.btn-secondary:hover {
    background: #5a6268;
}

.btn-danger {
    background: #dc3545;
    color: white;
}

.btn-danger:hover {
    background: #c82333;
}

.btn-sm {
    padding: 4px 8px;
    font-size: 12px;
}

.text-muted {
    color: #6c757d;
}

body.dark-mode .text-muted {
    color: #95a5a6;
}
`;
document.head.appendChild(entityBrowserStyles);

// Create global instance  
window.entityBrowser = new EntityBrowser();
window.EntityBrowser = EntityBrowser;