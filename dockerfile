FROM golang:1.24.5-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum config.go main.go config.yml /app/

RUN go build . 

FROM golang:1.24.5-alpine3.22 AS runner

WORKDIR /app

# We just need the binary to run
COPY --from=builder /app/gateway-purch /app/gateway-purch
COPY --from=builder /app/config.yml /app/config.yml
COPY LICENSE /app/LICENSE
# Run the binary
CMD ["/app/gateway-purch"]