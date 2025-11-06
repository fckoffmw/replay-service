
FROM golang:1.24.2-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/replay-service ./server/cmd/replay-service

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/replay-service /app/replay-service
# storage dir for files
VOLUME ["/app/storage"]
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/app/replay-service"]