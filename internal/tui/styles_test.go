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

	"github.com/charmbracelet/lipgloss"
)

func TestRenderCheckbox(t *testing.T) {
	// Test checked checkbox
	result := RenderCheckbox(true, "Test Label")
	if !strings.Contains(result, "☑") {
		t.Error("Checked checkbox should contain checked symbol")
	}
	if !strings.Contains(result, "Test Label") {
		t.Error("Checkbox should contain the label")
	}

	// Test unchecked checkbox
	result = RenderCheckbox(false, "Test Label")
	if !strings.Contains(result, "☐") {
		t.Error("Unchecked checkbox should contain unchecked symbol")
	}
	if !strings.Contains(result, "Test Label") {
		t.Error("Checkbox should contain the label")
	}
}

func TestRenderProgress(t *testing.T) {
	// Test normal progress
	result := RenderProgress(3, 10, "items")
	if !strings.Contains(result, "[3/10]") {
		t.Error("Progress should show current/total")
	}
	if !strings.Contains(result, "30%") {
		t.Error("Progress should show percentage")
	}
	if !strings.Contains(result, "items") {
		t.Error("Progress should show label")
	}

	// Test zero total (edge case)
	result = RenderProgress(0, 0, "items")
	if result != "" {
		t.Error("Progress with zero total should return empty string")
	}

	// Test completed progress
	result = RenderProgress(10, 10, "complete")
	if !strings.Contains(result, "100%") {
		t.Error("Complete progress should show 100%")
	}
}

func TestColorConstants(t *testing.T) {
	// Test that color constants are defined
	colors := []string{
		ColorPrimary,
		ColorSecondary,
		ColorSuccess,
		ColorError,
		ColorInfo,
		ColorMuted,
		ColorText,
	}

	for _, color := range colors {
		if color == "" {
			t.Error("Color constant should not be empty")
		}
	}
}

func TestStylesInitialization(t *testing.T) {
	// Test that all styles are initialized and have proper colors
	styles := []struct {
		name  string
		style lipgloss.Style
	}{
		{"TitleStyle", TitleStyle},
		{"QuestionStyle", QuestionStyle},
		{"SuccessStyle", SuccessStyle},
		{"ErrorStyle", ErrorStyle},
		{"InfoStyle", InfoStyle},
		{"MutedStyle", MutedStyle},
		{"InputStyle", InputStyle},
		{"ListItemStyle", ListItemStyle},
		{"SelectedItemStyle", SelectedItemStyle},
		{"DescriptionStyle", DescriptionStyle},
		{"CheckboxCheckedStyle", CheckboxCheckedStyle},
		{"CheckboxUncheckedStyle", CheckboxUncheckedStyle},
		{"ProgressStyle", ProgressStyle},
		{"BorderStyle", BorderStyle},
	}

	for _, s := range styles {
		// Just test that the styles can render something without panicking
		rendered := s.style.Render("test")
		if rendered == "" {
			t.Errorf("Style %s should render non-empty output", s.name)
		}
	}
}
