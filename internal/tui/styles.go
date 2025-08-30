package tui

// Copyright (C) 2025 Rizome Labs, Inc.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Color Palette - Using a modern, cohesive color scheme
const (
	// Primary Colors
	ColorAccent    = "#00D9FF" // Bright cyan for primary actions
	ColorPrimary   = "#874BFD" // Purple for highlights
	ColorSecondary = "#FF6B6B" // Coral for emphasis
	
	// UI Colors
	ColorSuccess   = "#10B981" // Green for success
	ColorWarning   = "#F59E0B" // Amber for warnings
	ColorError     = "#EF4444" // Red for errors
	ColorInfo      = "#3B82F6" // Blue for information
	
	// Neutral Colors
	ColorText      = "#E4E4E7" // Light gray for primary text
	ColorTextDim   = "#71717A" // Dimmed text
	ColorBorder    = "#3F3F46" // Border color
	ColorBg        = "#18181B" // Background
	ColorBgSubtle  = "#27272A" // Subtle background
)

// Unified Design System Styles
var (
	// Title Styles - For section headers
	TitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorAccent)).
		Bold(true).
		MarginBottom(1)
		
	// Subtitle style for secondary headers
	SubtitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText)).
		Italic(true)
	
	// Step indicator style (e.g., "Step 1/2")
	StepStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Background(lipgloss.Color(ColorBgSubtle)).
		Padding(0, 2).
		Bold(true).
		MarginBottom(1)

	// Focus indicator styles
	FocusedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorAccent)).
		Bold(true)
		
	UnfocusedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextDim))

	// Selection styles - Consistent across all components
	SelectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorAccent)).
		Bold(true)
		
	UnselectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText))
		
	// Item container styles
	ItemStyle = lipgloss.NewStyle().
		PaddingLeft(2)
		
	FocusedItemStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderForeground(lipgloss.Color(ColorAccent)).
		PaddingLeft(1)
		
	// Description styles
	DescriptionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextDim)).
		Italic(true).
		PaddingLeft(4)
		
	// Status indicator styles
	SuccessStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSuccess)).
		Bold(true)
		
	ErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorError)).
		Bold(true)
		
	WarningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorWarning)).
		Bold(true)
		
	InfoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorInfo))
		
	// Help text style
	HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextDim)).
		MarginTop(1)
		
	// Disabled style
	DisabledStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3F3F46")).
		Strikethrough(true)
		
	// Checkbox specific styles
	CheckedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSuccess))
		
	UncheckedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextDim))
		
	// Radio button specific styles  
	RadioSelectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary))
		
	RadioUnselectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextDim))
		
	// Box/Container styles
	BoxStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Padding(1, 2)
		
	// Progress indicator
	ProgressStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary))
		
	// Prompt styles
	PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Bold(true)
		
	// Input field style
	InputStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Padding(0, 1)
		
	FocusedInputStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorAccent)).
		Padding(0, 1)
		
	// Badge styles for status tags
	BadgeStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorBgSubtle)).
		Foreground(lipgloss.Color(ColorText)).
		Padding(0, 1).
		MarginLeft(1)
		
	BadgeSuccessStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorSuccess)).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		MarginLeft(1)
		
	BadgeWarningStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorWarning)).
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1).
		MarginLeft(1)
		
	// Cursor and selection indicators
	CursorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorAccent)).
		Bold(true)
		
	// List item styles (redefining for consistency)
	ListItemStyle = lipgloss.NewStyle().
		PaddingLeft(2)
		
	SelectedItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorAccent)).
		Bold(true).
		PaddingLeft(2)
		
	// Legacy style mappings for compatibility
	QuestionStyle = PromptStyle
	MutedStyle    = UnfocusedStyle
	BorderStyle   = BoxStyle
)

// Icons and symbols used consistently across components
const (
	IconCheck       = "✓"
	IconCross       = "✗"
	IconInfo        = "ℹ"
	IconWarning     = "⚠"
	IconArrowRight  = "→"
	IconArrowLeft   = "←"
	IconBullet      = "•"
	IconDot         = "·"
	
	// Selection indicators
	RadioEmpty      = "○"
	RadioFilled     = "●"
	CheckboxEmpty   = "☐"
	CheckboxFilled  = "☑"
	CheckboxPartial = "☒"
	
	// Cursor indicators
	CursorArrow     = "▶"
	CursorBlock     = "█"
	CursorLine      = "│"
)

// RenderCheckbox renders a checkbox with consistent styling
func RenderCheckbox(checked bool, label string, focused bool) string {
	var icon string
	var style lipgloss.Style
	
	if checked {
		icon = CheckboxFilled
		style = CheckedStyle
	} else {
		icon = CheckboxEmpty
		style = UncheckedStyle
	}
	
	if focused {
		style = style.Inherit(FocusedStyle)
	}
	
	return style.Render(fmt.Sprintf("%s %s", icon, label))
}

// RenderRadio renders a radio button with consistent styling
func RenderRadio(selected bool, label string, focused bool) string {
	var icon string
	var style lipgloss.Style
	
	if selected {
		icon = RadioFilled
		style = RadioSelectedStyle
	} else {
		icon = RadioEmpty
		style = RadioUnselectedStyle
	}
	
	if focused {
		style = style.Inherit(FocusedStyle)
	}
	
	return style.Render(fmt.Sprintf("%s %s", icon, label))
}

// RenderProgress renders a progress indicator
func RenderProgress(current, total int, label string) string {
	if total == 0 {
		return ""
	}
	
	percentage := float64(current) / float64(total) * 100
	
	// Create a visual progress bar
	barWidth := 20
	filled := int(float64(barWidth) * (percentage / 100))
	
	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "="
		} else {
			bar += "-"
		}
	}
	bar += "]"
	
	progressText := fmt.Sprintf("%s %.0f%% %s", bar, percentage, label)
	return ProgressStyle.Render(progressText)
}

// RenderStep renders a step indicator (e.g., "Step 1 of 2")
func RenderStep(current, total int, label string) string {
	stepText := fmt.Sprintf("Step %d of %d: %s", current, total, label)
	return StepStyle.Render(stepText)
}

// RenderBadge renders a status badge
func RenderBadge(text string, style lipgloss.Style) string {
	return style.Render(text)
}

// RenderHelp renders help text with consistent formatting
func RenderHelp(keys ...string) string {
	if len(keys) == 0 {
		return ""
	}
	
	helpText := ""
	for i, key := range keys {
		if i > 0 {
			helpText += fmt.Sprintf(" %s ", IconBullet)
		}
		helpText += key
	}
	
	return HelpStyle.Render(helpText)
}

// Legacy compatibility functions
func RenderCheckboxLegacy(checked bool, label string) string {
	return RenderCheckbox(checked, label, false)
}