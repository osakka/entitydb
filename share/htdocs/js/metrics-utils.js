// Utility functions for metrics display

// Calculate trend from historical data
async function calculateTrend(metricName, hours = 1) {
    try {
        const response = await fetch(`/api/v1/metrics/history/v2?metric_name=${metricName}&hours=${hours}`);
        if (!response.ok) return null;
        
        const data = await response.json();
        if (!data.data_points || data.data_points.length < 2) return null;
        
        // Sort by timestamp
        const points = data.data_points.sort((a, b) => 
            new Date(a.timestamp) - new Date(b.timestamp)
        );
        
        // Calculate percentage change from first to last
        const firstValue = points[0].value;
        const lastValue = points[points.length - 1].value;
        
        if (firstValue === 0) return null;
        
        const percentChange = ((lastValue - firstValue) / firstValue) * 100;
        
        return {
            change: percentChange,
            direction: percentChange > 0 ? 'up' : percentChange < 0 ? 'down' : 'stable',
            symbol: percentChange > 0 ? '↑' : percentChange < 0 ? '↓' : '→',
            formatted: `${percentChange > 0 ? '+' : ''}${percentChange.toFixed(1)}%`
        };
    } catch (error) {
        console.error('Error calculating trend:', error);
        return null;
    }
}

// Format trend for display
function formatTrend(trend) {
    if (!trend) return '';
    
    const color = trend.direction === 'up' ? '#10B981' : 
                  trend.direction === 'down' ? '#EF4444' : '#6B7280';
    
    return `<span style="color: ${color}">${trend.symbol} ${trend.formatted}</span>`;
}

// Get current metric value
async function getCurrentMetricValue(metricName) {
    try {
        const response = await fetch('/api/v1/metrics/current');
        if (!response.ok) return null;
        
        const data = await response.json();
        return data[metricName] || null;
    } catch (error) {
        console.error('Error fetching current metric:', error);
        return null;
    }
}

// Update stat card with real-time data
async function updateStatCard(cardElement, metricName, formatter = (v) => v) {
    const valueElement = cardElement.querySelector('.stat-value');
    const trendElement = cardElement.querySelector('.stat-trend');
    
    if (!valueElement) return;
    
    // Get current value
    const current = await getCurrentMetricValue(metricName);
    if (current) {
        valueElement.textContent = formatter(current.value);
    }
    
    // Get trend
    if (trendElement) {
        const trend = await calculateTrend(metricName, 1);
        if (trend) {
            trendElement.innerHTML = formatTrend(trend);
        }
    }
}

// Export functions
window.metricsUtils = {
    calculateTrend,
    formatTrend,
    getCurrentMetricValue,
    updateStatCard
};