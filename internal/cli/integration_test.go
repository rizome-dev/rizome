package cli

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

	"github.com/rizome-dev/rizome/internal/config"
	"github.com/rizome-dev/rizome/internal/sync"
)

func TestInitWithTemplate(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "rizome-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Test init with default template (non-interactive mode)
	err = runInitInteractive(false, "default")
	if err != nil {
		t.Fatalf("Failed to run init with template: %v", err)
	}

	// Verify RIZOME.md was created
	rizomePath := filepath.Join(tempDir, "RIZOME.md")
	if _, err := os.Stat(rizomePath); os.IsNotExist(err) {
		t.Error("RIZOME.md should be created")
	}

	// Read the content and verify it's from the default template
	content, err := os.ReadFile(rizomePath)
	if err != nil {
		t.Fatalf("Failed to read RIZOME.md: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "## Common Instructions") {
		t.Error("RIZOME.md should contain Common Instructions section")
	}

	if !strings.Contains(contentStr, "## Provider Overrides") {
		t.Error("RIZOME.md should contain Provider Overrides section")
	}

	// Verify all standard providers are included
	expectedProviders := []string{"CLAUDE", "QWEN", "CURSOR", "GEMINI", "WINDSURF"}
	for _, provider := range expectedProviders {
		if !strings.Contains(contentStr, "### "+provider) {
			t.Errorf("RIZOME.md should contain %s provider section", provider)
		}
	}
}

func TestSyncNonInteractive(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "rizome-sync-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test RIZOME.md file
	rizomeContent := `# Test RIZOME.md

## Common Instructions

These are test common instructions:
- Use best practices
- Follow conventions

## Provider Overrides

### CLAUDE
Claude-specific test instructions

### QWEN
Qwen-specific test instructions
`

	rizomePath := filepath.Join(tempDir, "RIZOME.md")
	err = os.WriteFile(rizomePath, []byte(rizomeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write RIZOME.md: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Test sync in non-interactive mode with specific providers
	err = runSyncInteractive(false, false, true, "CLAUDE,QWEN")
	if err != nil {
		t.Fatalf("Failed to run sync: %v", err)
	}

	// Verify that CLAUDE.md and QWEN.md were created
	claudePath := filepath.Join(tempDir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		t.Error("CLAUDE.md should be created")
	}

	qwenPath := filepath.Join(tempDir, "QWEN.md")
	if _, err := os.Stat(qwenPath); os.IsNotExist(err) {
		t.Error("QWEN.md should be created")
	}

	// Verify CURSOR.md was NOT created (not in provider list)
	cursorPath := filepath.Join(tempDir, "CURSOR.md")
	if _, err := os.Stat(cursorPath); err == nil {
		t.Error("CURSOR.md should not be created when not in provider list")
	}

	// Verify content of created files
	claudeContent, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("Failed to read CLAUDE.md: %v", err)
	}

	claudeStr := string(claudeContent)
	if !strings.Contains(claudeStr, "Common Instructions") {
		t.Error("CLAUDE.md should contain common instructions")
	}

	if !strings.Contains(claudeStr, "Claude-specific test instructions") {
		t.Error("CLAUDE.md should contain Claude-specific instructions")
	}

	if !strings.Contains(claudeStr, "Use best practices") {
		t.Error("CLAUDE.md should contain common instruction content")
	}
}

func TestTemplateSystemIntegration(t *testing.T) {
	// Create temporary directory for template manager
	tempDir, err := os.MkdirTemp("", "rizome-template-integration")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create template manager with custom config directory
	tm := &config.TemplateManager{}
	tm = &config.TemplateManager{} // Reset to use custom dir later

	// Set custom config directory
	configDir := filepath.Join(tempDir, ".rizome")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Test creating a custom template
	customTemplate := config.Template{
		Name:        "Integration Test Template",
		Description: "Template for integration testing",
		Content: `# Integration Test RIZOME.md

## Common Instructions
Integration test instructions

## Provider Overrides
### CLAUDE
Integration test Claude instructions`,
	}

	// Test the template loading and saving
	tm, err = config.NewTemplateManager()
	if err != nil {
		t.Fatalf("Failed to create template manager: %v", err)
	}

	err = tm.SaveTemplate("integration-test", customTemplate)
	if err != nil {
		t.Fatalf("Failed to save custom template: %v", err)
	}

	// Verify the template can be retrieved
	retrievedTemplate, err := tm.GetTemplate("integration-test")
	if err != nil {
		t.Fatalf("Failed to get custom template: %v", err)
	}

	if retrievedTemplate.Name != customTemplate.Name {
		t.Errorf("Expected template name '%s', got '%s'", customTemplate.Name, retrievedTemplate.Name)
	}

	// Test that we can list templates and our custom one is included
	templates, err := tm.ListTemplates()
	if err != nil {
		t.Fatalf("Failed to list templates: %v", err)
	}

	if _, exists := templates["integration-test"]; !exists {
		t.Error("Custom template should exist in template list")
	}

	// Should have at least default templates + custom template
	if len(templates) < 3 {
		t.Errorf("Expected at least 3 templates (including custom), got %d", len(templates))
	}
}

func TestSyncWithCustomProviders(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "rizome-custom-providers-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a RIZOME.md with custom provider overrides
	rizomeContent := `# Custom Providers Test

## Common Instructions
Common instructions for all providers

## Provider Overrides
### CLAUDE
Claude-specific instructions

### CUSTOM_PROVIDER
Custom provider instructions that shouldn't be synced by default
`

	rizomePath := filepath.Join(tempDir, "RIZOME.md")
	err = os.WriteFile(rizomePath, []byte(rizomeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write RIZOME.md: %v", err)
	}

	// Create sync manager to test the parsing
	syncManager, err := sync.New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create sync manager: %v", err)
	}

	config := syncManager.GetConfig()

	// Verify that the custom provider override was parsed
	if _, exists := config.ProviderOverrides["CUSTOM_PROVIDER"]; !exists {
		t.Error("Custom provider override should be parsed")
	}

	// Verify common instructions were parsed
	if config.CommonInstructions == "" {
		t.Error("Common instructions should be parsed")
	}

	if !strings.Contains(config.CommonInstructions, "Common instructions for all providers") {
		t.Error("Common instructions should contain expected content")
	}

	// Test syncing only standard providers (CUSTOM_PROVIDER should be ignored)
	results, err := syncManager.SyncProviders([]string{"CLAUDE"}, false, false)
	if err != nil {
		t.Fatalf("Failed to sync specific providers: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 sync result, got %d", len(results))
	}

	if results[0].Provider != "CLAUDE" {
		t.Errorf("Expected CLAUDE result, got %s", results[0].Provider)
	}

	// Verify CLAUDE.md was created with both common and specific content
	claudePath := filepath.Join(tempDir, "CLAUDE.md")
	claudeContent, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("Failed to read CLAUDE.md: %v", err)
	}

	claudeStr := string(claudeContent)
	if !strings.Contains(claudeStr, "Common instructions for all providers") {
		t.Error("CLAUDE.md should contain common instructions")
	}

	if !strings.Contains(claudeStr, "Claude-specific instructions") {
		t.Error("CLAUDE.md should contain provider-specific instructions")
	}

	// Verify that custom provider file was NOT created
	customPath := filepath.Join(tempDir, "CUSTOM_PROVIDER.md")
	if _, err := os.Stat(customPath); err == nil {
		t.Error("CUSTOM_PROVIDER.md should not be created automatically")
	}
}
