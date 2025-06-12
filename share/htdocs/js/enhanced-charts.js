/**
 * EntityDB Enhanced Charts Component
 * Real-time updating charts with smooth animations
 */

class EnhancedCharts {
    constructor() {
        this.charts = new Map();
        this.updateInterval = 5000; // 5 seconds
        this.maxDataPoints = 50;
        this.colors = {
            primary: '#3498db',
            success: '#2ecc71',
            warning: '#f39c12',
            danger: '#e74c3c',
            info: '#9b59b6',
            secondary: '#95a5a6'
        };
    }

    createChart(canvasId, config) {
        // Check if Chart.js is available
        if (typeof Chart === 'undefined') {
            console.warn('Chart.js not available, falling back to simple display');
            return this.createFallbackChart(canvasId, config);
        }

        const ctx = document.getElementById(canvasId);
        if (!ctx) return null;

        // Destroy existing chart
        if (this.charts.has(canvasId)) {
            this.charts.get(canvasId).destroy();
        }

        // Enhanced default config
        const enhancedConfig = {
            ...config,
            options: {
                responsive: true,
                maintainAspectRatio: false,
                animation: {
                    duration: 750,
                    easing: 'easeInOutQuart'
                },
                plugins: {
                    legend: {
                        display: config.type !== 'line' || (config.data.datasets && config.data.datasets.length > 1),
                        position: 'bottom',
                        labels: {
                            usePointStyle: true,
                            padding: 15,
                            font: {
                                size: 12
                            }
                        }
                    },
                    tooltip: {
                        mode: 'index',
                        intersect: false,
                        backgroundColor: 'rgba(0, 0, 0, 0.8)',
                        titleFont: {
                            size: 14,
                            weight: 'bold'
                        },
                        bodyFont: {
                            size: 13
                        },
                        padding: 12,
                        cornerRadius: 8,
                        displayColors: true,
                        callbacks: {
                            label: (context) => {
                                let label = context.dataset.label || '';
                                if (label) {
                                    label += ': ';
                                }
                                if (context.parsed.y !== null) {
                                    label += this.formatValue(context.parsed.y, context.dataset.unit);
                                }
                                return label;
                            }
                        }
                    }
                },
                scales: this.getScalesConfig(config),
                ...config.options
            }
        };

        const chart = new Chart(ctx, enhancedConfig);
        this.charts.set(canvasId, chart);
        return chart;
    }

    createFallbackChart(canvasId, config) {
        // Simple fallback when Chart.js is not available
        const canvas = document.getElementById(canvasId);
        if (!canvas) return null;

        const container = canvas.parentElement;
        container.innerHTML = `
            <div style="display: flex; align-items: center; justify-content: center; height: 100%; color: #6c757d;">
                <div style="text-align: center;">
                    <i class="fas fa-chart-line" style="font-size: 48px; margin-bottom: 16px; opacity: 0.3;"></i>
                    <p>Chart View</p>
                    <small>Chart.js loading...</small>
                </div>
            </div>
        `;
        return null;
    }

    getScalesConfig(config) {
        if (config.type === 'pie' || config.type === 'doughnut') {
            return {};
        }

        return {
            x: {
                grid: {
                    display: false
                },
                ticks: {
                    maxRotation: 45,
                    minRotation: 0,
                    autoSkip: true,
                    maxTicksLimit: 10
                }
            },
            y: {
                beginAtZero: true,
                grid: {
                    color: 'rgba(0, 0, 0, 0.05)',
                    drawBorder: false
                },
                ticks: {
                    callback: (value) => this.formatValue(value, config.unit)
                }
            }
        };
    }

    formatValue(value, unit) {
        if (unit === 'bytes') {
            return this.formatBytes(value);
        } else if (unit === 'ms') {
            return `${value.toFixed(0)}ms`;
        } else if (unit === 'percent') {
            return `${value.toFixed(1)}%`;
        } else if (value >= 1000000) {
            return `${(value / 1000000).toFixed(1)}M`;
        } else if (value >= 1000) {
            return `${(value / 1000).toFixed(1)}K`;
        }
        return value.toFixed(0);
    }

    formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    destroy(canvasId) {
        const chart = this.charts.get(canvasId);
        if (chart) {
            chart.destroy();
            this.charts.delete(canvasId);
        }
    }

    destroyAll() {
        this.charts.forEach(chart => chart.destroy());
        this.charts.clear();
    }
}

// Create global instance
window.enhancedCharts = new EnhancedCharts();