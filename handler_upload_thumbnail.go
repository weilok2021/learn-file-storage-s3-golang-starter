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
	const maxMemory int64 = 10 * 1024 * 1024
	r.ParseMultipartForm(maxMemory)
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse thumbnail form file", err)
		return
	}
	defer file.Close()
	mediaType := header.Header.Get("Content-Type")
	imgData, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to read image data in file", err)
		return
	}

	metadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to read video meta data from db", err)
		return
	}

	// check video ownership
	if userID != metadata.UserID {
		respondWithError(w, http.StatusUnauthorized, "Authorization Error: User does not own this video", err)
	}
	thumbnail := thumbnail{
		data:      imgData,
		mediaType: mediaType,
	}
	// Add the thumbnail to the global map, using the video's ID as the key
	videoThumbnails[videoID] = thumbnail
	url := fmt.Sprintf("http://localhost:8091/api/thumbnails/%s", videoIDString)
	metadata.ThumbnailURL = &url
	err = cfg.db.UpdateVideo(metadata)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to update video's thumbnail", err)
	}
	respondWithJSON(w, http.StatusOK, metadata)
}
