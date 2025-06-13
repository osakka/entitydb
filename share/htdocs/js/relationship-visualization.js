/**
 * EntityDB Relationship Visualization System
 * Interactive graph visualization for entity relationships
 * Version: v2.30.0+
 */

class RelationshipVisualization {
    constructor() {
        this.nodes = [];
        this.edges = [];
        this.selectedNodes = new Set();
        this.selectedEdges = new Set();
        this.simulation = null;
        this.svg = null;
        this.container = null;
        this.zoom = null;
        this.currentFilter = null;
        this.layoutMode = 'force';
        this.colorScheme = 'type';
        this.showLabels = true;
        this.showMetrics = true;
        this.animationEnabled = true;
        
        // Graph settings
        this.settings = {
            nodeSize: { min: 8, max: 30, default: 12 },
            linkDistance: 80,
            linkStrength: 0.3,
            chargeStrength: -300,
            centerForce: 0.1,
            collisionRadius: 15,
            maxNodes: 1000,
            maxEdges: 2000
        };

        this.init();
    }

    init() {
        this.setupColorSchemes();
        this.setupLayoutModes();
        this.loadVisualizationPreferences();
    }

    setupColorSchemes() {
        this.colorSchemes = {
            'type': {
                name: 'Entity Type',
                colors: {
                    'user': '#3498db',
                    'document': '#2ecc71', 
                    'task': '#e74c3c',
                    'note': '#f39c12',
                    'file': '#9b59b6',
                    'custom': '#34495e',
                    'default': '#95a5a6'
                }
            },
            'status': {
                name: 'Status',
                colors: {
                    'active': '#27ae60',
                    'inactive': '#e74c3c',
                    'pending': '#f39c12',
                    'draft': '#95a5a6',
                    'archived': '#7f8c8d',
                    'default': '#34495e'
                }
            },
            'degree': {
                name: 'Connection Degree',
                colors: {
                    'high': '#e74c3c',     // >10 connections
                    'medium': '#f39c12',   // 5-10 connections
                    'low': '#3498db',      // 1-4 connections
                    'isolated': '#95a5a6', // 0 connections
                    'default': '#34495e'
                }
            },
            'creation': {
                name: 'Creation Time',
                colors: {
                    'recent': '#2ecc71',    // Last 24h
                    'week': '#3498db',      // Last week
                    'month': '#f39c12',     // Last month
                    'older': '#95a5a6',     // Older
                    'default': '#34495e'
                }
            }
        };
    }

    setupLayoutModes() {
        this.layoutModes = {
            'force': {
                name: 'Force-Directed',
                description: 'Physics-based layout with natural clustering'
            },
            'hierarchy': {
                name: 'Hierarchical',
                description: 'Tree-like structure based on relationships'
            },
            'circular': {
                name: 'Circular',
                description: 'Nodes arranged in concentric circles'
            },
            'grid': {
                name: 'Grid',
                description: 'Organized grid layout'
            }
        };
    }

    // Main visualization interface
    openRelationshipVisualizer(entities = []) {
        const modalId = 'relationship-visualizer';
        const modalContent = this.createVisualizerContent();
        
        const modal = this.createModal(modalId, modalContent, {
            title: 'Relationship Visualization',
            size: 'extra-large'
        });

        this.showModal(modal);
        this.initializeVisualizer(entities);
    }

