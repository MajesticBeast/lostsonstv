package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	"github.com/golang-jwt/jwt/v4"
	"github.com/majesticbeast/lostsonstv/internal/muxgo"
	lstv "github.com/majesticbeast/lostsonstv/types"
	"github.com/realTristan/disgoauth"
)

type APIServer struct {
	listenAddr    string
	store         Storage
	muxApiAuth    lstv.MuxApiAuth
	discordClient *disgoauth.Client
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func (s *APIServer) Run() {
	router := chi.NewRouter()

	// JWT stuff
	tokenAuth := jwtauth.New("HS256", []byte(os.Getenv("JWT_SECRET")), nil)

	// Define and create FileServers
	clipsFs := http.FileServer(http.Dir(""))
	templatesFs := http.FileServer(http.Dir("./static"))
	router.Handle("/clips/temp/*", clipsFs)
	router.Handle("/*", templatesFs)

	///////////////////
	// 				 //
	// Public Routes //
	//				 //
	///////////////////

	router.Group(func(r chi.Router) {
		// Homepage
		router.Get("/", makeHTTPHandleFunc(s.handleIndex))
		// List all clips
		router.Get("/clips", makeHTTPHandleFunc(s.handleGetClips))
		// List single clip
		router.Get("/clips/{playbackId}", makeHTTPHandleFunc(s.handleGetClip))
		// Endpoint for Mux video status
		// router.Get("/clips/status", makeHTTPHandleFunc(s.handleMuxStatus))

		// Login
		router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLoginRequest)) // TODO: Fix crash if user goes directly to redirect
		router.HandleFunc("/redirect", makeHTTPHandleFunc(s.handleLoginRedirect))
	})

	//////////////////////
	//					//
	// Protected Routes //
	//					//
	//////////////////////

	router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator)

		// User dashboard
		r.Get("/dashboard", makeHTTPHandleFunc(s.handleDashboard))

		// Create clip
		r.Get("/clips/upload", makeHTTPHandleFunc(s.handleClipSubmission))
		r.Post("/clips/upload", makeHTTPHandleFunc(s.handleClipSubmission))

		// Delete clip
		router.Delete("/clips/{playbackId}", makeHTTPHandleFunc(s.handleDeleteClip))
	})

	// Login

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

func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "invalid token."})
}

func NewAPIServer(listenAddr string, store Storage, muxApiAuth lstv.MuxApiAuth, discordClient *disgoauth.Client) *APIServer {
	return &APIServer{
		listenAddr:    listenAddr,
		store:         store,
		muxApiAuth:    muxApiAuth,
		discordClient: discordClient,
	}
}

func (s *APIServer) handleIndex(w http.ResponseWriter, r *http.Request) error {

	t, err := template.ParseFiles("./static/templates/index.html")
	if err != nil {
		return fmt.Errorf("unable to parse templates: %s", err)
	}

	if err := t.Execute(w, nil); err != nil {
		return fmt.Errorf("unable to execute templates: %s", err)
	}

	return nil
}

func (s *APIServer) handleDashboard(w http.ResponseWriter, r *http.Request) error {
	t, err := template.ParseFiles("./static/templates/dashboard.html")
	if err != nil {
		return fmt.Errorf("unable to parse templates: %s", err)
	}

	cookie, err := r.Cookie("jwt")
	token := cookie.Value
	username, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	claims := username.Claims.(jwt.MapClaims)

	data := lstv.TemplateData{
		Username: claims["Username"].(string),
	}

	if err := t.Execute(w, data); err != nil {
		return fmt.Errorf("unable to execute templates: %s", err)
	}
	return nil
}

func (s *APIServer) handleLoginRequest(w http.ResponseWriter, r *http.Request) error {
	s.discordClient.RedirectHandler(w, r, "")

	return nil
}

func createJWT(user string) (string, error) {
	claims := &jwt.MapClaims{
		"ExpiresAt": 15000,
		"Username":  user,
	}

	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func (s *APIServer) handleLoginRedirect(w http.ResponseWriter, r *http.Request) error {
	code := r.URL.Query()["code"][0]
	accessToken, _ := s.discordClient.GetOnlyAccessToken(code)
	userData, _ := disgoauth.GetUserData(accessToken)
	for k, v := range userData {
		fmt.Printf("Key: %v\nValue: %v\n\n", k, v)
	}
	username := fmt.Sprintf("%v#%v", userData["username"].(string), userData["discriminator"])

	tokenString, err := createJWT(username)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
		// Secure: true,
		Name:  "jwt",
		Value: tokenString,
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)

	return WriteJSON(w, http.StatusOK, "login success")
}

func (s *APIServer) handleGetClip(w http.ResponseWriter, r *http.Request) error {
	// Get a single clip from database
	playbackId := chi.URLParam(r, "playbackId")
	clip, err := s.store.GetClipByPlaybackId(playbackId)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, clip)
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

	// Actual deletion...
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

	newForm := new(lstv.ClipSubmissionForm)
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

	createClipReq := lstv.CreateClipRequest{
		Title:       newForm.Title,
		Description: newForm.Description,
		Game:        newForm.Game,
		Tags:        newForm.Tags,
		Players:     newForm.Players,
		UploadedBy:  newForm.Username,
	}

	// Step 3: POST to Mux.com
	url := "http://thelostsons.net/" + tempFile.Name()
	playbackId, assetId, err := muxgo.PostVideoToMux(url, s.muxApiAuth)
	if err != nil {
		return err
	}

	// Step 4: Enter clip info into database
	clip := lstv.NewClip(playbackId, createClipReq.UploadedBy, createClipReq.Title, createClipReq.Description, createClipReq.Game, createClipReq.Tags, createClipReq.Players, assetId)
	if err := s.store.CreateClip(clip); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, clip)
}
