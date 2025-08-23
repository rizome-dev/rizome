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

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// listModel wraps bubbles list.Model to implement tea.Model
type listModel struct {
	list     list.Model
	selected list.Item
	finished bool
}

// NewListModel creates a new list selection model
func NewListModel(title string, items []list.Item) listModel {
	const defaultWidth = 80
	listHeight := min(len(items)*3+8, 20)

	l := list.New(items, list.NewDefaultDelegate(), defaultWidth, listHeight)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1).
		Bold(true)

	return listModel{
		list:     l,
		finished: false,
	}
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.finished = true
			return m, tea.Quit
		case "enter":
			m.selected = m.list.SelectedItem()
			m.finished = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listModel) View() string {
	return m.list.View()
}

// GetSelectedItem returns the selected item
func (m listModel) GetSelectedItem() list.Item {
	return m.selected
}

// IsFinished returns whether the selection is complete
func (m listModel) IsFinished() bool {
	return m.finished
}

// ListSelection runs a list selection and returns the selected item
func ListSelection(title string, items []list.Item) (list.Item, error) {
	model := NewListModel(title, items)
	p := tea.NewProgram(model)

	result, err := p.Run()
	if err != nil {
		return nil, err
	}

	if finalModel, ok := result.(listModel); ok {
		if !finalModel.IsFinished() || finalModel.GetSelectedItem() == nil {
			return nil, fmt.Errorf("no selection made")
		}
		return finalModel.GetSelectedItem(), nil
	}

	return nil, fmt.Errorf("unexpected model type")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
