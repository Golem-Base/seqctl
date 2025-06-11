package flags

import (
	"time"

	"github.com/urfave/cli/v2"
)

// EnvVarPrefix is the prefix for all environment variables
const EnvVarPrefix = "SEQCTL"

// PrefixEnvVar adds the app prefix to the environment variable
func PrefixEnvVar(name string) string {
	return EnvVarPrefix + "_" + name
}

// CLI Flags
var (
	// Config flag
	Config = &cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "Path to config file",
		EnvVars: []string{PrefixEnvVar("CONFIG")},
		Value:   "./config.toml",
	}

	// Kubernetes flags
	K8sConfig = &cli.StringFlag{
		Name:    "k8s-config",
		Usage:   "Path to kubeconfig file",
		Value:   "~/.kube/config",
		EnvVars: []string{PrefixEnvVar("K8S_CONFIG")},
	}

	K8sSelector = &cli.StringFlag{
		Name:    "k8s-selector",
		Usage:   "Label selector for finding sequencer pods",
		Value:   "golem-base.io/optimism-role in (sequencer,sequencer-bootstrap)",
		EnvVars: []string{PrefixEnvVar("K8S_SELECTOR")},
	}

	// Logging flags
	LogLevel = &cli.StringFlag{
		Name:    "log-level",
		Usage:   "Log level (debug, info, warn, error)",
		Value:   "error",
		EnvVars: []string{PrefixEnvVar("LOG_LEVEL")},
	}

	LogFormat = &cli.StringFlag{
		Name:    "log-format",
		Usage:   "Log format (text, json)",
		Value:   "text",
		EnvVars: []string{PrefixEnvVar("LOG_FORMAT")},
	}

	LogNoColor = &cli.BoolFlag{
		Name:    "log-no-color",
		Usage:   "Disable colored output in logs",
		Value:   false,
		EnvVars: []string{PrefixEnvVar("LOG_NO_COLOR")},
	}

	LogFile = &cli.StringFlag{
		Name:    "log-file",
		Usage:   "Path to log file (logs to stderr if empty, required for TUI mode)",
		Value:   "",
		EnvVars: []string{PrefixEnvVar("LOG_FILE")},
	}

	// TUI flags
	RefreshInterval = &cli.DurationFlag{
		Name:    "refresh-interval",
		Usage:   "Auto-refresh interval for TUI updates",
		Value:   2 * time.Second,
		EnvVars: []string{PrefixEnvVar("REFRESH_INTERVAL")},
	}

	AutoRefresh = &cli.BoolFlag{
		Name:    "auto-refresh",
		Usage:   "Enable auto-refresh in TUI on startup",
		Value:   true,
		EnvVars: []string{PrefixEnvVar("AUTO_REFRESH")},
	}
)

// Required flags
var requiredFlags = []cli.Flag{}

// Optional flags
var optionalFlags = []cli.Flag{
	Config,
	K8sConfig,
	K8sSelector,
	LogLevel,
	LogFormat,
	LogNoColor,
	LogFile,
	RefreshInterval,
	AutoRefresh,
}

// Flags contains all CLI flags
var Flags []cli.Flag

func init() {
	Flags = append(requiredFlags, optionalFlags...)
}
