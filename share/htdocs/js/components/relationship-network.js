/**
 * Relationship Network Component - Professional network visualization
 * Uses D3.js for interactive entity relationship graphs
 */
class RelationshipNetwork extends BaseComponent {
    get defaultOptions() {
        return {
            width: 800,
            height: 600,
            nodeRadius: 20,
            linkDistance: 100,
            linkStrength: 0.5,
            chargeStrength: -300,
            collideRadius: 25,
            maxDepth: 3,
            showLabels: true,
            interactive: true,
            zoom: true
        };
    }

    get defaultState() {
        return {
            focusEntityId: null,
            nodes: [],
            links: [],
            networkData: null,
            loading: false,
            error: null,
            selectedNode: null,
            depth: 1
        };
    }

    init() {
        super.init();
        this.setupD3();
        this.setupControls();
    }

    setupD3() {
        // Create SVG container
        this.svg = d3.select(this.container)
            .append('svg')
            .attr('class', 'network-svg')
            .attr('width', this.options.width)
            .attr('height', this.options.height);

        // Setup zoom if enabled
        if (this.options.zoom) {
            this.zoom = d3.zoom()
                .scaleExtent([0.1, 4])
                .on('zoom', (event) => {
                    this.networkGroup.attr('transform', event.transform);
                });
            
            this.svg.call(this.zoom);
        }

        // Create main group for network elements
        this.networkGroup = this.svg.append('g').attr('class', 'network-group');

        // Create groups for different elements
        this.linksGroup = this.networkGroup.append('g').attr('class', 'links');
        this.nodesGroup = this.networkGroup.append('g').attr('class', 'nodes');
        this.labelsGroup = this.networkGroup.append('g').attr('class', 'labels');

        // Setup force simulation
        this.simulation = d3.forceSimulation()
            .force('link', d3.forceLink().id(d => d.id).distance(this.options.linkDistance))
            .force('charge', d3.forceManyBody().strength(this.options.chargeStrength))
            .force('center', d3.forceCenter(this.options.width / 2, this.options.height / 2))
            .force('collide', d3.forceCollide().radius(this.options.collideRadius));
    }

    setupControls() {
        const controlsHtml = `
            <div class="network-controls">
                <div class="control-group">
                    <label>Depth:</label>
                    <select class="depth-selector">
                        <option value="1">1 Level</option>
                        <option value="2">2 Levels</option>
                        <option value="3">3 Levels</option>
                    </select>
                </div>
                
                <div class="control-group">
                    <button class="center-network-btn">Center View</button>
                    <button class="fit-network-btn">Fit to View</button>
                    <button class="reset-zoom-btn">Reset Zoom</button>
                </div>
                
                <div class="control-group">
                    <label>
                        <input type="checkbox" class="show-labels-toggle" ${this.options.showLabels ? 'checked' : ''}> 
                        Show Labels
                    </label>
                </div>
                
                <div class="network-info">
                    <span class="node-count">0 nodes</span>
                    <span class="link-count">0 links</span>
                </div>
            </div>
        `;

        this.container.insertAdjacentHTML('afterbegin', controlsHtml);
        this.setupControlListeners();
    }

    setupControlListeners() {
        // Depth selector
        this.addEventListener(this.find('.depth-selector'), 'change', (e) => {
            this.setState({ depth: parseInt(e.target.value) });
            if (this.state.focusEntityId) {
                this.loadNetwork(this.state.focusEntityId);
            }
        });

        // View controls
        this.addEventListener(this.find('.center-network-btn'), 'click', () => this.centerNetwork());
        this.addEventListener(this.find('.fit-network-btn'), 'click', () => this.fitToView());
        this.addEventListener(this.find('.reset-zoom-btn'), 'click', () => this.resetZoom());

        // Toggle labels
        this.addEventListener(this.find('.show-labels-toggle'), 'change', (e) => {
            this.options.showLabels = e.target.checked;
            this.updateLabels();
        });
    }

    async loadNetwork(entityId, depth = this.state.depth) {
        try {
            this.setLoading(true);
            this.setState({ 
                focusEntityId: entityId,
                depth: depth,
                error: null 
            });

            // Load network data from API
            const networkData = await apiClient.getEntityNetwork(entityId, depth);
            
            this.setState({ networkData });
            this.processNetworkData(networkData);
            this.renderNetwork();
            
        } catch (error) {
            this.handleError(error, 'RelationshipNetwork.loadNetwork');
            this.setState({ error: error.message });
        } finally {
            this.setLoading(false);
        }
    }

    processNetworkData(networkData) {
        // Process nodes
        const nodes = networkData.nodes.map(node => ({
            ...node,
            x: this.options.width / 2 + (Math.random() - 0.5) * 100,
            y: this.options.height / 2 + (Math.random() - 0.5) * 100,
            radius: node.is_focus ? this.options.nodeRadius * 1.5 : this.options.nodeRadius,
            color: this.getNodeColor(node)
        }));

        // Process links
        const links = networkData.edges.map(edge => ({
            ...edge,
            source: edge.source,
            target: edge.target,
            strength: edge.weight || 1
        }));

        this.setState({ nodes, links });
        this.updateNetworkInfo();
    }

