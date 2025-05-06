package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func (apiCfg *apiConfig) healthzHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (apiCfg *apiConfig) validateChirpHandler(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding json request: %s", err)
		w.WriteHeader(500)
		w.Write([]byte(`{"error": "Something went wrong"}`))
		return
	}

	if len(params.Body) > 140 {
		log.Println("Message too long received")
		w.WriteHeader(400)
		w.Write([]byte(`{"error": "Chirp is too long"}`))
		return
	}

	splitted := strings.Split(params.Body, " ")

	for i, word := range splitted {
		if isCurseWord(strings.ToLower(word)) {
			splitted[i] = "****"
		}
	}

	joined := strings.Join(splitted, " ")

	responseJson := fmt.Sprintf(`{"cleaned_body": "%s"}`, joined)

	log.Println("Valid message received")
	w.WriteHeader(200)
	w.Write([]byte(responseJson))

}

func isCurseWord(word string) bool {
	switch word {
	case
		"kerfuffle",
		"sharbert",
		"fornax":
		return true
	}
	return false
}
