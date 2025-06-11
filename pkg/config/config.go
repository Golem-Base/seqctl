package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/golem-base/seqctl/pkg/flags"
)

const (
	delimiter    = "."
	envSeparator = "_"
	koanfTag     = "koanf"
	envPrefix    = flags.EnvVarPrefix + envSeparator
)

// expandPath expands ~ to the user's home directory
func expandPath(path string) string {
	if path == "" || !strings.HasPrefix(path, "~/") {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If we can't get home dir, return path as-is
		return path
	}

	// Use filepath.Join to properly handle path separators
	return filepath.Join(homeDir, path[2:])
}

// K8sConfig holds Kubernetes-related configuration
type K8sConfig struct {
	AppLabel             string   `koanf:"app_label" toml:"app_label"`
	ConductorPort        int      `koanf:"conductor_port" toml:"conductor_port"`
	ConductorPortName    string   `koanf:"conductor_port_name" toml:"conductor_port_name"`
	ConfigPath           string   `koanf:"config_path" toml:"config_path"`
	ConnectionMode       string   `koanf:"connection_mode" toml:"connection_mode"`
	Namespaces           []string `koanf:"namespaces" toml:"namespaces"`
	NetworkLabel         string   `koanf:"network_label" toml:"network_label"`
	NodePort             int      `koanf:"node_port" toml:"node_port"`
	NodePortName         string   `koanf:"node_port_name" toml:"node_port_name"`
	RaftPort             int      `koanf:"raft_port" toml:"raft_port"`
	SequencerModeFilter  string   `koanf:"sequencer_mode_filter" toml:"sequencer_mode_filter"`
	SequencerRoleLabel   string   `koanf:"sequencer_role_label" toml:"sequencer_role_label"`
	SequencerVoterValues []string `koanf:"sequencer_voter_values" toml:"sequencer_voter_values"`
	ServiceSelector      string   `koanf:"service_selector" toml:"service_selector"`
	StatefulSetSelector  string   `koanf:"statefulset_selector" toml:"statefulset_selector"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level    string `koanf:"level" toml:"level"`
	Format   string `koanf:"format" toml:"format"`
	NoColor  bool   `koanf:"no_color" toml:"no_color"`
	FilePath string `koanf:"file_path" toml:"file_path"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Address string `koanf:"address" toml:"address"`
	Port    int    `koanf:"port" toml:"port"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	DiscoveryTTL string `koanf:"discovery_ttl" toml:"discovery_ttl"`
	StatusTTL    string `koanf:"status_ttl" toml:"status_ttl"`
}

// Config holds the application configuration
type Config struct {
	K8s    K8sConfig    `koanf:"k8s"`
	Log    LogConfig    `koanf:"log"`
	Server ServerConfig `koanf:"server"`
	Cache  CacheConfig  `koanf:"cache"`
}

// New creates a new Config instance with default values
func New() *Config {
	return &Config{
		K8s: K8sConfig{
			AppLabel:             flags.K8sAppLabel.Value,
			ConductorPort:        flags.K8sConductorPort.Value,
			ConductorPortName:    flags.K8sConductorPortName.Value,
			ConfigPath:           expandPath(flags.K8sConfig.Value),
			ConnectionMode:       flags.ConnectionMode.Value,
			Namespaces:           []string{},
			NetworkLabel:         flags.K8sNetworkLabel.Value,
			NodePort:             flags.K8sNodePort.Value,
			NodePortName:         flags.K8sNodePortName.Value,
			RaftPort:             flags.K8sRaftPort.Value,
			SequencerModeFilter:  flags.K8sSequencerModeFilter.Value,
			SequencerRoleLabel:   flags.K8sSequencerRoleLabel.Value,
			SequencerVoterValues: []string{"voter"}, // Default: only "voter" indicates voting member
			ServiceSelector:      flags.K8sServiceSelector.Value,
			StatefulSetSelector:  flags.K8sStatefulSetSelector.Value,
		},
		Log: LogConfig{
			FilePath: flags.LogFile.Value,
			Format:   flags.LogFormat.Value,
			Level:    flags.LogLevel.Value,
			NoColor:  flags.LogNoColor.Value,
		},
		Server: ServerConfig{
			Address: flags.ServerAddress.Value,
			Port:    flags.ServerPort.Value,
		},
		Cache: CacheConfig{
			DiscoveryTTL: "5m",
			StatusTTL:    "10s",
		},
	}
}
