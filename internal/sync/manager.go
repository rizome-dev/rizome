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
	ProjectOverview    string
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

// Standard providers to sync (fallback when registry is not available)
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

// GetAvailableProviders returns the list of all available providers from registry
func GetAvailableProviders() []string {
	tm, err := config.NewTemplateManager()
	if err != nil {
		// Fallback to standard providers if registry unavailable
		return standardProviders
	}
	
	allProviders, err := tm.GetAllProviders()
	if err != nil || len(allProviders) == 0 {
		// Fallback to standard providers if registry empty or error
		return standardProviders
	}
	
	return allProviders
}

// GetEnabledProviders returns the list of enabled providers from registry
func GetEnabledProviders() []string {
	tm, err := config.NewTemplateManager()
	if err != nil {
		// Fallback to standard providers if registry unavailable
		return standardProviders
	}
	
	enabledProviders, err := tm.GetEnabledProviders()
	if err != nil || len(enabledProviders) == 0 {
		// Fallback to standard providers if registry empty or error
		return standardProviders
	}
	
	return enabledProviders
}

// parseRizomeContent parses the content of RIZOME.md
func parseRizomeContent(content string) (*Config, error) {
	config := &Config{
		ProviderOverrides: make(map[string]string),
		Providers:         GetAvailableProviders(),
	}

	lines := strings.Split(content, "\n")
	var currentSection string
	var sectionContent strings.Builder
	var currentProvider string
	var overviewContent strings.Builder
	var inHeader bool = false
	var headerFound bool = false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for main header
		if strings.HasPrefix(trimmed, "# ") && !headerFound {
			headerFound = true
			inHeader = true
			continue // Skip the header line itself
		}

		// Check for main sections
		if strings.HasPrefix(trimmed, "## ") {
			// Save overview content if we were capturing it
			if inHeader {
				config.ProjectOverview = strings.TrimSpace(overviewContent.String())
				inHeader = false
			}

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

		// Capture overview content (between header and first ## section)
		if inHeader {
			overviewContent.WriteString(line + "\n")
		} else if currentSection != "" {
			// Add line to current section
			sectionContent.WriteString(line + "\n")
		}
	}

	// Save overview if we never hit a ## section
	if inHeader {
		config.ProjectOverview = strings.TrimSpace(overviewContent.String())
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

	// Add project overview if available
	if m.config.ProjectOverview != "" {
		content.WriteString(m.config.ProjectOverview)
		content.WriteString("\n\n")
	}

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

// SyncFromProvider performs sync from a provider file to other files
func (m *Manager) SyncFromProvider(sourceProvider string, destinations []string, dryRun, force bool) ([]SyncResult, error) {
	var results []SyncResult
	
	// Parse the source provider file
	sourceConfig, err := m.parseProviderFile(sourceProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source file %s.md: %w", sourceProvider, err)
	}
	
	// Update timestamp in source file (unless dry run)
	if !dryRun {
		if err := m.updateProviderTimestamp(sourceProvider); err != nil {
			// Don't fail the sync if timestamp update fails
		}
	}
	
	// Sync to each destination
	for _, dest := range destinations {
		if dest == "RIZOME" {
			// Syncing to RIZOME.md
			result := m.syncToRizome(sourceConfig, sourceProvider, dryRun, force)
			results = append(results, result)
		} else {
			// Syncing to another provider file
			result := m.syncProviderToProvider(sourceConfig, sourceProvider, dest, dryRun, force)
			results = append(results, result)
		}
	}
	
	return results, nil
}

// parseProviderFile parses a provider file and extracts configuration
func (m *Manager) parseProviderFile(provider string) (*Config, error) {
	filepath := filepath.Join(m.baseDir, fmt.Sprintf("%s.md", provider))
	
	content, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s.md not found in current directory", provider)
		}
		return nil, fmt.Errorf("failed to read %s.md: %w", provider, err)
	}
	
	return parseProviderContent(string(content), provider)
}