    getNodeColor(node) {
        const typeColors = {
            'user': '#00d9ff',
            'session': '#ff6b35', 
            'config': '#00ff88',
            'relationship': '#ffb800',
            'unknown': '#94a3b8'
        };

        let color = typeColors[node.type] || typeColors.unknown;
        
        if (node.is_focus) {
            color = '#ff3366'; // Special color for focus node
        }

        return color;
    }

    renderNetwork() {
        // Clear existing elements
        this.linksGroup.selectAll('*').remove();
        this.nodesGroup.selectAll('*').remove();
        this.labelsGroup.selectAll('*').remove();

        // Render links
        const links = this.linksGroup
            .selectAll('.link')
            .data(this.state.links)
            .enter()
            .append('line')
            .attr('class', 'link')
            .attr('stroke', '#2a3441')
            .attr('stroke-width', d => Math.sqrt(d.strength) * 2)
            .attr('stroke-opacity', 0.6);

        // Render nodes
        const nodes = this.nodesGroup
            .selectAll('.node')
            .data(this.state.nodes)
            .enter()
            .append('circle')
            .attr('class', 'node')
            .attr('r', d => d.radius)
            .attr('fill', d => d.color)
            .attr('stroke', '#1a1f2e')
            .attr('stroke-width', 2)
            .style('cursor', this.options.interactive ? 'pointer' : 'default');

        // Add glow effect for focus node
        nodes.filter(d => d.is_focus)
            .attr('filter', 'url(#glow)');

        // Setup node interactions if enabled
        if (this.options.interactive) {
            nodes
                .on('click', (event, d) => this.handleNodeClick(event, d))
                .on('dblclick', (event, d) => this.handleNodeDoubleClick(event, d))
                .on('mouseover', (event, d) => this.handleNodeHover(event, d))
                .on('mouseout', (event, d) => this.handleNodeOut(event, d))
                .call(this.getDragBehavior());
        }

        // Render labels if enabled
        if (this.options.showLabels) {
            this.renderLabels();
        }

        // Setup filters and effects
        this.setupSVGFilters();

        // Update simulation
        this.simulation
            .nodes(this.state.nodes)
            .on('tick', () => this.updatePositions());

        this.simulation.force('link')
            .links(this.state.links);

        this.simulation.alpha(1).restart();
    }

    renderLabels() {
        const labels = this.labelsGroup
            .selectAll('.label')
            .data(this.state.nodes)
            .enter()
            .append('text')
            .attr('class', 'label')
            .attr('text-anchor', 'middle')
            .attr('dy', '.35em')
            .attr('fill', '#e2e8f0')
            .attr('font-size', '12px')
            .attr('font-family', 'SF Pro Display, sans-serif')
            .text(d => d.name || d.id.substring(0, 8))
            .style('pointer-events', 'none');
    }

    updateLabels() {
        this.labelsGroup.selectAll('.label')
            .style('display', this.options.showLabels ? 'block' : 'none');
    }

    setupSVGFilters() {
        // Define glow filter
        const defs = this.svg.select('defs').empty() ? this.svg.append('defs') : this.svg.select('defs');
        
        const glowFilter = defs.append('filter')
            .attr('id', 'glow')
            .attr('x', '-50%')
            .attr('y', '-50%')
            .attr('width', '200%')
            .attr('height', '200%');

        glowFilter.append('feGaussianBlur')
            .attr('stdDeviation', '3')
            .attr('result', 'coloredBlur');

        const feMerge = glowFilter.append('feMerge');
        feMerge.append('feMergeNode').attr('in', 'coloredBlur');
        feMerge.append('feMergeNode').attr('in', 'SourceGraphic');
    }

    getDragBehavior() {
        return d3.drag()
            .on('start', (event, d) => {
                if (!event.active) this.simulation.alphaTarget(0.3).restart();
                d.fx = d.x;
                d.fy = d.y;
            })
            .on('drag', (event, d) => {
                d.fx = event.x;
                d.fy = event.y;
            })
            .on('end', (event, d) => {
                if (!event.active) this.simulation.alphaTarget(0);
                d.fx = null;
                d.fy = null;
            });
    }

    updatePositions() {
        // Update link positions
        this.linksGroup.selectAll('.link')
            .attr('x1', d => d.source.x)
            .attr('y1', d => d.source.y)
            .attr('x2', d => d.target.x)
            .attr('y2', d => d.target.y);

        // Update node positions
        this.nodesGroup.selectAll('.node')
            .attr('cx', d => d.x)
            .attr('cy', d => d.y);

        // Update label positions
        this.labelsGroup.selectAll('.label')
            .attr('x', d => d.x)
            .attr('y', d => d.y + d.radius + 15);
    }

    // Event handlers
    handleNodeClick(event, node) {
        this.setState({ selectedNode: node });
        this.emit('node-select', { node });
        
        // Visual feedback
        this.nodesGroup.selectAll('.node')
            .attr('stroke-width', d => d === node ? 4 : 2);
    }

