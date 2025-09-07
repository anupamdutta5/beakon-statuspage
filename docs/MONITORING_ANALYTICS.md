# Monitoring Analytics System

## Overview

The Monitoring Analytics System provides comprehensive monitoring data visualization and performance analytics similar to statuspage.io. It includes uptime tracking, response time monitoring, error rate analysis, and incident correlation with beautiful, interactive charts and graphs.

## Features

### 📊 **Comprehensive Analytics Dashboard**
- **Real-time Metrics**: Live monitoring data with automatic refresh
- **Interactive Charts**: Beautiful Chart.js visualizations
- **Multiple Time Periods**: 1 hour, 24 hours, 7 days, 30 days, 90 days
- **Service Comparison**: Side-by-side service performance comparison
- **Trend Analysis**: Historical data analysis and trend identification

### 📈 **Uptime Monitoring**
- **Overall Uptime**: System-wide uptime percentage
- **Service Uptime**: Individual service uptime tracking
- **Uptime History**: Historical uptime trends over time
- **Status Tracking**: Real-time service status monitoring
- **Incident Correlation**: Uptime impact from incidents and maintenance

### ⚡ **Response Time Analytics**
- **Average Response Time**: Mean response time across all services
- **Response Time Distribution**: Min, Max, P95, P99 percentiles
- **Response Time Trends**: Historical response time patterns
- **Performance Benchmarking**: Service performance comparison
- **Anomaly Detection**: Unusual response time patterns

### 🚨 **Error Rate Monitoring**
- **Error Rate Tracking**: Percentage of failed requests
- **Error Type Analysis**: Categorization of different error types
- **Error Trends**: Historical error rate patterns
- **Service Error Comparison**: Error rates across services
- **Error Impact Assessment**: Impact of errors on service health

### 📋 **Incident Analytics**
- **Incident Timeline**: Chronological incident tracking
- **Incident Impact**: Severity and duration analysis
- **Service Impact**: Which services are affected by incidents
- **Resolution Time**: Time to resolve incidents
- **Incident Patterns**: Recurring incident analysis

## Architecture

### Backend Services

#### MonitoringAnalyticsService
```go
type MonitoringAnalyticsService struct {
    db *gorm.DB
}
```

**Key Methods:**
- `GetServiceUptimeData()` - Retrieves uptime data for a service
- `GetServiceResponseTimeData()` - Gets response time metrics
- `GetServiceHealthData()` - Returns health status for all services
- `GetMonitoringMetrics()` - Comprehensive monitoring metrics
- `GetServiceStatusHistory()` - Historical status data
- `GetUptimeSummary()` - Overall uptime summary

#### Data Structures

**UptimeData**
```go
type UptimeData struct {
    ServiceID   uint      `json:"service_id"`
    ServiceName string    `json:"service_name"`
    Timestamp   time.Time `json:"timestamp"`
    Uptime      float64   `json:"uptime"`      // Percentage (0-100)
    Downtime    float64   `json:"downtime"`    // Percentage (0-100)
    Status      string    `json:"status"`      // "up", "down", "degraded"
    ResponseTime float64  `json:"response_time"` // Milliseconds
    ErrorRate   float64   `json:"error_rate"`   // Percentage (0-100)
}
```

**ResponseTimeData**
```go
type ResponseTimeData struct {
    ServiceID    uint      `json:"service_id"`
    Timestamp    time.Time `json:"timestamp"`
    ResponseTime float64   `json:"response_time"` // Milliseconds
    MinTime      float64   `json:"min_time"`
    MaxTime      float64   `json:"max_time"`
    AvgTime      float64   `json:"avg_time"`
    P95Time      float64   `json:"p95_time"`
    P99Time      float64   `json:"p99_time"`
}
```

