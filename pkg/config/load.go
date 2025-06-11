package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/golem-base/seqctl/pkg/flags"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/urfave/cli/v2"
)

// LoadConfig loads configuration from various sources in order of precedence:
// 1. Default values from struct
// 2. Environment variables (prefixed with SEQCTL_)
// 3. Configuration file (if provided)
// 4. Command-line flags (highest priority)
func LoadConfig(cliCtx *cli.Context) (*Config, error) {
	cfg := New()
	k := koanf.New(delimiter)

	// Load defaults from struct
	if err := k.Load(structs.Provider(cfg, koanfTag), nil); err != nil {
		return nil, fmt.Errorf("load defaults: %w", err)
	}

	// Load environment variables
	if err := loadEnvVars(k); err != nil {
		return nil, fmt.Errorf("load env vars: %w", err)
	}

	// Load config file if provided
	if configPath := cliCtx.String(flags.Config.Name); configPath != "" {
		if err := loadConfigFile(k, configPath); err != nil {
			return nil, err
		}
	}

	// Load CLI flags
	if err := loadCLIFlags(k, cliCtx); err != nil {
		return nil, fmt.Errorf("load CLI flags: %w", err)
	}

	// Unmarshal into the config struct
	if err := k.Unmarshal("", cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Expand paths after loading
	cfg.K8s.ConfigPath = expandPath(cfg.K8s.ConfigPath)
	cfg.Log.FilePath = expandPath(cfg.Log.FilePath)

	logFinalConfig(cfg)
	return cfg, nil
}

// loadEnvVars loads configuration from environment variables
func loadEnvVars(k *koanf.Koanf) error {
	envTransform := func(s string) string {
		return strings.ReplaceAll(
			strings.ToLower(strings.TrimPrefix(s, envPrefix)),
			envSeparator,
			delimiter,
		)
	}

	slog.Debug("Loading config from environment variables")
	return k.Load(env.Provider(envPrefix, delimiter, envTransform), nil)
}

// loadConfigFile loads configuration from a TOML file
func loadConfigFile(k *koanf.Koanf, path string) error {
	slog.Debug("Loading config from file", "path", path)
	if err := k.Load(file.Provider(path), toml.Parser()); err != nil {
		if strings.Contains(err.Error(), "no such file") {
			slog.Debug("Config file not found, skipping", "path", path)
			return nil
		}
		return fmt.Errorf("load config file %s: %w", path, err)
	}
	return nil
}

// flagMapping defines the mapping from CLI flags to koanf paths
var flagMapping = map[string]string{
	"k8s-config":                 "k8s.config_path",
	"k8s-statefulset-selector":   "k8s.statefulset_selector",
	"k8s-service-selector":       "k8s.service_selector",
	"k8s-connection-mode":        "k8s.connection_mode",
	"k8s-network-label":          "k8s.network_label",
	"k8s-app-label":              "k8s.app_label",
	"k8s-sequencer-role-label":   "k8s.sequencer_role_label",
	"k8s-sequencer-voter-values": "k8s.sequencer_voter_values",
	"k8s-sequencer-mode-filter":  "k8s.sequencer_mode_filter",
	"k8s-conductor-port":         "k8s.conductor_port",
	"k8s-node-port":              "k8s.node_port",
	"k8s-raft-port":              "k8s.raft_port",
	"k8s-conductor-port-name":    "k8s.conductor_port_name",
	"k8s-node-port-name":         "k8s.node_port_name",
	"log-level":                  "log.level",
	"log-format":                 "log.format",
	"log-no-color":               "log.no_color",
	"log-file":                   "log.file_path",
	"server-address":             "server.address",
	"server-port":                "server.port",
	"k8s-namespaces":             "k8s.namespaces",
	"cache-discovery-ttl":        "cache.discovery_ttl",
	"cache-status-ttl":           "cache.status_ttl",
}

// loadCLIFlags loads configuration from command-line flags
func loadCLIFlags(k *koanf.Koanf, cliCtx *cli.Context) error {
	for flagName, koanfPath := range flagMapping {
		if !cliCtx.IsSet(flagName) {
			continue
		}

		var value any
		switch flagName {
		case "log-no-color":
			value = cliCtx.Bool(flagName)
		case "server-port", "k8s-conductor-port", "k8s-node-port", "k8s-raft-port":
			value = cliCtx.Int(flagName)
		case "namespaces", "k8s-sequencer-voter-values":
			value = cliCtx.StringSlice(flagName)
		default:
			value = cliCtx.String(flagName)
		}

		if err := k.Set(koanfPath, value); err != nil {
			return fmt.Errorf("set flag %s: %w", flagName, err)
		}

		slog.Debug("Set CLI flag", "flag", flagName, "path", koanfPath, "value", value)
	}

	return nil
}

// logFinalConfig logs the final configuration for debugging
func logFinalConfig(cfg *Config) {
	slog.Debug("Configuration loaded",
		"k8s.config_path", cfg.K8s.ConfigPath,
		"k8s.statefulset_selector", cfg.K8s.StatefulSetSelector,
		"k8s.service_selector", cfg.K8s.ServiceSelector,
		"k8s.connection_mode", cfg.K8s.ConnectionMode,
		"k8s.namespaces", cfg.K8s.Namespaces,
		"log.level", cfg.Log.Level,
		"server.address", cfg.Server.Address,
		"server.port", cfg.Server.Port,
		"cache.discovery_ttl", cfg.Cache.DiscoveryTTL,
		"cache.status_ttl", cfg.Cache.StatusTTL)
}
