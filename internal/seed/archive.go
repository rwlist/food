package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rwlist/food/internal/blobs"
	"github.com/rwlist/food/internal/db"
)

type Archive struct {
	Name     string
	Type     string
	ID       int
	Messages []json.RawMessage
}

type Message struct {
	ID           int    `json:"id,omitempty"`
	Type         string `json:"type,omitempty"`
	Date         string `json:"date,omitempty"`
	DateUnixtime string `json:"date_unixtime,omitempty"`
	From         string `json:"from,omitempty"`
	FromID       string `json:"from_id,omitempty"`
	Photo        string `json:"photo,omitempty"`
	Text         any    `json:"text,omitempty"`
}

type ArchiveUploader struct {
	Dir     string
	Queries *db.Queries
	Photos  *blobs.Photos
}

func (a *ArchiveUploader) LoadFromArchive(ctx context.Context) error {
	archiveJSON, err := os.ReadFile(a.Dir + "/result.json")
	if err != nil {
		return err
	}

	// unmarshal archiveJSON to Archive struct
	var archive Archive
	if err := json.Unmarshal(archiveJSON, &archive); err != nil {
		return err
	}

	err = a.dbInsertPosts(ctx, archive)
	if err != nil {
		return err
	}

	err = a.uploadPhotos(ctx, archive)
	if err != nil {
		return err
	}

	return nil
}

func (a *ArchiveUploader) dbInsertPosts(ctx context.Context, archive Archive) error {
	for _, msg := range archive.Messages {
		var message Message
		if err := json.Unmarshal(msg, &message); err != nil {
			return err
		}

		unixtime, err := strconv.Atoi(message.DateUnixtime)
		if err != nil {
			return fmt.Errorf("failed to convert unixtime to int: %w", err)
		}

		timestamp := pgtype.Timestamp{
			Time:  time.Unix(int64(unixtime), 0),
			Valid: true,
		}

		text := ""
		if str, ok := message.Text.(string); ok {
			text = str
		} else if arr, ok := message.Text.([]interface{}); ok {
			for _, v := range arr {
				if str, ok := v.(string); ok {
					text += str + "\n"
				} else if mapStr, ok := v.(map[string]interface{}); ok {
					textValue, ok := mapStr["text"].(string)
					if !ok {
						return fmt.Errorf("unexpected type of mapStr[text]: %T", mapStr["text"])
					}
					text += textValue
				}
			}
		} else {
			return fmt.Errorf("unexpected type of message.Text: %T", message.Text)
		}

		rows, err := a.Queries.InsertFoodPost(ctx, db.InsertFoodPostParams{
			MessageID:   int64(message.ID),
			PostTs:      timestamp,
			PhotoKey:    "",
			CaptionText: text,
			Raw:         msg,
		})
		if err != nil {
			return fmt.Errorf("failed to insert post: %w", err)
		}

		slog.Info("imported post", "rows", rows, "msg", message)
	}
	return nil
}

func (a *ArchiveUploader) uploadPhotos(ctx context.Context, archive Archive) error {
	for _, msg := range archive.Messages {
		var message Message
		if err := json.Unmarshal(msg, &message); err != nil {
			return err
		}

		if message.Photo == "" {
			continue
		}

		post, err := a.Queries.FindFoodPostByMessageId(ctx, int64(message.ID))
		if err != nil {
			return fmt.Errorf("failed to find post by message_id: %w", err)
		}

		if post.PhotoKey != "" {
			continue
		}

		photoPath := fmt.Sprintf("%s/%s", a.Dir, message.Photo)

		// upload photo to r2
		photoKey, err := a.Photos.UploadFile(ctx, photoPath)
		if err != nil {
			return fmt.Errorf("failed to upload photo: %w", err)
		}

		// update photo_key in db
		rows, err := a.Queries.UpdateFoodPostSetPhoto(ctx, db.UpdateFoodPostSetPhotoParams{
			MessageID: int64(message.ID),
			PhotoKey:  photoKey,
		})
		if err != nil {
			return fmt.Errorf("failed to update photo_key: %w", err)
		}

		if rows != 1 {
			return fmt.Errorf("unexpected rows affected: %d", rows)
		}
	}
	return nil
}