**ServiceHealthData**
```go
type ServiceHealthData struct {
    ServiceID     uint      `json:"service_id"`
    ServiceName   string    `json:"service_name"`
    Status        string    `json:"status"`
    Uptime        float64   `json:"uptime"`
    ResponseTime  float64   `json:"response_time"`
    ErrorRate     float64   `json:"error_rate"`
    LastChecked   time.Time `json:"last_checked"`
    IncidentCount int       `json:"incident_count"`
    MaintenanceCount int    `json:"maintenance_count"`
}
```

### API Endpoints

#### Uptime Analytics
```http
GET /api/v1/tenant/monitoring/uptime-summary
```
Returns overall uptime summary for all services.

**Response:**
```json
{
  "summary": {
    "total_services": 5,
    "up_services": 4,
    "degraded_services": 1,
    "down_services": 0,
    "average_uptime": 99.5,
    "overall_status": "up",
    "last_updated": "2024-01-15T10:30:00Z"
  }
}
```

#### Service Uptime Data
```http
GET /api/v1/tenant/monitoring/services/:serviceId/uptime?period=24h
```
Returns detailed uptime data for a specific service.

**Response:**
```json
{
  "service_id": 1,
  "period": "24h",
  "data": [
    {
      "service_id": 1,
      "service_name": "API Service",
      "timestamp": "2024-01-15T10:00:00Z",
      "uptime": 99.8,
      "downtime": 0.2,
      "status": "up",
      "response_time": 150.5,
      "error_rate": 0.1
    }
  ]
}
```

#### Response Time Data
```http
GET /api/v1/tenant/monitoring/services/:serviceId/response-time?period=24h
```
Returns response time metrics for a service.

**Response:**
```json
{
  "service_id": 1,
  "period": "24h",
  "data": [
    {
      "service_id": 1,
      "timestamp": "2024-01-15T10:00:00Z",
      "response_time": 150.0,
      "min_time": 120.0,
      "max_time": 225.0,
      "avg_time": 150.0,
      "p95_time": 180.0,
      "p99_time": 210.0
    }
  ]
}
```

#### Service Health Data
```http
GET /api/v1/tenant/monitoring/services/health
```
Returns health status for all services.

**Response:**
```json
{
  "services": [
    {
      "service_id": 1,
      "service_name": "API Service",
      "status": "up",
      "uptime": 99.8,
      "response_time": 150.0,
      "error_rate": 0.1,
      "last_checked": "2024-01-15T10:28:00Z",
      "incident_count": 2,
      "maintenance_count": 1
    }
  ]
}
```

#### Comprehensive Metrics
```http
GET /api/v1/tenant/monitoring/metrics?period=24h
```
Returns comprehensive monitoring metrics.

**Response:**
```json
{
  "metrics": {
    "period": "24h",
    "start_time": "2024-01-14T10:30:00Z",
    "end_time": "2024-01-15T10:30:00Z",
    "overall_uptime": 99.5,
    "services": [...],
    "uptime_history": [...],
    "response_time_history": [...],
    "incidents": [...],
    "maintenances": [...]
  }
}
```

## Frontend Interface

### Analytics Dashboard

#### Overview Section
- **Total Services**: Count of all monitored services
- **Average Uptime**: Overall system uptime percentage
- **Overall Status**: System health status (up/degraded/down)
- **Last Updated**: Data refresh timestamp

#### Analytics Tabs

**1. Uptime Tab**
- Overall uptime line chart
- Service uptime comparison bar chart
- Service uptime details table
- Historical uptime trends

**2. Response Time Tab**
- Average response time line chart
- Response time distribution chart
- Service response time details table
- Performance metrics (Min, Max, P95, P99)

**3. Errors Tab**
- Error rate over time chart
- Error types distribution chart
- Service error details table
- Error impact analysis

**4. Incidents Tab**
- Incidents timeline chart
- Incident impact analysis
- Recent incidents table
- Incident resolution tracking

#### Interactive Features

