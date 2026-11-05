# Production image. Build from apps/api: needs no JS tooling or other workspaces.
#   docker build -t turgorepo-api apps/api
FROM golang:1.27.3-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ ./cmd/server ./cmd/migrate ./cmd/seed

# Includes CA certificates for TLS to Neon, RDS and Redis, and has no shell.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/ /app/
COPY openapi.json /app/openapi.json

ENV GIN_MODE=release
ENV PORT=5000
ENV OPENAPI_PATH=/app/openapi.json

EXPOSE 5000

CMD ["/app/server"]
