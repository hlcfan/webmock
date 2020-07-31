## Webmock

This package creates a HTTP server, it stubs requests. Inspired by [bblimke/webmock](https://github.com/bblimke/webmock)

## Examples

### Stubbed request based on method, url

```go
server := webmock.New()
baseURL := server.URL()

server.Start()

server.Stub("GET", "/hello", "ok")
server.Stub("POST", "/hello", "ok-post")

// curl http://server/hello
// curl -XPOST http://server/hello
```

### Stubbed request based on method, url and headers

```go
server.Stub(
    "GET",
    "/hello",
    "ok",
    webmock.WithHeaders("Content-Type: application/json")
)

// curl -H "Content-Type: application/json" http://server/hello
```

### 
