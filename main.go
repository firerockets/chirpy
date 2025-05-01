package main

import (
	"log"
	"net/http"
)

const (
	filepathRoot   = "./"
	filepathAssets = "./assets"
	port           = "8080"
)

func main() {

	mux := http.NewServeMux()

	// mux.Handle("/assets", http.FileServer(http.Dir(filepathAssets)))

	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot))))
	// mux.Handle("/app/assets", http.StripPrefix("/app", http.FileServer(http.Dir(filepathAssets))))
	mux.HandleFunc("/healthz", healthzHandler)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}

func healthzHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
