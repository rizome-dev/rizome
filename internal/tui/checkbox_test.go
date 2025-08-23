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
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCheckboxOption(t *testing.T) {
	option := CheckboxOption{
		Label:       "Test Option",
		Description: "Test Description",
		Value:       "test-value",
		Checked:     true,
	}

	if option.Label != "Test Option" {
		t.Errorf("Expected label 'Test Option', got '%s'", option.Label)
	}

	if option.Description != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", option.Description)
	}

	if option.Value != "test-value" {
		t.Errorf("Expected value 'test-value', got '%s'", option.Value)
	}

	if !option.Checked {
		t.Error("Expected option to be checked")
	}
}

func TestNewCheckboxModel(t *testing.T) {
	options := []CheckboxOption{
		{Label: "Option 1", Value: "opt1", Checked: true},
		{Label: "Option 2", Value: "opt2", Checked: false},
		{Label: "Option 3", Value: "opt3", Checked: true},
	}

	model := NewCheckboxModel("Test Title", options)

	if model.title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", model.title)
	}

	if len(model.options) != 3 {
		t.Errorf("Expected 3 options, got %d", len(model.options))
	}

	if model.cursor != 0 {
		t.Errorf("Expected cursor to start at 0, got %d", model.cursor)
	}

	if model.confirmed {
		t.Error("Model should not be confirmed initially")
	}

	if model.cancelled {
		t.Error("Model should not be cancelled initially")
	}
}

func TestCheckboxModelNavigation(t *testing.T) {
	options := []CheckboxOption{
		{Label: "Option 1", Value: "opt1", Checked: false},
		{Label: "Option 2", Value: "opt2", Checked: false},
		{Label: "Option 3", Value: "opt3", Checked: false},
	}

	model := NewCheckboxModel("Test", options)

	// Test moving down
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(checkboxModel)
	if model.cursor != 1 {
		t.Errorf("Expected cursor at 1 after down, got %d", model.cursor)
	}

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(checkboxModel)
	if model.cursor != 2 {
		t.Errorf("Expected cursor at 2 after down, got %d", model.cursor)
	}

	// Test not going past end
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(checkboxModel)
	if model.cursor != 2 {
		t.Errorf("Expected cursor to stay at 2 at end, got %d", model.cursor)
	}

	// Test moving up
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updatedModel.(checkboxModel)
	if model.cursor != 1 {
		t.Errorf("Expected cursor at 1 after up, got %d", model.cursor)
	}

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updatedModel.(checkboxModel)
	if model.cursor != 0 {
		t.Errorf("Expected cursor at 0 after up, got %d", model.cursor)
	}

	// Test not going past beginning
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updatedModel.(checkboxModel)
	if model.cursor != 0 {
		t.Errorf("Expected cursor to stay at 0 at beginning, got %d", model.cursor)
	}
}

func TestCheckboxModelToggling(t *testing.T) {
	options := []CheckboxOption{
		{Label: "Option 1", Value: "opt1", Checked: false},
		{Label: "Option 2", Value: "opt2", Checked: true},
	}

	model := NewCheckboxModel("Test", options)

	// Test toggling first option (currently unchecked)
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updatedModel.(checkboxModel)
	if !model.options[0].Checked {
		t.Error("First option should be checked after toggle")
	}

	// Move to second option and toggle it (currently checked)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(checkboxModel)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updatedModel.(checkboxModel)
	if model.options[1].Checked {
		t.Error("Second option should be unchecked after toggle")
	}
}

func TestCheckboxModelSelectAll(t *testing.T) {
	options := []CheckboxOption{
		{Label: "Option 1", Value: "opt1", Checked: false},
		{Label: "Option 2", Value: "opt2", Checked: false},
		{Label: "Option 3", Value: "opt3", Checked: true},
	}

	model := NewCheckboxModel("Test", options)

	// Test select all
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updatedModel, _ := model.Update(keyMsg)
	model = updatedModel.(checkboxModel)

	for i, option := range model.options {
		if !option.Checked {
			t.Errorf("Option %d should be checked after select all", i)
		}
	}
}

func TestCheckboxModelDeselectAll(t *testing.T) {
	options := []CheckboxOption{
		{Label: "Option 1", Value: "opt1", Checked: true},
		{Label: "Option 2", Value: "opt2", Checked: true},
		{Label: "Option 3", Value: "opt3", Checked: false},
	}

	model := NewCheckboxModel("Test", options)

	// Test deselect all (capital A)
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}}
	updatedModel, _ := model.Update(keyMsg)
	model = updatedModel.(checkboxModel)

	for i, option := range model.options {
		if option.Checked {
			t.Errorf("Option %d should be unchecked after deselect all", i)
		}
	}
}

func TestCheckboxModelConfirmAndCancel(t *testing.T) {
	options := []CheckboxOption{
		{Label: "Option 1", Value: "opt1", Checked: true},
	}

	// Test confirmation
	model := NewCheckboxModel("Test", options)
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(checkboxModel)

	if !model.IsConfirmed() {
		t.Error("Model should be confirmed after Enter")
	}

	if cmd == nil {
		t.Error("Should return a command after confirmation")
	}

	// Test cancellation
	model = NewCheckboxModel("Test", options)
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	updatedModel, cmd = model.Update(keyMsg)
	model = updatedModel.(checkboxModel)

	if !model.IsCancelled() {
		t.Error("Model should be cancelled after 'q'")
	}

	if cmd == nil {
		t.Error("Should return a command after cancellation")
	}
}

func TestCheckboxModelGetSelectedValues(t *testing.T) {
	options := []CheckboxOption{
		{Label: "Option 1", Value: "opt1", Checked: true},
		{Label: "Option 2", Value: "opt2", Checked: false},
		{Label: "Option 3", Value: "opt3", Checked: true},
	}

	model := NewCheckboxModel("Test", options)
	selectedValues := model.GetSelectedValues()

	expectedValues := []string{"opt1", "opt3"}
	if len(selectedValues) != len(expectedValues) {
		t.Errorf("Expected %d selected values, got %d", len(expectedValues), len(selectedValues))
	}

	for i, expected := range expectedValues {
		if i >= len(selectedValues) || selectedValues[i] != expected {
			t.Errorf("Expected selected value %d to be '%s', got '%s'", i, expected, selectedValues[i])
		}
	}
}

func TestCheckboxModelView(t *testing.T) {
	options := []CheckboxOption{
		{Label: "Option 1", Value: "opt1", Checked: true, Description: "First option"},
		{Label: "Option 2", Value: "opt2", Checked: false, Description: "Second option"},
	}

	model := NewCheckboxModel("Test Title", options)
	view := model.View()

	// Check that the view contains the title
	if !strings.Contains(view, "Test Title") {
		t.Error("View should contain the title")
	}

	// Check that the view contains the options
	if !strings.Contains(view, "Option 1") {
		t.Error("View should contain Option 1")
	}

	if !strings.Contains(view, "Option 2") {
		t.Error("View should contain Option 2")
	}

	// Check that the view contains descriptions
	if !strings.Contains(view, "First option") {
		t.Error("View should contain first option description")
	}

	if !strings.Contains(view, "Second option") {
		t.Error("View should contain second option description")
	}

	// Check that the view contains checkboxes
	if !strings.Contains(view, "☑") {
		t.Error("View should contain checked checkbox")
	}

	if !strings.Contains(view, "☐") {
		t.Error("View should contain unchecked checkbox")
	}

	// Check that the view contains instructions
	if !strings.Contains(view, "Space: toggle") {
		t.Error("View should contain toggle instructions")
	}
}
