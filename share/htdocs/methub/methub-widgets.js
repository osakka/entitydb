// MetHub Widget System - Pluggable widget renderers

class MetHubWidgets {
    constructor() {
        this.charts = new Map();
        this.updateCallbacks = new Map();
    }

    // Register a chart instance
    registerChart(widgetId, chart) {
        this.charts.set(widgetId, chart);
    }

    // Get or create chart
    getChart(widgetId) {
        return this.charts.get(widgetId);
    }

    // Destroy chart
    destroyChart(widgetId) {
        const chart = this.charts.get(widgetId);
        if (chart) {
            chart.destroy();
            this.charts.delete(widgetId);
        }
    }

    // Render widget based on type
    async renderWidget(widget, metrics, container) {
        switch (widget.type) {
            case 'gauge':
                return this.renderGauge(widget, metrics, container);
            case 'line':
                return this.renderLineChart(widget, metrics, container);
            case 'bar':
                return this.renderBarChart(widget, metrics, container);
            case 'value':
                return this.renderSingleValue(widget, metrics, container);
            case 'table':
                return this.renderTable(widget, metrics, container);
            case 'heatmap':
                return this.renderHeatmap(widget, metrics, container);
            default:
                return '<div class="error">Unknown widget type</div>';
        }
    }

    // Gauge widget
    renderGauge(widget, metrics, container) {
        const latest = metrics[0];
        if (!latest) return '<div class="no-data">No data</div>';

        const value = latest.value;
        const max = widget.max || 100;
        const percentage = (value / max) * 100;
        
        // Determine color based on thresholds
        let color = '#4ade80'; // green
        if (widget.thresholds) {
            if (value >= widget.thresholds.critical) color = '#ef4444'; // red
            else if (value >= widget.thresholds.warning) color = '#f59e0b'; // yellow
        }

        const html = `
            <div class="gauge-widget">
                <div class="gauge-container">
                    <svg viewBox="0 0 200 100" class="gauge-svg">
                        <path d="M 30 90 A 70 70 0 0 1 170 90" 
                              fill="none" 
                              stroke="#e5e7eb" 
                              stroke-width="15" />
                        <path d="M 30 90 A 70 70 0 0 1 170 90" 
                              fill="none" 
                              stroke="${color}" 
                              stroke-width="15"
                              stroke-dasharray="${percentage * 2.2} 220"
                              class="gauge-fill" />
                    </svg>
                    <div class="gauge-value">
                        <span class="value">${value.toFixed(1)}</span>
                        <span class="unit">${latest.unit || '%'}</span>
                    </div>
                </div>
                <div class="gauge-label">${latest.name}</div>
            </div>
        `;

        if (container) container.innerHTML = html;
        return html;
    }

