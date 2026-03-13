// Fernwood - A lightweight agentic coding harness forked from PicoClaw
// License: MIT
//
// Copyright (c) 2026 Fernwood contributors

package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/strand1/fernwood/pkg/skills"
	"github.com/strand1/fernwood/pkg/utils"
	"gopkg.in/yaml.v3"
)

// SkillMetadata holds a skill's metadata from skill.yaml or SKILL.md frontmatter
type SkillMetadata struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Version     string   `yaml:"version" json:"version"`
	Enabled     bool     `yaml:"enabled" json:"enabled"`
	WhenToUse   []string `yaml:"when_to_use" json:"when_to_use,omitempty"`
	AutoTrigger bool     `yaml:"auto_trigger" json:"auto_trigger"`
	Registry    string   `yaml:"registry" json:"registry,omitempty"`
	Slug        string   `yaml:"slug" json:"slug,omitempty"`
	InstalledAt string   `yaml:"installed_at" json:"installed_at,omitempty"`
	UpdatedAt   string   `yaml:"updated_at" json:"updated_at,omitempty"`
}

// RegisterSkillCommands registers all skill commands to the registry.
// workspace: directory where skills/ lives
func RegisterSkillCommands(registry *CommandRegistry, workspace string) {
	skillsDir := filepath.Join(workspace, "skills")

	registry.Register("skill search", "Search for installable skills from registries", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("skill search: usage: skill search <query> [--limit N] [--registry clawhub]")
		}

		// Parse arguments
		query := ""
		limit := 5
		registryName := "clawhub"

		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--limit", "-l":
				if i+1 < len(args) {
					if l, err := parseInt(args[i+1]); err == nil && l > 0 && l <= 20 {
						limit = l
					}
					i++
				}
			case "--registry", "-r":
				if i+1 < len(args) {
					registryName = args[i+1]
					i++
				}
			default:
				if query == "" {
					query = args[i]
				}
			}
		}

		if query == "" {
			return "", fmt.Errorf("skill search: query is required")
		}

		return cmdSkillSearch(workspace, query, limit, registryName)
	})

	registry.Register("skill install", "Install a skill from a registry", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("skill install: usage: skill install <slug> [--registry clawhub] [--force]")
		}

		slug := ""
		registryName := "clawhub"
		force := false

		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--registry", "-r":
				if i+1 < len(args) {
					registryName = args[i+1]
					i++
				}
			case "--force", "-f":
				force = true
			default:
				if slug == "" {
					slug = args[i]
				}
			}
		}

		if slug == "" {
			return "", fmt.Errorf("skill install: slug is required")
		}

		return cmdSkillInstall(workspace, slug, registryName, force)
	})

	registry.Register("skill list", "List installed skills", func(args []string, stdin string) (string, error) {
		showInstalled := false
		showAvailable := false
		registryName := ""

		for _, arg := range args {
			switch arg {
			case "--installed", "-i":
				showInstalled = true
			case "--available", "-a":
				showAvailable = true
			case "--registry", "-r":
				// Next arg would be registry name, but we'd need more complex parsing
				// For now, just flag presence
				registryName = "clawhub"
			}
		}

		// Default: show installed if no flags
		if !showInstalled && !showAvailable {
			showInstalled = true
		}

		return cmdSkillList(skillsDir, showInstalled, showAvailable, registryName)
	})

	registry.Register("skill info", "Show detailed information about a skill", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("skill info: usage: skill info <slug|name>")
		}

		return cmdSkillInfo(skillsDir, args[0])
	})

	registry.Register("skill update", "Update an installed skill", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("skill update: usage: skill update <slug|name> [--force]")
		}

		slug := args[0]
		force := false

		for i := 1; i < len(args); i++ {
			if args[i] == "--force" || args[i] == "-f" {
				force = true
			}
		}

		return cmdSkillUpdate(workspace, slug, force)
	})

	registry.Register("skill uninstall", "Uninstall/remove a skill", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("skill uninstall: usage: skill uninstall <name>")
		}

		return cmdSkillUninstall(skillsDir, args[0])
	})

	registry.Register("skill enable", "Enable a disabled skill", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("skill enable: usage: skill enable <name>")
		}

		return cmdSkillEnable(skillsDir, args[0], true)
	})

	registry.Register("skill disable", "Disable a skill without uninstalling", func(args []string, stdin string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("skill disable: usage: skill disable <name>")
		}

		return cmdSkillEnable(skillsDir, args[0], false)
	})

	// Aliases
	registry.RegisterAlias("skill ls", "skill list")
	registry.RegisterAlias("skill rm", "skill uninstall")
}

