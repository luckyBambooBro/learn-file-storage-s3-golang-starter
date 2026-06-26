package main

import (
	"net/http"

	"internal/auth"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	//Set an upload limit of 1 GB (1 << 30 bytes) using http.MaxBytesReader.
	r.Body = http.MaxBytesReader(w, r.Body, 10 << 30)
	
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID) 
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to retrieve video metadata", err)
	}
	if userID != video.UserID {
		respondWithError(w, http.StatusUnauthorized, "unauthorised user access for requested video", nil)
	}

	const maxMemory = 32 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.Formfile()

	


}