// parseProviderContent parses the content of a provider file
func parseProviderContent(content string, provider string) (*Config, error) {
	config := &Config{
		ProviderOverrides: make(map[string]string),
		Providers:         GetAvailableProviders(),
	}
	
	lines := strings.Split(content, "\n")
	var currentSection string
	var sectionContent strings.Builder
	var overviewContent strings.Builder
	var inHeader bool = false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Skip comment lines and metadata
		if strings.HasPrefix(trimmed, "<!--") {
			continue
		}
		
		// Skip header line
		if strings.HasPrefix(trimmed, "# ") {
			inHeader = true
			continue
		}
		
		// Skip "managed by Rizome" notice
		if strings.Contains(trimmed, "managed by Rizome") || 
		   strings.Contains(trimmed, "Do not edit directly") ||
		   strings.Contains(trimmed, "Update RIZOME.md") {
			continue
		}
		
		// Check for sections
		if strings.HasPrefix(trimmed, "## ") {
			// Save overview if we were capturing it
			if inHeader {
				config.ProjectOverview = strings.TrimSpace(overviewContent.String())
				inHeader = false
			}
			
			// Save previous section
			if currentSection != "" {
				content := strings.TrimSpace(sectionContent.String())
				if currentSection == "common" {
					config.CommonInstructions = content
				} else if currentSection == "provider" {
					config.ProviderOverrides[provider] = content
				}
			}
			
			// Start new section
			sectionContent.Reset()
			
			sectionTitle := strings.ToLower(strings.TrimSpace(trimmed[3:]))
			if strings.Contains(sectionTitle, "common") {
				currentSection = "common"
			} else if strings.Contains(sectionTitle, provider) || 
			          strings.Contains(sectionTitle, "specific") {
				currentSection = "provider"
			} else {
				currentSection = ""
			}
			continue
		}
		
		// Capture content
		if inHeader {
			overviewContent.WriteString(line + "\n")
		} else if currentSection != "" {
			sectionContent.WriteString(line + "\n")
		}
	}
	
	// Save final section
	if inHeader {
		config.ProjectOverview = strings.TrimSpace(overviewContent.String())
	}
	if currentSection != "" {
		content := strings.TrimSpace(sectionContent.String())
		if currentSection == "common" {
			config.CommonInstructions = content
		} else if currentSection == "provider" {
			config.ProviderOverrides[provider] = content
		}
	}
	
	return config, nil
}

// syncToRizome syncs provider config to RIZOME.md
func (m *Manager) syncToRizome(sourceConfig *Config, sourceProvider string, dryRun, force bool) SyncResult {
	rizomePath := filepath.Join(m.baseDir, "RIZOME.md")
	result := SyncResult{Provider: "RIZOME"}
	
	// Check if RIZOME.md exists
	_, err := os.Stat(rizomePath)
	fileExists := !os.IsNotExist(err)
	
	if fileExists && !force && !dryRun {
		// TODO: Add interactive prompt for overwrite
	}
	
	// Generate RIZOME.md content from source config
	content := m.generateRizomeContent(sourceConfig, sourceProvider)
	
	if dryRun {
		result.Created = !fileExists
		result.Updated = fileExists
		return result
	}
	
	// Write file
	err = os.WriteFile(rizomePath, []byte(content), 0644)
	if err != nil {
		result.Error = fmt.Errorf("failed to write RIZOME.md: %w", err)
		return result
	}
	
	result.Created = !fileExists
	result.Updated = fileExists
	
	return result
}

