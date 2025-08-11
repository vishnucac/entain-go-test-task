FROM golang:1.23 AS builder

WORKDIR /app

# Install dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/main.go

# Step 2: Run the app
FROM alpine:3.20

WORKDIR /root/

COPY --from=builder /app/app .

RUN adduser -D appuser
USER appuser

EXPOSE 8080

CMD ["./app"]
