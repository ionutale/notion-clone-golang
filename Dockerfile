FROM node:22-alpine AS frontend-builder
WORKDIR /app
RUN npm install -g pnpm
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install
COPY web/ .
RUN pnpm build

FROM golang:1.26-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/build ./web/build
RUN go build -o bin/notion-clone .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend-builder /app/bin/notion-clone .
COPY --from=backend-builder /app/web/build ./web/build
COPY --from=backend-builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./notion-clone"]
