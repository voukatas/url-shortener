package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/voukatas/url-shortener/internal/model"
	"github.com/voukatas/url-shortener/internal/store"
	"github.com/voukatas/url-shortener/internal/url_converter"
	"github.com/voukatas/url-shortener/pkg/cache"
	"github.com/voukatas/url-shortener/pkg/logger"
)

type URLShortener struct {
	Store  store.Store
	Router *http.ServeMux
	Config *model.Config
	Logger logger.Logger
	Cache  cache.Cache
}

func NewServer(store store.Store, router *http.ServeMux, config *model.Config, logger logger.Logger, cache cache.Cache) *URLShortener {
	return &URLShortener{
		Store:  store,
		Router: router,
		Config: config,
		Logger: logger,
		Cache:  cache,
	}
}

func (server *URLShortener) SetupHandlers() {
	server.Router.HandleFunc("GET /short/get/{url}", server.RedirectURL)
	server.Router.HandleFunc("POST /short/post", server.CreateShortURL)
}

func (server *URLShortener) RedirectURL(w http.ResponseWriter, r *http.Request) {

	shortUrl := r.PathValue("url")
	server.Logger.Debug("RedirectURL", "url", shortUrl, "address", server.getClientIP(r))

	if shortUrl == "" {
		http.Error(w, "Bad Request: Missing or invalid URL", http.StatusBadRequest)
		server.Logger.Warn("RedirectURL Bad Request: Missing or invalid URL")
		return
	}

	// retrieve value from cache
	if url, err := server.Cache.Get(shortUrl); err == nil {
		server.Logger.Info("RedirectURL - Cache Get found", "url", url)
		http.Redirect(w, r, url, http.StatusFound)
		return

	}

	decodedID := url_converter.DecodeShortCode(shortUrl, server.Config.XorSecretKey)
	server.Logger.Info("RedirectURL", "Decoded ID", decodedID)

	longUrl, err := server.Store.Lookup(decodedID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		server.Logger.Error("DecodeShortCode", "error", err)
		return
	}

	// store it in cache
	server.Cache.Set(shortUrl, longUrl)
	server.Logger.Info("RedirectURL - Cache Set triggered", "shortUrl", shortUrl, "longUrl", longUrl)

	http.Redirect(w, r, longUrl, http.StatusFound)
}

func (server *URLShortener) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var url model.Url
	if err := json.NewDecoder(r.Body).Decode(&url); err != nil {
		server.Logger.Error("Failed to decode JSON", "error", err, "address", server.getClientIP(r))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if !strings.HasPrefix(url.Url, "http://") && !strings.HasPrefix(url.Url, "https://") {
		server.Logger.Error("Missing Protocol", "address", server.getClientIP(r))
		http.Error(w, "Protocol Missing", http.StatusBadRequest)
		return

	}

	id, err := server.Store.Shorten(url.Url)
	if err != nil {
		server.Logger.Error("CreateShortURL", "error", err)
		return
	}

	// Encode the ID
	shortCode := url_converter.EncodeID(id, server.Config.XorSecretKey)
	server.Logger.Debug("CreateShortURL", "Original ID", id, "Long URL", url.Url, "Short Code", shortCode, "address", server.getClientIP(r))

	response := model.ShortUrlResponse{LongUrl: url.Url, ShortUrl: shortCode}

	// store it in cache
	//server.Cache.Set(shortCode, response.LongUrl)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201
	if err := json.NewEncoder(w).Encode(response); err != nil {
		server.Logger.Error("Failed to encode response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

}

func (server *URLShortener) getClientIP(r *http.Request) string {
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		server.Logger.Debug("getClientIP", "X-Real-IP", realIP)
		return realIP
	}

	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// the first ip from the list is the client ip, the rest are proxies etc..
		server.Logger.Debug("getClientIP", "Full X-Forwarded-For", forwarded)
		return strings.Split(forwarded, ",")[0]
	}

	// Fallback to RemoteAddr if no headers are set, this will return the localhost ip if used in conjuction with a reverse proxy
	server.Logger.Debug("getClientIP fallback to RemoteAddr")
	return r.RemoteAddr
}
