package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
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

	const maxMemory = 10 << 20

	//dunno why lesson doesnt error check the following, but ill follow suit for now
	r.ParseMultipartForm(maxMemory)
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to parse multipart form", err)
		return
	}
	defer file.Close()
	mediaType := header.Header.Get("Content-Type")
	if mediaType == "" {
		respondWithError(w, http.StatusBadRequest, "missing content-type header", nil)
		return
	}
	imgData, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to read file", err)
		return
	}


	//1.7 start
	parsedType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to parse content type for file extension", err)
		return
	}

	ext, err := mime.ExtensionsByType(parsedType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to obtain file ext", err)
	}

	thumbFilePath := filepath.Join(cfg.assetsRoot, videoID.String(), ext[0])
	if thumbFilePath == "" {
		respondWithError(w, http.StatusInternalServerError, "error creating thumbnail path", nil)
		return
	}

	thumbFile, err := os.Create(thumbFilePath)
	if thumbFilePath == "" {
		respondWithError(w, http.StatusInternalServerError, "error creating thumbnail file", nil)
		return
	}

	//copy file
	_, err = io.Copy(thumbFile, file)




	//1.7 end

	videoMetadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldnt find videot", err)
		return
	}
	//unauthorised access if logged in user is not owner of the video
	if videoMetadata.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "unauthorised access. current user is not user that owns video", nil)
		return
	}

	//update video url
	videoMetadata.ThumbnailURL = &dataURL64

	if err = cfg.db.UpdateVideo(videoMetadata); err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to update video in database", err)
		return
	}
	respondWithJSON(w, http.StatusOK, videoMetadata)
}
