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

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// promptModel is a simple text input model for prompts
type promptModel struct {
	textInput textinput.Model
	question  string
	err       error
}

func initialPromptModel(question string) promptModel {
	ti := textinput.New()
	ti.Placeholder = "Type your answer..."
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 50

	return promptModel{
		textInput: ti,
		question:  question,
		err:       nil,
	}
}

func (m promptModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			return m, tea.Quit
		}

	case error:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m promptModel) View() string {
	// Style for the question
	questionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	// Style for the input area
	inputStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		questionStyle.Render(m.question),
		inputStyle.Render(m.textInput.View()),
		"(Press Enter to submit, Esc to cancel)",
	)
}

// Prompt asks the user a question and returns their answer
func Prompt(question string) (string, error) {
	p := tea.NewProgram(initialPromptModel(question))
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	if m, ok := m.(promptModel); ok {
		return strings.TrimSpace(m.textInput.Value()), nil
	}

	return "", fmt.Errorf("unexpected model type")
}

// confirmModel is a simple yes/no confirmation model
type confirmModel struct {
	question string
	answer   bool
}

func initialConfirmModel(question string) confirmModel {
	return confirmModel{
		question: question,
		answer:   false,
	}
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.answer = true
			return m, tea.Quit
		case "n", "N":
			m.answer = false
			return m, tea.Quit
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	// Style for the question
	questionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	// Style for the options
	optionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	return fmt.Sprintf(
		"%s\n\n%s",
		questionStyle.Render(m.question),
		optionStyle.Render("(y/n)"),
	)
}

// Confirm asks the user a yes/no question
func Confirm(question string) (bool, error) {
	p := tea.NewProgram(initialConfirmModel(question))
	m, err := p.Run()
	if err != nil {
		return false, err
	}

	if m, ok := m.(confirmModel); ok {
		return m.answer, nil
	}

	return false, fmt.Errorf("unexpected model type")
}
