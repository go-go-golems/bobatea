package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/sparkline"
)

// ComprehensiveDemo tests all sparkline features mentioned in the documentation
func ComprehensiveDemo() {
	fmt.Println("🔍 Comprehensive Sparkline Test")
	fmt.Println("Testing ALL features from documentation")
	fmt.Println("=====================================")

	// Test all four visual styles as documented
	testAllStyles()

	// Test color coding with different ranges
	testColorCoding()

	// Test configuration options
	testConfigurationOptions()

	// Test real-time behavior
	testRealTimeBehavior()

	// Test Bubble Tea integration
	testBubbleTeaIntegration()

	// Test memory management
	testMemoryManagement()
}

func testAllStyles() {
	fmt.Println("\n📊 Testing All Visual Styles")
	fmt.Println("-----------------------------")

	// Generate test data that shows patterns clearly
	data := generateTestPattern()

	styles := []struct {
		style sparkline.Style
		name  string
		desc  string
	}{
		{sparkline.StyleBars, "Bars", "Traditional bar chart using Unicode block characters"},
		{sparkline.StyleDots, "Dots", "Scatter plot representation with positioned dots"},
		{sparkline.StyleLine, "Line", "Connected line chart with directional indicators"},
		{sparkline.StyleFilled, "Filled", "Area chart showing regions under the curve"},
	}

	for _, s := range styles {
		config := sparkline.Config{
			Width:     50,
			Height:    6,
			Style:     s.style,
			Title:     fmt.Sprintf("%s Style", s.name),
			ShowValue: true,
		}

		chart := sparkline.New(config)
		chart.SetData(data)

		fmt.Printf("\n%s (%s):\n", s.name, s.desc)
		fmt.Println(chart.Render())
	}
}

func testColorCoding() {
	fmt.Println("\n🎨 Testing Color-Based Alerting")
	fmt.Println("-------------------------------")

	// Create styles as documented
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))  // Success
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Warning
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))    // Critical

	// Test different threshold configurations
	scenarios := []struct {
		name   string
		ranges []sparkline.ColorRange
		data   []float64
	}{
		{
			name: "CPU Temperature Monitoring",
			ranges: []sparkline.ColorRange{
				{Min: 0, Max: 60, Style: greenStyle},
				{Min: 60, Max: 80, Style: yellowStyle},
				{Min: 80, Max: math.Inf(1), Style: redStyle},
			},
			data: []float64{45, 55, 65, 75, 85, 90, 70, 60, 50, 85, 95, 40},
		},
		{
			name: "Response Time Alerts",
			ranges: []sparkline.ColorRange{
				{Min: 0, Max: 100, Style: greenStyle},
				{Min: 100, Max: 200, Style: yellowStyle},
				{Min: 200, Max: math.Inf(1), Style: redStyle},
			},
			data: []float64{50, 80, 150, 60, 250, 180, 90, 300, 70, 120, 40, 160},
		},
	}

	for _, scenario := range scenarios {
		config := sparkline.Config{
			Width:       45,
			Height:      5,
			Style:       sparkline.StyleBars,
			Title:       scenario.name,
			ShowValue:   true,
			ShowMinMax:  true,
			ColorRanges: scenario.ranges,
		}

		chart := sparkline.New(config)
		chart.SetData(scenario.data)

		fmt.Printf("\n%s:\n", scenario.name)
		fmt.Println(chart.Render())
	}
}

func testConfigurationOptions() {
	fmt.Println("\n⚙️  Testing Configuration Options")
	fmt.Println("--------------------------------")

	data := generateTestPattern()

	configs := []struct {
		name   string
		config sparkline.Config
	}{
		{
			name: "Minimal Configuration",
			config: sparkline.Config{
				Width:  20,
				Height: 3,
				Style:  sparkline.StyleBars,
			},
		},
		{
			name: "Maximum Features",
			config: sparkline.Config{
				Width:        60,
				Height:       8,
				MaxPoints:    100,
				Style:        sparkline.StyleLine,
				Title:        "Full Featured Chart",
				ShowValue:    true,
				ShowMinMax:   true,
				DefaultStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
			},
		},
		{
			name: "No Title/Values",
			config: sparkline.Config{
				Width:      30,
				Height:     4,
				Style:      sparkline.StyleFilled,
				ShowValue:  false,
				ShowMinMax: false,
			},
		},
	}

	for _, cfg := range configs {
		chart := sparkline.New(cfg.config)
		chart.SetData(data)

		fmt.Printf("\n%s:\n", cfg.name)
		fmt.Println(chart.Render())
	}
}