    createVisualizerContent() {
        return `
            <div class="relationship-visualizer">
                <!-- Control Panel -->
                <div class="visualization-controls">
                    <div class="control-group">
                        <h4 class="control-title">Visualization</h4>
                        <div class="control-buttons">
                            <button class="btn btn-sm btn-secondary" onclick="relationshipVisualization.resetZoom()">
                                <i class="fas fa-expand-arrows-alt"></i> Reset View
                            </button>
                            <button class="btn btn-sm btn-secondary" onclick="relationshipVisualization.centerGraph()">
                                <i class="fas fa-crosshairs"></i> Center
                            </button>
                            <button class="btn btn-sm btn-secondary" onclick="relationshipVisualization.fitToScreen()">
                                <i class="fas fa-compress-arrows-alt"></i> Fit Screen
                            </button>
                        </div>
                    </div>

                    <div class="control-group">
                        <h4 class="control-title">Layout</h4>
                        <div class="layout-selector">
                            ${Object.entries(this.layoutModes).map(([key, mode]) => `
                                <button class="layout-btn ${key === 'force' ? 'active' : ''}" 
                                        data-layout="${key}"
                                        title="${mode.description}">
                                    ${mode.name}
                                </button>
                            `).join('')}
                        </div>
                    </div>

                    <div class="control-group">
                        <h4 class="control-title">Colors</h4>
                        <select id="color-scheme" class="form-input control-select">
                            ${Object.entries(this.colorSchemes).map(([key, scheme]) => `
                                <option value="${key}" ${key === 'type' ? 'selected' : ''}>${scheme.name}</option>
                            `).join('')}
                        </select>
                    </div>

                    <div class="control-group">
                        <h4 class="control-title">Display</h4>
                        <div class="display-options">
                            <label class="control-option">
                                <input type="checkbox" id="show-labels" checked>
                                <span>Show Labels</span>
                            </label>
                            <label class="control-option">
                                <input type="checkbox" id="show-metrics" checked>
                                <span>Node Metrics</span>
                            </label>
                            <label class="control-option">
                                <input type="checkbox" id="enable-animation" checked>
                                <span>Animations</span>
                            </label>
                        </div>
                    </div>

                    <div class="control-group">
                        <h4 class="control-title">Filters</h4>
                        <div class="filter-controls">
                            <input type="text" id="node-filter" class="form-input" placeholder="Filter nodes...">
                            <select id="relationship-filter" class="form-input">
                                <option value="">All Relationships</option>
                                <option value="parent">Parent-Child</option>
                                <option value="related">Related</option>
                                <option value="reference">References</option>
                                <option value="custom">Custom</option>
                            </select>
                        </div>
                    </div>
                </div>

                <!-- Visualization Area -->
                <div class="visualization-area">
                    <div class="graph-container" id="graph-container">
                        <div class="graph-placeholder">
                            <i class="fas fa-project-diagram placeholder-icon"></i>
                            <p>Loading relationship graph...</p>
                        </div>
                    </div>

                    <!-- Graph Info Panel -->
                    <div class="graph-info" id="graph-info">
                        <div class="info-section">
                            <h4>Graph Statistics</h4>
                            <div class="stats-grid">
                                <div class="stat-item">
                                    <span class="stat-label">Nodes:</span>
                                    <span class="stat-value" id="node-count">0</span>
                                </div>
                                <div class="stat-item">
                                    <span class="stat-label">Edges:</span>
                                    <span class="stat-value" id="edge-count">0</span>
                                </div>
                                <div class="stat-item">
                                    <span class="stat-label">Density:</span>
                                    <span class="stat-value" id="graph-density">0%</span>
                                </div>
                                <div class="stat-item">
                                    <span class="stat-label">Components:</span>
                                    <span class="stat-value" id="component-count">0</span>
                                </div>
                            </div>
                        </div>

                        <div class="info-section">
                            <h4>Selection Info</h4>
                            <div id="selection-info">
                                <p class="text-muted">No nodes selected</p>
                            </div>
                        </div>

                        <div class="info-section">
                            <h4>Legend</h4>
                            <div id="color-legend" class="color-legend">
                                <!-- Color legend will be populated dynamically -->
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Entity Details Panel -->
                <div class="entity-details" id="entity-details" style="display: none;">
                    <div class="details-header">
                        <h4 id="details-title">Entity Details</h4>
                        <button class="btn btn-sm btn-ghost" onclick="relationshipVisualization.closeEntityDetails()">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>
                    <div class="details-content" id="details-content">
                        <!-- Entity details will be populated dynamically -->
                    </div>
                </div>

                <div class="modal-actions">
                    <button type="button" class="btn btn-secondary" onclick="relationshipVisualization.exportGraph()">
                        <i class="fas fa-download"></i> Export
                    </button>
                    <button type="button" class="btn btn-secondary" onclick="relationshipVisualization.saveLayout()">
                        <i class="fas fa-save"></i> Save Layout
                    </button>
                    <button type="button" class="btn btn-secondary" onclick="relationshipVisualization.closeVisualizer('${modalId}')">
                        Close
                    </button>
                </div>
            </div>
        `;
    }

    async initializeVisualizer(entities = []) {
        try {
            // Setup D3 visualization
            this.setupD3Visualization();
            
            // Load relationship data
            await this.loadRelationshipData(entities);
            
            // Setup event listeners
            this.setupEventListeners();
            
            // Initialize layout
            this.updateLayout();
            
            // Update UI
            this.updateColorLegend();
            this.updateGraphStatistics();
            
        } catch (error) {
            console.error('Failed to initialize relationship visualizer:', error);
            this.showNotification('Failed to load relationship data', 'error');
        }
    }

    setupD3Visualization() {
        const container = document.getElementById('graph-container');
        const width = container.clientWidth;
        const height = container.clientHeight;

        // Clear existing content
        container.innerHTML = '';

        // Create SVG
        this.svg = d3.select(container)
            .append('svg')
            .attr('width', width)
            .attr('height', height)
            .style('background', 'var(--bg-card)');

        // Setup zoom behavior
        this.zoom = d3.zoom()
            .scaleExtent([0.1, 4])
            .on('zoom', (event) => {
                this.g.attr('transform', event.transform);
            });

        this.svg.call(this.zoom);

        // Create main group for graph elements
        this.g = this.svg.append('g');

        // Create groups for different elements
        this.linkGroup = this.g.append('g').attr('class', 'links');
        this.nodeGroup = this.g.append('g').attr('class', 'nodes');
        this.labelGroup = this.g.append('g').attr('class', 'labels');

        // Store dimensions
        this.width = width;
        this.height = height;
    }

    async loadRelationshipData(entities = []) {
        const apiClient = window.apiClient || (window.EntityDBClient ? new window.EntityDBClient() : null);
        if (!apiClient) {
            throw new Error('API client not available');
        }

        try {
            const token = localStorage.getItem('entitydb-admin-token');
            if (token && apiClient.setToken) {
                apiClient.setToken(token);
            }

            // If no specific entities provided, load from entity browser
            if (entities.length === 0 && window.entityBrowserEnhanced) {
                entities = window.entityBrowserEnhanced.entities.slice(0, 100); // Limit for performance
            }

            // Convert entities to nodes
            this.nodes = entities.map(entity => ({
                id: entity.id,
                entity: entity,
                title: this.getEntityTitle(entity),
                type: this.getEntityType(entity),
                status: this.getEntityStatus(entity),
                createdAt: entity.created_at,
                tags: entity.tags || [],
                degree: 0,
                x: Math.random() * this.width,
                y: Math.random() * this.height
            }));

            // Load relationships
            const relationships = await this.loadEntityRelationships(entities);
            this.edges = relationships.map(rel => ({
                id: rel.id,
                source: rel.from_entity_id,
                target: rel.to_entity_id,
                type: rel.relationship_type || 'related',
                label: rel.relationship_type || 'related',
                weight: 1,
                relationship: rel
            }));

            // Calculate node degrees
            this.calculateNodeDegrees();

            // Filter out nodes without edges if needed
            this.filterIsolatedNodes();

            console.log(`Loaded ${this.nodes.length} nodes and ${this.edges.length} edges`);

        } catch (error) {
            console.error('Failed to load relationship data:', error);
            // Create sample data for demo
            this.createSampleData();
        }
    }

