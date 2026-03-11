# Go

## Project setup

Deployer detects Go projects by the presence of `go.mod`.

### Minimal example

```go
package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message": "Hello from Deployer!"}`)
	})

	fmt.Printf("Listening on :%s\n", port)
	http.ListenAndServe(":"+port, nil)
}
```

Initialize the module:

```bash
go mod init myapp
```

## Port binding

Read the port from the `PORT` environment variable and listen on `0.0.0.0`:

```go
http.ListenAndServe(":"+port, nil)  // binds to 0.0.0.0 by default
```

## Deploy

```bash
deployer init
deployer deploy
```

Deployer builds your Go binary with `go build` and runs it inside an Alpine-based container.

## Using a custom Dockerfile

For more control, use a multi-stage build:

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

## Using popular frameworks

### Gin

```go
r := gin.Default()
r.GET("/", func(c *gin.Context) {
    c.JSON(200, gin.H{"message": "hello"})
})
r.Run(":" + os.Getenv("PORT"))
```

### Fiber

```go
app := fiber.New()
app.Get("/", func(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"message": "hello"})
})
app.Listen(":" + os.Getenv("PORT"))
```

## Adding a database

```bash
deployer db create postgres
deployer db link <db-id> <app-id>
```

```go
db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
```

## Common issues

| Problem | Solution |
|---------|----------|
| `CGO_ENABLED` errors | Set `CGO_ENABLED=0` in your Dockerfile |
| Binary not found | Ensure the build output path matches the `CMD` |
| TLS errors | Add `ca-certificates` in the final stage |
