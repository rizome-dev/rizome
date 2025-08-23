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
}

// NewCheckboxModel creates a new checkbox selection model
func NewCheckboxModel(title string, options []CheckboxOption) checkboxModel {
	return checkboxModel{
		title:   title,
		options: options,
		cursor:  0,
	}
}

func (m checkboxModel) Init() tea.Cmd {
	return nil
}

func (m checkboxModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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

		case " ":
			// Toggle checkbox
			if m.cursor >= 0 && m.cursor < len(m.options) {
				m.options[m.cursor].Checked = !m.options[m.cursor].Checked
			}

		case "a":
			// Select all
			for i := range m.options {
				m.options[i].Checked = true
			}

		case "A":
			// Deselect all
			for i := range m.options {
				m.options[i].Checked = false
			}
		}
	}

	return m, nil
}

func (m checkboxModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render(m.title))
	b.WriteString("\n\n")

	// Options
	for i, option := range m.options {
		// Cursor indicator
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		// Checkbox
		checkbox := RenderCheckbox(option.Checked, option.Label)

		// Style based on cursor position
		if m.cursor == i {
			line := fmt.Sprintf("%s %s", cursor, checkbox)
			b.WriteString(SelectedItemStyle.Render(line))
		} else {
			line := fmt.Sprintf("%s %s", cursor, checkbox)
			b.WriteString(ListItemStyle.Render(line))
		}

		// Description if available
		if option.Description != "" {
			b.WriteString("\n")
			description := fmt.Sprintf("   %s", option.Description)
			b.WriteString(DescriptionStyle.Render(description))
		}

		b.WriteString("\n")
	}

	// Instructions
	b.WriteString("\n")
	instructions := MutedStyle.Render("Space: toggle • a: select all • A: deselect all • Enter: confirm • q: quit")
	b.WriteString(instructions)

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

// IsConfirmed returns true if the user confirmed the selection
func (m checkboxModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns true if the user cancelled the selection
func (m checkboxModel) IsCancelled() bool {
	return m.cancelled
}

// CheckboxSelection runs a checkbox selection and returns the selected values
func CheckboxSelection(title string, options []CheckboxOption) ([]string, error) {
	model := NewCheckboxModel(title, options)
	p := tea.NewProgram(model)

	result, err := p.Run()
	if err != nil {
		return nil, err
	}

	if finalModel, ok := result.(checkboxModel); ok {
		if finalModel.IsCancelled() {
			return nil, fmt.Errorf("selection cancelled")
		}
		return finalModel.GetSelectedValues(), nil
	}

	return nil, fmt.Errorf("unexpected model type")
}