    async loadEntityRelationships(entities) {
        const apiClient = window.apiClient || (window.EntityDBClient ? new window.EntityDBClient() : null);
        const entityIds = entities.map(e => e.id);
        const relationships = [];

        try {
            // Load relationships for each entity
            for (const entityId of entityIds.slice(0, 50)) { // Limit for performance
                try {
                    const response = await apiClient.get('/api/v1/entity-relationships', {
                        entity_id: entityId,
                        limit: 20
                    });
                    
                    if (response.relationships) {
                        relationships.push(...response.relationships);
                    }
                } catch (error) {
                    console.warn(`Failed to load relationships for entity ${entityId}:`, error);
                }
            }

            return relationships;
        } catch (error) {
            console.error('Failed to load entity relationships:', error);
            return [];
        }
    }

    createSampleData() {
        // Create sample nodes for demonstration
        const sampleTypes = ['user', 'document', 'task', 'note'];
        const sampleStatuses = ['active', 'inactive', 'pending'];
        
        this.nodes = Array.from({length: 20}, (_, i) => ({
            id: `sample-${i}`,
            title: `Entity ${i + 1}`,
            type: sampleTypes[Math.floor(Math.random() * sampleTypes.length)],
            status: sampleStatuses[Math.floor(Math.random() * sampleStatuses.length)],
            degree: 0,
            x: Math.random() * this.width,
            y: Math.random() * this.height,
            entity: { id: `sample-${i}`, tags: [`type:${sampleTypes[Math.floor(Math.random() * sampleTypes.length)]}`] }
        }));

        // Create sample edges
        this.edges = [];
        for (let i = 0; i < 30; i++) {
            const source = Math.floor(Math.random() * this.nodes.length);
            const target = Math.floor(Math.random() * this.nodes.length);
            
            if (source !== target) {
                this.edges.push({
                    id: `edge-${i}`,
                    source: this.nodes[source].id,
                    target: this.nodes[target].id,
                    type: 'related',
                    label: 'related',
                    weight: 1
                });
            }
        }

        this.calculateNodeDegrees();
    }

    calculateNodeDegrees() {
        // Reset degrees
        this.nodes.forEach(node => node.degree = 0);
        
        // Count connections
        this.edges.forEach(edge => {
            const sourceNode = this.nodes.find(n => n.id === edge.source);
            const targetNode = this.nodes.find(n => n.id === edge.target);
            
            if (sourceNode) sourceNode.degree++;
            if (targetNode) targetNode.degree++;
        });
    }

    filterIsolatedNodes() {
        if (this.nodes.length > 50) {
            // Remove isolated nodes if we have too many nodes
            this.nodes = this.nodes.filter(node => node.degree > 0);
        }
    }

    updateLayout() {
        if (!this.svg || this.nodes.length === 0) return;

        switch (this.layoutMode) {
            case 'force':
                this.applyForceLayout();
                break;
            case 'hierarchy':
                this.applyHierarchicalLayout();
                break;
            case 'circular':
                this.applyCircularLayout();
                break;
            case 'grid':
                this.applyGridLayout();
                break;
            default:
                this.applyForceLayout();
        }
    }

    applyForceLayout() {
        // Setup force simulation
        this.simulation = d3.forceSimulation(this.nodes)
            .force('link', d3.forceLink(this.edges)
                .id(d => d.id)
                .distance(this.settings.linkDistance)
                .strength(this.settings.linkStrength))
            .force('charge', d3.forceManyBody()
                .strength(this.settings.chargeStrength))
            .force('center', d3.forceCenter(this.width / 2, this.height / 2)
                .strength(this.settings.centerForce))
            .force('collision', d3.forceCollide()
                .radius(this.settings.collisionRadius));

        this.renderGraph();
        
        if (this.animationEnabled) {
            this.simulation.on('tick', () => {
                this.updatePositions();
            });
        } else {
            // Run simulation silently for static layout
            this.simulation.stop();
            for (let i = 0; i < 300; ++i) this.simulation.tick();
            this.updatePositions();
        }
    }

    applyHierarchicalLayout() {
        // Simplified hierarchical layout
        const root = this.nodes.find(n => n.degree === Math.max(...this.nodes.map(n => n.degree))) || this.nodes[0];
        const levels = this.calculateHierarchyLevels(root);
        
        levels.forEach((levelNodes, level) => {
            const y = (level + 1) * (this.height / (levels.length + 1));
            levelNodes.forEach((node, index) => {
                node.x = (index + 1) * (this.width / (levelNodes.length + 1));
                node.y = y;
                node.fx = node.x;
                node.fy = node.y;
            });
        });

        this.renderGraph();
        this.updatePositions();
    }

    applyCircularLayout() {
        const centerX = this.width / 2;
        const centerY = this.height / 2;
        const radius = Math.min(this.width, this.height) * 0.3;
        
        this.nodes.forEach((node, index) => {
            const angle = (index / this.nodes.length) * 2 * Math.PI;
            node.x = centerX + radius * Math.cos(angle);
            node.y = centerY + radius * Math.sin(angle);
            node.fx = node.x;
            node.fy = node.y;
        });

        this.renderGraph();
        this.updatePositions();
    }

