package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/voukatas/url-shortener/internal/config"
	"github.com/voukatas/url-shortener/internal/store"
	"github.com/voukatas/url-shortener/internal/url_converter"
	"github.com/voukatas/url-shortener/pkg/cache"
	"github.com/voukatas/url-shortener/pkg/logger"
)

//helper functions

func waitForServerToStart(url string, timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return nil

		} else {
			fmt.Println(err.Error())
		}
		time.Sleep(100 * time.Millisecond)

	}
	return fmt.Errorf("server failed to start after :%v", timeout)
}

func TestAPI(t *testing.T) {
	config, err := config.LoadConfig("../../short_conf.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	store, err := store.NewStore("file:test.db?mode=rwc")
	if err != nil {
		log.Printf("%v", err.Error())
		return
	}
	defer func() {
		log.Println("DB Closed properly")
		store.Close()
		defer os.Remove("test.db")
	}()
	store.SetStoreOptions()

	url_converter.InitBase62Array(shuffleKey)
	// logging
	slogger, cleanup := logger.SetupLogger(config.LogFilename, config.LogLevel, config.Production)
	defer func() {
		cleanup()
		os.Remove("test.log")
	}()

	// cache
	cache := cache.NewCache(config.CacheCapacity)

	server := NewServer(store, http.NewServeMux(), config, slogger, cache)
	//server := NewServer(store, http.NewServeMux(), config, slogger)
	server.SetupHandlers()

	httpServer := &http.Server{
		Addr:    ":5000",
		Handler: server.Router,
	}

	go func() {
		log.Println("Server Started")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("http server failed : %v", err.Error())
		}
	}()

	// shutdown the server
	defer func() {
		log.Println("Graceful shutdown!")
		ctx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server Shutdown Failed: %+v", err)
		}
	}()

	//time.Sleep(1 * time.Second)
	serverReadyUrl := "http://localhost:5000/test"
	if err := waitForServerToStart(serverReadyUrl, 5*time.Second); err != nil {
		t.Error(err)
		return
	}

	// test POST
	postUrl := "http://localhost:5000/short/post"
	jsonStr := []byte(`{"url":"http://example.com"}`)
	expectedPostResult := `{"long_url":"http://example.com","short_url":"ZxDf"}`

	req, err := http.NewRequest("POST", postUrl, bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(body))
	var expected, actual map[string]string
	if err := json.Unmarshal([]byte(expectedPostResult), &expected); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}

	if err := json.Unmarshal(body, &actual); err != nil {
		t.Fatalf("Failed to unmarshal response JSON: %v", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected: %v received: %v", expected, actual)
	}

	url := "http://localhost:5000/short/get/ZxDf"

	// test GET

	req, err = http.NewRequest("HEAD", url, nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}

	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// prevent redirection
			return http.ErrUseLastResponse
		},
	}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	expectedGet := "http://example.com"
	location := resp.Header.Get("Location")
	if location != expectedGet || resp.Status != "302 Found" {
		t.Errorf("expected: %v received: %v", expected, location)
	}

}
