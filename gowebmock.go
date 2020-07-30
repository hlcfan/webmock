package gowebmock

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

type mockServer struct {
	addr   string
	server *http.Server
	routes map[string][]byte
}

func NewMockServer() *mockServer {
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
		routes: make(map[string][]byte),
	}
}

func (s *mockServer) URL() string {
	return fmt.Sprintf("http://%s", s.addr)
}

func (s *mockServer) Serve() {
	s.server.Handler = s
	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("fail to serve : ", err)
	}

	s.addr = s.server.Addr
}

func (s *mockServer) Stub(method, url string, response []byte) {
	s.routes[method+url] = response
	s.server.Handler = s
}

func (s *mockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.URL.Path
	resp := s.routes[method+path]

	if len(resp) > 0 {
		io.WriteString(w, string(resp))
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
