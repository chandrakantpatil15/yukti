// Dashboard JavaScript for Yukti FinOps
let costTrendChart, resourceDistributionChart;
let currentFilters = {};

// Initialize dashboard
document.addEventListener('DOMContentLoaded', function() {
    initializeCharts();
    loadInitialData();
    updateLastUpdated();
    
    // Set up filter event listeners
    setupFilterListeners();
});

function setupFilterListeners() {
    const filters = ['resourceTypeFilter', 'environmentFilter', 'instanceTypeFilter', 'timeRangeFilter'];
    filters.forEach(filterId => {
        document.getElementById(filterId).addEventListener('change', applyFilters);
    });
}

function initializeCharts() {
    // Cost Trend Chart
    const costCtx = document.getElementById('costTrendChart').getContext('2d');
    costTrendChart = new Chart(costCtx, {
        type: 'line',
        data: {
            labels: [],
            datasets: [{
                label: 'Daily Cost ($)',
                data: [],
                borderColor: 'rgb(59, 130, 246)',
                backgroundColor: 'rgba(59, 130, 246, 0.1)',
                tension: 0.4
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        callback: function(value) {
                            return '$' + value.toFixed(2);
                        }
                    }
                }
            }
        }
    });

    // Resource Distribution Chart
    const resourceCtx = document.getElementById('resourceDistributionChart').getContext('2d');
    resourceDistributionChart = new Chart(resourceCtx, {
        type: 'doughnut',
        data: {
            labels: [],
            datasets: [{
                data: [],
                backgroundColor: [
                    '#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6'
                ]
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    position: 'bottom'
                }
            }
        }
    });
}

async function loadInitialData() {
    try {
        await Promise.all([
            loadKPIs(),
            loadCostTrend(),
            loadResourceDistribution(),
            loadResources(),
            loadRecommendations(),
            loadInstanceTypes()
        ]);
    } catch (error) {
        console.error('Error loading initial data:', error);
        showError('Failed to load dashboard data');
    }
}

async function loadKPIs() {
    try {
        const response = await fetch('/api/v1/cost/summary?days=30');
        const data = await response.json();
        
        document.getElementById('totalCost').textContent = '$' + (data.total_cost || 0).toFixed(2);
        document.getElementById('resourceCount').textContent = data.resource_count || 0;
        
        // Load recommendations for KPIs
        const recResponse = await fetch('/api/v1/recommendations');
        const recommendations = await recResponse.json();
        
        const totalSavings = recommendations.reduce((sum, rec) => sum + parseFloat(rec.potential_savings || 0), 0);
        document.getElementById('potentialSavings').textContent = '$' + totalSavings.toFixed(2);
        document.getElementById('recommendationCount').textContent = recommendations.length;
        
    } catch (error) {
        console.error('Error loading KPIs:', error);
    }
}

async function loadCostTrend() {
    try {
        const days = document.getElementById('timeRangeFilter').value || 30;
        const response = await fetch(`/api/v1/costs?days=${days}`);
        const costs = await response.json();
        
        // Group costs by date
        const costByDate = {};
        costs.forEach(cost => {
            const date = new Date(cost.date).toLocaleDateString();
            costByDate[date] = (costByDate[date] || 0) + parseFloat(cost.cost_usd);
        });
        
        const labels = Object.keys(costByDate).sort();
        const data = labels.map(date => costByDate[date]);
        
        costTrendChart.data.labels = labels;
        costTrendChart.data.datasets[0].data = data;
        costTrendChart.update();
        
    } catch (error) {
        console.error('Error loading cost trend:', error);
    }
}

async function loadResourceDistribution() {
    try {
        const response = await fetch('/api/v1/resources');
        const resources = await response.json();
        
        // Group by instance type
        const distribution = {};
        resources.forEach(resource => {
            const type = resource.instance_type || 'unknown';
            distribution[type] = (distribution[type] || 0) + 1;
        });
        
        const labels = Object.keys(distribution);
        const data = Object.values(distribution);
        
        resourceDistributionChart.data.labels = labels;
        resourceDistributionChart.data.datasets[0].data = data;
        resourceDistributionChart.update();
        
    } catch (error) {
        console.error('Error loading resource distribution:', error);
    }
}

