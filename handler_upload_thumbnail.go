package main

import (
	"encoding/base64"
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

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	// "thumbnail" should match the HTML form input name
	file, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	fileMediaType := fileHeader.Header.Get("Content-Type")
	if fileMediaType == "" {
		respondWithError(w, http.StatusBadRequest, "Unable to get form file media type", err)
		return
	}

	// `file` is an `io.Reader` that we can read from to get the image data
	fileData, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid file", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something is wrong", err)
		return
	}

	if userID != video.UserID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	fileDataStr := base64.StdEncoding.EncodeToString(fileData)
	newThumbnailDataURL := fmt.Sprintf("data:%s;base64,%s", fileMediaType, fileDataStr)
	video.ThumbnailURL = &newThumbnailDataURL

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something is wrong", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
