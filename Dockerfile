FROM --platform=$BUILDPLATFORM golang:1.20 AS builder

WORKDIR /app

# Download dependencies first for better caching.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/okta-operator /okta-operator
USER 65532:65532
ENTRYPOINT ["/okta-operator"]
