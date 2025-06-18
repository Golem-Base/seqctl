package config

import (
	"fmt"

	"github.com/golem-base/seqctl/pkg/flags"
	"github.com/golem-base/seqctl/pkg/ui/tui/styles"
)

const (
	Delimiter    = "."
	KoanfTag     = "koanf"
	EnvSeparator = "_"
	EnvPrefix    = flags.EnvVarPrefix + EnvSeparator
)

// ThemeName represents available themes
type ThemeName string

const (
	ThemeDefault         ThemeName = "default"
	ThemeCatppuccinMocha ThemeName = "catppuccin-mocha"
)

// IconStyle represents available icon styles
type IconStyle string

const (
	IconStyleDefault IconStyle = "default"
)

// UIConfig holds TUI configuration options
type UIConfig struct {
	Theme     ThemeName `koanf:"theme" json:"theme" yaml:"theme" toml:"theme"`
	IconStyle IconStyle `koanf:"icon_style" json:"icon_style" yaml:"icon_style" toml:"icon_style"`
}

// Config holds the application configuration
type Config struct {
	K8sConfig   string   `koanf:"k8s_config"`
	K8sSelector string   `koanf:"k8s_selector"`
	LogLevel    string   `koanf:"log_level"`
	LogFormat   string   `koanf:"log_format"`
	LogNoColor  bool     `koanf:"log_no_color"`
	LogFile     string   `koanf:"log_file"`
	UI          UIConfig `koanf:"ui"`
}

// New creates a new Config instance with default values
func New() *Config {
	return &Config{
		K8sSelector: flags.K8sSelector.Value,
		LogLevel:    flags.LogLevel.Value,
		LogFormat:   flags.LogFormat.Value,
		LogNoColor:  flags.LogNoColor.Value,
		UI: UIConfig{
			Theme:     ThemeDefault,
			IconStyle: IconStyleDefault,
		},
	}
}

// GetTheme returns the theme instance based on the configured theme name
func (ui *UIConfig) GetTheme() (*styles.Theme, error) {
	switch ui.Theme {
	case ThemeDefault:
		return styles.Default(), nil
	case ThemeCatppuccinMocha:
		return styles.CatppuccinMocha(), nil
	default:
		return nil, fmt.Errorf("unknown theme: %s", ui.Theme)
	}
}

// GetIcons returns the icon set based on the configured icon style
func (ui *UIConfig) GetIcons() (*styles.Icons, error) {
	switch ui.IconStyle {
	case IconStyleDefault:
		return styles.DefaultIcons(), nil
	default:
		return nil, fmt.Errorf("unknown icon style: %s", ui.IconStyle)
	}
}

// Validate checks if the UI configuration values are valid
func (ui *UIConfig) Validate() error {
	if _, err := ui.GetTheme(); err != nil {
		return fmt.Errorf("invalid theme: %w", err)
	}

	if _, err := ui.GetIcons(); err != nil {
		return fmt.Errorf("invalid icon style: %w", err)
	}

	return nil
}
