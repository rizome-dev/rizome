package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RadioOption represents a single radio button option
type RadioOption struct {
	Label       string
	Description string
	Value       string
	Disabled    bool
}

// radioModel represents the state of the radio button selection
type radioModel struct {
	options     []RadioOption
	selected    int
	title       string
	quitting    bool
	aborted     bool
	width       int
	height      int
}

// radioKeyMap defines the key bindings for the radio selection
type radioKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Quit   key.Binding
}

var radioKeys = radioKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("enter/space", "select"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q/esc", "quit"),
	),
}

func (m radioModel) Init() tea.Cmd {
	return nil
}

func (m radioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, radioKeys.Quit):
			m.quitting = true
			m.aborted = true
			return m, tea.Quit

		case key.Matches(msg, radioKeys.Up):
			// Move up, skipping disabled options
			for i := m.selected - 1; i >= 0; i-- {
				if !m.options[i].Disabled {
					m.selected = i
					break
				}
			}

		case key.Matches(msg, radioKeys.Down):
			// Move down, skipping disabled options
			for i := m.selected + 1; i < len(m.options); i++ {
				if !m.options[i].Disabled {
					m.selected = i
					break
				}
			}

		case key.Matches(msg, radioKeys.Select):
			if !m.options[m.selected].Disabled {
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m radioModel) View() string {
	if m.quitting {
		if m.aborted {
			return ""
		}
		// Show the selected option when confirmed
		return fmt.Sprintf("%s Selected: %s\n", 
			SuccessStyle.Render(IconCheck),
			m.options[m.selected].Label)
	}

	var s strings.Builder
	
	// Title with consistent styling
	s.WriteString(TitleStyle.Render(m.title))
	s.WriteString("\n\n")

	// Calculate how many items we can display
	availableHeight := m.height - 8 // Reserve space for title, help, etc.
	itemsToShow := len(m.options)
	startIdx := 0
	
	// Implement scrolling if needed
	if availableHeight < len(m.options)*2 { // Approximate 2 lines per item
		itemsToShow = availableHeight / 2
		// Center the cursor in the visible area
		if m.selected > itemsToShow/2 {
			startIdx = m.selected - itemsToShow/2
			if startIdx+itemsToShow > len(m.options) {
				startIdx = len(m.options) - itemsToShow
			}
		}
	}

	// Render visible options
	for i := startIdx; i < startIdx+itemsToShow && i < len(m.options); i++ {
		option := m.options[i]
		isFocused := m.selected == i
		
		// Build the item line
		var itemLine strings.Builder
		
		// Cursor indicator
		if isFocused {
			itemLine.WriteString(CursorStyle.Render(CursorArrow))
			itemLine.WriteString(" ")
		} else {
			itemLine.WriteString("  ")
		}
		
		// Radio button and label
		var radioIcon string
		if isFocused {
			radioIcon = RadioFilled
		} else {
			radioIcon = RadioEmpty
		}
		
		// Apply appropriate styling
		var lineStyle lipgloss.Style
		if option.Disabled {
			lineStyle = DisabledStyle
		} else if isFocused {
			lineStyle = RadioSelectedStyle.Copy().Bold(true)
		} else {
			lineStyle = UnselectedStyle
		}
		
		itemLine.WriteString(lineStyle.Render(fmt.Sprintf("%s %s", radioIcon, option.Label)))
		
		// Add a badge if this is the default/recommended option
		if i == 0 && !option.Disabled {
			badge := BadgeStyle.Render("default")
			itemLine.WriteString(" ")
			itemLine.WriteString(badge)
		}
		
		s.WriteString(itemLine.String())
		s.WriteString("\n")
		
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
			
			s.WriteString(desc)
			s.WriteString("\n")
		}
	}
	
	// Show scroll indicators if needed
	if startIdx > 0 || startIdx+itemsToShow < len(m.options) {
		scrollInfo := MutedStyle.Render(fmt.Sprintf("\n[%d-%d of %d items]", 
			startIdx+1, 
			min(startIdx+itemsToShow, len(m.options)), 
			len(m.options)))
		s.WriteString(scrollInfo)
	}

	// Help text
	s.WriteString("\n")
	helpText := RenderHelp(
		"↑/↓: navigate",
		"enter: select",
		"q: cancel",
	)
	s.WriteString(helpText)

	return s.String()
}

// RadioSelection displays a radio button selection interface and returns the selected value
func RadioSelection(title string, options []RadioOption, defaultSelection int) (string, error) {
	// Ensure default selection is valid and not disabled
	if defaultSelection < 0 || defaultSelection >= len(options) {
		defaultSelection = 0
	}
	
	// Find first non-disabled option if default is disabled
	if options[defaultSelection].Disabled {
		for i := 0; i < len(options); i++ {
			if !options[i].Disabled {
				defaultSelection = i
				break
			}
		}
	}

	m := radioModel{
		options:  options,
		selected: defaultSelection,
		title:    title,
		width:    80, // Default width
		height:   24, // Default height
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	final := finalModel.(radioModel)
	if final.aborted {
		return "", fmt.Errorf("selection aborted")
	}

	return final.options[final.selected].Value, nil
}

