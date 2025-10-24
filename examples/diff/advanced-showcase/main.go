package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/diff"
)

// InfraProvider simulates infrastructure configuration changes
type InfraProvider struct {
	title string
	items []diff.DiffItem
}

func (p *InfraProvider) Title() string          { return p.title }
func (p *InfraProvider) Items() []diff.DiffItem { return p.items }

// ServiceItem represents a microservice with configuration changes
type ServiceItem struct {
	id         string
	name       string
	categories []diff.Category
}

func (s *ServiceItem) ID() string                  { return s.id }
func (s *ServiceItem) Name() string                { return s.name }
func (s *ServiceItem) Categories() []diff.Category { return s.categories }

// ConfigCategory represents a category of configuration changes
type ConfigCategory struct {
	name    string
	changes []diff.Change
}

func (c *ConfigCategory) Name() string           { return c.name }
func (c *ConfigCategory) Changes() []diff.Change { return c.changes }

// ConfigChange represents a single configuration change with advanced formatting
type ConfigChange struct {
	path        string
	status      diff.ChangeStatus
	beforeValue any
	afterValue  any
	sensitive   bool
	dataType    string // "json", "yaml", "size", "duration", "url", "secret"
}

func (c *ConfigChange) Path() string              { return c.path }
func (c *ConfigChange) Status() diff.ChangeStatus { return c.status }
func (c *ConfigChange) Before() any               { return c.formatValue(c.beforeValue) }
func (c *ConfigChange) After() any                { return c.formatValue(c.afterValue) }
func (c *ConfigChange) Sensitive() bool           { return c.sensitive }

