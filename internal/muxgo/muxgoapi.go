package muxgo

import (
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
