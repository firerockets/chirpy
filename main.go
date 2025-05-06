package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

const (
	filepathRoot   = "./"
	filepathAssets = "./assets"
	port           = "8080"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	cfg.fileserverHits.Add(1)
	return next
}

func main() {

	apiCfg := apiConfig{}
	mux := http.NewServeMux()

	mux.HandleFunc("/app/", apiCfg.appHandler)
	mux.HandleFunc("GET /api/healthz", apiCfg.healthzHandler)
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.validateChirpHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}

func (apiCfg *apiConfig) appHandler(w http.ResponseWriter, req *http.Request) {
	fileServerHandler := http.FileServer(http.Dir(filepathRoot))
	apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServerHandler)).ServeHTTP(w, req)
}