**Time Period Selection**
- 1 Hour: Real-time monitoring
- 24 Hours: Daily analysis
- 7 Days: Weekly trends
- 30 Days: Monthly patterns
- 90 Days: Quarterly analysis

**Service Detail Modal**
- Individual service deep-dive
- Detailed metrics and charts
- Historical performance data
- Incident and maintenance history

**Chart Interactions**
- Hover tooltips with detailed data
- Zoom and pan capabilities
- Data point selection
- Export functionality

### Chart Types

#### Line Charts
- **Uptime Trends**: Historical uptime over time
- **Response Time**: Performance trends
- **Error Rates**: Error rate patterns
- **Incident Timeline**: Incident occurrence over time

#### Bar Charts
- **Service Comparison**: Side-by-side service metrics
- **Uptime Comparison**: Service uptime percentages
- **Error Distribution**: Error types breakdown
- **Incident Impact**: Incident severity analysis

#### Area Charts
- **Uptime Visualization**: Filled uptime areas
- **Response Time Distribution**: Performance ranges
- **Error Rate Trends**: Error rate over time

## Data Visualization

### Chart.js Integration

The system uses Chart.js for beautiful, interactive charts:

```javascript
// Overall Uptime Chart
new Chart(ctx, {
    type: 'line',
    data: {
        labels: timeLabels,
        datasets: [{
            label: 'Uptime %',
            data: uptimeData,
            borderColor: '#4CAF50',
            backgroundColor: 'rgba(76, 175, 80, 0.1)',
            tension: 0.4
        }]
    },
    options: {
        responsive: true,
        scales: {
            y: {
                beginAtZero: false,
                min: 95,
                max: 100
            }
        }
    }
});
```

### Color Scheme

