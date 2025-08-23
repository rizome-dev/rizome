package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseRizomeContent(t *testing.T) {
	content := `# RIZOME.md

## Common Instructions

These are common instructions for all providers:
- Use Go best practices
- Follow the existing patterns

## Provider Overrides

### CLAUDE
Claude-specific instructions:
- Focus on clean architecture

### QWEN  
Qwen-specific instructions:
- Optimize for performance
`

	config, err := parseRizomeContent(content)
	if err != nil {
		t.Fatalf("Failed to parse content: %v", err)
	}

	if config.CommonInstructions == "" {
		t.Error("Common instructions should not be empty")
	}

	if !strings.Contains(config.CommonInstructions, "Use Go best practices") {
		t.Error("Common instructions should contain expected text")
	}

	claudeOverride, exists := config.ProviderOverrides["CLAUDE"]
	if !exists {
		t.Error("Claude override should exist")
	}

	if !strings.Contains(claudeOverride, "Focus on clean architecture") {
		t.Error("Claude override should contain expected text")
	}

	qwenOverride, exists := config.ProviderOverrides["QWEN"]
	if !exists {
		t.Error("Qwen override should exist")
	}

	if !strings.Contains(qwenOverride, "Optimize for performance") {
		t.Error("Qwen override should contain expected text")
	}
}

func TestGenerateProviderContent(t *testing.T) {
	config := &Config{
		CommonInstructions: "Common instructions here",
		ProviderOverrides: map[string]string{
			"CLAUDE": "Claude specific instructions",
		},
		Providers: standardProviders,
	}

	manager := &Manager{config: config}

	// Test Claude content generation
	content := manager.generateProviderContent("CLAUDE")

	if !strings.Contains(content, "CLAUDE.md") {
		t.Error("Content should contain provider name in header")
	}

	if !strings.Contains(content, "Common instructions here") {
		t.Error("Content should contain common instructions")
	}

	if !strings.Contains(content, "Claude specific instructions") {
		t.Error("Content should contain provider-specific instructions")
	}

	// Test provider without specific overrides
	content = manager.generateProviderContent("QWEN")

	if !strings.Contains(content, "QWEN.md") {
		t.Error("Content should contain provider name in header")
	}

	if !strings.Contains(content, "Common instructions here") {
		t.Error("Content should contain common instructions")
	}

	if strings.Contains(content, "QWEN-Specific Instructions") {
		t.Error("Content should not contain provider-specific section when no overrides exist")
	}
}

