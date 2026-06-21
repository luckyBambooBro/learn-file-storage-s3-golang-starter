package main

import (
	"io"
	"net/http"

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

	//1.7 start
	ext, err := getFileExtension(mediaType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "", err)
		return
	}
	if ext != ".jpeg" && ext != ".jpg" && ext != ".png" {
		respondWithError(w, http.StatusBadRequest, "upload needs to be an image", nil)
		return
	}

	thumbFilePath, thumbFile, err := cfg.createAssetFile(ext)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to create thumbnail folder/file", err)
		return
	}
	defer thumbFile.Close()
	//copy file
	_, err = io.Copy(thumbFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error copying thumbnail file", nil)
		return
	}

	//1.7 checkpoint

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldnt find video", err)
		return
	}
	//unauthorised access if logged in user is not owner of the video
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "unauthorised access. current user is not user that owns video", nil)
		return
	}

	//update video url
	url := cfg.getAssetURL(thumbFilePath)
	video.ThumbnailURL = &url

	if err = cfg.db.UpdateVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to update video in database", err)
		return
	}
	respondWithJSON(w, http.StatusOK, video)
}
