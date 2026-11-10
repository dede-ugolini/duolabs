# syntax=docker/dockerfile:1

# ---------- Build the React frontend ----------
FROM node:22-alpine AS frontend

WORKDIR /src/frontend

# Install dependencies first so this layer is cached when source changes.
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
# Empty base URL: the Go server serves the built app, so requests are same-origin.
ARG VITE_API_BASE=""
ENV VITE_API_BASE=$VITE_API_BASE
RUN npm run build

# ---------- Build the Go backend ----------
FROM golang:1.26-alpine AS backend

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# glebarez/sqlite is pure Go, so a static binary needs no cgo.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/duolabs .

# ---------- Runtime ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
	&& adduser -D -u 10001 app \
	&& mkdir -p /data /app/static

WORKDIR /app

COPY --from=backend /out/duolabs /app/duolabs
COPY --from=frontend /src/frontend/dist/ /app/static/

RUN chown -R app:app /app /data

ENV PORT=8081 \
	DB_PATH=/data/duolabs.db \
	STATIC_DIR=/app/static

USER app

VOLUME ["/data"]
EXPOSE 8081

ENTRYPOINT ["/app/duolabs"]