    // Line chart widget
    renderLineChart(widget, metrics, container) {
        if (!container) return '';
        
        // Aggregate metrics
        const aggregated = widget.api.aggregateMetrics(metrics, widget.interval || '1m');
        
        // Prepare chart data
        const chartData = {
            labels: aggregated.map(m => new Date(m.timestamp / 1000000).toLocaleTimeString()),
            datasets: [{
                label: widget.metricName,
                data: aggregated.map(m => m.value),
                borderColor: '#3b82f6',
                backgroundColor: 'rgba(59, 130, 246, 0.1)',
                tension: 0.4,
                fill: true
            }]
        };

        // Create canvas
        const canvasId = `chart-${widget.id}`;
        container.innerHTML = `<canvas id="${canvasId}"></canvas>`;
        
        // Destroy existing chart
        this.destroyChart(widget.id);
        
        // Create new chart
        const ctx = document.getElementById(canvasId).getContext('2d');
        const chart = new Chart(ctx, {
            type: 'line',
            data: chartData,
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            callback: function(value) {
                                return value + (metrics[0]?.unit || '');
                            }
                        }
                    },
                    x: {
                        ticks: {
                            maxTicksLimit: 6
                        }
                    }
                }
            }
        });
        
        this.registerChart(widget.id, chart);
        return '';
    }

    // Bar chart widget
    renderBarChart(widget, metrics, container) {
        if (!container) return '';
        
        // Group by a dimension (e.g., device for disk metrics)
        const groups = {};
        metrics.forEach(m => {
            const key = m.device || m.mount || m.host || 'default';
            if (!groups[key] || m.timestamp > groups[key].timestamp) {
                groups[key] = m;
            }
        });
        
        const labels = Object.keys(groups);
        const data = labels.map(k => groups[k].value);
        
        // Create canvas
        const canvasId = `chart-${widget.id}`;
        container.innerHTML = `<canvas id="${canvasId}"></canvas>`;
        
        // Destroy existing chart
        this.destroyChart(widget.id);
        
        // Create new chart
        const ctx = document.getElementById(canvasId).getContext('2d');
        const chart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [{
                    label: widget.metricName,
                    data: data,
                    backgroundColor: '#3b82f6'
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            callback: function(value) {
                                return value + (metrics[0]?.unit || '');
                            }
                        }
                    }
                }
            }
        });
        
        this.registerChart(widget.id, chart);
        return '';
    }

    // Single value widget
    renderSingleValue(widget, metrics, container) {
        const latest = metrics[0];
        if (!latest) return '<div class="no-data">No data</div>';

        const value = latest.value;
        const unit = latest.unit || '';
        
        // Calculate trend if we have historical data
        let trend = '';
        if (metrics.length > 1) {
            const previous = metrics[1].value;
            const change = ((value - previous) / previous) * 100;
            const trendIcon = change > 0 ? 'fa-arrow-up' : 'fa-arrow-down';
            const trendColor = change > 0 ? '#ef4444' : '#4ade80';
            trend = `
                <div class="value-trend" style="color: ${trendColor}">
                    <i class="fas ${trendIcon}"></i>
                    ${Math.abs(change).toFixed(1)}%
                </div>
            `;
        }

        const html = `
            <div class="value-widget">
                <div class="value-main">
                    <span class="value">${value.toFixed(2)}</span>
                    <span class="unit">${unit}</span>
                </div>
                ${trend}
                <div class="value-label">${latest.name}</div>
                <div class="value-time">${new Date(latest.timestamp / 1000000).toLocaleTimeString()}</div>
            </div>
        `;

        if (container) container.innerHTML = html;
        return html;
    }

    // Table widget
    renderTable(widget, metrics, container) {
        if (metrics.length === 0) return '<div class="no-data">No data</div>';

        // Get latest value for each unique metric
        const latestMetrics = {};
        metrics.forEach(m => {
            const key = `${m.host}-${m.name}-${m.device || m.mount || ''}`;
            if (!latestMetrics[key] || m.timestamp > latestMetrics[key].timestamp) {
                latestMetrics[key] = m;
            }
        });

        const rows = Object.values(latestMetrics).map(m => `
            <tr>
                <td>${m.host}</td>
                <td>${m.name}</td>
                <td>${m.value.toFixed(2)} ${m.unit || ''}</td>
                <td>${new Date(m.timestamp / 1000000).toLocaleTimeString()}</td>
            </tr>
        `).join('');

        const html = `
            <div class="table-widget">
                <table>
                    <thead>
                        <tr>
                            <th>Host</th>
                            <th>Metric</th>
                            <th>Value</th>
                            <th>Time</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${rows}
                    </tbody>
                </table>
            </div>
        `;

        if (container) container.innerHTML = html;
        return html;
    }

    // Heatmap widget
    renderHeatmap(widget, metrics, container) {
        // Group metrics by host and time buckets
        const hosts = [...new Set(metrics.map(m => m.host))];
        const bucketSize = 60000000000; // 1 minute in nanoseconds
        
        const data = {};
        metrics.forEach(m => {
            const bucket = Math.floor(m.timestamp / bucketSize) * bucketSize;
            const key = `${m.host}-${bucket}`;
            data[key] = m.value;
        });

        // Find time range
        const timestamps = metrics.map(m => m.timestamp);
        const minTime = Math.min(...timestamps);
        const maxTime = Math.max(...timestamps);
        const buckets = [];
        
        for (let t = minTime; t <= maxTime; t += bucketSize) {
            buckets.push(t);
        }

        // Generate heatmap cells
        const cells = hosts.map(host => {
            return buckets.map(bucket => {
                const value = data[`${host}-${bucket}`];
                if (value === undefined) return '<td class="no-data"></td>';
                
                // Color based on value (customize based on metric)
                const intensity = Math.min(value / (widget.max || 100), 1);
                const color = `rgba(239, 68, 68, ${intensity})`;
                
                return `<td style="background: ${color}" title="${host}: ${value}"></td>`;
            }).join('');
        });

        const html = `
            <div class="heatmap-widget">
                <table>
                    <tbody>
                        ${hosts.map((host, i) => `
                            <tr>
                                <th>${host}</th>
                                ${cells[i]}
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            </div>
        `;

        if (container) container.innerHTML = html;
        return html;
    }
}