    applyGridLayout() {
        const cols = Math.ceil(Math.sqrt(this.nodes.length));
        const cellWidth = this.width / cols;
        const cellHeight = this.height / Math.ceil(this.nodes.length / cols);
        
        this.nodes.forEach((node, index) => {
            const row = Math.floor(index / cols);
            const col = index % cols;
            node.x = col * cellWidth + cellWidth / 2;
            node.y = row * cellHeight + cellHeight / 2;
            node.fx = node.x;
            node.fy = node.y;
        });

        this.renderGraph();
        this.updatePositions();
    }

    calculateHierarchyLevels(root) {
        const visited = new Set();
        const levels = [];
        const queue = [{node: root, level: 0}];
        
        while (queue.length > 0) {
            const {node, level} = queue.shift();
            
            if (visited.has(node.id)) continue;
            visited.add(node.id);
            
            if (!levels[level]) levels[level] = [];
            levels[level].push(node);
            
            // Find connected nodes
            this.edges.forEach(edge => {
                let connectedNode = null;
                if (edge.source === node.id) {
                    connectedNode = this.nodes.find(n => n.id === edge.target);
                } else if (edge.target === node.id) {
                    connectedNode = this.nodes.find(n => n.id === edge.source);
                }
                
                if (connectedNode && !visited.has(connectedNode.id)) {
                    queue.push({node: connectedNode, level: level + 1});
                }
            });
        }
        
        return levels;
    }

    renderGraph() {
        // Clear previous elements
        this.linkGroup.selectAll('*').remove();
        this.nodeGroup.selectAll('*').remove();
        this.labelGroup.selectAll('*').remove();

        // Render links
        this.renderLinks();
        
        // Render nodes
        this.renderNodes();
        
        // Render labels if enabled
        if (this.showLabels) {
            this.renderLabels();
        }
    }

    renderLinks() {
        const links = this.linkGroup.selectAll('line')
            .data(this.edges)
            .enter()
            .append('line')
            .attr('class', 'graph-link')
            .attr('stroke', '#95a5a6')
            .attr('stroke-width', d => Math.sqrt(d.weight) * 2)
            .attr('stroke-opacity', 0.6);

        // Add link labels for important relationships
        const linkLabels = this.linkGroup.selectAll('text')
            .data(this.edges.filter(d => d.type !== 'related'))
            .enter()
            .append('text')
            .attr('class', 'link-label')
            .attr('text-anchor', 'middle')
            .attr('font-size', '10px')
            .attr('fill', '#7f8c8d')
            .text(d => d.label);

        this.links = links;
        this.linkLabels = linkLabels;
    }

    renderNodes() {
        const nodes = this.nodeGroup.selectAll('circle')
            .data(this.nodes)
            .enter()
            .append('circle')
            .attr('class', 'graph-node')
            .attr('r', d => this.getNodeSize(d))
            .attr('fill', d => this.getNodeColor(d))
            .attr('stroke', '#ffffff')
            .attr('stroke-width', 2)
            .style('cursor', 'pointer');

        // Add node interactions
        nodes
            .on('click', (event, d) => this.handleNodeClick(event, d))
            .on('dblclick', (event, d) => this.handleNodeDoubleClick(event, d))
            .on('mouseover', (event, d) => this.handleNodeMouseOver(event, d))
            .on('mouseout', (event, d) => this.handleNodeMouseOut(event, d));

        // Add drag behavior
        if (this.layoutMode === 'force') {
            nodes.call(d3.drag()
                .on('start', (event, d) => this.dragStart(event, d))
                .on('drag', (event, d) => this.dragMove(event, d))
                .on('end', (event, d) => this.dragEnd(event, d)));
        }

        this.nodes_selection = nodes;
    }

    renderLabels() {
        const labels = this.labelGroup.selectAll('text')
            .data(this.nodes)
            .enter()
            .append('text')
            .attr('class', 'node-label')
            .attr('text-anchor', 'middle')
            .attr('font-size', '12px')
            .attr('font-weight', '500')
            .attr('fill', '#2c3e50')
            .attr('dy', '0.35em')
            .text(d => d.title.length > 15 ? d.title.substring(0, 15) + '...' : d.title)
            .style('pointer-events', 'none');

        this.labels = labels;
    }

    updatePositions() {
        if (this.links) {
            this.links
                .attr('x1', d => {
                    const source = this.nodes.find(n => n.id === d.source.id || n.id === d.source);
                    return source ? source.x : 0;
                })
                .attr('y1', d => {
                    const source = this.nodes.find(n => n.id === d.source.id || n.id === d.source);
                    return source ? source.y : 0;
                })
                .attr('x2', d => {
                    const target = this.nodes.find(n => n.id === d.target.id || n.id === d.target);
                    return target ? target.x : 0;
                })
                .attr('y2', d => {
                    const target = this.nodes.find(n => n.id === d.target.id || n.id === d.target);
                    return target ? target.y : 0;
                });
        }

        if (this.linkLabels) {
            this.linkLabels
                .attr('x', d => {
                    const source = this.nodes.find(n => n.id === d.source.id || n.id === d.source);
                    const target = this.nodes.find(n => n.id === d.target.id || n.id === d.target);
                    return source && target ? (source.x + target.x) / 2 : 0;
                })
                .attr('y', d => {
                    const source = this.nodes.find(n => n.id === d.source.id || n.id === d.source);
                    const target = this.nodes.find(n => n.id === d.target.id || n.id === d.target);
                    return source && target ? (source.y + target.y) / 2 : 0;
                });
        }

        if (this.nodes_selection) {
            this.nodes_selection
                .attr('cx', d => d.x)
                .attr('cy', d => d.y);
        }

        if (this.labels) {
            this.labels
                .attr('x', d => d.x)
                .attr('y', d => d.y + this.getNodeSize(d) + 15);
        }
    }

