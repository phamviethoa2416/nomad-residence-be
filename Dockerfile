FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/app ./internal/cmd/server

FROM alpine:3.22 AS runtime

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S -G app app

ENV CONFIG_PATH=/app \
    TZ=Asia/Ho_Chi_Minh

COPY --from=builder /out/app /app/app
COPY config.yaml /app/config.yaml

USER app

EXPOSE 8080

ENTRYPOINT ["/app/app"]