// Advanced value formatting based on data type
func (c *ConfigChange) formatValue(value any) any {
	if value == nil {
		return nil
	}

	switch c.dataType {
	case "json":
		if str, ok := value.(string); ok {
			var obj any
			if json.Unmarshal([]byte(str), &obj) == nil {
				if formatted, err := json.MarshalIndent(obj, "", "  "); err == nil {
					return string(formatted)
				}
			}
		}
	case "size":
		if num, ok := value.(int64); ok {
			return formatBytes(num)
		}
	case "duration":
		if str, ok := value.(string); ok {
			if d, err := time.ParseDuration(str); err == nil {
				return d.String()
			}
		}
	case "url":
		if str, ok := value.(string); ok {
			return strings.ReplaceAll(str, "://", "://\n  ")
		}
	case "secret":
		// Always mark secrets as sensitive
		c.sensitive = true
	}

	return value
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func main() {
	var (
		large     = flag.Bool("large", false, "generate large dataset for performance testing")
		noSearch  = flag.Bool("no-search", false, "disable search functionality")
		noFilters = flag.Bool("no-filters", false, "disable status filters")
		theme     = flag.String("theme", "default", "color theme: default, dark, light")
	)
	flag.Parse()

	var items []diff.DiffItem
	if *large {
		items = generateLargeDataset()
	} else {
		items = generateAdvancedDataset()
	}

	provider := &InfraProvider{
		title: "Infrastructure Configuration Changes",
		items: items,
	}

	config := diff.DefaultConfig()
	config.Title = "🚀 Infrastructure Deployment Preview"

	// Apply theme
	switch *theme {
	case "dark":
		// Keep default - already dark-friendly
	case "light":
		config.Title = "☀️ " + config.Title
	}

	// Apply flags
	if *noSearch {
		config.EnableSearch = false
	}
	if *noFilters {
		config.EnableStatusFilters = false
	}

	model := diff.NewModel(provider, config)

	program := tea.NewProgram(model, tea.WithContext(context.Background()), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}

// generateAdvancedDataset creates a comprehensive example with various data types
func generateAdvancedDataset() []diff.DiffItem {
	return []diff.DiffItem{
		&ServiceItem{
			id:   "api-gateway",
			name: "api-gateway",
			categories: []diff.Category{
				&ConfigCategory{
					name: "Environment Variables",
					changes: []diff.Change{
						&ConfigChange{
							path:        "LOG_LEVEL",
							status:      diff.ChangeStatusUpdated,
							beforeValue: "info",
							afterValue:  "debug",
							dataType:    "string",
						},
						&ConfigChange{
							path:        "RATE_LIMIT_PER_MINUTE",
							status:      diff.ChangeStatusUpdated,
							beforeValue: 1000,
							afterValue:  2000,
							dataType:    "number",
						},
						&ConfigChange{
							path:        "API_TIMEOUT",
							status:      diff.ChangeStatusUpdated,
							beforeValue: "30s",
							afterValue:  "45s",
							dataType:    "duration",
						},
					},
				},
				&ConfigCategory{
					name: "Resource Limits",
					changes: []diff.Change{
						&ConfigChange{
							path:        "memory_limit",
							status:      diff.ChangeStatusUpdated,
							beforeValue: int64(536870912),  // 512MB
							afterValue:  int64(1073741824), // 1GB
							dataType:    "size",
						},
						&ConfigChange{
							path:        "cpu_limit",
							status:      diff.ChangeStatusUpdated,
							beforeValue: "0.5",
							afterValue:  "1.0",
							dataType:    "string",
						},
					},
				},
				&ConfigCategory{
					name: "Secrets",
					changes: []diff.Change{
						&ConfigChange{
							path:        "JWT_SECRET",
							status:      diff.ChangeStatusUpdated,
							beforeValue: "old-jwt-secret-key-here",
							afterValue:  "new-jwt-secret-key-rotated",
							sensitive:   true,
							dataType:    "secret",
						},
						&ConfigChange{
							path:       "DATABASE_PASSWORD",
							status:     diff.ChangeStatusAdded,
							afterValue: "super-secure-password-123",
							sensitive:  true,
							dataType:   "secret",
						},
					},
				},
			},
		},
		&ServiceItem{
			id:   "user-service",
			name: "user-service",
			categories: []diff.Category{
				&ConfigCategory{
					name: "Database Configuration",
					changes: []diff.Change{
						&ConfigChange{
							path:        "database_url",
							status:      diff.ChangeStatusUpdated,
							beforeValue: "postgresql://localhost:5432/users",
							afterValue:  "postgresql://db-cluster:5432/users",
							dataType:    "url",
						},
						&ConfigChange{
							path:        "connection_pool_size",
							status:      diff.ChangeStatusUpdated,
							beforeValue: 10,
							afterValue:  20,
							dataType:    "number",
						},
					},
				},
				&ConfigCategory{
					name: "Feature Flags",
					changes: []diff.Change{
						&ConfigChange{
							path:        "features",
							status:      diff.ChangeStatusUpdated,
							beforeValue: `{"new_auth": false, "email_verification": true, "rate_limiting": false}`,
							afterValue:  `{"new_auth": true, "email_verification": true, "rate_limiting": true, "social_login": true}`,
							dataType:    "json",
						},
					},
				},
				&ConfigCategory{
					name: "Caching",
					changes: []diff.Change{
						&ConfigChange{
							path:       "redis_url",
							status:     diff.ChangeStatusAdded,
							afterValue: "redis://cache-cluster:6379/0",
							dataType:   "url",
						},
						&ConfigChange{
							path:       "cache_ttl",
							status:     diff.ChangeStatusAdded,
							afterValue: "24h",
							dataType:   "duration",
						},
					},
				},
			},
		},
		&ServiceItem{
			id:   "notification-service",
			name: "notification-service",
			categories: []diff.Category{
				&ConfigCategory{
					name: "Email Configuration",
					changes: []diff.Change{
						&ConfigChange{
							path:        "SMTP_HOST",
							status:      diff.ChangeStatusUpdated,
							beforeValue: "smtp.local.dev",
							afterValue:  "smtp.production.com",
							dataType:    "string",
						},
						&ConfigChange{
							path:        "EMAIL_API_KEY",
							status:      diff.ChangeStatusUpdated,
							beforeValue: "dev-key-12345",
							afterValue:  "prod-key-67890",
							sensitive:   true,
							dataType:    "secret",
						},
					},
				},
				&ConfigCategory{
					name: "Queue Configuration",
					changes: []diff.Change{
						&ConfigChange{
							path:        "queue_size",
							status:      diff.ChangeStatusUpdated,
							beforeValue: 100,
							afterValue:  500,
							dataType:    "number",
						},
						&ConfigChange{
							path:        "worker_timeout",
							status:      diff.ChangeStatusUpdated,
							beforeValue: "5m",
							afterValue:  "10m",
							dataType:    "duration",
						},
					},
				},
				&ConfigCategory{
					name: "Storage",
					changes: []diff.Change{
						&ConfigChange{
							path:        "max_attachment_size",
							status:      diff.ChangeStatusUpdated,
							beforeValue: int64(10485760), // 10MB
							afterValue:  int64(52428800), // 50MB
							dataType:    "size",
						},
					},
				},
			},
		},
		&ServiceItem{
			id:   "analytics-service",
			name: "analytics-service",
			categories: []diff.Category{
				&ConfigCategory{
					name: "Data Processing",
					changes: []diff.Change{
						&ConfigChange{
							path:       "pipeline_config",
							status:     diff.ChangeStatusAdded,
							afterValue: `{"batch_size": 1000, "parallel_workers": 4, "retry_attempts": 3, "output_format": "parquet"}`,
							dataType:   "json",
						},
						&ConfigChange{
							path:       "data_retention",
							status:     diff.ChangeStatusAdded,
							afterValue: "2160h", // 90 days
							dataType:   "duration",
						},
					},
				},
				&ConfigCategory{
					name: "Machine Learning",
					changes: []diff.Change{
						&ConfigChange{
							path:       "model_memory_limit",
							status:     diff.ChangeStatusAdded,
							afterValue: int64(8589934592), // 8GB
							dataType:   "size",
						},
						&ConfigChange{
							path:       "training_timeout",
							status:     diff.ChangeStatusAdded,
							afterValue: "6h",
							dataType:   "duration",
						},
					},
				},
			},
		},
	}
}

// generateLargeDataset creates a large dataset for performance testing
func generateLargeDataset() []diff.DiffItem {
	var items []diff.DiffItem

	serviceTypes := []string{"api", "worker", "scheduler", "proxy", "cache", "db", "queue", "monitor"}
	environments := []string{"staging", "production", "development", "testing"}

	itemCount := 0
	for _, env := range environments {
		for i, serviceType := range serviceTypes {
			for replica := 1; replica <= 5; replica++ {
				itemCount++
				serviceName := fmt.Sprintf("%s-%s-%02d", serviceType, env, replica)

				var categories []diff.Category

				// Environment Variables Category
				var envChanges []diff.Change
				for j := 0; j < 10; j++ {
					var status diff.ChangeStatus
					switch j % 3 {
					case 0:
						status = diff.ChangeStatusAdded
					case 1:
						status = diff.ChangeStatusRemoved
					default:
						status = diff.ChangeStatusUpdated
					}
					envChanges = append(envChanges, &ConfigChange{
						path:        fmt.Sprintf("VAR_%d", j),
						status:      status,
						beforeValue: fmt.Sprintf("old_value_%d", j),
						afterValue:  fmt.Sprintf("new_value_%d", j),
						dataType:    "string",
					})
				}
				categories = append(categories, &ConfigCategory{
					name:    "Environment Variables",
					changes: envChanges,
				})

				// Resource Limits Category
				var resourceChanges []diff.Change
				for j := 0; j < 5; j++ {
					resourceChanges = append(resourceChanges, &ConfigChange{
						path:        fmt.Sprintf("resource_%d", j),
						status:      diff.ChangeStatusUpdated,
						beforeValue: int64((j + 1) * 1024 * 1024 * 512),  // Various MB sizes
						afterValue:  int64((j + 1) * 1024 * 1024 * 1024), // Various GB sizes
						dataType:    "size",
					})
				}
				categories = append(categories, &ConfigCategory{
					name:    "Resource Limits",
					changes: resourceChanges,
				})

				// Secrets Category (every 3rd service)
				if i%3 == 0 {
					var secretChanges []diff.Change
					for j := 0; j < 3; j++ {
						secretChanges = append(secretChanges, &ConfigChange{
							path:        fmt.Sprintf("SECRET_%d", j),
							status:      diff.ChangeStatusUpdated,
							beforeValue: fmt.Sprintf("old-secret-%d-%s", j, serviceName),
							afterValue:  fmt.Sprintf("new-secret-%d-%s", j, serviceName),
							sensitive:   true,
							dataType:    "secret",
						})
					}
					categories = append(categories, &ConfigCategory{
						name:    "Secrets",
						changes: secretChanges,
					})
				}

				items = append(items, &ServiceItem{
					id:         serviceName,
					name:       serviceName,
					categories: categories,
				})
			}
		}
	}

	log.Printf("Generated %d services with large dataset", itemCount)
	return items
}
