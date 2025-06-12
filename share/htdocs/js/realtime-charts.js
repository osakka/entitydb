/**
 * EntityDB Real-time Charts Module
 * Provides real-time chart updates for dashboard
 */

class RealtimeCharts {
    constructor() {
        this.charts = new Map();
        this.updateInterval = 5000; // 5 seconds
        this.dataRetention = 50; // Keep last 50 data points
    }

    async initializeDashboardCharts() {
        // Check if API client is available
        if (!window.apiClient) {
            console.warn('API client not available for real-time charts');
            return;
        }

        // Initialize basic charts
        await this.initializeSystemOverview();
        
        // Start real-time updates
        this.startUpdates();
    }

    async initializeSystemOverview() {
        // Simple system overview chart
        const canvas = document.getElementById('metrics-overview-chart');
        if (!canvas) return;

        try {
            const response = await window.apiClient.get('/health');
            const data = response.data || response;
            
            this.createSimpleMetricsDisplay(canvas, data);
        } catch (error) {
            console.error('Failed to load system metrics:', error);
            this.createErrorDisplay(canvas);
        }
    }

    createSimpleMetricsDisplay(canvas, data) {
        const container = canvas.parentElement;
        container.innerHTML = `
            <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(120px, 1fr)); gap: 16px; padding: 20px;">
                <div style="text-align: center;">
                    <div style="font-size: 24px; font-weight: 600; color: #3498db;">
                        ${data.metrics?.entity_count || 0}
                    </div>
                    <div style="font-size: 12px; color: #6c757d; margin-top: 4px;">
                        Entities
                    </div>
                </div>
                <div style="text-align: center;">
                    <div style="font-size: 24px; font-weight: 600; color: #2ecc71;">
                        ${data.metrics?.goroutines || 0}
                    </div>
                    <div style="font-size: 12px; color: #6c757d; margin-top: 4px;">
                        Goroutines
                    </div>
                </div>
                <div style="text-align: center;">
                    <div style="font-size: 24px; font-weight: 600; color: #f39c12;">
                        ${this.formatBytes(data.metrics?.memory_usage?.alloc_bytes || 0)}
                    </div>
                    <div style="font-size: 12px; color: #6c757d; margin-top: 4px;">
                        Memory
                    </div>
                </div>
                <div style="text-align: center;">
                    <div style="font-size: 24px; font-weight: 600; color: #9b59b6;">
                        ${data.uptime || '0h 0m'}
                    </div>
                    <div style="font-size: 12px; color: #6c757d; margin-top: 4px;">
                        Uptime
                    </div>
                </div>
            </div>
        `;
    }

    createErrorDisplay(canvas) {
        const container = canvas.parentElement;
        container.innerHTML = `
            <div style="display: flex; align-items: center; justify-content: center; height: 100%; color: #6c757d;">
                <div style="text-align: center;">
                    <i class="fas fa-exclamation-triangle" style="font-size: 48px; margin-bottom: 16px; opacity: 0.3;"></i>
                    <p>Unable to load metrics</p>
                    <small>Check server connection</small>
                </div>
            </div>
        `;
    }

    startUpdates() {
        // Update system overview periodically
        setInterval(async () => {
            await this.updateSystemOverview();
        }, this.updateInterval);
    }

    async updateSystemOverview() {
        const canvas = document.getElementById('metrics-overview-chart');
        if (!canvas || !window.apiClient) return;

        try {
            const response = await window.apiClient.get('/health');
            const data = response.data || response;
            this.createSimpleMetricsDisplay(canvas, data);
        } catch (error) {
            console.error('Failed to update system metrics:', error);
        }
    }

    formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    destroy() {
        // Cleanup any charts
        this.charts.clear();
    }
}

// Create global instance
window.realtimeCharts = new RealtimeCharts();