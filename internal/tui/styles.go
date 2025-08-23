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

// Shared color scheme
const (
	ColorPrimary   = "62"  // Blue for titles and highlights
	ColorSecondary = "205" // Pink for questions and emphasis
	ColorSuccess   = "34"  // Green for success messages
	ColorError     = "196" // Red for errors
	ColorInfo      = "33"  // Yellow for info
	ColorMuted     = "240" // Gray for muted text
	ColorText      = "230" // Light text for titles
)

// Common styles used across the application
var (
	// Title style for list headers and main titles
	TitleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorPrimary)).
			Foreground(lipgloss.Color(ColorText)).
			Padding(0, 1).
			Bold(true)

	// Question style for prompts
	QuestionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorSecondary))

	// Success message style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSuccess)).
			Bold(true)

	// Error message style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError)).
			Bold(true)

	// Info message style
	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorInfo)).
			Bold(true)

	// Muted text style for hints and secondary info
	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted))

	// Input box style
	InputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorPrimary)).
			Padding(0, 1)

	// List item style
	ListItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	// Selected list item style
	SelectedItemStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(ColorPrimary)).
				Foreground(lipgloss.Color(ColorText)).
				PaddingLeft(2)

	// Description style for list items
	DescriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorMuted)).
				PaddingLeft(2)

	// Checkbox style - checked
	CheckboxCheckedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorSuccess)).
				Bold(true)

	// Checkbox style - unchecked
	CheckboxUncheckedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorMuted))

	// Progress bar style
	ProgressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPrimary))

	// Border style for containers
	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorPrimary)).
			Padding(1)
)

// RenderCheckbox renders a checkbox with the given checked state and label
func RenderCheckbox(checked bool, label string) string {
	checkbox := "☐"
	style := CheckboxUncheckedStyle

	if checked {
		checkbox = "☑"
		style = CheckboxCheckedStyle
	}

	return style.Render(checkbox + " " + label)
}

// RenderProgress renders a simple progress indicator
func RenderProgress(current, total int, label string) string {
	if total == 0 {
		return ""
	}

	percentage := float64(current) / float64(total) * 100
	progressText := fmt.Sprintf("[%d/%d] %.0f%% %s", current, total, percentage, label)

	return ProgressStyle.Render(progressText)
}