    // Node interaction handlers
    handleNodeClick(event, node) {
        event.stopPropagation();
        
        if (event.ctrlKey || event.metaKey) {
            // Multi-select
            if (this.selectedNodes.has(node.id)) {
                this.selectedNodes.delete(node.id);
            } else {
                this.selectedNodes.add(node.id);
            }
        } else {
            // Single select
            this.selectedNodes.clear();
            this.selectedNodes.add(node.id);
        }
        
        this.updateNodeSelection();
        this.updateSelectionInfo();
    }

    handleNodeDoubleClick(event, node) {
        event.stopPropagation();
        this.showEntityDetails(node);
    }

    handleNodeMouseOver(event, node) {
        // Highlight connected nodes and edges
        this.highlightConnections(node);
        
        // Show tooltip
        this.showNodeTooltip(event, node);
    }

    handleNodeMouseOut(event, node) {
        // Remove highlights
        this.clearHighlights();
        
        // Hide tooltip
        this.hideNodeTooltip();
    }

    // Drag handlers for force layout
    dragStart(event, d) {
        if (!event.active && this.simulation) {
            this.simulation.alphaTarget(0.3).restart();
        }
        d.fx = d.x;
        d.fy = d.y;
    }

    dragMove(event, d) {
        d.fx = event.x;
        d.fy = event.y;
    }

    dragEnd(event, d) {
        if (!event.active && this.simulation) {
            this.simulation.alphaTarget(0);
        }
        d.fx = null;
        d.fy = null;
    }

    // Visual helpers
    getNodeSize(node) {
        if (!this.showMetrics) return this.settings.nodeSize.default;
        
        const minSize = this.settings.nodeSize.min;
        const maxSize = this.settings.nodeSize.max;
        const maxDegree = Math.max(...this.nodes.map(n => n.degree));
        
        if (maxDegree === 0) return this.settings.nodeSize.default;
        
        const normalizedDegree = node.degree / maxDegree;
        return minSize + (maxSize - minSize) * normalizedDegree;
    }

    getNodeColor(node) {
        const scheme = this.colorSchemes[this.colorScheme];
        if (!scheme) return scheme.colors.default || '#34495e';
        
        switch (this.colorScheme) {
            case 'type':
                return scheme.colors[node.type] || scheme.colors.default;
            case 'status':
                return scheme.colors[node.status] || scheme.colors.default;
            case 'degree':
                if (node.degree > 10) return scheme.colors.high;
                if (node.degree > 4) return scheme.colors.medium;
                if (node.degree > 0) return scheme.colors.low;
                return scheme.colors.isolated;
            case 'creation':
                return this.getCreationTimeColor(node, scheme);
            default:
                return scheme.colors.default;
        }
    }

    getCreationTimeColor(node, scheme) {
        if (!node.createdAt) return scheme.colors.default;
        
        const now = Date.now() * 1000000; // Convert to nanoseconds
        const age = now - node.createdAt;
        const day = 24 * 60 * 60 * 1000 * 1000000;
        const week = 7 * day;
        const month = 30 * day;
        
        if (age < day) return scheme.colors.recent;
        if (age < week) return scheme.colors.week;
        if (age < month) return scheme.colors.month;
        return scheme.colors.older;
    }

    // Entity data helpers
    getEntityTitle(entity) {
        if (!entity.tags) return `Entity ${entity.id.substring(0, 8)}`;
        
        const titleTag = entity.tags.find(t => this.stripTimestamp(t).startsWith('title:'));
        if (titleTag) {
            return this.stripTimestamp(titleTag).split(':').slice(1).join(':');
        }
        
        const nameTag = entity.tags.find(t => this.stripTimestamp(t).startsWith('name:'));
        if (nameTag) {
            return this.stripTimestamp(nameTag).split(':').slice(1).join(':');
        }
        
        return `Entity ${entity.id.substring(0, 8)}`;
    }

    getEntityType(entity) {
        if (!entity.tags) return 'default';
        
        const typeTag = entity.tags.find(t => this.stripTimestamp(t).startsWith('type:'));
        if (typeTag) {
            return this.stripTimestamp(typeTag).split(':')[1] || 'default';
        }
        
        return 'default';
    }

    getEntityStatus(entity) {
        if (!entity.tags) return 'default';
        
        const statusTag = entity.tags.find(t => this.stripTimestamp(t).startsWith('status:'));
        if (statusTag) {
            return this.stripTimestamp(statusTag).split(':')[1] || 'default';
        }
        
        return 'default';
    }

    stripTimestamp(tag) {
        if (typeof tag !== 'string') return tag;
        const pipeIndex = tag.indexOf('|');
        return pipeIndex !== -1 ? tag.substring(pipeIndex + 1) : tag;
    }

    // UI update methods
    updateNodeSelection() {
        if (this.nodes_selection) {
            this.nodes_selection
                .attr('stroke', d => this.selectedNodes.has(d.id) ? '#e74c3c' : '#ffffff')
                .attr('stroke-width', d => this.selectedNodes.has(d.id) ? 3 : 2);
        }
    }

