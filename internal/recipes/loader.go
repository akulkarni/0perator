package recipes

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Loader loads and manages recipes
type Loader struct {
	recipesDir string
	recipes    map[string]*Recipe
}

// NewLoader creates a new recipe loader
func NewLoader(recipesDir string) *Loader {
	return &Loader{
		recipesDir: recipesDir,
		recipes:    make(map[string]*Recipe),
	}
}

// LoadAll loads all recipes from the recipes directory
func (l *Loader) LoadAll() error {
	// Create recipes directory if it doesn't exist
	if err := os.MkdirAll(l.recipesDir, 0755); err != nil {
		return fmt.Errorf("failed to create recipes directory: %w", err)
	}

	// Find all .yaml and .yml files
	patterns := []string{
		filepath.Join(l.recipesDir, "*.yaml"),
		filepath.Join(l.recipesDir, "*.yml"),
		filepath.Join(l.recipesDir, "*/*.yaml"),  // Subdirectories
		filepath.Join(l.recipesDir, "*/*.yml"),
	}

	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			log.Printf("Warning: failed to glob pattern %s: %v", pattern, err)
			continue
		}

		for _, file := range files {
			if err := l.LoadRecipe(file); err != nil {
				log.Printf("Warning: failed to load recipe %s: %v", file, err)
				continue
			}
		}
	}

	log.Printf("Loaded %d recipes from %s", len(l.recipes), l.recipesDir)
	return nil
}

// LoadRecipe loads a single recipe file
func (l *Loader) LoadRecipe(path string) error {
	recipe, err := LoadRecipe(path)
	if err != nil {
		return err
	}

	// Generate a key for the recipe
	key := l.recipeKey(recipe.Name)

	// Check for duplicates
	if existing, exists := l.recipes[key]; exists {
		return fmt.Errorf("duplicate recipe name '%s' (already loaded from another file)", existing.Name)
	}

	l.recipes[key] = recipe
	log.Printf("Loaded recipe: %s - %s", recipe.Name, recipe.Desc)

	return nil
}

// Get returns a recipe by name
func (l *Loader) Get(name string) (*Recipe, error) {
	key := l.recipeKey(name)
	recipe, exists := l.recipes[key]
	if !exists {
		return nil, fmt.Errorf("recipe '%s' not found", name)
	}
	return recipe, nil
}

// List returns all loaded recipes
func (l *Loader) List() []*Recipe {
	recipes := []*Recipe{}
	for _, recipe := range l.recipes {
		recipes = append(recipes, recipe)
	}
	return recipes
}

// ListNames returns all recipe names
func (l *Loader) ListNames() []string {
	names := []string{}
	for _, recipe := range l.recipes {
		names = append(names, recipe.Name)
	}
	return names
}

// Search finds recipes matching a query
func (l *Loader) Search(query string) []*Recipe {
	query = strings.ToLower(query)
	matches := []*Recipe{}

	for _, recipe := range l.recipes {
		// Search in name, description, and steps
		if strings.Contains(strings.ToLower(recipe.Name), query) ||
		   strings.Contains(strings.ToLower(recipe.Desc), query) {
			matches = append(matches, recipe)
			continue
		}

		// Search in steps
		for _, step := range recipe.Steps {
			if strings.Contains(strings.ToLower(step), query) {
				matches = append(matches, recipe)
				break
			}
		}
	}

	return matches
}

// Reload reloads all recipes from disk
func (l *Loader) Reload() error {
	l.recipes = make(map[string]*Recipe)
	return l.LoadAll()
}

// recipeKey generates a normalized key for a recipe name
func (l *Loader) recipeKey(name string) string {
	// Convert to lowercase and replace spaces with underscores
	key := strings.ToLower(name)
	key = strings.ReplaceAll(key, " ", "_")
	key = strings.ReplaceAll(key, "-", "_")
	return key
}

// GetRecipePath returns the expected path for a recipe file
func (l *Loader) GetRecipePath(name string) string {
	filename := l.recipeKey(name) + ".yaml"
	return filepath.Join(l.recipesDir, filename)
}

// SaveRecipe saves a recipe to disk
func (l *Loader) SaveRecipe(recipe *Recipe) error {
	path := l.GetRecipePath(recipe.Name)

	// Create YAML content
	content := fmt.Sprintf(`name: %s
desc: %s
inputs:
`, recipe.Name, recipe.Desc)

	// Add inputs
	for name, def := range recipe.Inputs {
		content += fmt.Sprintf("  %s: %s\n", name, def)
	}

	// Add steps
	content += "\nsteps:\n"
	for _, step := range recipe.Steps {
		content += fmt.Sprintf("  - %s\n", step)
	}

	// Write to file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to save recipe: %w", err)
	}

	// Add to loaded recipes
	l.recipes[l.recipeKey(recipe.Name)] = recipe

	return nil
}

// Package-level helper functions for easy access

var defaultLoader *Loader

// getDefaultLoader returns the default loader, creating it if necessary
func getDefaultLoader() *Loader {
	if defaultLoader == nil {
		// Look for recipes in common locations
		possiblePaths := []string{
			"recipes",
			"../recipes",
			"../../recipes",
			filepath.Join(os.Getenv("HOME"), ".0perator", "recipes"),
		}

		var recipesDir string
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				recipesDir = path
				break
			}
		}

		if recipesDir == "" {
			recipesDir = "recipes" // Default to local recipes dir
		}

		defaultLoader = NewLoader(recipesDir)
		if err := defaultLoader.LoadAll(); err != nil {
			log.Printf("Warning: failed to load recipes: %v", err)
		}
	}
	return defaultLoader
}

// Load loads a recipe by name using the default loader
func Load(name string) (*Recipe, error) {
	return getDefaultLoader().Get(name)
}

// List returns all recipe names using the default loader
func List() ([]string, error) {
	return getDefaultLoader().ListNames(), nil
}