// cmdSkillSearch searches for skills using the registry manager
func cmdSkillSearch(workspace, query string, limit int, registryName string) (string, error) {
	// Load config to get registry settings
	cfg, err := loadSkillsConfig(workspace)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	// Create registry manager
	registryMgr := createRegistryManager(cfg)

	// Search
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := registryMgr.SearchAll(ctx, query, limit)
	if err != nil {
		return "", fmt.Errorf("skill search failed: %w", err)
	}

	if len(results) == 0 {
		return fmt.Sprintf("No skills found for query: %q", query), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d skills for %q:\n\n", len(results), query))

	for i, r := range results {
		sb.WriteString(fmt.Sprintf("%d. **%s** (v%s)\n", i+1, r.Slug, r.Version))
		sb.WriteString(fmt.Sprintf("   %s\n", r.Summary))
		if r.RegistryName != "" {
			sb.WriteString(fmt.Sprintf("   Registry: %s\n", r.RegistryName))
		}
		if i < len(results)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

// cmdSkillInstall installs a skill from a registry
func cmdSkillInstall(workspace, slug, registryName string, force bool) (string, error) {
	// Validate slug
	if err := utils.ValidateSkillIdentifier(slug); err != nil {
		return "", fmt.Errorf("invalid slug: %w", err)
	}

	// Load config and create registry manager
	cfg, err := loadSkillsConfig(workspace)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	registryMgr := createRegistryManager(cfg)

	// Get the specific registry
	reg := registryMgr.GetRegistry(registryName)
	if reg == nil {
		return "", fmt.Errorf("registry %q not found", registryName)
	}

	// Check if already installed
	skillsDir := filepath.Join(workspace, "skills")
	targetDir := filepath.Join(skillsDir, slug)
	if !force {
		if _, err := os.Stat(targetDir); err == nil {
			return fmt.Sprintf("Skill %q is already installed at %s. Use --force to reinstall.", slug, targetDir), nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Download and install from registry
	_, err = reg.DownloadAndInstall(ctx, slug, "", targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to install skill: %w", err)
	}

	return fmt.Sprintf("Successfully installed skill %q from %s", slug, registryName), nil
}

// cmdSkillList lists installed skills
func cmdSkillList(skillsDir string, showInstalled, showAvailable bool, registryName string) (string, error) {
	var sb strings.Builder

	if showInstalled {
		// Scan local skills directory
		entries, err := os.ReadDir(skillsDir)
		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("failed to read skills directory: %w", err)
		}

		if len(entries) == 0 {
			sb.WriteString("No skills installed.\n")
		} else {
			sb.WriteString("Installed skills:\n\n")

			var skillInfos []string
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}

				skillPath := filepath.Join(skillsDir, entry.Name())
				meta, err := loadSkillMetadata(skillPath)
				if err != nil {
					// Fallback to directory name
					skillInfos = append(skillInfos, fmt.Sprintf("  %s", entry.Name()))
					continue
				}

				status := ""
				if !meta.Enabled {
					status = " (disabled)"
				}

				desc := meta.Description
				if desc == "" {
					desc = "No description"
				}

				skillInfos = append(skillInfos, fmt.Sprintf("  %s — %s%s", entry.Name(), desc, status))
			}

			// Sort alphabetically
			sort.Strings(skillInfos)
			for _, info := range skillInfos {
				sb.WriteString(info + "\n")
			}
		}
	}

	if showAvailable {
		sb.WriteString("\nAvailable from registry (use 'skill search <query>' to find more):\n")
		sb.WriteString("  Use 'skill install <slug>' to install\n")
	}

	return sb.String(), nil
}

// cmdSkillInfo shows detailed information about a skill
func cmdSkillInfo(skillsDir, identifier string) (string, error) {
	// Check if it's a local skill
	skillPath := filepath.Join(skillsDir, identifier)
	if info, err := os.Stat(skillPath); err == nil && info.IsDir() {
		return showLocalSkillInfo(skillPath)
	}

	// Not a local skill, try to fetch from registry
	return fmt.Sprintf("Skill %q not found locally. Use 'skill search %s' to find it in registries.", identifier, identifier), nil
}

// cmdSkillUpdate updates an installed skill
func cmdSkillUpdate(workspace, slug string, force bool) (string, error) {
	// Determine skill name from slug
	skillName := slug
	if idx := strings.LastIndex(slug, "/"); idx >= 0 {
		skillName = slug[idx+1:]
	}

	skillsDir := filepath.Join(workspace, "skills")
	skillPath := filepath.Join(skillsDir, skillName)

	// Check if skill exists
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		return "", fmt.Errorf("skill %q not found. Install it first with 'skill install %s'", skillName, slug)
	}

	// Load metadata to get registry info
	meta, err := loadSkillMetadata(skillPath)
	if err != nil {
		return "", fmt.Errorf("failed to load skill metadata: %w", err)
	}

	registryName := meta.Registry
	if registryName == "" {
		registryName = "clawhub"
	}

	// Reinstall from registry
	return cmdSkillInstall(workspace, slug, registryName, true)
}

// cmdSkillUninstall removes a skill
func cmdSkillUninstall(skillsDir, skillName string) (string, error) {
	skillPath := filepath.Join(skillsDir, skillName)

	// Check if skill exists
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		return "", fmt.Errorf("skill %q not found", skillName)
	}

	// Remove the directory
	if err := os.RemoveAll(skillPath); err != nil {
		return "", fmt.Errorf("failed to remove skill: %w", err)
	}

	return fmt.Sprintf("Successfully uninstalled skill %q", skillName), nil
}

