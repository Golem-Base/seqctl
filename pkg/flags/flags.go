package flags

import (
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

	K8sNetworkLabel = &cli.StringFlag{
		Name:    "k8s-network-label",
		Usage:   "Label key for network identification",
		Value:   "golem-base.io/eth-network",
		EnvVars: []string{PrefixEnvVar("K8S_NETWORK_LABEL")},
	}

	K8sRoleLabel = &cli.StringFlag{
		Name:    "k8s-role-label",
		Usage:   "Label key for sequencer role identification",
		Value:   "golem-base.io/optimism-role",
		EnvVars: []string{PrefixEnvVar("K8S_ROLE_LABEL")},
	}

	K8sAppLabel = &cli.StringFlag{
		Name:    "k8s-app-label",
		Usage:   "Label key for app identification (used to match services)",
		Value:   "app",
		EnvVars: []string{PrefixEnvVar("K8S_APP_LABEL")},
	}

	// Logging flags
	LogLevel = &cli.StringFlag{
		Name:    "log-level",
		Usage:   "Log level (debug, info, warn, error)",
		Value:   "info",
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
		Usage:   "Path to log file (logs to stderr if empty)",
		Value:   "",
		EnvVars: []string{PrefixEnvVar("LOG_FILE")},
	}

	// Web server flags
	WebAddress = &cli.StringFlag{
		Name:    "address",
		Usage:   "Web server listen address",
		Value:   "0.0.0.0",
		EnvVars: []string{PrefixEnvVar("WEB_ADDRESS")},
	}

	WebPort = &cli.IntFlag{
		Name:    "port",
		Usage:   "Web server port",
		Value:   8080,
		EnvVars: []string{PrefixEnvVar("WEB_PORT")},
	}

	WebRefreshInterval = &cli.IntFlag{
		Name:    "refresh-interval",
		Usage:   "Auto-refresh interval in seconds (minimum 1)",
		Value:   5,
		EnvVars: []string{PrefixEnvVar("WEB_REFRESH_INTERVAL")},
	}

	// Connection mode flag
	ConnectionMode = &cli.StringFlag{
		Name:    "connection-mode",
		Usage:   "Kubernetes connection mode: proxy, direct, pod-scan, or auto",
		Value:   "auto",
		EnvVars: []string{PrefixEnvVar("CONNECTION_MODE")},
	}

	// Namespaces flag for scanning specific namespaces
	Namespaces = &cli.StringSliceFlag{
		Name:    "namespaces",
		Usage:   "Kubernetes namespaces to scan for sequencers (empty means all namespaces)",
		EnvVars: []string{PrefixEnvVar("NAMESPACES")},
	}

	// Port configuration flags
	K8sConductorPort = &cli.IntFlag{
		Name:    "k8s-conductor-port",
		Usage:   "Default conductor RPC port",
		Value:   8555,
		EnvVars: []string{PrefixEnvVar("K8S_CONDUCTOR_PORT")},
	}

	K8sNodePort = &cli.IntFlag{
		Name:    "k8s-node-port",
		Usage:   "Default node RPC port",
		Value:   9545,
		EnvVars: []string{PrefixEnvVar("K8S_NODE_PORT")},
	}

	K8sRaftPort = &cli.IntFlag{
		Name:    "k8s-raft-port",
		Usage:   "Default raft consensus port",
		Value:   50050,
		EnvVars: []string{PrefixEnvVar("K8S_RAFT_PORT")},
	}

	K8sConductorPortName = &cli.StringFlag{
		Name:    "k8s-conductor-port-name",
		Usage:   "Service port name for conductor RPC",
		Value:   "cndctr-rpc",
		EnvVars: []string{PrefixEnvVar("K8S_CONDUCTOR_PORT_NAME")},
	}

	K8sNodePortName = &cli.StringFlag{
		Name:    "k8s-node-port-name",
		Usage:   "Service port name for node RPC",
		Value:   "op-node-rpc",
		EnvVars: []string{PrefixEnvVar("K8S_NODE_PORT_NAME")},
	}

	// Role identifier flags
	K8sSequencerRole = &cli.StringFlag{
		Name:    "k8s-sequencer-role",
		Usage:   "Role identifier for sequencers",
		Value:   "sequencer",
		EnvVars: []string{PrefixEnvVar("K8S_SEQUENCER_ROLE")},
	}

	K8sBootstrapRole = &cli.StringFlag{
		Name:    "k8s-bootstrap-role",
		Usage:   "Role identifier for bootstrap nodes",
		Value:   "bootstrap",
		EnvVars: []string{PrefixEnvVar("K8S_BOOTSTRAP_ROLE")},
	}
)

// CommonFlags are flags shared by all commands
var CommonFlags = []cli.Flag{
	Config,
	K8sConfig,
	K8sSelector,
	K8sNetworkLabel,
	K8sRoleLabel,
	K8sAppLabel,
	K8sConductorPort,
	K8sNodePort,
	K8sRaftPort,
	K8sConductorPortName,
	K8sNodePortName,
	K8sSequencerRole,
	K8sBootstrapRole,
	LogLevel,
	LogFormat,
	LogNoColor,
	LogFile,
}

// WebFlags are flags specific to the Web command
var WebFlags = []cli.Flag{
	WebAddress,
	WebPort,
	WebRefreshInterval,
	ConnectionMode,
	Namespaces,
}

// Flags contains all CLI flags (for backward compatibility)
var Flags []cli.Flag

func init() {
	Flags = append(CommonFlags, WebFlags...)
}
