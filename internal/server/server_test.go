package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/voukatas/url-shortener/internal/model"
	"github.com/voukatas/url-shortener/internal/url_converter"
)

var (
	expectedGetUrl = "http://example.com"
	id             = int64(1)
)

// mock db
type mockStore struct {
}

func (store *mockStore) Shorten(string) (int64, error) {
	return id, nil
}
func (store *mockStore) Lookup(int64) (string, error) {
	return expectedGetUrl, nil
}
func (store *mockStore) Close() {
}
func (store *mockStore) SetStoreOptions() {
}

// mock logger
type mockLogger struct {
}

// No Operation
func (s *mockLogger) Debug(msg string, args ...interface{}) {}
func (s *mockLogger) Info(msg string, args ...interface{})  {}
func (s *mockLogger) Warn(msg string, args ...interface{})  {}
func (s *mockLogger) Error(msg string, args ...interface{}) {}

// mock cache
type mockCache struct {
	getFuncCalled bool
	setFuncCalled bool
}

func (lru *mockCache) Set(key string, value string) {
	lru.setFuncCalled = true
}
func (lru *mockCache) Get(key string) (string, error) {
	lru.getFuncCalled = true
	return "", errors.New("Key not found")
}

var shuffleKey = "your_key"

func TestRedirectURLSuccess(t *testing.T) {
	url_converter.InitBase62Array(shuffleKey)
	server := NewServer(&mockStore{}, http.NewServeMux(), &model.Config{XorSecretKey: 15489079}, &mockLogger{}, &mockCache{})

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("url", "12zPr")
	resp := httptest.NewRecorder()

	server.RedirectURL(resp, req)
	if resp.Code != http.StatusFound {
		t.Errorf("expected %v received %v", http.StatusFound, resp.Code)
	}
	if resp.Header().Get("Location") != expectedGetUrl {
		t.Errorf("expected %v received %v", expectedGetUrl, resp.Header().Get("Location"))
	}

}

func TestRedirectURLSuccessWithCacheUse(t *testing.T) {
	url_converter.InitBase62Array(shuffleKey)
	mCache := &mockCache{}
	server := NewServer(&mockStore{}, http.NewServeMux(), &model.Config{XorSecretKey: 15489079}, &mockLogger{}, mCache)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("url", "12zPr")
	resp := httptest.NewRecorder()

	server.RedirectURL(resp, req)
	if resp.Code != http.StatusFound {
		t.Errorf("expected %v received %v", http.StatusFound, resp.Code)
	}
	if resp.Header().Get("Location") != expectedGetUrl {
		t.Errorf("expected %v received %v", expectedGetUrl, resp.Header().Get("Location"))
	}

	if !mCache.getFuncCalled || !mCache.setFuncCalled {
		t.Error("Unexpected - Cache was not used")
	}

}

func TestRedirectURLStatusBadRequest(t *testing.T) {

	url_converter.InitBase62Array(shuffleKey)
	//server := NewServer(&mockStore{}, http.NewServeMux())
	server := NewServer(&mockStore{}, http.NewServeMux(), &model.Config{XorSecretKey: 15489079}, &mockLogger{}, &mockCache{})

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	server.RedirectURL(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected %v received %v", http.StatusBadRequest, resp.Code)
	}
	if resp.Header().Get("Location") != "" {
		t.Errorf("expected %v received %v", expectedGetUrl, resp.Header().Get("Location"))
	}

}
func TestCreateShortURLSuccess(t *testing.T) {
	url_converter.InitBase62Array(shuffleKey)
	//server := NewServer(&mockStore{}, http.NewServeMux())
	server := NewServer(&mockStore{}, http.NewServeMux(), &model.Config{XorSecretKey: 15489079}, &mockLogger{}, &mockCache{})

	body := []byte(`{"url": "http://example.com"}`)
	req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()

	server.CreateShortURL(resp, req)
	rspBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	expectedPostResult := `{"long_url":"http://example.com","short_url":"neBlT"}`

	var expected, actual map[string]string
	if err := json.Unmarshal([]byte(expectedPostResult), &expected); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}

	if err := json.Unmarshal(rspBody, &actual); err != nil {
		t.Fatalf("Failed to unmarshal response JSON: %v", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected: %v received: %v", expected, actual)
	}

	if resp.Code != http.StatusCreated {
		t.Errorf("expected: %v received: %v", http.StatusCreated, resp.Code)
	}
}

func TestCreateShortURLBadRequest(t *testing.T) {
	url_converter.InitBase62Array(shuffleKey)
	server := NewServer(&mockStore{}, http.NewServeMux(), &model.Config{XorSecretKey: 15489079}, &mockLogger{}, &mockCache{})

	body := []byte("invalid_json")
	req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()

	server.CreateShortURL(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected: %v received: %v", http.StatusBadRequest, resp.Code)
	}
}

func TestCreateShortURLMissingProtocol(t *testing.T) {
	url_converter.InitBase62Array(shuffleKey)
	server := NewServer(&mockStore{}, http.NewServeMux(), &model.Config{XorSecretKey: 15489079}, &mockLogger{}, &mockCache{})

	body := []byte(`{"url": "example.com"}`)
	req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()

	server.CreateShortURL(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected: %v received: %v", http.StatusBadRequest, resp.Code)
	}
}