func testRealTimeBehavior() {
	fmt.Println("\n🔄 Testing Real-time Behavior")
	fmt.Println("-----------------------------")

	config := sparkline.Config{
		Width:     30,
		Height:    5,
		MaxPoints: 15, // Small buffer to demonstrate rolling window
		Style:     sparkline.StyleBars,
		Title:     "Rolling Window Demo",
		ShowValue: true,
	}

	chart := sparkline.New(config)

	fmt.Println("Demonstrating sliding window behavior:")
	fmt.Printf("MaxPoints: %d, adding 20 data points\n", config.MaxPoints)

	for i := 0; i < 20; i++ {
		value := 50 + 30*math.Sin(float64(i)*0.3) + rng.Float64()*10
		chart.AddPoint(value)

		data := chart.GetData()
		fmt.Printf("Point %2d: %.1f (buffer size: %d)\n", i+1, value, len(data))

		// Show chart every 5 points
		if (i+1)%5 == 0 {
			fmt.Println(chart.Render())
			fmt.Println()
		}
	}
}

func testBubbleTeaIntegration() {
	fmt.Println("\n🫧 Testing Bubble Tea Integration")
	fmt.Println("--------------------------------")

	config := sparkline.Config{
		Width:  40,
		Height: 5,
		Style:  sparkline.StyleBars,
		Title:  "Bubble Tea Model Test",
	}

	chart := sparkline.New(config)
	data := generateTestPattern()
	chart.SetData(data)

	// Test that sparkline implements tea.Model interface
	var model tea.Model = chart

	// Test Init
	cmd := model.Init()
	fmt.Printf("Init() returned: %v (should be nil)\n", cmd)

	// Test Update
	model, cmd = model.Update(nil)
	fmt.Printf("Update() preserved model and returned cmd: %v\n", cmd)

	// Test View
	view := model.View()
	fmt.Println("View() output:")
	fmt.Println(view)
}

func testMemoryManagement() {
	fmt.Println("\n💾 Testing Memory Management")
	fmt.Println("-----------------------------")

	config := sparkline.Config{
		Width:     20,
		Height:    4,
		MaxPoints: 10, // Small buffer for testing
		Style:     sparkline.StyleBars,
		Title:     "Memory Test",
		ShowValue: true,
	}

	chart := sparkline.New(config)

	fmt.Printf("MaxPoints: %d\n", config.MaxPoints)

	// Add more points than MaxPoints to test FIFO behavior
	for i := 0; i < 15; i++ {
		value := float64(i * 10)
		chart.AddPoint(value)

		data := chart.GetData()
		minVal, maxVal := chart.GetMinMax()

		fmt.Printf("Added %.0f: buffer size=%d, min=%.0f, max=%.0f\n",
			value, len(data), minVal, maxVal)
	}

	fmt.Println("\nFinal chart:")
	fmt.Println(chart.Render())

	// Test Clear functionality
	chart.Clear()
	data := chart.GetData()
	fmt.Printf("\nAfter Clear(): buffer size=%d\n", len(data))
	fmt.Println(chart.Render())
}

func generateTestPattern() []float64 {
	// Generate a pattern that shows interesting behavior in all chart types
	data := make([]float64, 30)
	for i := range data {
		// Combination of sine wave and random noise
		base := 50 + 30*math.Sin(float64(i)*0.3)
		noise := (rng.Float64() - 0.5) * 10
		data[i] = base + noise

		// Add some spikes
		if i%7 == 0 {
			data[i] += 20
		}
	}
	return data
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano())) // #nosec G404 // demo code
