# seqctl

A CLI tool for managing Optimism conductor sequencer clusters in Kubernetes environments.

> [!IMPORTANT]
> `seqctl` is alpha software and may cause service disruption on your sequencer cluster if used improperly!

## Features

- **Interactive TUI**: Real-time monitoring and control of sequencer networks
- **Network Discovery**: Automatic discovery of sequencer networks from Kubernetes
- **Sequencer Management**: Pause, resume, transfer leadership, and control sequencer operations
- **Cluster Operations**: Manage Raft cluster membership and voting rights
- **Health Monitoring**: Track conductor and sequencer health status

## Installation

### Prerequisites

- Go 1.22 or higher
- Access to a Kubernetes cluster with Optimism sequencers deployed
- `kubectl` configured with appropriate permissions

### Building from Source

```bash
# Clone the repository
git clone https://github.com/golem-base/seqctl.git
cd seqctl

# Build the binary
just build

# Or using go directly
go build -o bin/seqctl ./cmd
```

### Using Nix

If you have Nix installed:

```bash
# Enter development shell
nix develop

# Build the project
just build
```

## Usage

### Interactive TUI (Default)

The TUI provides a real-time view of your sequencer networks:

```bash
# Launch TUI for a specific network
seqctl <network-name>

# Launch TUI with specific kubeconfig
seqctl --k8s-config ~/.kube/config <network-name>
```

## Configuration

Configuration can be provided through multiple sources (in order of precedence):

1. Command-line flags
2. Environment variables (prefixed with `SEQCTL_`)
3. Configuration file

### Configuration Options

```toml
# config.toml example
[k8s]
config = "/path/to/kubeconfig"  # Path to kubeconfig (uses in-cluster config if empty)
namespace = ""                  # Namespace to search (empty = all namespaces)
selector = "app=op-conductor"   # Label selector for sequencer pods

[log]
level = "info"                  # Log level: debug, info, warn, error
format = "text"                 # Log format: text or json
```

### Environment Variables

```bash
export SEQCTL_K8S_CONFIG="/path/to/kubeconfig"
export SEQCTL_K8S_SELECTOR="app=op-conductor"
export SEQCTL_LOG_LEVEL="debug"
export SEQCTL_REFRESH_INTERVAL="5s"
export SEQCTL_AUTO_REFRESH="true"
```

## Kubernetes Setup

### Required Permissions

The tool requires the following Kubernetes permissions:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: seqctl
rules:
  - apiGroups: ["apps"]
    resources: ["statefulsets"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["services", "pods", "pods/log"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["pods/proxy"]
    verbs: ["get", "create"]
```

### Expected Kubernetes Resources

The tool discovers sequencers from StatefulSets with the following structure:

- StatefulSets labeled with the configured selector
- Services matching the StatefulSet names
- Container named `op-conductor` with conductor RPC on port 8545
- Container named `op-node` with node RPC on port 8547

## Development

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Troubleshooting

### Common Issues

**Cannot connect to Kubernetes cluster**

- Ensure your kubeconfig is valid: `kubectl cluster-info`
- Check if you have the required permissions: `kubectl auth can-i list statefulsets`

**No networks discovered**

- Verify the label selector matches your StatefulSets: `kubectl get statefulsets -l app=op-conductor`
- Check if sequencers are in the expected namespace

**RPC connection failures**

- Ensure the Kubernetes API proxy is enabled
- Check if the pods are running: `kubectl get pods -l app=op-conductor`
- Verify the container ports are correctly configured

## License

See [LICENSE](./LICENSE) for more information.
