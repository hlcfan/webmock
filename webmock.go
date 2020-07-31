package webmock

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type mockServer struct {
	addr   string
	server *http.Server
	routes []*route
}

type route struct {
	domain          string
	method          string
	path            string
	query           string
	requestHeaders  map[string]string
	statusCode      int
	response        string
	responseHeaders map[string]string
	//TODO: payload
}

// FuncOption is the option for a route
type FuncOption func(*route)

// New creates a mock server, it will listen on a unoccupied port
func New() *mockServer {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal("allocate listen address fail: ", err)
	}

	addr := listener.Addr().String()

	listener.Close()

	srv := &http.Server{
		Addr: addr,
	}

	return &mockServer{
		addr:   addr,
		server: srv,
		routes: make([]*route, 0),
	}
}

// Start starts the mock server in a goroutine
func (s *mockServer) Start() {
	s.server.Handler = s

	go func() {
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal("fail to serve : ", err)
		}

		s.addr = s.server.Addr
	}()
}

// Stop stops the mock server
func (s *mockServer) Stop() {
	if s.server != nil {
		s.server.Shutdown(context.TODO())
	}
}

// URL returns the base URL of the mock server
func (s *mockServer) URL() string {
	return fmt.Sprintf("http://%s", s.addr)
}

// Stub loads stub requests into routes
// TODO: load cassettes
func (s *mockServer) Stub(method, uri string, response string, options ...FuncOption) {
	url, err := url.Parse(uri)
	if err != nil {
		log.Fatal("invalid url: ", err)
	}

	r := &route{
		domain:   url.Host,
		path:     url.Path,
		method:   method,
		query:    url.RawQuery,
		response: response,
	}

	for _, opt := range options {
		opt(r)
	}

	s.routes = append(s.routes, r)

	s.server.Handler = s
}

// ServeHTTP implements the server.Handler
// It go over all existing routes and find the one matches and render response
// based on the found route
func (s *mockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var routeFound *route

	for _, route := range s.routes {
		if routeMatch(route, r) {
			routeFound = route
		}
	}

	if routeFound == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	for headerKey, headerVal := range routeFound.responseHeaders {
		w.Header().Set(headerKey, headerVal)
	}

	statusCode := routeFound.statusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	w.WriteHeader(statusCode)

	io.WriteString(w, routeFound.response)
}

func routeMatch(route *route, r *http.Request) bool {
	if route.path == r.URL.Path &&
		route.method == r.Method &&
		route.query == r.URL.RawQuery &&
		headersMatch(route.requestHeaders, r.Header) {
		return true
	}

	return false
}

func headersMatch(routeHeaders map[string]string, requestHeader http.Header) bool {
	for key, val := range routeHeaders {
		if val != requestHeader.Get(key) {
			return false
		}
	}

	return true
}

// WithHeaders specifies headers to be matched
func WithHeaders(headerStr string) FuncOption {
	headers := strings.Split(headerStr, ";")
	headerMap := make(map[string]string)

	for _, header := range headers {
		pair := strings.Split(header, ":")
		headerMap[strings.TrimSpace(pair[0])] = strings.TrimSpace(pair[1])
	}

	return func(r *route) {
		r.requestHeaders = headerMap
	}
}

// WithResponse specifies response to be rendered
func WithResponse(code int, response string, headers map[string]string) FuncOption {
	return func(r *route) {
		if len(http.StatusText(code)) > 0 {
			r.statusCode = code
		}

		if len(response) > 0 {
			r.response = response
		}

		if len(headers) > 0 {
			r.responseHeaders = headers
		}
	}
}
