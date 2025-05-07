package main

import (
	"fmt"
	"log"
	"net/http"
)

func (apiCfg *apiConfig) metricsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	body := fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
	`, apiCfg.fileserverHits.Load())

	w.Write([]byte(body))
}

func (apiCfg *apiConfig) resetHandler(w http.ResponseWriter, req *http.Request) {

	if apiCfg.platform != "dev" {
		respondWithError(w, "This operation is forbiden", http.StatusForbidden)
		log.Println("Tried to run reset from non-dev env.")
		return
	}

	apiCfg.dbQueries.DeleteAllUsers(req.Context())

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	apiCfg.fileserverHits.Store(0)
}
