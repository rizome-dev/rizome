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
	"slices"
	"strings"
)

// Provider represents an AI provider with metadata and configuration
type Provider struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Enabled     bool   `yaml:"enabled"`
	Category    string `yaml:"category,omitempty"`
}

// ProviderRegistry manages the collection of available providers
type ProviderRegistry struct {
	Providers []Provider `yaml:"providers"`
}

// GetDefaultProviders returns the default set of providers with descriptions
func GetDefaultProviders() []Provider {
	return []Provider{
		{
			Name:        "CLAUDE",
			Description: "Anthropic's Claude AI assistant",
			Enabled:     true,
			Category:    "Chat",
		},
		{
			Name:        "QWEN",
			Description: "Alibaba's Qwen AI model",
			Enabled:     true,
			Category:    "Chat",
		},
		{
			Name:        "CURSOR",
			Description: "AI-powered code editor",
			Enabled:     true,
			Category:    "Code Editor",
		},
		{
			Name:        "GEMINI",
			Description: "Google's Gemini AI model",
			Enabled:     true,
			Category:    "Chat",
		},
		{
			Name:        "WINDSURF",
			Description: "Codeium's AI-powered IDE",
			Enabled:     true,
			Category:    "Code Editor",
		},
	}
}

// GetEnabledProviders returns only the providers that are enabled
func (pr *ProviderRegistry) GetEnabledProviders() []string {
	var enabled []string
	for _, provider := range pr.Providers {
		if provider.Enabled {
			enabled = append(enabled, provider.Name)
		}
	}
	return enabled
}

// GetAllProviders returns all provider names regardless of enabled status
func (pr *ProviderRegistry) GetAllProviders() []string {
	var all []string
	for _, provider := range pr.Providers {
		all = append(all, provider.Name)
	}
	return all
}

// GetProvider returns a specific provider by name
func (pr *ProviderRegistry) GetProvider(name string) (*Provider, bool) {
	name = strings.ToUpper(name)
	for i, provider := range pr.Providers {
		if strings.ToUpper(provider.Name) == name {
			return &pr.Providers[i], true
		}
	}
	return nil, false
}

// UpdateProvider updates an existing provider or adds it if it doesn't exist
func (pr *ProviderRegistry) UpdateProvider(provider Provider) {
	provider.Name = strings.ToUpper(provider.Name)
	
	for i, existing := range pr.Providers {
		if strings.ToUpper(existing.Name) == provider.Name {
			pr.Providers[i] = provider
			return
		}
	}
	
	// Provider doesn't exist, add it
	pr.Providers = append(pr.Providers, provider)
}

// RemoveProvider removes a provider by name
func (pr *ProviderRegistry) RemoveProvider(name string) bool {
	name = strings.ToUpper(name)
	for i, provider := range pr.Providers {
		if strings.ToUpper(provider.Name) == name {
			pr.Providers = append(pr.Providers[:i], pr.Providers[i+1:]...)
			return true
		}
	}
	return false
}

// SetProviderEnabled enables or disables a provider
func (pr *ProviderRegistry) SetProviderEnabled(name string, enabled bool) error {
	name = strings.ToUpper(name)
	for i, provider := range pr.Providers {
		if strings.ToUpper(provider.Name) == name {
			pr.Providers[i].Enabled = enabled
			return nil
		}
	}
	return fmt.Errorf("provider '%s' not found", name)
}

// GetProvidersByCategory returns providers filtered by category
func (pr *ProviderRegistry) GetProvidersByCategory(category string) []Provider {
	var filtered []Provider
	for _, provider := range pr.Providers {
		if strings.EqualFold(provider.Category, category) {
			filtered = append(filtered, provider)
		}
	}
	return filtered
}

// GetCategories returns all unique categories
func (pr *ProviderRegistry) GetCategories() []string {
	categorySet := make(map[string]bool)
	for _, provider := range pr.Providers {
		if provider.Category != "" {
			categorySet[provider.Category] = true
		}
	}
	
	var categories []string
	for category := range categorySet {
		categories = append(categories, category)
	}
	
	slices.Sort(categories)
	return categories
}

// Validate ensures the provider registry is in a valid state
func (pr *ProviderRegistry) Validate() error {
	if pr.Providers == nil {
		return fmt.Errorf("providers list cannot be nil")
	}
	
	nameSet := make(map[string]bool)
	for _, provider := range pr.Providers {
		if provider.Name == "" {
			return fmt.Errorf("provider name cannot be empty")
		}
		
		upperName := strings.ToUpper(provider.Name)
		if nameSet[upperName] {
			return fmt.Errorf("duplicate provider name: %s", provider.Name)
		}
		nameSet[upperName] = true
	}
	
	return nil
}

// SortProviders sorts providers by name
func (pr *ProviderRegistry) SortProviders() {
	slices.SortFunc(pr.Providers, func(a, b Provider) int {
		return strings.Compare(a.Name, b.Name)
	})
}