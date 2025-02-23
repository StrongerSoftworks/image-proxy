package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/StrongerSoftworks/image-proxy/internal/handlers"
	"github.com/joho/godotenv"
)

func healthCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/status" {
			w.WriteHeader(http.StatusOK)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func requestFilterMiddleware(next http.Handler, allowedURLs []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		referer := r.Header.Get("Referer")

		urls := sliceToMap(allowedURLs)

		if origin != "" && urls[origin] {
			next.ServeHTTP(w, r)
			return
		}
		if referer != "" && urls[referer] {
			next.ServeHTTP(w, r)
			return
		}

		// Block request if no valid Origin or Referer
		http.Error(w, "Forbidden", http.StatusForbidden)
	})
}

func sliceToMap(slice []string) map[string]bool {
	m := make(map[string]bool)
	for _, v := range slice {
		m[v] = true
	}
	return m
}

func main() {
	// Determine which environment file to load
	env := os.Getenv("GO_ENV") // Set this to "production", "development" or "local"
	envFile := ".env"          // Default to local
	if env == "production" {
		envFile = ".env.production"
	} else if env == "development" {
		envFile = ".env.development"
	}

	// Load the chosen .env file
	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("Error loading %s file: %v", envFile, err)
	}

	// Read environment variables
	allowedURLs := os.Getenv("ALLOWED_ORIGINS")
	storageMode := os.Getenv("STORAGE_MODE")

	var requestHandler handlers.ImageProxyRequestHandler
	if storageMode == "s3" {
		requestHandler = handlers.NewS3RequestHanlder()
	} else {
		requestHandler = handlers.NewLocalRequestHandler()
	}

	requestHandler.Init()

	log.Println("Allowed domains: " + allowedURLs)
	log.Println("Server is running on port 8080...")

	mux := http.NewServeMux()
	mux.HandleFunc("/proxy", requestHandler.Handler)
	http.ListenAndServe(":8080", healthCheckMiddleware(requestFilterMiddleware(mux, strings.Split(allowedURLs, ","))))
}
