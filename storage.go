package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Storage interface {
	CreateClip(*Clip) error
	DeleteClip(string) error
	GetClipByPlaybackId(string) (*Clip, error)
	GetAllClips() ([]*Clip, error)
}

type PostgresStore struct {
	db *pgx.Conn
}

func NewPostgresStore(dbConnStr string) (*PostgresStore, error) {

	db, err := pgx.Connect(context.Background(), dbConnStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.createClipsTable()
}

func (s *PostgresStore) createClipsTable() error {
	query := `CREATE TABLE IF NOT EXISTS clips (
		playback_id varchar(200) UNIQUE NOT NULL,
		uploaded_by varchar(35) NOT NULL,
		title varchar(60) NOT NULL,
		description varchar(120) NOT NULL,
		game varchar(60) NOT NULL,
		tags varchar(100) NOT NULL,
		players varchar(255) NOT NULL,
		date_uploaded date NOT NULL,
		PRIMARY KEY ("playback_id")
	)`

	_, err := s.db.Exec(context.Background(), query)
	return err
}

func (s *PostgresStore) CreateClip(clip *Clip) error {

	query := `INSERT INTO clips VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	resp, err := s.db.Exec(context.Background(), query,
		clip.PlaybackId,
		clip.UploadedBy,
		clip.Title,
		clip.Description,
		clip.Game,
		clip.Tags,
		clip.Players,
		clip.DateUploaded,
		clip.AssetId)

	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", resp)

	return nil
}

func (s *PostgresStore) DeleteClip(playbackId string) error {
	query := "DELETE FROM clips WHERE playback_id = $1"

	_, err := s.db.Exec(context.Background(), query, playbackId)
	if err != nil {
		return fmt.Errorf("error deleting video with id `%s`, does it even exist, bro?", playbackId)
	}

	return nil
}

func (s *PostgresStore) GetClipByPlaybackId(playbackId string) (*Clip, error) {
	query := "SELECT * FROM clips WHERE playback_id = $1"

	clip := &Clip{}
	err := s.db.QueryRow(context.Background(), query, playbackId).Scan(&clip.PlaybackId, &clip.UploadedBy, &clip.Title, &clip.Description, &clip.Game, &clip.Tags, &clip.Players, &clip.DateUploaded, &clip.AssetId)
	if err != nil {
		return nil, fmt.Errorf("video with id `%s` does not exist.", playbackId)
	}

	return clip, nil
}

func (s *PostgresStore) GetAllClips() ([]*Clip, error) {
	query := "SELECT * FROM clips"

	clips := []*Clip{}
	rows, _ := s.db.Query(context.Background(), query)
	for rows.Next() {
		clip := new(Clip)
		if err := rows.Scan(&clip.PlaybackId, &clip.UploadedBy, &clip.Title, &clip.Description, &clip.Game, &clip.Tags, &clip.Players, &clip.DateUploaded, &clip.AssetId); err != nil {
			return nil, err
		}

		clips = append(clips, clip)
	}

	return clips, nil
}
