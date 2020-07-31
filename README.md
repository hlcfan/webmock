## Webmock

This package creates a HTTP server, it stubs requests. Inspired by [bblimke/webmock](https://github.com/bblimke/webmock). It's useful when writing integration tests while code reply on external requests. With *webmock*, just point the endpoint/host to mock server and stub the requests.

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
    webmock.WithHeaders("Content-Type: application/json"),
)

// curl -H "Content-Type: application/json" http://server/hello
```

### Stubbing with custom response

```go
server.Stub(
  "GET",
  "/abc?foo=bar",
  "",
  webmock.WithHeaders("Accept: application/json"),
  webmock.WithResponse(500, "Ah oh", map[string]string{
    "Access-Control-Allow-Origin": "*",
    "Content-Type":                "application/xml",
  }),
)

// curl -i -H "Content-Type: application/json" http://server/abc?foo=bar
```
