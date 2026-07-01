package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
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

	//create temp file and copy uploaded file over
	tempFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "internal error, could not create temporary file", err)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err = io.Copy(tempFile, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to copy file", err)
		return
	}
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not reset file pointer", err)
		return
	}
	//put file into s3
	assetPath := getAssetPath(mediatype)
	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket: aws.String(cfg.s3Bucket),
		Key: aws.String(assetPath),
		Body: tempFile,
		ContentType: aws.String(mediatype),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "aws PutObject error", err)
		return
	}

	videoURL := cfg.getObjectURL(assetPath)
	video.VideoURL = &videoURL

	if err = cfg.db.UpdateVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to update video URL", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)

}

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v",
		"error",
		"-print_format",
		"json",
		"-show_streams",
		filePath)

	buf := &bytes.Buffer{}
	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		return "", err
	}
	stdoutBytes := buf.Bytes()

	v := &FFProbeData{}
	if err := json.Unmarshal(stdoutBytes, v); err != nil {
		return "", err
	}
}

/*

Create a function getVideoAspectRatio(filePath string) (string, error) that takes a file path and returns the aspect ratio as a string.

It should use exec.Command to run the same ffprobe command as above. In this case, the command is ffprobe and the arguments are -v, error, -print_format, json, -show_streams, and the file path.

Set the resulting exec.Cmd's Stdout field to a pointer to a new bytes.Buffer.
.Run() the command

Unmarshal the stdout of the command from the buffer's .Bytes into a JSON struct so that you can get the width and height fields.
I did a bit of math to determine the ratio, then returned one of three strings: 16:9, 9:16, or other.

*/