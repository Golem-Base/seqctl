package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/golem-base/seqctl/pkg/flags"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
	"github.com/urfave/cli/v2"
)

// LoadConfig loads configuration from various sources
func LoadConfig(cliCtx *cli.Context) (*Config, error) {
	// Create a new instance with defaults
	cfg := New()

	// Initialize koanf
	k := koanf.New(Delimiter)

	// Load defaults from struct
	if err := k.Load(structs.Provider(cfg, KoanfTag), nil); err != nil {
		slog.Error("Failed to load defaults", "error", err)
		return nil, fmt.Errorf("failed to load default configuration values: %w", err)
	}

	// Define environment variable transformer
	envTransform := func(s string) string {
		return strings.ReplaceAll(
			strings.ToLower(
				strings.TrimPrefix(s, EnvPrefix),
			),
			EnvSeparator,
			Delimiter,
		)
	}

	// Load from environment variables
	slog.Debug("Loading config from environment variables")
	if err := k.Load(env.Provider(EnvPrefix, Delimiter, envTransform), nil); err != nil {
		slog.Error("Failed to load environment variables", "error", err)
		return nil, fmt.Errorf("failed to load configuration from environment variables with prefix %s: %w", EnvPrefix, err)
	}

	// Load from config file if provided
	configPath := cliCtx.String(flags.Config.Name)
	if configPath != "" {
		slog.Debug("Loading config from file", "path", configPath)
		if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				slog.Error("Failed to load config file", "path", configPath, "error", err)
				return nil, fmt.Errorf("failed to load configuration from file %s: %w", configPath, err)
			}
			slog.Debug("Config file not found, skipping", "path", configPath)
		}
	}

	// Create flag set for command line args
	fs := pflag.NewFlagSet("config", pflag.ContinueOnError)

	// Add flags from CLI context
	flagsAdded := false

	// Map of flag names to their koanf keys
	flagMap := map[string]string{
		flags.K8sConfig.Name:   "k8s_config",
		flags.K8sSelector.Name: "k8s_selector",
		flags.LogLevel.Name:    "log_level",
		flags.LogFormat.Name:   "log_format",
		flags.LogNoColor.Name:  "log_no_color",
		flags.LogFile.Name:     "log_file",
	}

	// Process each flag
	for flagName, koanfKey := range flagMap {
		if cliCtx.IsSet(flagName) {
			flagsAdded = true
			if flagName == flags.LogNoColor.Name {
				fs.Bool(koanfKey, cliCtx.Bool(flagName), "")
				fs.Set(koanfKey, strings.ToLower(strings.TrimSpace(cliCtx.String(flagName))))
				slog.Debug("Added CLI flag", "name", flagName, "koanf_key", koanfKey, "value", cliCtx.Bool(flagName))
			} else {
				fs.String(koanfKey, cliCtx.String(flagName), "")
				fs.Set(koanfKey, cliCtx.String(flagName))
				slog.Debug("Added CLI flag", "name", flagName, "koanf_key", koanfKey, "value", cliCtx.String(flagName))
			}
		}
	}

	// Only load flags if any were set
	if flagsAdded {
		slog.Debug("Loading config from CLI flags")
		if err := k.Load(posflag.Provider(fs, Delimiter, k), nil); err != nil {
			slog.Error("Failed to load CLI flags", "error", err)
			return nil, fmt.Errorf("failed to load configuration from CLI flags: %w", err)
		}
	}

	// Unmarshal into the config struct
	if err := k.Unmarshal("", cfg); err != nil {
		slog.Error("Failed to unmarshal config", "error", err)
		return nil, fmt.Errorf("failed to unmarshal configuration into struct: %w", err)
	}

	// Log the final configuration
	slog.Debug("Configuration loaded",
		"k8s_config", cfg.K8sConfig,
		"k8s_selector", cfg.K8sSelector,
		"log_level", cfg.LogLevel)

	return cfg, nil
}
