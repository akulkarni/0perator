package skills

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"sync"

	"github.com/adrg/frontmatter"
)

//go:embed all:*
var skillsFS embed.FS

var (
	cachedSkillsList string
	listSkillsOnce   sync.Once
	listSkillsErr    error
)

// Skill represents a skill with its metadata
type Skill struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Body        string `yaml:"-"`
}

// GetSkill returns the full Skill by name (directory name)
func GetSkill(name string) (Skill, error) {
	return parseSkill(name)
}

// ListSkills returns all available skills formatted as "{name} - {description}" newline delimited
func ListSkills() (string, error) {
	listSkillsOnce.Do(func() {
		cachedSkillsList, listSkillsErr = computeListSkills()
	})
	return cachedSkillsList, listSkillsErr
}

func computeListSkills() (string, error) {
	entries, err := skillsFS.ReadDir(".")
	if err != nil {
		return "", err
	}

	var lines []string
	for _, entry := range entries {
		if entry.IsDir() {
			path := entry.Name() + "/SKILL.md"
			if _, err := fs.Stat(skillsFS, path); err == nil {
				skill, err := parseSkill(entry.Name())
				if err != nil {
					return "", err
				}
				lines = append(lines, fmt.Sprintf("%s - %s", skill.Name, skill.Description))
			}
		}
	}
	return strings.Join(lines, "\n"), nil
}

// parseSkill reads and parses a skill from its directory
func parseSkill(dirName string) (Skill, error) {
	path := dirName + "/SKILL.md"
	content, err := skillsFS.ReadFile(path)
	if err != nil {
		return Skill{}, fmt.Errorf("skill '%s' not found", dirName)
	}

	var skill Skill
	body, err := frontmatter.Parse(bytes.NewReader(content), &skill)
	if err != nil {
		return Skill{}, fmt.Errorf("skill '%s': failed to parse frontmatter: %w", dirName, err)
	}
	if skill.Name == "" {
		return Skill{}, fmt.Errorf("skill '%s': name is required in frontmatter", dirName)
	}
	if skill.Description == "" {
		return Skill{}, fmt.Errorf("skill '%s': description is required in frontmatter", dirName)
	}
	if skill.Name != dirName {
		return Skill{}, fmt.Errorf("skill '%s': frontmatter name '%s' must match directory name", dirName, skill.Name)
	}
	skill.Body = string(body)

	return skill, nil
}
