//nolint:gochecknoglobals
package tui

import "github.com/charmbracelet/lipgloss"

const (
	ANSIColorWhite  = lipgloss.Color("15")
	ANSIColorCyan   = lipgloss.Color("14")
	ANSIColorPurple = lipgloss.Color("13")
	ANSIColorBlue   = lipgloss.Color("12")
	ANSIColorYellow = lipgloss.Color("11")
	ANSIColorGreen  = lipgloss.Color("10")
	ANSIColorRed    = lipgloss.Color("9")
	ANSIColorGray   = lipgloss.Color("8")
)

const (
	IconInfo = "ℹ️"
	IconWarn = "⚠"
)

var Info = lipgloss.NewStyle().
	Foreground(ANSIColorCyan).Render

var Warn = lipgloss.NewStyle().
	Foreground(ANSIColorYellow).Render
