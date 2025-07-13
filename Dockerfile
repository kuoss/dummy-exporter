FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o dummy-exporter ./cmd/dummy-exporter

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/dummy-exporter /app/dummy-exporter
COPY etc/exporter.yaml /app/etc/exporter.yaml
EXPOSE 9100

CMD ["/app/dummy-exporter"]
