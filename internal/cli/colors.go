package cli

import "fmt"

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorAccent = "\033[38;2;255;79;0m" // #ff4f00
	colorGray   = "\033[90m"
)

// Accent colors text with the brand accent (#ff4f00)
func accent(s string) string {
	return colorAccent + s + colorReset
}

// Gray dims text
func gray(s string) string {
	return colorGray + s + colorReset
}

// ProgressBar creates a progress bar
func progressBar(width int, percent float64) string {
	filled := int(float64(width) * percent)
	empty := width - filled

	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}

	// Only the filled portion is accented
	if filled > 0 {
		filledBar := ""
		for i := 0; i < filled; i++ {
			filledBar += "█"
		}
		emptyBar := ""
		for i := 0; i < empty; i++ {
			emptyBar += "░"
		}
		return accent(filledBar) + emptyBar
	}

	return bar
}

// FullProgressBar creates a full progress bar (100%)
func fullProgressBar(width int) string {
	bar := ""
	for i := 0; i < width; i++ {
		bar += "█"
	}
	return accent(bar)
}
