package main

import (
	"fmt"
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


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20

	//dunno why lesson doesnt error check the following, but ill follow suit for now
	r.ParseMultipartForm(maxMemory)
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to parse multipart form", err)
	}
	defer file.Close()
	mediaType := header.Header.Get("Content-Type")
	imgData, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to read file", err)
		return
	}
	videoMetadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "video with provided uuid does not exist", err)
		return
	}
	//unauthorised access if logged in user is not owner of the video
	if videoMetadata.CreateVideoParams.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "unauthorised access. current user is not user that owns video", nil)
		return
	}

	videoThumbnail := thumbnail{
		data: imgData,
		mediaType: mediaType,
	}

	videoThumbnails[videoID] = videoThumbnail

	//update video url
	thumbnailURL := "http://localhost:" + cfg.port + "/api/thumbnails/" + videoID.String()
	videoMetadata.ThumbnailURL = &thumbnailURL
	if err = cfg.db.UpdateVideo(videoMetadata); err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to update database with video metadata", err)
		return
	}


	respondWithJSON(w, http.StatusOK, videoMetadata)
}
