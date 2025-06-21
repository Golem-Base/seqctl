# seqctl

Web-based control panel for managing Optimism conductor sequencer clusters in Kubernetes environments.

## Features

- **Real-time Monitoring** - View sequencer health, status, and leadership in real-time
- **Cluster Operations** - Pause/resume conductors, transfer leadership, manage Raft membership
- **Multi-Provider Support** - Kubernetes today, extensible to Docker and AWS
- **RESTful API** - Complete API for automation and integration
- **WebSocket Updates** - Real-time dashboard updates without polling
- **Clean Architecture** - Layered design for maintainability and testing

## Screenshots

<p align="center">
  <img src="docs/assets/dashboard.png" alt="Dashboard" width="600">
</p>

## Quick Start

### Using Pre-built Binary

```bash
# Download the latest release
curl -L https://github.com/golem-base/seqctl/releases/latest/download/seqctl_linux_amd64.tar.gz | tar xz

# Run the web server
./seqctl web --k8s-selector "app=op-conductor"
```

### Using Docker

```bash
docker run -p 8080:8080 \
  -v ~/.kube/config:/app/.kube/config:ro \
  golemnetwork/seqctl:latest \
  web --k8s-config /app/.kube/config
```

### From Source

```bash
# Clone the repository
git clone https://github.com/golem-base/seqctl
cd seqctl

# Build with Just
just build

# Or with Go directly
go build -o bin/seqctl ./cmd/seqctl

# Run the web server
./bin/seqctl web
```

Access the dashboard at `http://localhost:8080`

## Configuration

Configure via CLI flags, environment variables, or config file:

```bash
# CLI flags
seqctl web --port 8080 --k8s-selector "app=op-conductor"

# Environment variables
export SEQCTL_WEB_PORT=8080
export SEQCTL_K8S_SELECTOR="app=op-conductor"
seqctl web

# Config file
seqctl web --config ./config.toml
```

See [Configuration Guide](https://golem-base.github.io/seqctl/configuration/) for all options.

## Use Cases

### Monitor Sequencer Health

View real-time status of all sequencers in your network, including sync status, leadership, and Raft cluster health.

### Perform Maintenance

Safely pause conductors for maintenance, transfer leadership before updates, and resume operations without downtime.

### Manage Failures

Override leader status during split-brain scenarios, force sequencers active, or halt problematic nodes.

### Automate Operations

Use the REST API to integrate with your existing automation, monitoring, and alerting systems.

## Contributing

We welcome contributions! Please see our [Contributing Guide](https://golem-base.github.io/seqctl/developer-guide/contributing/) for details.

```bash
# Setup development environment
just dev

# Run tests
just test

# Build and run locally
just run
```

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Built with love by the [Golem Base](https://github.com/golem-base) team.

Special thanks to the [Optimism](https://optimism.io) community for the conductor sequencer architecture.
