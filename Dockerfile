FROM golang:1.18 as builder
WORKDIR /app
COPY ./cloudclimbers-slack-bot/go.mod ./cloudclimbers-slack-bot/go.sum ./
RUN go mod download
COPY ./cloudclimbers-slack-bot/ .
RUN go build -o cloudclimbers-slack-bot ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/cloudclimbers-slack-bot .
EXPOSE 8080
CMD ["./cloudclimbers-slack-bot"]
