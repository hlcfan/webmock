package webmock

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v2"
)

// MockServer holds mock server and routes
type MockServer struct {
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
	body            string
	responseHeaders map[string]string
	//TODO: payload
}

// FuncOption is the option for a route
type FuncOption func(*route)

type cassetteRoute struct {
	Request  httpRequest  `yaml:"request"`
	Response httpResponse `yaml:"response"`
}

type cassetteRoutes []cassetteRoute

type httpRequest struct {
	Method  string            `yaml:"method"`
	Path    string            `yaml:"path"`
	Headers map[string]string `yaml:"headers"`
}

type httpResponse struct {
	Status  int               `yaml:"status"`
	Headers map[string]string `yaml:"headers"`
	Body    string            `yaml:"body"`
}

// New creates a mock server, it will listen on a unoccupied port
func New() *MockServer {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal("allocate listen address fail: ", err)
	}

	addr := listener.Addr().String()

	listener.Close()

	srv := &http.Server{
		Addr: addr,
	}

	return &MockServer{
		addr:   addr,
		server: srv,
		routes: make([]*route, 0),
	}
}

// Start starts the mock server in a goroutine
func (s *MockServer) Start() {
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
func (s *MockServer) Stop() {
	if s.server != nil {
		s.server.Shutdown(context.TODO())
	}
}

// Reset resets mocked routes
func (s *MockServer) Reset() {
	s.routes = make([]*route, 0)
	s.server.Handler = s
}

// URL returns the base URL of the mock server
func (s *MockServer) URL() string {
	return fmt.Sprintf("http://%s", s.addr)
}

// Stub loads stub requests into routes
func (s *MockServer) Stub(method, uri string, response string, options ...FuncOption) {
	url, err := url.Parse(uri)
	if err != nil {
		log.Fatal("invalid url: ", err)
	}

	r := &route{
		domain: url.Host,
		path:   url.Path,
		method: method,
		query:  url.RawQuery,
		body:   response,
	}

	for _, opt := range options {
		opt(r)
	}

	s.routes = append(s.routes, r)

	s.server.Handler = s
}

// LoadCassette loads cassette files
func (s *MockServer) LoadCassette(path string) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatal("file or dir not exist: ", path)
		} else {
			log.Fatalf("reading file or dir `%s` fail: %s", path, err)
		}
	}

	var routes []*route

	switch mode := stat.Mode(); {
	case mode.IsDir():
		r := loadCassettes(path)
		routes = append(routes, r...)
	case mode.IsRegular():
		r := loadCassette(path)
		routes = append(routes, r...)
	}

	s.routes = append(s.routes, routes...)
	s.server.Handler = s
}

func loadCassettes(dirPath string) []*route {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatal("read dir fails: ", err)
	}

	var routes []*route
	for _, f := range files {
		r := loadCassette(path.Join(dirPath, f.Name()))
		routes = append(routes, r...)
	}

	return routes
}

func loadCassette(filePath string) []*route {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("fail to read file: %s, err: %s", filePath, err)
	}

	var cassettes cassetteRoutes
	err = yaml.NewDecoder(file).Decode(&cassettes)
	if err != nil {
		log.Fatalf("invalid yaml format: %s, err: %s", filePath, err)
	}

	var routes []*route

	for _, c := range cassettes {
		url, err := url.Parse(c.Request.Path)
		if err != nil {
			log.Fatalf("invalid url: %s, err: %s", c.Request.Path, err)
		}

		r := &route{
			path:            url.Path,
			method:          strings.ToUpper(c.Request.Method),
			query:           url.RawQuery,
			body:            c.Response.Body,
			responseHeaders: c.Response.Headers,
			statusCode:      c.Response.Status,
		}

		routes = append(routes, r)
	}

	return routes
}

// ServeHTTP implements the server.Handler
// It go over all existing routes and find the one matches and render response
// based on the found route
func (s *MockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	io.WriteString(w, routeFound.body)
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
			r.body = response
		}

		if len(headers) > 0 {
			r.responseHeaders = headers
		}
	}
}
