package gowebmock_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/hlcfan/gowebmock"
)

func TestWebMock(t *testing.T) {
	server := gowebmock.New()
	baseURL := server.URL()
	fmt.Println("===", baseURL)
	server.Start()

	defer server.Stop()

	client := &http.Client{}

	t.Run("It serves stub http requests with GET", func(t *testing.T) {
		server.Stub("GET", "/abc", "ok")

		resp, err := http.Get(baseURL + "/abc")
		if err != nil {
			panic(err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("unexpected response status, want: %d, got: %d", http.StatusOK, resp.StatusCode)
		}

		defer resp.Body.Close()

		respBodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		respBody := string(respBodyBytes)
		if respBody != "ok" {
			t.Errorf("unexpected response body, want: %s, got: %s", "ok", respBody)
		}
	})

	t.Run("It serves stub http requests with POST", func(t *testing.T) {
		server.Stub("POST", "/post", "ok post")

		payload := []byte(`{"name":"Alex"}`)
		req, err := http.NewRequest("POST", baseURL+"/post", bytes.NewBuffer(payload))
		if err != nil {
			panic(err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("unexpected response status, want: %d, got: %d", http.StatusOK, resp.StatusCode)
		}

		defer resp.Body.Close()

		respBodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		respBody := string(respBodyBytes)
		if respBody != "ok post" {
			t.Errorf("unexpected response body, want: %s, got: %s", "ok", respBody)
		}
	})

	t.Run("It serves stub http requests when query parameters match", func(t *testing.T) {
		url := "/get?foo=bar&a=b"
		response := "ok with query parameters"

		server.Stub("GET", url, response)

		resp, err := http.Get(baseURL + "/get?foo=bar&a=b")
		if err != nil {
			panic(err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("unexpected response status, want: %d, got: %d", http.StatusOK, resp.StatusCode)
		}

		defer resp.Body.Close()

		respBodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		respBody := string(respBodyBytes)
		if respBody != response {
			t.Errorf("unexpected response body, want: %s, got: %s", "ok", respBody)
		}
	})

	t.Run("It serves stub http requests when query parameters don't match", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/get?foo=bar")
		if err != nil {
			panic(err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("unexpected response status, want: %d, got: %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("It serves stub http requests when headers match", func(t *testing.T) {
		url := "/get"
		response := "ok with headers"

		server.Stub("GET", url, response, gowebmock.WithHeaders("Accept-Encoding: gzip,deflate"))

		req, err := http.NewRequest("GET", baseURL+url, nil)
		if err != nil {
			panic(err)
		}

		req.Header.Set("Accept-Encoding", "gzip,deflate")

		resp, err := client.Do(req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("unexpected response status, want: %d, got: %d", http.StatusOK, resp.StatusCode)
		}

		defer resp.Body.Close()

		respBodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		respBody := string(respBodyBytes)
		if respBody != response {
			t.Errorf("unexpected response body, want: %s, got: %s", "ok", respBody)
		}
	})

	t.Run("It serves stub http requests when headers don't match", func(t *testing.T) {
		url := "/get"
		response := "ok with headers"

		server.Stub("GET", url, response, gowebmock.WithHeaders("Accept-Encoding: gzip,deflate"))

		req, err := http.NewRequest("GET", baseURL+url, nil)
		if err != nil {
			panic(err)
		}

		req.Header.Set("Accept-Encoding", "gzip")

		resp, err := client.Do(req)

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("unexpected response status, want: %d, got: %d", http.StatusOK, resp.StatusCode)
		}
	})
}
