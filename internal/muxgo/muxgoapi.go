package muxgo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	db "github.com/jackc/pgx/v5"
	lstv "github.com/majesticbeast/lostsonstv/types"
	muxgo "github.com/muxinc/mux-go"
)

func CreateMuxGoClient(MUX_TOKEN_ID, MUX_TOKEN_SECRET string) *muxgo.APIClient {
	client := muxgo.NewAPIClient(
		muxgo.NewConfiguration(
			muxgo.WithBasicAuth(MUX_TOKEN_ID, MUX_TOKEN_SECRET),
		))

	return client
}

func CreateAsset(client *muxgo.APIClient, url string) (muxgo.AssetResponse, error) {
	asset, err := client.AssetsApi.CreateAsset(muxgo.CreateAssetRequest{
		Input: []muxgo.InputSettings{
			muxgo.InputSettings{
				Url: url,
			},
		},
		PlaybackPolicy: []muxgo.PlaybackPolicy{
			"public",
		},
	})

	return asset, err
}

func DeleteAsset(client *muxgo.APIClient, assetId string) error {
	err := client.AssetsApi.DeleteAsset(assetId)
	if err != nil {
		return err
	}
	_, err = client.AssetsApi.GetAsset(assetId)

	return err
}

func ReceiveVideoStatus(w http.ResponseWriter, r *http.Request) {

	// Only POST allowed
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(fmt.Sprint(http.StatusNotImplemented) + ": Method not allowed."))
		return
	}

	// Get the hook data
	jsonResult, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}

	var hook VideoAsset
	json.Unmarshal([]byte(jsonResult), &hook)

	log.Printf("Incoming webhook from: %s\n", r.RemoteAddr)

	if hook.Type == "video.asset.ready" {
		log.Println(hook.Type)
		log.Println("Upload ID: ", hook.Data.UploadID)

		dbString := os.Getenv("DB_CONN_STR")

		conn, err := db.Connect(context.Background(), dbString)
		if err != nil {
			fmt.Println("unable to connecto to db: ", err)
		}

		defer conn.Close(context.Background())

		_, err = conn.Exec(context.Background(), "update thelostsons_clips SET playback_id = $1 WHERE unique_id = $2", hook.Data.PlaybackIds[0].ID, hook.Data.UploadID)
	}
}

func PostVideoToMux(url string, muxAuth lstv.MuxApiAuth) (string, string, error) {
	client := CreateMuxGoClient(muxAuth.Id, muxAuth.Token)
	asset, err := CreateAsset(client, url)
	if err != nil {
		return "", "", fmt.Errorf("error sending video to host: %s", err)
	}

	playbackId := asset.Data.PlaybackIds[0].Id
	assetid := asset.Data.Id

	return playbackId, assetid, nil
}
