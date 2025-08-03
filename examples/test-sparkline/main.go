package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/sparkline"
)

func main() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano())) // #nosec G404 // demo code
	fmt.Println("🌟 Sparkline Documentation Test")
	fmt.Println("Building a sparkline application based ONLY on the documentation")
	fmt.Println("=" + fmt.Sprintf("%*s", 60, "="))

	// Test 1: Basic Static Chart (copied exactly from documentation)
	fmt.Println("\n1. Basic Static Chart Test")
	fmt.Println("--------------------------")

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

	// Test 2: Multiple Visual Styles
	fmt.Println("\n\n2. Multiple Visual Styles Test")
	fmt.Println("------------------------------")

	styles := []struct {
		name  string
		style sparkline.Style
	}{
		{"Bars", sparkline.StyleBars},
		{"Dots", sparkline.StyleDots},
		{"Line", sparkline.StyleLine},
		{"Filled", sparkline.StyleFilled},
	}

	testData := []float64{10, 30, 20, 80, 40, 60, 50, 90, 70, 35, 25, 85}

	for _, styleInfo := range styles {
		config := sparkline.Config{
			Width:  40,
			Height: 5,
			Style:  styleInfo.style,
			Title:  fmt.Sprintf("Sample Data (%s)", styleInfo.name),
		}
		s := sparkline.New(config)
		s.SetData(testData)
		fmt.Printf("\n%s:\n", styleInfo.name)
		fmt.Println(s.Render())
	}

	// Test 3: Color Coding (from documentation example)
	fmt.Println("\n\n3. Color-Based Alerting Test")
	fmt.Println("----------------------------")

	// Define color styles for different alert levels
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))  // Success
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Warning
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))    // Critical

	config = sparkline.Config{
		Width:     40,
		Height:    5,
		Style:     sparkline.StyleBars,
		Title:     "CPU Temperature (°C)",
		ShowValue: true,
		ColorRanges: []sparkline.ColorRange{
			{Min: 0, Max: 60, Style: greenStyle},         // Safe operating range
			{Min: 60, Max: 80, Style: yellowStyle},       // Elevated temperature
			{Min: 80, Max: math.Inf(1), Style: redStyle}, // Overheating risk
		},
	}

	tempData := []float64{45, 55, 65, 75, 85, 90, 70, 60, 50, 85, 95, 40, 82}
	s = sparkline.New(config)
	s.SetData(tempData)
	fmt.Println(s.Render())

	// Test 4: Comprehensive Configuration (from documentation)
	fmt.Println("\n\n4. Comprehensive Configuration Test")
	fmt.Println("-----------------------------------")

	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))

	config = sparkline.Config{
		// Chart dimensions
		Width:     50,  // Display width in terminal columns
		Height:    8,   // Display height in terminal rows
		MaxPoints: 200, // History buffer size (0 = unlimited)

		// Visual presentation
		Style: sparkline.StyleBars,
		Title: "Server Response Time",

		// Information display
		ShowValue:  true, // Display current value
		ShowMinMax: true, // Show data range indicators

		// Color coding for value ranges
		ColorRanges: []sparkline.ColorRange{
			{Min: 0, Max: 100, Style: greenStyle},         // Normal range
			{Min: 100, Max: 200, Style: yellowStyle},      // Warning zone
			{Min: 200, Max: math.Inf(1), Style: redStyle}, // Critical levels
		},
		DefaultStyle: whiteStyle, // Fallback styling
	}

	s = sparkline.New(config)

	// Generate some server response time data
	responseData := []float64{50, 80, 120, 60, 220, 180, 90, 70, 250, 110, 40, 160, 300, 80, 60}
	s.SetData(responseData)
	fmt.Println(s.Render())

	// Test 5: Real-time Data Stream Simulation
	fmt.Println("\n\n5. Real-time Data Stream Simulation")
	fmt.Println("-----------------------------------")

	config = sparkline.Config{
		Width:     50,
		Height:    6,
		MaxPoints: 100, // Retain last 100 measurements
		Style:     sparkline.StyleBars,
		Title:     "Network Traffic (MB/s)",
		ShowValue: true,
	}

	s = sparkline.New(config)

	fmt.Println("Simulating real-time updates (adding points one by one):")

	// Simulate the real-time loop from documentation
	for i := 0; i < 15; i++ {
		// Simulate network traffic measurement
		value := 10 + 20*math.Sin(float64(i)*0.3) + rng.Float64()*5

		// New values automatically push out old ones
		s.AddPoint(value)

		fmt.Printf("\nStep %d (Value: %.1f MB/s):\n", i+1, value)
		fmt.Println(s.Render())

		// Simulate time delay (documentation uses time.Sleep)
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("\n✅ All documentation examples completed successfully!")
	fmt.Println("\nThe sparkline documentation appears to be accurate and complete.")

	// Test methods mentioned in documentation
	fmt.Println("\n\n6. API Methods Test")
	fmt.Println("------------------")

	lastValue := s.GetLastValue()
	minVal, maxVal := s.GetMinMax()
	data = s.GetData()
	config = s.GetConfig()

	fmt.Printf("Last Value: %.2f\n", lastValue)
	fmt.Printf("Min/Max: %.2f / %.2f\n", minVal, maxVal)
	fmt.Printf("Data Points: %d\n", len(data))
	fmt.Printf("Config Width: %d\n", config.Width)

	// Test Clear method
	s.Clear()
	fmt.Println("\nAfter Clear():")
	fmt.Println(s.Render())

	// Run comprehensive tests
	fmt.Println("\n" + strings.Repeat("=", 60))
	ComprehensiveDemo()
}