    handleNodeDoubleClick(event, node) {
        if (!node.is_focus) {
            this.loadNetwork(node.id, this.state.depth);
        }
        this.emit('node-focus', { node });
    }

    handleNodeHover(event, node) {
        // Highlight connected nodes and links
        const connectedNodeIds = new Set();
        
        this.state.links.forEach(link => {
            if (link.source.id === node.id) {
                connectedNodeIds.add(link.target.id);
            } else if (link.target.id === node.id) {
                connectedNodeIds.add(link.source.id);
            }
        });

        // Dim non-connected elements
        this.nodesGroup.selectAll('.node')
            .style('opacity', d => d === node || connectedNodeIds.has(d.id) ? 1 : 0.3);

        this.linksGroup.selectAll('.link')
            .style('opacity', d => d.source === node || d.target === node ? 1 : 0.1);

        // Show tooltip
        this.showTooltip(event, node);
    }

    handleNodeOut(event, node) {
        // Reset opacity
        this.nodesGroup.selectAll('.node').style('opacity', 1);
        this.linksGroup.selectAll('.link').style('opacity', 0.6);
        
        // Hide tooltip
        this.hideTooltip();
    }

    showTooltip(event, node) {
        const tooltip = this.getOrCreateTooltip();
        
        tooltip.innerHTML = `
            <div class="tooltip-header">
                <strong>${node.name || node.id.substring(0, 12)}</strong>
                <span class="tooltip-type">${node.type}</span>
            </div>
            <div class="tooltip-body">
                <div>Connections: ${node.connections}</div>
                <div>Distance: ${node.distance}</div>
                ${node.is_focus ? '<div class="focus-indicator">Focus Entity</div>' : ''}
            </div>
        `;

        tooltip.style.display = 'block';
        tooltip.style.left = (event.pageX + 10) + 'px';
        tooltip.style.top = (event.pageY - 10) + 'px';
    }

    hideTooltip() {
        const tooltip = this.container.querySelector('.network-tooltip');
        if (tooltip) {
            tooltip.style.display = 'none';
        }
    }

    getOrCreateTooltip() {
        let tooltip = this.container.querySelector('.network-tooltip');
        if (!tooltip) {
            tooltip = this.createElement('div', 'network-tooltip');
            this.container.appendChild(tooltip);
        }
        return tooltip;
    }

    // View control methods
    centerNetwork() {
        if (this.options.zoom && this.state.nodes.length > 0) {
            const centerX = this.options.width / 2;
            const centerY = this.options.height / 2;
            
            const transform = d3.zoomIdentity.translate(centerX, centerY);
            this.svg.transition().duration(750).call(this.zoom.transform, transform);
        }
    }

    fitToView() {
        if (this.options.zoom && this.state.nodes.length > 0) {
            const bounds = this.getNetworkBounds();
            const width = bounds.right - bounds.left;
            const height = bounds.bottom - bounds.top;
            const midX = (bounds.left + bounds.right) / 2;
            const midY = (bounds.top + bounds.bottom) / 2;
            
            const scale = Math.min(
                this.options.width / width,
                this.options.height / height
            ) * 0.8; // Add some padding
            
            const transform = d3.zoomIdentity
                .translate(this.options.width / 2, this.options.height / 2)
                .scale(scale)
                .translate(-midX, -midY);
            
            this.svg.transition().duration(750).call(this.zoom.transform, transform);
        }
    }

    resetZoom() {
        if (this.options.zoom) {
            this.svg.transition().duration(750).call(this.zoom.transform, d3.zoomIdentity);
        }
    }

    getNetworkBounds() {
        const padding = 50;
        return {
            left: Math.min(...this.state.nodes.map(d => d.x)) - padding,
            right: Math.max(...this.state.nodes.map(d => d.x)) + padding,
            top: Math.min(...this.state.nodes.map(d => d.y)) - padding,
            bottom: Math.max(...this.state.nodes.map(d => d.y)) + padding
        };
    }

    updateNetworkInfo() {
        const nodeCount = this.find('.node-count');
        const linkCount = this.find('.link-count');
        
        if (nodeCount) nodeCount.textContent = `${this.state.nodes.length} nodes`;
        if (linkCount) linkCount.textContent = `${this.state.links.length} links`;
    }

    onStateChange(oldState, newState) {
        if (oldState.loading !== newState.loading) {
            this.container.classList.toggle('loading', newState.loading);
        }

        if (oldState.depth !== newState.depth) {
            const depthSelector = this.find('.depth-selector');
            if (depthSelector) {
                depthSelector.value = newState.depth;
            }
        }
    }

    resize(width, height) {
        this.options.width = width;
        this.options.height = height;
        
        this.svg
            .attr('width', width)
            .attr('height', height);
        
        this.simulation
            .force('center', d3.forceCenter(width / 2, height / 2))
            .alpha(0.3)
            .restart();
    }

    destroy() {
        if (this.simulation) {
            this.simulation.stop();
        }
        super.destroy();
    }
}

// Export component
window.RelationshipNetwork = RelationshipNetwork;