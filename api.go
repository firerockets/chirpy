package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/firerockets/chirpy/internal/auth"
	"github.com/firerockets/chirpy/internal/database"
	"github.com/google/uuid"
)

func (apiCfg *apiConfig) healthzHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (apiCfg *apiConfig) createUserHandler(w http.ResponseWriter, req *http.Request) {
	var params userRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, "Something went wrong while parsing the request body", http.StatusInternalServerError)
		log.Printf("Error decoding json request: %s\n", err)
		return
	}

	hashedPass, err := auth.HashPassword(params.Password)

	if err != nil {
		respondWithError(w, "Something went wrong while setting the password", http.StatusInternalServerError)
		log.Printf("Error hashing password: %s\n", err)
		return
	}

	usr, err := apiCfg.dbQueries.CreateUser(req.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPass,
	})

	if err != nil {
		respondWithError(w, "Something went wrong while creating the user in the db", http.StatusInternalServerError)
		log.Printf("Error creating user: %s\n", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, userResponse{
		ID:        usr.ID.String(),
		CreatedAt: usr.CreatedAt,
		UpdatedAt: usr.UpdatedAt,
		Email:     usr.Email,
	})

	log.Println("User created sucessfully.")
}

func (apiCfg *apiConfig) updateUserHandler(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)

	if err != nil {
		respondWithError(w, "Something went wrong", http.StatusUnauthorized)
		log.Printf("Error loading token from header: %s\n", err)
		return
	}

	usrID, err := auth.ValidadeJWT(token, apiCfg.secret)

	if err != nil {
		respondWithError(w, "Unauthorized", http.StatusUnauthorized)
		log.Printf("Error validating token: %s\n", err)
		return
	}

	type userRequest struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	var params userRequest

	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&params)

	if err != nil {
		respondWithError(w, "Something went wrong while parsing the request body", http.StatusInternalServerError)
		log.Printf("Error decoding json request: %s\n", err)
		return
	}

	hashedPass, err := auth.HashPassword(params.Password)

	if err != nil {
		respondWithError(w, "Something went wrong while setting the password", http.StatusInternalServerError)
		log.Printf("Error hashing password: %s\n", err)
		return
	}

	usr, err := apiCfg.dbQueries.UpdateUserForId(req.Context(), database.UpdateUserForIdParams{
		ID:             usrID,
		Email:          params.Email,
		HashedPassword: hashedPass,
	})

	if err != nil {
		respondWithError(w, "Unauthorized", http.StatusUnauthorized)
		log.Printf("No user found for id: %s\n", err)
		return
	}

	respondWithJSON(w, http.StatusOK, userResponse{
		ID:        usr.ID.String(),
		CreatedAt: usr.CreatedAt,
		UpdatedAt: usr.UpdatedAt,
		Email:     usr.Email,
	})

	log.Println("User updated sucessfully.")

}

