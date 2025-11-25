package sites

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/xerudro/DASHBOARD-v2/internal/models"
)

// TemplateConfig represents a deployment template configuration
type TemplateConfig struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Framework   string                 `json:"framework,omitempty"`
	Defaults    map[string]interface{} `json:"defaults"`
	Required    []string               `json:"required"`
	Optional    []string               `json:"optional"`
}

// TemplateManager manages deployment templates
type TemplateManager struct {
	templates map[string]*TemplateConfig
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[string]*TemplateConfig),
	}
	tm.loadTemplates()
	return tm
}

// loadTemplates loads all available deployment templates
func (tm *TemplateManager) loadTemplates() {
	tm.templates = map[string]*TemplateConfig{
		"wordpress": {
			ID:          "wordpress",
			Name:        "WordPress",
			Description: "Popular CMS with one-click installation",
			Type:        models.SiteTypeWordPress,
			Defaults: map[string]interface{}{
				"php_version":   "8.2",
				"database_type": "mysql",
				"admin_user":    "admin",
				"site_title":    "My WordPress Site",
				"plugins":       []string{"woocommerce", "contact-form-7"},
				"themes":        []string{"astra", "generatepress"},
			},
			Required: []string{"admin_email", "admin_pass", "site_title"},
			Optional: []string{"plugins", "themes", "multisite"},
		},
		"laravel": {
			ID:          "laravel",
			Name:        "Laravel",
			Description: "PHP framework for modern web applications",
			Type:        models.SiteTypeLaravel,
			Framework:   "laravel",
			Defaults: map[string]interface{}{
				"php_version":    "8.2",
				"database_type":  "mysql",
				"app_env":        "production",
				"app_debug":      "false",
				"cache_driver":   "redis",
				"session_driver": "redis",
				"queue_driver":   "redis",
			},
			Required: []string{"app_key", "git_repo"},
			Optional: []string{"git_branch", "env_vars"},
		},
		"nodejs": {
			ID:          "nodejs",
			Name:        "Node.js",
			Description: "JavaScript runtime for server-side applications",
			Type:        models.SiteTypeNodeJS,
			Defaults: map[string]interface{}{
				"nodejs_version": "18",
				"app_port":       "3000",
				"pm2_instances":  "1",
				"start_command":  "npm start",
				"build_command":  "npm run build",
			},
			Required: []string{"git_repo"},
			Optional: []string{"git_branch", "start_command", "build_command", "env_vars"},
		},
		"nextjs": {
			ID:          "nextjs",
			Name:        "Next.js",
			Description: "React framework for production web applications",
			Type:        models.SiteTypeNodeJS,
			Framework:   "nextjs",
			Defaults: map[string]interface{}{
				"nodejs_version": "18",
				"app_port":       "3000",
				"start_command":  "npm start",
				"build_command":  "npm run build",
			},
			Required: []string{"git_repo"},
			Optional: []string{"git_branch", "env_vars"},
		},
		"nuxtjs": {
			ID:          "nuxtjs",
			Name:        "Nuxt.js",
			Description: "Vue.js framework for universal applications",
			Type:        models.SiteTypeNodeJS,
			Framework:   "nuxtjs",
			Defaults: map[string]interface{}{
				"nodejs_version": "18",
				"app_port":       "3000",
				"start_command":  "npm start",
				"build_command":  "npm run build",
			},
			Required: []string{"git_repo"},
			Optional: []string{"git_branch", "env_vars"},
		},
		"static": {
			ID:          "static",
			Name:        "Static Site",
			Description: "HTML/CSS/JS static websites",
			Type:        models.SiteTypeStatic,
			Defaults: map[string]interface{}{
				"build_tool":    "none",
				"build_command": "",
			},
			Required: []string{"git_repo"},
			Optional: []string{"git_branch", "build_tool", "build_command"},
		},
		"hugo": {
			ID:          "hugo",
			Name:        "Hugo",
			Description: "Fast static site generator",
			Type:        models.SiteTypeStatic,
			Framework:   "hugo",
			Defaults: map[string]interface{}{
				"build_tool":    "hugo",
				"build_command": "hugo --minify",
			},
			Required: []string{"git_repo"},
			Optional: []string{"git_branch", "build_command"},
		},
		"jekyll": {
			ID:          "jekyll",
			Name:        "Jekyll",
			Description: "Ruby-based static site generator",
			Type:        models.SiteTypeStatic,
			Framework:   "jekyll",
			Defaults: map[string]interface{}{
				"build_tool":    "jekyll",
				"build_command": "jekyll build",
			},
			Required: []string{"git_repo"},
			Optional: []string{"git_branch", "build_command"},
		},
		"gatsby": {
			ID:          "gatsby",
			Name:        "Gatsby",
			Description: "React-based static site generator",
			Type:        models.SiteTypeStatic,
			Framework:   "gatsby",
			Defaults: map[string]interface{}{
				"nodejs_version": "18",
				"build_tool":     "npm",
				"build_command":  "npm run build",
			},
			Required: []string{"git_repo"},
			Optional: []string{"git_branch", "build_command"},
		},
		"php": {
			ID:          "php",
			Name:        "PHP Application",
			Description: "Generic PHP web application",
			Type:        models.SiteTypePHP,
			Defaults: map[string]interface{}{
				"php_version":   "8.2",
				"framework":     "none",
				"database_type": "mysql",
			},
			Required: []string{"git_repo"},
			Optional: []string{"git_branch", "framework", "database_type", "env_vars"},
		},
		"codeigniter": {
			ID:          "codeigniter",
			Name:        "CodeIgniter",
			Description: "PHP framework for rapid development",
			Type:        models.SiteTypePHP,
			Framework:   "codeigniter",
			Defaults: map[string]interface{}{
				"php_version":   "8.2",
				"database_type": "mysql",
			},
			Required: []string{"git_repo"},
			Optional: []string{"git_branch", "env_vars"},
		},
		"symfony": {
			ID:          "symfony",
			Name:        "Symfony",
			Description: "PHP framework for enterprise applications",
			Type:        models.SiteTypePHP,
			Framework:   "symfony",
			Defaults: map[string]interface{}{
				"php_version":   "8.2",
				"database_type": "mysql",
			},
			Required: []string{"git_repo"},
			Optional: []string{"git_branch", "env_vars"},
		},
		"cakephp": {
			ID:          "cakephp",
			Name:        "CakePHP",
			Description: "PHP framework for rapid development",
			Type:        models.SiteTypePHP,
			Framework:   "cakephp",
			Defaults: map[string]interface{}{
				"php_version":   "8.2",
				"database_type": "mysql",
			},
			Required: []string{"git_repo"},
			Optional: []string{"git_branch", "env_vars"},
		},
	}
}