**Status Colors:**
- **Up**: Green (#4CAF50)
- **Degraded**: Orange (#FF9800)
- **Down**: Red (#F44336)
- **Unknown**: Gray (#9E9E9E)

**Chart Colors:**
- **Primary**: Blue (#2196F3)
- **Success**: Green (#4CAF50)
- **Warning**: Orange (#FF9800)
- **Error**: Red (#F44336)

### Responsive Design

- **Mobile-First**: Optimized for mobile devices
- **Tablet Support**: Responsive layouts for tablets
- **Desktop Enhanced**: Full-featured desktop experience
- **Touch Interactions**: Touch-friendly chart interactions

## Performance Features

### Real-time Updates
- **Live Data**: Real-time monitoring data
- **Auto Refresh**: Automatic data updates
- **WebSocket Support**: Live data streaming (future)
- **Caching**: Efficient data caching

### Data Aggregation
- **Time-based Aggregation**: Data grouped by time periods
- **Service Aggregation**: Multi-service metrics
- **Statistical Analysis**: Min, Max, Average, Percentiles
- **Trend Calculation**: Historical trend analysis

### Optimization
- **Lazy Loading**: Charts loaded on demand
- **Data Pagination**: Large datasets paginated
- **Chart Caching**: Cached chart configurations
- **Efficient Rendering**: Optimized chart rendering

## Integration

### Service Monitoring
- **Health Checks**: Regular service health monitoring
- **Performance Metrics**: Response time and throughput
- **Error Tracking**: Error rate and type monitoring
- **Incident Correlation**: Link incidents to service impact

### Database Integration
- **PostgreSQL**: Primary data storage
- **Time Series Data**: Efficient time-based queries
- **Indexing**: Optimized database indexes
- **Data Retention**: Configurable data retention policies

### External Monitoring
- **API Integration**: External monitoring service APIs
- **Webhook Support**: Real-time data ingestion
- **Custom Metrics**: User-defined monitoring metrics
- **Third-party Tools**: Integration with monitoring tools

## Configuration

### Monitoring Settings
```json
{
  "monitoring": {
    "data_retention_days": 90,
    "refresh_interval_seconds": 60,
    "chart_update_interval_seconds": 30,
    "max_data_points": 1000,
    "default_period": "24h"
  }
}
```

### Chart Configuration
```json
{
  "charts": {
    "uptime": {
      "min_value": 95,
      "max_value": 100,
      "color": "#4CAF50"
    },
    "response_time": {
      "min_value": 0,
      "max_value": 1000,
      "color": "#2196F3"
    },
    "error_rate": {
      "min_value": 0,
      "max_value": 10,
      "color": "#F44336"
    }
  }
}
```

## Usage Examples

### Basic Analytics Dashboard
```javascript
// Load analytics data
async function loadAnalytics() {
    const [summary, metrics, health] = await Promise.all([
        fetch('/api/v1/tenant/monitoring/uptime-summary'),
        fetch('/api/v1/tenant/monitoring/metrics?period=24h'),
        fetch('/api/v1/tenant/monitoring/services/health')
    ]);
    
    const summaryData = await summary.json();
    const metricsData = await metrics.json();
    const healthData = await health.json();
    
    renderAnalyticsDashboard(summaryData, metricsData, healthData);
}
```

### Service Detail View
```javascript
// View service details
function viewServiceDetails(serviceId) {
    const modal = document.getElementById('service-detail-modal');
    modal.style.display = 'block';
    
    // Load service-specific data
    loadServiceData(serviceId);
}
```

### Period Change Handler
```javascript
// Change time period
function changeAnalyticsPeriod() {
    const period = document.getElementById('analytics-period').value;
    
    // Reload data with new period
    loadAnalyticsData(period);
}
```

## Best Practices

### Data Visualization
1. **Clear Labels**: Use descriptive chart labels
2. **Consistent Colors**: Maintain consistent color scheme
3. **Appropriate Scales**: Set appropriate Y-axis scales
4. **Interactive Elements**: Add hover tooltips and interactions
5. **Responsive Design**: Ensure charts work on all devices

### Performance
1. **Data Pagination**: Paginate large datasets
2. **Chart Caching**: Cache chart configurations
3. **Lazy Loading**: Load charts on demand
4. **Efficient Queries**: Optimize database queries
5. **Real-time Updates**: Use efficient update mechanisms

### User Experience
1. **Loading States**: Show loading indicators
2. **Error Handling**: Handle data loading errors gracefully
3. **Empty States**: Show appropriate empty state messages
4. **Accessibility**: Ensure charts are accessible
5. **Mobile Optimization**: Optimize for mobile devices

## Troubleshooting

### Common Issues

#### Charts Not Loading
- Check Chart.js library is loaded
- Verify canvas elements exist
- Check for JavaScript errors
- Ensure data is properly formatted

#### Data Not Updating
- Check API endpoints are working
- Verify authentication tokens
- Check network connectivity
- Review data format compatibility

#### Performance Issues
- Reduce data points for large datasets
- Implement data pagination
- Use chart caching
- Optimize database queries

### Debug Mode
```javascript
// Enable debug mode
window.analyticsDebug = true;

// Debug chart initialization
if (window.analyticsDebug) {
    console.log('Initializing chart:', chartId);
    console.log('Chart data:', chartData);
}
```

## Future Enhancements

### Planned Features
- **Real-time WebSocket Updates**: Live data streaming
- **Advanced Analytics**: Machine learning insights
- **Custom Dashboards**: User-configurable dashboards
- **Alert Integration**: Automated alerting based on metrics
- **Export Functionality**: Export charts and data
- **Mobile App**: Native mobile application
- **API Documentation**: Interactive API documentation
- **Performance Optimization**: Advanced caching and optimization

### Integration Roadmap
- **Prometheus Integration**: Prometheus metrics integration
- **Grafana Compatibility**: Grafana dashboard compatibility
- **Slack Integration**: Slack notifications and updates
- **Email Reports**: Automated email reports
- **Webhook Support**: Custom webhook integrations
- **Third-party APIs**: Integration with external monitoring services
