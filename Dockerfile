FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bin/notion-clone .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/notion-clone .
COPY --from=builder /app/web/build ./web/build
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./notion-clone"]
