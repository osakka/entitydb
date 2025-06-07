// Enhanced chart functionality for EntityDB metrics

class MetricsChartManager {
    constructor() {
        this.charts = {};
        this.timeRanges = {
            '1h': { hours: 1, label: 'Last Hour', interval: '1min' },
            '6h': { hours: 6, label: 'Last 6 Hours', interval: '1min' },
            '24h': { hours: 24, label: 'Last 24 Hours', interval: '1hour' },
            '7d': { hours: 168, label: 'Last 7 Days', interval: '1hour' },
            '30d': { hours: 720, label: 'Last 30 Days', interval: '1day' }
        };
        this.chartTypes = {
            'line': { label: 'Line Chart', icon: 'fa-chart-line' },
            'bar': { label: 'Bar Chart', icon: 'fa-chart-bar' },
            'area': { label: 'Area Chart', icon: 'fa-chart-area' }
        };
        this.currentTimeRange = '6h';
        this.currentChartType = 'line';
    }

    // Initialize chart controls
    initializeControls(containerId) {
        const container = document.getElementById(containerId);
        if (!container) return;

        // Create time range selector
        const timeRangeHtml = `
            <div class="chart-controls" style="display: flex; gap: 12px; margin-bottom: 16px; align-items: center;">
                <div class="time-range-selector">
                    <label style="font-size: 14px; color: #6b7280; margin-right: 8px;">Time Range:</label>
                    <select id="${containerId}-timerange" class="form-select" style="padding: 6px 12px; border: 1px solid #e5e7eb; border-radius: 6px;">
                        ${Object.entries(this.timeRanges).map(([key, value]) => 
                            `<option value="${key}" ${key === this.currentTimeRange ? 'selected' : ''}>${value.label}</option>`
                        ).join('')}
                    </select>
                </div>
                <div class="chart-type-selector">
                    <label style="font-size: 14px; color: #6b7280; margin-right: 8px;">Chart Type:</label>
                    <div class="btn-group" role="group">
                        ${Object.entries(this.chartTypes).map(([key, value]) => 
                            `<button type="button" class="btn btn-sm ${key === this.currentChartType ? 'btn-primary' : 'btn-outline'}" 
                                     data-chart-type="${key}" title="${value.label}">
                                <i class="fas ${value.icon}"></i>
                            </button>`
                        ).join('')}
                    </div>
                </div>
                <div class="chart-actions" style="margin-left: auto;">
                    <button class="btn btn-sm btn-outline" id="${containerId}-refresh" title="Refresh">
                        <i class="fas fa-sync-alt"></i>
                    </button>
                    <button class="btn btn-sm btn-outline" id="${containerId}-download" title="Download">
                        <i class="fas fa-download"></i>
                    </button>
                </div>
            </div>
        `;

        // Insert controls before the chart
        container.insertAdjacentHTML('afterbegin', timeRangeHtml);

        // Add event listeners
        this.attachEventListeners(containerId);
    }

