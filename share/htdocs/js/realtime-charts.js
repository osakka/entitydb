// Real-time chart initialization with actual metrics data

async function fetchMetricHistory(metricName, hours = 1) {
    try {
        // Use the real metrics history endpoint
        const response = await fetch(`/api/v1/metrics/history?metric_name=${metricName}&hours=${hours}`);
        if (!response.ok) {
            console.error(`Failed to fetch metric history for ${metricName}:`, response.status);
            
            // Fall back to system metrics for current value
            const sysResponse = await fetch('/api/v1/system/metrics');
            if (!sysResponse.ok) {
                return null;
            }
            const metrics = await sysResponse.json();
            
            // Map metric names to system metric values
            let currentValue = 0;
            switch(metricName) {
                case 'database_size':
                    currentValue = metrics.storage?.file_sizes?.database_bytes || 0;
                    break;
                case 'wal_size':
                    currentValue = metrics.storage?.file_sizes?.wal_bytes || 0;
                    break;
                case 'index_size':
                    currentValue = metrics.storage?.file_sizes?.index_bytes || 0;
                    break;
                case 'memory_alloc':
                    currentValue = metrics.memory?.alloc_bytes || 0;
                    break;
                case 'memory_heap_alloc':
                    currentValue = metrics.memory?.heap_alloc_bytes || 0;
                    break;
                case 'entity_count_total':
                    currentValue = metrics.database?.total_entities || 0;
                    break;
                default:
                    console.warn(`Unknown metric: ${metricName}`);
                    return null;
            }
            
            // Return single current value
            return {
                metric_name: metricName,
                data_points: [{
                    timestamp: new Date().toISOString(),
                    value: currentValue
                }],
                unit: metricName.includes('bytes') || metricName.includes('size') ? 'bytes' : 'count'
            };
        }
        
        const data = await response.json();
        
        // If no historical data, add current value from system metrics
        if (!data.data_points || data.data_points.length === 0) {
            const sysResponse = await fetch('/api/v1/system/metrics');
            if (sysResponse.ok) {
                const metrics = await sysResponse.json();
                let currentValue = 0;
                
                switch(metricName) {
                    case 'database_size':
                        currentValue = metrics.storage?.file_sizes?.database_bytes || 0;
                        break;
                    case 'wal_size':
                        currentValue = metrics.storage?.file_sizes?.wal_bytes || 0;
                        break;
                    case 'index_size':
                        currentValue = metrics.storage?.file_sizes?.index_bytes || 0;
                        break;
                    case 'memory_alloc':
                        currentValue = metrics.memory?.alloc_bytes || 0;
                        break;
                    case 'memory_heap_alloc':
                        currentValue = metrics.memory?.heap_alloc_bytes || 0;
                        break;
                    case 'entity_count_total':
                        currentValue = metrics.database?.total_entities || 0;
                        break;
                }
                
                if (currentValue > 0) {
                    data.data_points = [{
                        timestamp: new Date().toISOString(),
                        value: currentValue
                    }];
                }
            }
        }
        
        return data;
    } catch (error) {
        console.error(`Error fetching metric data for ${metricName}:`, error);
        return null;
    }
}

function prepareChartData(metricHistory, label, color) {
    if (!metricHistory || !metricHistory.data_points) {
        return {
            labels: [],
            datasets: []
        };
    }
    
    // Sort data points by timestamp
    const dataPoints = metricHistory.data_points.sort((a, b) => 
        new Date(a.timestamp) - new Date(b.timestamp)
    );
    
    // Prepare labels and values
    const labels = dataPoints.map(point => {
        const date = new Date(point.timestamp);
        return date.toLocaleTimeString('en-US', { 
            hour: '2-digit', 
            minute: '2-digit'
        });
    });
    
    const values = dataPoints.map(point => point.value);
    
    return {
        labels: labels,
        datasets: [{
            label: label,
            data: values,
            borderColor: color,
            backgroundColor: color + '20',
            tension: 0.1,
            fill: true
        }]
    };
}

