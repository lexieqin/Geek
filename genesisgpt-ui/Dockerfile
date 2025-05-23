# UI Server Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY proxy-server.go .
RUN go build -o ui-server proxy-server.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/ui-server .
COPY index.html .

EXPOSE 3000

CMD ["./ui-server"]