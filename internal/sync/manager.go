package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/rizome-dev/rizome/internal/config"
)

// Manager handles synchronization of RIZOME.md with provider files
type Manager struct {
	baseDir string
	config  *Config
}

// Config represents parsed RIZOME.md configuration
type Config struct {
	CommonInstructions string
	ProviderOverrides  map[string]string
	Providers          []string
}

// SyncResult represents the result of syncing a provider file
type SyncResult struct {
	Provider string
	Created  bool
	Updated  bool
	Error    error
}

// Standard providers to sync
var standardProviders = []string{"CLAUDE", "QWEN", "CURSOR", "GEMINI", "WINDSURF"}

// New creates a new sync manager
func New(baseDir string) (*Manager, error) {
	m := &Manager{
		baseDir: baseDir,
	}

	config, err := m.parseRizomeFile()
	if err != nil {
		return nil, err
	}

	m.config = config
	return m, nil
}

// parseRizomeFile parses RIZOME.md and extracts configuration
func (m *Manager) parseRizomeFile() (*Config, error) {
	rizomePath := filepath.Join(m.baseDir, "RIZOME.md")

	content, err := os.ReadFile(rizomePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("RIZOME.md not found in current directory")
		}
		return nil, fmt.Errorf("failed to read RIZOME.md: %w", err)
	}

	return parseRizomeContent(string(content))
}

// parseRizomeContent parses the content of RIZOME.md
func parseRizomeContent(content string) (*Config, error) {
	config := &Config{
		ProviderOverrides: make(map[string]string),
		Providers:         standardProviders,
	}

	lines := strings.Split(content, "\n")
	var currentSection string
	var sectionContent strings.Builder
	var currentProvider string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for main sections
		if strings.HasPrefix(trimmed, "## ") {
			// Save previous section
			if currentSection != "" {
				content := strings.TrimSpace(sectionContent.String())
				if currentSection == "common" {
					config.CommonInstructions = content
				} else if currentProvider != "" {
					config.ProviderOverrides[currentProvider] = content
				}
			}

			// Start new section
			sectionContent.Reset()
			currentProvider = ""

			sectionTitle := strings.ToLower(strings.TrimSpace(trimmed[3:]))
			if strings.Contains(sectionTitle, "common") {
				currentSection = "common"
			} else if strings.Contains(sectionTitle, "provider") || strings.Contains(sectionTitle, "override") {
				currentSection = "providers"
			} else {
				currentSection = ""
			}
			continue
		}

		// Check for provider subsections
		if strings.HasPrefix(trimmed, "### ") && currentSection == "providers" {
			// Save previous provider
			if currentProvider != "" {
				content := strings.TrimSpace(sectionContent.String())
				config.ProviderOverrides[currentProvider] = content
			}

			// Start new provider
			sectionContent.Reset()
			currentProvider = strings.ToUpper(strings.TrimSpace(trimmed[4:]))
			continue
		}

		// Add line to current section
		if currentSection != "" {
			sectionContent.WriteString(line + "\n")
		}
	}

	// Save final section
	if currentSection != "" {
		content := strings.TrimSpace(sectionContent.String())
		if currentSection == "common" {
			config.CommonInstructions = content
		} else if currentProvider != "" {
			config.ProviderOverrides[currentProvider] = content
		}
	}

	return config, nil
}

// GetConfig returns the parsed configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// Sync performs the synchronization operation for all providers
func (m *Manager) Sync(dryRun, force bool) ([]SyncResult, error) {
	var results []SyncResult

	// Update RIZOME.md timestamp first (unless dry run)
	if !dryRun {
		if err := m.updateRizomeTimestamp(); err != nil {
			// Don't fail the sync if timestamp update fails, just continue
			// This ensures sync works even if RIZOME.md is read-only or has issues
		}
	}

	for _, provider := range m.config.Providers {
		result := m.syncProvider(provider, dryRun, force)
		results = append(results, result)
	}

	return results, nil
}

// SyncProviders performs the synchronization operation for specific providers
func (m *Manager) SyncProviders(providers []string, dryRun, force bool) ([]SyncResult, error) {
	var results []SyncResult

	// Update RIZOME.md timestamp first (unless dry run)
	if !dryRun {
		if err := m.updateRizomeTimestamp(); err != nil {
			// Don't fail the sync if timestamp update fails, just continue
			// This ensures sync works even if RIZOME.md is read-only or has issues
		}
	}

	for _, provider := range providers {
		result := m.syncProvider(provider, dryRun, force)
		results = append(results, result)
	}

	return results, nil
}

// syncProvider syncs a single provider file
func (m *Manager) syncProvider(provider string, dryRun, force bool) SyncResult {
	filename := fmt.Sprintf("%s.md", provider)
	filepath := filepath.Join(m.baseDir, filename)

	result := SyncResult{Provider: provider}

	// Check if file exists
	_, err := os.Stat(filepath)
	fileExists := !os.IsNotExist(err)

	if fileExists && !force && !dryRun {
		// TODO: Add interactive prompt for overwrite
		// For now, we'll update if it exists
	}

	// Generate content
	content := m.generateProviderContent(provider)

	if dryRun {
		result.Created = !fileExists
		result.Updated = fileExists
		return result
	}

	// Write file
	err = os.WriteFile(filepath, []byte(content), 0644)
	if err != nil {
		result.Error = fmt.Errorf("failed to write %s: %w", filename, err)
		return result
	}

	result.Created = !fileExists
	result.Updated = fileExists

	return result
}

// generateProviderContent generates the content for a provider file
func (m *Manager) generateProviderContent(provider string) string {
	var content strings.Builder

	// Add header
	content.WriteString(fmt.Sprintf("# %s.md\n\n", provider))
	content.WriteString("This file is managed by Rizome CLI. Do not edit directly.\n")
	content.WriteString("Update RIZOME.md and run 'rizome sync' instead.\n\n")

	// Add common instructions if available
	if m.config.CommonInstructions != "" {
		content.WriteString("## Common Instructions\n\n")
		content.WriteString(m.config.CommonInstructions)
		content.WriteString("\n\n")
	}

	// Add provider-specific overrides if available
	if override, exists := m.config.ProviderOverrides[provider]; exists && override != "" {
		content.WriteString(fmt.Sprintf("## %s-Specific Instructions\n\n", provider))
		content.WriteString(override)
		content.WriteString("\n")
	}

	// Inject timestamp for model grounding
	return config.InjectTimestamp(content.String())
}

// updateRizomeTimestamp updates the Last Updated timestamp in RIZOME.md
func (m *Manager) updateRizomeTimestamp() error {
	rizomePath := filepath.Join(m.baseDir, "RIZOME.md")
	
	// Read current content
	content, err := os.ReadFile(rizomePath)
	if err != nil {
		return fmt.Errorf("failed to read RIZOME.md: %w", err)
	}
	
	// Inject/update timestamp
	updatedContent := config.InjectTimestamp(string(content))
	
	// Write back to file
	err = os.WriteFile(rizomePath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to update RIZOME.md timestamp: %w", err)
	}
	
	return nil
}
