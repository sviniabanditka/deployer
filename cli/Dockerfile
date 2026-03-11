FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/deployer .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/deployer /usr/local/bin/deployer
ENTRYPOINT ["deployer"]
