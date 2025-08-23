# RIZOME.md

This is the master configuration file for the Rizome CLI project.

## Common Instructions

These are common instructions that apply to all AI providers:

- This is a Golang CLI tool built using charmbracelet tools and cobra
- The main command is `rizome sync` which synchronizes configuration files
- Follow Go best practices and proper error handling
- Use the existing code patterns from the opun project
- Maintain clean, readable, and well-documented code

## Provider Overrides

### CLAUDE
Claude-specific instructions:
- Focus on clean architecture and separation of concerns
- Use proper dependency injection patterns
- Ensure comprehensive error handling

### QWEN
Qwen-specific instructions:
- Pay attention to performance optimizations
- Use efficient algorithms and data structures
- Consider memory usage in implementations

### CURSOR  
Cursor-specific instructions:
- Emphasize code readability and maintainability
- Provide clear inline documentation
- Use descriptive variable and function names

### GEMINI
Gemini-specific instructions:
- Focus on modularity and reusability
- Implement proper testing strategies
- Consider edge cases in implementations