    // Attach event listeners to controls
    attachEventListeners(containerId) {
        // Time range change
        const timeRangeSelect = document.getElementById(`${containerId}-timerange`);
        if (timeRangeSelect) {
            timeRangeSelect.addEventListener('change', (e) => {
                this.currentTimeRange = e.target.value;
                this.refreshChart(containerId);
            });
        }

        // Chart type change
        const chartTypeButtons = document.querySelectorAll(`#${containerId} .btn-group button`);
        chartTypeButtons.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const chartType = e.currentTarget.getAttribute('data-chart-type');
                this.currentChartType = chartType;
                
                // Update button states
                chartTypeButtons.forEach(b => b.classList.remove('btn-primary'));
                chartTypeButtons.forEach(b => b.classList.add('btn-outline'));
                e.currentTarget.classList.remove('btn-outline');
                e.currentTarget.classList.add('btn-primary');
                
                this.refreshChart(containerId);
            });
        });

        // Refresh button
        const refreshBtn = document.getElementById(`${containerId}-refresh`);
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.refreshChart(containerId);
            });
        }

        // Download button
        const downloadBtn = document.getElementById(`${containerId}-download`);
        if (downloadBtn) {
            downloadBtn.addEventListener('click', () => {
                this.downloadChart(containerId);
            });
        }
    }

    // Create an enhanced chart with real-time data
    async createEnhancedChart(canvasId, metricNames, options = {}) {
        const canvas = document.getElementById(canvasId);
        if (!canvas) return;

        // Destroy existing chart
        if (this.charts[canvasId]) {
            this.charts[canvasId].destroy();
        }

        // Fetch data for all metrics
        const datasets = [];
        const timeRange = this.timeRanges[this.currentTimeRange];
        
        for (let i = 0; i < metricNames.length; i++) {
            const metricName = metricNames[i];
            const data = await this.fetchMetricData(metricName, timeRange);
            
            if (data && data.data_points && data.data_points.length > 0) {
                datasets.push({
                    label: options.labels?.[i] || metricName,
                    data: data.data_points.map(p => ({
                        x: new Date(p.timestamp),
                        y: p.value
                    })),
                    borderColor: options.colors?.[i] || this.getColor(i),
                    backgroundColor: this.currentChartType === 'area' ? 
                        this.getColor(i, 0.1) : 'transparent',
                    fill: this.currentChartType === 'area',
                    tension: 0.1
                });
            }
        }

        // Create chart configuration
        const config = {
            type: this.currentChartType === 'area' ? 'line' : this.currentChartType,
            data: { datasets },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                interaction: {
                    mode: 'index',
                    intersect: false
                },
                plugins: {
                    title: {
                        display: options.title ? true : false,
                        text: options.title
                    },
                    legend: {
                        display: true,
                        position: 'top'
                    },
                    tooltip: {
                        callbacks: {
                            label: (context) => {
                                let label = context.dataset.label || '';
                                if (label) {
                                    label += ': ';
                                }
                                const value = context.parsed.y;
                                if (options.formatValue) {
                                    label += options.formatValue(value);
                                } else {
                                    label += value.toFixed(2);
                                }
                                return label;
                            }
                        }
                    }
                },
                scales: {
                    x: {
                        type: 'time',
                        time: {
                            unit: this.getTimeUnit(timeRange.hours),
                            displayFormats: {
                                minute: 'HH:mm',
                                hour: 'MMM D, HH:mm',
                                day: 'MMM D'
                            }
                        },
                        title: {
                            display: true,
                            text: 'Time'
                        }
                    },
                    y: {
                        beginAtZero: options.beginAtZero !== false,
                        title: {
                            display: true,
                            text: options.yAxisLabel || 'Value'
                        },
                        ticks: {
                            callback: options.formatValue || ((value) => value)
                        }
                    }
                }
            }
        };

        // Create the chart
        this.charts[canvasId] = new Chart(canvas.getContext('2d'), config);
        return this.charts[canvasId];
    }

    // Fetch metric data from the API
    async fetchMetricData(metricName, timeRange) {
        try {
            const aggregation = timeRange.interval === 'raw' ? 'raw' : timeRange.interval.replace('1', '');
            const response = await fetch(`/api/v1/metrics/history?metric_name=${metricName}&hours=${timeRange.hours}&aggregation=${aggregation}`);
            
            if (!response.ok) {
                console.error(`Failed to fetch metric ${metricName}: ${response.status}`);
                return null;
            }
            
            return await response.json();
        } catch (error) {
            console.error(`Error fetching metric ${metricName}:`, error);
            return null;
        }
    }

    // Refresh chart with current settings
    async refreshChart(containerId) {
        const chartConfig = this.getChartConfig(containerId);
        if (chartConfig) {
            await this.createEnhancedChart(
                chartConfig.canvasId,
                chartConfig.metrics,
                chartConfig.options
            );
        }
    }

    // Download chart as image
    downloadChart(containerId) {
        const canvasId = this.getCanvasId(containerId);
        const chart = this.charts[canvasId];
        if (chart) {
            const url = chart.toBase64Image();
            const link = document.createElement('a');
            link.download = `${canvasId}-${new Date().toISOString()}.png`;
            link.href = url;
            link.click();
        }
    }

    // Helper: Get time unit for Chart.js
    getTimeUnit(hours) {
        if (hours <= 1) return 'minute';
        if (hours <= 24) return 'hour';
        return 'day';
    }

    // Helper: Get color
    getColor(index, alpha = 1) {
        const colors = [
            `rgba(59, 130, 246, ${alpha})`,   // Blue
            `rgba(16, 185, 129, ${alpha})`,   // Green
            `rgba(245, 158, 11, ${alpha})`,   // Orange
            `rgba(239, 68, 68, ${alpha})`,    // Red
            `rgba(139, 92, 246, ${alpha})`,   // Purple
            `rgba(236, 72, 153, ${alpha})`,   // Pink
            `rgba(14, 165, 233, ${alpha})`,   // Sky
            `rgba(168, 85, 247, ${alpha})`    // Violet
        ];
        return colors[index % colors.length];
    }

    // Store chart configurations for refresh
    chartConfigs = {};

    setChartConfig(containerId, canvasId, metrics, options) {
        this.chartConfigs[containerId] = { canvasId, metrics, options };
    }

    getChartConfig(containerId) {
        return this.chartConfigs[containerId];
    }

    getCanvasId(containerId) {
        return this.chartConfigs[containerId]?.canvasId;
    }
}

// Create global instance
window.metricsChartManager = new MetricsChartManager();

// Helper function to format bytes
function formatBytes(bytes, decimals = 2) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

// Helper function to format duration
function formatDuration(ms) {
    if (ms < 1) return ms.toFixed(3) + ' ms';
    if (ms < 1000) return ms.toFixed(1) + ' ms';
    if (ms < 60000) return (ms / 1000).toFixed(1) + ' s';
    return (ms / 60000).toFixed(1) + ' min';
}