async function loadResources() {
    try {
        const response = await fetch('/api/v1/resources');
        const resources = await response.json();
        
        const tbody = document.getElementById('resourcesTable');
        tbody.innerHTML = '';
        
        resources.slice(0, 10).forEach(resource => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${resource.resource_id}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${resource.instance_type}</td>
                <td class="px-6 py-4 whitespace-nowrap">
                    <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${getEnvironmentColor(resource.environment)}">
                        ${resource.environment}
                    </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                    <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${getStatusColor(resource.status)}">
                        ${resource.status}
                    </span>
                </td>
            `;
            tbody.appendChild(row);
        });
        
    } catch (error) {
        console.error('Error loading resources:', error);
    }
}

async function loadRecommendations() {
    try {
        const response = await fetch('/api/v1/recommendations');
        const recommendations = await response.json();
        
        const tbody = document.getElementById('recommendationsTable');
        tbody.innerHTML = '';
        
        recommendations.slice(0, 10).forEach(rec => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${rec.recommendation_type}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-green-600 font-medium">$${parseFloat(rec.potential_savings || 0).toFixed(2)}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${Math.round((rec.confidence || 0) * 100)}%</td>
            `;
            tbody.appendChild(row);
        });
        
    } catch (error) {
        console.error('Error loading recommendations:', error);
    }
}

async function loadInstanceTypes() {
    try {
        const response = await fetch('/api/v1/resources');
        const resources = await response.json();
        
        const instanceTypes = [...new Set(resources.map(r => r.instance_type))].sort();
        const select = document.getElementById('instanceTypeFilter');
        
        // Clear existing options except "All Instance Types"
        select.innerHTML = '<option value="">All Instance Types</option>';
        
        instanceTypes.forEach(type => {
            const option = document.createElement('option');
            option.value = type;
            option.textContent = type;
            select.appendChild(option);
        });
        
    } catch (error) {
        console.error('Error loading instance types:', error);
    }
}

function applyFilters() {
    currentFilters = {
        resourceType: document.getElementById('resourceTypeFilter').value,
        environment: document.getElementById('environmentFilter').value,
        instanceType: document.getElementById('instanceTypeFilter').value,
        timeRange: document.getElementById('timeRangeFilter').value
    };
    
    // Reload data with filters
    loadInitialData();
}

function clearFilters() {
    document.getElementById('resourceTypeFilter').value = '';
    document.getElementById('environmentFilter').value = '';
    document.getElementById('instanceTypeFilter').value = '';
    document.getElementById('timeRangeFilter').value = '30';
    
    currentFilters = {};
    loadInitialData();
}

function refreshData() {
    loadInitialData();
    updateLastUpdated();
    showSuccess('Data refreshed successfully');
}

function updateLastUpdated() {
    document.getElementById('lastUpdated').textContent = new Date().toLocaleTimeString();
}

function getEnvironmentColor(environment) {
    const colors = {
        'prod': 'bg-red-100 text-red-800',
        'staging': 'bg-yellow-100 text-yellow-800',
        'dev': 'bg-blue-100 text-blue-800',
        'test': 'bg-green-100 text-green-800'
    };
    return colors[environment] || 'bg-gray-100 text-gray-800';
}

function getStatusColor(status) {
    const colors = {
        'running': 'bg-green-100 text-green-800',
        'stopped': 'bg-red-100 text-red-800',
        'pending': 'bg-yellow-100 text-yellow-800'
    };
    return colors[status] || 'bg-gray-100 text-gray-800';
}

function showSuccess(message) {
    // Simple success notification
    const notification = document.createElement('div');
    notification.className = 'fixed top-4 right-4 bg-green-500 text-white px-6 py-3 rounded-md shadow-lg z-50';
    notification.textContent = message;
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.remove();
    }, 3000);
}

function showError(message) {
    // Simple error notification
    const notification = document.createElement('div');
    notification.className = 'fixed top-4 right-4 bg-red-500 text-white px-6 py-3 rounded-md shadow-lg z-50';
    notification.textContent = message;
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.remove();
    }, 5000);
}

// Auto-refresh every 5 minutes
setInterval(refreshData, 5 * 60 * 1000);