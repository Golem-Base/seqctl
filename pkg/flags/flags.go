package flags

import "github.com/urfave/cli/v2"

// EnvVarPrefix is the prefix for all environment variables
const EnvVarPrefix = "SEQCTL"

// PrefixEnvVar adds the app prefix to the environment variable
func PrefixEnvVar(name string) string {
	return EnvVarPrefix + "_" + name
}

// Config flags
var (
	Config = &cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "Path to config file",
		EnvVars: []string{PrefixEnvVar("CONFIG")},
		Value:   "./config.toml",
	}
)

// Logging flags
var (
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
)

// Kubernetes flags
var (
	K8sConfig = &cli.StringFlag{
		Name:    "k8s-config",
		Usage:   "Path to kubeconfig file",
		Value:   "~/.kube/config",
		EnvVars: []string{PrefixEnvVar("K8S_CONFIG")},
	}
	K8sStatefulSetSelector = &cli.StringFlag{
		Name:    "k8s-statefulset-selector",
		Usage:   "Label selector for finding sequencer StatefulSets",
		Value:   "golem-base.io/optimism-role in (sequencer)",
		EnvVars: []string{PrefixEnvVar("K8S_STATEFULSET_SELECTOR")},
	}
	K8sServiceSelector = &cli.StringFlag{
		Name:    "k8s-service-selector",
		Usage:   "Label selector for finding sequencer Services (defaults to statefulset-selector)",
		Value:   "golem-base.io/optimism-role in (sequencer)",
		EnvVars: []string{PrefixEnvVar("K8S_SERVICE_SELECTOR")},
	}
	K8sNetworkLabel = &cli.StringFlag{
		Name:    "k8s-network-label",
		Usage:   "Label key for network identification",
		Value:   "golem-base.io/eth-network",
		EnvVars: []string{PrefixEnvVar("K8S_NETWORK_LABEL")},
	}
	K8sAppLabel = &cli.StringFlag{
		Name:    "k8s-app-label",
		Usage:   "Label key for app identification (used to match services)",
		Value:   "app",
		EnvVars: []string{PrefixEnvVar("K8S_APP_LABEL")},
	}
	K8sSequencerRoleLabel = &cli.StringFlag{
		Name:    "k8s-sequencer-role-label",
		Usage:   "Label key for sequencer role identification (voter/non-voter)",
		Value:   "golem-base.io/sequencer-role",
		EnvVars: []string{PrefixEnvVar("K8S_SEQUENCER_ROLE_LABEL")},
	}
	K8sSequencerVoterValues = &cli.StringSliceFlag{
		Name:    "k8s-sequencer-voter-values",
		Usage:   "Label values that indicate a voting member (default: voter)",
		Value:   cli.NewStringSlice("voter"),
		EnvVars: []string{PrefixEnvVar("K8S_SEQUENCER_VOTER_VALUES")},
	}
	K8sSequencerModeFilter = &cli.StringFlag{
		Name:    "k8s-sequencer-mode-filter",
		Usage:   "Filter resources by sequencer mode label (e.g. 'golem-base.io/sequencer-mode=ha'). Empty means no filtering",
		Value:   "",
		EnvVars: []string{PrefixEnvVar("K8S_SEQUENCER_MODE_FILTER")},
	}
	ConnectionMode = &cli.StringFlag{
		Name:    "k8s-connection-mode",
		Usage:   "Kubernetes connection mode: proxy, direct, or auto",
		Value:   "auto",
		EnvVars: []string{PrefixEnvVar("K8S_CONNECTION_MODE")},
	}
	Namespaces = &cli.StringSliceFlag{
		Name:    "k8s-namespaces",
		Usage:   "Kubernetes namespaces to scan for sequencers (empty means all namespaces)",
		EnvVars: []string{PrefixEnvVar("K8S_NAMESPACES")},
	}
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
)

// Server flags
var (
	ServerAddress = &cli.StringFlag{
		Name:    "server-address",
		Usage:   "Server listen address",
		Value:   "0.0.0.0",
		EnvVars: []string{PrefixEnvVar("SERVER_ADDRESS")},
	}
	ServerPort = &cli.IntFlag{
		Name:    "server-port",
		Usage:   "Server port",
		Value:   8080,
		EnvVars: []string{PrefixEnvVar("SERVER_PORT")},
	}
)

// Cache flags
var (
	CacheDiscoveryTTL = &cli.StringFlag{
		Name:    "cache-discovery-ttl",
		Usage:   "How long to cache network discovery (e.g. 5m, 30s)",
		Value:   "5m",
		EnvVars: []string{PrefixEnvVar("CACHE_DISCOVERY_TTL")},
	}
	CacheStatusTTL = &cli.StringFlag{
		Name:    "cache-status-ttl",
		Usage:   "How long before refreshing network status (e.g. 10s, 1m)",
		Value:   "10s",
		EnvVars: []string{PrefixEnvVar("CACHE_STATUS_TTL")},
	}
)

// ConfigFlags returns configuration-related flags
func ConfigFlags() []cli.Flag {
	return []cli.Flag{Config}
}

// LoggingFlags returns logging-related flags
func LoggingFlags() []cli.Flag {
	return []cli.Flag{LogLevel, LogFormat, LogNoColor, LogFile}
}

// K8sFlags returns all Kubernetes-related flags
func K8sFlags() []cli.Flag {
	return []cli.Flag{
		K8sConfig,
		K8sStatefulSetSelector,
		K8sServiceSelector,
		K8sNetworkLabel,
		K8sAppLabel,
		K8sSequencerRoleLabel,
		K8sSequencerVoterValues,
		K8sSequencerModeFilter,
		ConnectionMode,
		Namespaces,
		K8sConductorPort,
		K8sNodePort,
		K8sRaftPort,
		K8sConductorPortName,
		K8sNodePortName,
	}
}

// ServerFlags returns server specific flags
func ServerFlags() []cli.Flag {
	return []cli.Flag{ServerAddress, ServerPort}
}

// CacheFlags returns cache-related flags
func CacheFlags() []cli.Flag {
	return []cli.Flag{CacheDiscoveryTTL, CacheStatusTTL}
}

// ServeCommandFlags returns all flags needed for the serve command
func ServeCommandFlags() []cli.Flag {
	var flags []cli.Flag
	flags = append(flags, ConfigFlags()...)
	flags = append(flags, LoggingFlags()...)
	flags = append(flags, ServerFlags()...)
	flags = append(flags, K8sFlags()...)
	flags = append(flags, CacheFlags()...)
	return flags
}
