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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Template represents a RIZOME.md template
type Template struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Content     string `yaml:"content"`
}

// TemplateConfig represents the structure of the template configuration file
type TemplateConfig struct {
	Templates map[string]Template `yaml:"templates"`
}

// DefaultTemplate returns the default RIZOME.md template
func DefaultTemplate() Template {
	return Template{
		Name:        "Default Template",
		Description: "Standard RIZOME.md template with all supported providers",
		Content: `# RIZOME.md

Project overview and context.

## Common Instructions

Instructions that apply to all providers:
- Project type and technology stack
- Coding standards and conventions
- General best practices

## Provider Overrides

### CLAUDE
Claude-specific instructions

### QWEN
Qwen-specific instructions

### CURSOR
Cursor-specific instructions

### GEMINI
Gemini-specific instructions

### WINDSURF
Windsurf-specific instructions`,
	}
}

// Fallback text for empty template sections
const (
	DefaultCommonInstructions = `Instructions that apply to all providers:
- Project type and technology stack
- Coding standards and conventions
- General best practices`

	DefaultProviderInstructions = `Provider-specific instructions and preferences`
)

// GetDefaultTemplates returns the default set of templates
func GetDefaultTemplates() map[string]Template {
	return map[string]Template{
		"default": DefaultTemplate(),
	}
}

// TemplateManager handles template operations
type TemplateManager struct {
	configDir  string
	configFile string
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() (*TemplateManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".rizome")
	configFile := filepath.Join(configDir, "config.yaml")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &TemplateManager{
		configDir:  configDir,
		configFile: configFile,
	}, nil
}

// LoadConfig loads the template configuration from the config file
func (tm *TemplateManager) LoadConfig() (*TemplateConfig, error) {
	// Check if config file exists
	if _, err := os.Stat(tm.configFile); os.IsNotExist(err) {
		// Create default config with default templates
		config := &TemplateConfig{
			Templates: GetDefaultTemplates(),
		}

		// Save the default config
		if err := tm.SaveConfig(config); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}

		return config, nil
	}

	// Load existing config
	data, err := os.ReadFile(tm.configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config TemplateConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure default templates exist
	if config.Templates == nil {
		config.Templates = make(map[string]Template)
	}

	// Add missing default templates
	defaults := GetDefaultTemplates()
	for key, template := range defaults {
		if _, exists := config.Templates[key]; !exists {
			config.Templates[key] = template
		}
	}

	return &config, nil
}

// SaveConfig saves the template configuration to the config file
func (tm *TemplateManager) SaveConfig(config *TemplateConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(tm.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ListTemplates returns all available templates
func (tm *TemplateManager) ListTemplates() (map[string]Template, error) {
	config, err := tm.LoadConfig()
	if err != nil {
		return nil, err
	}

	return config.Templates, nil
}

// GetTemplate returns a specific template by key
func (tm *TemplateManager) GetTemplate(key string) (*Template, error) {
	config, err := tm.LoadConfig()
	if err != nil {
		return nil, err
	}

	template, exists := config.Templates[key]
	if !exists {
		return nil, fmt.Errorf("template '%s' not found", key)
	}

	return &template, nil
}

// SaveTemplate saves a template with the given key
func (tm *TemplateManager) SaveTemplate(key string, template Template) error {
	config, err := tm.LoadConfig()
	if err != nil {
		return err
	}

	config.Templates[key] = template

	return tm.SaveConfig(config)
}

// DeleteTemplate removes a template by key
func (tm *TemplateManager) DeleteTemplate(key string) error {
	config, err := tm.LoadConfig()
	if err != nil {
		return err
	}

	if _, exists := config.Templates[key]; !exists {
		return fmt.Errorf("template '%s' not found", key)
	}

	delete(config.Templates, key)

	return tm.SaveConfig(config)
}

// TemplateExists checks if a template exists
func (tm *TemplateManager) TemplateExists(key string) (bool, error) {
	config, err := tm.LoadConfig()
	if err != nil {
		return false, err
	}

	_, exists := config.Templates[key]
	return exists, nil
}

// InjectTimestamp adds or updates the current date timestamp for AI model grounding
func InjectTimestamp(content string) string {
	now := time.Now()
	
	// Format: <!-- Current Date: 2025-08-23 14:35:42 UTC -->
	newTimestamp := fmt.Sprintf("<!-- Current Date: %s -->", now.UTC().Format("2006-01-02 15:04:05 UTC"))
	
	lines := strings.Split(content, "\n")
	
	// Look for existing timestamp patterns to replace
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "<!-- Current Date:") ||
		   strings.HasPrefix(trimmedLine, "<!-- Last Updated:") || 
		   strings.HasPrefix(trimmedLine, "<!-- Generated:") {
			// Replace the existing timestamp
			lines[i] = newTimestamp
			return strings.Join(lines, "\n")
		}
	}
	
	// No existing timestamp found, add at the beginning
	if strings.TrimSpace(content) == "" {
		return newTimestamp + "\n\n"
	}
	
	// Check if content starts with other metadata comments
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "<!--") {
		// Find the end of the first comment block and insert after it
		for i, line := range lines {
			if strings.Contains(line, "-->") && 
			   !strings.HasPrefix(strings.TrimSpace(line), "<!-- Current Date:") &&
			   !strings.HasPrefix(strings.TrimSpace(line), "<!-- Last Updated:") &&
			   !strings.HasPrefix(strings.TrimSpace(line), "<!-- Generated:") {
				// Insert timestamp after the existing comment
				result := make([]string, 0, len(lines)+2)
				result = append(result, lines[:i+1]...)
				result = append(result, newTimestamp)
				result = append(result, "")
				result = append(result, lines[i+1:]...)
				return strings.Join(result, "\n")
			}
		}
	}
	
	// Default: prepend timestamp at the very beginning
	return newTimestamp + "\n\n" + content
}
