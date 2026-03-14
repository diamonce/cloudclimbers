# Build stage
FROM golang:1.25.8-alpine AS builder

WORKDIR /app

COPY ./cloudclimbers-slack-bot/go.mod ./cloudclimbers-slack-bot/go.sum ./

RUN go get golang.org/x/crypto@v0.35.0 && go mod download

COPY ./cloudclimbers-slack-bot/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o cloudclimbers-slack-bot ./cmd/main.go

# Final stage
FROM alpine:3.21

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder /app/cloudclimbers-slack-bot .

# Run as non-root user
USER appuser

EXPOSE 8080

CMD ["./cloudclimbers-slack-bot"]
