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
RUN adduser -S appuser appuser

RUN chown -R appuser /app/config
USER appuser

ENV APP_NAME=PLACE_HOLDER

ENTRYPOINT ["/app/http-ping","launchHttp","--appName=backend","--config=/app/config/sam-ping.yaml"]
EXPOSE 5115
