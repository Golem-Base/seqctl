# Default command to show help
default:
    @just --list

# Build seqctl
build-seqctl:
    go build -o bin/seqctl cmd/seqctl/main.go

# Build all binaries
build: build-seqctl

# Build the seqctl Docker image
docker-build-seqctl version="latest":
    docker build -f Dockerfile -t quay.io/golemnetwork/seqctl:{{version}} .

# Build all Docker images
docker-build: docker-build-seqctl

# Clean all artifacts
clean:
    rm -rf ./bin
