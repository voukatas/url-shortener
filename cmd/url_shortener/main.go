package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/voukatas/url-shortener/internal/config"
	"github.com/voukatas/url-shortener/internal/server"
	"github.com/voukatas/url-shortener/internal/store"
	"github.com/voukatas/url-shortener/internal/url_converter"
	"github.com/voukatas/url-shortener/pkg/cache"
	"github.com/voukatas/url-shortener/pkg/logger"
)

func main() {
	// conf
	config, err := config.LoadConfig("short_conf.json")
	if err != nil {
		fmt.Printf("Error loading config: %v", err)
		return
	}

	fmt.Printf("Loaded config: %+v\n", config)

	// init obfuscation
	url_converter.InitBase62Array(config.ShuffleKey)

	// logging
	slogger, cleanup := logger.SetupLogger(config.LogFilename, config.LogLevel, config.Production)
	defer cleanup()

	// store
	store, err := store.NewStore(config.DBFilename + "?mode=rwc")
	if err != nil {
		slogger.Error("Store init", "error", err.Error())
		return
	}
	defer func() {
		slogger.Error("DB Closed properly")
		store.Close()
	}()
	store.SetStoreOptions()

	// cache
	cache := cache.NewCache(config.CacheCapacity)

	server := server.NewServer(store, http.NewServeMux(), config, slogger, cache)
	server.SetupHandlers()

	httpServer := &http.Server{
		Addr:    config.Address,
		Handler: server.Router,
	}

	// Catch interruptions like ctrl-c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		server.Logger.Error("Server Started")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			server.Logger.Error("http server failed", "error", err.Error())
		}
	}()

	// wait for a shutdown signal
	<-signalChan
	// gracefull shutdown
	server.Logger.Error("Gracefull shutdown!")

	ctx := context.Background()
	shutdownCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		server.Logger.Error("Server Shutdown Failed", "error", err)
	}

	server.Logger.Error("Server exited normally")
}
