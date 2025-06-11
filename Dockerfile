FROM golang:1.24.2 AS builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY pkg/ pkg/

# Build with version information
ARG GIT_COMMIT=""
ARG GIT_DATE=""
ARG VERSION="v0.1.0"

# Build populator
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X github.com/golem-base/seqctl/pkg/version.GitCommit=${GIT_COMMIT} \
              -X github.com/golem-base/seqctl/pkg/version.GitDate=${GIT_DATE} \
              -X github.com/golem-base/seqctl/pkg/version.Version=${VERSION}" \
    -a -o seqctl cmd/seqctl/main.go

# Use distroless as minimal base image to package the populator binary
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/seqctl .
USER 65532:65532

ENTRYPOINT ["/seqctl"]
