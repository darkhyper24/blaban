FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/cmd
RUN go build -o /app/bin/api-gateway main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/bin/api-gateway .
EXPOSE 8080
CMD ["./api-gateway"]