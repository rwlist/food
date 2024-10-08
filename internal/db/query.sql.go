// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const findFoodPostByMessageId = `-- name: FindFoodPostByMessageId :one
SELECT message_id, post_ts, photo_key, caption_text, raw FROM food_posts WHERE message_id = $1
`

func (q *Queries) FindFoodPostByMessageId(ctx context.Context, messageID int64) (FoodPost, error) {
	row := q.db.QueryRow(ctx, findFoodPostByMessageId, messageID)
	var i FoodPost
	err := row.Scan(
		&i.MessageID,
		&i.PostTs,
		&i.PhotoKey,
		&i.CaptionText,
		&i.Raw,
	)
	return i, err
}

const findFoodPostsByDateRange = `-- name: FindFoodPostsByDateRange :many
SELECT message_id, post_ts, photo_key, caption_text, raw FROM food_posts
WHERE post_ts >= $1 AND post_ts <= $2 ORDER BY post_ts ASC
`

type FindFoodPostsByDateRangeParams struct {
	FromTs  pgtype.Timestamp
	UntilTs pgtype.Timestamp
}

func (q *Queries) FindFoodPostsByDateRange(ctx context.Context, arg FindFoodPostsByDateRangeParams) ([]FoodPost, error) {
	rows, err := q.db.Query(ctx, findFoodPostsByDateRange, arg.FromTs, arg.UntilTs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FoodPost
	for rows.Next() {
		var i FoodPost
		if err := rows.Scan(
			&i.MessageID,
			&i.PostTs,
			&i.PhotoKey,
			&i.CaptionText,
			&i.Raw,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertFoodPost = `-- name: InsertFoodPost :execrows
INSERT INTO food_posts (
    message_id, post_ts, photo_key, caption_text, raw
) VALUES (
    $1, $2, $3, $4, $5
)
ON CONFLICT (message_id) DO NOTHING
RETURNING message_id, post_ts, photo_key, caption_text, raw
`

type InsertFoodPostParams struct {
	MessageID   int64
	PostTs      pgtype.Timestamp
	PhotoKey    string
	CaptionText string
	Raw         []byte
}

func (q *Queries) InsertFoodPost(ctx context.Context, arg InsertFoodPostParams) (int64, error) {
	result, err := q.db.Exec(ctx, insertFoodPost,
		arg.MessageID,
		arg.PostTs,
		arg.PhotoKey,
		arg.CaptionText,
		arg.Raw,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const updateFoodPostSetPhoto = `-- name: UpdateFoodPostSetPhoto :execrows
UPDATE food_posts SET photo_key = $2 WHERE message_id = $1
`

type UpdateFoodPostSetPhotoParams struct {
	MessageID int64
	PhotoKey  string
}

func (q *Queries) UpdateFoodPostSetPhoto(ctx context.Context, arg UpdateFoodPostSetPhotoParams) (int64, error) {
	result, err := q.db.Exec(ctx, updateFoodPostSetPhoto, arg.MessageID, arg.PhotoKey)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
