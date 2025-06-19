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
	k := koanf.New(delimiter)

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
			delimiter,
		)
	}

	// Load from environment variables
	slog.Debug("Loading config from environment variables")
	if err := k.Load(env.Provider(EnvPrefix, delimiter, envTransform), nil); err != nil {
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
		flags.K8sConfig.Name:            "k8s.config_path",
		flags.K8sSelector.Name:          "k8s.selector",
		flags.ConnectionMode.Name:       "k8s.connection_mode",
		flags.K8sNetworkLabel.Name:      "k8s.network_label",
		flags.K8sRoleLabel.Name:         "k8s.role_label",
		flags.K8sAppLabel.Name:          "k8s.app_label",
		flags.K8sConductorPort.Name:     "k8s.conductor_port",
		flags.K8sNodePort.Name:          "k8s.node_port",
		flags.K8sRaftPort.Name:          "k8s.raft_port",
		flags.K8sConductorPortName.Name: "k8s.conductor_port_name",
		flags.K8sNodePortName.Name:      "k8s.node_port_name",
		flags.K8sSequencerRole.Name:     "k8s.sequencer_role",
		flags.K8sBootstrapRole.Name:     "k8s.bootstrap_role",
		flags.LogLevel.Name:             "log.level",
		flags.LogFormat.Name:            "log.format",
		flags.LogNoColor.Name:           "log.no_color",
		flags.LogFile.Name:              "log.file_path",
		flags.WebAddress.Name:           "web.address",
		flags.WebPort.Name:              "web.port",
	}

	// Process each flag
	for flagName, koanfKey := range flagMap {
		if cliCtx.IsSet(flagName) {
			flagsAdded = true
			if flagName == flags.LogNoColor.Name {
				fs.Bool(koanfKey, cliCtx.Bool(flagName), "")
				fs.Set(koanfKey, strings.ToLower(strings.TrimSpace(cliCtx.String(flagName))))
				slog.Debug("Added CLI flag", "name", flagName, "koanf_key", koanfKey, "value", cliCtx.Bool(flagName))
			} else if flagName == flags.WebPort.Name ||
				flagName == flags.K8sConductorPort.Name ||
				flagName == flags.K8sNodePort.Name ||
				flagName == flags.K8sRaftPort.Name {
				fs.Int(koanfKey, cliCtx.Int(flagName), "")
				fs.Set(koanfKey, fmt.Sprintf("%d", cliCtx.Int(flagName)))
				slog.Debug("Added CLI flag", "name", flagName, "koanf_key", koanfKey, "value", cliCtx.Int(flagName))
			} else {
				fs.String(koanfKey, cliCtx.String(flagName), "")
				fs.Set(koanfKey, cliCtx.String(flagName))
				slog.Debug("Added CLI flag", "name", flagName, "koanf_key", koanfKey, "value", cliCtx.String(flagName))
			}
		}
	}

	// Handle namespaces separately since it's a StringSlice
	if cliCtx.IsSet(flags.Namespaces.Name) {
		namespaces := cliCtx.StringSlice(flags.Namespaces.Name)
		if err := k.Set("k8s.namespaces", namespaces); err != nil {
			slog.Error("Failed to set namespaces", "error", err)
		}
		slog.Debug("Added CLI flag", "name", flags.Namespaces.Name, "value", namespaces)
	}

	// Only load flags if any were set
	if flagsAdded {
		slog.Debug("Loading config from CLI flags")
		if err := k.Load(posflag.Provider(fs, delimiter, k), nil); err != nil {
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
		"k8s.config_path", cfg.K8s.ConfigPath,
		"k8s.selector", cfg.K8s.Selector,
		"k8s.connection_mode", cfg.K8s.ConnectionMode,
		"k8s.namespaces", cfg.K8s.Namespaces,
		"log.level", cfg.Log.Level,
		"web.address", cfg.Web.Address,
		"web.port", cfg.Web.Port)

	return cfg, nil
}
