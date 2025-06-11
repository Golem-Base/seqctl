# Stage 1: Build frontend
FROM oven/bun:1 AS frontend-builder

WORKDIR /app

COPY web/package.json ./
COPY web/bun.lock ./

# Install frontend dependencies
RUN bun install

# Copy frontend source
COPY web/ ./

# Build frontend
RUN bun run build

# Stage 2: Build backend
FROM golang:1.24.2 AS backend-builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY pkg/ pkg/

# Copy built frontend from frontend-builder stage
COPY --from=frontend-builder /app/dist pkg/server/dist

# Build with version information
ARG GIT_COMMIT=""
ARG GIT_DATE=""
ARG VERSION="v0.1.0"

# Build seqctl binary with embedded frontend
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X github.com/golem-base/seqctl/pkg/version.GitCommit=${GIT_COMMIT} \
    -X github.com/golem-base/seqctl/pkg/version.GitDate=${GIT_DATE} \
    -X github.com/golem-base/seqctl/pkg/version.Version=${VERSION}" \
    -a -o seqctl cmd/seqctl/main.go

# Stage 3: Final minimal image
FROM gcr.io/distroless/static:nonroot

# Add labels for container metadata
LABEL org.opencontainers.image.title="seqctl" \
    org.opencontainers.image.description="Web-based control panel for managing Optimism conductor sequencer clusters" \
    org.opencontainers.image.vendor="golem-base" \
    org.opencontainers.image.source="https://github.com/golem-base/seqctl"

WORKDIR /
COPY --from=backend-builder /workspace/seqctl .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/seqctl"]
CMD ["serve"]
