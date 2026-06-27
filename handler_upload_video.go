package main

import (
	"mime"
	"net/http"

	"internal/auth"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	//Set an upload limit of 1 GB (1 << 30 bytes) using http.MaxBytesReader.
	r.Body = http.MaxBytesReader(w, r.Body, 1 << 30)
	
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
		return
	}
	if userID != video.UserID {
		respondWithError(w, http.StatusUnauthorized, "unauthorised user access for requested video", nil)
		return
	}

	const maxMemory = 32 << 20
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to parse form", err)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to parse form", err)
		return
	}
	defer file.Close()
	
	mediatype, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to parse media type", err)
		return
	}
	if mediatype != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "must upload video/mp4 file", nil)
		return
	}



}
