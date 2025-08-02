FROM golang:1.24.5-alpine3.22 AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY . /app/

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo 

FROM alpine:latest AS runner

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/gateway-purch .
COPY --from=builder /app/config.yml .
COPY --from=builder /app/LICENSE .