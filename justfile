# Default command to show help
default:
    @just --list

# Build templ templates
templ-generate:
    templ generate

# Build seqctl
build-seqctl:
    go build -o bin/seqctl cmd/seqctl/main.go

# Build all binaries
build: templ-generate build-seqctl

# Generate OpenAPI specification
swagger:
    swag init \
        -g pkg/swagger/apigen.go \
        -d . \
        -o pkg/swagger \
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

# Clean all artifacts
clean:
    rm -rf ./bin
