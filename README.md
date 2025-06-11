# Seqctl

A modern web-based control panel for managing Optimism conductor sequencer clusters in Kubernetes environments. Built with React and Go, seqctl provides a single-binary deployment with an embedded SPA frontend.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.22-blue.svg)
![React](https://img.shields.io/badge/React-v19-61DAFB.svg)
![Bun](https://img.shields.io/badge/Bun-latest-F9F1E1.svg)
![Tailwind CSS](https://img.shields.io/badge/Tailwind%20CSS-v4-38B2AC.svg)

## Features

`seqctl` provides a modern web interface to manage and monitor sequencer clusters deployed in Kubernetes:

### Core Capabilities

- **Real-time Monitoring**: View sequencer status with auto-refresh
- **Conductor Operations**: Pause/resume conductor services
- **Leadership Management**: Transfer, resign, or override leadership
- **Sequencer Control**: Force active state or halt sequencers
- **Cluster Management**: Add/remove cluster members
- **Multi-Network Support**: Manage multiple sequencer networks
- **RESTful API**: Full API access for automation

### UI Features

- **Modern React SPA**: Fast, responsive single-page application
- **Client-side Routing**: Seamless navigation with React Router
- **Component-based UI**: Modular design with Radix UI components
- **State Management**: Efficient state handling with Zustand
- **Real-time Updates**: WebSocket support for live data
- **Dark Mode**: Built-in theme support with Tailwind CSS

## Requirements

- Go 1.22 or higher
- Bun runtime (for building the frontend)
- Kubernetes cluster with Optimism sequencers
- Access to Kubernetes API (kubeconfig or in-cluster)

## Installation

### Using Go Install

```bash
go install github.com/golem-base/seqctl/cmd/seqctl@latest
```

### Building from Source

```bash
git clone https://github.com/golem-base/seqctl.git
cd seqctl

# Install frontend dependencies and build
cd web && bun install && bun run build && cd ..

# Copy frontend build to server
mkdir -p pkg/server/dist
cp -r web/dist/* pkg/server/dist/

# Build Go binary with embedded frontend
go build -o seqctl ./cmd/seqctl
```

### Using Just (recommended)

```bash
# Build everything (frontend + backend)
just build

# Or build components separately
just build-web    # Build React frontend only
just build-seqctl # Build Go binary only
```

## Quick Start

1. **Basic Usage**

   ```bash
   # Start web server on default port 8080
   seqctl web
   ```

2. **Custom Configuration**

   ```bash
   seqctl serve \
     --port 9090 \
     --k8s-selector "app=op-conductor" \
     --namespaces "optimism,base"
   ```

3. **Access the Dashboard**
   ```
   http://localhost:8080
   ```

## API Reference

### Networks

```
GET    /api/v1/networks                    # List all networks
GET    /api/v1/networks/{network}          # Get network details
GET    /api/v1/networks/{network}/sequencers # List sequencers
```

### Sequencer Operations

```
POST   /api/v1/sequencers/{id}/pause       # Pause conductor
POST   /api/v1/sequencers/{id}/resume      # Resume conductor
POST   /api/v1/sequencers/{id}/transfer-leader # Transfer leadership
POST   /api/v1/sequencers/{id}/resign-leader   # Resign leadership
POST   /api/v1/sequencers/{id}/override-leader # Override leader
POST   /api/v1/sequencers/{id}/force-active    # Force active state
POST   /api/v1/sequencers/{id}/halt            # Halt sequencer
```

### Membership Management

```
PUT    /api/v1/sequencers/{id}/membership  # Add cluster member
DELETE /api/v1/sequencers/{id}/membership  # Remove cluster member
```

### Health & WebSocket

```
GET    /health                             # Health check
GET    /ws                                 # WebSocket for real-time updates
```

## Configuration

Configuration can be provided through (in order of precedence):

1. Command-line flags
2. Environment variables (prefixed with `SEQCTL_`)
3. Configuration file (`config.toml`)

### Command-Line Flags

#### Web Server

```
--address          Server listen address (default: "0.0.0.0")
--port             Server port (default: 8080)
```

#### Kubernetes

```
--k8s-config                Path to kubeconfig file
--k8s-statefulset-selector  Label selector for StatefulSets (default: "golem-base.io/optimism-role in (sequencer)")
--k8s-service-selector      Label selector for Services (default: same as statefulset-selector)
--connection-mode           Connection mode: auto|proxy|direct (default: "auto")
--namespaces                Comma-separated namespaces (empty = all)
```

#### Labels

```
--k8s-network-label  Network identification label (default: "golem-base.io/eth-network")
--k8s-role-label     Role identification label (default: "golem-base.io/optimism-role")
--k8s-app-label      App identification label (default: "app")
```

#### Logging

```
--log-level        Log level: debug|info|warn|error (default: "info")
--log-format       Log format: text|json (default: "text")
--log-no-color     Disable colored output
--log-file         Path to log file
```

### Configuration File

Create a `config.toml`:

```toml
[web]
address = "0.0.0.0"
port = 8080

[k8s]
config = "/path/to/kubeconfig"
selector = "app=op-conductor"
connection_mode = "auto"
namespaces = ["optimism", "base"]
network_label = "golem-base.io/eth-network"
role_label = "golem-base.io/optimism-role"
app_label = "app"

[log]
level = "info"
format = "text"
file = "/var/log/seqctl.log"
```

### Environment Variables

```bash
export SEQCTL_WEB_PORT=9090
export SEQCTL_K8S_SELECTOR="app=op-conductor"
export SEQCTL_K8S_NAMESPACES="optimism,base"
export SEQCTL_LOG_LEVEL=debug
```

## Provider Support

### Current Providers

- **Kubernetes**: Full support for StatefulSets and Services

### Adding a Provider

Implement the `Provider` interface:

```go
type Provider interface {
    DiscoverNetworks(ctx context.Context) (map[string]*network.Network, error)
    Name() string
}
```

## UI Technology Stack

## Development

### Project Structure

```
.
├── cmd/seqctl/    # Main application entry point
├── pkg/
│   ├── app/       # Application orchestration
│   ├── config/    # Configuration management
│   ├── flags/     # CLI flag definitions
│   ├── log/       # Structured logging
│   ├── network/   # Network domain model
│   ├── provider/  # Infrastructure providers
│   ├── repository/# Data access with caching
│   ├── sequencer/ # Sequencer domain model
│   ├── server/    # HTTP server implementation
│   │   ├── dist/  # Embedded React build
│   │   ├── handlers/  # API handlers
│   │   └── server.go  # Server setup
│   ├── swagger/   # OpenAPI documentation
│   └── version/   # Version information
└── web/           # React frontend application
    ├── src/
    │   ├── components/  # React components
    │   ├── hooks/       # Custom React hooks
    │   ├── lib/         # Utilities and helpers
    │   ├── store/       # Zustand state stores
    │   └── types/       # TypeScript types
    ├── dist/      # Build output
    └── package.json
```

### Building

```bash
# Full build (recommended)
just build

# Manual build steps
## 1. Build frontend
cd web && bun install && bun run build

## 2. Copy to server directory
mkdir -p ../pkg/server/dist
cp -r dist/* ../pkg/server/dist/

## 3. Build Go binary
cd .. && go build -o bin/seqctl ./cmd/seqctl
```

### Development Mode

```bash
# Run both frontend and backend in development mode
just dev

# Or run separately:
just dev-web  # Frontend on http://localhost:3000
just dev-go   # Backend on http://localhost:8080
```

### Running Tests

```bash
go test ./...
```

## Docker

### Building Image

```bash
# Build with Docker Buildx (recommended)
just docker-build

# Or build manually
docker build -t seqctl:latest .
```

### Running Container

```bash
docker run -p 8080:8080 \
  -v ~/.kube/config:/kube/config:ro \
  -e SEQCTL_K8S_CONFIG=/kube/config \
  seqctl:latest
```

**Note**: The Docker image contains both the Go binary and the embedded React frontend, creating a single self-contained deployment unit.

## Monitoring & Observability

- **Health Endpoint**: `/health` for liveness checks
- **Structured Logging**: JSON format available for log aggregation
- **WebSocket Updates**: Real-time data streaming at `/api/v1/ws`
- **Frontend Error Tracking**: Integrated error boundary handling
- **API Response Times**: Logged via Chi middleware
- **Metrics**: Prometheus metrics (planned)

## Security

- **Authentication**: Inherits from Kubernetes RBAC
- **TLS Support**: Configure via reverse proxy
- **CORS**: Enabled for API access
- **Input Validation**: All API inputs validated

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache 2 License - see the [LICENSE](LICENSE) file for details.