// cmdSkillEnable enables or disables a skill
func cmdSkillEnable(skillsDir, skillName string, enabled bool) (string, error) {
	skillPath := filepath.Join(skillsDir, skillName)

	// Check if skill exists
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		return "", fmt.Errorf("skill %q not found", skillName)
	}

	// Load or create skill.yaml
	meta, err := loadSkillMetadata(skillPath)
	if err != nil {
		// Create new metadata
		meta = &SkillMetadata{
			Name:    skillName,
			Enabled: enabled,
		}
	} else {
		meta.Enabled = enabled
	}

	// Write updated metadata
	if err := saveSkillMetadata(skillPath, meta); err != nil {
		return "", fmt.Errorf("failed to update skill metadata: %w", err)
	}

	action := "enabled"
	if !enabled {
		action = "disabled"
	}

	return fmt.Sprintf("Skill %q %s", skillName, action), nil
}

// Helper functions

func loadSkillMetadata(skillPath string) (*SkillMetadata, error) {
	// Try skill.yaml first
	yamlPath := filepath.Join(skillPath, "skill.yaml")
	if data, err := os.ReadFile(yamlPath); err == nil {
		var meta SkillMetadata
		if err := yaml.Unmarshal(data, &meta); err != nil {
			return nil, fmt.Errorf("failed to parse skill.yaml: %w", err)
		}
		return &meta, nil
	}

	// Fallback to SKILL.md frontmatter
	mdPath := filepath.Join(skillPath, "SKILL.md")
	if data, err := os.ReadFile(mdPath); err == nil {
		meta := parseSkillMarkdownFrontmatter(string(data))
		return meta, nil
	}

	return nil, fmt.Errorf("no skill.yaml or SKILL.md found")
}

func saveSkillMetadata(skillPath string, meta *SkillMetadata) error {
	yamlPath := filepath.Join(skillPath, "skill.yaml")

	data, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return os.WriteFile(yamlPath, data, 0o644)
}

func parseSkillMarkdownFrontmatter(content string) *SkillMetadata {
	meta := &SkillMetadata{
		Enabled: true,
	}

	if !strings.HasPrefix(content, "---\n") {
		return meta
	}

	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return meta
	}

	frontMatter := content[4 : 4+end]
	if err := yaml.Unmarshal([]byte(frontMatter), meta); err != nil {
		return meta
	}

	return meta
}

func showLocalSkillInfo(skillPath string) (string, error) {
	meta, err := loadSkillMetadata(skillPath)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Skill: %s\n", meta.Name))
	if meta.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", meta.Description))
	}
	if meta.Version != "" {
		sb.WriteString(fmt.Sprintf("Version: %s\n", meta.Version))
	}
	sb.WriteString(fmt.Sprintf("Status: %s\n", map[bool]string{true: "enabled", false: "disabled"}[meta.Enabled]))

	if len(meta.WhenToUse) > 0 {
		sb.WriteString("\nWhen to use:\n")
		for _, w := range meta.WhenToUse {
			sb.WriteString(fmt.Sprintf("  - %s\n", w))
		}
	}

	// Show SKILL.md preview
	mdPath := filepath.Join(skillPath, "SKILL.md")
	if data, err := os.ReadFile(mdPath); err == nil {
		content := string(data)
		// Show first 500 chars after frontmatter
		if idx := strings.Index(content, "\n---\n"); idx >= 0 {
			content = content[idx+5:]
		}
		if len(content) > 500 {
			content = content[:500] + "..."
		}
		sb.WriteString(fmt.Sprintf("\nPreview:\n%s\n", content))
	}

	return sb.String(), nil
}

func loadSkillsConfig(workspace string) (*skills.RegistryConfig, error) {
	// For now, return a default config
	// In production, this would load from config.json
	return &skills.RegistryConfig{
		MaxConcurrentSearches: 3,
		ClawHub: skills.ClawHubConfig{
			Enabled: true,
			BaseURL: "https://clawhub.io",
		},
	}, nil
}

func createRegistryManager(cfg *skills.RegistryConfig) *skills.RegistryManager {
	return skills.NewRegistryManagerFromConfig(*cfg)
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
