---
Title: Sparkline Component
Slug: sparkline
Short: Terminal data visualization component for displaying trends in compact charts
Topics:
- components
- visualization  
- monitoring
- charts
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Sparkline Component

A flexible, feature-rich sparkline component for Go terminal applications. Sparklines are miniature charts that show data trends in a compact format - perfect for dashboards, monitoring tools, and CLI applications that need to display time-series data efficiently.

## What are Sparklines?

Sparklines are word-sized graphics that display quantitative information in a condensed format. They show the essential shape of data variation without traditional chart elements like axes, labels, or legends. Originally conceived by Edward Tufte, they excel at revealing patterns, trends, and outliers in datasets while occupying minimal screen real estate.

Think of sparklines as "data thumbnails" - they communicate the story of your data's behavior over time using just a few terminal characters.

## Core Features

The bobatea sparkline component provides a complete solution for terminal-based data visualization:

- **Multiple Visual Styles**: Bar charts, dot plots, line graphs, and filled area charts
- **Real-time Capability**: Sliding window behavior for live data streams
- **Smart Color Coding**: Value-based color ranges for instant visual alerts
- **Memory Management**: Automatic data pruning with configurable history limits
- **Bubble Tea Integration**: Native support for interactive terminal applications
- **Flexible Configuration**: Comprehensive styling and display options

## Quick Start

### Basic Static Chart

The simplest way to create a sparkline is to configure it with static data:

```go
package main

import (
    "fmt"
    "github.com/go-go-golems/bobatea/pkg/sparkline"
)

func main() {
    // Create a basic sparkline configuration
    config := sparkline.Config{
        Width:  30,
        Height: 4,
        Style:  sparkline.StyleBars,
        Title:  "CPU Usage",
    }
    
    s := sparkline.New(config)
    
    // Add data points representing CPU usage over time
    data := []float64{45, 67, 23, 89, 56, 78, 34, 90, 12, 67}
    s.SetData(data)
    
    fmt.Println(s.Render())
}
```

**Output:**
```
CPU Usage
▃▅▂▇▄▆▃█▁▅▃▆▄▇▅▆▃▇▂▅▃▆▄▇▅▆▃▇▂▅
```

### Real-time Data Streams

For monitoring applications, sparklines provide an efficient way to display live metrics:

```go
config := sparkline.Config{
    Width:     50,
    Height:    6,
    MaxPoints: 100,  // Retain last 100 measurements
    Style:     sparkline.StyleBars,
    Title:     "Network Traffic (MB/s)",
    ShowValue: true,
}

s := sparkline.New(config)

// Continuously update with new measurements
for {
    value := getCurrentNetworkTraffic()
    
    // New values automatically push out old ones
    s.AddPoint(value)
    
    fmt.Print("\033[H\033[2J") // Clear screen
    fmt.Println(s.Render())
    
    time.Sleep(1 * time.Second)
}
```

## Visual Styles

Choose the rendering style that best represents your data type and use case:

- **`StyleBars`**: Traditional bar chart using Unicode block characters - ideal for discrete measurements
- **`StyleDots`**: Scatter plot representation with positioned dots - good for sparse data or outlier detection
- **`StyleLine`**: Connected line chart with directional indicators - excellent for continuous trends
- **`StyleFilled`**: Area chart showing regions under the curve - emphasizes cumulative patterns

## Configuration Reference

The `Config` struct offers comprehensive control over sparkline appearance and behavior:

```go
config := sparkline.Config{
    // Chart dimensions
    Width:     50,    // Display width in terminal columns
    Height:    8,     // Display height in terminal rows
    MaxPoints: 200,   // History buffer size (0 = unlimited)
    
    // Visual presentation
    Style: sparkline.StyleBars,
    Title: "Server Response Time",
    
    // Information display
    ShowValue:  true,  // Display current value
    ShowMinMax: true,  // Show data range indicators
    
    // Color coding for value ranges
    ColorRanges: []sparkline.ColorRange{
        {Min: 0, Max: 100, Style: greenStyle},    // Normal range
        {Min: 100, Max: 200, Style: yellowStyle}, // Warning zone
        {Min: 200, Max: math.Inf(1), Style: redStyle}, // Critical levels
    },
    DefaultStyle: whiteStyle, // Fallback styling
}
```

## Color-Based Alerting

Sparklines support automatic color coding based on value thresholds, enabling instant visual feedback:

```go
import "github.com/charmbracelet/lipgloss"

// Define color styles for different alert levels
greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))   // Success
yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))  // Warning
redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))     // Critical

config := sparkline.Config{
    Width:  40,
    Height: 5,
    Style:  sparkline.StyleBars,
    Title:  "CPU Temperature (°C)",
    ShowValue: true,
    ColorRanges: []sparkline.ColorRange{
        {Min: 0, Max: 60, Style: greenStyle},      // Safe operating range
        {Min: 60, Max: 80, Style: yellowStyle},    // Elevated temperature
        {Min: 80, Max: math.Inf(1), Style: redStyle}, // Overheating risk
    },
}
```

This configuration automatically applies green styling for safe temperatures (0-60°C), yellow for elevated readings (60-80°C), and red for critical levels (80°C+).

## Bubble Tea Integration

Sparklines integrate seamlessly with interactive terminal applications built using Bubble Tea:

```go
type Model struct {
    sparkline *sparkline.Sparkline
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case TickMsg:
        // Fetch latest metric and update chart
        newValue := getMetric()
        m.sparkline.AddPoint(newValue)
        return m, m.tick()
    }
    return m, nil
}

func (m Model) View() string {
    return m.sparkline.View() // Renders the sparkline for display
}
```

## Memory Management

The sparkline component automatically manages memory consumption to support long-running applications:

- **Sliding Window**: When the data buffer exceeds `MaxPoints`, the oldest measurements are discarded
- **FIFO Behavior**: New data points push out historical ones, maintaining recent context
- **Bounded Storage**: Memory usage remains constant regardless of application runtime

This design makes sparklines suitable for monitoring applications that run indefinitely without memory leaks.

## Common Use Cases

Sparklines are particularly effective for:

**System Monitoring**
- CPU utilization, memory consumption, disk I/O rates
- Network throughput, connection counts, error frequencies

**Application Performance**  
- API response times, request rates, cache hit ratios
- Database query performance, queue depths, processing latencies

**Business Intelligence**
- Sales performance, user engagement, conversion funnels  
- Revenue trends, customer acquisition, retention metrics

**DevOps and Infrastructure**
- Build success rates, deployment frequencies, uptime percentages
- Log volume analysis, alert frequencies, service health scores

**IoT and Sensor Data**
- Environmental readings, device status, energy consumption
- Location tracking, usage patterns, anomaly detection

## Examples and Demos

The bobatea repository includes comprehensive examples demonstrating sparkline capabilities:

**Static Demo**
```bash
go run ./examples/sparkline-test demo
```

**Interactive TUI**  
```bash
go run ./examples/sparkline-test
```

The interactive demo showcases all visual styles with simulated real-time data generation, allowing experimentation with different configurations and update frequencies.

## Contributing

The sparkline component is part of the [bobatea](https://github.com/go-go-golems/bobatea) project, which welcomes contributions, bug reports, and feature requests. Please refer to the main repository for contribution guidelines and license information.
