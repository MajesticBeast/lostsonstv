package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/majesticbeast/lostsonstv/internal/muxgo"
)

type APIServer struct {
	listenAddr string
	store      Storage
	muxApiAuth muxApiAuth
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func (s *APIServer) Run() {
	router := chi.NewRouter()

	// Define and create FileServers
	clipsFs := http.FileServer(http.Dir("."))
	templatesFs := http.FileServer(http.Dir("./static/templates"))
	router.Handle("/clips/temp/*", clipsFs)
	router.Handle("/clips/upload", templatesFs)

	// List all clips
	router.Get("/clips", makeHTTPHandleFunc(s.handleGetClips))

	// Get and delete single clips
	router.Get("/clips/{playbackId}", makeHTTPHandleFunc(s.handleGetClip))
	router.Delete("/clips/{playbackId}", makeHTTPHandleFunc(s.handleDeleteClip))

	// Clip creation
	router.Get("/clips/upload", makeHTTPHandleFunc(s.handleClipSubmission))
	router.Post("/clips/upload", makeHTTPHandleFunc(s.handleClipSubmission))

	log.Println("Lost Sons TV server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func NewAPIServer(listenAddr string, store Storage, muxApiAuth muxApiAuth) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
		muxApiAuth: muxApiAuth,
	}
}

func (s *APIServer) handleGetClip(w http.ResponseWriter, r *http.Request) error {
	// Pull single clip from database - view or delete based on HTTP method

	if r.Method == "GET" {
		playbackId := chi.URLParam(r, "playbackId")
		clip, err := s.store.GetClipByPlaybackId(playbackId)
		if err != nil {
			return err
		}

		return WriteJSON(w, http.StatusOK, clip)
	}

	return fmt.Errorf("method now allowed: %s", r.Method)

}

func (s *APIServer) handleGetClips(w http.ResponseWriter, r *http.Request) error {
	// Pull all clips from database

	clips, err := s.store.GetAllClips()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, clips)
}

func (s *APIServer) handleDeleteClip(w http.ResponseWriter, r *http.Request) error {
	// Hard delete of a single clip

	playbackId := chi.URLParam(r, "playbackId")

	client := muxgo.CreateMuxGoClient(s.muxApiAuth.Id, s.muxApiAuth.Token)

	clip, err := s.store.GetClipByPlaybackId(playbackId)
	if err != nil {
		return err
	}

	muxgo.DeleteAsset(client, clip.AssetId)
	err = s.store.DeleteClip(playbackId)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]string{"deleted": playbackId})
}

func (s *APIServer) handleClipSubmission(w http.ResponseWriter, r *http.Request) error {
	// Serve the HTML upload form or redirect if form is submitted

	// See if user already submitted form
	if r.Method == http.MethodPost {
		return s.handleCreateClip(w, r)
	}

	t, err := template.ParseFiles("./static/templates/uploadClip.html")
	if err != nil {
		return fmt.Errorf("unable to parse templates: %s", err)
	}

	if err := t.Execute(w, nil); err != nil {
		return fmt.Errorf("unable to execute templates: %s", err)
	}

	return nil
}

func (s *APIServer) handleCreateClip(w http.ResponseWriter, r *http.Request) error {
	// Create a single clip

	// Step 1: Receive html web form POST request with user submitted data and video
	r.ParseMultipartForm(40 << 20)

	newForm := new(ClipSubmissionForm)
	newForm.Title = r.FormValue("title")
	newForm.Description = r.FormValue("description")
	newForm.Game = r.FormValue("game")
	newForm.Tags = r.FormValue("tags")
	newForm.Players = r.FormValue("players")
	newForm.Username = r.FormValue("username")

	file, header, err := r.FormFile("videofile")
	if err != nil {
		return fmt.Errorf("error retrieving video clip: %s", err)
	}
	defer file.Close()

	tempFile, err := os.Create("clips/temp/" + header.Filename)
	if err != nil {
		return fmt.Errorf("error creating temp file: %s", err)
	}
	defer tempFile.Close()

	log.Println(tempFile.Name())

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading video clip: %s", err)
	}

	tempFile.Write(fileBytes)

	createClipReq := CreateClipRequest{
		Title:       newForm.Title,
		Description: newForm.Description,
		Game:        newForm.Game,
		Tags:        newForm.Tags,
		Players:     newForm.Players,
		UploadedBy:  newForm.Username,
	}

	// Step 3: POST to Mux.com
	url := "http://thelostsons.net/" + tempFile.Name()
	client := muxgo.CreateMuxGoClient(s.muxApiAuth.Id, s.muxApiAuth.Token)
	asset, err := muxgo.CreateAsset(client, url)
	if err != nil {
		return fmt.Errorf("error sending video to host: %s", err)
	}

	createClipReq.PlaybackId = asset.Data.PlaybackIds[0].Id
	createClipReq.AssetId = asset.Data.Id

	// Step 4: Enter clip info into database
	clip := NewClip(createClipReq.PlaybackId, createClipReq.UploadedBy, createClipReq.Title, createClipReq.Description, createClipReq.Game, createClipReq.Tags, createClipReq.Players, createClipReq.AssetId)
	if err := s.store.CreateClip(clip); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, clip)
}
