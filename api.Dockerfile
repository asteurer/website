FROM golang:1.19 AS builder

WORKDIR /app

# Copy only the necessary files for building the application
COPY go.mod .
COPY go.sum .
COPY cmd/ ./cmd

# Download Go modules before adding the entire context
RUN go mod download

# Build the application as a static binary
RUN CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' -o /main ./cmd

FROM alpine

WORKDIR /app

COPY --from=builder /main /app/main
COPY cmd/media /app/media
COPY cmd/templates /app/templates
COPY cmd/sql /app/sql

EXPOSE 8080

ENTRYPOINT ["/app/main"]