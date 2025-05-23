# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o k8sgpt main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/k8sgpt .

# K8sGpt needs OPENAI_API_KEY
ENV OPENAI_API_KEY=""

EXPOSE 8080

CMD ["./k8sgpt", "chat"]