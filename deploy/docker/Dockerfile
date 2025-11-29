
FROM golang:1.24.2-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/replay-service ./server/cmd/replay-service

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /out/replay-service /app/replay-service
EXPOSE 8080
ENTRYPOINT ["/app/replay-service"]