    updateSelectionInfo() {
        const container = document.getElementById('selection-info');
        if (!container) return;
        
        if (this.selectedNodes.size === 0) {
            container.innerHTML = '<p class="text-muted">No nodes selected</p>';
            return;
        }
        
        const selectedNodeData = this.nodes.filter(n => this.selectedNodes.has(n.id));
        const html = `
            <div class="selection-summary">
                <p><strong>${this.selectedNodes.size}</strong> nodes selected</p>
                <div class="selected-nodes">
                    ${selectedNodeData.map(node => `
                        <div class="selected-node">
                            <span class="node-color" style="background: ${this.getNodeColor(node)}"></span>
                            <span class="node-title">${node.title}</span>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
        
        container.innerHTML = html;
    }

    updateGraphStatistics() {
        document.getElementById('node-count').textContent = this.nodes.length;
        document.getElementById('edge-count').textContent = this.edges.length;
        
        // Calculate graph density
        const maxEdges = this.nodes.length * (this.nodes.length - 1) / 2;
        const density = maxEdges > 0 ? ((this.edges.length / maxEdges) * 100).toFixed(1) : '0';
        document.getElementById('graph-density').textContent = density + '%';
        
        // Calculate connected components (simplified)
        document.getElementById('component-count').textContent = this.calculateConnectedComponents();
    }

    calculateConnectedComponents() {
        const visited = new Set();
        let components = 0;
        
        for (const node of this.nodes) {
            if (!visited.has(node.id)) {
                this.dfsComponent(node.id, visited);
                components++;
            }
        }
        
        return components;
    }

    dfsComponent(nodeId, visited) {
        visited.add(nodeId);
        
        this.edges.forEach(edge => {
            let connectedId = null;
            if (edge.source === nodeId) connectedId = edge.target;
            else if (edge.target === nodeId) connectedId = edge.source;
            
            if (connectedId && !visited.has(connectedId)) {
                this.dfsComponent(connectedId, visited);
            }
        });
    }

    updateColorLegend() {
        const container = document.getElementById('color-legend');
        if (!container) return;
        
        const scheme = this.colorSchemes[this.colorScheme];
        if (!scheme) return;
        
        const html = Object.entries(scheme.colors)
            .filter(([key]) => key !== 'default')
            .map(([key, color]) => `
                <div class="legend-item">
                    <span class="legend-color" style="background: ${color}"></span>
                    <span class="legend-label">${key}</span>
                </div>
            `).join('');
        
        container.innerHTML = html;
    }

