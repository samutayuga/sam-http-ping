FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./

RUN go mod download


COPY cmd ./cmd
COPY main.go ./main.go



RUN go mod tidy
RUN go build -o /http-ping .
# Final Stage - Stage 2
FROM alpine:3.14.2 as baseImage

WORKDIR /app


COPY --from=builder /http-ping ./http-ping
COPY config ./config
RUN adduser -S appuser

RUN GRPC_HEALTH_PROBE_VERSION=v0.4.13 && \
    wget -qO/app/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /app/grpc_health_probe

ENTRYPOINT ["/app/http-ping","-config=/app/config/sam-ping.yaml"]
EXPOSE 5115