// GetTemplate returns a template by ID
func (tm *TemplateManager) GetTemplate(id string) (*TemplateConfig, error) {
	template, exists := tm.templates[id]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	return template, nil
}

// ListTemplates returns all available templates
func (tm *TemplateManager) ListTemplates() []*TemplateConfig {
	var templates []*TemplateConfig
	for _, template := range tm.templates {
		templates = append(templates, template)
	}
	return templates
}

// ListTemplatesByType returns templates filtered by site type
func (tm *TemplateManager) ListTemplatesByType(siteType string) []*TemplateConfig {
	var templates []*TemplateConfig
	for _, template := range tm.templates {
		if template.Type == siteType {
			templates = append(templates, template)
		}
	}
	return templates
}

// ValidateTemplate validates a template configuration
func (tm *TemplateManager) ValidateTemplate(templateID string, config map[string]interface{}) error {
	template, err := tm.GetTemplate(templateID)
	if err != nil {
		return err
	}

	// Check required fields
	for _, required := range template.Required {
		if _, exists := config[required]; !exists {
			return fmt.Errorf("required field missing: %s", required)
		}
	}

	// Validate specific fields
	if err := tm.validateTemplateFields(template, config); err != nil {
		return err
	}

	return nil
}

// validateTemplateFields validates template-specific fields
func (tm *TemplateManager) validateTemplateFields(template *TemplateConfig, config map[string]interface{}) error {
	switch template.Type {
	case models.SiteTypeWordPress:
		return tm.validateWordPressConfig(config)
	case models.SiteTypeLaravel:
		return tm.validateLaravelConfig(config)
	case models.SiteTypeNodeJS:
		return tm.validateNodeJSConfig(config)
	case models.SiteTypeStatic:
		return tm.validateStaticConfig(config)
	case models.SiteTypePHP:
		return tm.validatePHPConfig(config)
	}
	return nil
}

// validateWordPressConfig validates WordPress-specific configuration
func (tm *TemplateManager) validateWordPressConfig(config map[string]interface{}) error {
	if email, ok := config["admin_email"].(string); ok {
		if !strings.Contains(email, "@") {
			return fmt.Errorf("invalid admin email format")
		}
	}

	if title, ok := config["site_title"].(string); ok {
		if len(strings.TrimSpace(title)) == 0 {
			return fmt.Errorf("site title cannot be empty")
		}
	}

	return nil
}

// validateLaravelConfig validates Laravel-specific configuration
func (tm *TemplateManager) validateLaravelConfig(config map[string]interface{}) error {
	if appKey, ok := config["app_key"].(string); ok {
		if len(appKey) < 32 {
			return fmt.Errorf("laravel app key must be at least 32 characters")
		}
	}

	if gitRepo, ok := config["git_repo"].(string); ok {
		if !strings.HasPrefix(gitRepo, "http") && !strings.HasPrefix(gitRepo, "git@") {
			return fmt.Errorf("invalid git repository URL")
		}
	}

	return nil
}

