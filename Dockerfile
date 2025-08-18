# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build application binary
RUN CGO_ENABLED=0 GOOS=linux go build -o notiflow ./cmd/notiflow

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/notiflow .

VOLUME ["/etc/notiflow"]

EXPOSE 8080

ENTRYPOINT ["./notiflow"]
CMD ["-config.file=/etc/notiflow/config.yaml"]