// File: internal/tui/ui/theme.go
package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette — single source of truth for the TUI.
const (
	ColorPurpleDark  = "#5B3CC4"
	ColorPurple      = "#7D56F4"
	ColorPurpleLight = "#9F7AEA"
	ColorPurpleMuted = "#B794F6"

	ColorSurfaceDark  = "#1a1a2e"
	ColorSurfaceLight = "#2a2a3e"
	ColorSurfaceAlt   = "#1e1e30"

	ColorTextPrimary   = "#FAFAFA"
	ColorTextSecondary = "#888888"
	ColorTextDim       = "#555555"

	ColorStatusGreen = "#22c55e"
	ColorStatusAmber = "#f59e0b"
	ColorStatusRed   = "#ef4444"
	ColorError       = "#FF0000"
)

const (
	maxContentWidth = 90
	contentMargin   = 10
)

func ContentWidth(termWidth int) int {
	return max(10, min(maxContentWidth, termWidth-contentMargin))
}

var (
	BannerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorPurple)).
			PaddingLeft(2).
			PaddingRight(2)

	ContentBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorSurfaceLight))

	ModalBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorPurple)).
			PaddingLeft(1).
			PaddingRight(1).
			PaddingTop(1).
			PaddingBottom(1)

	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPurpleMuted)).
			Bold(true)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextDim))

	TabUnderlineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorPurple))

	TextPrimaryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextPrimary))

	TextSecondaryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextSecondary))

	TextDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextDim))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError))

	StatusGreenStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorStatusGreen))

	StatusRedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorStatusRed))

	SectionHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorPurpleLight)).
				Bold(true)

	ProviderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPurpleMuted))

	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextDim)).
				Bold(true)

	SelectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(ColorSurfaceLight))

	EvenRowStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorSurfaceDark))

	OddRowStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorSurfaceAlt))

	HintKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPurple))

	HintDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextDim))

	LabelTagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPurpleMuted)).
			Background(lipgloss.Color(ColorSurfaceLight)).
			PaddingLeft(1).
			PaddingRight(1)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPurple))

	BreadcrumbSepStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextDim))

	BreadcrumbActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorPurpleLight)).
				Bold(true)
)

var gradientColors = []lipgloss.Color{
	lipgloss.Color(ColorPurpleDark),
	lipgloss.Color(ColorPurple),
	lipgloss.Color(ColorPurpleLight),
	lipgloss.Color(ColorPurpleMuted),
}

func GradientText(text string, colors []lipgloss.Color) string {
	runes := []rune(text)
	n := len(runes)
	if n == 0 || len(colors) == 0 {
		return text
	}
	if len(colors) == 1 {
		return lipgloss.NewStyle().Foreground(colors[0]).Render(text)
	}

	var b strings.Builder
	for i, r := range runes {
		pos := float64(i) / float64(max(1, n-1)) * float64(len(colors)-1)
		idx := min(int(pos), len(colors)-1)
		style := lipgloss.NewStyle().Foreground(colors[idx])
		b.WriteString(style.Render(string(r)))
	}
	return b.String()
}
