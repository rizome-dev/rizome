#!/usr/bin/env bash

set -euo pipefail

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Command line options
FORCE_REMOVE_CONFIG=false

# Print colored output
print_error() { echo -e "${RED}✗ $1${NC}" >&2; }
print_success() { echo -e "${GREEN}✓ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠ $1${NC}"; }
print_info() { echo -e "$1"; }

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --force)
                FORCE_REMOVE_CONFIG=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Show help message
show_help() {
    echo "Rizome Uninstaller"
    echo "=================="
    echo
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  --force    Remove all configuration files without prompting"
    echo "  -h, --help Show this help message"
    echo
    echo "By default, configuration files in ~/.rizome are preserved for testing."
    echo "Use --force to remove all traces including configuration."
}

# Check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        print_error "This script should not be run as root"
        exit 1
    fi
}

# Detect if installed via Homebrew
check_brew_installation() {
    if command -v brew >/dev/null 2>&1; then
        if brew list rizome >/dev/null 2>&1; then
            print_warning "Rizome appears to be installed via Homebrew"
            print_info "Please run: brew uninstall rizome"
            print_info ""
            read -p "Continue with manual uninstall anyway? (y/N) " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                exit 0
            fi
        fi
    fi
}

# Remove binary
remove_binary() {
    local binary_paths=(
        "/usr/local/bin/rizome"
        "$HOME/.local/bin/rizome"
        "$HOME/bin/rizome"
        "/opt/homebrew/bin/rizome"
    )
    
    local removed=false
    for path in "${binary_paths[@]}"; do
        if [[ -f "$path" ]]; then
            print_info "Removing binary: $path"
            if rm -f "$path" 2>/dev/null || sudo rm -f "$path"; then
                print_success "Removed binary: $path"
                removed=true
            else
                print_error "Failed to remove binary: $path"
            fi
        fi
    done
    
    if [[ "$removed" == false ]]; then
        print_warning "No rizome binary found in standard locations"
    fi
}

# Remove configuration directory
remove_config() {
    local config_dir="$HOME/.rizome"
    
    if [[ -d "$config_dir" ]]; then
        print_info "Found configuration directory: $config_dir"
        
        if [[ "$FORCE_REMOVE_CONFIG" == true ]]; then
            print_info "Force removing configuration directory..."
            rm -rf "$config_dir"
            print_success "Removed configuration directory"
        else
            # Check if directory has content
            if [[ -n "$(ls -A "$config_dir" 2>/dev/null)" ]]; then
                print_warning "Configuration directory contains:"
                ls -la "$config_dir" | head -10
                if [[ $(ls -1 "$config_dir" | wc -l) -gt 9 ]]; then
                    print_info "... and $(( $(ls -1 "$config_dir" | wc -l) - 9 )) more items"
                fi
                echo
                print_info "By default, configuration files are preserved for testing purposes."
                print_info "Use --force flag to remove all configuration files."
                echo
                read -p "Remove configuration directory anyway? (y/N) " -n 1 -r
                echo
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    rm -rf "$config_dir"
                    print_success "Removed configuration directory"
                else
                    print_info "Keeping configuration directory for testing"
                fi
            else
                rm -rf "$config_dir"
                print_success "Removed empty configuration directory"
            fi
        fi
    else
        print_info "No configuration directory found"
    fi
}

# Remove shell completions
remove_completions() {
    local completion_files=(
        # Bash completions
        "/usr/local/etc/bash_completion.d/rizome"
        "/etc/bash_completion.d/rizome"
        "$HOME/.local/share/bash-completion/completions/rizome"
        
        # Zsh completions
        "/usr/local/share/zsh/site-functions/_rizome"
        "/usr/share/zsh/site-functions/_rizome"
        "$HOME/.zsh/completions/_rizome"
        
        # Fish completions
        "/usr/local/share/fish/vendor_completions.d/rizome.fish"
        "/usr/share/fish/vendor_completions.d/rizome.fish"
        "$HOME/.config/fish/completions/rizome.fish"
        
        # Homebrew completions
        "/opt/homebrew/share/zsh/site-functions/_rizome"
        "/opt/homebrew/etc/bash_completion.d/rizome"
    )
    
    local removed=false
    for file in "${completion_files[@]}"; do
        if [[ -f "$file" ]]; then
            if rm -f "$file" 2>/dev/null || sudo rm -f "$file"; then
                print_success "Removed completion file: $file"
                removed=true
            else
                print_error "Failed to remove completion file: $file"
            fi
        fi
    done
    
    if [[ "$removed" == false ]]; then
        print_info "No shell completion files found"
    fi
}

# Remove from PATH if added to shell configs
check_path_modifications() {
    local shell_configs=(
        "$HOME/.bashrc"
        "$HOME/.bash_profile"
        "$HOME/.zshrc"
        "$HOME/.profile"
        "$HOME/.zprofile"
    )
    
    local found=false
    for config in "${shell_configs[@]}"; do
        if [[ -f "$config" ]] && grep -q "rizome" "$config" 2>/dev/null; then
            print_warning "Found 'rizome' references in $config"
            found=true
        fi
    done
    
    if [[ "$found" == true ]]; then
        print_info "Please manually check and remove any rizome-related PATH modifications from your shell configuration files"
    fi
}

# Remove any RIZOME.md files (optional cleanup)
check_rizome_files() {
    local found_files=()
    
    # Look for RIZOME.md files in common project directories
    if command -v find >/dev/null 2>&1; then
        while IFS= read -r -d '' file; do
            found_files+=("$file")
        done < <(find "$HOME" -name "RIZOME.md" -type f -not -path "*/.*" -print0 2>/dev/null | head -20)
    fi
    
    if [[ ${#found_files[@]} -gt 0 ]]; then
        print_warning "Found RIZOME.md files in your projects:"
        for file in "${found_files[@]}"; do
            print_info "  $file"
        done
        print_info ""
        print_info "These project files were created by rizome and can be kept for your projects."
        print_info "Remove them manually if no longer needed."
    fi
}

# Main uninstall function
main() {
    parse_args "$@"
    
    echo "Rizome Uninstaller"
    echo "=================="
    echo
    
    check_root
    check_brew_installation
    
    print_info "This will remove:"
    print_info "  • Rizome binary"
    if [[ "$FORCE_REMOVE_CONFIG" == true ]]; then
        print_info "  • Configuration directory (~/.rizome) [FORCED]"
    else
        print_info "  • Configuration directory (~/.rizome) [will prompt, preserved by default]"
    fi
    print_info "  • Shell completion files"
    echo
    
    if [[ "$FORCE_REMOVE_CONFIG" == false ]]; then
        print_info "Note: Configuration files are preserved by default for testing."
        print_info "Use --force flag to remove all configuration without prompting."
        echo
    fi
    
    read -p "Continue with uninstall? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Uninstall cancelled"
        exit 0
    fi
    
    echo
    remove_binary
    remove_config
    remove_completions
    check_path_modifications
    check_rizome_files
    
    echo
    if [[ "$FORCE_REMOVE_CONFIG" == true ]]; then
        print_success "Rizome has been completely uninstalled"
    else
        print_success "Rizome has been uninstalled (configuration preserved)"
    fi
    
    # Check if rizome is still in PATH
    if command -v rizome >/dev/null 2>&1; then
        print_warning "rizome command is still available in your PATH"
        print_info "Location: $(command -v rizome)"
        print_info "You may need to restart your shell or manually remove this file"
    fi
}

# Run main function
main "$@"