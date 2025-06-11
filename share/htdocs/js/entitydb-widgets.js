/**
 * EntityDB Widget Library
 * A comprehensive collection of monitoring widgets for EntityDB
 * Compatible with Vue 3 and ApexCharts
 */

(function (window) {
    'use strict';

    // Widget configuration defaults
    const WIDGET_DEFAULTS = {
        refreshInterval: 30000, // 30 seconds
        chartHeight: 300,
        colors: {
            primary: '#3498db',
            success: '#10b981',
            warning: '#f59e0b',
            danger: '#ef4444',
            info: '#3b82f6',
            purple: '#9b59b6',
            dark: '#2c3e50'
        },
        darkModeColors: {
            primary: '#5DADE2',
            success: '#48C9B0',
            warning: '#F8C471',
            danger: '#EC7063',
            info: '#5DADE2',
            purple: '#BB8FCE',
            dark: '#34495E'
        }
    };

    // Time range options
    const TIME_RANGES = [
        { label: 'Last 5 minutes', value: '5m', minutes: 5 },
        { label: 'Last 15 minutes', value: '15m', minutes: 15 },
        { label: 'Last 1 hour', value: '1h', minutes: 60 },
        { label: 'Last 6 hours', value: '6h', minutes: 360 },
        { label: 'Last 24 hours', value: '24h', minutes: 1440 },
        { label: 'Last 7 days', value: '7d', minutes: 10080 },
        { label: 'Last 30 days', value: '30d', minutes: 43200 }
    ];

    // Base widget class
    class EntityDBWidget {
        constructor(options) {
            this.id = options.id || `widget-${Date.now()}`;
            this.type = options.type;
            this.title = options.title;
            this.refreshInterval = options.refreshInterval || WIDGET_DEFAULTS.refreshInterval;
            this.timeRange = options.timeRange || '1h';
            this.isDarkMode = options.isDarkMode || false;
            this.apiEndpoint = options.apiEndpoint;
            this.chart = null;
            this.refreshTimer = null;
            this.isLoading = false;
            this.error = null;
            this.data = null;
        }

        get colors() {
            return this.isDarkMode ? WIDGET_DEFAULTS.darkModeColors : WIDGET_DEFAULTS.colors;
        }

        async fetchData() {
            this.isLoading = true;
            this.error = null;
            
            try {
                const headers = {};
                const token = localStorage.getItem('entitydb-admin-token');
                if (token) {
                    headers['Authorization'] = `Bearer ${token}`;
                }

                const response = await fetch(this.apiEndpoint, { headers });
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }
                
                this.data = await response.json();
                this.onDataUpdate(this.data);
            } catch (error) {
                this.error = error.message;
                this.onError(error);
            } finally {
                this.isLoading = false;
            }
        }

        onDataUpdate(data) {
            // Override in subclasses
        }

        onError(error) {
            console.error(`Widget ${this.id} error:`, error);
        }

        start() {
            this.fetchData();
            if (this.refreshInterval > 0) {
                this.refreshTimer = setInterval(() => this.fetchData(), this.refreshInterval);
            }
        }

        stop() {
            if (this.refreshTimer) {
                clearInterval(this.refreshTimer);
                this.refreshTimer = null;
            }
        }

        destroy() {
            this.stop();
            if (this.chart) {
                this.chart.destroy();
                this.chart = null;
            }
        }

        exportData() {
            const dataStr = JSON.stringify(this.data, null, 2);
            const dataBlob = new Blob([dataStr], { type: 'application/json' });
            const url = URL.createObjectURL(dataBlob);
            const link = document.createElement('a');
            link.href = url;
            link.download = `${this.type}-${new Date().toISOString()}.json`;
            link.click();
            URL.revokeObjectURL(url);
        }
    }

    // Query Performance Widget
    class QueryPerformanceWidget extends EntityDBWidget {
        constructor(options) {
            super({
                ...options,
                type: 'query-performance',
                title: options.title || 'Query Performance',
                apiEndpoint: `/api/v1/metrics/history/v2?metric_name=query_duration_ms&hours=${options.hours || 1}`
            });
        }

        onDataUpdate(data) {
            const chartData = this.processQueryData(data);
            this.updateChart(chartData);
        }

        processQueryData(data) {
            if (!data.data_points || data.data_points.length === 0) {
                return { categories: [], series: [] };
            }

            const points = data.data_points.sort((a, b) => 
                new Date(a.timestamp) - new Date(b.timestamp)
            );

            return {
                categories: points.map(p => new Date(p.timestamp)),
                series: [{
                    name: 'Query Duration (ms)',
                    data: points.map(p => p.value)
                }]
            };
        }

        updateChart(chartData) {
            const options = {
                series: chartData.series,
                chart: {
                    type: 'line',
                    height: WIDGET_DEFAULTS.chartHeight,
                    toolbar: { show: true },
                    background: 'transparent',
                    animations: {
                        enabled: true,
                        easing: 'easeinout',
                        speed: 800
                    }
                },
                stroke: {
                    curve: 'smooth',
                    width: 3
                },
                xaxis: {
                    type: 'datetime',
                    categories: chartData.categories
                },
                yaxis: {
                    title: {
                        text: 'Duration (ms)'
                    }
                },
                colors: [this.colors.primary],
                theme: {
                    mode: this.isDarkMode ? 'dark' : 'light'
                },
                tooltip: {
                    x: {
                        format: 'dd MMM HH:mm'
                    }
                }
            };

            if (this.chart) {
                this.chart.updateOptions(options);
            } else {
                this.chart = new ApexCharts(document.querySelector(`#${this.id}`), options);
                this.chart.render();
            }
        }
    }

    // Storage Metrics Widget
    class StorageMetricsWidget extends EntityDBWidget {
        constructor(options) {
            super({
                ...options,
                type: 'storage-metrics',
                title: options.title || 'Storage Performance',
                apiEndpoint: '/api/v1/system/metrics'
            });
        }

        onDataUpdate(data) {
            const metrics = data.storage_metrics || {};
            this.updateGauges(metrics);
        }

        updateGauges(metrics) {
            const readLatency = metrics.read_latency_ms || 0;
            const writeLatency = metrics.write_latency_ms || 0;
            const walSize = (metrics.wal_size_mb || 0).toFixed(2);
            const compressionRatio = (metrics.compression_ratio || 1).toFixed(2);

            const options = {
                series: [
                    {
                        name: 'Read Latency',
                        data: [readLatency]
                    },
                    {
                        name: 'Write Latency',
                        data: [writeLatency]
                    }
                ],
                chart: {
                    type: 'bar',
                    height: WIDGET_DEFAULTS.chartHeight,
                    toolbar: { show: false },
                    background: 'transparent'
                },
                plotOptions: {
                    bar: {
                        horizontal: true,
                        distributed: true,
                        dataLabels: {
                            position: 'top'
                        }
                    }
                },
                dataLabels: {
                    enabled: true,
                    formatter: (val) => `${val}ms`,
                    style: {
                        fontSize: '12px'
                    }
                },
                colors: [this.colors.success, this.colors.warning],
                xaxis: {
                    categories: ['Read', 'Write'],
                    max: Math.max(readLatency, writeLatency) * 1.5
                },
                yaxis: {
                    labels: {
                        style: {
                            fontSize: '14px'
                        }
                    }
                },
                theme: {
                    mode: this.isDarkMode ? 'dark' : 'light'
                },
                annotations: {
                    texts: [
                        {
                            x: 10,
                            y: 10,
                            text: `WAL: ${walSize}MB | Compression: ${compressionRatio}x`,
                            textAnchor: 'start',
                            fontSize: '12px',
                            foreColor: this.isDarkMode ? '#e1e8ed' : '#6c757d'
                        }
                    ]
                }
            };

            if (this.chart) {
                this.chart.updateOptions(options);
            } else {
                this.chart = new ApexCharts(document.querySelector(`#${this.id}`), options);
                this.chart.render();
            }
        }
    }

    // Memory Analytics Widget
    class MemoryAnalyticsWidget extends EntityDBWidget {
        constructor(options) {
            super({
                ...options,
                type: 'memory-analytics',
                title: options.title || 'Memory Analytics',
                apiEndpoint: '/api/v1/system/metrics'
            });
        }

        onDataUpdate(data) {
            const memory = data.memory_metrics || {};
            this.updateMemoryChart(memory);
        }

        updateMemoryChart(memory) {
            const alloc = (memory.alloc_mb || 0).toFixed(2);
            const sys = (memory.sys_mb || 0).toFixed(2);
            const gcPause = (memory.gc_pause_ms || 0).toFixed(2);
            const heapInUse = (memory.heap_inuse_mb || 0).toFixed(2);

            const options = {
                series: [{
                    name: 'Memory Usage',
                    data: [
                        { x: 'Allocated', y: parseFloat(alloc) },
                        { x: 'System', y: parseFloat(sys) },
                        { x: 'Heap In Use', y: parseFloat(heapInUse) }
                    ]
                }],
                chart: {
                    type: 'treemap',
                    height: WIDGET_DEFAULTS.chartHeight,
                    toolbar: { show: false },
                    background: 'transparent'
                },
                colors: [this.colors.primary, this.colors.info, this.colors.purple],
                plotOptions: {
                    treemap: {
                        distributed: true,
                        enableShades: true
                    }
                },
                dataLabels: {
                    enabled: true,
                    formatter: (text, op) => {
                        return `${text}: ${op.value}MB`;
                    }
                },
                theme: {
                    mode: this.isDarkMode ? 'dark' : 'light'
                },
                title: {
                    text: `GC Pause: ${gcPause}ms`,
                    align: 'right',
                    style: {
                        fontSize: '12px',
                        color: this.isDarkMode ? '#e1e8ed' : '#6c757d'
                    }
                }
            };

            if (this.chart) {
                this.chart.updateOptions(options);
            } else {
                this.chart = new ApexCharts(document.querySelector(`#${this.id}`), options);
                this.chart.render();
            }
        }
    }

    // Session Metrics Widget
    class SessionMetricsWidget extends EntityDBWidget {
        constructor(options) {
            super({
                ...options,
                type: 'session-metrics',
                title: options.title || 'Session Analytics',
                apiEndpoint: '/api/v1/rbac/metrics'
            });
        }

        onDataUpdate(data) {
            const sessions = data.session_metrics || {};
            this.updateSessionChart(sessions);
        }

        updateSessionChart(sessions) {
            const activeSessions = sessions.active_sessions || 0;
            const totalLogins = sessions.total_logins || 0;
            const failedLogins = sessions.failed_logins || 0;
            const successRate = totalLogins > 0 
                ? ((totalLogins - failedLogins) / totalLogins * 100).toFixed(1)
                : 100;

            const options = {
                series: [parseFloat(successRate)],
                chart: {
                    type: 'radialBar',
                    height: WIDGET_DEFAULTS.chartHeight,
                    toolbar: { show: false },
                    background: 'transparent'
                },
                plotOptions: {
                    radialBar: {
                        hollow: {
                            size: '70%'
                        },
                        dataLabels: {
                            name: {
                                offsetY: -10,
                                color: this.isDarkMode ? '#e1e8ed' : '#374151',
                                fontSize: '13px'
                            },
                            value: {
                                color: this.colors.success,
                                fontSize: '30px',
                                show: true,
                                formatter: (val) => `${val}%`
                            }
                        }
                    }
                },
                fill: {
                    type: 'gradient',
                    gradient: {
                        shade: 'dark',
                        type: 'horizontal',
                        shadeIntensity: 0.5,
                        gradientToColors: [this.colors.success],
                        inverseColors: true,
                        opacityFrom: 1,
                        opacityTo: 1,
                        stops: [0, 100]
                    }
                },
                stroke: {
                    lineCap: 'round'
                },
                labels: ['Success Rate'],
                annotations: {
                    texts: [
                        {
                            x: '50%',
                            y: '75%',
                            text: `${activeSessions} Active Sessions`,
                            textAnchor: 'middle',
                            fontSize: '14px',
                            foreColor: this.isDarkMode ? '#e1e8ed' : '#6c757d'
                        }
                    ]
                }
            };

            if (this.chart) {
                this.chart.updateOptions(options);
            } else {
                this.chart = new ApexCharts(document.querySelector(`#${this.id}`), options);
                this.chart.render();
            }
        }
    }

    // Error Tracking Widget
    class ErrorTrackingWidget extends EntityDBWidget {
        constructor(options) {
            super({
                ...options,
                type: 'error-tracking',
                title: options.title || 'Error Analytics',
                apiEndpoint: `/api/v1/metrics/history/v2?metric_name=error_count&hours=${options.hours || 24}`
            });
        }

        onDataUpdate(data) {
            const chartData = this.processErrorData(data);
            this.updateErrorChart(chartData);
        }

        processErrorData(data) {
            if (!data.data_points || data.data_points.length === 0) {
                return { categories: [], series: [] };
            }

            // Group errors by hour
            const hourlyErrors = {};
            data.data_points.forEach(point => {
                const hour = new Date(point.timestamp);
                hour.setMinutes(0, 0, 0);
                const key = hour.toISOString();
                hourlyErrors[key] = (hourlyErrors[key] || 0) + point.value;
            });

            const sortedHours = Object.keys(hourlyErrors).sort();
            
            return {
                categories: sortedHours.map(h => new Date(h)),
                series: [{
                    name: 'Errors',
                    data: sortedHours.map(h => hourlyErrors[h])
                }]
            };
        }

        updateErrorChart(chartData) {
            const options = {
                series: chartData.series,
                chart: {
                    type: 'area',
                    height: WIDGET_DEFAULTS.chartHeight,
                    toolbar: { show: true },
                    background: 'transparent',
                    zoom: {
                        enabled: true
                    }
                },
                dataLabels: {
                    enabled: false
                },
                stroke: {
                    curve: 'smooth',
                    width: 2
                },
                fill: {
                    type: 'gradient',
                    gradient: {
                        shadeIntensity: 1,
                        opacityFrom: 0.7,
                        opacityTo: 0.3,
                        stops: [0, 90, 100]
                    }
                },
                xaxis: {
                    type: 'datetime',
                    categories: chartData.categories
                },
                yaxis: {
                    title: {
                        text: 'Error Count'
                    },
                    min: 0
                },
                colors: [this.colors.danger],
                theme: {
                    mode: this.isDarkMode ? 'dark' : 'light'
                },
                tooltip: {
                    x: {
                        format: 'dd MMM HH:mm'
                    }
                }
            };

            if (this.chart) {
                this.chart.updateOptions(options);
            } else {
                this.chart = new ApexCharts(document.querySelector(`#${this.id}`), options);
                this.chart.render();
            }
        }
    }

    // Temporal Metrics Widget
    class TemporalMetricsWidget extends EntityDBWidget {
        constructor(options) {
            super({
                ...options,
                type: 'temporal-metrics',
                title: options.title || 'Temporal Operations',
                apiEndpoint: '/api/v1/system/metrics'
            });
        }

        onDataUpdate(data) {
            const temporal = data.temporal_metrics || {};
            this.updateTemporalChart(temporal);
        }

        updateTemporalChart(temporal) {
            const asOfQueries = temporal.as_of_queries || 0;
            const historyQueries = temporal.history_queries || 0;
            const diffQueries = temporal.diff_queries || 0;
            const changesQueries = temporal.changes_queries || 0;

            const options = {
                series: [{
                    data: [asOfQueries, historyQueries, diffQueries, changesQueries]
                }],
                chart: {
                    type: 'bar',
                    height: WIDGET_DEFAULTS.chartHeight,
                    toolbar: { show: false },
                    background: 'transparent'
                },
                plotOptions: {
                    bar: {
                        borderRadius: 8,
                        horizontal: true,
                        distributed: true
                    }
                },
                dataLabels: {
                    enabled: true
                },
                colors: [
                    this.colors.primary,
                    this.colors.success,
                    this.colors.warning,
                    this.colors.info
                ],
                xaxis: {
                    categories: ['As-Of', 'History', 'Diff', 'Changes']
                },
                theme: {
                    mode: this.isDarkMode ? 'dark' : 'light'
                }
            };

            if (this.chart) {
                this.chart.updateOptions(options);
            } else {
                this.chart = new ApexCharts(document.querySelector(`#${this.id}`), options);
                this.chart.render();
            }
        }
    }

    // Relationship Metrics Widget
    class RelationshipMetricsWidget extends EntityDBWidget {
        constructor(options) {
            super({
                ...options,
                type: 'relationship-metrics',
                title: options.title || 'Entity Relationships',
                apiEndpoint: '/api/v1/system/metrics'
            });
        }

        onDataUpdate(data) {
            const relationships = data.relationship_metrics || {};
            this.updateRelationshipChart(relationships);
        }

        updateRelationshipChart(relationships) {
            const totalRelationships = relationships.total_relationships || 0;
            const avgRelationsPerEntity = (relationships.avg_relations_per_entity || 0).toFixed(2);
            const queryLatency = (relationships.query_latency_ms || 0).toFixed(2);

            const options = {
                series: [
                    {
                        name: 'Metrics',
                        data: [
                            totalRelationships,
                            parseFloat(avgRelationsPerEntity) * 100, // Scale for visibility
                            parseFloat(queryLatency) * 10 // Scale for visibility
                        ]
                    }
                ],
                chart: {
                    type: 'radar',
                    height: WIDGET_DEFAULTS.chartHeight,
                    toolbar: { show: false },
                    background: 'transparent'
                },
                xaxis: {
                    categories: ['Total Relations', 'Avg per Entity (x100)', 'Query Latency (x10ms)']
                },
                colors: [this.colors.purple],
                theme: {
                    mode: this.isDarkMode ? 'dark' : 'light'
                },
                fill: {
                    opacity: 0.2
                },
                stroke: {
                    width: 2
                },
                markers: {
                    size: 4
                }
            };

            if (this.chart) {
                this.chart.updateOptions(options);
            } else {
                this.chart = new ApexCharts(document.querySelector(`#${this.id}`), options);
                this.chart.render();
            }
        }
    }

    // Tag Index Metrics Widget
    class TagIndexMetricsWidget extends EntityDBWidget {
        constructor(options) {
            super({
                ...options,
                type: 'tag-index-metrics',
                title: options.title || 'Tag Index Performance',
                apiEndpoint: '/api/v1/system/metrics'
            });
        }

        onDataUpdate(data) {
            const tagIndex = data.tag_index_metrics || {};
            this.updateTagIndexChart(tagIndex);
        }

        updateTagIndexChart(tagIndex) {
            const indexSize = tagIndex.index_size_mb || 0;
            const lookupLatency = tagIndex.lookup_latency_ms || 0;
            const rebuildTime = tagIndex.last_rebuild_time_ms || 0;
            const totalTags = tagIndex.total_tags || 0;

            const options = {
                series: [{
                    name: 'Performance',
                    data: [
                        { x: 'Index Size', y: indexSize, fillColor: this.colors.primary },
                        { x: 'Lookup Time', y: lookupLatency, fillColor: this.colors.success },
                        { x: 'Rebuild Time', y: rebuildTime / 1000, fillColor: this.colors.warning }, // Convert to seconds
                        { x: 'Total Tags', y: totalTags / 1000, fillColor: this.colors.info } // Scale down
                    ]
                }],
                chart: {
                    type: 'scatter',
                    height: WIDGET_DEFAULTS.chartHeight,
                    toolbar: { show: false },
                    background: 'transparent',
                    zoom: {
                        enabled: true,
                        type: 'xy'
                    }
                },
                xaxis: {
                    type: 'category',
                    tickPlacement: 'on'
                },
                yaxis: {
                    title: {
                        text: 'Value'
                    },
                    labels: {
                        formatter: (val) => {
                            if (val < 1) return `${(val * 1000).toFixed(0)}ms`;
                            if (val > 1000) return `${(val / 1000).toFixed(1)}k`;
                            return val.toFixed(1);
                        }
                    }
                },
                theme: {
                    mode: this.isDarkMode ? 'dark' : 'light'
                },
                markers: {
                    size: 20,
                    hover: {
                        size: 25
                    }
                },
                tooltip: {
                    custom: function({ series, seriesIndex, dataPointIndex, w }) {
                        const data = w.config.series[seriesIndex].data[dataPointIndex];
                        let value = data.y;
                        let unit = '';
                        
                        switch(data.x) {
                            case 'Index Size':
                                unit = 'MB';
                                break;
                            case 'Lookup Time':
                                unit = 'ms';
                                break;
                            case 'Rebuild Time':
                                unit = 's';
                                break;
                            case 'Total Tags':
                                value = value * 1000;
                                unit = '';
                                break;
                        }
                        
                        return `<div class="apexcharts-tooltip-custom">
                            <span>${data.x}: ${value.toFixed(2)}${unit}</span>
                        </div>`;
                    }
                }
            };

            if (this.chart) {
                this.chart.updateOptions(options);
            } else {
                this.chart = new ApexCharts(document.querySelector(`#${this.id}`), options);
                this.chart.render();
            }
        }
    }

    // Heatmap Widget for Time-based Patterns
    class ActivityHeatmapWidget extends EntityDBWidget {
        constructor(options) {
            super({
                ...options,
                type: 'activity-heatmap',
                title: options.title || 'Activity Heatmap',
                apiEndpoint: `/api/v1/metrics/history/v2?metric_name=request_count&hours=${options.hours || 168}` // 7 days
            });
        }

        onDataUpdate(data) {
            const heatmapData = this.processHeatmapData(data);
            this.updateHeatmap(heatmapData);
        }

        processHeatmapData(data) {
            if (!data.data_points || data.data_points.length === 0) {
                return { series: [] };
            }

            // Create 7x24 grid (days x hours)
            const grid = Array(7).fill(null).map(() => Array(24).fill(0));
            
            data.data_points.forEach(point => {
                const date = new Date(point.timestamp);
                const day = date.getDay();
                const hour = date.getHours();
                grid[day][hour] += point.value;
            });

            const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
            
            return {
                series: days.map((day, dayIndex) => ({
                    name: day,
                    data: grid[dayIndex].map((value, hour) => ({
                        x: `${hour}:00`,
                        y: value
                    }))
                }))
            };
        }

        updateHeatmap(heatmapData) {
            const options = {
                series: heatmapData.series,
                chart: {
                    type: 'heatmap',
                    height: WIDGET_DEFAULTS.chartHeight,
                    toolbar: { show: false },
                    background: 'transparent'
                },
                dataLabels: {
                    enabled: false
                },
                colors: [this.colors.primary],
                xaxis: {
                    type: 'category',
                    labels: {
                        show: true,
                        rotate: -45,
                        rotateAlways: false,
                        hideOverlappingLabels: true
                    }
                },
                theme: {
                    mode: this.isDarkMode ? 'dark' : 'light'
                },
                plotOptions: {
                    heatmap: {
                        radius: 2,
                        enableShades: true,
                        shadeIntensity: 0.5,
                        colorScale: {
                            ranges: [
                                { from: 0, to: 10, color: this.colors.success, name: 'Low' },
                                { from: 11, to: 50, color: this.colors.warning, name: 'Medium' },
                                { from: 51, to: 100, color: this.colors.danger, name: 'High' },
                                { from: 101, to: 99999, color: this.colors.dark, name: 'Very High' }
                            ]
                        }
                    }
                }
            };

            if (this.chart) {
                this.chart.updateOptions(options);
            } else {
                this.chart = new ApexCharts(document.querySelector(`#${this.id}`), options);
                this.chart.render();
            }
        }
    }

    // Real-time Line Chart Widget
    class RealtimeMetricsWidget extends EntityDBWidget {
        constructor(options) {
            super({
                ...options,
                type: 'realtime-metrics',
                title: options.title || 'Real-time Metrics',
                refreshInterval: options.refreshInterval || 5000, // 5 seconds for real-time
                apiEndpoint: '/api/v1/system/metrics'
            });
            
            this.maxDataPoints = 50;
            this.dataBuffer = [];
        }

        onDataUpdate(data) {
            const timestamp = new Date();
            const value = this.extractMetricValue(data);
            
            this.dataBuffer.push({ x: timestamp, y: value });
            
            // Keep only last N points
            if (this.dataBuffer.length > this.maxDataPoints) {
                this.dataBuffer.shift();
            }
            
            this.updateRealtimeChart();
        }

        extractMetricValue(data) {
            // Override in subclasses to extract specific metric
            return Math.random() * 100;
        }

        updateRealtimeChart() {
            const options = {
                series: [{
                    name: 'Value',
                    data: this.dataBuffer
                }],
                chart: {
                    type: 'line',
                    height: WIDGET_DEFAULTS.chartHeight,
                    toolbar: { show: false },
                    background: 'transparent',
                    animations: {
                        enabled: true,
                        easing: 'linear',
                        dynamicAnimation: {
                            speed: 1000
                        }
                    }
                },
                stroke: {
                    curve: 'smooth',
                    width: 3
                },
                xaxis: {
                    type: 'datetime',
                    range: 5 * 60 * 1000, // 5 minutes window
                    labels: {
                        datetimeUTC: false
                    }
                },
                yaxis: {
                    decimalsInFloat: 2
                },
                colors: [this.colors.primary],
                theme: {
                    mode: this.isDarkMode ? 'dark' : 'light'
                },
                legend: {
                    show: false
                }
            };

            if (this.chart) {
                this.chart.updateSeries([{
                    data: this.dataBuffer
                }]);
            } else {
                this.chart = new ApexCharts(document.querySelector(`#${this.id}`), options);
                this.chart.render();
            }
        }
    }

    // Widget Factory
    class WidgetFactory {
        static createWidget(type, options) {
            switch (type) {
                case 'query-performance':
                    return new QueryPerformanceWidget(options);
                case 'storage-metrics':
                    return new StorageMetricsWidget(options);
                case 'memory-analytics':
                    return new MemoryAnalyticsWidget(options);
                case 'session-metrics':
                    return new SessionMetricsWidget(options);
                case 'error-tracking':
                    return new ErrorTrackingWidget(options);
                case 'temporal-metrics':
                    return new TemporalMetricsWidget(options);
                case 'relationship-metrics':
                    return new RelationshipMetricsWidget(options);
                case 'tag-index-metrics':
                    return new TagIndexMetricsWidget(options);
                case 'activity-heatmap':
                    return new ActivityHeatmapWidget(options);
                case 'realtime-metrics':
                    return new RealtimeMetricsWidget(options);
                default:
                    throw new Error(`Unknown widget type: ${type}`);
            }
        }

        static getAvailableWidgets() {
            return [
                {
                    type: 'query-performance',
                    name: 'Query Performance',
                    description: 'Monitor query latency and throughput',
                    icon: 'fas fa-tachometer-alt',
                    defaultSize: { w: 6, h: 8 }
                },
                {
                    type: 'storage-metrics',
                    name: 'Storage Metrics',
                    description: 'Track storage performance and WAL size',
                    icon: 'fas fa-hdd',
                    defaultSize: { w: 6, h: 8 }
                },
                {
                    type: 'memory-analytics',
                    name: 'Memory Analytics',
                    description: 'Memory usage and GC statistics',
                    icon: 'fas fa-memory',
                    defaultSize: { w: 6, h: 8 }
                },
                {
                    type: 'session-metrics',
                    name: 'Session Metrics',
                    description: 'Authentication and session analytics',
                    icon: 'fas fa-users',
                    defaultSize: { w: 4, h: 8 }
                },
                {
                    type: 'error-tracking',
                    name: 'Error Tracking',
                    description: 'Monitor error rates and patterns',
                    icon: 'fas fa-exclamation-triangle',
                    defaultSize: { w: 8, h: 8 }
                },
                {
                    type: 'temporal-metrics',
                    name: 'Temporal Operations',
                    description: 'Temporal query performance',
                    icon: 'fas fa-history',
                    defaultSize: { w: 6, h: 8 }
                },
                {
                    type: 'relationship-metrics',
                    name: 'Relationship Metrics',
                    description: 'Entity relationship statistics',
                    icon: 'fas fa-project-diagram',
                    defaultSize: { w: 6, h: 8 }
                },
                {
                    type: 'tag-index-metrics',
                    name: 'Tag Index Performance',
                    description: 'Tag indexing and lookup metrics',
                    icon: 'fas fa-tags',
                    defaultSize: { w: 6, h: 8 }
                },
                {
                    type: 'activity-heatmap',
                    name: 'Activity Heatmap',
                    description: 'Visualize activity patterns over time',
                    icon: 'fas fa-th',
                    defaultSize: { w: 12, h: 8 }
                },
                {
                    type: 'realtime-metrics',
                    name: 'Real-time Monitor',
                    description: 'Live streaming metrics',
                    icon: 'fas fa-stream',
                    defaultSize: { w: 6, h: 8 }
                }
            ];
        }
    }

    // Vue 3 Component for Widgets
    const EntityDBWidgetComponent = {
        props: {
            widgetConfig: {
                type: Object,
                required: true
            },
            isDarkMode: {
                type: Boolean,
                default: false
            }
        },
        
        template: `
            <div class="entitydb-widget" :class="widgetClass">
                <div class="widget-header">
                    <h3 class="widget-title">{{ widget.title }}</h3>
                    <div class="widget-controls">
                        <select v-if="showTimeRange" v-model="timeRange" @change="onTimeRangeChange" class="time-range-selector">
                            <option v-for="range in timeRanges" :key="range.value" :value="range.value">
                                {{ range.label }}
                            </option>
                        </select>
                        <button @click="refresh" class="widget-btn" title="Refresh">
                            <i class="fas fa-sync-alt" :class="{ 'fa-spin': isLoading }"></i>
                        </button>
                        <button @click="exportData" class="widget-btn" title="Export Data">
                            <i class="fas fa-download"></i>
                        </button>
                        <button @click="toggleFullscreen" class="widget-btn" title="Fullscreen">
                            <i class="fas fa-expand"></i>
                        </button>
                    </div>
                </div>
                
                <div class="widget-body">
                    <div v-if="error" class="widget-error">
                        <i class="fas fa-exclamation-circle"></i>
                        <span>{{ error }}</span>
                        <button @click="refresh" class="retry-btn">Retry</button>
                    </div>
                    
                    <div v-else-if="isLoading && !widget" class="widget-loading">
                        <i class="fas fa-spinner fa-spin"></i>
                        <span>Loading...</span>
                    </div>
                    
                    <div v-else class="widget-chart" :id="chartId"></div>
                </div>
                
                <div v-if="showFooter" class="widget-footer">
                    <span class="last-updated">Last updated: {{ lastUpdated }}</span>
                    <span class="auto-refresh" v-if="widget && widget.refreshInterval > 0">
                        Auto-refresh: {{ formatInterval(widget.refreshInterval) }}
                    </span>
                </div>
            </div>
        `,
        
        data() {
            return {
                widget: null,
                chartId: `widget-chart-${Date.now()}`,
                isLoading: false,
                error: null,
                timeRange: '1h',
                timeRanges: TIME_RANGES,
                lastUpdated: null,
                isFullscreen: false
            };
        },
        
        computed: {
            widgetClass() {
                return {
                    'dark-mode': this.isDarkMode,
                    'loading': this.isLoading,
                    'has-error': this.error,
                    'fullscreen': this.isFullscreen
                };
            },
            
            showTimeRange() {
                return ['query-performance', 'error-tracking', 'activity-heatmap'].includes(this.widgetConfig.type);
            },
            
            showFooter() {
                return true;
            }
        },
        
        mounted() {
            this.initWidget();
        },
        
        beforeUnmount() {
            if (this.widget) {
                this.widget.destroy();
            }
        },
        
        methods: {
            initWidget() {
                const options = {
                    ...this.widgetConfig,
                    id: this.chartId,
                    isDarkMode: this.isDarkMode,
                    hours: this.getHoursFromTimeRange(this.timeRange)
                };
                
                try {
                    this.widget = WidgetFactory.createWidget(this.widgetConfig.type, options);
                    this.widget.start();
                    this.lastUpdated = new Date().toLocaleTimeString();
                } catch (error) {
                    this.error = error.message;
                    console.error('Failed to initialize widget:', error);
                }
            },
            
            refresh() {
                if (this.widget) {
                    this.widget.fetchData();
                    this.lastUpdated = new Date().toLocaleTimeString();
                }
            },
            
            exportData() {
                if (this.widget) {
                    this.widget.exportData();
                }
            },
            
            toggleFullscreen() {
                this.isFullscreen = !this.isFullscreen;
                
                if (this.isFullscreen) {
                    this.$el.requestFullscreen();
                } else {
                    document.exitFullscreen();
                }
            },
            
            onTimeRangeChange() {
                if (this.widget) {
                    this.widget.destroy();
                    this.initWidget();
                }
            },
            
            getHoursFromTimeRange(range) {
                const found = TIME_RANGES.find(r => r.value === range);
                return found ? found.minutes / 60 : 1;
            },
            
            formatInterval(ms) {
                const seconds = ms / 1000;
                if (seconds < 60) return `${seconds}s`;
                const minutes = seconds / 60;
                if (minutes < 60) return `${minutes}m`;
                const hours = minutes / 60;
                return `${hours}h`;
            }
        },
        
        watch: {
            isDarkMode(newVal) {
                if (this.widget) {
                    this.widget.isDarkMode = newVal;
                    this.widget.fetchData();
                }
            }
        }
    };

    // Export to global scope
    window.EntityDBWidgets = {
        WidgetFactory,
        EntityDBWidget,
        QueryPerformanceWidget,
        StorageMetricsWidget,
        MemoryAnalyticsWidget,
        SessionMetricsWidget,
        ErrorTrackingWidget,
        TemporalMetricsWidget,
        RelationshipMetricsWidget,
        TagIndexMetricsWidget,
        ActivityHeatmapWidget,
        RealtimeMetricsWidget,
        EntityDBWidgetComponent,
        TIME_RANGES,
        WIDGET_DEFAULTS
    };

})(window);