# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o genesisgpt main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/genesisgpt .

# GenesisGpt needs OPENAI_API_KEY
ENV OPENAI_API_KEY=""

EXPOSE 8080

CMD ["./genesisgpt", "chat"]