func TestSyncDryRun(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "rizome-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test RIZOME.md file
	rizomeContent := `# Test RIZOME.md

## Common Instructions
Test common instructions

## Provider Overrides
### CLAUDE
Test Claude instructions
`

	rizomePath := filepath.Join(tempDir, "RIZOME.md")
	err = os.WriteFile(rizomePath, []byte(rizomeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write RIZOME.md: %v", err)
	}

	// Create manager and test dry run
	manager, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	results, err := manager.Sync(true, false) // dry run
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Check that results indicate files would be created
	if len(results) != len(standardProviders) {
		t.Errorf("Expected %d results, got %d", len(standardProviders), len(results))
	}

	for _, result := range results {
		if result.Error != nil {
			t.Errorf("Unexpected error for %s: %v", result.Provider, result.Error)
		}
		if !result.Created {
			t.Errorf("Expected %s to be marked as created in dry run", result.Provider)
		}
	}

	// Verify no files were actually created
	for _, provider := range standardProviders {
		filename := filepath.Join(tempDir, provider+".md")
		if _, err := os.Stat(filename); err == nil {
			t.Errorf("File %s should not exist after dry run", filename)
		}
	}
}

func TestGetAvailableProviders(t *testing.T) {
	providers := GetAvailableProviders()
	
	if len(providers) == 0 {
		t.Error("Should return at least one provider")
	}
	
	// Should at least include standard providers as fallback
	expectedProviders := []string{"CLAUDE", "QWEN", "CURSOR", "GEMINI", "WINDSURF"}
	for _, expected := range expectedProviders {
		found := false
		for _, provider := range providers {
			if provider == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected provider %s not found in available providers", expected)
		}
	}
}

func TestGetEnabledProviders(t *testing.T) {
	providers := GetEnabledProviders()
	
	if len(providers) == 0 {
		t.Error("Should return at least one enabled provider")
	}
	
	// Should be a subset of available providers
	availableProviders := GetAvailableProviders()
	for _, enabled := range providers {
		found := false
		for _, available := range availableProviders {
			if enabled == available {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Enabled provider %s should be in available providers list", enabled)
		}
	}
}

func TestParseRizomeContentWithRegistryProviders(t *testing.T) {
	// Test that parsing uses provider registry
	content := `# RIZOME.md

## Common Instructions
Common instructions here

## Provider Overrides
### CLAUDE
Claude instructions
`

	config, err := parseRizomeContent(content)
	if err != nil {
		t.Fatalf("Failed to parse content: %v", err)
	}

	// Should use providers from registry, not just hardcoded ones
	if len(config.Providers) == 0 {
		t.Error("Config should have providers from registry")
	}

	// Should include available providers
	availableProviders := GetAvailableProviders()
	if len(config.Providers) != len(availableProviders) {
		t.Errorf("Expected %d providers from registry, got %d", len(availableProviders), len(config.Providers))
	}

	for _, expected := range availableProviders {
		found := false
		for _, provider := range config.Providers {
			if provider == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected provider %s from registry not found in config", expected)
		}
	}
}

func TestSyncWithRegistryFallback(t *testing.T) {
	// This test verifies that sync works even if registry is not available
	// and falls back to standard providers
	
	tempDir, err := os.MkdirTemp("", "rizome-sync-fallback-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test RIZOME.md file
	rizomeContent := `# Test RIZOME.md

## Common Instructions
Test common instructions

## Provider Overrides
### CLAUDE
Test Claude instructions
`

	rizomePath := filepath.Join(tempDir, "RIZOME.md")
	err = os.WriteFile(rizomePath, []byte(rizomeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write RIZOME.md: %v", err)
	}

	// Create manager
	manager, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test dry run to ensure it works with provider registry integration
	results, err := manager.Sync(true, false)
	if err != nil {
		t.Fatalf("Sync with registry integration failed: %v", err)
	}

	// Should have results for available providers
	availableProviders := GetAvailableProviders()
	if len(results) != len(availableProviders) {
		t.Errorf("Expected %d results for available providers, got %d", len(availableProviders), len(results))
	}

	// All results should be successful
	for _, result := range results {
		if result.Error != nil {
			t.Errorf("Unexpected error for provider %s: %v", result.Provider, result.Error)
		}
	}
}

func TestSyncProviders(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "rizome-sync-providers-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test RIZOME.md file
	rizomeContent := `# Test RIZOME.md

## Common Instructions
Test common instructions
`

	rizomePath := filepath.Join(tempDir, "RIZOME.md")
	err = os.WriteFile(rizomePath, []byte(rizomeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write RIZOME.md: %v", err)
	}

	// Create manager
	manager, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test syncing specific providers
	selectedProviders := []string{"CLAUDE", "QWEN"}
	results, err := manager.SyncProviders(selectedProviders, true, false) // dry run
	if err != nil {
		t.Fatalf("SyncProviders failed: %v", err)
	}

	// Should have results only for selected providers
	if len(results) != len(selectedProviders) {
		t.Errorf("Expected %d results for selected providers, got %d", len(selectedProviders), len(results))
	}

	// Verify results are for correct providers
	for _, result := range results {
		found := false
		for _, selected := range selectedProviders {
			if result.Provider == selected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Unexpected result for provider %s", result.Provider)
		}
		if result.Error != nil {
			t.Errorf("Unexpected error for provider %s: %v", result.Provider, result.Error)
		}
	}
}
