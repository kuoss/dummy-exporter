FROM golang:1.23 AS builder

ARG VERSION

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-X 'main.Version=$VERSION'" -a -o dummy-exporter cmd/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/dummy-exporter    .
COPY --from=builder /workspace/etc/exporter.yaml ./etc/
USER 65532:65532

ENTRYPOINT ["/dummy-exporter"]
