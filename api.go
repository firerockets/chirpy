package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/firerockets/chirpy/internal/database"
	"github.com/google/uuid"
)

func (apiCfg *apiConfig) healthzHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (apiCfg *apiConfig) createUserHandler(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	var params parameters
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, "Something went wrong while parsing the request body", http.StatusInternalServerError)
		log.Printf("Error decoding json request: %s\n", err)
		return
	}

	usr, err := apiCfg.dbQueries.CreateUser(req.Context(), params.Email)

	if err != nil {
		respondWithError(w, "Something went wrong while creating the user in the db", http.StatusInternalServerError)
		log.Printf("Error creating user: %s\n", err)
		return
	}

	type userResponse struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	respondWithJSON(w, http.StatusCreated, userResponse{
		ID:        usr.ID.String(),
		CreatedAt: usr.CreatedAt,
		UpdatedAt: usr.UpdatedAt,
		Email:     usr.Email,
	})

	log.Println("User created sucessfully.")
}

func (apiCfg *apiConfig) createChirpHandler(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error decoding json request: %s\n", err)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, "Chirp is too long", http.StatusBadRequest)
		log.Println("Message too long received")
		return
	}

	log.Println("Valid message received")

	chirp, err := apiCfg.dbQueries.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: params.UserId,
	})

	if err != nil {
		respondWithError(w, "Error creating Chirp", http.StatusInternalServerError)
		log.Printf("Error inserting chirp into the database: %s\n", err)
		return
	}

	type responseJson struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    string    `json:"user_id"`
	}

	respondWithJSON(w, http.StatusCreated, responseJson{
		ID:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID.String(),
	})

	log.Println("Chirp created in the database")
}

func cleanBody(body string) string {
	splitted := strings.Split(body, " ")

	for i, word := range splitted {
		if isCurseWord(strings.ToLower(word)) {
			splitted[i] = "****"
		}
	}

	return strings.Join(splitted, " ")
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
