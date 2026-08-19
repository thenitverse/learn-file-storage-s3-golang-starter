package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse multipart form", err)
		return
	}
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()
	value := header.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(value)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse mediaType", err)
		return
	}
	if mediatype != "image/jpeg" && mediatype != "image/png" {
		respondWithError(w, http.StatusBadRequest, "not a correct media type", nil)
		return
	}

	// Suppose mediatype is "image/png"
	parts := strings.Split(mediatype, "/")
	// parts becomes []string{"image", "png"}

	// The second element (parts[1]) is the extension name: "png"
	ext := "." + parts[1]
	// ext becomes ".png"
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get a video", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unable to find user with this id ", nil)
		return
	}
	fileName := fmt.Sprintf("%s%s", videoID, ext)
	filePath := filepath.Join(cfg.assetsRoot, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to find file ", err)
		return
		// handle error
	}
	defer dst.Close()
	_, err = io.Copy(dst, file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to copy file", err)
		return
	}
	url := fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, fileName)
	video.ThumbnailURL = &url
	// it will be something like data:image/png;base64,iVBORW0KGgoAAAAn...

	err = cfg.db.UpdateVideo(video) //saved the dataurl string as a video.thumbnailurl in data base
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get URL", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
