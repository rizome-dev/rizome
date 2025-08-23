package config

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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultTemplate(t *testing.T) {
	template := DefaultTemplate()

	if template.Name != "Default Template" {
		t.Errorf("Expected name 'Default Template', got '%s'", template.Name)
	}

	if template.Description == "" {
		t.Error("Default template should have a description")
	}

	if template.Content == "" {
		t.Error("Default template should have content")
	}

	// Check that content contains expected sections
	if !strings.Contains(template.Content, "## Common Instructions") {
		t.Error("Default template should contain Common Instructions section")
	}

	if !strings.Contains(template.Content, "## Provider Overrides") {
		t.Error("Default template should contain Provider Overrides section")
	}

	// Check that all standard providers are included
	expectedProviders := []string{"CLAUDE", "QWEN", "CURSOR", "GEMINI", "WINDSURF"}
	for _, provider := range expectedProviders {
		if !strings.Contains(template.Content, "### "+provider) {
			t.Errorf("Default template should contain %s provider section", provider)
		}
	}
}

func TestGetDefaultTemplates(t *testing.T) {
	templates := GetDefaultTemplates()

	if len(templates) != 1 {
		t.Errorf("Expected exactly 1 default template, got %d", len(templates))
	}

	// Check default template exists
	defaultTemplate, exists := templates["default"]
	if !exists {
		t.Error("Default templates should include 'default' template")
	}
	if defaultTemplate.Name != "Default Template" {
		t.Error("Default template should have correct name")
	}
}

func TestTemplateManager(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "rizome-template-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create template manager with custom config directory
	tm := &TemplateManager{
		configDir:  tempDir,
		configFile: filepath.Join(tempDir, "config.yaml"),
	}

	// Test loading config (should create default config)
	config, err := tm.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.Templates == nil {
		t.Error("Config should have templates map")
	}

	// Should have default templates
	if len(config.Templates) < 1 {
		t.Errorf("Expected at least 1 template in default config, got %d", len(config.Templates))
	}

	// Check that default template exists
	_, exists := config.Templates["default"]
	if !exists {
		t.Error("Default config should contain 'default' template")
	}
}

func TestTemplateManagerCRUD(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "rizome-template-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create template manager with custom config directory
	tm := &TemplateManager{
		configDir:  tempDir,
		configFile: filepath.Join(tempDir, "config.yaml"),
	}

	// Test creating a new template
	testTemplate := Template{
		Name:        "Test Template",
		Description: "A test template",
		Content:     "# Test RIZOME.md\n\nTest content",
	}

	err = tm.SaveTemplate("test", testTemplate)
	if err != nil {
		t.Fatalf("Failed to save template: %v", err)
	}

	// Test retrieving the template
	retrievedTemplate, err := tm.GetTemplate("test")
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}

	if retrievedTemplate.Name != testTemplate.Name {
		t.Errorf("Expected name '%s', got '%s'", testTemplate.Name, retrievedTemplate.Name)
	}

	if retrievedTemplate.Description != testTemplate.Description {
		t.Errorf("Expected description '%s', got '%s'", testTemplate.Description, retrievedTemplate.Description)
	}

	if retrievedTemplate.Content != testTemplate.Content {
		t.Errorf("Expected content '%s', got '%s'", testTemplate.Content, retrievedTemplate.Content)
	}

	// Test listing templates
	templates, err := tm.ListTemplates()
	if err != nil {
		t.Fatalf("Failed to list templates: %v", err)
	}

	// Should have at least the test template plus default templates
	if len(templates) < 2 {
		t.Errorf("Expected at least 2 templates, got %d", len(templates))
	}

	// Check that our test template is in the list
	listedTemplate, exists := templates["test"]
	if !exists {
		t.Error("Test template should exist in template list")
	}

	if listedTemplate.Name != testTemplate.Name {
		t.Error("Listed template should match saved template")
	}

	// Test template existence check
	exists, err = tm.TemplateExists("test")
	if err != nil {
		t.Fatalf("Failed to check template existence: %v", err)
	}
	if !exists {
		t.Error("Test template should exist")
	}

	exists, err = tm.TemplateExists("nonexistent")
	if err != nil {
		t.Fatalf("Failed to check template existence: %v", err)
	}
	if exists {
		t.Error("Nonexistent template should not exist")
	}

	// Test deleting the template
	err = tm.DeleteTemplate("test")
	if err != nil {
		t.Fatalf("Failed to delete template: %v", err)
	}

	// Verify template is deleted
	_, err = tm.GetTemplate("test")
	if err == nil {
		t.Error("Getting deleted template should return an error")
	}

	exists, err = tm.TemplateExists("test")
	if err != nil {
		t.Fatalf("Failed to check template existence after deletion: %v", err)
	}
	if exists {
		t.Error("Deleted template should not exist")
	}
}

func TestTemplateManagerUpdate(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "rizome-template-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create template manager with custom config directory
	tm := &TemplateManager{
		configDir:  tempDir,
		configFile: filepath.Join(tempDir, "config.yaml"),
	}

	// Create initial template
	originalTemplate := Template{
		Name:        "Original Template",
		Description: "Original description",
		Content:     "Original content",
	}

	err = tm.SaveTemplate("update-test", originalTemplate)
	if err != nil {
		t.Fatalf("Failed to save original template: %v", err)
	}

	// Update the template
	updatedTemplate := Template{
		Name:        "Updated Template",
		Description: "Updated description",
		Content:     "Updated content",
	}

	err = tm.SaveTemplate("update-test", updatedTemplate)
	if err != nil {
		t.Fatalf("Failed to update template: %v", err)
	}

	// Verify the template was updated
	retrievedTemplate, err := tm.GetTemplate("update-test")
	if err != nil {
		t.Fatalf("Failed to get updated template: %v", err)
	}

	if retrievedTemplate.Name != updatedTemplate.Name {
		t.Errorf("Expected updated name '%s', got '%s'", updatedTemplate.Name, retrievedTemplate.Name)
	}

	if retrievedTemplate.Description != updatedTemplate.Description {
		t.Errorf("Expected updated description '%s', got '%s'", updatedTemplate.Description, retrievedTemplate.Description)
	}

	if retrievedTemplate.Content != updatedTemplate.Content {
		t.Errorf("Expected updated content '%s', got '%s'", updatedTemplate.Content, retrievedTemplate.Content)
	}
}

func TestTemplateManagerInvalidOperations(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "rizome-template-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create template manager with custom config directory
	tm := &TemplateManager{
		configDir:  tempDir,
		configFile: filepath.Join(tempDir, "config.yaml"),
	}

	// Test getting non-existent template
	_, err = tm.GetTemplate("nonexistent")
	if err == nil {
		t.Error("Getting non-existent template should return an error")
	}

	// Test deleting non-existent template
	err = tm.DeleteTemplate("nonexistent")
	if err == nil {
		t.Error("Deleting non-existent template should return an error")
	}
}
