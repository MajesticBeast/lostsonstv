package muxgo

import "time"

type VideoAsset struct {
	Type      string `json:"type"`
	RequestID any    `json:"request_id"`
	Object    struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	} `json:"object"`
	ID          string `json:"id"`
	Environment struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"environment"`
	Data struct {
		UploadID string `json:"upload_id"`
		Tracks   []struct {
			Type             string `json:"type"`
			MaxWidth         int    `json:"max_width,omitempty"`
			MaxHeight        int    `json:"max_height,omitempty"`
			MaxFrameRate     int    `json:"max_frame_rate,omitempty"`
			ID               string `json:"id"`
			MaxChannels      int    `json:"max_channels,omitempty"`
			MaxChannelLayout string `json:"max_channel_layout,omitempty"`
		} `json:"tracks"`
		Test        bool   `json:"test"`
		Status      string `json:"status"`
		PlaybackIds []struct {
			Policy string `json:"policy"`
			ID     string `json:"id"`
		} `json:"playback_ids"`
		Mp4Support          string  `json:"mp4_support"`
		MaxStoredResolution string  `json:"max_stored_resolution"`
		MaxStoredFrameRate  int     `json:"max_stored_frame_rate"`
		MasterAccess        string  `json:"master_access"`
		ID                  string  `json:"id"`
		Duration            float64 `json:"duration"`
		CreatedAt           int     `json:"created_at"`
		AspectRatio         string  `json:"aspect_ratio"`
	} `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	Attempts  []struct {
		WebhookID          int `json:"webhook_id"`
		ResponseStatusCode int `json:"response_status_code"`
		ResponseHeaders    struct {
			XCacheLookup  string `json:"x-cache-lookup"`
			XCache        string `json:"x-cache"`
			Date          string `json:"date"`
			ContentLength string `json:"content-length"`
			Connection    string `json:"connection"`
		} `json:"response_headers"`
		ResponseBody any       `json:"response_body"`
		MaxAttempts  int       `json:"max_attempts"`
		ID           string    `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		Address      string    `json:"address"`
	} `json:"attempts"`
	AccessorSource any `json:"accessor_source"`
	Accessor       any `json:"accessor"`
}
