package main

import (
	"time"
)

type ClipSubmissionForm struct {
	Title       string
	Description string
	Game        string
	Tags        string
	Players     string
	Username    string
}

type CreateClipRequest struct {
	PlaybackId  string `json:"playbackId"`
	UploadedBy  string `json:"uploadedBy"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Game        string `json:"game"`
	Tags        string `json:"tags"`
	Players     string `json:"players"`
}

type Clip struct {
	PlaybackId   string    `json:"playbackId"`
	UploadedBy   string    `json:"uploadedBy"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Game         string    `json:"game"`
	Tags         string    `json:"tags"`
	Players      string    `json:"players"`
	DateUploaded time.Time `json:"dateUploaded"`
}

func NewClip(playbackId, uploadedBy, title, description, game, tags, players string) *Clip {
	return &Clip{
		PlaybackId:   playbackId,
		UploadedBy:   uploadedBy,
		Title:        title,
		Description:  description,
		Game:         game,
		Tags:         tags,
		Players:      players,
		DateUploaded: time.Now(),
	}
}