// validateNodeJSConfig validates Node.js-specific configuration
func (tm *TemplateManager) validateNodeJSConfig(config map[string]interface{}) error {
	if port, ok := config["app_port"].(string); ok {
		if port != "" && !isValidPort(port) {
			return fmt.Errorf("invalid port number: %s", port)
		}
	}

	if gitRepo, ok := config["git_repo"].(string); ok {
		if !strings.HasPrefix(gitRepo, "http") && !strings.HasPrefix(gitRepo, "git@") {
			return fmt.Errorf("invalid git repository URL")
		}
	}

	return nil
}

// validateStaticConfig validates static site configuration
func (tm *TemplateManager) validateStaticConfig(config map[string]interface{}) error {
	if gitRepo, ok := config["git_repo"].(string); ok {
		if !strings.HasPrefix(gitRepo, "http") && !strings.HasPrefix(gitRepo, "git@") {
			return fmt.Errorf("invalid git repository URL")
		}
	}

	if buildTool, ok := config["build_tool"].(string); ok {
		validTools := []string{"none", "hugo", "jekyll", "npm", "yarn", "gulp", "grunt"}
		if !contains(validTools, buildTool) {
			return fmt.Errorf("unsupported build tool: %s", buildTool)
		}
	}

	return nil
}

// validatePHPConfig validates PHP application configuration
func (tm *TemplateManager) validatePHPConfig(config map[string]interface{}) error {
	if gitRepo, ok := config["git_repo"].(string); ok {
		if !strings.HasPrefix(gitRepo, "http") && !strings.HasPrefix(gitRepo, "git@") {
			return fmt.Errorf("invalid git repository URL")
		}
	}

	if framework, ok := config["framework"].(string); ok {
		validFrameworks := []string{"none", "laravel", "codeigniter", "symfony", "cakephp", "zend", "slim"}
		if !contains(validFrameworks, framework) {
			return fmt.Errorf("unsupported PHP framework: %s", framework)
		}
	}

	return nil
}

// PrepareDeploymentConfig prepares the final deployment configuration
func (tm *TemplateManager) PrepareDeploymentConfig(templateID string, userConfig map[string]interface{}) (map[string]interface{}, error) {
	template, err := tm.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// Start with defaults
	config := make(map[string]interface{})
	for key, value := range template.Defaults {
		config[key] = value
	}

	// Override with user configuration
	for key, value := range userConfig {
		config[key] = value
	}

	// Validate the final configuration
	if err := tm.ValidateTemplate(templateID, config); err != nil {
		return nil, err
	}

	// Add template metadata
	config["_template_id"] = template.ID
	config["_template_type"] = string(template.Type)
	config["_template_framework"] = template.Framework

	log.Info().
		Str("template_id", templateID).
		Interface("config", config).
		Msg("Deployment configuration prepared")

	return config, nil
}

// GetAnsiblePlaybook returns the appropriate Ansible playbook for a template
func (tm *TemplateManager) GetAnsiblePlaybook(templateID string) (string, error) {
	template, err := tm.GetTemplate(templateID)
	if err != nil {
		return "", err
	}

	var playbookName string
	switch template.Type {
	case models.SiteTypeWordPress:
		playbookName = "deploy-wordpress.yml"
	case models.SiteTypeLaravel:
		playbookName = "deploy-laravel.yml"
	case models.SiteTypeNodeJS:
		switch template.Framework {
		case "nextjs":
			playbookName = "deploy-nextjs.yml"
		case "nuxtjs":
			playbookName = "deploy-nuxtjs.yml"
		default:
			playbookName = "deploy-nodejs.yml"
		}
	case models.SiteTypeStatic:
		switch template.Framework {
		case "hugo":
			playbookName = "deploy-hugo.yml"
		case "jekyll":
			playbookName = "deploy-jekyll.yml"
		case "gatsby":
			playbookName = "deploy-gatsby.yml"
		default:
			playbookName = "deploy-static.yml"
		}
	case models.SiteTypePHP:
		switch template.Framework {
		case "codeigniter":
			playbookName = "deploy-codeigniter.yml"
		case "symfony":
			playbookName = "deploy-symfony.yml"
		case "cakephp":
			playbookName = "deploy-cakephp.yml"
		default:
			playbookName = "deploy-php.yml"
		}
	default:
		return "", fmt.Errorf("no playbook available for template: %s", templateID)
	}

	return filepath.Join("automation", "playbooks", playbookName), nil
}

// Helper functions

// isValidPort checks if a port string is valid
func isValidPort(port string) bool {
	// Simple validation - could be enhanced
	return len(port) > 0 && len(port) <= 5
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
