package webmock

import (
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
	domain   string
	method   string
	path     string
	query    string
	headers  map[string]string
	response string
	//TODO: status code, payload
}

type routeOptions struct {
	Logging bool
}

// FuncOption is the option for a route
type FuncOption func(*route)

// WithHeaders specifies headers to be matched
func WithHeaders(headerStr string) FuncOption {
	headers := strings.Split(headerStr, ";")
	headerMap := make(map[string]string)

	for _, header := range headers {
		pair := strings.Split(header, ":")
		headerMap[strings.TrimSpace(pair[0])] = strings.TrimSpace(pair[1])
	}

	return func(r *route) {
		r.headers = headerMap
	}
}

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
		s.server.Close()
	}
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

	fmt.Printf("===Route: %#v\n", r)

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

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, routeFound.response)
}

func routeMatch(route *route, r *http.Request) bool {
	// fmt.Printf("===Route: %#v\n", route)
	// fmt.Printf("===Request: %#v\n", r.Header)

	if route.path == r.URL.Path &&
		route.method == r.Method &&
		route.query == r.URL.RawQuery &&
		headersMatch(route.headers, r.Header) {
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

func (s *mockServer) URL() string {
	return fmt.Sprintf("http://%s", s.addr)
}
