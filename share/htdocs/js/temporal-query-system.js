/**
 * EntityDB Temporal Query System
 * Advanced temporal navigation and time-based entity queries
 * Version: v2.30.0+
 */

class TemporalQuerySystem {
    constructor() {
        this.currentTimestamp = null;
        this.timelineData = [];
        this.timelineCache = new Map();
        this.selectedTimeRange = null;
        this.temporalFilters = [];
        this.animationSpeed = 1000; // ms
        this.maxTimelinePoints = 100;
        this.autoRefresh = false;
        this.refreshInterval = null;
        
        this.init();
    }

    init() {
        this.setupTimeNavigation();
        this.setupKeyboardShortcuts();
        this.loadTemporalPreferences();
    }

    setupTimeNavigation() {
        this.timeNavigationModes = {
            'live': {
                label: 'Live',
                icon: 'fas fa-play-circle',
                description: 'View current state in real-time'
            },
            'point': {
                label: 'Point in Time',
                icon: 'fas fa-clock',
                description: 'View entities at a specific timestamp'
            },
            'range': {
                label: 'Time Range',
                icon: 'fas fa-history',
                description: 'Compare entities across a time period'
            },
            'changes': {
                label: 'Change Tracking',
                icon: 'fas fa-code-branch',
                description: 'Track entity modifications over time'
            }
        };
    }

    setupKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            if (e.ctrlKey || e.metaKey) {
                switch(e.key) {
                    case 't':
                        e.preventDefault();
                        this.openTemporalNavigator();
                        break;
                    case 'ArrowLeft':
                        if (e.shiftKey) {
                            e.preventDefault();
                            this.navigateBackward();
                        }
                        break;
                    case 'ArrowRight':
                        if (e.shiftKey) {
                            e.preventDefault();
                            this.navigateForward();
                        }
                        break;
                }
            }
        });
    }

    // Main temporal query interface
    async performTemporalQuery(entityIds, options = {}) {
        const queryType = options.queryType || 'as-of';
        const timestamp = options.timestamp || Date.now() * 1000000; // Convert to nanoseconds
        const apiClient = window.apiClient || (window.EntityDBClient ? new window.EntityDBClient() : null);
        
        if (!apiClient) {
            throw new Error('API client not available');
        }

        try {
            const token = localStorage.getItem('entitydb-admin-token');
            if (token && apiClient.setToken) {
                apiClient.setToken(token);
            }

            let results;
            switch (queryType) {
                case 'as-of':
                    results = await this.queryAsOf(entityIds, timestamp, apiClient);
                    break;
                case 'history':
                    results = await this.queryHistory(entityIds, options, apiClient);
                    break;
                case 'changes':
                    results = await this.queryChanges(entityIds, options, apiClient);
                    break;
                case 'diff':
                    results = await this.queryDiff(entityIds, options, apiClient);
                    break;
                default:
                    throw new Error(`Unknown query type: ${queryType}`);
            }

            // Cache results
            this.cacheTemporalResults(queryType, entityIds, options, results);
            return results;
        } catch (error) {
            console.error('Temporal query failed:', error);
            throw error;
        }
    }

    async queryAsOf(entityIds, timestamp, apiClient) {
        const results = [];
        
        for (const entityId of entityIds) {
            try {
                const response = await apiClient.getEntityAsOf(entityId, timestamp);
                results.push(response);
            } catch (error) {
                console.warn(`Failed to query entity ${entityId} as-of ${timestamp}:`, error);
                results.push({ id: entityId, error: error.message });
            }
        }
        
        return results;
    }

    async queryHistory(entityIds, options, apiClient) {
        const results = [];
        const limit = options.limit || 50;
        const startTime = options.startTime;
        const endTime = options.endTime;
        
        for (const entityId of entityIds) {
            try {
                const params = {
                    id: entityId,
                    limit: limit
                };
                
                if (startTime) params.start_time = startTime;
                if (endTime) params.end_time = endTime;
                
                const response = await apiClient.getEntityHistory(entityId, params);
                results.push({
                    id: entityId,
                    history: response.history || [],
                    total_changes: response.total_changes || 0
                });
            } catch (error) {
                console.warn(`Failed to query history for entity ${entityId}:`, error);
                results.push({ id: entityId, error: error.message });
            }
        }
        
        return results;
    }

    async queryChanges(entityIds, options, apiClient) {
        const results = [];
        const startTime = options.startTime || (Date.now() - 24 * 60 * 60 * 1000) * 1000000; // Last 24h
        const endTime = options.endTime || Date.now() * 1000000;
        
        for (const entityId of entityIds) {
            try {
                const response = await apiClient.getEntityChanges({
                    id: entityId,
                    start_time: startTime,
                    end_time: endTime
                });
                results.push({
                    id: entityId,
                    changes: response.changes || []
                });
            } catch (error) {
                console.warn(`Failed to query changes for entity ${entityId}:`, error);
                results.push({ id: entityId, error: error.message });
            }
        }
        
        return results;
    }

    async queryDiff(entityIds, options, apiClient) {
        const results = [];
        const fromTime = options.fromTime;
        const toTime = options.toTime || Date.now() * 1000000;
        
        if (!fromTime) {
            throw new Error('fromTime is required for diff queries');
        }
        
        for (const entityId of entityIds) {
            try {
                const response = await apiClient.getEntityDiff(entityId, fromTime, toTime);
                results.push({
                    id: entityId,
                    diff: response
                });
            } catch (error) {
                console.warn(`Failed to query diff for entity ${entityId}:`, error);
                results.push({ id: entityId, error: error.message });
            }
        }
        
        return results;
    }

    // Timeline visualization
    openTemporalNavigator() {
        const modalId = 'temporal-navigator';
        const modalContent = this.createTemporalNavigatorContent(modalId);
        
        const modal = this.createModal(modalId, modalContent, {
            title: 'Temporal Navigator',
            size: 'extra-large'
        });

        this.showModal(modal);
        this.initializeTemporalNavigator();
    }

    createTemporalNavigatorContent(modalId) {
        return `
            <div class="temporal-navigator">
                <!-- Time Navigation Controls -->
                <div class="temporal-controls">
                    <div class="control-section">
                        <h4 class="control-title">Navigation Mode</h4>
                        <div class="navigation-modes">
                            ${Object.entries(this.timeNavigationModes).map(([key, mode]) => `
                                <button class="nav-mode-btn ${key === 'live' ? 'active' : ''}" data-mode="${key}">
                                    <i class="${mode.icon}"></i>
                                    <span class="mode-label">${mode.label}</span>
                                    <small class="mode-desc">${mode.description}</small>
                                </button>
                            `).join('')}
                        </div>
                    </div>

                    <div class="control-section">
                        <h4 class="control-title">Time Selection</h4>
                        <div class="time-controls">
                            <div class="time-input-group">
                                <label class="time-label">From:</label>
                                <input type="datetime-local" id="temporal-from" class="time-input">
                                <button class="btn btn-sm btn-secondary" onclick="temporalQuerySystem.setRelativeTime('from', '-1h')">1h ago</button>
                                <button class="btn btn-sm btn-secondary" onclick="temporalQuerySystem.setRelativeTime('from', '-1d')">1d ago</button>
                                <button class="btn btn-sm btn-secondary" onclick="temporalQuerySystem.setRelativeTime('from', '-1w')">1w ago</button>
                            </div>
                            <div class="time-input-group">
                                <label class="time-label">To:</label>
                                <input type="datetime-local" id="temporal-to" class="time-input">
                                <button class="btn btn-sm btn-secondary" onclick="temporalQuerySystem.setRelativeTime('to', 'now')">Now</button>
                            </div>
                        </div>
                    </div>

                    <div class="control-section">
                        <h4 class="control-title">Timeline Options</h4>
                        <div class="timeline-options">
                            <label class="option-item">
                                <input type="checkbox" id="show-all-entities" checked>
                                <span>Show All Entities</span>
                            </label>
                            <label class="option-item">
                                <input type="checkbox" id="show-modifications">
                                <span>Highlight Modifications</span>
                            </label>
                            <label class="option-item">
                                <input type="checkbox" id="auto-refresh">
                                <span>Auto Refresh</span>
                            </label>
                            <label class="option-item">
                                <input type="checkbox" id="smooth-animations" checked>
                                <span>Smooth Animations</span>
                            </label>
                        </div>
                    </div>
                </div>

                <!-- Timeline Visualization -->
                <div class="timeline-container">
                    <div class="timeline-header">
                        <div class="timeline-info">
                            <span id="timeline-status">Ready to explore temporal data</span>
                        </div>
                        <div class="timeline-actions">
                            <button class="btn btn-sm btn-secondary" onclick="temporalQuerySystem.zoomOut()">
                                <i class="fas fa-search-minus"></i> Zoom Out
                            </button>
                            <button class="btn btn-sm btn-secondary" onclick="temporalQuerySystem.zoomIn()">
                                <i class="fas fa-search-plus"></i> Zoom In
                            </button>
                            <button class="btn btn-sm btn-secondary" onclick="temporalQuerySystem.resetZoom()">
                                <i class="fas fa-expand-arrows-alt"></i> Reset
                            </button>
                        </div>
                    </div>
                    
                    <div class="timeline-visualization" id="timeline-viz">
                        <div class="timeline-placeholder">
                            <i class="fas fa-clock timeline-placeholder-icon"></i>
                            <p>Select entities and time range to visualize temporal data</p>
                        </div>
                    </div>
                </div>

                <!-- Entity Selection -->
                <div class="entity-selection">
                    <div class="selection-header">
                        <h4 class="selection-title">Selected Entities</h4>
                        <button class="btn btn-sm btn-primary" onclick="temporalQuerySystem.addEntityToTimeline()">
                            <i class="fas fa-plus"></i> Add Entity
                        </button>
                    </div>
                    <div class="selected-entities" id="selected-entities">
                        <div class="no-entities">
                            <span class="text-muted">No entities selected for temporal analysis</span>
                        </div>
                    </div>
                </div>

                <div class="modal-actions">
                    <button type="button" class="btn btn-secondary" onclick="temporalQuerySystem.closeTemporalNavigator('${modalId}')">
                        Close
                    </button>
                    <button type="button" class="btn btn-primary" onclick="temporalQuerySystem.applyTemporalQuery('${modalId}')">
                        <i class="fas fa-search"></i> Query Timeline
                    </button>
                </div>
            </div>
        `;
    }

    initializeTemporalNavigator() {
        // Set default time values
        const now = new Date();
        const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
        
        document.getElementById('temporal-from').value = this.formatDateTimeLocal(oneHourAgo);
        document.getElementById('temporal-to').value = this.formatDateTimeLocal(now);

        // Setup mode switching
        document.querySelectorAll('.nav-mode-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                document.querySelectorAll('.nav-mode-btn').forEach(b => b.classList.remove('active'));
                e.target.closest('.nav-mode-btn').classList.add('active');
                this.switchNavigationMode(e.target.closest('.nav-mode-btn').dataset.mode);
            });
        });

        // Setup auto-refresh toggle
        document.getElementById('auto-refresh').addEventListener('change', (e) => {
            this.toggleAutoRefresh(e.target.checked);
        });

        // Initialize with current selected entities if any
        if (window.entityBrowserEnhanced && window.entityBrowserEnhanced.selectedEntities.size > 0) {
            this.addSelectedEntitiesToTimeline();
        }
    }

    switchNavigationMode(mode) {
        this.currentNavigationMode = mode;
        
        // Update UI based on mode
        const timeControls = document.querySelector('.time-controls');
        const timelineOptions = document.querySelector('.timeline-options');
        
        switch (mode) {
            case 'live':
                timeControls.style.display = 'none';
                this.startLiveMode();
                break;
            case 'point':
                timeControls.style.display = 'block';
                document.getElementById('temporal-to').style.display = 'none';
                break;
            case 'range':
            case 'changes':
                timeControls.style.display = 'block';
                document.getElementById('temporal-to').style.display = 'block';
                break;
        }
    }

    setRelativeTime(field, relativeTime) {
        const now = new Date();
        let targetTime;
        
        switch (relativeTime) {
            case 'now':
                targetTime = now;
                break;
            case '-1h':
                targetTime = new Date(now.getTime() - 60 * 60 * 1000);
                break;
            case '-1d':
                targetTime = new Date(now.getTime() - 24 * 60 * 60 * 1000);
                break;
            case '-1w':
                targetTime = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
                break;
            default:
                targetTime = now;
        }
        
        const fieldId = field === 'from' ? 'temporal-from' : 'temporal-to';
        document.getElementById(fieldId).value = this.formatDateTimeLocal(targetTime);
    }

    formatDateTimeLocal(date) {
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');
        
        return `${year}-${month}-${day}T${hours}:${minutes}`;
    }

    addSelectedEntitiesToTimeline() {
        if (!window.entityBrowserEnhanced) return;
        
        const selectedIds = Array.from(window.entityBrowserEnhanced.selectedEntities);
        const selectedEntityObjects = window.entityBrowserEnhanced.filteredEntities.filter(entity => 
            selectedIds.includes(entity.id)
        );

        selectedEntityObjects.forEach(entity => {
            this.addEntityToTimelineAnalysis(entity);
        });
    }

    addEntityToTimelineAnalysis(entity) {
        const container = document.getElementById('selected-entities');
        const noEntitiesMsg = container.querySelector('.no-entities');
        
        if (noEntitiesMsg) {
            noEntitiesMsg.remove();
        }

        const entityItem = document.createElement('div');
        entityItem.className = 'timeline-entity-item';
        entityItem.dataset.entityId = entity.id;
        entityItem.innerHTML = `
            <div class="entity-item-header">
                <div class="entity-item-info">
                    <strong class="entity-item-title">${this.getEntityTitle(entity)}</strong>
                    <small class="entity-item-id">${entity.id}</small>
                </div>
                <div class="entity-item-actions">
                    <button class="btn btn-sm btn-ghost" onclick="temporalQuerySystem.viewEntityHistory('${entity.id}')" title="View History">
                        <i class="fas fa-history"></i>
                    </button>
                    <button class="btn btn-sm btn-ghost btn-danger" onclick="temporalQuerySystem.removeEntityFromTimeline('${entity.id}')" title="Remove">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
            </div>
            <div class="entity-timeline-preview" id="preview-${entity.id}">
                <div class="timeline-loading">
                    <i class="fas fa-spinner fa-spin"></i>
                    <span>Loading temporal data...</span>
                </div>
            </div>
        `;

        container.appendChild(entityItem);
        
        // Load preview timeline data
        this.loadEntityTimelinePreview(entity.id);
    }

    async loadEntityTimelinePreview(entityId) {
        try {
            const endTime = Date.now() * 1000000;
            const startTime = endTime - (24 * 60 * 60 * 1000 * 1000000); // Last 24 hours

            const changes = await this.performTemporalQuery([entityId], {
                queryType: 'changes',
                startTime: startTime,
                endTime: endTime
            });

            this.renderEntityTimelinePreview(entityId, changes[0]);
        } catch (error) {
            console.error('Failed to load timeline preview:', error);
            this.renderEntityTimelineError(entityId, error.message);
        }
    }

    renderEntityTimelinePreview(entityId, changeData) {
        const preview = document.getElementById(`preview-${entityId}`);
        if (!preview) return;

        if (changeData.error) {
            this.renderEntityTimelineError(entityId, changeData.error);
            return;
        }

        const changes = changeData.changes || [];
        
        if (changes.length === 0) {
            preview.innerHTML = `
                <div class="timeline-no-changes">
                    <i class="fas fa-info-circle"></i>
                    <span>No changes in the last 24 hours</span>
                </div>
            `;
            return;
        }

        // Create mini timeline
        const timelineHTML = `
            <div class="mini-timeline">
                <div class="timeline-track">
                    ${changes.slice(0, 10).map((change, index) => `
                        <div class="timeline-point" 
                             style="left: ${(index / Math.max(changes.length - 1, 1)) * 100}%"
                             title="${new Date(change.timestamp / 1000000).toLocaleString()}: ${change.type || 'Modified'}">
                        </div>
                    `).join('')}
                </div>
                <div class="timeline-labels">
                    <span class="timeline-start">24h ago</span>
                    <span class="timeline-end">Now</span>
                </div>
            </div>
            <div class="timeline-stats">
                <span class="stat-item">
                    <i class="fas fa-edit"></i>
                    ${changes.length} changes
                </span>
                <span class="stat-item">
                    <i class="fas fa-clock"></i>
                    Last: ${this.formatRelativeTime(changes[changes.length - 1]?.timestamp)}
                </span>
            </div>
        `;

        preview.innerHTML = timelineHTML;
    }

    renderEntityTimelineError(entityId, errorMessage) {
        const preview = document.getElementById(`preview-${entityId}`);
        if (!preview) return;

        preview.innerHTML = `
            <div class="timeline-error">
                <i class="fas fa-exclamation-triangle"></i>
                <span>Error: ${errorMessage}</span>
            </div>
        `;
    }

    // Timeline visualization methods
    async applyTemporalQuery(modalId) {
        const fromInput = document.getElementById('temporal-from');
        const toInput = document.getElementById('temporal-to');
        const showAllEntities = document.getElementById('show-all-entities').checked;
        
        const fromTime = fromInput.value ? new Date(fromInput.value).getTime() * 1000000 : null;
        const toTime = toInput.value ? new Date(toInput.value).getTime() * 1000000 : Date.now() * 1000000;

        const selectedEntityElements = document.querySelectorAll('.timeline-entity-item');
        const entityIds = Array.from(selectedEntityElements).map(el => el.dataset.entityId);

        if (entityIds.length === 0 && !showAllEntities) {
            this.showNotification('Please select entities or enable "Show All Entities"', 'warning');
            return;
        }

        try {
            document.getElementById('timeline-status').textContent = 'Loading temporal data...';
            
            let queryEntityIds = entityIds;
            if (showAllEntities && window.entityBrowserEnhanced) {
                queryEntityIds = window.entityBrowserEnhanced.entities.map(e => e.id);
            }

            const results = await this.performTemporalQuery(queryEntityIds, {
                queryType: this.currentNavigationMode === 'point' ? 'as-of' : 'changes',
                startTime: fromTime,
                endTime: toTime,
                timestamp: fromTime || toTime
            });

            this.renderFullTimeline(results, fromTime, toTime);
            document.getElementById('timeline-status').textContent = 
                `Showing ${results.length} entities from ${new Date(fromTime / 1000000).toLocaleString()} to ${new Date(toTime / 1000000).toLocaleString()}`;
        } catch (error) {
            console.error('Temporal query failed:', error);
            this.showNotification('Temporal query failed: ' + error.message, 'error');
            document.getElementById('timeline-status').textContent = 'Query failed';
        }
    }

    renderFullTimeline(results, fromTime, toTime) {
        const container = document.getElementById('timeline-viz');
        const timeSpan = toTime - fromTime;
        
        if (timeSpan <= 0 || results.length === 0) {
            container.innerHTML = `
                <div class="timeline-placeholder">
                    <i class="fas fa-info-circle"></i>
                    <p>No temporal data found for the selected time range</p>
                </div>
            `;
            return;
        }

        // Create timeline visualization
        const timelineHTML = `
            <div class="full-timeline">
                <div class="timeline-scale">
                    ${this.generateTimeScale(fromTime, toTime)}
                </div>
                <div class="timeline-entities">
                    ${results.map((result, index) => this.renderEntityTimeline(result, fromTime, toTime, index)).join('')}
                </div>
            </div>
        `;

        container.innerHTML = timelineHTML;
        this.setupTimelineInteractions();
    }

    generateTimeScale(fromTime, toTime) {
        const timeSpan = toTime - fromTime;
        const numTicks = 10;
        const tickInterval = timeSpan / numTicks;
        
        let ticks = [];
        for (let i = 0; i <= numTicks; i++) {
            const tickTime = fromTime + (i * tickInterval);
            const position = (i / numTicks) * 100;
            
            ticks.push(`
                <div class="time-tick" style="left: ${position}%">
                    <div class="tick-mark"></div>
                    <div class="tick-label">${this.formatTimeLabel(tickTime)}</div>
                </div>
            `);
        }
        
        return ticks.join('');
    }

    renderEntityTimeline(result, fromTime, toTime, index) {
        const entityId = result.id;
        const changes = result.changes || [];
        const timeSpan = toTime - fromTime;
        
        if (result.error) {
            return `
                <div class="entity-timeline error">
                    <div class="entity-timeline-header">
                        <span class="entity-name">${entityId}</span>
                        <span class="error-badge">Error</span>
                    </div>
                    <div class="entity-timeline-track error">
                        <span class="error-message">${result.error}</span>
                    </div>
                </div>
            `;
        }

        const timelinePoints = changes.map(change => {
            const position = ((change.timestamp - fromTime) / timeSpan) * 100;
            return `
                <div class="timeline-change-point" 
                     style="left: ${Math.max(0, Math.min(100, position))}%"
                     data-timestamp="${change.timestamp}"
                     data-type="${change.type || 'modification'}"
                     title="${new Date(change.timestamp / 1000000).toLocaleString()}: ${change.type || 'Modified'}">
                </div>
            `;
        }).join('');

        return `
            <div class="entity-timeline" data-entity-id="${entityId}">
                <div class="entity-timeline-header">
                    <span class="entity-name">${this.getEntityTitle({id: entityId})}</span>
                    <span class="change-count">${changes.length} changes</span>
                </div>
                <div class="entity-timeline-track">
                    ${timelinePoints}
                    <div class="timeline-base-line"></div>
                </div>
            </div>
        `;
    }

    setupTimelineInteractions() {
        // Add click handlers for timeline points
        document.querySelectorAll('.timeline-change-point').forEach(point => {
            point.addEventListener('click', (e) => {
                const timestamp = e.target.dataset.timestamp;
                const entityId = e.target.closest('.entity-timeline').dataset.entityId;
                this.showChangeDetails(entityId, timestamp);
            });
        });

        // Add hover effects
        document.querySelectorAll('.entity-timeline').forEach(timeline => {
            timeline.addEventListener('mouseenter', (e) => {
                e.target.classList.add('highlighted');
            });
            
            timeline.addEventListener('mouseleave', (e) => {
                e.target.classList.remove('highlighted');
            });
        });
    }

    async showChangeDetails(entityId, timestamp) {
        // Implementation for showing detailed change information
        try {
            const beforeTimestamp = parseInt(timestamp) - 1;
            const afterTimestamp = parseInt(timestamp);
            
            const diff = await this.performTemporalQuery([entityId], {
                queryType: 'diff',
                fromTime: beforeTimestamp,
                toTime: afterTimestamp
            });

            this.displayChangeDiff(entityId, timestamp, diff[0]);
        } catch (error) {
            console.error('Failed to load change details:', error);
            this.showNotification('Failed to load change details', 'error');
        }
    }

    displayChangeDiff(entityId, timestamp, diffData) {
        // Create a modal or sidebar showing the diff
        const modalId = 'change-details';
        const modalContent = `
            <div class="change-details">
                <div class="change-header">
                    <h3>Change Details</h3>
                    <div class="change-meta">
                        <span class="entity-id">Entity: ${entityId}</span>
                        <span class="timestamp">Time: ${new Date(parseInt(timestamp) / 1000000).toLocaleString()}</span>
                    </div>
                </div>
                <div class="change-content">
                    ${this.renderDiffContent(diffData)}
                </div>
            </div>
        `;

        const modal = this.createModal(modalId, modalContent, {
            title: 'Change Details',
            size: 'large'
        });

        this.showModal(modal);
    }

    renderDiffContent(diffData) {
        if (diffData.error) {
            return `<div class="diff-error">Error: ${diffData.error}</div>`;
        }

        const diff = diffData.diff || {};
        
        return `
            <div class="diff-container">
                <div class="diff-section">
                    <h4>Tags Changed</h4>
                    <div class="tag-changes">
                        ${this.renderTagChanges(diff.tags_added, diff.tags_removed)}
                    </div>
                </div>
                <div class="diff-section">
                    <h4>Content Changed</h4>
                    <div class="content-changes">
                        ${this.renderContentChanges(diff.content_before, diff.content_after)}
                    </div>
                </div>
            </div>
        `;
    }

    renderTagChanges(added = [], removed = []) {
        if (added.length === 0 && removed.length === 0) {
            return '<span class="text-muted">No tag changes</span>';
        }

        return `
            <div class="tag-changes">
                ${added.length > 0 ? `
                    <div class="tags-added">
                        <strong class="change-type added">Added:</strong>
                        ${added.map(tag => `<span class="tag-badge added">${tag}</span>`).join('')}
                    </div>
                ` : ''}
                ${removed.length > 0 ? `
                    <div class="tags-removed">
                        <strong class="change-type removed">Removed:</strong>
                        ${removed.map(tag => `<span class="tag-badge removed">${tag}</span>`).join('')}
                    </div>
                ` : ''}
            </div>
        `;
    }

    renderContentChanges(before, after) {
        if (!before && !after) {
            return '<span class="text-muted">No content changes</span>';
        }

        try {
            const beforeContent = before ? atob(before) : '';
            const afterContent = after ? atob(after) : '';
            
            if (beforeContent === afterContent) {
                return '<span class="text-muted">Content unchanged</span>';
            }

            return `
                <div class="content-diff">
                    <div class="content-before">
                        <strong>Before:</strong>
                        <pre class="content-preview">${this.escapeHtml(beforeContent.substring(0, 500))}${beforeContent.length > 500 ? '...' : ''}</pre>
                    </div>
                    <div class="content-after">
                        <strong>After:</strong>
                        <pre class="content-preview">${this.escapeHtml(afterContent.substring(0, 500))}${afterContent.length > 500 ? '...' : ''}</pre>
                    </div>
                </div>
            `;
        } catch (e) {
            return '<span class="text-muted">Binary content changed</span>';
        }
    }

    // Utility methods
    formatTimeLabel(timestamp) {
        const date = new Date(timestamp / 1000000);
        return date.toLocaleTimeString();
    }

    formatRelativeTime(timestamp) {
        if (!timestamp) return 'Unknown';
        
        const now = Date.now() * 1000000;
        const diff = now - timestamp;
        const minutes = Math.floor(diff / (60 * 1000 * 1000000));
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);
        
        if (days > 0) return `${days}d ago`;
        if (hours > 0) return `${hours}h ago`;
        if (minutes > 0) return `${minutes}m ago`;
        return 'Just now';
    }

    getEntityTitle(entity) {
        if (!entity.tags) return `Entity ${entity.id.substring(0, 8)}`;
        
        const titleTag = entity.tags?.find(t => this.stripTimestamp(t).startsWith('title:'));
        if (titleTag) {
            return this.stripTimestamp(titleTag).split(':').slice(1).join(':');
        }
        
        const nameTag = entity.tags?.find(t => this.stripTimestamp(t).startsWith('name:'));
        if (nameTag) {
            return this.stripTimestamp(nameTag).split(':').slice(1).join(':');
        }
        
        return `Entity ${entity.id.substring(0, 8)}`;
    }

    stripTimestamp(tag) {
        if (typeof tag !== 'string') return tag;
        const pipeIndex = tag.indexOf('|');
        return pipeIndex !== -1 ? tag.substring(pipeIndex + 1) : tag;
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Timeline control methods
    zoomIn() {
        // Implementation for zooming into timeline
        this.showNotification('Timeline zoom in - coming soon', 'info');
    }

    zoomOut() {
        // Implementation for zooming out of timeline
        this.showNotification('Timeline zoom out - coming soon', 'info');
    }

    resetZoom() {
        // Implementation for resetting timeline zoom
        this.showNotification('Timeline zoom reset - coming soon', 'info');
    }

    navigateBackward() {
        // Implementation for navigating backward in time
        this.showNotification('Navigate backward - coming soon', 'info');
    }

    navigateForward() {
        // Implementation for navigating forward in time
        this.showNotification('Navigate forward - coming soon', 'info');
    }

    startLiveMode() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
        }
        
        this.refreshInterval = setInterval(() => {
            if (this.currentNavigationMode === 'live') {
                this.refreshLiveData();
            }
        }, 5000); // Refresh every 5 seconds
    }

    async refreshLiveData() {
        // Implementation for live data refresh
        try {
            // Refresh current view with latest data
            const container = document.getElementById('timeline-viz');
            if (container && !container.querySelector('.timeline-placeholder')) {
                // Re-query current entities
                await this.applyTemporalQuery();
            }
        } catch (error) {
            console.error('Live refresh failed:', error);
        }
    }

    toggleAutoRefresh(enabled) {
        this.autoRefresh = enabled;
        if (enabled) {
            this.startLiveMode();
        } else if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }

    // Cache management
    cacheTemporalResults(queryType, entityIds, options, results) {
        const cacheKey = `${queryType}_${entityIds.join(',')}_${JSON.stringify(options)}`;
        this.timelineCache.set(cacheKey, {
            results: results,
            timestamp: Date.now(),
            ttl: 300000 // 5 minutes
        });

        // Clean old cache entries
        this.cleanCache();
    }

    cleanCache() {
        const now = Date.now();
        for (const [key, value] of this.timelineCache.entries()) {
            if (now - value.timestamp > value.ttl) {
                this.timelineCache.delete(key);
            }
        }
    }

    // Modal system integration
    createModal(id, content, options = {}) {
        const modal = document.createElement('div');
        modal.id = id;
        modal.className = `modal ${options.size || 'medium'}`;
        modal.innerHTML = `
            <div class="modal-backdrop" onclick="temporalQuerySystem.closeModal('${id}')"></div>
            <div class="modal-dialog">
                <div class="modal-header">
                    <h2 class="modal-title">${options.title || 'Modal'}</h2>
                    <button class="modal-close" onclick="temporalQuerySystem.closeModal('${id}')">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                <div class="modal-body">
                    ${content}
                </div>
            </div>
        `;

        const container = document.getElementById('entity-modal-container') || document.body;
        container.appendChild(modal);
        
        return modal;
    }

    showModal(modal) {
        modal.classList.add('show');
        document.body.classList.add('modal-open');
    }

    closeModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.remove('show');
            setTimeout(() => {
                modal.remove();
                document.body.classList.remove('modal-open');
            }, 300);
        }
    }

    closeTemporalNavigator(modalId) {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
        this.closeModal(modalId);
    }

    // Additional methods
    addEntityToTimeline() {
        // Implementation to add entity through search/selection
        this.showNotification('Add entity dialog - coming soon', 'info');
    }

    removeEntityFromTimeline(entityId) {
        const item = document.querySelector(`.timeline-entity-item[data-entity-id="${entityId}"]`);
        if (item) {
            item.remove();
            
            // Show "no entities" message if none left
            const container = document.getElementById('selected-entities');
            if (container.children.length === 0) {
                container.innerHTML = '<div class="no-entities"><span class="text-muted">No entities selected for temporal analysis</span></div>';
            }
        }
    }

    viewEntityHistory(entityId) {
        // Implementation for detailed entity history view
        this.showNotification(`View history for ${entityId} - coming soon`, 'info');
    }

    // Preferences management
    loadTemporalPreferences() {
        try {
            const prefs = localStorage.getItem('entitydb-temporal-preferences');
            if (prefs) {
                const parsed = JSON.parse(prefs);
                this.animationSpeed = parsed.animationSpeed || 1000;
                this.maxTimelinePoints = parsed.maxTimelinePoints || 100;
                this.autoRefresh = parsed.autoRefresh || false;
            }
        } catch (e) {
            console.warn('Failed to load temporal preferences:', e);
        }
    }

    saveTemporalPreferences() {
        const prefs = {
            animationSpeed: this.animationSpeed,
            maxTimelinePoints: this.maxTimelinePoints,
            autoRefresh: this.autoRefresh
        };
        localStorage.setItem('entitydb-temporal-preferences', JSON.stringify(prefs));
    }

    showNotification(message, type = 'info') {
        if (window.notificationSystem) {
            window.notificationSystem.show(message, type);
        } else {
            console.log(`${type}: ${message}`);
        }
    }
}

// Initialize the temporal query system
if (typeof window !== 'undefined') {
    window.TemporalQuerySystem = TemporalQuerySystem;
    
    // Wait for DOM to be ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            window.temporalQuerySystem = new TemporalQuerySystem();
        });
    } else {
        window.temporalQuerySystem = new TemporalQuerySystem();
    }
}