func (apiCfg *apiConfig) loginHandler(w http.ResponseWriter, req *http.Request) {
	type loginRequest struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	var params loginRequest

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, "Something went wrong while parsing the request body", http.StatusInternalServerError)
		log.Printf("Error decoding json request: %s\n", err)
		return
	}

	user, err := apiCfg.dbQueries.GetUserByEmail(req.Context(), params.Email)

	if err != nil {
		respondWithError(w, "Invalid user credentials", http.StatusInternalServerError)
		log.Printf("No user found for the email - %s: %s\n", params.Email, err)
		return
	}

	err = auth.CheckPasswordHash(user.HashedPassword, params.Password)

	if err != nil {
		respondWithError(w, "Invalid user password", http.StatusUnauthorized)
		log.Printf("Password doesn't match with hashed value: %s\n", err)
		return
	}

	jwt, err := auth.MakeJWT(user.ID, apiCfg.secret, time.Duration(1)*time.Hour)

	if err != nil {
		respondWithError(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error generating token: %s\n", err)
		return
	}

	refreshTokenString, err := auth.MakeRefreshToken()

	if err != nil {
		respondWithError(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error generating refresh token: %s\n", err)
		return
	}

	_, err = apiCfg.dbQueries.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token:     refreshTokenString,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Duration(60*24) * time.Hour),
		RevokedAt: sql.NullTime{Valid: false},
	})

	if err != nil {
		respondWithError(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error creating refresh token in the database: %s\n", err)
		return
	}

	type loginResponse struct {
		ID           string    `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	respondWithJSON(w, http.StatusOK, loginResponse{
		ID:           user.ID.String(),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        jwt,
		RefreshToken: refreshTokenString,
	})
}

func (apiCfg *apiConfig) refreshTokenHandler(w http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)

	if err != nil {
		respondWithError(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error loading token from header: %s\n", err)
		return
	}

	tokenObj, err := apiCfg.dbQueries.GetRefreshTokenByToken(req.Context(), refreshToken)

	if err != nil {
		respondWithError(w, "Unauthorized", http.StatusUnauthorized)
		log.Printf("Error looking up refresh token: %s\n", err)
		return
	}

	if tokenObj.RevokedAt.Valid {
		respondWithError(w, "Unauthorized - token revoked", http.StatusUnauthorized)
		return
	}

	if time.Now().After(tokenObj.ExpiresAt) {
		respondWithError(w, "Unauthorized - token expired", http.StatusUnauthorized)
		return
	}

	usr, err := apiCfg.dbQueries.GetUserById(req.Context(), tokenObj.UserID)

	if err != nil {
		respondWithError(w, "Unauthorized", http.StatusUnauthorized)
		log.Printf("Error looking up user: %s\n", err)
		return
	}

	jwt, err := auth.MakeJWT(usr.ID, apiCfg.secret, time.Duration(1)*time.Hour)

	if err != nil {
		respondWithError(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error generating token: %s\n", err)
		return
	}

	type responseJSON struct {
		Token string `json:"token"`
	}

	respondWithJSON(w, http.StatusOK, responseJSON{
		Token: jwt,
	})
}

func (apiCfg *apiConfig) revokeRefreshTokenHandler(w http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)

	if err != nil {
		respondWithError(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error loading token from header: %s\n", err)
		return
	}

	err = apiCfg.dbQueries.RevokeRefreshToken(req.Context(), refreshToken)

	if err != nil {
		respondWithError(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error error updating data base with revoked token: %s\n", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (apiCfg *apiConfig) createChirpHandler(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)

	if err != nil {
		respondWithError(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error loading token from header: %s\n", err)
		return
	}

	usrID, err := auth.ValidadeJWT(token, apiCfg.secret)

	if err != nil {
		respondWithError(w, "Unauthorized", http.StatusUnauthorized)
		log.Printf("Error validating token: %s\n", err)
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}

	err = decoder.Decode(&params)
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
		UserID: usrID,
	})

	if err != nil {
		respondWithError(w, "Error creating Chirp", http.StatusInternalServerError)
		log.Printf("Error inserting chirp into the database: %s\n", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, chirpResponse{
		ID:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID.String(),
	})

	log.Println("Chirp created in the database")
}

func (apiCfg *apiConfig) getChirpsHandler(w http.ResponseWriter, req *http.Request) {
	chirps, err := apiCfg.dbQueries.GetChirps(req.Context())

	if err != nil {
		respondWithError(w, "Error getting chirps", http.StatusInternalServerError)
		log.Printf("Error fetching chirps from database: %s\n", err)
		return
	}

	chirpsResponse := []chirpResponse{}

	for _, c := range chirps {
		chirpsResponse = append(chirpsResponse, chirpResponse{
			ID:        c.ID.String(),
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID.String(),
		})
	}

	respondWithJSON(w, http.StatusOK, chirpsResponse)
}

func (apiCfg *apiConfig) getChirpByIdHandler(w http.ResponseWriter, req *http.Request) {
	chirpID, err := uuid.Parse(req.PathValue("chirpID"))

	if err != nil {
		respondWithError(w, "Invalid ID", http.StatusInternalServerError)
		log.Printf("Error validating UUID: %s\n", err)
		return
	}

	chirp, err := apiCfg.dbQueries.GetChirpById(req.Context(), chirpID)

	if err != nil {
		respondWithError(w, "Error getting chirp from the database", http.StatusNotFound)
		log.Printf("Chirp id not found: %s\n", err)
		return
	}

	respondWithJSON(w, http.StatusOK, chirpResponse{
		ID:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID.String(),
	})

	log.Printf("Successfuly returned chirp object")
}

func (apiCfg *apiConfig) deleteChirpByIdHandler(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)

	if err != nil {
		respondWithError(w, "Something went wrong", http.StatusUnauthorized)
		log.Printf("Error loading token from header: %s\n", err)
		return
	}

	usrID, err := auth.ValidadeJWT(token, apiCfg.secret)

	if err != nil {
		respondWithError(w, "Unauthorized", http.StatusUnauthorized)
		log.Printf("Error validating token: %s\n", err)
		return
	}

	chirpID, err := uuid.Parse(req.PathValue("chirpID"))

	if err != nil {
		respondWithError(w, "Invalid ID", http.StatusInternalServerError)
		log.Printf("Error validating UUID: %s\n", err)
		return
	}

	chirp, err := apiCfg.dbQueries.GetChirpById(req.Context(), chirpID)

	if err != nil {
		respondWithError(w, "Error getting chirp from the database", http.StatusNotFound)
		log.Printf("Chirp id not found: %s\n", err)
		return
	}

	if chirp.UserID != usrID {
		respondWithError(w, "Forbiden", http.StatusForbidden)
		log.Println("Chirp does not belong to this user")
		return
	}

	err = apiCfg.dbQueries.DeleteChirpById(req.Context(), chirpID)

	if err != nil {
		respondWithError(w, "Error getting chirp from the database", http.StatusNotFound)
		log.Printf("Error deleting chirp: %s\n", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type chirpResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    string    `json:"user_id"`
}

type userRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type userResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
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
