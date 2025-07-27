# Build Frontend
FROM node:18-alpine AS frontend
WORKDIR /app/ui
COPY ui/package.json ui/pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install
COPY ui/. .
RUN pnpm build

# Build Backend
FROM golang:1.23-alpine AS backend
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/. .
RUN go build -o monkey-code-server ./cmd/server

# Final Stage
FROM alpine:latest
WORKDIR /app
COPY --from=frontend /app/ui/dist /app/ui/dist
COPY --from=backend /app/monkey-code-server .
COPY --from=backend /app/config/config.yaml /app/config/config.yaml
COPY --from=backend /app/migration /app/migration

RUN mkdir -p /app/static && touch /app/static/.machine_id

EXPOSE 8888

CMD ["./monkey-code-server"]