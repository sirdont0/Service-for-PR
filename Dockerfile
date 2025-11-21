# builder
FROM golang:1.20-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app ./cmd/server

# runtime
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /app /app
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app"]