async function initRealTimeCharts() {
    console.log('Initializing real-time charts...');
    
    // Storage Chart - Database and WAL sizes
    const storageCanvas = document.getElementById('storageChart');
    if (storageCanvas && storageCanvas.offsetParent !== null) {
        try {
            // Fetch real metrics
            const [dbSizeHistory, walSizeHistory, indexSizeHistory] = await Promise.all([
                fetchMetricHistory('database_size', 6),
                fetchMetricHistory('wal_size', 6),
                fetchMetricHistory('index_size', 6)
            ]);
            
            // Prepare datasets
            const datasets = [];
            
            if (dbSizeHistory) {
                const dbData = dbSizeHistory.data_points.map(p => ({
                    x: new Date(p.timestamp),
                    y: p.value / (1024 * 1024) // Convert to MB
                }));
                datasets.push({
                    label: 'Database (MB)',
                    data: dbData,
                    borderColor: '#3B82F6',
                    backgroundColor: 'rgba(59, 130, 246, 0.1)',
                    tension: 0.1
                });
            }
            
            if (walSizeHistory) {
                const walData = walSizeHistory.data_points.map(p => ({
                    x: new Date(p.timestamp),
                    y: p.value / (1024 * 1024) // Convert to MB
                }));
                datasets.push({
                    label: 'WAL (MB)',
                    data: walData,
                    borderColor: '#10B981',
                    backgroundColor: 'rgba(16, 185, 129, 0.1)',
                    tension: 0.1
                });
            }
            
            if (indexSizeHistory) {
                const indexData = indexSizeHistory.data_points.map(p => ({
                    x: new Date(p.timestamp),
                    y: p.value / (1024 * 1024) // Convert to MB
                }));
                datasets.push({
                    label: 'Index (MB)',
                    data: indexData,
                    borderColor: '#F59E0B',
                    backgroundColor: 'rgba(245, 158, 11, 0.1)',
                    tension: 0.1
                });
            }
            
            const ctx = storageCanvas.getContext('2d');
            new Chart(ctx, {
                type: 'line',
                data: { datasets },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        x: {
                            type: 'time',
                            time: {
                                unit: 'hour',
                                displayFormats: {
                                    hour: 'HH:mm'
                                }
                            },
                            title: {
                                display: true,
                                text: 'Time'
                            }
                        },
                        y: {
                            beginAtZero: true,
                            title: {
                                display: true,
                                text: 'Size (MB)'
                            }
                        }
                    },
                    plugins: {
                        legend: {
                            position: 'top'
                        },
                        title: {
                            display: true,
                            text: 'Storage Growth Over Time'
                        }
                    }
                }
            });
            console.log('Storage chart created with real data');
        } catch (error) {
            console.error('Error creating storage chart:', error);
        }
    }
    
    // Memory Usage Chart
    const memoryCanvas = document.getElementById('memoryChart');
    if (memoryCanvas && memoryCanvas.offsetParent !== null) {
        try {
            const memoryHistory = await fetchMetricHistory('memory_alloc', 1);
            
            if (memoryHistory) {
                const memoryData = memoryHistory.data_points.map(p => ({
                    x: new Date(p.timestamp),
                    y: p.value / (1024 * 1024) // Convert to MB
                }));
                
                const ctx = memoryCanvas.getContext('2d');
                new Chart(ctx, {
                    type: 'line',
                    data: {
                        datasets: [{
                            label: 'Memory Usage (MB)',
                            data: memoryData,
                            borderColor: '#8B5CF6',
                            backgroundColor: 'rgba(139, 92, 246, 0.1)',
                            tension: 0.1,
                            fill: true
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        scales: {
                            x: {
                                type: 'time',
                                time: {
                                    unit: 'minute',
                                    displayFormats: {
                                        minute: 'HH:mm'
                                    }
                                },
                                title: {
                                    display: true,
                                    text: 'Time'
                                }
                            },
                            y: {
                                beginAtZero: true,
                                title: {
                                    display: true,
                                    text: 'Memory (MB)'
                                }
                            }
                        },
                        plugins: {
                            legend: {
                                position: 'top'
                            },
                            title: {
                                display: true,
                                text: 'Memory Usage Over Time'
                            }
                        }
                    }
                });
                console.log('Memory chart created with real data');
            }
        } catch (error) {
            console.error('Error creating memory chart:', error);
        }
    }
    
    // Entity Growth Chart
    const entityCanvas = document.getElementById('entityGrowthChart');
    if (entityCanvas && entityCanvas.offsetParent !== null) {
        try {
            const entityHistory = await fetchMetricHistory('entity_count_total', 24);
            
            if (entityHistory) {
                const entityData = entityHistory.data_points.map(p => ({
                    x: new Date(p.timestamp),
                    y: p.value
                }));
                
                const ctx = entityCanvas.getContext('2d');
                new Chart(ctx, {
                    type: 'line',
                    data: {
                        datasets: [{
                            label: 'Total Entities',
                            data: entityData,
                            borderColor: '#06B6D4',
                            backgroundColor: 'rgba(6, 182, 212, 0.1)',
                            tension: 0.1,
                            fill: true,
                            stepped: true
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        scales: {
                            x: {
                                type: 'time',
                                time: {
                                    unit: 'hour',
                                    displayFormats: {
                                        hour: 'MMM D, HH:mm'
                                    }
                                },
                                title: {
                                    display: true,
                                    text: 'Time'
                                }
                            },
                            y: {
                                beginAtZero: true,
                                title: {
                                    display: true,
                                    text: 'Entity Count'
                                }
                            }
                        },
                        plugins: {
                            legend: {
                                position: 'top'
                            },
                            title: {
                                display: true,
                                text: 'Entity Growth Over Time'
                            }
                        }
                    }
                });
                console.log('Entity growth chart created with real data');
            }
        } catch (error) {
            console.error('Error creating entity growth chart:', error);
        }
    }
}

// Export for use in main app
window.initRealTimeCharts = initRealTimeCharts;