    // Event listeners setup
    setupEventListeners() {
        // Layout buttons
        document.querySelectorAll('.layout-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                document.querySelectorAll('.layout-btn').forEach(b => b.classList.remove('active'));
                e.target.classList.add('active');
                this.layoutMode = e.target.dataset.layout;
                this.updateLayout();
            });
        });

        // Color scheme selector
        document.getElementById('color-scheme').addEventListener('change', (e) => {
            this.colorScheme = e.target.value;
            this.updateNodeColors();
            this.updateColorLegend();
        });

        // Display options
        document.getElementById('show-labels').addEventListener('change', (e) => {
            this.showLabels = e.target.checked;
            this.toggleLabels();
        });

        document.getElementById('show-metrics').addEventListener('change', (e) => {
            this.showMetrics = e.target.checked;
            this.updateNodeSizes();
        });

        document.getElementById('enable-animation').addEventListener('change', (e) => {
            this.animationEnabled = e.target.checked;
            if (this.simulation) {
                if (this.animationEnabled) {
                    this.simulation.restart();
                } else {
                    this.simulation.stop();
                }
            }
        });

        // Filters
        document.getElementById('node-filter').addEventListener('input', (e) => {
            this.filterNodes(e.target.value);
        });

        document.getElementById('relationship-filter').addEventListener('change', (e) => {
            this.filterRelationships(e.target.value);
        });

        // SVG click to clear selection
        if (this.svg) {
            this.svg.on('click', () => {
                this.selectedNodes.clear();
                this.updateNodeSelection();
                this.updateSelectionInfo();
            });
        }
    }

    updateNodeColors() {
        if (this.nodes_selection) {
            this.nodes_selection.attr('fill', d => this.getNodeColor(d));
        }
    }

    updateNodeSizes() {
        if (this.nodes_selection) {
            this.nodes_selection.attr('r', d => this.getNodeSize(d));
        }
    }

    toggleLabels() {
        if (this.labels) {
            this.labels.style('display', this.showLabels ? 'block' : 'none');
        }
    }

    // Control methods
    resetZoom() {
        if (this.svg && this.zoom) {
            this.svg.transition().duration(750).call(
                this.zoom.transform,
                d3.zoomIdentity
            );
        }
    }

    centerGraph() {
        if (!this.svg || this.nodes.length === 0) return;
        
        const bounds = this.getGraphBounds();
        const centerX = (bounds.minX + bounds.maxX) / 2;
        const centerY = (bounds.minY + bounds.maxY) / 2;
        
        const translateX = this.width / 2 - centerX;
        const translateY = this.height / 2 - centerY;
        
        this.svg.transition().duration(750).call(
            this.zoom.transform,
            d3.zoomIdentity.translate(translateX, translateY)
        );
    }

    fitToScreen() {
        if (!this.svg || this.nodes.length === 0) return;
        
        const bounds = this.getGraphBounds();
        const graphWidth = bounds.maxX - bounds.minX;
        const graphHeight = bounds.maxY - bounds.minY;
        
        if (graphWidth === 0 || graphHeight === 0) return;
        
        const scale = Math.min(
            this.width / graphWidth,
            this.height / graphHeight
        ) * 0.9; // 90% to add some padding
        
        const centerX = (bounds.minX + bounds.maxX) / 2;
        const centerY = (bounds.minY + bounds.maxY) / 2;
        
        this.svg.transition().duration(750).call(
            this.zoom.transform,
            d3.zoomIdentity
                .translate(this.width / 2, this.height / 2)
                .scale(scale)
                .translate(-centerX, -centerY)
        );
    }

    getGraphBounds() {
        const xs = this.nodes.map(d => d.x);
        const ys = this.nodes.map(d => d.y);
        
        return {
            minX: Math.min(...xs),
            maxX: Math.max(...xs),
            minY: Math.min(...ys),
            maxY: Math.max(...ys)
        };
    }

    // Filtering methods
    filterNodes(query) {
        if (!query.trim()) {
            this.showAllNodes();
            return;
        }
        
        const lowerQuery = query.toLowerCase();
        this.nodes.forEach(node => {
            const matches = node.title.toLowerCase().includes(lowerQuery) ||
                           node.type.toLowerCase().includes(lowerQuery) ||
                           node.status.toLowerCase().includes(lowerQuery);
            
            node.filtered = !matches;
        });
        
        this.updateNodeVisibility();
    }

    filterRelationships(type) {
        if (!type) {
            this.showAllEdges();
            return;
        }
        
        this.edges.forEach(edge => {
            edge.filtered = edge.type !== type;
        });
        
        this.updateEdgeVisibility();
    }

    showAllNodes() {
        this.nodes.forEach(node => node.filtered = false);
        this.updateNodeVisibility();
    }

    showAllEdges() {
        this.edges.forEach(edge => edge.filtered = false);
        this.updateEdgeVisibility();
    }

    updateNodeVisibility() {
        if (this.nodes_selection) {
            this.nodes_selection.style('opacity', d => d.filtered ? 0.1 : 1);
        }
        if (this.labels) {
            this.labels.style('opacity', d => d.filtered ? 0.1 : 1);
        }
    }

    updateEdgeVisibility() {
        if (this.links) {
            this.links.style('opacity', d => d.filtered ? 0.1 : 0.6);
        }
        if (this.linkLabels) {
            this.linkLabels.style('opacity', d => d.filtered ? 0.1 : 1);
        }
    }

    // Additional features
    highlightConnections(node) {
        const connectedEdges = this.edges.filter(e => 
            e.source === node.id || e.target === node.id ||
            e.source.id === node.id || e.target.id === node.id
        );
        
        const connectedNodeIds = new Set();
        connectedEdges.forEach(edge => {
            connectedNodeIds.add(edge.source.id || edge.source);
            connectedNodeIds.add(edge.target.id || edge.target);
        });
        
        // Highlight connected nodes
        if (this.nodes_selection) {
            this.nodes_selection.style('opacity', d => 
                d.id === node.id || connectedNodeIds.has(d.id) ? 1 : 0.3
            );
        }
        
        // Highlight connected edges
        if (this.links) {
            this.links.style('opacity', d => 
                connectedEdges.includes(d) ? 1 : 0.1
            );
        }
    }

    clearHighlights() {
        if (this.nodes_selection) {
            this.nodes_selection.style('opacity', d => d.filtered ? 0.1 : 1);
        }
        if (this.links) {
            this.links.style('opacity', d => d.filtered ? 0.1 : 0.6);
        }
    }

    showNodeTooltip(event, node) {
        // Create or update tooltip
        let tooltip = d3.select('body').select('.graph-tooltip');
        if (tooltip.empty()) {
            tooltip = d3.select('body').append('div')
                .attr('class', 'graph-tooltip')
                .style('position', 'absolute')
                .style('background', 'rgba(0, 0, 0, 0.8)')
                .style('color', 'white')
                .style('padding', '8px 12px')
                .style('border-radius', '4px')
                .style('font-size', '12px')
                .style('pointer-events', 'none')
                .style('opacity', 0);
        }
        
        const html = `
            <strong>${node.title}</strong><br>
            Type: ${node.type}<br>
            Status: ${node.status}<br>
            Connections: ${node.degree}
        `;
        
        tooltip.html(html)
            .style('left', (event.pageX + 10) + 'px')
            .style('top', (event.pageY - 10) + 'px')
            .transition()
            .duration(200)
            .style('opacity', 1);
    }

    hideNodeTooltip() {
        d3.select('.graph-tooltip')
            .transition()
            .duration(200)
            .style('opacity', 0);
    }

    showEntityDetails(node) {
        const panel = document.getElementById('entity-details');
        const title = document.getElementById('details-title');
        const content = document.getElementById('details-content');
        
        if (!panel || !title || !content) return;
        
        title.textContent = node.title;
        
        const html = `
            <div class="entity-detail-content">
                <div class="detail-section">
                    <h5>Basic Information</h5>
                    <div class="detail-grid">
                        <div class="detail-item">
                            <span class="detail-label">ID:</span>
                            <span class="detail-value">${node.id}</span>
                        </div>
                        <div class="detail-item">
                            <span class="detail-label">Type:</span>
                            <span class="detail-value">${node.type}</span>
                        </div>
                        <div class="detail-item">
                            <span class="detail-label">Status:</span>
                            <span class="detail-value">${node.status}</span>
                        </div>
                        <div class="detail-item">
                            <span class="detail-label">Connections:</span>
                            <span class="detail-value">${node.degree}</span>
                        </div>
                    </div>
                </div>
                
                <div class="detail-section">
                    <h5>Tags</h5>
                    <div class="entity-tags">
                        ${node.tags.map(tag => `<span class="tag-badge">${this.stripTimestamp(tag)}</span>`).join('')}
                    </div>
                </div>
                
                <div class="detail-section">
                    <h5>Connected Entities</h5>
                    <div class="connected-entities">
                        ${this.getConnectedEntities(node).map(conn => `
                            <div class="connected-entity">
                                <span class="connection-type">${conn.type}</span>
                                <span class="connected-title">${conn.title}</span>
                            </div>
                        `).join('') || '<p class="text-muted">No connections found</p>'}
                    </div>
                </div>
            </div>
        `;
        
        content.innerHTML = html;
        panel.style.display = 'block';
    }

    getConnectedEntities(node) {
        return this.edges
            .filter(edge => 
                edge.source === node.id || edge.target === node.id ||
                edge.source.id === node.id || edge.target.id === node.id
            )
            .map(edge => {
                const connectedId = (edge.source.id || edge.source) === node.id ? 
                    (edge.target.id || edge.target) : (edge.source.id || edge.source);
                const connectedNode = this.nodes.find(n => n.id === connectedId);
                
                return {
                    id: connectedId,
                    title: connectedNode ? connectedNode.title : connectedId,
                    type: edge.type || 'related'
                };
            });
    }

    closeEntityDetails() {
        const panel = document.getElementById('entity-details');
        if (panel) {
            panel.style.display = 'none';
        }
    }

    // Export and save methods
    exportGraph() {
        const graphData = {
            nodes: this.nodes,
            edges: this.edges,
            settings: this.settings,
            layout: this.layoutMode,
            colorScheme: this.colorScheme,
            timestamp: Date.now()
        };
        
        const blob = new Blob([JSON.stringify(graphData, null, 2)], { 
            type: 'application/json' 
        });
        
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `entitydb_graph_${new Date().toISOString().slice(0, 10)}.json`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);
        
        this.showNotification('Graph exported successfully', 'success');
    }

    saveLayout() {
        const layout = {
            nodes: this.nodes.map(n => ({ id: n.id, x: n.x, y: n.y })),
            mode: this.layoutMode,
            colorScheme: this.colorScheme,
            settings: this.settings
        };
        
        localStorage.setItem('entitydb-graph-layout', JSON.stringify(layout));
        this.showNotification('Layout saved', 'success');
    }

    loadSavedLayout() {
        try {
            const saved = localStorage.getItem('entitydb-graph-layout');
            if (saved) {
                const layout = JSON.parse(saved);
                
                // Apply saved positions
                layout.nodes.forEach(savedNode => {
                    const node = this.nodes.find(n => n.id === savedNode.id);
                    if (node) {
                        node.x = savedNode.x;
                        node.y = savedNode.y;
                    }
                });
                
                this.layoutMode = layout.mode || 'force';
                this.colorScheme = layout.colorScheme || 'type';
                
                this.updateLayout();
                this.showNotification('Layout loaded', 'success');
            }
        } catch (e) {
            console.warn('Failed to load saved layout:', e);
        }
    }

    // Preferences management
    loadVisualizationPreferences() {
        try {
            const prefs = localStorage.getItem('entitydb-visualization-preferences');
            if (prefs) {
                const parsed = JSON.parse(prefs);
                this.layoutMode = parsed.layoutMode || 'force';
                this.colorScheme = parsed.colorScheme || 'type';
                this.showLabels = parsed.showLabels !== false;
                this.showMetrics = parsed.showMetrics !== false;
                this.animationEnabled = parsed.animationEnabled !== false;
                this.settings = { ...this.settings, ...parsed.settings };
            }
        } catch (e) {
            console.warn('Failed to load visualization preferences:', e);
        }
    }

    saveVisualizationPreferences() {
        const prefs = {
            layoutMode: this.layoutMode,
            colorScheme: this.colorScheme,
            showLabels: this.showLabels,
            showMetrics: this.showMetrics,
            animationEnabled: this.animationEnabled,
            settings: this.settings
        };
        localStorage.setItem('entitydb-visualization-preferences', JSON.stringify(prefs));
    }

    // Modal integration
    createModal(id, content, options = {}) {
        const modal = document.createElement('div');
        modal.id = id;
        modal.className = `modal ${options.size || 'large'}`;
        modal.innerHTML = `
            <div class="modal-backdrop" onclick="relationshipVisualization.closeVisualizer('${id}')"></div>
            <div class="modal-dialog">
                <div class="modal-header">
                    <h2 class="modal-title">${options.title || 'Relationship Visualization'}</h2>
                    <button class="modal-close" onclick="relationshipVisualization.closeVisualizer('${id}')">
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

    closeVisualizer(modalId) {
        // Clean up D3 elements
        if (this.simulation) {
            this.simulation.stop();
        }
        
        // Save preferences
        this.saveVisualizationPreferences();
        
        // Close modal
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.remove('show');
            setTimeout(() => {
                modal.remove();
                document.body.classList.remove('modal-open');
            }, 300);
        }
    }

    showNotification(message, type = 'info') {
        if (window.notificationSystem) {
            window.notificationSystem.show(message, type);
        } else {
            console.log(`${type}: ${message}`);
        }
    }
}

// Initialize the relationship visualization system
if (typeof window !== 'undefined') {
    window.RelationshipVisualization = RelationshipVisualization;
    
    const initVisualization = () => {
        // Load D3.js if not already loaded
        if (typeof d3 === 'undefined') {
            const script = document.createElement('script');
            script.src = 'https://d3js.org/d3.v7.min.js';
            script.onload = () => {
                window.relationshipVisualization = new RelationshipVisualization();
            };
            document.head.appendChild(script);
        } else {
            window.relationshipVisualization = new RelationshipVisualization();
        }
    };
    
    // Wait for DOM to be ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initVisualization);
    } else {
        initVisualization();
    }
}