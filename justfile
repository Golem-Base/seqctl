# Default command to show help
default:
    @just --list

# Build React app
build-web:
    cd web && bun install && bun run build
    mkdir -p pkg/server/dist
    cp -r web/dist/* pkg/server/dist/

# Build seqctl
build-seqctl:
    go build -o bin/seqctl cmd/seqctl/main.go

# Build all binaries
build: build-web build-seqctl

# Generate OpenAPI specification
swagger:
    swag init \
        -g pkg/server/swagger/apigen.go \
        -d . \
        -o pkg/server/swagger \
        --parseDependency \
        --parseInternal \
        --parseDepth 2

# Build the seqctl Docker image
docker-build target="default" tag="latest":
    docker buildx bake --file build.hcl {{ target }} \
        --set "*.args.GIT_COMMIT=$(just git-commit)" \
        --set "*.args.GIT_DATE=$(just git-date)"

# Get git commit
git-commit:
    @git rev-parse --short HEAD 2>/dev/null || echo "unknown"

# Get git date
git-date:
    @git log -1 --format=%cd --date=short 2>/dev/null || echo "unknown"

# Development mode - run React dev server
dev-web:
    cd web && bun run dev

# Development mode - run Go server (with air if available)
dev-go:
    @if command -v air >/dev/null 2>&1; then \
        echo "Running with air for hot reload..."; \
        air; \
    else \
        echo "Running without hot reload (install air for hot reload)..."; \
        go run cmd/seqctl/main.go web; \
    fi

# Development mode - run both servers in parallel
dev:
    #!/usr/bin/env bash
    echo "Starting development servers..."
    echo "React dev server will run on http://localhost:3000"
    echo "Go API server will run on http://localhost:8080"
    echo ""

    # Function to kill both processes on exit
    cleanup() {
        echo -e "\nShutting down servers..."
        kill $WEB_PID $GO_PID 2>/dev/null
        exit
    }

    # Set up trap for clean exit
    trap cleanup INT TERM EXIT

    # Start web server in background
    (cd web && bun run dev) &
    WEB_PID=$!

    # Give the web server a moment to start
    sleep 2

    # Start Go server in foreground
    if command -v air >/dev/null 2>&1; then
        air
    else
        go run cmd/seqctl/main.go web
    fi

# Clean all artifacts
clean:
    rm -rf ./bin
    rm -rf ./web/dist
    rm -rf ./pkg/server/dist

# Lint the code
lint:
    revive --config revive.toml ./...
