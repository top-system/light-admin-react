# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Cache module downloads.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Pure-Go build (pgx + modernc sqlite are cgo-free), statically linked.
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags "-w -s" -o /out/echo-admin .

# ---- runtime stage ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 10001 app
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /out/echo-admin /app/echo-admin
# Migrations are embedded in the binary; only runtime config is copied.
COPY config /app/config

USER app
EXPOSE 8080

# Override the config path via --config if needed.
ENTRYPOINT ["/app/echo-admin"]
CMD ["runserver", "--config=./config/config.yaml", "--casbin_model=./config/casbin_model.conf"]
