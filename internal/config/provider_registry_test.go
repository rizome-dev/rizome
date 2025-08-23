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
	"reflect"
	"slices"
	"testing"
)

func TestGetDefaultProviders(t *testing.T) {
	providers := GetDefaultProviders()

	if len(providers) == 0 {
		t.Error("GetDefaultProviders should return at least one provider")
	}

	expectedProviders := []string{"CLAUDE", "QWEN", "CURSOR", "GEMINI", "WINDSURF"}
	if len(providers) != len(expectedProviders) {
		t.Errorf("Expected %d providers, got %d", len(expectedProviders), len(providers))
	}

	for _, expected := range expectedProviders {
		found := false
		for _, provider := range providers {
			if provider.Name == expected {
				found = true
				if !provider.Enabled {
					t.Errorf("Default provider %s should be enabled by default", expected)
				}
				if provider.Description == "" {
					t.Errorf("Default provider %s should have a description", expected)
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected provider %s not found in default providers", expected)
		}
	}
}

func TestProviderRegistryBasicOperations(t *testing.T) {
	registry := &ProviderRegistry{
		Providers: GetDefaultProviders(),
	}

	// Test GetEnabledProviders
	enabled := registry.GetEnabledProviders()
	if len(enabled) == 0 {
		t.Error("Should have enabled providers by default")
	}

	// Test GetAllProviders
	all := registry.GetAllProviders()
	if len(all) != len(registry.Providers) {
		t.Errorf("GetAllProviders should return %d providers, got %d", len(registry.Providers), len(all))
	}

	// Test GetProvider
	provider, exists := registry.GetProvider("CLAUDE")
	if !exists {
		t.Error("CLAUDE provider should exist")
	}
	if provider.Name != "CLAUDE" {
		t.Errorf("Expected provider name CLAUDE, got %s", provider.Name)
	}

	// Test case insensitive search
	provider, exists = registry.GetProvider("claude")
	if !exists {
		t.Error("Provider search should be case insensitive")
	}

	// Test non-existent provider
	_, exists = registry.GetProvider("NONEXISTENT")
	if exists {
		t.Error("Non-existent provider should not be found")
	}
}

func TestProviderRegistryEnableDisable(t *testing.T) {
	registry := &ProviderRegistry{
		Providers: GetDefaultProviders(),
	}

	// Disable a provider
	err := registry.SetProviderEnabled("CLAUDE", false)
	if err != nil {
		t.Fatalf("Failed to disable CLAUDE provider: %v", err)
	}

	// Check that provider is disabled
	provider, exists := registry.GetProvider("CLAUDE")
	if !exists {
		t.Error("CLAUDE provider should still exist")
	}
	if provider.Enabled {
		t.Error("CLAUDE provider should be disabled")
	}

	// Check enabled providers list doesn't include disabled provider
	enabled := registry.GetEnabledProviders()
	for _, name := range enabled {
		if name == "CLAUDE" {
			t.Error("Disabled provider should not appear in enabled list")
		}
	}

	// Re-enable the provider
	err = registry.SetProviderEnabled("CLAUDE", true)
	if err != nil {
		t.Fatalf("Failed to re-enable CLAUDE provider: %v", err)
	}

	provider, exists = registry.GetProvider("CLAUDE")
	if !exists {
		t.Error("CLAUDE provider should exist")
	}
	if !provider.Enabled {
		t.Error("CLAUDE provider should be enabled")
	}

	// Test setting status for non-existent provider
	err = registry.SetProviderEnabled("NONEXISTENT", true)
	if err == nil {
		t.Error("Setting status for non-existent provider should return error")
	}
}

func TestProviderRegistryUpdateAndRemove(t *testing.T) {
	registry := &ProviderRegistry{
		Providers: GetDefaultProviders(),
	}

	// Add a new provider
	newProvider := Provider{
		Name:        "TEST_PROVIDER",
		Description: "Test AI provider",
		Enabled:     true,
		Category:    "Test",
	}

	registry.UpdateProvider(newProvider)

	// Verify provider was added
	provider, exists := registry.GetProvider("TEST_PROVIDER")
	if !exists {
		t.Error("New provider should be added")
	}
	if provider.Description != "Test AI provider" {
		t.Errorf("Expected description 'Test AI provider', got '%s'", provider.Description)
	}

	// Update existing provider
	updatedProvider := Provider{
		Name:        "TEST_PROVIDER",
		Description: "Updated test provider",
		Enabled:     false,
		Category:    "Updated",
	}

	registry.UpdateProvider(updatedProvider)

	// Verify provider was updated
	provider, exists = registry.GetProvider("TEST_PROVIDER")
	if !exists {
		t.Error("Updated provider should exist")
	}
	if provider.Description != "Updated test provider" {
		t.Errorf("Expected updated description, got '%s'", provider.Description)
	}
	if provider.Enabled {
		t.Error("Provider should be disabled after update")
	}

	// Remove the provider
	removed := registry.RemoveProvider("TEST_PROVIDER")
	if !removed {
		t.Error("Provider should be successfully removed")
	}

	// Verify provider was removed
	_, exists = registry.GetProvider("TEST_PROVIDER")
	if exists {
		t.Error("Removed provider should not exist")
	}

	// Try to remove non-existent provider
	removed = registry.RemoveProvider("NONEXISTENT")
	if removed {
		t.Error("Removing non-existent provider should return false")
	}
}

func TestProviderRegistryCategories(t *testing.T) {
	providers := []Provider{
		{Name: "CLAUDE", Category: "Chat", Enabled: true},
		{Name: "CURSOR", Category: "Code Editor", Enabled: true},
		{Name: "TEST1", Category: "Chat", Enabled: true},
		{Name: "TEST2", Category: "Code Editor", Enabled: false},
		{Name: "TEST3", Category: "", Enabled: true}, // No category
	}

	registry := &ProviderRegistry{Providers: providers}

	// Test GetCategories
	categories := registry.GetCategories()
	expectedCategories := []string{"Chat", "Code Editor"}
	slices.Sort(expectedCategories)

	if !reflect.DeepEqual(categories, expectedCategories) {
		t.Errorf("Expected categories %v, got %v", expectedCategories, categories)
	}

	// Test GetProvidersByCategory
	chatProviders := registry.GetProvidersByCategory("Chat")
	if len(chatProviders) != 2 {
		t.Errorf("Expected 2 Chat providers, got %d", len(chatProviders))
	}

	codeProviders := registry.GetProvidersByCategory("Code Editor")
	if len(codeProviders) != 2 {
		t.Errorf("Expected 2 Code Editor providers, got %d", len(codeProviders))
	}

	// Test case insensitive category search
	chatProviders = registry.GetProvidersByCategory("chat")
	if len(chatProviders) != 2 {
		t.Error("Category search should be case insensitive")
	}

	// Test non-existent category
	unknownProviders := registry.GetProvidersByCategory("Unknown")
	if len(unknownProviders) != 0 {
		t.Errorf("Expected 0 providers for unknown category, got %d", len(unknownProviders))
	}
}

func TestProviderRegistryValidation(t *testing.T) {
	// Test valid registry
	validRegistry := &ProviderRegistry{
		Providers: GetDefaultProviders(),
	}

	err := validRegistry.Validate()
	if err != nil {
		t.Errorf("Valid registry should pass validation: %v", err)
	}

	// Test nil providers
	nilRegistry := &ProviderRegistry{
		Providers: nil,
	}

	err = nilRegistry.Validate()
	if err == nil {
		t.Error("Registry with nil providers should fail validation")
	}

	// Test empty provider name
	emptyNameRegistry := &ProviderRegistry{
		Providers: []Provider{
			{Name: "", Description: "Test", Enabled: true},
		},
	}

	err = emptyNameRegistry.Validate()
	if err == nil {
		t.Error("Registry with empty provider name should fail validation")
	}

	// Test duplicate provider names
	duplicateRegistry := &ProviderRegistry{
		Providers: []Provider{
			{Name: "CLAUDE", Description: "First", Enabled: true},
			{Name: "CLAUDE", Description: "Second", Enabled: false},
		},
	}

	err = duplicateRegistry.Validate()
	if err == nil {
		t.Error("Registry with duplicate provider names should fail validation")
	}

	// Test case insensitive duplicate detection
	caseInsensitiveDuplicateRegistry := &ProviderRegistry{
		Providers: []Provider{
			{Name: "CLAUDE", Description: "First", Enabled: true},
			{Name: "claude", Description: "Second", Enabled: false},
		},
	}

	err = caseInsensitiveDuplicateRegistry.Validate()
	if err == nil {
		t.Error("Registry with case insensitive duplicate names should fail validation")
	}
}

func TestProviderRegistrySortProviders(t *testing.T) {
	providers := []Provider{
		{Name: "ZEBRA", Enabled: true},
		{Name: "ALPHA", Enabled: true},
		{Name: "BETA", Enabled: false},
	}

	registry := &ProviderRegistry{Providers: providers}
	registry.SortProviders()

	expectedOrder := []string{"ALPHA", "BETA", "ZEBRA"}
	for i, provider := range registry.Providers {
		if provider.Name != expectedOrder[i] {
			t.Errorf("Expected provider at position %d to be %s, got %s", i, expectedOrder[i], provider.Name)
		}
	}
}

func TestProviderNameCaseHandling(t *testing.T) {
	registry := &ProviderRegistry{}

	// Add provider with mixed case
	provider := Provider{
		Name:    "MixedCase",
		Enabled: true,
	}

	registry.UpdateProvider(provider)

	// Verify provider name is stored in uppercase
	storedProvider, exists := registry.GetProvider("MIXEDCASE")
	if !exists {
		t.Error("Provider should be found by uppercase name")
	}
	if storedProvider.Name != "MIXEDCASE" {
		t.Errorf("Expected provider name to be uppercase, got %s", storedProvider.Name)
	}

	// Test case insensitive retrieval
	_, exists = registry.GetProvider("mixedcase")
	if !exists {
		t.Error("Provider should be found by lowercase name")
	}

	_, exists = registry.GetProvider("MixedCase")
	if !exists {
		t.Error("Provider should be found by mixed case name")
	}
}