// generateRizomeContent generates RIZOME.md content from a config
func (m *Manager) generateRizomeContent(srcConfig *Config, sourceProvider string) string {
	var content strings.Builder
	
	// Add header
	content.WriteString("# RIZOME.md\n\n")
	
	// Add project overview if available
	if srcConfig.ProjectOverview != "" {
		content.WriteString(srcConfig.ProjectOverview)
		content.WriteString("\n\n")
	}
	
	// Add common instructions
	if srcConfig.CommonInstructions != "" {
		content.WriteString("## Common Instructions\n\n")
		content.WriteString(srcConfig.CommonInstructions)
		content.WriteString("\n\n")
	}
	
	// Add provider overrides section
	content.WriteString("## Provider Overrides\n\n")
	
	// Add the source provider's specific instructions if available
	if override, exists := srcConfig.ProviderOverrides[sourceProvider]; exists && override != "" {
		content.WriteString(fmt.Sprintf("### %s\n", sourceProvider))
		content.WriteString(override)
		content.WriteString("\n\n")
	}
	
	// Add placeholders for other providers
	for _, provider := range srcConfig.Providers {
		if provider != sourceProvider {
			content.WriteString(fmt.Sprintf("### %s\n", provider))
			content.WriteString(fmt.Sprintf("%s-specific instructions\n\n", provider))
		}
	}
	
	// Inject timestamp
	return config.InjectTimestamp(content.String())
}

// syncProviderToProvider syncs from one provider to another
func (m *Manager) syncProviderToProvider(sourceConfig *Config, sourceProvider, destProvider string, dryRun, force bool) SyncResult {
	filename := fmt.Sprintf("%s.md", destProvider)
	filepath := filepath.Join(m.baseDir, filename)
	result := SyncResult{Provider: destProvider}
	
	// Check if file exists
	_, err := os.Stat(filepath)
	fileExists := !os.IsNotExist(err)
	
	if fileExists && !force && !dryRun {
		// TODO: Add interactive prompt for overwrite
	}
	
	// Generate content for destination provider
	// Use source config but apply for destination provider
	var content strings.Builder
	
	// Add header
	content.WriteString(fmt.Sprintf("# %s.md\n\n", destProvider))
	content.WriteString("This file is managed by Rizome CLI. Do not edit directly.\n")
	content.WriteString("Update RIZOME.md and run 'rizome sync' instead.\n\n")
	
	// Add project overview if available
	if sourceConfig.ProjectOverview != "" {
		content.WriteString(sourceConfig.ProjectOverview)
		content.WriteString("\n\n")
	}
	
	// Add common instructions if available
	if sourceConfig.CommonInstructions != "" {
		content.WriteString("## Common Instructions\n\n")
		content.WriteString(sourceConfig.CommonInstructions)
		content.WriteString("\n\n")
	}
	
	// Add provider-specific instructions from source if it's for this provider
	if override, exists := sourceConfig.ProviderOverrides[sourceProvider]; exists && override != "" {
		// Copy source provider's specific instructions as common for destination
		content.WriteString(fmt.Sprintf("## %s-Specific Instructions\n\n", destProvider))
		content.WriteString(fmt.Sprintf("(Copied from %s)\n\n", sourceProvider))
		content.WriteString(override)
		content.WriteString("\n")
	}
	
	// Inject timestamp
	finalContent := config.InjectTimestamp(content.String())
	
	if dryRun {
		result.Created = !fileExists
		result.Updated = fileExists
		return result
	}
	
	// Write file
	err = os.WriteFile(filepath, []byte(finalContent), 0644)
	if err != nil {
		result.Error = fmt.Errorf("failed to write %s: %w", filename, err)
		return result
	}
	
	result.Created = !fileExists
	result.Updated = fileExists
	
	return result
}

// updateProviderTimestamp updates the timestamp in a provider file
func (m *Manager) updateProviderTimestamp(provider string) error {
	filepath := filepath.Join(m.baseDir, fmt.Sprintf("%s.md", provider))
	
	// Read current content
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read %s.md: %w", provider, err)
	}
	
	// Inject/update timestamp
	updatedContent := config.InjectTimestamp(string(content))
	
	// Write back to file
	err = os.WriteFile(filepath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to update %s.md timestamp: %w", provider, err)
	}
	
	return nil
}
