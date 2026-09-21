# syntax=docker/dockerfile:1.7

############################
# Build API
############################
FROM golang:1.24.1-alpine3.21 AS api-builder

WORKDIR /app/api

COPY api/go.mod api/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY api ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go test ./... && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server .

############################
# Build frontend
############################
FROM node:22.14.0-alpine3.21 AS frontend-builder

WORKDIR /app

RUN corepack enable && corepack prepare yarn@1.22.22 --activate
COPY frontend/package.json frontend/yarn.lock ./
RUN --mount=type=cache,target=/yarn-cache \
    YARN_CACHE_FOLDER=/yarn-cache yarn install --frozen-lockfile --network-timeout 600000

COPY frontend ./
RUN yarn build

############################
# Build runtime image
############################
FROM alpine:3.21.3

ARG VERSION=dev
ARG REVISION=unknown
ENV SERVICE_VERSION="${VERSION}"
LABEL org.opencontainers.image.title="go-url" \
      org.opencontainers.image.description="A Go URL shortener with a React frontend" \
      org.opencontainers.image.source="https://github.com/kjwardy/go-url" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${REVISION}"

RUN apk add --no-cache ca-certificates dumb-init && \
    addgroup -S app && \
    adduser -S -G app app

WORKDIR /go/bin

COPY --from=api-builder --chown=app:app /out/server ./server
COPY --from=frontend-builder --chown=app:app /app/build ./public
COPY --chown=app:app entrypoint.sh ./entrypoint.sh

USER app
EXPOSE 1323

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/go/bin/entrypoint.sh"]
