# Advanced Diff Showcase

This example demonstrates advanced features and patterns for the bobatea diff component in a real-world infrastructure configuration management scenario.

## Features Demonstrated

### 🎯 **Complex Data Types**
- **JSON formatting**: Pretty-printed JSON with indentation
- **Size formatting**: Human-readable byte sizes (KB, MB, GB)
- **Duration formatting**: Time duration parsing and display
- **URL formatting**: Breaking URLs across lines for readability
- **Secret handling**: Automatic sensitive value detection and redaction

### 🏗️ **Infrastructure Simulation**
- **Multiple microservices**: API Gateway, User Service, Notification Service, Analytics Service
- **Various configuration categories**: Environment Variables, Resource Limits, Secrets, Database Config, etc.
- **Realistic change scenarios**: Memory limit increases, feature flag rollouts, secret rotations

### 📊 **Performance Testing**
- **Large dataset generation**: Use `--large` flag to generate hundreds of services
- **Stress testing**: Test search and filtering performance with large datasets
- **Memory efficiency**: Validate UI responsiveness with substantial data

### 🎨 **Advanced Customization**
- **Custom value formatters**: Different formatting strategies per data type
- **Themed output**: Support for different visual themes
- **Feature toggles**: Runtime flags to disable search/filters
- **Smart categorization**: Logical grouping of related configuration changes

## Usage

### Basic Usage
```bash
go run main.go
```

### Large Dataset Performance Testing
```bash
go run main.go --large
```
Generates 160+ services with 2000+ configuration changes to test UI performance.

### Feature Toggles
```bash
# Disable search functionality
go run main.go --no-search

# Disable status filters
go run main.go --no-filters

# Combine flags
go run main.go --large --no-search
```

### Theme Options
```bash
# Default theme (dark-friendly)
go run main.go --theme=default

# Light theme
go run main.go --theme=light

# Dark theme (explicit)
go run main.go --theme=dark
```

## Key Patterns Demonstrated

### 1. **Custom Value Formatting**
```go
// Different formatters based on data type
func (c *ConfigChange) formatValue(value any) any {
    switch c.dataType {
    case "json":
        // Pretty-print JSON
    case "size":
        // Human-readable bytes (512MB, 1.5GB)
    case "duration":
        // Parse and format durations
    case "url":
        // Break URLs for readability
    case "secret":
        // Mark as sensitive automatically
    }
}
```

### 2. **Realistic Data Generation**
The example simulates real infrastructure changes:
- Memory limit increases during scaling
- Feature flag rollouts with JSON configuration
- Secret rotations for security
- Database connection updates during migration
- Performance tuning parameter adjustments

### 3. **Large Dataset Handling**
Test the diff component's performance with:
- 160+ service instances
- 4 environments × 8 service types × 5 replicas
- 2000+ individual configuration changes
- Multiple categories per service

### 4. **Smart Categorization**
Configuration changes are logically grouped:
- **Environment Variables**: Runtime configuration
- **Resource Limits**: Memory, CPU, storage constraints
- **Secrets**: Sensitive credentials and keys
- **Database Configuration**: Connection strings, pools
- **Feature Flags**: JSON-formatted feature toggles
- **Caching**: Redis URLs, TTL settings

## Search Examples

Try these search queries to see advanced filtering:
- `memory` - Shows all memory-related changes
- `secret` - Filters to sensitive credential changes  
- `json` - Shows JSON configuration updates
- `postgresql` - Database connection changes
- `timeout` - Duration-related timeouts
- `api-gateway` - Specific service filtering

## Performance Notes

The `--large` flag generates a substantial dataset to validate:
- **Search performance**: Real-time filtering across 2000+ changes
- **Memory usage**: Efficient handling of large item collections  
- **Render performance**: Smooth scrolling and updates
- **Filter performance**: Status toggle responsiveness

This simulates real-world usage where infrastructure tools might display hundreds of configuration changes across multiple services and environments.

## Extension Ideas

This example can be extended to demonstrate:
- **Custom renderers**: Service-specific change visualization
- **Export functionality**: Generate deployment summaries
- **Integration**: Load data from Terraform, Kubernetes, Docker Compose
- **Validation**: Highlight potentially risky changes
- **Approval workflows**: Mark changes as reviewed/approved
