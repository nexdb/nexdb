FROM --platform=$BUILDPLATFORM golang:1.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -o nexdb cmd/server/main.go

# Stage 2: Create the final Docker image
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/nexdb .

EXPOSE 9000
CMD ["./nexdb"]
