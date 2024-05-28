# Build stage
FROM golang:1.22.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o create_flux .

# Final stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/create_flux .

EXPOSE 8085

CMD ["./create_flux"]
