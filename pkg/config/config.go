package config

import (
	"github.com/golem-base/seqctl/pkg/flags"
)

const (
	delimiter    = "."
	KoanfTag     = "koanf"
	EnvSeparator = "_"
	EnvPrefix    = flags.EnvVarPrefix + EnvSeparator
)

// K8sConfig holds Kubernetes-related configuration
type K8sConfig struct {
	ConfigPath     string   `koanf:"config_path" json:"config_path" yaml:"config_path" toml:"config_path"`
	Selector       string   `koanf:"selector" json:"selector" yaml:"selector" toml:"selector"`
	ConnectionMode string   `koanf:"connection_mode" json:"connection_mode" yaml:"connection_mode" toml:"connection_mode"`
	Namespaces     []string `koanf:"namespaces" json:"namespaces" yaml:"namespaces" toml:"namespaces"`
	NetworkLabel   string   `koanf:"network_label" json:"network_label" yaml:"network_label" toml:"network_label"`
	RoleLabel      string   `koanf:"role_label" json:"role_label" yaml:"role_label" toml:"role_label"`
	AppLabel       string   `koanf:"app_label" json:"app_label" yaml:"app_label" toml:"app_label"`

	// Port configuration
	ConductorPort     int    `koanf:"conductor_port" json:"conductor_port" yaml:"conductor_port" toml:"conductor_port"`
	NodePort          int    `koanf:"node_port" json:"node_port" yaml:"node_port" toml:"node_port"`
	RaftPort          int    `koanf:"raft_port" json:"raft_port" yaml:"raft_port" toml:"raft_port"`
	ConductorPortName string `koanf:"conductor_port_name" json:"conductor_port_name" yaml:"conductor_port_name" toml:"conductor_port_name"`
	NodePortName      string `koanf:"node_port_name" json:"node_port_name" yaml:"node_port_name" toml:"node_port_name"`

	// Role identifiers
	SequencerRole string `koanf:"sequencer_role" json:"sequencer_role" yaml:"sequencer_role" toml:"sequencer_role"`
	BootstrapRole string `koanf:"bootstrap_role" json:"bootstrap_role" yaml:"bootstrap_role" toml:"bootstrap_role"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level    string `koanf:"level" json:"level" yaml:"level" toml:"level"`
	Format   string `koanf:"format" json:"format" yaml:"format" toml:"format"`
	NoColor  bool   `koanf:"no_color" json:"no_color" yaml:"no_color" toml:"no_color"`
	FilePath string `koanf:"file_path" json:"file_path" yaml:"file_path" toml:"file_path"`
}

// WebConfig holds web server configuration
type WebConfig struct {
	Address         string `koanf:"address" json:"address" yaml:"address" toml:"address"`
	Port            int    `koanf:"port" json:"port" yaml:"port" toml:"port"`
	RefreshInterval int    `koanf:"refresh_interval" json:"refresh_interval" yaml:"refresh_interval" toml:"refresh_interval"`
}

// Config holds the application configuration
type Config struct {
	K8s K8sConfig `koanf:"k8s"`
	Log LogConfig `koanf:"log"`
	Web WebConfig `koanf:"web"`
}

// New creates a new Config instance with default values
func New() *Config {
	return &Config{
		K8s: K8sConfig{
			ConfigPath:     flags.K8sConfig.Value,
			Selector:       flags.K8sSelector.Value,
			ConnectionMode: "auto",
			Namespaces:     []string{},
			NetworkLabel:   flags.K8sNetworkLabel.Value,
			RoleLabel:      flags.K8sRoleLabel.Value,
			AppLabel:       flags.K8sAppLabel.Value,
			// Port defaults
			ConductorPort:     flags.K8sConductorPort.Value,
			NodePort:          flags.K8sNodePort.Value,
			RaftPort:          flags.K8sRaftPort.Value,
			ConductorPortName: flags.K8sConductorPortName.Value,
			NodePortName:      flags.K8sNodePortName.Value,
			// Role defaults
			SequencerRole: flags.K8sSequencerRole.Value,
			BootstrapRole: flags.K8sBootstrapRole.Value,
		},
		Log: LogConfig{
			Level:    flags.LogLevel.Value,
			Format:   flags.LogFormat.Value,
			NoColor:  flags.LogNoColor.Value,
			FilePath: flags.LogFile.Value,
		},
		Web: WebConfig{
			Address:         flags.WebAddress.Value,
			Port:            flags.WebPort.Value,
			RefreshInterval: flags.WebRefreshInterval.Value,
		},
	}
}
