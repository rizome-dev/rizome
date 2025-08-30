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
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CheckboxOption represents a single checkbox option
type CheckboxOption struct {
	Label       string
	Description string
	Value       string
	Checked     bool
}

// checkboxModel handles checkbox selection for multiple items
type checkboxModel struct {
	title     string
	options   []CheckboxOption
	cursor    int
	confirmed bool
	cancelled bool
	width     int
	height    int
}

// NewCheckboxModel creates a new checkbox selection model
func NewCheckboxModel(title string, options []CheckboxOption) checkboxModel {
	return checkboxModel{
		title:   title,
		options: options,
		cursor:  0,
		width:   80, // Default width
		height:  24, // Default height
	}
}

func (m checkboxModel) Init() tea.Cmd {
	return nil
}

func (m checkboxModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.cancelled = true
			return m, tea.Quit

		case "enter":
			m.confirmed = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}

		case " ", "x":
			// Toggle checkbox
			if m.cursor >= 0 && m.cursor < len(m.options) {
				m.options[m.cursor].Checked = !m.options[m.cursor].Checked
			}

		case "a":
			// Select all
			for i := range m.options {
				m.options[i].Checked = true
			}

		case "A", "n":
			// Deselect all (n for none)
			for i := range m.options {
				m.options[i].Checked = false
			}
			
		case "i":
			// Invert selection
			for i := range m.options {
				m.options[i].Checked = !m.options[i].Checked
			}
		}
	}

	return m, nil
}

func (m checkboxModel) View() string {
	if m.confirmed || m.cancelled {
		return ""
	}
	
	var b strings.Builder

	// Title with consistent styling
	b.WriteString(TitleStyle.Render(m.title))
	b.WriteString("\n\n")

	// Calculate how many items we can display
	availableHeight := m.height - 8 // Reserve space for title, help, etc.
	itemsToShow := len(m.options)
	startIdx := 0
	
	// Implement scrolling if needed
	if availableHeight < len(m.options)*2 { // Approximate 2 lines per item
		itemsToShow = availableHeight / 2
		// Center the cursor in the visible area
		if m.cursor > itemsToShow/2 {
			startIdx = m.cursor - itemsToShow/2
			if startIdx+itemsToShow > len(m.options) {
				startIdx = len(m.options) - itemsToShow
			}
		}
	}

	// Render visible options
	for i := startIdx; i < startIdx+itemsToShow && i < len(m.options); i++ {
		option := m.options[i]
		isFocused := m.cursor == i
		
		// Build the item line
		var itemLine strings.Builder
		
		// Cursor indicator
		if isFocused {
			itemLine.WriteString(CursorStyle.Render(CursorArrow))
			itemLine.WriteString(" ")
		} else {
			itemLine.WriteString("  ")
		}
		
		// Checkbox and label
		var checkboxIcon string
		if option.Checked {
			checkboxIcon = CheckboxFilled
		} else {
			checkboxIcon = CheckboxEmpty
		}
		
		// Apply appropriate styling
		var lineStyle lipgloss.Style
		if isFocused {
			if option.Checked {
				lineStyle = CheckedStyle.Copy().Bold(true)
			} else {
				lineStyle = FocusedStyle
			}
		} else {
			if option.Checked {
				lineStyle = CheckedStyle
			} else {
				lineStyle = UnselectedStyle
			}
		}
		
		itemLine.WriteString(lineStyle.Render(fmt.Sprintf("%s %s", checkboxIcon, option.Label)))
		
		// Add status badge if checked
		if option.Checked {
			badge := BadgeSuccessStyle.Render("selected")
			itemLine.WriteString(" ")
			itemLine.WriteString(badge)
		}
		
		b.WriteString(itemLine.String())
		b.WriteString("\n")
		
		// Description on the next line if available and item is focused
		if option.Description != "" && isFocused {
			desc := DescriptionStyle.Copy().PaddingLeft(4).Render(option.Description)
			
			// Wrap description if too long
			maxWidth := m.width - 6
			if maxWidth > 0 && lipgloss.Width(desc) > maxWidth {
				desc = DescriptionStyle.Copy().
					PaddingLeft(4).
					Width(maxWidth).
					Render(option.Description)
			}
			
			b.WriteString(desc)
			b.WriteString("\n")
		}
	}
	
	// Show scroll indicators if needed
	if startIdx > 0 || startIdx+itemsToShow < len(m.options) {
		scrollInfo := MutedStyle.Render(fmt.Sprintf("\n[%d-%d of %d items]", 
			startIdx+1, 
			min(startIdx+itemsToShow, len(m.options)), 
			len(m.options)))
		b.WriteString(scrollInfo)
	}

	// Selected count summary
	selectedCount := 0
	for _, opt := range m.options {
		if opt.Checked {
			selectedCount++
		}
	}
	
	b.WriteString("\n\n")
	summary := fmt.Sprintf("%d of %d selected", selectedCount, len(m.options))
	b.WriteString(InfoStyle.Render(summary))

	// Help text
	b.WriteString("\n")
	helpText := RenderHelp(
		"space: toggle",
		"a: all",
		"n: none", 
		"i: invert",
		"enter: confirm",
		"q: cancel",
	)
	b.WriteString(helpText)

	return b.String()
}

// GetSelectedValues returns the values of all checked options
func (m checkboxModel) GetSelectedValues() []string {
	var selected []string
	for _, option := range m.options {
		if option.Checked {
			selected = append(selected, option.Value)
		}
	}
	return selected
}

// CheckboxSelection displays a checkbox selection interface and returns selected values
func CheckboxSelection(title string, options []CheckboxOption) ([]string, error) {
	m := NewCheckboxModel(title, options)

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	final := finalModel.(checkboxModel)
	if final.cancelled {
		return nil, fmt.Errorf("selection cancelled")
	}

	return final.GetSelectedValues(), nil
}

